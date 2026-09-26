# Pico-Clock-Green TinyGo  テトリス

Waveshare Pico-Clock-Green の LED マトリクスと3つのボタンを使用した Tetris 風ゲームです。

## ハードウェア

- Raspberry Pi Pico / RP2040
- LED マトリクス: Pico-Clock-Green の 8×24 ドット表示
- LEFT: GP2
- CENTER: GP17
- RIGHT: GP15
- Buzzer: GP14

## ゲーム画面

ゲーム上の論理盤面は **幅7ドット × 高さ22ドット**です。
Pico-Clock-Green の物理表示は8×24ドットなので、ゲーム盤を90度回転し、物理的な22×7ドット領域へ表示しています。

## 操作

### 待機画面

電源投入後は `TETRIS` のスクロール表示とデモプレイを繰り返します。

- LEFT: ゲーム開始
- CENTER: **長押し**で設定変更モードへ移行
- RIGHT: ゲーム開始
- 短押しの CENTER は操作受付音のみ鳴ります。

### 設定変更モード

CENTER ボタンで設定項目を確定し、次の項目へ移ります。

1. 落下スピード
   - LEFT: 高速側へ変更
   - RIGHT: 低速側へ変更
   - 表示: `FAST SPEED` / `MEDIUM SPEED` / `SLOW SPEED`
   - 初期値: `MEDIUM SPEED`
   - CENTER: 決定して効果音設定へ移行
2. 効果音
   - LEFT: `SE OFF`
   - RIGHT: `SE ON`
   - 表示: `SE OFF` / `SE ON`
   - 初期値: `SE OFF`
   - CENTER: 決定して待機画面へ戻る

設定変更中は現在の設定値をスクロール表示します。
設定値は電源を切るまで保持されます。

すべてのボタン操作には、1980Hz・20ms の受付音を使用します。
受付音は効果音設定が OFF でも鳴ります。

### ゲーム中

- LEFT: 左へ移動
- CENTER: 短押しで時計回りに回転、長押しで高速落下
- RIGHT: 右へ移動
- 効果音設定が ON の場合、ゲーム開始音・操作音・着地音・ライン消去音を鳴らします。

ブロックが接地したとき、横一列が埋まっていればその行を消去します。

## 得点

- 1ライン: 10点
- 2ライン: 20点
- 3ライン: 40点
- 4ライン: 80点

ゲームオーバー後は `GAME OVER !!` を表示し、その後 `SCORE n ` を3回表示して、再びタイトル画面へ戻ります。

## 設定

ゲームの難易度と操作性を調整するには、`config.go` で定義されている以下の変数の設定を変更して下さい。

- `speedFastInterval`: 高速設定の落下間隔
- `speedMediumInterval`: 中速設定の落下間隔
- `speedSlowInterval`: 低速設定の落下間隔
- `centerLongPressDuration`: CENTER長押し判定時間
- `fastFallInterval`: 長押し中の高速落下速度
- `buttonDebounceTime`: ボタンのデバウンス時間
- `scrollInterval`: スクロール速度
- `messageInterval`: メッセージ間隔
- `startToneDuration`: ゲーム開始音
