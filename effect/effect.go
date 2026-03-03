// Package effect defines the Effect interface and built-in procedural effects.
package effect

import "github.com/bealmot/cynthia/canvas"

// Effect is a procedural visual effect that renders to a canvas.
type Effect interface {
	// Step advances the effect by one frame.
	Step(frame uint64, dt float64)

	// Render draws the effect to the canvas pixel buffer.
	Render(c *canvas.Canvas)

	// SetParams updates effect parameters (typically spring-interpolated by the director).
	SetParams(params map[string]float64)

	// Params returns the current parameter values.
	Params() map[string]float64

	// Name returns the effect's identifier.
	Name() string
}
