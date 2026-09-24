package main

import "machine"

// Buttons は、元ファームウェアの短押し50 ms／長押し300 ms判定を、
// 割り込み処理ではなく小さなポーリング状態機械として実装します。
type Buttons struct {
	set, up, down                      machine.Pin
	setCount, upCount, downCount       uint16
	setHandled, upHandled, downHandled bool
	app                                *App
}

// NewButtons は前面の3個のボタンを内部プルアップ入力として初期化します。
func NewButtons(app *App) *Buttons {
	for _, p := range []machine.Pin{pinSet, pinUp, pinDown} {
		p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
	return &Buttons{set: pinSet, up: pinUp, down: pinDown, app: app}
}

// Poll は3個のボタンを1回サンプリングします。約1 ms 周期で呼び出します。
func (b *Buttons) Poll() {
	b.pollOne(b.set,  &b.setCount,  &b.setHandled,  'S')
	b.pollOne(b.up,   &b.upCount,   &b.upHandled,   'U')
	b.pollOne(b.down, &b.downCount, &b.downHandled, 'D')
}

func (b *Buttons) pollOne(pin machine.Pin, count *uint16, handled *bool, code byte) {
	if !pin.Get() {
		if *count < 1000 {
			(*count)++
		}
		if *count >= 300 && !*handled {
			*handled = true
			b.app.ButtonAction(code, true)
		}
		return
	}
	if *count >= 50 && *count < 300 && !*handled {
		b.app.ButtonAction(code, false)
	}
	*count = 0
	*handled = false
}
