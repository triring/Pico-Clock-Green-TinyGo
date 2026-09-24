// Pico-Clock-Green 用 TinyGo ストップウォッチです。
//
// RTC は使用せず、Pico の CPU 上で経過時間だけを管理します。
// 表示は MM:SS、最大 99 分 59 秒です。

package main

import (
	"time"
)

func main() {
	display := NewDisplay()
	display.Start()

	buzzer, err := NewBuzzer()
	if err != nil {
		// ブザーを初期化できない場合でも、ストップウォッチ本体は動作させます。
		buzzer = nil
	}

	app := NewApp(display, buzzer)
	app.render()

	buttons := NewButtons(app)

	// ボタン入力は約 1 ms 周期でポーリングします。
	buttonTicker := time.NewTicker(time.Millisecond)
	defer buttonTicker.Stop()

	// ストップウォッチの計測は 10 ms 周期で確認します。
	// Tick 内では実際の経過時間を time.Now() から求めるため、
	// タイマー処理の遅延が累積しないようにしています。
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
