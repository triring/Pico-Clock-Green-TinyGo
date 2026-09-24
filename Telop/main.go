package main

import "time"

func main() {
	display := NewDisplay()
	display.Start()

	app := NewApp(display)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		app.Tick()
	}
}
