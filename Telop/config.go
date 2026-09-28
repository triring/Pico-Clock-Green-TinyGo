package main

import (
	"machine"
	"time"
)

const (
	pinOE  = machine.GP13
	pinSDI = machine.GP11
	pinCLK = machine.GP10
	pinLE  = machine.GP12
	pinA0  = machine.GP16
	pinA1  = machine.GP18
	pinA2  = machine.GP22
)

const (
	// telopScrollInterval は、テロップを1ピクセル移動する間隔です。
	// telopScrollInterval = 50 * time.Millisecond
	telopScrollInterval = 100 * time.Millisecond
	// telopScrollInterval = 200 * time.Millisecond
	

	// telopLineInterval は、1行のスクロール終了から次の行を開始するまでの待ち時間です。
	telopLineInterval = 1200 * time.Millisecond

	// telopStartX は、テロップを画面右端から開始する位置です。
	telopStartX = displayColumns

	// telopTextBufferSize は、1行のテロップを保持するバッファ容量です。
	telopTextBufferSize = 1024
)

const (
	displayColumns    = 24
	displayRows       = 8
	displayBufferSize = 112
	displayOffset     = 2
)
