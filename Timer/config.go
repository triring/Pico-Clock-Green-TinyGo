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

// カウントダウンタイマーの GPIO 割り当てです。
const (
	pinMode      = machine.GP2
	pinUp        = machine.GP17
	pinStartStop = machine.GP15
	pinBuzzer    = machine.GP14
)

// 待機中のスクロール表示設定です。
const (
	// standbyScrollText は、1分以上キー入力がないときに表示する文字列です。
	standbyScrollText = "COUNT DOWN TIMER"

	// standbyScrollDelay は、キー入力がない状態からスクロールを開始するまでの時間です。
	standbyScrollDelay = time.Minute

	// standbyScrollInterval は、スクロールを1ピクセル進める間隔です。
	standbyScrollInterval = 150 * time.Millisecond

	// standbyScrollStartX は、スクロール文字列の開始位置です。
	// 表示領域より右側から文字列を流し始めます。
	standbyScrollStartX = displayColumns

	// scrollTextBufferSize は、スクロール表示用文字列を保持するバッファの容量です。
	scrollTextBufferSize = 2048
)

// カウントダウンタイマーの設定値です。
// 実機で操作感を調整する場合は、ここだけを変更します。
const (
	// longPressDuration は長押しと判定する押下時間です。
	longPressDuration = 500 * time.Millisecond

	// upRepeatInterval は UP 長押し時の連続加算間隔です。
	upRepeatInterval = 100 * time.Millisecond

	// settingBlinkInterval は設定中の選択桁を点滅させる間隔です。
	settingBlinkInterval = 250 * time.Millisecond

	// modeStepToneDuration は MODE 短押しで設定桁を切り替える音です。
	modeStepToneDuration = 50 * time.Millisecond

	// modeExitToneDuration は最後の桁から MODE 短押しで設定を終了する音です。
	modeExitToneDuration = 500 * time.Millisecond

	// modeEnterToneDuration は設定モードへ入る音です。
	modeEnterToneDuration = 100 * time.Millisecond

	// modeLeaveToneDuration は設定モードから長押しで抜ける音です。
	modeLeaveToneDuration = 300 * time.Millisecond

	// upToneDuration は UP 短押しの音です。
	upToneDuration = 50 * time.Millisecond

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
