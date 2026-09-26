package main

import (
	"math"
	"sync"
	"time"

	"machine"
	"tinygo.org/x/drivers/tone"
)

// Buzzer は GP14 に接続されたブザーを制御します。
type Buzzer struct {
	speaker tone.Speaker
	mu      sync.Mutex
}

// NewBuzzer は GP14 のブザーを初期化します。
func NewBuzzer() (*Buzzer, error) {
	speaker, err := tone.New(machine.PWM7, pinBuzzer)
	if err != nil {
		return nil, err
	}
	return &Buzzer{speaker: speaker}, nil
}

// Play は指定した長さだけ 1980Hz の音を鳴らします。
// ゲーム開始時など、呼び出し側で待ってもよい報知音に使用します。
func (b *Buzzer) Play(duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.setFrequency(buzzerFrequency)
	time.Sleep(duration)
	b.speaker.Stop()
}

// sound は指定した周波数と音長（ミリ秒）の音をバックグラウンドで鳴らします。
//
// ゲーム本体の処理を停止させないため、音の再生処理は goroutine で実行します。
// 複数の音が短時間に要求された場合は、ブザーを排他制御して順番に再生します。
func (b *Buzzer) sound(frequency uint64, durationMS int) {
	go func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		b.setFrequency(frequency)
		time.Sleep(time.Duration(durationMS) * time.Millisecond)
		b.speaker.Stop()
	}()
}

// setFrequency は指定周波数に対応するPWM周期を設定します。
func (b *Buzzer) setFrequency(frequency uint64) {
	if frequency == 0 {
		b.speaker.Stop()
		return
	}
	period := uint64(math.Round(1e9 / float64(frequency)))
	b.speaker.SetPeriod(period)
}
