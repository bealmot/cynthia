package effect

import (
	"github.com/bealmot/cynthia/canvas"
)

// Langton implements Langton's Ant — a 2-state 2D cellular automaton
// that produces emergent complexity from just 2 rules:
//   - On white: turn right 90°, flip color, move forward
//   - On black: turn left 90°, flip color, move forward
//
// After ~10,000 steps of apparent chaos, a "highway" pattern emerges.
type Langton struct {
	grid   []bool
	w, h   int
	antX   int
	antY   int
	antDir int // 0=up, 1=right, 2=down, 3=left
	accum  float64
	params map[string]float64
}

func init() {
	Register("langton", func() Effect { return NewLangton() })
}

func NewLangton() *Langton {
	return &Langton{
		params: map[string]float64{
			"speed": 50.0,
		},
	}
}

func (l *Langton) Name() string { return "langton" }

func (l *Langton) ensureSize(w, h int) {
	if l.w == w && l.h == h {
		return
	}
	l.w, l.h = w, h
	l.grid = make([]bool, w*h)
	l.antX = w / 2
	l.antY = h / 2
	l.antDir = 0
}

func (l *Langton) step() {
	idx := l.antY*l.w + l.antX

	if l.grid[idx] {
		// On black: turn left.
		l.antDir = (l.antDir + 3) % 4
	} else {
		// On white: turn right.
		l.antDir = (l.antDir + 1) % 4
	}

	// Flip cell color.
	l.grid[idx] = !l.grid[idx]

	// Move forward (toroidal).
	switch l.antDir {
	case 0:
		l.antY = (l.antY - 1 + l.h) % l.h
	case 1:
		l.antX = (l.antX + 1) % l.w
	case 2:
		l.antY = (l.antY + 1) % l.h
	case 3:
		l.antX = (l.antX - 1 + l.w) % l.w
	}
}

func (l *Langton) Step(frame uint64, dt float64) {
	if l.w == 0 || l.h == 0 {
		return
	}

	speed := l.params["speed"]
	l.accum += dt * speed
	for l.accum >= 1.0 {
		l.step()
		l.accum -= 1.0
	}
}

func (l *Langton) Render(c *canvas.Canvas) {
	l.ensureSize(c.Width, c.Height)

	on := canvas.Hex("#FF8800")
	off := canvas.Hex("#0A0A12")
	ant := canvas.Hex("#FF0044")

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if x == l.antX && y == l.antY {
				c.SetPixel(x, y, ant)
			} else if l.grid[y*l.w+x] {
				c.SetPixel(x, y, on)
			} else {
				c.SetPixel(x, y, off)
			}
		}
	}
}

func (l *Langton) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := l.params[k]; exists {
			l.params[k] = v
		}
	}
}

func (l *Langton) Params() map[string]float64 {
	out := make(map[string]float64, len(l.params))
	for k, v := range l.params {
		out[k] = v
	}
	return out
}
