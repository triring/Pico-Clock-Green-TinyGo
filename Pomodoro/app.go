package main

import "time"

// timerMode はポモドーロタイマーの時間モードを表します。
type timerMode uint8

const (
	modeWork timerMode = iota
	modeRest
)

// timerState はタイマーの動作状態を表します。
type timerState uint8

const (
	timerIdle timerState = iota
	timerRunning
	timerPaused
	timerAlarm
)

// App はポモドーロタイマー全体の状態を保持します。
type App struct {
	display *Display
	buzzer  *Buzzer

	state timerState
	mode  timerMode

	// remainingSeconds は現在表示・計測している残り時間です。
	remainingSeconds uint32

	lastTick time.Time

	// standbySince はキー入力待ちを開始した時刻です。
	standbySince time.Time

	// scrolling は待機中の文字列をスクロール表示していることを示します。
	scrolling bool

	// scrollText はスクロール表示する文字列を保持するバッファです。
	scrollText [scrollTextBufferSize]byte

	// scrollTextLength は scrollText に格納されている文字数です。
	scrollTextLength int

	// scrollSteps はスクロールを開始してからの移動回数です。
	scrollSteps int

	// lastScroll は直前にスクロールした時刻です。
	lastScroll time.Time
}

// NewApp はポモドーロタイマーを初期化します。
// 起動直後は25分モードで、LEDパネルに 25MIN を表示します。
func NewApp(display *Display, buzzer *Buzzer) *App {
	a := &App{
		display:          display,
		buzzer:           buzzer,
		state:            timerIdle,
		mode:             modeWork,
		remainingSeconds: uint32(workDuration / time.Second),
		standbySince:     time.Now(),
	}
	a.setScrollText(standbyScrollText)
	a.renderMode()
	return a
}

// ModeShortPress は MODE 短押しを処理します。
// カウントダウン中および終了音中は、入力処理側で別扱いにするためここでは無視します。
func (a *App) ModeShortPress() {
	if a.state != timerIdle && a.state != timerPaused {
		return
	}

	if a.mode == modeWork {
		a.mode = modeRest
	} else {
		a.mode = modeWork
	}
	a.remainingSeconds = a.modeDurationSeconds()
	a.startStandby()
	a.renderMode()
}

// StartStopShortPress は START/STOP 短押しを処理します。
func (a *App) StartStopShortPress() {
	switch a.state {
	case timerIdle:
		if a.remainingSeconds == 0 {
			a.remainingSeconds = a.modeDurationSeconds()
		}
		a.state = timerRunning
		a.lastTick = time.Now()
		a.stopScroll()
		a.play(startToneDuration)
		a.render()

	case timerRunning:
		a.state = timerPaused
		a.startStandby()
		a.play(stopToneDuration)
		a.render()

	case timerPaused:
		a.state = timerRunning
		a.lastTick = time.Now()
		a.stopScroll()
		a.play(startToneDuration)
		a.render()
	}
}

// StartStopLongPress は START/STOP 長押しを処理します。
// 現在のモードの設定時間へ戻し、表示もその時間に戻します。
func (a *App) StartStopLongPress() {
	if a.state == timerAlarm {
		a.HandleAlarmButton()
		return
	}

	a.state = timerIdle
	a.remainingSeconds = a.modeDurationSeconds()
	a.lastTick = time.Time{}
	a.startStandby()
	a.play(resetToneDuration)
	a.renderMode()
}

// HandleAlarmButton は終了音中のボタン操作を処理します。
// 終了音だけを中断し、押されたボタン自身の通常操作は実行しません。
func (a *App) HandleAlarmButton() {
	if a.state != timerAlarm {
		return
	}
	if a.buzzer != nil {
		a.buzzer.StopAlarm()
	}
	a.state = timerIdle
	a.remainingSeconds = a.modeDurationSeconds()
	a.startStandby()
	a.renderMode()
}

// Tick は定期的に呼び出され、カウントダウンと待機中のスクロールを更新します。
func (a *App) Tick() {
	switch a.state {
	case timerRunning:
		a.tickRunning()
	case timerAlarm:
		a.tickAlarm()
	case timerIdle, timerPaused:
		a.tickStandbyScroll()
	}
}

func (a *App) tickAlarm() {
	// 終了音が正常に最後まで再生されたら、次のモードへ自動切替します。
	// 終了音中にボタンが押された場合は HandleAlarmButton が timerIdle に
	// 戻すため、この処理は実行されません。
	if a.buzzer != nil && a.buzzer.IsAlarmActive() {
		return
	}

	if a.mode == modeWork {
		a.mode = modeRest
	} else {
		a.mode = modeWork
	}
	a.remainingSeconds = a.modeDurationSeconds()
	a.state = timerIdle
	a.startStandby()
	a.renderMode()
}

func (a *App) tickRunning() {
	now := time.Now()
	elapsed := now.Sub(a.lastTick)
	if elapsed < 0 {
		a.lastTick = now
		return
	}

	seconds := uint32(elapsed / time.Second)
	if seconds == 0 {
		return
	}
	a.lastTick = a.lastTick.Add(time.Duration(seconds) * time.Second)

	if seconds >= a.remainingSeconds {
		a.remainingSeconds = 0
		a.state = timerAlarm
		a.render()
		if a.buzzer != nil {
			a.buzzer.PlayPomodoroEnd(a.mode)
		} else {
			a.tickAlarm()
		}
		return
	}

	a.remainingSeconds -= seconds
	a.render()
}

func (a *App) modeDurationSeconds() uint32 {
	if a.mode == modeWork {
		return uint32(workDuration / time.Second)
	}
	return uint32(restDuration / time.Second)
}

func (a *App) modeLabel() string {
	if a.mode == modeWork {
		return "25MIN"
	}
	return " 5MIN"
}

// renderMode は待機中に現在のモードを表示します。
func (a *App) renderMode() {
	a.display.Clear()
	text := a.modeLabel()
	for i := 0; i < len(text); i++ {
		// MINの表示のNが見切れるので、2ドット補正する。
		if 'N' == text[i] {
			a.display.DisplayChar(byte(i*5) - 2, text[i])
		} else {
			a.display.DisplayChar(byte(i*5), text[i])
		}
	}
}

// render は現在の残り時間を mm:ss 形式で表示します。
func (a *App) render() {
	minutes := a.remainingSeconds / 60
	seconds := a.remainingSeconds % 60
	a.display.Clear()
	a.display.DisplayChar(0, byte('0'+minutes/10))
	a.display.DisplayChar(5, byte('0'+minutes%10))
	a.display.DisplayChar(10, ':')
	a.display.DisplayChar(13, byte('0'+seconds/10))
	a.display.DisplayChar(18, byte('0'+seconds%10))
}

// HandleScrollButton は、スクロール表示中のボタン操作を処理します。
// スクロールだけを中断し、そのボタン自身の通常操作は実行しません。
func (a *App) HandleScrollButton() {
	if !a.scrolling {
		return
	}
	a.stopScroll()
	a.startStandby()
	if a.state == timerPaused {
		a.render()
	} else {
		a.renderMode()
	}
}

// IsScrolling は、現在スクロール表示中かどうかを返します。
func (a *App) IsScrolling() bool {
	return a.scrolling
}

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
		if a.state == timerPaused {
			a.render()
		} else {
			a.renderMode()
		}
	}
}

func (a *App) startStandby() {
	a.standbySince = time.Now()
	a.scrolling = false
	a.scrollSteps = 0
}

func (a *App) setScrollText(text string) {
	length := len(text)
	if length > len(a.scrollText) {
		length = len(a.scrollText)
	}
	copy(a.scrollText[:length], text[:length])
	a.scrollTextLength = length
}

func (a *App) scrollTotalSteps() int {
	return standbyScrollStartX + a.scrollTextLength*5
}

func (a *App) startScroll(now time.Time) {
	a.scrolling = true
	a.scrollSteps = 0
	a.lastScroll = now
	a.renderScroll()
}

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

func (a *App) stopScroll() {
	a.scrolling = false
	a.scrollSteps = 0
}

func (a *App) play(duration time.Duration) {
	if a.buzzer != nil {
		a.buzzer.Play(duration)
	}
}
