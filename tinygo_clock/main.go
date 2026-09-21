// Pico-Clock-Green 用 TinyGo ファームウェアです。
//
// Raspberry Pi Pico 向け Waveshare サンプルを TinyGo へ移植しています。
// RTCを使用せず、ソフトウェア時計として動作します。
// tinygo flash -target=pico -size=short .
// tinygo build -o tinygo_clock.uf2 --target pico --size short .

package main

import (
	"machine"
	"time"
)

func main() {
	// Configure the buzzer before anything else so it cannot float during startup.
	pinBuzz.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinBuzz.Low()

	display := NewDisplay()
	display.Start()

	// RTCを使用しないため、電源投入時は既定日時から開始します。
	// 実際の現在時刻は設定モードで合わせてください。
	app := NewApp(display)
	app.render()

	buttons := NewButtons(app)

	// ボタン入力はメインgoroutineでポーリングします。
	buttonTicker := time.NewTicker(time.Millisecond)
	defer buttonTicker.Stop()

	oneSecond := time.NewTicker(time.Second)
	defer oneSecond.Stop()
	scrollTicker := time.NewTicker(47 * time.Millisecond)		// 最初の設定は、50msだったが、他のTickerとの競合を低減させるため素数にする。 
	defer scrollTicker.Stop()
	maintenanceTicker := time.NewTicker(53 * time.Millisecond)	// 最初の設定は、50msだったが、他のTickerとの競合を低減させるため素数にする。
	defer maintenanceTicker.Stop()

	for {
		select {
		case <-buttonTicker.C:
			buttons.Poll()
		case <-oneSecond.C:
			app.Tick()
		case <-scrollTicker.C:
			app.ServiceScroll()
		case <-maintenanceTicker.C:
			app.ServiceFeedback()
			app.CheckTimeout()
			app.serviceBeep()
		}
	}
}
