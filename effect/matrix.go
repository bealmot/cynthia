package effect

import (
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Matrix implements digital rain (Matrix-style falling characters).
type Matrix struct {
	drops  []matrixDrop
	params map[string]float64
	rng    *rand.Rand
	w, h   int
}

type matrixDrop struct {
	col    int
	head   float64 // y position of the head (fractional)
	speed  float64
	length int
}

func init() {
	Register("matrix", func() Effect { return NewMatrix() })
}

func NewMatrix() *Matrix {
	return &Matrix{
		params: map[string]float64{
			"speed":   1.0,
			"density": 0.4, // fraction of columns that have active drops
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (m *Matrix) Name() string { return "matrix" }

func (m *Matrix) ensureSize(w, h int) {
	if m.w == w && m.h == h {
		return
	}
	m.w, m.h = w, h
	m.drops = nil
}

func (m *Matrix) Step(frame uint64, dt float64) {
	if m.w == 0 || m.h == 0 {
		return
	}

	speed := m.params["speed"]
	density := m.params["density"]

	// Spawn new drops
	targetDrops := int(float64(m.w) * density)
	for len(m.drops) < targetDrops {
		m.drops = append(m.drops, matrixDrop{
			col:    m.rng.Intn(m.w),
			head:   -float64(m.rng.Intn(m.h)),
			speed:  (0.5 + m.rng.Float64()) * speed,
			length: 5 + m.rng.Intn(15),
		})
	}

	// Update drops
	alive := m.drops[:0]
	for _, d := range m.drops {
		d.head += d.speed * dt * 20 // pixels per second
		if int(d.head)-d.length < m.h {
			alive = append(alive, d)
		}
	}
	m.drops = alive
}

func (m *Matrix) Render(c *canvas.Canvas) {
	m.ensureSize(c.Width, c.Height)
	c.Clear(canvas.Black)

	green := canvas.Hex("#00FF41")  // matrix green
	bright := canvas.Hex("#CCFFCC") // bright head

	for _, d := range m.drops {
		headY := int(d.head)
		for i := 0; i < d.length; i++ {
			y := headY - i
			if y < 0 || y >= c.Height {
				continue
			}
			x := d.col
			if x < 0 || x >= c.Width {
				continue
			}

			// Head is brightest, fades along tail
			t := float64(i) / float64(d.length)
			var col canvas.Color
			if i == 0 {
				col = bright
			} else {
				fade := 1.0 - t
				col = canvas.Color{
					R: green.R * fade,
					G: green.G * fade,
					B: green.B * fade,
					A: green.A,
				}
			}
			c.SetPixel(x, y, col)
		}
	}
}

func (m *Matrix) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, ok := m.params[k]; ok {
			m.params[k] = v
		}
	}
}

func (m *Matrix) Params() map[string]float64 {
	out := make(map[string]float64, len(m.params))
	for k, v := range m.params {
		out[k] = v
	}
	return out
}
