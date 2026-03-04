package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Tunnel renders a textured fly-through tunnel effect.
// Each pixel is mapped to polar coordinates (angle, depth) using atan2 and 1/dist,
// then scrolled over time to create the illusion of forward movement.
type Tunnel struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("tunnel", func() Effect { return NewTunnel() })
}

func NewTunnel() *Tunnel {
	return &Tunnel{
		params: map[string]float64{
			"speed": 1.0,
			"scale": 4.0,
			"twist": 0.5,
		},
	}
}

func (t *Tunnel) Name() string { return "tunnel" }

func (t *Tunnel) Step(frame uint64, dt float64) {
	t.time += dt * t.params["speed"]
}

func (t *Tunnel) Render(c *canvas.Canvas) {
	scale := t.params["scale"]
	twist := t.params["twist"]
	tm := t.time

	pal := canvas.Palette{
		canvas.Hex("#0D0221"),
		canvas.Hex("#1A1A4E"),
		canvas.Hex("#3D2B7F"),
		canvas.Hex("#7B4FBF"),
		canvas.Hex("#D4A5FF"),
		canvas.Hex("#7B4FBF"),
		canvas.Hex("#3D2B7F"),
		canvas.Hex("#1A1A4E"),
	}

	fw := float64(c.Width)
	fh := float64(c.Height)
	cx := fw / 2.0
	cy := fh / 2.0

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy

			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 0.5 {
				dist = 0.5
			}

			angle := math.Atan2(dy, dx) / (2 * math.Pi)
			depth := 1.0 / dist * fw * 0.5

			// Texture coordinates with time scroll and twist.
			u := angle*scale + tm*twist
			v := depth*scale + tm

			// Checkerboard-ish pattern via two overlapping sines.
			pattern := math.Sin(u*math.Pi*2) * math.Sin(v*math.Pi*2)
			t := (pattern + 1.0) * 0.5

			// Distance fog: darken with distance from center.
			fog := 1.0 - math.Min(1.0, 8.0/dist)
			if fog < 0 {
				fog = 0
			}

			col := pal.Sample(t)
			col.R *= fog
			col.G *= fog
			col.B *= fog

			c.SetPixel(x, y, col)
		}
	}
}

func (t *Tunnel) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := t.params[k]; exists {
			t.params[k] = v
		}
	}
}

func (t *Tunnel) Params() map[string]float64 {
	out := make(map[string]float64, len(t.params))
	for k, v := range t.params {
		out[k] = v
	}
	return out
}
