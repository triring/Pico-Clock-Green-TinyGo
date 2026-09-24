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

// ストップウォッチのボタン割り当てです。
const (
	pinStartStop = machine.GP2
	pinReset     = machine.GP15
	pinBuzzer    = machine.GP14
)

// ストップウォッチのスクロール表示設定です。
const (
	// standbyScrollText は、待機時間が長くなったときに表示する文字列です。
	standbyScrollText = "STOP WATCH  STOP WATCH  STOP WATCH  "

	// standbyScrollInterval は、スクロールを1ピクセル進める間隔です。
	standbyScrollInterval = 150 * time.Millisecond

	// standbyScrollStartX は、スクロール文字列の開始位置です。
	// 表示領域より右側から文字列を流し始めます。
	standbyScrollStartX = displayColumns

	// scrollTextBufferSize は、スクロール表示用文字列を保持するバッファの容量です。
	// ASCII文字列を最大2048バイトまで保持できます。
	scrollTextBufferSize = 2048
)

// 表示バッファは元の Pico-Clock-Green と同じサイズを使用します。
const (
	displayColumns    = 24
	displayRows       = 8
	displayBufferSize = 112
	displayOffset     = 2
)
