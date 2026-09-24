package main

import (
	"machine"
	"time"
)

// button は1個のボタンの押下状態を管理します。
type button struct {
	pin machine.Pin

	pressedAt  uint32
	lastRepeat uint32
	pressed    bool
	long       bool
}

// Buttons は MODE、UP、START/STOP の3個のボタンをポーリングします。
type Buttons struct {
	mode      button
	up        button
	startStop button

	app *App
}

// NewButtons は3個のボタンを初期化します。
func NewButtons(app *App) *Buttons {
	pinMode.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinUp.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinStartStop.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	return &Buttons{
		mode:      button{pin: pinMode},
		up:        button{pin: pinUp},
		startStop: button{pin: pinStartStop},
		app:       app,
	}
}

// Poll はボタンを1回サンプリングします。
// 約1ms周期で呼び出すことを想定しています。
func (b *Buttons) Poll() {
	now := millis()

	// スクロール表示中は、いずれかのスイッチを押した時点で
	// スクロールだけを中断し、そのスイッチの通常操作は消費します。
	if b.app.IsScrolling() && (!b.mode.pin.Get() || !b.up.pin.Get() || !b.startStop.pin.Get()) {
		b.app.HandleScrollButton()
		return
	}

	b.pollMode(now)
	b.pollUp(now)
	b.pollStartStop(now)
}

func (b *Buttons) pollMode(now uint32) {
	p := &b.mode
	if !p.pin.Get() {
		if !p.pressed {
			p.pressed = true
			p.long = false
			p.pressedAt = now
			return
		}
		if b.app.state == timerAlarm {
			b.app.HandleAlarmButton()
			p.long = true
			return
		}
		if !p.long && elapsedMillis(now, p.pressedAt) >= uint32(longPressDuration/time.Millisecond) {
			p.long = true
			b.app.ModeLongPress()
		}
		return
	}

	if !p.pressed {
		return
	}
	if !p.long {
		if b.app.state == timerAlarm {
			b.app.HandleAlarmButton()
		} else {
			b.app.ModeShortPress()
		}
	}
	p.pressed = false
	p.long = false
}

func (b *Buttons) pollUp(now uint32) {
	p := &b.up
	if !p.pin.Get() {
		if !p.pressed {
			p.pressed = true
			p.long = false
			p.pressedAt = now
			p.lastRepeat = now
			return
		}
		if b.app.state == timerAlarm {
			b.app.HandleAlarmButton()
			p.long = true
			return
		}
		if !p.long && elapsedMillis(now, p.pressedAt) >= uint32(longPressDuration/time.Millisecond) {
			p.long = true
			p.lastRepeat = now
			b.app.UpRepeat()
			return
		}
		if p.long && elapsedMillis(now, p.lastRepeat) >= uint32(upRepeatInterval/time.Millisecond) {
			p.lastRepeat = now
			b.app.UpRepeat()
		}
		return
	}

	if !p.pressed {
		return
	}
	if !p.long {
		if b.app.state == timerAlarm {
			b.app.HandleAlarmButton()
		} else {
			b.app.UpShortPress()
		}
	}
	p.pressed = false
	p.long = false
}

func (b *Buttons) pollStartStop(now uint32) {
	p := &b.startStop
	if !p.pin.Get() {
		if !p.pressed {
			p.pressed = true
			p.long = false
			p.pressedAt = now
			return
		}
		if b.app.state == timerAlarm {
			b.app.HandleAlarmButton()
			p.long = true
			return
		}
		if !p.long && elapsedMillis(now, p.pressedAt) >= uint32(longPressDuration/time.Millisecond) {
			p.long = true
			b.app.StartStopLongPress()
		}
		return
	}

	if !p.pressed {
		return
	}
	if !p.long {
		if b.app.state == timerAlarm {
			b.app.HandleAlarmButton()
		} else {
			b.app.StartStopShortPress()
		}
	}
	p.pressed = false
	p.long = false
}

func millis() uint32 {
	return uint32(time.Now().UnixNano() / int64(time.Millisecond))
}

func elapsedMillis(now, then uint32) uint32 {
	return now - then
}
