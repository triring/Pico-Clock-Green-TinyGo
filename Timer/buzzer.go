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
	queue   chan soundRequest

	cancelAlarm chan struct{}
	mu          sync.Mutex
}

type soundRequest struct {
	duration time.Duration
	endSound bool
}

const (
	// buzzerFrequency は操作音および JIS S 0013 終了音の基本周波数です。
	buzzerFrequency = 1980
)

// NewBuzzer は GP14 のブザーを初期化します。
func NewBuzzer() (*Buzzer, error) {
	speaker, err := tone.New(machine.PWM7, pinBuzzer)
	if err != nil {
		return nil, err
	}
	b := &Buzzer{
		speaker: speaker,
		queue:   make(chan soundRequest, 8),
	}
	go b.run()
	return b, nil
}

func (b *Buzzer) run() {
	for request := range b.queue {
		if request.endSound {
			b.playEndSound()
			continue
		}
		b.setFrequency(buzzerFrequency)
		time.Sleep(request.duration)
		b.speaker.Stop()
	}
}

func (b *Buzzer) setFrequency(frequency uint32) {
	if frequency == 0 {
		b.speaker.Stop()
		return
	}
	period := uint64(math.Round(1e9 / float64(frequency)))
	b.speaker.SetPeriod(period)
}

// Play は指定された長さの操作音を再生します。
func (b *Buzzer) Play(duration time.Duration) {
	select {
	case b.queue <- soundRequest{duration: duration}:
	default:
	}
}

// PlayJISS0013End は JIS S 0013:2022 の終了音(近)(4)を繰り返し再生します。
// 再生中に StopAlarm が呼ばれると直ちに停止します。
func (b *Buzzer) PlayJISS0013End() {
	b.mu.Lock()
	if b.cancelAlarm != nil {
		b.mu.Unlock()
		return
	}
	cancel := make(chan struct{})
	b.cancelAlarm = cancel
	b.mu.Unlock()

	select {
	case b.queue <- soundRequest{endSound: true}:
	default:
		b.mu.Lock()
		b.cancelAlarm = nil
		b.mu.Unlock()
	}
}

// StopAlarm は JIS 終了音を中断します。
func (b *Buzzer) StopAlarm() {
	b.mu.Lock()
	cancel := b.cancelAlarm
	b.cancelAlarm = nil
	b.mu.Unlock()
	if cancel != nil {
		select {
		case cancel <- struct{}{}:
		default:
		}
	}
}

func (b *Buzzer) playEndSound() {
	b.mu.Lock()
	cancel := b.cancelAlarm
	b.mu.Unlock()
	if cancel == nil {
		return
	}

	for i := 0; i < 10; i++ {
		if b.waitOrCancel(cancel, 100*time.Millisecond, buzzerFrequency) {
			return
		}
		if b.waitOrCancel(cancel, 100*time.Millisecond, 0) {
			return
		}
		if b.waitOrCancel(cancel, 500*time.Millisecond, buzzerFrequency) {
			return
		}
		if b.waitOrCancel(cancel, 500*time.Millisecond, 0) {
			return
		}
	}

	b.speaker.Stop()
	b.mu.Lock()
	if b.cancelAlarm == cancel {
		b.cancelAlarm = nil
	}
	b.mu.Unlock()
}

func (b *Buzzer) waitOrCancel(cancel <-chan struct{}, duration time.Duration, frequency uint32) bool {
	if frequency == 0 {
		b.speaker.Stop()
	} else {
		b.setFrequency(frequency)
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-cancel:
		b.speaker.Stop()
		return true
	case <-timer.C:
		return false
	}
}
