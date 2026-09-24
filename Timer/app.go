package main

import "time"

// timerState はカウントダウンタイマーの状態を表します。
type timerState uint8

const (
	timerIdle timerState = iota
	timerSetting
	timerRunning
	timerPaused
	timerAlarm
)

// settingDigit は設定対象の桁を表します。
type settingDigit uint8

const (
	digitTenMinutes settingDigit = iota
	digitMinutes
	digitTenSeconds
	digitSeconds
)

const maxTimerSeconds = 99*60 + 59

// App はカウントダウンタイマー全体の状態を保持します。
type App struct {
	display *Display
	buzzer  *Buzzer

	state timerState

	// configuredSeconds は MODE で最後に設定して確定した時間です。
	configuredSeconds uint32
	// remainingSeconds は現在表示・計測している残り時間です。
	remainingSeconds uint32

	lastTick time.Time

	selectedDigit settingDigit
	blinkOn       bool
	lastBlink     time.Time

	// standbySince は、キー入力待ちを開始した時刻です。
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

// NewApp はカウントダウンタイマーを初期化します。
// 起動直後の表示は 00:00 とします。
func NewApp(display *Display, buzzer *Buzzer) *App {
	a := &App{
		display:           display,
		buzzer:            buzzer,
		state:             timerIdle,
		configuredSeconds: 0,
		remainingSeconds:  0,
		blinkOn:           true,
		standbySince:      time.Now(),
	}
	a.setScrollText(standbyScrollText)
	a.render()
	return a
}

// ModeLongPress は MODE 長押しを処理します。
// カウントダウン中は仕様により無視します。
func (a *App) ModeLongPress() {
	switch a.state {
	case timerIdle, timerPaused:
		a.enterSettingMode()
	case timerSetting:
		a.leaveSettingMode(false)
	}
}

// ModeShortPress は MODE 短押しを処理します。
func (a *App) ModeShortPress() {
	if a.state != timerSetting {
		return
	}
	if a.selectedDigit == digitSeconds {
		a.leaveSettingMode(true)
		return
	}
	a.selectedDigit++
	a.blinkOn = true
	a.lastBlink = time.Now()
	a.play(modeStepToneDuration)
	a.renderSetting()
}

func (a *App) enterSettingMode() {
	if a.state == timerRunning || a.state == timerAlarm {
		return
	}
	a.state = timerSetting
	a.remainingSeconds = a.configuredSeconds
	a.selectedDigit = digitTenMinutes
	a.blinkOn = true
	a.lastBlink = time.Now()
	a.play(modeEnterToneDuration)
	a.renderSetting()
}

func (a *App) leaveSettingMode(shortExit bool) {
	a.configuredSeconds = a.remainingSeconds
	a.state = timerIdle
	a.blinkOn = true
	a.startStandby()
	if shortExit {
		a.play(modeExitToneDuration)
	} else {
		a.play(modeLeaveToneDuration)
	}
	a.render()
}

// UpShortPress は UP 短押しを処理します。
func (a *App) UpShortPress() {
	if a.state != timerSetting {
		return
	}
	a.incrementSelectedDigit()
	a.play(upToneDuration)
	a.renderSetting()
}

// UpRepeat は UP 長押しによる高速加算を1回分処理します。
func (a *App) UpRepeat() {
	if a.state != timerSetting {
		return
	}
	a.incrementSelectedDigit()
	a.renderSetting()
}

func (a *App) incrementSelectedDigit() {
	minutes := a.remainingSeconds / 60
	seconds := a.remainingSeconds % 60

	switch a.selectedDigit {
	case digitTenMinutes:
		v := minutes / 10
		v = (v + 1) % 10
		minutes = v*10 + minutes%10
	case digitMinutes:
		v := minutes % 10
		v = (v + 1) % 10
		minutes = (minutes/10)*10 + v
	case digitTenSeconds:
		v := seconds / 10
		v = (v + 1) % 6
		seconds = v*10 + seconds%10
	case digitSeconds:
		v := seconds % 10
		v = (v + 1) % 10
		seconds = (seconds/10)*10 + v
	}

	a.remainingSeconds = minutes*60 + seconds
}

// StartStopShortPress は START/STOP 短押しを処理します。
func (a *App) StartStopShortPress() {
	switch a.state {
	case timerIdle:
		if a.remainingSeconds == 0 {
			a.remainingSeconds = a.configuredSeconds
		}
		if a.remainingSeconds == 0 {
			return
		}
		a.state = timerRunning
		a.lastTick = time.Now()
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
		a.play(startToneDuration)
		a.render()
	}
}

// StartStopLongPress は START/STOP 長押しを処理します。
// 設定モード中は無視し、通常時は記憶した設定時間へ戻します。
func (a *App) StartStopLongPress() {
	if a.state == timerAlarm {
		a.HandleAlarmButton()
		return
	}
	if a.state == timerSetting {
		return
	}

	// START/STOP 長押しは、現在のタイマーを完全にリセットします。
	// 記憶していた設定時間も 00:00 に戻すため、次の短押しで
	// リセット前の時間が再び読み込まれることはありません。
	a.state = timerIdle
	a.configuredSeconds = 0
	a.remainingSeconds = 0
	a.lastTick = time.Time{}
	a.startStandby()
	a.play(resetToneDuration)
	a.render()
}

// HandleAlarmButton は終了音中のボタン操作を処理します。
// ボタン操作は音を止めるだけで、ボタン自体の機能は実行しません。
func (a *App) HandleAlarmButton() {
	if a.state != timerAlarm {
		return
	}
	if a.buzzer != nil {
		a.buzzer.StopAlarm()
	}
	a.state = timerIdle
	// 終了音を中断した時点で、記憶している設定時間を表示し、
	// 次の START/STOP 操作で再スタートできる状態にします。
	a.remainingSeconds = a.configuredSeconds
	a.startStandby()
	a.render()
}

// Tick は定期的に呼び出され、設定中の点滅とカウントダウンを更新します。
func (a *App) Tick() {
	switch a.state {
	case timerSetting:
		a.tickSetting()
	case timerRunning:
		a.tickRunning()
	case timerIdle:
		a.tickStandbyScroll()
	}
}

func (a *App) tickSetting() {
	now := time.Now()
	if now.Sub(a.lastBlink) >= settingBlinkInterval {
		a.lastBlink = a.lastBlink.Add(settingBlinkInterval)
		a.blinkOn = !a.blinkOn
		a.renderSetting()
	}
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
			a.buzzer.PlayJISS0013End()
		}
		return
	}
	a.remainingSeconds -= seconds
	a.render()
}

// HandleScrollButton は、スクロール表示中のボタン操作を処理します。
// いずれかのボタンが押された場合はスクロールだけを中断し、
// そのボタン自身の通常操作は実行しません。
func (a *App) HandleScrollButton() {
	if !a.scrolling {
		return
	}
	a.stopScroll()
	a.startStandby()
	a.render()
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
		a.render()
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

func (a *App) renderSetting() {
	minutes := a.remainingSeconds / 60
	seconds := a.remainingSeconds % 60
	a.display.Clear()

	chars := [5]byte{
		byte('0' + minutes/10),
		byte('0' + minutes%10),
		':',
		byte('0' + seconds/10),
		byte('0' + seconds%10),
	}
	positions := [5]byte{0, 5, 10, 13, 18}

	displayIndex := [4]int{0, 1, 3, 4}

	for i := 0; i < len(chars); i++ {
		selected := (i == displayIndex[int(a.selectedDigit)])
		if selected && !a.blinkOn {
			continue
		}
		a.display.DisplayChar(positions[i], chars[i])
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

func (a *App) play(duration time.Duration) {
	if a.buzzer != nil {
		a.buzzer.Play(duration)
	}
}
