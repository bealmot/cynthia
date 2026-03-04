package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Metaballs renders blobby organic shapes using inverse-distance field summation.
// Multiple moving points generate fields that merge smoothly when close together.
type Metaballs struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("metaballs", func() Effect { return NewMetaballs() })
}

func NewMetaballs() *Metaballs {
	return &Metaballs{
		params: map[string]float64{
			"speed":     0.7,
			"count":     6,
			"threshold": 1.0,
			"radius":    0.15,
		},
	}
}

func (m *Metaballs) Name() string { return "metaballs" }

func (m *Metaballs) Step(frame uint64, dt float64) {
	m.time += dt * m.params["speed"]
}

func (m *Metaballs) Render(c *canvas.Canvas) {
	count := int(m.params["count"])
	threshold := m.params["threshold"]
	radius := m.params["radius"]
	t := m.time

	if count < 1 {
		count = 6
	}

	// Compute ball positions — each ball orbits on its own trajectory.
	type ball struct{ x, y float64 }
	balls := make([]ball, count)
	for i := 0; i < count; i++ {
		fi := float64(i)
		phase := fi * 2.0 * math.Pi / float64(count)
		balls[i] = ball{
			x: 0.5 + 0.3*math.Sin(t*0.7+phase)*math.Cos(t*0.3+fi),
			y: 0.5 + 0.3*math.Cos(t*0.5+phase)*math.Sin(t*0.4+fi*1.3),
		}
	}

	pal := canvas.Palette{
		canvas.Hex("#0A0020"),
		canvas.Hex("#1B0A4A"),
		canvas.Hex("#FF2D95"),
		canvas.Hex("#FF6B6B"),
		canvas.Hex("#FFBE76"),
		canvas.Hex("#FFFFFF"),
	}

	fw := float64(c.Width)
	fh := float64(c.Height)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Sum inverse-distance fields from all balls.
			var field float64
			for _, b := range balls {
				dx := fx - b.x
				dy := fy - b.y
				dist := dx*dx + dy*dy
				if dist < 1e-6 {
					dist = 1e-6
				}
				field += radius * radius / dist
			}

			// Map field value to color.
			v := field / threshold
			if v > 1.0 {
				v = 1.0
			}
			c.SetPixel(x, y, pal.Sample(v))
		}
	}
}

func (m *Metaballs) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := m.params[k]; exists {
			m.params[k] = v
		}
	}
}

func (m *Metaballs) Params() map[string]float64 {
	out := make(map[string]float64, len(m.params))
	for k, v := range m.params {
		out[k] = v
	}
	return out
}
