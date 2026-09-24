package main

// ストップウォッチは RTC を使用せず、Pico の CPU 上で経過時間を管理します。
// 計測値は 00:00 ～ 99:59 の範囲で表示します。

import "time"

// stopwatchState はストップウォッチの状態を表します。
type stopwatchState uint8

const (
	// stopwatchStopped は停止中または待機中の状態です。
	stopwatchStopped stopwatchState = iota
	// stopwatchRunning は計測中の状態です。
	stopwatchRunning
)

const (
	// standbyScrollDelay はスタートせずに待機してから文字列を表示するまでの時間です。
	standbyScrollDelay = time.Minute

	// standbyScrollStartX は、スクロール文字列を表示領域の右端から開始する位置です。
//	standbyScrollStartX = displayColumns
)

// App はストップウォッチ全体の状態を保持します。
type App struct {
	display *Display
	buzzer  *Buzzer

	state stopwatchState

	// elapsedSeconds は現在までに計測した経過秒数です。
	elapsedSeconds uint32

	// lastTick は計測中に前回 1 秒処理を行った時刻です。
	lastTick time.Time

	// standbySince は停止状態になってからの待機開始時刻です。
	standbySince time.Time

	// scrolling は待機中の文字列スクロール表示中であることを示します。
	scrolling bool

	// scrollText はスクロール表示する文字列を保持するバッファです。
	// 容量は config.go の scrollTextBufferSize で定義しています。
	scrollText [scrollTextBufferSize]byte

	// scrollTextLength は scrollText に格納されている文字数です。
	scrollTextLength int

	// scrollSteps はスクロールを開始してからの移動回数です。
	scrollSteps int

	// lastScroll は直前にスクロールした時刻です。
	lastScroll time.Time
}

// NewApp はストップウォッチを初期化します。
func NewApp(display *Display, buzzer *Buzzer) *App {
	now := time.Now()
	a := &App{
		display:      display,
		buzzer:       buzzer,
		state:        stopwatchStopped,
		standbySince: now,
	}
	a.setScrollText(standbyScrollText)
	return a
}

// ToggleStartStop は計測を開始／停止します。
// 停止中に押すと計測開始、計測中に押すと一時停止します。
// もう一度押すと停止した位置から計測を再開します。
func (a *App) ToggleStartStop() {
	if a.state == stopwatchRunning {
		a.state = stopwatchStopped
		a.startStandby()
		if a.buzzer != nil {
			a.buzzer.StopSound()
		}
		return
	}

	// スクロール表示中であっても、スタートボタンを押せば直ちに計測へ移行します。
	a.stopScroll()
	a.state = stopwatchRunning
	if a.buzzer != nil {
		a.buzzer.StartSound()
	}
	a.lastTick = time.Now()
	a.render()
}

// Reset は計測値を 00:00 に戻し、内部カウンタも 0 にします。
func (a *App) Reset() {
	a.stopScroll()
	a.state = stopwatchStopped
	if a.buzzer != nil {
		a.buzzer.ResetSound()
	}
	a.elapsedSeconds = 0
	a.lastTick = time.Time{}
	a.startStandby()
	a.render()
}

// Tick は定期的に呼び出され、計測中なら経過時間を更新します。
// 停止状態で1分以上待機している場合は、STOP WATCH をスクロール表示します。
func (a *App) Tick() {
	if a.state == stopwatchRunning {
		a.tickStopwatch()
		return
	}

	a.tickStandbyScroll()
}

// tickStopwatch は計測中の経過時間を更新します。
func (a *App) tickStopwatch() {
	now := time.Now()
	elapsed := now.Sub(a.lastTick)

	// 起動直後などの異常な時刻差は無視します。
	if elapsed < 0 {
		a.lastTick = now
		return
	}

	seconds := uint32(elapsed / time.Second)
	if seconds == 0 {
		return
	}

	a.lastTick = a.lastTick.Add(time.Duration(seconds) * time.Second)

	// 99分59秒を超えたら、最大値で停止します。
	const maxSeconds = 99*60 + 59
	if a.elapsedSeconds+seconds >= maxSeconds {
		a.elapsedSeconds = maxSeconds
		a.state = stopwatchStopped
		a.startStandby()
	} else {
		a.elapsedSeconds += seconds
	}

	a.render()
}

// tickStandbyScroll は停止中の待機時間を監視し、必要なら文字列をスクロールします。
func (a *App) tickStandbyScroll() {
	now := time.Now()

	if !a.scrolling {
		if now.Sub(a.standbySince) >= standbyScrollDelay {
			a.startScroll(now)
		}
		return
	}

	if now.Sub(a.lastScroll) < standbyScrollInterval {
		return
	}

	a.lastScroll = now
	a.scrollSteps++
	a.renderScroll()

	if a.scrollSteps >= a.scrollTotalSteps() {
		a.stopScroll()
		a.startStandby()
		a.render()
	}
}

// startStandby は停止状態の待機時間を現在から測り直します。
func (a *App) startStandby() {
	a.standbySince = time.Now()
	a.scrolling = false
	a.scrollSteps = 0
}

// setScrollText は、スクロール表示用文字列を内部バッファへコピーします。
// バッファ容量を超える文字列は、容量に収まる範囲だけを使用します。
func (a *App) setScrollText(text string) {
	length := len(text)
	if length > len(a.scrollText) {
		length = len(a.scrollText)
	}
	copy(a.scrollText[:length], text[:length])
	a.scrollTextLength = length
}

// scrollTotalSteps は、文字列全体が表示領域を通過し終わるまでの移動回数を返します。
func (a *App) scrollTotalSteps() int {
	return standbyScrollStartX + a.scrollTextLength*5
}

// startScroll は、設定された文字列のスクロール表示を開始します。
func (a *App) startScroll(now time.Time) {
	a.scrolling = true
	a.scrollSteps = 0
	a.lastScroll = now
	a.renderScroll()
}

// renderScroll は、現在のスクロール位置に合わせて表示領域を描画します。
func (a *App) renderScroll() {
	a.display.Clear()
	for i := 0; i < a.scrollTextLength; i++ {
		x := standbyScrollStartX - a.scrollSteps + i*5
		if x >= displayColumns || x <= -5 {
			continue
		}
		a.display.DisplayCharClipped(x, a.scrollText[i])
	}
}

// stopScroll はスクロール表示を終了します。
func (a *App) stopScroll() {
	a.scrolling = false
	a.scrollSteps = 0
}

// render は現在の計測時間を mm:ss 形式で表示します。
func (a *App) render() {
	minutes := a.elapsedSeconds / 60
	seconds := a.elapsedSeconds % 60

	a.display.Clear()

	// MM:SS
	a.display.DisplayChar(0, byte('0'+minutes/10))
	a.display.DisplayChar(5, byte('0'+minutes%10))
	a.display.DisplayChar(10, ':')
	a.display.DisplayChar(13, byte('0'+seconds/10))
	a.display.DisplayChar(18, byte('0'+seconds%10))
}
