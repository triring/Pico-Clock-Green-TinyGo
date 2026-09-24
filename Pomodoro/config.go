// Copyright (c) 2026
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"machine"
	"time"
)

// Waveshare Pico-Clock-Green の LED 表示制御用 GPIO です。
const (
	pinOE  = machine.GP13
	pinSDI = machine.GP11
	pinCLK = machine.GP10
	pinLE  = machine.GP12
	pinA0  = machine.GP16
	pinA1  = machine.GP18
	pinA2  = machine.GP22
)

// ポモドーロタイマーの GPIO 割り当てです。
const (
	pinStartStop = machine.GP2
	pinMode      = machine.GP15
	pinBuzzer    = machine.GP14
)

// ポモドーロタイマーの基本時間です。
const (
	workDuration = 25 * time.Minute
	restDuration =  5 * time.Minute
)

// 待機中のスクロール表示設定です。
const (
	// standbyScrollText は、1分以上キー入力がないときに表示する文字列です。
	standbyScrollText = "POMODORO TIMER  POMODORO TIMER  POMODORO TIMER  "

	// standbyScrollDelay は、キー入力がない状態からスクロールを開始するまでの時間です。
	standbyScrollDelay = time.Minute

	// standbyScrollInterval は、スクロールを1ピクセル進める間隔です。
	standbyScrollInterval = 150 * time.Millisecond

	// standbyScrollStartX は、スクロール文字列の開始位置です。
	standbyScrollStartX = displayColumns

	// scrollTextBufferSize は、スクロール表示用文字列を保持するバッファの容量です。
	scrollTextBufferSize = 2048
)

// 操作音の設定値です。
const (
	// startToneDuration はカウントダウン開始・再開の音です。
	startToneDuration = 50 * time.Millisecond

	// stopToneDuration はカウントダウン停止の音です。
	stopToneDuration = 100 * time.Millisecond

	// resetToneDuration は START/STOP 長押しによるリセットの音です。
	resetToneDuration = 1000 * time.Millisecond
)

// 表示バッファは元の Pico-Clock-Green と同じサイズを使用します。
const (
	displayColumns    = 24
	displayRows       = 8
	displayBufferSize = 112
	displayOffset     = 2
)

// longPressDuration は長押しと判定する押下時間です。
const longPressDuration = 500 * time.Millisecond
