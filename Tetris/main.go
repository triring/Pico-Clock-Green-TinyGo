package main

import "time"

func main() {
	display := NewDisplay()
	display.Start()

	buzzer, err := NewBuzzer()
	if err != nil {
		return
	}

	app := NewApp(display, buzzer)
	buttons := NewButtons(app)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		buttons.Poll()
		app.Tick()
	}
}
