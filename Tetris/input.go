package main

import (
	"machine"
	"time"
)

// Buttons は Tetris の3個のボタンをポーリングします。
// ボタンは内部プルアップ入力で、押すと Low になります。
type Buttons struct {
	left   machine.Pin
	center machine.Pin
	right  machine.Pin

	leftCount   uint16
	centerCount uint16
	rightCount  uint16

	leftHandled   bool
	centerHandled bool
	rightHandled  bool

	centerPressedAt time.Time
	centerFastFall  bool

	app *App
}

// NewButtons は3個のボタンを初期化します。
func NewButtons(app *App) *Buttons {
	pinLeft.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinCenter.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinRight.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	return &Buttons{left: pinLeft, center: pinCenter, right: pinRight, app: app}
}

// Poll は3個のボタンを1回サンプリングします。
func (b *Buttons) Poll() {
	now := time.Now()
	b.pollOne(b.left, &b.leftCount, &b.leftHandled, func() { b.app.ButtonPressed(buttonLeft) })
	b.pollCenter(now)
	b.pollOne(b.right, &b.rightCount, &b.rightHandled, func() { b.app.ButtonPressed(buttonRight) })
}

func (b *Buttons) pollCenter(now time.Time) {
	if !b.center.Get() {
		if b.centerCount < 1000 {
			b.centerCount++
		}
		if b.centerCount >= uint16(buttonDebounceTime/time.Millisecond) {
			if !b.centerHandled {
				b.centerHandled = true
				b.centerPressedAt = now
				b.centerFastFall = false
				return
			}
			if !b.centerFastFall && now.Sub(b.centerPressedAt) >= centerLongPressDuration {
				b.centerFastFall = true
				if b.app.IsWaiting() {
					b.app.CenterLongPressed()
				} else {
					b.app.SetFastFall(true)
				}
			}
		}
		return
	}

	if b.centerFastFall {
		b.app.SetFastFall(false)
	}
	if b.centerHandled && !b.centerFastFall {
		b.app.ButtonPressed(buttonCenter)
	}
	b.centerCount = 0
	b.centerHandled = false
	b.centerFastFall = false
}

// pollOne はチャタリングを除去し、1回の押下を1イベントとして通知します。
func (b *Buttons) pollOne(pin machine.Pin, count *uint16, handled *bool, action func()) {
	if !pin.Get() {
		if *count < 1000 {
			(*count)++
		}
		if *count >= uint16(buttonDebounceTime/time.Millisecond) && !*handled {
			*handled = true
			action()
		}
		return
	}
	*count = 0
	*handled = false
}
