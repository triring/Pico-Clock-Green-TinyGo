package main

import "time"

// telop は、順番に表示するテロップの一覧です。
// 1行ずつ右から左へスクロール表示し、表示終了後に一定時間待ってから次の行へ進みます。
var telop = [...]string{
/*
	南極探検隊 隊員募集広告
	求む男子。至難の旅。
	僅かな報酬、極寒、暗黒の長い日々、絶えざる危険
	生還の保証無し。
	成功の暁には名誉と賞賛を得る
	アーネスト・シャクルトン
*/
	"MEN WANTED FOR HAZARDOUS JOURNEY.",
	"SMALL WAGES, BITTER COLD, LONG MONTHS OF COMPLETE DARKNESS, CONSTANT DANGER,",
	"SAFE RETURN DOUBTFUL.",
	"HONOR AND RECOGNITION IN CASE OF SUCCESS.",
	"ERNEST SHACKLETON",

/*
	"ALPHA",
	"BRAVO",
	"CHARLIE",
	"DELTA",
	"ECHO",
	"FOXTROT",
	"GOLF",
	"HOTEL",
	"INDIA",
	"JULIET",
	"KILO",
	"LIMA",
	"MIKE",
	"NOVEMBER",
	"OSCAR",
	"PAPA",
	"QUEBEC",
	"ROMEO",
	"SIERRA",
	"TANGO",
	"UNIFORM",
	"VICTOR",
	"WHISKEY",
	"X-RAY",
	"YANKEE",
	"ZULU",
*/
	//	"ABCDEFGHIJKLMNOPQRSTUVWXYZ.",
	/*
	"THE TEN COMMANDMENTS",
	"YOU SHALL HAVE NO OTHER GODS BEFORE ME.",
	"YOU SHALL NOT MAKE FOR YOURSELF AN IDOL.",
	"YOU SHALL NOT MISUSE THE NAME OF THE LORD YOUR GOD.",
	"REMEMBER THE SABBATH DAY BY KEEPING IT HOLY.",
	"HONOR YOUR FATHER AND YOUR MOTHER.",
	"YOU SHALL NOT MURDER.",
	"YOU SHALL NOT COMMIT ADULTERY.",
	"YOU SHALL NOT STEAL.",
	"YOU SHALL NOT GIVE FALSE TESTIMONY AGAINST YOUR NEIGHBOR'S HOUSE, WIFE, OR POSSESSIONS.",
*/
}

// App は、テロップ表示全体の状態を管理します。
type App struct {
	display    *Display
	lineIndex  int
	text       [telopTextBufferSize]byte
	textLength int
	scrollX    int
	scrolling  bool
	nextLineAt time.Time
	lastScroll time.Time
}

// NewApp はテロップ表示を初期化し、最初の行から表示を開始します。
func NewApp(display *Display) *App {
	a := &App{display: display}
	a.startLine(0, time.Now())
	return a
}

// Tick はテロップのスクロールと行間待ち時間を更新します。
func (a *App) Tick() {
	now := time.Now()
	if !a.scrolling {
		if now.Before(a.nextLineAt) {
			return
		}
		a.startLine((a.lineIndex+1)%len(telop), now)
		return
	}

	if now.Sub(a.lastScroll) < telopScrollInterval {
		return
	}

	a.lastScroll = now
	a.scrollX--
	a.render()

	if a.scrollX < -a.textPixelWidth() {
		a.scrolling = false
		a.nextLineAt = now.Add(telopLineInterval)
	}
}

func (a *App) startLine(index int, now time.Time) {
	a.lineIndex = index
	a.setText(telop[index])
	a.scrollX = telopStartX
	a.scrolling = true
	a.lastScroll = now
	a.render()
}

func (a *App) setText(s string) {
	n := len(s)
	if n > len(a.text) {
		n = len(a.text)
	}
	copy(a.text[:n], s[:n])
	a.textLength = n
}

func (a *App) render() {
	a.display.Clear()
	x := a.scrollX
	for i := 0; i < a.textLength; i++ {
		if x < displayColumns && x > -5 {
			a.display.DisplayCharClipped(x, a.text[i])
		}

		// MMQTVWX は 5 ドット幅いっぱいにデザインされているため、
		// 次の文字との間に 1 ドット分の間隔を設けます。
		x += telopCharWidth(a.text[i])
	}
}

// telopCharWidth は、指定した文字を含めた次の文字までの幅を返します。
// MMQTVWX の後ろには、5 ドット幅に加えて 1 ドットの間隔を設けます。
func telopCharWidth(c byte) int {
	switch c {
	case 'M', 'Q', 'T', 'V', 'W', 'X':
		return 6
	default:
		return 5
	}
}

// textPixelWidth は、現在のテロップ1行を表示するために必要な横幅を返します。
func (a *App) textPixelWidth() int {
	width := 0
	for i := 0; i < a.textLength; i++ {
		width += telopCharWidth(a.text[i])
	}
	return width
}
