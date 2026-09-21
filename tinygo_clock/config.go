// Copyright (c) 2026
// SPDX-License-Identifier: BSD-3-Clause

package main

import "machine"

// Waveshare Pico-Clock-Green の GPIO 割り当てです。
const (
	pinOE   = machine.GP13
	pinSDI  = machine.GP11
	pinCLK  = machine.GP10
	pinLE   = machine.GP12
	pinA0   = machine.GP16
	pinA1   = machine.GP18
	pinA2   = machine.GP22
	pinSet  = machine.GP2
	pinUp   = machine.GP17
	pinDown = machine.GP15
	pinBuzz = machine.GP14
)

// 表示バッファは24列×8行に加えて、元ファームウェアと同じスクロール領域を確保します。
// 1行あたり8バイトを使用します。
const (
	displayColumns    = 24
	displayRows       = 8
	displayBufferSize = 112
	displayOffset     = 2
)
