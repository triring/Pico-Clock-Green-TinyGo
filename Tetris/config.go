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

// Tetris の入力ボタンです。
const (
	pinLeft   = machine.GP2
	pinCenter = machine.GP17
	pinRight  = machine.GP15
	pinBuzzer = machine.GP14
)

// speedSetting はゲーム中のブロック落下速度を表します。
type speedSetting uint8

const (
	speedFast speedSetting = iota
	speedMedium
	speedSlow
)

// Tetris のゲーム設定です。
const (
	boardWidth  = 7
	boardHeight = 22

	// 落下速度の設定値です。
	speedFastInterval   = 250 * time.Millisecond
	speedMediumInterval = 500 * time.Millisecond
	speedSlowInterval   = 1000 * time.Millisecond

	// defaultSpeedSetting は初期状態の落下速度です。
	defaultSpeedSetting = speedMedium

	// defaultSoundEnabled は初期状態の効果音設定です。
	defaultSoundEnabled = false

	// centerLongPressDuration は CENTER 長押しを高速落下として判定する時間です。
	centerLongPressDuration = 200 * time.Millisecond

	// fastFallInterval は CENTER 長押し中のブロック落下間隔です。
	fastFallInterval = 50 * time.Millisecond

	// demoFallInterval はデモ画面でのブロック落下間隔です。
	demoFallInterval = 300 * time.Millisecond

	// buttonDebounceTime はボタン入力を確定するまでの時間です。
	buttonDebounceTime = 10 * time.Millisecond

	// buttonAcceptFrequency は、ボタン操作の受付音に使用する周波数です。
	buttonAcceptFrequency = 1980

	// buttonAcceptDuration は、ボタン操作の受付音の長さです。
	buttonAcceptDuration = 20 * time.Millisecond

	// startToneDuration はゲーム開始時の報知音です。
	startToneDuration = 500 * time.Millisecond

	// buzzerFrequency はゲーム開始音などに使用する周波数です。
	buzzerFrequency = 1980
)

// 起動画面・ゲーム終了画面のスクロール設定です。
const (
	titleScrollText   = "TETRIS"
	gameOverText      = "GAME OVER !!"
	scoreScrollPrefix = "SCORE "
	scoreScrollRepeat = 3

	// scrollInterval はスクロールを1ドット移動する間隔です。
	scrollInterval = 100 * time.Millisecond

	// messageInterval は1つのスクロール表示が終わってから次を開始するまでの時間です。
	messageInterval = 1 * time.Second

	textBufferSize = 128

	// 設定画面に表示する文字列です。
	settingSpeedFastText   = "FAST SPEED"
	settingSpeedMediumText = "MEDIUM SPEED"
	settingSpeedSlowText   = "SLOW SPEED"
	settingSoundOnText     = "SE ON"
	settingSoundOffText    = "SE OFF"
)

// 得点設定です。
const (
	lineScore1 = 10
	lineScore2 = 20
	lineScore3 = 40
	lineScore4 = 80
)

// LED マトリクスの物理的な表示領域です。
const (
	displayColumns    = 24
	displayRows       = 8
	displayBufferSize = 112
	// displayOffset は左端1ビットを曜日表示用の予約領域として除外します。
	displayOffset = 2
	// gameDisplayRows は最下行を AM/PM 等の状態表示領域として除外したゲーム表示高さです。
	gameDisplayRows = 7
)
