package effect

import (
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Life implements Conway's Game of Life as a procedural effect.
// Uses double-buffered bool grids with toroidal wrapping.
type Life struct {
	grid, next []bool
	w, h       int
	accum      float64
	params     map[string]float64
	rng        *rand.Rand
	seeded     bool
}

func init() {
	Register("life", func() Effect { return NewLife() })
}

func NewLife() *Life {
	return &Life{
		params: map[string]float64{
			"speed":   5.0,
			"density": 0.35,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (l *Life) Name() string { return "life" }

func (l *Life) ensureSize(w, h int) {
	if l.w == w && l.h == h {
		return
	}
	l.w, l.h = w, h
	l.grid = make([]bool, w*h)
	l.next = make([]bool, w*h)
	l.seeded = false
}

func (l *Life) seed() {
	density := l.params["density"]
	for i := range l.grid {
		l.grid[i] = l.rng.Float64() < density
	}
	l.seeded = true
}

// wrap performs toroidal coordinate wrapping.
func (l *Life) wrap(x, y int) int {
	x = ((x % l.w) + l.w) % l.w
	y = ((y % l.h) + l.h) % l.h
	return y*l.w + x
}

func (l *Life) neighbors(x, y int) int {
	n := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if l.grid[l.wrap(x+dx, y+dy)] {
				n++
			}
		}
	}
	return n
}

func (l *Life) step() {
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			n := l.neighbors(x, y)
			idx := y*l.w + x
			alive := l.grid[idx]
			// B3/S23: born with 3, survive with 2 or 3
			l.next[idx] = n == 3 || (alive && n == 2)
		}
	}
	l.grid, l.next = l.next, l.grid
}

func (l *Life) Step(frame uint64, dt float64) {
	if l.w == 0 || l.h == 0 {
		return
	}
	if !l.seeded {
		l.seed()
	}

	speed := l.params["speed"]
	l.accum += dt * speed
	for l.accum >= 1.0 {
		l.step()
		l.accum -= 1.0
	}
}

func (l *Life) Render(c *canvas.Canvas) {
	l.ensureSize(c.Width, c.Height)

	alive := canvas.Hex("#00FF88")
	dead := canvas.Hex("#0A0A0A")

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if l.grid[y*l.w+x] {
				c.SetPixel(x, y, alive)
			} else {
				c.SetPixel(x, y, dead)
			}
		}
	}
}

func (l *Life) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := l.params[k]; exists {
			l.params[k] = v
		}
	}
}

func (l *Life) Params() map[string]float64 {
	out := make(map[string]float64, len(l.params))
	for k, v := range l.params {
		out[k] = v
	}
	return out
}
