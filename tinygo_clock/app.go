package main

import (
	"time"
)

type uiMode uint8

const (
	modeClock uiMode = iota
	modeSettings
)

type settingItem uint8

const (
	// 時刻設定は「年 → 月 → 日 → 時 → 分」の順で行います。
	settingYear settingItem = iota + 1
	settingMonth
	settingDay
	settingHour
	settingMinute
	settingBeep
	settingScroll
	settingFormat
	settingChime
)

// App は時計アプリケーション全体の状態を保持します。
// 元の C プログラムに多数存在したグローバルフラグをまとめることで、保守性を高めています。
type App struct {
	display *Display
	mode    uiMode
	setting settingItem

	now           time.Time
	hour12        bool
	beepEnabled   bool
	scrollEnabled bool
	chimeEnabled  bool

	lastInput             time.Time
	scrollDue             int
	scrollActive          bool
	scrollCounter         int
	scrollSteps           int
	beepUntil             time.Time
	chimeCount            int
	settingFeedbackUntil  time.Time
	settingFeedbackPhase  int
	settingFeedbackActive bool
	settingBlinkVisible   bool
	settingBlinkCounter   int
}

// NewApp は時計アプリケーションの状態を生成します。
// 起動時は既定日時から開始し、設定モードで「年 → 月 → 日 → 時 → 分」を合わせます。
func NewApp(display *Display) *App {
	a := &App{
		display:             display,
		now:                 time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		beepEnabled:         true,
		scrollEnabled:       true,
		chimeEnabled:        false,
		lastInput:           time.Now(),
		settingBlinkVisible: true,
	}
	return a
}

func (a *App) beep() {
	if !a.beepEnabled {
		return
	}
	pinBuzz.High()
	a.beepUntil = time.Now().Add(80 * time.Millisecond)
}

func (a *App) serviceBeep() {
	if !a.beepUntil.IsZero() && time.Now().After(a.beepUntil) {
		pinBuzz.Low()
		a.beepUntil = time.Time{}
	}
}

// startSettingFeedback は設定モードへの移行を知らせるフィードバックを開始します。
// 約1秒間、LED表示を点滅させ、同時に短いブザー音を鳴らします。
// 通常の「ビープ音OFF」設定に関係なく、モード移行の通知として動作します。
func (a *App) startSettingFeedback() {
	now := time.Now()
	a.settingFeedbackUntil = now.Add(900 * time.Millisecond)
	a.settingFeedbackPhase = 0
	a.settingFeedbackActive = true
	a.settingBlinkVisible = true
	a.settingBlinkCounter = 0

	// 設定モードに入ったことを明確に知らせるため、強制的に短い確認音を出す。
	pinBuzz.High()
	a.beepUntil = now.Add(120 * time.Millisecond)

	// 最初は表示を点灯させ、ServiceFeedback で点滅させる。
	a.display.SetEnabled(true)
}

// ServiceFeedback は設定モード移行時の画面点滅を50 ms周期で更新します。
func (a *App) ServiceFeedback() {
	// 設定モード移行直後の画面全体の点滅を処理します。
	if a.settingFeedbackActive {
		now := time.Now()
		if now.After(a.settingFeedbackUntil) {
			a.settingFeedbackActive = false
			a.display.SetEnabled(true)
			a.settingBlinkVisible = true
			a.settingBlinkCounter = 0
			a.render()
		} else {
			a.settingFeedbackPhase++
			// 100 ms周期程度で表示をON/OFFして、明確な点滅として見せる。
			a.display.SetEnabled((a.settingFeedbackPhase/2)%2 == 0)
		}
		return
	}

	// 設定モードでは、現在選択中の設定項目だけを点滅させます。
	// 50 ms周期で呼ばれるため、5回ごと（約250 ms）に表示/消灯を切り替えます。
	if a.mode != modeSettings {
		return
	}
	a.settingBlinkCounter++
	if a.settingBlinkCounter >= 5 {
		a.settingBlinkCounter = 0
		a.settingBlinkVisible = !a.settingBlinkVisible
		a.render()
	}
}

// Tick は1秒周期でソフトウェア時計を進め、時報と表示を処理します。
// DS3231などの外部RTCにはアクセスしません。
func (a *App) Tick() {
	a.serviceBeep()
	a.now = a.now.Add(time.Second)

	if a.chimeEnabled && a.now.Minute() == 0 && a.now.Second() == 0 {
		a.chimeCount = 5
	}
	if a.chimeCount > 0 {
		a.beep()
		a.chimeCount--
	}

	if a.mode == modeClock {
		a.scrollDue++
		if a.scrollEnabled && a.scrollDue >= 180 && !a.scrollActive {
			a.startScroll()
			a.scrollDue = 0
		}
	}
	a.render()
}

func (a *App) startScroll() {
	a.scrollActive = true
	a.scrollCounter = 0
	a.scrollSteps = 0
	a.display.ClearFrom(24)
	a.display.DisplayChar(32, byte('2'))
	a.display.DisplayChar(37, byte('0'))
	a.display.DisplayChar(42, byte('0'+a.now.Year()/10%10))
	a.display.DisplayChar(47, byte('0'+a.now.Year()%10))
	a.display.DisplayChar(52, '-')
	a.display.DisplayChar(55, byte('0'+int(a.now.Month())/10))
	a.display.DisplayChar(60, byte('0'+int(a.now.Month())%10))
	a.display.DisplayChar(65, '-')
	a.display.DisplayChar(68, byte('0'+a.now.Day()/10))
	a.display.DisplayChar(73, byte('0'+a.now.Day()%10))
	// DS3231を使用しないため、温度表示は行いません。
}

// ServiceScroll はスクロール表示を進めます。50 ms 周期で呼び出します。
func (a *App) ServiceScroll() {
	if !a.scrollActive {
		return
	}
	a.scrollCounter++
	if a.scrollCounter < 3 {
		return
	}
	a.scrollCounter = 0
	a.display.ScrollLeft()
	a.scrollSteps++
	// メッセージが十分に左へ移動したらスクロール表示を終了します。
	if a.scrollSteps > 120 {
		a.scrollSteps = 0
		a.scrollActive = false
		a.render()
	}
}

// render は現在の UI 状態に応じて24列の表示内容を更新します。
func (a *App) render() {
	switch a.mode {
	case modeClock:
		a.renderClock()
	case modeSettings:
		a.renderSettings()
	}
}

func (a *App) renderClock() {
	if a.scrollActive {
		return
	}
	a.display.Clear()
	h := a.now.Hour()
	if a.hour12 {
		if h == 0 {
			h = 12
		} else if h > 12 {
			h -= 12
		}
	}
	a.display.DisplayChar(0, byte('0'+h/10))
	a.display.DisplayChar(5, byte('0'+h%10))
	a.display.DisplayChar(10, ':')
	a.display.DisplayChar(13, byte('0'+a.now.Minute()/10))
	a.display.DisplayChar(18, byte('0'+a.now.Minute()%10))
	a.display.SetWeekdayIndicator(goWeekdayToClock(a.now.Weekday()))
	a.display.SetStatus(4, a.hour12 && a.now.Hour() < 12)
	a.display.SetStatus(5, a.hour12 && a.now.Hour() >= 12)
	a.display.SetStatus(0, a.scrollEnabled)
	a.display.SetStatus(7, a.chimeEnabled)
}

func goWeekdayToClock(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}

func (a *App) renderSettings() {
	a.display.Clear()

	// settingBlinkVisible が false のときは、現在選択中の項目を表示しません。
	// これにより「どの項目を変更できる状態なのか」が一目で分かります。
	visible := a.settingBlinkVisible

	switch a.setting {
	case settingYear:
		if visible {
			a.display.DisplayChar(0, byte('0'+a.now.Year()/1000))
			a.display.DisplayChar(5, byte('0'+a.now.Year()/100%10))
			a.display.DisplayChar(10, byte('0'+a.now.Year()/10%10))
			a.display.DisplayChar(15, byte('0'+a.now.Year()%10))
		}

	case settingMonth:
		if visible {
			a.display.DisplayChar(0, byte('0'+int(a.now.Month())/10))
			a.display.DisplayChar(5, byte('0'+int(a.now.Month())%10))
		}

	case settingDay:
		if visible {
			a.display.DisplayChar(0, byte('0'+a.now.Day()/10))
			a.display.DisplayChar(5, byte('0'+a.now.Day()%10))
		}

	case settingHour:
		if visible {
			a.display.DisplayChar(0, byte('0'+a.now.Hour()/10))
			a.display.DisplayChar(5, byte('0'+a.now.Hour()%10))
		}
		a.display.DisplayChar(10, ':')
		a.display.DisplayChar(13, byte('0'+a.now.Minute()/10))
		a.display.DisplayChar(18, byte('0'+a.now.Minute()%10))

	case settingMinute:
		a.display.DisplayChar(0, byte('0'+a.now.Hour()/10))
		a.display.DisplayChar(5, byte('0'+a.now.Hour()%10))
		a.display.DisplayChar(10, ':')
		if visible {
			a.display.DisplayChar(13, byte('0'+a.now.Minute()/10))
			a.display.DisplayChar(18, byte('0'+a.now.Minute()%10))
		}

	case settingBeep: // Beep音のOn/Off設定
		a.display.DisplayChar(0, 'B')
		a.display.DisplayChar(5, 'P')
		a.display.DisplayChar(10, ':')
		if visible {
			//	a.display.DisplayChar(13, ternaryByte(a.beepEnabled, 'N', 'F'))
			a.display.DisplayChar(13, ternaryByte(a.beepEnabled, 'O', 'O'))
			a.display.DisplayChar(18, ternaryByte(a.beepEnabled, 'N', 'F'))
		}

	case settingScroll: // スクロール表示の設定
		a.display.DisplayChar(0, 'D')
		a.display.DisplayChar(5, 'P')
		a.display.DisplayChar(10, ':')
		if visible {
			//	a.display.DisplayChar(18, ternaryByte(a.scrollEnabled, 'N', 'F'))
			a.display.DisplayChar(13, ternaryByte(a.scrollEnabled, 'O', 'O'))
			a.display.DisplayChar(18, ternaryByte(a.scrollEnabled, 'N', 'F'))
		}

	case settingFormat: // 12時間・24時間表示の変更
		a.display.DisplayChar(0, 'M')
		a.display.DisplayChar(6, 'D')
		a.display.DisplayChar(11, ':')
		if visible {
			//	a.display.DisplayChar(14, ternaryByte(a.hour12, '1', '2'))
			a.display.DisplayChar(14, ternaryByte(a.hour12, '1', '2'))
			a.display.DisplayChar(18, ternaryByte(a.hour12, '2', '4'))
		}

	case settingChime: // 時報のOn/Off設定
		a.display.DisplayChar(0, 'F')
		a.display.DisplayChar(5, 'T')
		a.display.DisplayChar(10, ':')
		if visible {
			//	a.display.DisplayChar(18, ternaryByte(a.chimeEnabled, 'N', 'F'))
			a.display.DisplayChar(13, ternaryByte(a.chimeEnabled, 'O', 'O'))
			a.display.DisplayChar(18, ternaryByte(a.chimeEnabled, 'N', 'F'))
		}
	}
}

func ternaryByte(on bool, yes, no byte) byte {
	if on {
		return yes
	}
	return no
}

// ButtonAction は、短押しまたは長押しされたボタンイベントを処理します。
func (a *App) ButtonAction(button byte, long bool) {
	a.lastInput = time.Now()
	if long {
		if button == 'D' && a.mode != modeClock {
			a.exitMode()
			a.beep()
		}
		a.render()
		return
	}

	switch a.mode {
	case modeClock:
		switch button {
		case 'S':
			a.mode = modeSettings
			a.setting = settingYear
			a.settingBlinkVisible = true
			a.settingBlinkCounter = 0
			a.startSettingFeedback()
		}
	case modeSettings:
		a.settingAction(button)
	}
	a.beep()
	a.render()
}

func (a *App) settingAction(button byte) {
	if button == 'S' {
		// 分設定を確定した瞬間を「00秒」とします。
		// UP/DOWN操作中は秒を変更せず、最後のSETで秒をゼロにすることで、
		// 外部RTCなしでも時刻合わせの誤差を最小限にします。
		if a.setting == settingMinute {
			a.now = time.Date(
				a.now.Year(),
				a.now.Month(),
				a.now.Day(),
				a.now.Hour(),
				a.now.Minute(),
				0,
				0,
				a.now.Location(),
			)
		}

		a.setting++
		a.settingBlinkVisible = true
		a.settingBlinkCounter = 0
		if a.setting > settingChime {
			a.exitMode()
			return
		}
		return
	}
	if button != 'U' && button != 'D' {
		return
	}
	delta := 1
	if button == 'D' {
		delta = -1
	}
	switch a.setting {
	case settingYear:
		a.adjustDate(delta, 0)
	case settingMonth:
		a.adjustDate(delta, 1)
	case settingDay:
		a.adjustDate(delta, 2)
	case settingHour:
		a.adjustClock(delta, true)
	case settingMinute:
		a.adjustClock(delta, false)
	case settingBeep:
		a.beepEnabled = !a.beepEnabled
	case settingScroll:
		a.scrollEnabled = !a.scrollEnabled
	case settingFormat:
		a.hour12 = !a.hour12
	case settingChime:
		a.chimeEnabled = !a.chimeEnabled
	}
}

func (a *App) adjustClock(delta int, hour bool) {
	step := time.Minute
	if hour {
		step = time.Hour
	}

	t := a.now.Add(time.Duration(delta) * step)
	if !validClockDateTime(t) {
		return
	}
	a.now = t
}

func (a *App) adjustDate(delta, part int) {
	t := a.now
	switch part {
	case 0:
		// 年を変更すると2月29日が存在しなくなる場合があるため、
		// 変更後の年の月末日に日を丸めます。
		t = addYearClamped(t, delta)
	case 1:
		// 月を変更すると、31日→30日や2月などで日が範囲外になるため、
		// 変更後の月の末日に日を丸めます。
		t = addMonthClamped(t, delta)
	case 2:
		t = t.AddDate(0, 0, delta)
	default:
		return
	}

	if !validClockDateTime(t) {
		return
	}
	a.now = t
}

// isLeapYear はグレゴリオ暦のうるう年を判定します。
func isLeapYear(year int) bool {
	return year%400 == 0 || (year%4 == 0 && year%100 != 0)
}

// daysInMonth は指定した年月の日数を返します。
func daysInMonth(year int, month time.Month) int {
	switch month {
	case time.April, time.June, time.September, time.November:
		return 30
	case time.February:
		if isLeapYear(year) { // うるう年のチェックを行い、その結果から、2月の最終日を決める。
			return 29
		}
		return 28
	default:
		return 31
	}
}

// clampDay は、指定した年月に存在する範囲へ日を丸めます。
func clampDay(t time.Time) time.Time {
	day := t.Day()
	maxDay := daysInMonth(t.Year(), t.Month())
	if day > maxDay {
		day = maxDay
	}
	return time.Date(t.Year(), t.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// addYearClamped は年を変更し、2月29日などが存在しなくなる場合は月末日に丸めます。
func addYearClamped(t time.Time, delta int) time.Time {
	day := t.Day()
	year := t.Year() + delta
	if day > daysInMonth(year, t.Month()) {
		day = daysInMonth(year, t.Month())
	}
	return time.Date(year, t.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// addMonthClamped は月を変更し、変更後の月に存在しない日は月末日に丸めます。
func addMonthClamped(t time.Time, delta int) time.Time {
	year, month := t.Year(), t.Month()
	monthIndex := int(month) - 1 + delta
	year += monthIndex / 12
	monthIndex %= 12
	if monthIndex < 0 {
		monthIndex += 12
		year--
	}
	month = time.Month(monthIndex + 1)

	day := t.Day()
	maxDay := daysInMonth(year, month)
	if day > maxDay {
		day = maxDay
	}
	return time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// validClockDateTime は、ソフトウェア時計として使用できる日時か確認します。
// 月ごとの日数と、うるう年の2月29日まで厳密に判定します。
func validClockDateTime(t time.Time) bool {
	y := t.Year()
	m := t.Month()
	d := t.Day()
	return y >= 2020 && y <= 2099 &&
		m >= time.January && m <= time.December &&
		d >= 1 && d <= daysInMonth(y, m) &&
		t.Hour() >= 0 && t.Hour() <= 23 &&
		t.Minute() >= 0 && t.Minute() <= 59 &&
		t.Second() >= 0 && t.Second() <= 59
}

func (a *App) exitMode() {
	a.settingFeedbackActive = false
	a.settingBlinkVisible = true
	a.settingBlinkCounter = 0
	a.display.SetEnabled(true)
	a.mode = modeClock
	a.setting = 0
	a.scrollDue = 0
	a.scrollActive = false
	a.render()
}

// CheckTimeout は、10秒間操作がなければ設定画面を終了します。
func (a *App) CheckTimeout() {
	if a.mode != modeClock && time.Since(a.lastInput) >= 10*time.Second {
		a.exitMode()
	}
}
