package main

import (
	"sync"
	"time"

	"machine"
)

// Display は、Waveshare の元ファームウェアと同じ4線式シフトインターフェースで
// 24×8 LED マトリクスを直接駆動します。
type Display struct {
	buf     [displayBufferSize]byte
	mu      sync.Mutex
	row     byte
	enabled bool
}

// NewDisplay は、LED マトリクス制御用 GPIO を初期化して Display を生成します。
func NewDisplay() *Display {
	for _, p := range []machine.Pin{pinOE, pinSDI, pinCLK, pinLE, pinA0, pinA1, pinA2} {
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	pinOE.High() // output disabled while initializing
	pinSDI.Low()
	pinCLK.Low()
	pinLE.Low()
	pinA0.Low()
	pinA1.Low()
	pinA2.Low()
	return &Display{enabled: true}
}

// Start は、1 ms 周期の LED マトリクス多重化リフレッシュ処理を開始します。
func (d *Display) Start() {
	go func() {
		t := time.NewTicker(time.Millisecond)
		defer t.Stop()
		for range t.C {
			d.refresh()
		}
	}()
}

func (d *Display) refresh() {
	d.mu.Lock()
	row := d.row
	var data [4]byte
	for i := 0; i < 4; i++ {
		data[i] = d.buf[8*i+int(row)]
	}
	d.row = (d.row + 1) & 7
	enabled := d.enabled
	d.mu.Unlock()

	// Disable the LED outputs while shifting data and selecting the row.
	if !enabled {
		pinOE.High()
		return
	}
	pinOE.High()
	for i := 0; i < 4; i++ {
		sendByteLSB(data[i])
	}
	pinLE.High()
	pinLE.Low()
	pinA0.Set((row & 1) != 0)
	pinA1.Set((row & 2) != 0)
	pinA2.Set((row & 4) != 0)

	// Fixed brightness: keep the LED output enabled for every scan.
	pinOE.Low()
}

func sendByteLSB(v byte) {
	for i := 0; i < 8; i++ {
		pinCLK.Low()
		pinSDI.Set((v & 1) != 0)
		v >>= 1
		pinCLK.High()
	}
}

// SetPixel はTetrisの論理表示領域の指定した1ドットを点灯または消灯します。
// x は0～21、y は0～6の範囲で使用します。
// 物理表示では、文字列表示メソッドと同じく左端1ビットを除外し、
// 物理行1～7へ描画します。
func (d *Display) SetPixel(x, y int, on bool) {
	if x < 0 || x >= displayColumns-displayOffset || y < 0 || y >= gameDisplayRows {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	physicalX := x + displayOffset
	byteIndex := physicalX / 8
	bitIndex := physicalX % 8
	// 文字列表示（DisplayCharClipped）と同じく、先頭の1行を予約して
	// 実際のマトリックス表示領域は物理行1～7を使用します。
	physicalY := y + 1
	index := 8*byteIndex + physicalY
	mask := byte(1 << bitIndex)
	if on {
		d.buf[index] |= mask
	} else {
		d.buf[index] &^= mask
	}
}

// Clear は、表示バッファ全体を消去します。
func (d *Display) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.buf {
		d.buf[i] = 0
	}
}

// SetEnabled は LED マトリクスの発光を有効または無効にします。
// 表示バッファの内容は保持したまま、出力だけを停止します。
// 設定モードへの移行時など、一時的な画面点滅に使用します。
func (d *Display) SetEnabled(enabled bool) {
	d.mu.Lock()
	d.enabled = enabled
	d.mu.Unlock()
	if !enabled {
		pinOE.High()
	}
}

// DisplayChar は、元の C ファームウェアと同じ論理列位置に1文字を書き込みます。
// 時計表示用の予約領域を考慮して描画します。
func (d *Display) DisplayChar(x byte, c byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	x += displayOffset
	j := int(x / 8)
	k := int(x % 8)
	glyph := glyphFor(c)
	if j < 0 || j >= displayBufferSize/8 {
		return
	}
	for i := 0; i < gameDisplayRows; i++ {
		row := glyph[i-1]
		if k > 0 {
			d.buf[8*j+i] = (d.buf[8*j+i] & (0xff >> (8 - k))) | (row << k)
			if j < displayBufferSize/8-1 {
				d.buf[8*j+8+i] = (d.buf[8*j+8+i] & (0xff << (8 - k))) | (row >> (8 - k))
			}
		} else {
			d.buf[8*j+i] = row
		}
	}
}

// DisplayCharClipped は、指定した論理列位置に文字を描画します。
// 表示領域の左右からはみ出した部分は自動的に切り捨てます。
// スクロール表示で、文字が画面端を通過するときに使用します。
func (d *Display) DisplayCharClipped(x int, c byte) {
	glyph := glyphFor(c)

	d.mu.Lock()
	defer d.mu.Unlock()

	for row := 0; row < gameDisplayRows; row++ {
		bits := glyph[row]
		for bit := 0; bit < 5; bit++ {
			if bits&(1<<bit) == 0 {
				continue
			}

			logicalX := x + bit
			if logicalX < 0 || logicalX >= displayColumns {
				continue
			}

			physicalX := logicalX + displayOffset
			byteIndex := physicalX / 8
			bitIndex := physicalX % 8
			d.buf[8*byteIndex+row+1] |= 1 << bitIndex
		}
	}
}

// ClearFrom は、指定列から8列おきに表示位置を空白にします。
func (d *Display) ClearFrom(x byte) {
	for x < displayBufferSize {
		d.DisplayChar(x, ' ')
		x += 8
	}
}

// SetWeekdayIndicator は、指定された曜日の LED インジケータを点灯します。
func (d *Display) SetWeekdayIndicator(weekday int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Preserve the first two status bits and clear all weekday bits first.
	for _, i := range []int{0, 8, 16} {
		mask := byte(0)
		if i == 0 {
			mask = 0x03
		}
		d.buf[i] &= mask
	}
	for i := 0; i < 7; i++ {
		d.clearWeekdayBit(i)
	}
	if weekday < 1 || weekday > 7 {
		return
	}
	switch weekday {
	case 1:
		d.buf[0] |= 1<<3 | 1<<4
	case 2:
		d.buf[0] |= 1<<6 | 1<<7
	case 3:
		d.buf[8] |= 1<<1 | 1<<2
	case 4:
		d.buf[8] |= 1<<4 | 1<<5
	case 5:
		d.buf[8] |= 1 << 7
		d.buf[16] |= 1 << 0
	case 6:
		d.buf[16] |= 1<<2 | 1<<3
	case 7:
		d.buf[16] |= 1<<5 | 1<<6
	}
}

func (d *Display) clearWeekdayBit(i int) {
	switch i {
	case 0:
		d.buf[0] &= ^byte((1 << 3) | (1 << 4))
	case 1:
		d.buf[0] &= ^byte((1 << 6) | (1 << 7))
	case 2:
		d.buf[8] &= ^byte((1 << 1) | (1 << 2))
	case 3:
		d.buf[8] &= ^byte((1 << 4) | (1 << 5))
	case 4:
		d.buf[8] &= ^byte(1 << 7)
		d.buf[16] &= ^byte(1 << 0)
	case 5:
		d.buf[16] &= ^byte((1 << 2) | (1 << 3))
	case 6:
		d.buf[16] &= ^byte((1 << 5) | (1 << 6))
	}
}

// SetStatus は、元ファームウェアにある状態表示 LED を点灯または消灯します。
// index は 0=スクロール、4=AM、5=PM、7=時報を表します。
// その他のインデックスは表示互換性のため残していますが、本アプリでは使用しません。
func (d *Display) SetStatus(index int, on bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var p *byte
	var mask byte
	switch index {
	case 0:
		p = &d.buf[0]
		mask = 0x03
	case 1:
		p = &d.buf[1]
		mask = 0x03
	case 2:
		p = &d.buf[2]
		mask = 0x03
	case 3:
		p = &d.buf[3]
		mask = 0x01
	case 4:
		p = &d.buf[4]
		mask = 0x01
	case 5:
		p = &d.buf[4]
		mask = 0x02
	case 6:
		p = &d.buf[5]
		mask = 0x03
	case 7:
		p = &d.buf[6]
		mask = 0x03
	case 8:
		p = &d.buf[7]
		mask = 0x03
	default:
		return
	}
	if on {
		*p |= mask
	} else {
		*p &^= mask
	}
}

// ScrollLeft は、曜日・機能表示用の下位2ビットを保持したまま、表示内容を1ピクセル左へ移動します。
func (d *Display) ScrollLeft() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := 0; i < gameDisplayRows; i++ {
		status := d.buf[i] & 0x03
		for j := 0; j < displayBufferSize/8-1; j++ {
			d.buf[8*j+i] = d.buf[8*j+i]>>1 | d.buf[8*j+8+i]<<7
		}
		d.buf[8*(displayBufferSize/8-1)+i] >>= 1
		d.buf[i] = (d.buf[i] &^ 0x03) | status
	}
}
