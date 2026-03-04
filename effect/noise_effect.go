package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// NoiseEffect renders animated simplex noise with palette mapping.
// Uses fbm (fractional Brownian motion) for layered organic patterns.
type NoiseEffect struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("noise", func() Effect { return NewNoiseEffect() })
}

func NewNoiseEffect() *NoiseEffect {
	return &NoiseEffect{
		params: map[string]float64{
			"speed":       0.5,
			"scale":       3.0,
			"octaves":     4,
			"lacunarity":  2.0,
			"persistence": 0.5,
		},
	}
}

func (n *NoiseEffect) Name() string { return "noise" }

func (n *NoiseEffect) Step(frame uint64, dt float64) {
	n.time += dt * n.params["speed"]
}

func (n *NoiseEffect) Render(c *canvas.Canvas) {
	scale := n.params["scale"]
	octaves := int(n.params["octaves"])
	lacunarity := n.params["lacunarity"]
	persistence := n.params["persistence"]

	pal := canvas.Palette{
		canvas.Hex("#0B0B2A"),
		canvas.Hex("#1B3A4B"),
		canvas.Hex("#2D6A4F"),
		canvas.Hex("#52B788"),
		canvas.Hex("#95D5B2"),
		canvas.Hex("#D8F3DC"),
		canvas.Hex("#52B788"),
		canvas.Hex("#1B3A4B"),
	}

	fw := float64(c.Width)
	fh := float64(c.Height)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh * scale
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw * scale

			v := fbm(fx, fy, n.time, octaves, lacunarity, persistence)
			// Normalize from [-1,1] to [0,1]
			v = (v + 1.0) * 0.5
			v = math.Mod(v, 1.0)
			if v < 0 {
				v += 1.0
			}

			c.SetPixel(x, y, pal.Sample(v))
		}
	}
}

func (n *NoiseEffect) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := n.params[k]; exists {
			n.params[k] = v
		}
	}
}

func (n *NoiseEffect) Params() map[string]float64 {
	out := make(map[string]float64, len(n.params))
	for k, v := range n.params {
		out[k] = v
	}
	return out
}
