// Plasma demo — full-screen animated plasma at 30fps. Ctrl+C to exit.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/effect"

	"golang.org/x/term"
)

func main() {
	// Get terminal size
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		cols, rows = 80, 24
	}

	c := canvas.New(cols, rows, canvas.ModeHalfBlock)
	w := canvas.NewWriter(os.Stdout, canvas.ProfileTrueColor)
	plasma := effect.NewPlasma()

	// Handle Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// Alternate screen buffer
	fmt.Fprint(os.Stdout, "\x1b[?1049h\x1b[2J")
	defer fmt.Fprint(os.Stdout, "\x1b[?1049l")

	targetFPS := 30
	frameDur := time.Second / time.Duration(targetFPS)
	dt := 1.0 / float64(targetFPS)
	var frame uint64

	for {
		select {
		case <-sig:
			return
		default:
		}

		start := time.Now()

		plasma.Step(frame, dt)
		plasma.Render(c)
		c.Rasterize()
		w.Render(c)

		frame++

		elapsed := time.Since(start)
		if elapsed < frameDur {
			time.Sleep(frameDur - elapsed)
		}
	}
}
