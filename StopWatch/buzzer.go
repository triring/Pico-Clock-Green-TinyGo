package main

import (
	"math"
	"time"

	"machine"
	"tinygo.org/x/drivers/tone"
)

// Buzzer は GP14 に接続されたブザーを制御します。
//
// 音源には TinyGo の tone ドライバを使用し、RP2040 の PWM7 を利用します。
type Buzzer struct {
	speaker tone.Speaker
	queue   chan time.Duration
}

const (
	// buzzerFrequency は JIS S 0013 の推奨帯域を考慮した通知音の周波数です。
	// 6オクターブのシ (B6) に相当する約1980 Hzです。
	buzzerFrequency = 1980

	// 通知音の長さです。
	startToneDuration = 50 * time.Millisecond
	stopToneDuration  = 100 * time.Millisecond
	resetToneDuration = 500 * time.Millisecond
)

// NewBuzzer は GP14 のブザーを初期化します。
func NewBuzzer() (*Buzzer, error) {
	speaker, err := tone.New(machine.PWM7, pinBuzzer)
	if err != nil {
		return nil, err
	}

	b := &Buzzer{
		speaker: speaker,
		queue:   make(chan time.Duration, 4),
	}

	go b.run()
	return b, nil
}

// run は通知音を順番に再生します。
func (b *Buzzer) run() {
	for duration := range b.queue {
		period := uint64(math.Round(1e9 / float64(buzzerFrequency)))
		b.speaker.SetPeriod(period)
		time.Sleep(duration)
		b.speaker.Stop()
	}
}

// Play は指定された長さの通知音を再生要求します。
// 呼び出し元を待たせないため、実際の再生は専用 goroutine で行います。
func (b *Buzzer) Play(duration time.Duration) {
	select {
	case b.queue <- duration:
	default:
		// ボタン操作が非常に短時間に集中した場合は、古い通知音を
		// 追加で蓄積せず、現在のキューを優先します。
	}
}

// StartSound は受付・スタート音を再生します。
func (b *Buzzer) StartSound() {
	b.Play(startToneDuration)
}

// StopSound は停止音を再生します。
func (b *Buzzer) StopSound() {
	b.Play(stopToneDuration)
}

// ResetSound はリセット音を再生します。
// JIS S 0013 にリセット音の規定はないため、終了音(近)の1.0秒を使用します。
func (b *Buzzer) ResetSound() {
	b.Play(resetToneDuration)
}
