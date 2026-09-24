# Pico-Clock-Green TinyGo ポモドーロタイマー

Waveshare Pico-Clock-Green を Raspberry Pi Pico / TinyGo で動作させるポモドーロタイマーです。

## 仕様

- 25分の作業時間と5分の休憩時間を切り替えて使用します。
- 起動時は25MINモードです。
- 待機中は `25MIN` または `5MIN` を表示します。
- START/STOP 短押しでカウントダウン開始・停止・再開を行います。
- START/STOP 長押しで現在モードの設定時間へリセットします。
- MODE 短押しで25MIN / 5MINを切り替えます。
- カウントダウン中のMODE操作は無視します。
- 終了時はモードごとの終了音を10回再生し、終了音完了後に次のモードへ自動切替します。
- 終了音中にMODEまたはSTART/STOPを押すと、終了音だけを中断します。押したボタンの通常操作は実行しません。
- 待機中に1分間入力がない場合は `POMODORO TIMER` をスクロール表示します。
- スクロール中にいずれかのボタンを押すとスクロールを中断して待機表示へ戻ります。

## GPIO

| 機能 | GPIO |
|---|---:|
| START/STOP | GP2 |
| MODE | GP15 |
| ブザー | GP14 |

LEDマトリクスのGPIOは Waveshare Pico-Clock-Green の既存配線を使用します。

## 操作

| 操作 | 動作 |
|---|---|
| START/STOP 短押し | 開始 / 停止 / 再開 |
| START/STOP 長押し | 現在モードの時間へリセット |
| MODE 短押し | 25MIN ↔ 5MIN |
| MODE 長押し | 無効 |

## 終了音

### 25MIN

400ms ON → 800ms OFF を5回。

### 5MIN

100ms ON → 100ms OFF → 500ms ON → 500ms OFF を5回。

基本周波数は `buzzerFrequency = 1980` Hz です。

## ファイル構成

| ファイル | 機能 |
|---|---|
| `main.go` | メインループとタイマー周期処理 |
| `config.go` | GPIO、時間、スクロール等の設定値 |
| `app.go` | ポモドーロタイマー本体、状態遷移、表示、スクロール |
| `input.go` | START/STOP、MODEのボタン入力処理 |
| `buzzer.go` | 操作音とポモドーロ終了音 |
| `display.go` | LEDマトリクス表示制御 |
| `font.go` | LEDマトリクス用フォント |

## 動作状態

```text
                  MODE短押し
        +----------------------------+
        |                            |
        v                            |
     25MIN待機 <-----------------> 5MIN待機
        |                            |
        | START短押し               | START短押し
        v                            v
     カウントダウン              カウントダウン
        |                            |
        | START短押し               | START短押し
        v                            v
       PAUSE                        PAUSE
        |                            |
        +------ START短押し --------+

25MIN終了 → 終了音 → 5MIN待機
5MIN終了  → 終了音 → 25MIN待機
```

終了音中に MODE または START/STOP を押すと、終了音を中断して現在のモードの待機表示へ戻ります。
