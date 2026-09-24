// Pico-Clock-Green 用 TinyGo ポモドーロタイマーです。
// RTC は使用せず、Pico の CPU 上で経過時間だけを管理します。

package main

import "time"

func main() {
	display := NewDisplay()
	display.Start()

	buzzer, err := NewBuzzer()
	if err != nil {
		buzzer = nil
	}

	app := NewApp(display, buzzer)
	buttons := NewButtons(app)

	buttonTicker := time.NewTicker(time.Millisecond)
	defer buttonTicker.Stop()

	tickTicker := time.NewTicker(10 * time.Millisecond)
	defer tickTicker.Stop()

	for {
		select {
		case <-buttonTicker.C:
			buttons.Poll()
		case <-tickTicker.C:
			app.Tick()
		}
	}
}
