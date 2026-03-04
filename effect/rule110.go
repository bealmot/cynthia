package effect

import (
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Rule110 renders a 1D elementary cellular automaton as a scrolling 2D image.
// Each row is one generation, newest at the bottom.
// Rule 110 is proven Turing-complete — complex behavior from simple rules.
type Rule110 struct {
	grid   []bool
	w, h   int
	row    int // current write row (circular buffer)
	accum  float64
	params map[string]float64
	rng    *rand.Rand
	seeded bool
}

func init() {
	Register("rule110", func() Effect { return NewRule110() })
}

func NewRule110() *Rule110 {
	return &Rule110{
		params: map[string]float64{
			"speed": 10.0,
			"rule":  110,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (r *Rule110) Name() string { return "rule110" }

func (r *Rule110) ensureSize(w, h int) {
	if r.w == w && r.h == h {
		return
	}
	r.w, r.h = w, h
	r.grid = make([]bool, w*h)
	r.row = 0
	r.seeded = false
}

func (r *Rule110) seed() {
	// Start with a single cell on the right edge (classic Rule 110 initial condition).
	firstRow := r.row * r.w
	r.grid[firstRow+r.w-1] = true
	r.seeded = true
}

func (r *Rule110) step() {
	rule := int(r.params["rule"])
	if rule < 0 || rule > 255 {
		rule = 110
	}

	currentRow := r.row
	nextRow := (r.row + 1) % r.h

	for x := 0; x < r.w; x++ {
		// Read 3-cell neighborhood from current row (toroidal).
		left := r.grid[currentRow*r.w+(x-1+r.w)%r.w]
		center := r.grid[currentRow*r.w+x]
		right := r.grid[currentRow*r.w+(x+1)%r.w]

		// Encode neighborhood as 3-bit number.
		var pattern int
		if left {
			pattern |= 4
		}
		if center {
			pattern |= 2
		}
		if right {
			pattern |= 1
		}

		// Apply rule: check if bit `pattern` is set in rule number.
		r.grid[nextRow*r.w+x] = (rule>>pattern)&1 == 1
	}

	r.row = nextRow
}

func (r *Rule110) Step(frame uint64, dt float64) {
	if r.w == 0 || r.h == 0 {
		return
	}
	if !r.seeded {
		r.seed()
	}

	speed := r.params["speed"]
	r.accum += dt * speed
	for r.accum >= 1.0 {
		r.step()
		r.accum -= 1.0
	}
}

func (r *Rule110) Render(c *canvas.Canvas) {
	r.ensureSize(c.Width, c.Height)

	alive := canvas.Hex("#00FFAA")
	dead := canvas.Hex("#0A0A12")

	// Render with newest generation at the bottom.
	for dy := 0; dy < c.Height; dy++ {
		// Map display row to circular buffer row.
		srcRow := (r.row - c.Height + 1 + dy + r.h*2) % r.h
		for x := 0; x < c.Width; x++ {
			if r.grid[srcRow*r.w+x] {
				c.SetPixel(x, dy, alive)
			} else {
				c.SetPixel(x, dy, dead)
			}
		}
	}
}

func (r *Rule110) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := r.params[k]; exists {
			r.params[k] = v
		}
	}
}

func (r *Rule110) Params() map[string]float64 {
	out := make(map[string]float64, len(r.params))
	for k, v := range r.params {
		out[k] = v
	}
	return out
}
