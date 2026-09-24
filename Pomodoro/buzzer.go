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
	mode     timerMode
	endSound bool
	duration time.Duration
}

const (
	// buzzerFrequency は操作音および終了音の基本周波数です。
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
			b.playPomodoroEndSound(request.mode)
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

// PlayPomodoroEnd は現在のモードに対応したポモドーロ終了音を再生します。
func (b *Buzzer) PlayPomodoroEnd(mode timerMode) {
	b.mu.Lock()
	if b.cancelAlarm != nil {
		b.mu.Unlock()
		return
	}
	cancel := make(chan struct{})
	b.cancelAlarm = cancel
	b.mu.Unlock()

	select {
	case b.queue <- soundRequest{mode: mode, endSound: true}:
	default:
		b.mu.Lock()
		b.cancelAlarm = nil
		b.mu.Unlock()
	}
}

// StopAlarm は終了音を中断します。
// IsAlarmActive は終了音の再生中かどうかを返します。
func (b *Buzzer) IsAlarmActive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.cancelAlarm != nil
}

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

func (b *Buzzer) playPomodoroEndSound(mode timerMode) {
	b.mu.Lock()
	cancel := b.cancelAlarm
	b.mu.Unlock()
	if cancel == nil {
		return
	}

	if mode == modeWork {
		// 25MIN 終了音: 400ms ON, 800ms OFF を5回。
		for i := 0; i < 5; i++ {
			if b.waitOrCancel(cancel, 400*time.Millisecond, buzzerFrequency) {
				return
			}
			if b.waitOrCancel(cancel, 800*time.Millisecond, 0) {
				return
			}
		}
	} else {
		// 5MIN 終了音: 100ms ON/OFF、500ms ON/OFF を5回。
		for i := 0; i < 5; i++ {
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
