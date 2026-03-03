package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Gradient renders animated gradient sweeps and color cycling.
type Gradient struct {
	time   float64
	params map[string]float64
	pal    canvas.Palette
}

func init() {
	Register("gradient", func() Effect { return NewGradient() })
}

func NewGradient() *Gradient {
	return &Gradient{
		params: map[string]float64{
			"speed": 0.3,
			"angle": 0.0, // 0 = horizontal, 0.5 = diagonal, 1 = vertical
			"scale": 1.0,
		},
		pal: canvas.Palette{
			canvas.Hex("#12101A"), // deep violet
			canvas.Hex("#1A1726"),
			canvas.Hex("#221F30"),
			canvas.Hex("#2A2540"),
			canvas.Hex("#1A1726"),
			canvas.Hex("#12101A"),
		},
	}
}

func (g *Gradient) Name() string { return "gradient" }

func (g *Gradient) Step(frame uint64, dt float64) {
	g.time += dt * g.params["speed"]
}

func (g *Gradient) Render(c *canvas.Canvas) {
	angle := g.params["angle"] * math.Pi
	scale := g.params["scale"]

	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	fw := float64(c.Width)
	fh := float64(c.Height)

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw
			fy := float64(y) / fh

			// Project position along gradient angle
			t := fx*cosA + fy*sinA
			t = t*scale + g.time
			t = math.Mod(t, 1.0)
			if t < 0 {
				t += 1.0
			}

			c.SetPixel(x, y, g.pal.Sample(t))
		}
	}
}

func (g *Gradient) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, ok := g.params[k]; ok {
			g.params[k] = v
		}
	}
}

func (g *Gradient) Params() map[string]float64 {
	out := make(map[string]float64, len(g.params))
	for k, v := range g.params {
		out[k] = v
	}
	return out
}
