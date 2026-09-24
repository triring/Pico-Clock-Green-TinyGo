package main

import "machine"

// Buttons はスタート／ストップとリセットの2個のボタンをポーリングします。
// ボタンは内部プルアップ入力で、押すと Low になります。
type Buttons struct {
	startStop machine.Pin
	reset     machine.Pin

	startStopCount uint16
	resetCount     uint16

	startStopHandled bool
	resetHandled     bool

	app *App
}

// NewButtons は2個のボタンを初期化します。
func NewButtons(app *App) *Buttons {
	pinStartStop.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinReset.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	return &Buttons{
		startStop: pinStartStop,
		reset:     pinReset,
		app:       app,
	}
}

// Poll はボタンを1回サンプリングします。
// 約 1 ms 周期で呼び出すことを想定しています。
// 50 ms 以上の押下を1回のボタン操作として扱います。
func (b *Buttons) Poll() {
	b.pollOne(b.startStop, &b.startStopCount, &b.startStopHandled, func() {
		b.app.ToggleStartStop()
		b.app.render()
	})

	b.pollOne(b.reset, &b.resetCount, &b.resetHandled, func() {
		b.app.Reset()
	})
}

// pollOne は1個のボタンのチャタリングを含む連続入力を簡易的に処理します。
func (b *Buttons) pollOne(
	pin machine.Pin,
	count *uint16,
	handled *bool,
	action func(),
) {
	if !pin.Get() {
		if *count < 1000 {
			(*count)++
		}

		if *count >= 50 && !*handled {
			*handled = true
			action()
		}
		return
	}

	*count = 0
	*handled = false
}
