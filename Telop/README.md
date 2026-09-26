# Pico-Clock-Green TinyGo テロップ

Waveshare Pico-Clock-Green を使ったテロップ表示アプリです。

## 動作

登録された文字列を1行ずつ右から左へスクロール表示します。
1行のスクロールが終わると設定した時間だけ待ち、次の行を表示します。
最後の行まで表示すると、先頭の行へ戻って繰り返します。

## 設定

`config.go` の以下を実機で調整できます。

- `telopScrollInterval`: スクロール速度
- `telopLineInterval`: 1行終了後、次の行を表示するまでの待ち時間

テロップ本文は `app.go` の `telop` 配列を編集します。
以下がそのサンプルです。  

```go
/*
	南極探検隊 隊員募集広告
	求む男子。至難の旅。
	僅かな報酬、極寒、暗黒の長い日々、絶えざる危険
	生還の保証無し。
	成功の暁には名誉と賞賛を得る
	アーネスト・シャクルトン
*/
var telop = [...]string{
	"MEN WANTED FOR HAZARDOUS JOURNEY.",
	"SMALL WAGES, BITTER COLD, LONG MONTHS OF COMPLETE DARKNESS, CONSTANT DANGER,",
	"SAFE RETURN DOUBTFUL.",
	"HONOR AND RECOGNITION IN CASE OF SUCCESS.",
	"ERNEST SHACKLETON",
}
```