package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Ripple renders concentric water rings using the 2D wave equation.
// Two height buffers are alternated each frame — each cell averages
// its 4 neighbors minus the previous value, with damping.
type Ripple struct {
	buf1, buf2 []float64
	w, h       int
	time       float64
	accum      float64
	params     map[string]float64
}

func init() {
	Register("ripple", func() Effect { return NewRipple() })
}

func NewRipple() *Ripple {
	return &Ripple{
		params: map[string]float64{
			"speed":   1.0,
			"damping": 0.98,
			"rate":    2.0,
		},
	}
}

func (r *Ripple) Name() string { return "ripple" }

func (r *Ripple) ensureSize(w, h int) {
	if r.w == w && r.h == h {
		return
	}
	r.w, r.h = w, h
	r.buf1 = make([]float64, w*h)
	r.buf2 = make([]float64, w*h)
}

func (r *Ripple) Step(frame uint64, dt float64) {
	if r.w == 0 || r.h == 0 {
		return
	}

	speed := r.params["speed"]
	damping := r.params["damping"]
	rate := r.params["rate"]

	r.time += dt * speed
	r.accum += dt * speed * rate

	// Drop new ripple periodically.
	for r.accum >= 1.0 {
		r.accum -= 1.0
		// Drop position orbits and drifts.
		cx := r.w/2 + int(float64(r.w)*0.3*math.Sin(r.time*0.7))
		cy := r.h/2 + int(float64(r.h)*0.3*math.Cos(r.time*0.5))
		if cx >= 1 && cx < r.w-1 && cy >= 1 && cy < r.h-1 {
			r.buf1[cy*r.w+cx] = 1.0
		}
	}

	// Wave equation step.
	for y := 1; y < r.h-1; y++ {
		for x := 1; x < r.w-1; x++ {
			idx := y*r.w + x
			val := (r.buf1[idx-1]+r.buf1[idx+1]+r.buf1[idx-r.w]+r.buf1[idx+r.w])/2.0 - r.buf2[idx]
			val *= damping
			r.buf2[idx] = val
		}
	}

	// Swap buffers.
	r.buf1, r.buf2 = r.buf2, r.buf1
}

func (r *Ripple) Render(c *canvas.Canvas) {
	r.ensureSize(c.Width, c.Height)

	deepColor := canvas.Hex("#001133")
	shallowColor := canvas.Hex("#0066CC")
	highColor := canvas.Hex("#99CCFF")

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			v := r.buf1[y*r.w+x]

			var col canvas.Color
			if v > 0 {
				t := math.Min(v*3.0, 1.0)
				col = shallowColor.Lerp(highColor, t)
			} else {
				t := math.Min(-v*3.0, 1.0)
				col = deepColor.Lerp(shallowColor, 1.0-t)
			}
			c.SetPixel(x, y, col)
		}
	}
}

func (r *Ripple) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := r.params[k]; exists {
			r.params[k] = v
		}
	}
}

func (r *Ripple) Params() map[string]float64 {
	out := make(map[string]float64, len(r.params))
	for k, v := range r.params {
		out[k] = v
	}
	return out
}
