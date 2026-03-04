package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Moire renders overlapping circular interference patterns.
// Two or more concentric ring patterns with offset centers create
// striking moiré fringes that animate as centers drift.
type Moire struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("moire", func() Effect { return NewMoire() })
}

func NewMoire() *Moire {
	return &Moire{
		params: map[string]float64{
			"speed":    0.5,
			"scale":    20.0,
			"centers":  3,
			"contrast": 1.5,
		},
	}
}

func (m *Moire) Name() string { return "moire" }

func (m *Moire) Step(frame uint64, dt float64) {
	m.time += dt * m.params["speed"]
}

func (m *Moire) Render(c *canvas.Canvas) {
	scale := m.params["scale"]
	centers := int(m.params["centers"])
	contrast := m.params["contrast"]
	t := m.time

	if centers < 2 {
		centers = 3
	}

	pal := canvas.Palette{
		canvas.Hex("#000000"),
		canvas.Hex("#003366"),
		canvas.Hex("#0066CC"),
		canvas.Hex("#FFFFFF"),
		canvas.Hex("#0066CC"),
		canvas.Hex("#003366"),
		canvas.Hex("#000000"),
	}

	fw := float64(c.Width)
	fh := float64(c.Height)

	// Compute drifting center positions.
	type center struct{ x, y float64 }
	pts := make([]center, centers)
	for i := 0; i < centers; i++ {
		fi := float64(i)
		phase := fi * 2.0 * math.Pi / float64(centers)
		pts[i] = center{
			x: 0.5 + 0.25*math.Sin(t*0.7+phase),
			y: 0.5 + 0.25*math.Cos(t*0.5+phase*1.3),
		}
	}

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Sum sine-of-distance from each center.
			var sum float64
			for _, p := range pts {
				dx := fx - p.x
				dy := fy - p.y
				dist := math.Sqrt(dx*dx + dy*dy)
				sum += math.Sin(dist * scale * math.Pi)
			}

			// Normalize and apply contrast.
			v := sum / float64(centers)
			v = (v*contrast + 1.0) * 0.5
			v = math.Max(0, math.Min(1, v))

			c.SetPixel(x, y, pal.Sample(v))
		}
	}
}

func (m *Moire) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := m.params[k]; exists {
			m.params[k] = v
		}
	}
}

func (m *Moire) Params() map[string]float64 {
	out := make(map[string]float64, len(m.params))
	for k, v := range m.params {
		out[k] = v
	}
	return out
}
