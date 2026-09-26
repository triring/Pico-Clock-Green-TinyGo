# Pico-Clock-Green TinyGo

[waveshare](https://www.waveshare.com/)の[Pico-Clock-Green](https://www.waveshare.com/wiki/Pico-Clock-Green)というLEDクロック用には、C++で書かれたサンプルが公開されています。これを参考に、AIさんの力を借りながら、tinygoでいくつかのサンプルを作成しました。  

## 1. 概要

[Pico-Clock-Green](https://www.waveshare.com/wiki/Pico-Clock-Green)は、[waveshare](https://www.waveshare.com/)が製造・販売する、Raspberry Pi Pico（ラズベリーパイ・ピコ）を使用したLEDドットマトリックス電子時計キットです。  
オリジナルは、内蔵のRTCを使用して時刻を刻み、アラームやタイマーの機能を持っていました。
最初は、オリジナルのC++でかかれた時計のサンプルをtinygoで書き換えようとしました。
しかし、内蔵のRTC機能を使うと以下のような不具合が発生し、どうしても修正することが出来ず実装を諦めました。  

* RTCのバックアップ機能が正常に動かない。
* RTCへの時刻設定中に表示が0になり、それ以降、設定作業ができなくなる。

そこで、基本機能に絞って、ソフトウェアのみでいくつかのサンプルを実装してみました。  

* デジタル時計
* ポモドーロタイマー
* ストップウォッチ
* テロップ表示
* テトリス
* タイマー

| 項目 | 仕様 |
|---|---|
| 対象ハードウェア | Waveshare Pico-Clock-Green |
| CPU | Raspberry Pi Pico / RP2040 |
| 開発環境 | TinyGo |
| 時計方式 | ソフトウェア時計 |
| RTC | 使用しない |
| 電源断後の時刻保持 | 不可 |

それぞれのプログラムの機能や詳細な使い方については、それぞれのディレクトリ内のREADME.mdをお読み下さい。

* [デジタル時計](./Clock/README.md)  
* [ポモドーロタイマー](./Pomodoro/README.md)  
* [ストップウォッチ](./StopWatch/README.md)  
* [テロップ表示](./Telop/README.md)  
* [テトリス](./Tetris/README.md)  
* [カウントダウンタイマー](./Timer/README.md)  

## 2. ソースコードについて

ソースコードは、以下のディレクトリに、アプリ別に保存されています。

``` bash
> tree /a /f

+---Clock
|       app.go
|       Clock.uf2
|       config.go
|       display.go
|       font.go
|       input.go
|       main.go
|       README.md
|
+---Pomodoro
|       app.go
|       buzzer.go
|       config.go
|       display.go
|       font.go
|       input.go
|       main.go
|       Pomodoro.uf2
|       README.md
|
+---StopWatch
|       app.go
|       buzzer.go
|       config.go
|       display.go
|       font.go
|       input.go
|       main.go
|       README.md
|       StopWatch.uf2
|
+---Telop
|       app.go
|       config.go
|       display.go
|       font.go
|       main.go
|       README.md
|       Telop.uf2
|
+---Tetris
|       buzzer.go
|       config.go
|       display.go
|       font.go
|       game.go
|       input.go
|       main.go
|       pieces.go
|       README.md
|       Tetris.uf2
|
\---Timer
        app.go
        buzzer.go
        config.go
        display.go
        font.go
        input.go
        main.go
        README.md
        Timer.uf2

```
## 3. プログラムのコンパイルと転送について

### ソースコードの入手

以下のコマンドで、このサイトの内容をローカルドライブにコピーして下さい。  

```bash
> git clone https://github.com/triring/Pico-Clock-Green-TinyGo.git
```

### コンパイルと転送

まず、開発用PCとPico-Clock-GreenをUSBケーブルで接続して下さい。  
次に、Githubから入手したファイルを保存してあるディレクトリに移動して下さい。  
その中の使用したいアプリのディレクトリに移動して、以下のコマンドを実行して下さい。  
ここでは、Clockを使う例を紹介します。  
コンパイルが完了すると、生成した実行用バイナリはPico-Clock-Greenに転送されます。  

```bash
> cd tinygo_clock
> tinygo flash -target=pico -size=short .
   code    data     bss |   flash     ram
  41688     888    6116 |   42576    7004
```

また、実行用バイナリを転送できない場合は、以下のコマンドで、実行用バイナリを作成し、手作業でPico-Clock-Greenに転送して下さい。  

```bash
> cd tinygo_clock
> tinygo build -o Clock.uf2 --target pico --size short .
   code    data     bss |   flash     ram
  41688     888    6116 |   42576    7004
> dir *.uf2
-a----        2026/09/21     22:01          85504 Clock.uf2
```

### フォント構造

`font.go` のフォントは、文字を直接検索できる map 構造になっています。

概念的には次のような構造です。

```text
文字
 ↓
Glyph
 ↓
7行分のドットデータ
```
C++のサンプルでは1次元配列＋文字比較用 `switch` 方式となっていましたが、
このTinygo版では、文字から Glyph を直接取得できる構造になっています。

---

## 15. ハードウェア接続

Pico-Clock-Green の主要なGPIO割り当ては次のとおりです。  

| 機能 | RP2040 GPIO |
|---|---:|
| LED OE | GP13 |
| LED SDI | GP11 |
| LED CLK | GP10 |
| LED LE | GP12 |
| A0 | GP16 |
| A1 | GP18 |
| A2 | GP22 |
| SW1 | GP2 |
| SW2 | GP17 |
| SW3 | GP15 |
| Buzzer | GP14 |
| Light ADC | GP26 / ADC0 |

---

## Author

* @triring

## License

### 基本ライセンス  

*transmitter* is under [MIT license](https://en.wikipedia.org/wiki/MIT_License).

### 追加ライセンス

[Poul-Henning Kamp](https://people.freebsd.org/%7Ephk/) 氏が提唱しているBEER-WAREライセンスを踏襲し配布する。  

### "THE BEER-WARE LICENSE" (Revision 42)

<akio@triring.net> wrote this file. As long as you retain this notice you
can do whatever you want with this stuff. If we meet some day, and you think this stuff is worth it, you can buy me a beer in return.
Copyright (c) 2024 Akio MIWA @triring  

### "THE BEER-WARE LICENSE" (第42版)

このファイルは、<akio@triring.net> が書きました。あなたがこの条文を載せている限り、あなたはソフトウェアをどのようにでも扱うことができます。
もし、いつか私達が出会った時、あなたがこのソフトに価値があると感じたなら、見返りとして私にビールを奢ることができます。  
Copyright (c) 2024 Akio MIWA @triring  
