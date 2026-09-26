package main

// Tetris の7種類のブロック番号です。
const (
	pieceI = iota
	pieceO
	pieceT
	pieceS
	pieceZ
	pieceJ
	pieceL
)

// basePiece は各ブロックの初期4×4配置です。
var basePiece = [7][4]Point{
	// I
	{{0, 1}, {1, 1}, {2, 1}, {3, 1}},
	// O
	{{1, 0}, {2, 0}, {1, 1}, {2, 1}},
	// T
	{{1, 0}, {0, 1}, {1, 1}, {2, 1}},
	// S
	{{1, 0}, {2, 0}, {0, 1}, {1, 1}},
	// Z
	{{0, 0}, {1, 0}, {1, 1}, {2, 1}},
	// J
	{{0, 0}, {0, 1}, {1, 1}, {2, 1}},
	// L
	{{2, 0}, {0, 1}, {1, 1}, {2, 1}},
}

// pieceCells は指定ブロックの回転後4セルを返します。
func pieceCells(kind, rot int) [4]Point {
	cells := basePiece[kind]
	if kind == pieceO {
		return cells
	}

	for r := 0; r < rot&3; r++ {
		for i := 0; i < len(cells); i++ {
			cells[i] = Point{x: 3 - cells[i].y, y: cells[i].x}
		}
	}

	// 回転後に左上へ詰めることで、壁際での操作を自然にします。
	minX := 4
	minY := 4
	for _, cell := range cells {
		if cell.x < minX {
			minX = cell.x
		}
		if cell.y < minY {
			minY = cell.y
		}
	}
	for i := 0; i < len(cells); i++ {
		cells[i].x -= minX
		cells[i].y -= minY
	}

	return cells
}

// 擬似乱数生成器です。ハードウェア乱数に依存せずTinyGoで安定して動作させます。
var randomState uint32 = 0x6D2B79F5

// nextRandomValue はデモ中の自動操作にも使用する擬似乱数値を返します。
func nextRandomValue() uint32 {
	x := randomState
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	randomState = x
	return x
}

func nextPieceKind() int {
	return int(nextRandomValue() % 7)
}
