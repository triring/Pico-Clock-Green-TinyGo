package main

import "time"

// Point は Tetris ブロック内の1ドットの座標です。
type Point struct {
	x int
	y int
}

// Piece は現在落下中のブロックです。
type Piece struct {
	kind int
	rot  int
	x    int
	y    int
}

// ボタンの種類です。
type buttonID uint8

const (
	buttonLeft buttonID = iota
	buttonCenter
	buttonRight
)

// AppState はアプリケーションの状態です。
type AppState uint8

const (
	stateTitle AppState = iota
	stateDemo
	statePlaying
	stateGameOver
	stateSettingSpeed
	stateSettingSound
)

// App は Tetris 全体の状態を管理します。
type App struct {
	display *Display
	buzzer  *Buzzer

	state AppState

	// board[y][x] が true のとき、その位置に積み上げ済みブロックがあります。
	board [boardHeight][boardWidth]bool

	piece Piece
	score int

	lastFall time.Time
	fastFall bool

	// ゲーム設定です。
	speedSetting speedSetting
	soundEnabled bool

	// デモ画面用の状態です。
	demoLastMove time.Time

	// スクロール表示用の状態です。
	message               [textBufferSize]byte
	messageLength         int
	scrollX               int
	scrolling             bool
	nextMessageAt         time.Time
	lastScroll            time.Time
	messageScrollInterval time.Duration

	// ゲームオーバー時の表示段階です。
	gameOverStage int
	gameOverCount int
}

// NewApp は Tetris を初期化し、タイトルスクロールを開始します。
func NewApp(display *Display, buzzer *Buzzer) *App {
	a := &App{
		display:      display,
		buzzer:       buzzer,
		state:        stateTitle,
		speedSetting: defaultSpeedSetting,
		soundEnabled: defaultSoundEnabled,
	}
	a.startMessage(titleScrollText, time.Now(), scrollInterval)
	return a
}

// Tick は一定周期で呼び出され、スクロールまたはブロック落下を更新します。
func (a *App) Tick() {
	now := time.Now()

	switch a.state {
	case stateTitle:
		if !a.scrolling && !now.Before(a.nextMessageAt) {
			a.startDemo(now)
		} else {
			a.tickMessage(now)
		}
	case stateDemo:
		a.tickDemo(now)
	case stateSettingSpeed, stateSettingSound:
		a.tickSettings(now)
	case statePlaying:
		a.tickGame(now)
	case stateGameOver:
		a.tickGameOver(now)
	}
}

// ButtonPressed はボタン入力を状態に応じて処理します。
func (a *App) ButtonPressed(button buttonID) {
	// 通常のボタン操作が確定した時点で受付音を鳴らします。
	a.buttonAccepted()

	switch a.state {
	case stateTitle, stateDemo:
		// 待機中の CENTER は長押し専用なので、通常の短押しでは何もしません。
		if button == buttonCenter {
			return
		}
		a.startGame()
	case stateSettingSpeed:
		switch button {
		case buttonLeft:
			a.changeSpeed(-1)
		case buttonRight:
			a.changeSpeed(1)
		case buttonCenter:
			a.startSoundSetting(time.Now())
		}
	case stateSettingSound:
		switch button {
		case buttonLeft:
			a.soundEnabled = false
			a.showSoundSetting(time.Now())
		case buttonRight:
			a.soundEnabled = true
			a.showSoundSetting(time.Now())
		case buttonCenter:
			a.startTitle(time.Now())
		}
	case statePlaying:
		switch button {
		case buttonLeft:
			a.movePiece(-1)
		case buttonCenter:
			a.rotatePiece()
		case buttonRight:
			a.movePiece(1)
		}
	case stateGameOver:
		// ゲームオーバー表示中の入力は無視します。
	}
}

// startSettings は待機画面から設定変更モードへ移行します。
func (a *App) startSettings(now time.Time) {
	a.stopMessage()
	a.state = stateSettingSpeed
	a.showSpeedSetting(now)
}

// tickSettings は設定画面のスクロール表示を更新します。
func (a *App) tickSettings(now time.Time) {
	if a.scrolling {
		a.tickMessage(now)
		return
	}
	if !now.Before(a.nextMessageAt) {
		if a.state == stateSettingSpeed {
			a.showSpeedSetting(now)
		} else {
			a.showSoundSetting(now)
		}
	}
}

// changeSpeed は落下速度を1段階変更します。
// 左キーで高速側、右キーで低速側へ移動し、端では反対側へ循環します。
func (a *App) changeSpeed(delta int) {
	value := int(a.speedSetting) + delta
	if value < int(speedFast) {
		value = int(speedSlow)
	} else if value > int(speedSlow) {
		value = int(speedFast)
	}
	a.speedSetting = speedSetting(value)
	a.showSpeedSetting(time.Now())
}

// showSpeedSetting は現在の落下速度をスクロール表示します。
func (a *App) showSpeedSetting(now time.Time) {
	a.startMessage(a.speedText(), now, scrollInterval)
}

// startSoundSetting は落下速度を確定して効果音設定へ移行します。
func (a *App) startSoundSetting(now time.Time) {
	a.state = stateSettingSound
	a.showSoundSetting(now)
}

// showSoundSetting は現在の効果音設定をスクロール表示します。
func (a *App) showSoundSetting(now time.Time) {
	value := settingSoundOffText
	if a.soundEnabled {
		value = settingSoundOnText
	}
	a.startMessage(value, now, scrollInterval)
}

// speedText は現在の落下速度を表示用の文字列に変換します。
func (a *App) speedText() string {
	switch a.speedSetting {
	case speedFast:
		return settingSpeedFastText
	case speedSlow:
		return settingSpeedSlowText
	default:
		return settingSpeedMediumText
	}
}

// fallInterval は現在の設定に対応する通常落下間隔を返します。
func (a *App) fallInterval() time.Duration {
	switch a.speedSetting {
	case speedFast:
		return speedFastInterval
	case speedSlow:
		return speedSlowInterval
	default:
		return speedMediumInterval
	}
}

// buttonAccepted はボタン操作の受付音を鳴らします。
func (a *App) buttonAccepted() {
	if a.buzzer != nil {
		a.buzzer.sound(buttonAcceptFrequency, int(buttonAcceptDuration/time.Millisecond))
	}
}

// IsWaiting は待機画面またはデモ画面であることを返します。
func (a *App) IsWaiting() bool {
	return a.state == stateTitle || a.state == stateDemo
}

// CenterLongPressed は待機中の CENTER 長押しを設定変更開始として処理します。
func (a *App) CenterLongPressed() {
	if !a.IsWaiting() {
		return
	}
	a.buttonAccepted()
	a.startSettings(time.Now())
}

// SetFastFall は CENTER ボタン長押しによる高速落下状態を設定します。
func (a *App) SetFastFall(enabled bool) {
	if a.state != statePlaying {
		return
	}
	if enabled && !a.fastFall {
		// CENTER 長押しが高速落下として確定した時点で受付音を鳴らします。
		a.buttonAccepted()
	}
	a.fastFall = enabled
	if enabled {
		a.lastFall = time.Now()
	}
}

// startDemo はデモプレイ用の盤面を初期化してデモを開始します。
func (a *App) startDemo(now time.Time) {
	a.stopMessage()
	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			a.board[y][x] = false
		}
	}
	a.score = 0
	a.fastFall = false
	a.state = stateDemo
	a.demoLastMove = now
	if !a.spawnPiece() {
		a.startTitle(now)
		return
	}
	a.lastFall = now
	a.renderGame()
}

// tickDemo はデモプレイ中のブロックを自動操作します。
// ブロックはランダムに左右へ移動しながら、通常よりやや速く落下します。
func (a *App) tickDemo(now time.Time) {
	if now.Sub(a.lastFall) < demoFallInterval {
		return
	}
	a.lastFall = now

	// 落下するたびにランダムな方向を選び、可能なら1マス移動します。
	if now.Sub(a.demoLastMove) >= demoFallInterval {
		a.demoLastMove = now
		x := nextRandomValue() % 3
		if x == 0 {
			a.moveDemoPiece(-1)
		} else if x == 1 {
			a.moveDemoPiece(1)
		}
	}

	if a.canPlace(a.piece, a.piece.x, a.piece.y+1, a.piece.rot) {
		a.piece.y++
		a.renderGame()
		return
	}

	a.lockPiece()
	a.clearLines()
	if !a.spawnPiece() {
		a.startTitle(now)
		return
	}
	a.demoLastMove = now
	a.renderGame()
}

// moveDemoPiece はデモ中のブロックを左右へ自動移動します。
func (a *App) moveDemoPiece(dx int) {
	if a.canPlace(a.piece, a.piece.x+dx, a.piece.y, a.piece.rot) {
		a.piece.x += dx
		a.renderGame()
	}
}

// startTitle はデモ終了後にタイトルスクロールへ戻します。
func (a *App) startTitle(now time.Time) {
	a.state = stateTitle
	a.fastFall = false
	a.startMessage(titleScrollText, now, scrollInterval)
}

// startGame は盤面を初期化して新しいゲームを開始します。
func (a *App) startGame() {
	a.stopMessage()
	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			a.board[y][x] = false
		}
	}
	a.score = 0
	a.gameOverStage = 0
	a.gameOverCount = 0
	a.state = statePlaying
	a.fastFall = false
	a.spawnPiece()
	a.lastFall = time.Now()
	if a.soundEnabled && a.buzzer != nil {
		a.buzzer.Play(startToneDuration)
	}
	a.renderGame()
}

// sound は指定した周波数と音長（ミリ秒）の効果音をバックグラウンドで再生します。
func (a *App) sound(frequency uint64, durationMS int) {
	if a.soundEnabled && a.buzzer != nil {
		a.buzzer.sound(frequency, durationMS)
	}
}

// tickGame は一定時間ごとにブロックを1段落下させます。
func (a *App) tickGame(now time.Time) {
	fallInterval := a.fallInterval()
	if a.fastFall {
		fallInterval = fastFallInterval
	}
	if now.Sub(a.lastFall) < fallInterval {
		return
	}
	a.lastFall = now

	if a.canPlace(a.piece, a.piece.x, a.piece.y+1, a.piece.rot) {
		a.piece.y++
		a.renderGame()
		return
	}

	a.lockPiece()
	a.sound(440, 100)
	a.clearLines()
	if !a.spawnPiece() {
		a.beginGameOver(now)
		return
	}
	a.renderGame()
}

// spawnPiece は新しいブロックを上端に配置します。
func (a *App) spawnPiece() bool {
	kind := nextPieceKind()
	a.piece = Piece{
		kind: kind,
		rot:  0,
		x:    (boardWidth - 4) / 2,
		y:    0,
	}

	return a.canPlace(a.piece, a.piece.x, a.piece.y, a.piece.rot)
}

// movePiece は現在のブロックを左右に1マス移動します。
func (a *App) movePiece(dx int) {
	if a.canPlace(a.piece, a.piece.x+dx, a.piece.y, a.piece.rot) {
		a.piece.x += dx
		a.sound(880, 20)
		a.renderGame()
	}
}

// rotatePiece は現在のブロックを時計回りに90度回転します。
func (a *App) rotatePiece() {
	newRot := (a.piece.rot + 1) & 3
	if a.piece.kind == pieceO {
		newRot = 0
	}

	// 壁際でも回転しやすいよう、左右1マス分だけ簡易的に位置を調整します。
	for _, dx := range []int{0, -1, 1, -2, 2} {
		if a.canPlace(a.piece, a.piece.x+dx, a.piece.y, newRot) {
			a.piece.rot = newRot
			a.piece.x += dx
			a.sound(1860, 50)
			a.renderGame()
			return
		}
	}
}

// lockPiece は現在のブロックを盤面へ固定します。
func (a *App) lockPiece() {
	cells := pieceCells(a.piece.kind, a.piece.rot)
	for _, cell := range cells {
		x := a.piece.x + cell.x
		y := a.piece.y + cell.y
		if x >= 0 && x < boardWidth && y >= 0 && y < boardHeight {
			a.board[y][x] = true
		}
	}
}

// clearLines は埋まった行を消し、上の行を下へ詰めます。
func (a *App) clearLines() {
	cleared := 0

	for y := boardHeight - 1; y >= 0; y-- {
		full := true
		for x := 0; x < boardWidth; x++ {
			if !a.board[y][x] {
				full = false
				break
			}
		}
		if !full {
			continue
		}

		cleared++
		for yy := y; yy > 0; yy-- {
			a.board[yy] = a.board[yy-1]
		}
		for x := 0; x < boardWidth; x++ {
			a.board[0][x] = false
		}
		y++
	}

	if cleared > 0 {
		a.sound(220, 400)
	}

	switch cleared {
	case 1:
		a.score += lineScore1
	case 2:
		a.score += lineScore2
	case 3:
		a.score += lineScore3
	case 4:
		a.score += lineScore4
	}
}

// canPlace は指定位置にブロックを配置できるか判定します。
func (a *App) canPlace(piece Piece, x, y, rot int) bool {
	cells := pieceCells(piece.kind, rot)
	for _, cell := range cells {
		px := x + cell.x
		py := y + cell.y
		if px < 0 || px >= boardWidth || py < 0 || py >= boardHeight {
			return false
		}
		if a.board[py][px] {
			return false
		}
	}
	return true
}

// renderGame は7×22の論理ゲーム盤を、Pico-Clock-Green の24×8表示へ90度回転して描画します。
// 物理表示は24列×8行なので、ゲーム盤は22列×7行の領域を使用します。
func (a *App) renderGame() {
	a.display.Clear()

	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			if a.board[y][x] {
				a.setBoardPixel(x, y)
			}
		}
	}

	cells := pieceCells(a.piece.kind, a.piece.rot)
	if a.state == statePlaying || a.state == stateDemo {
		for _, cell := range cells {
			x := a.piece.x + cell.x
			y := a.piece.y + cell.y
			if x >= 0 && x < boardWidth && y >= 0 && y < boardHeight {
				a.setBoardPixel(x, y)
			}
		}
	}
}

// setBoardPixel は論理的な7×22座標を物理LEDマトリクスへ変換します。
func (a *App) setBoardPixel(x, y int) {
	// 論理盤面を90度回転して、物理的な22×7領域へ配置します。
	physicalX := boardHeight - 1 - y
	physicalY := x
	a.display.SetPixel(physicalX, physicalY, true)
}

// beginGameOver はゲームオーバー画面へ遷移します。
func (a *App) beginGameOver(now time.Time) {
	a.state = stateGameOver
	a.fastFall = false
	a.gameOverStage = 0
	a.gameOverCount = 0
	a.startMessage(gameOverText, now, scrollInterval)
}

// tickGameOver は GAME OVER と SCORE の表示を順番に処理します。
func (a *App) tickGameOver(now time.Time) {
	if a.scrolling {
		a.tickMessage(now)
		return
	}

	if now.Before(a.nextMessageAt) {
		return
	}

	switch a.gameOverStage {
	case 0:
		a.gameOverStage = 1
		a.startMessage(a.scoreText(), now, scrollInterval)
	case 1:
		a.gameOverCount++
		if a.gameOverCount >= scoreScrollRepeat {
			a.gameOverStage = 2
			a.nextMessageAt = now.Add(messageInterval)
		} else {
			a.startMessage(a.scoreText(), now, scrollInterval)
		}
	case 2:
		a.state = stateTitle
		a.startMessage(titleScrollText, now, scrollInterval)
	}
}

// scoreText は現在の得点を "SCORE n " 形式に変換します。
func (a *App) scoreText() string {
	var buf [32]byte
	prefix := scoreScrollPrefix
	n := copy(buf[:], prefix)

	if a.score == 0 {
		buf[n] = '0'
		n++
	} else {
		var digits [12]byte
		value := a.score
		count := 0
		for value > 0 {
			digits[count] = byte('0' + value%10)
			value /= 10
			count++
		}
		for i := count - 1; i >= 0; i-- {
			buf[n] = digits[i]
			n++
		}
	}
	buf[n] = ' '
	n++
	return string(buf[:n])
}

// startMessage は指定文字列のスクロールを開始します。
func (a *App) startMessage(text string, now time.Time, scrollInterval time.Duration) {
	a.setMessage(text)
	a.scrollX = displayColumns
	a.scrolling = true
	a.lastScroll = now
	a.messageScrollInterval = scrollInterval
	a.renderMessage()
}

// stopMessage はスクロール表示を終了します。
func (a *App) stopMessage() {
	a.scrolling = false
}

// tickMessage は文字列を1ドットずつ左へ移動します。
func (a *App) tickMessage(now time.Time) {
	if !a.scrolling {
		return
	}
	if now.Sub(a.lastScroll) < a.messageScrollInterval {
		return
	}

	a.lastScroll = now
	a.scrollX--
	a.renderMessage()

	if a.scrollX < -a.messagePixelWidth() {
		a.scrolling = false
		if a.state == stateTitle {
			a.nextMessageAt = now.Add(messageInterval)
		} else {
			a.nextMessageAt = now.Add(messageInterval)
		}
	}
}

// setMessage はスクロール用バッファへ文字列をコピーします。
func (a *App) setMessage(text string) {
	n := len(text)
	if n > len(a.message) {
		n = len(a.message)
	}
	copy(a.message[:n], text[:n])
	a.messageLength = n
}

// renderMessage は現在のスクロール位置を描画します。
func (a *App) renderMessage() {
	a.display.Clear()
	x := a.scrollX
	for i := 0; i < a.messageLength; i++ {
		if x < displayColumns && x > -5 {
			a.display.DisplayCharClipped(x, a.message[i])
		}
		x += telopCharWidth(a.message[i])
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

// messagePixelWidth は文字列全体の表示幅を返します。
func (a *App) messagePixelWidth() int {
	width := 0
	for i := 0; i < a.messageLength; i++ {
		width += telopCharWidth(a.message[i])
	}
	return width
}
