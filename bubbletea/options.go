// Package bubbletea provides a Bubble Tea v2 adapter for the Cynthia effects engine.
package bubbletea

import "github.com/bealmot/cynthia/canvas"

// Config holds initialization options for the Cynthia model.
type Config struct {
	FPS          int                // target frames per second (default 60)
	RenderMode   canvas.RenderMode  // half-block, braille, or ASCII
	ColorProfile canvas.ColorProfile // truecolor, 256, 16, or none
	InitialMood  string             // starting mood preset (default "calm")
	Cols         int                // initial terminal columns
	Rows         int                // initial terminal rows
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		FPS:          60,
		RenderMode:   canvas.ModeHalfBlock,
		ColorProfile: canvas.ProfileTrueColor,
		InitialMood:  "calm",
		Cols:         80,
		Rows:         24,
	}
}
