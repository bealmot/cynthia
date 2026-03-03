package effect

import (
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Fire implements the DOOM PSX fire propagation algorithm.
// A 1D heat array at the bottom propagates upward with random
// lateral spread and decay — producing convincing flames.
type Fire struct {
	heat   []float64
	w, h   int
	params map[string]float64
	rng    *rand.Rand
}

func init() {
	Register("fire", func() Effect { return NewFire() })
}

func NewFire() *Fire {
	return &Fire{
		params: map[string]float64{
			"intensity": 1.0,
			"decay":     0.04,
			"spread":    1.5,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (f *Fire) Name() string { return "fire" }

func (f *Fire) ensureSize(w, h int) {
	if f.w == w && f.h == h {
		return
	}
	f.w, f.h = w, h
	f.heat = make([]float64, w*h)
}

func (f *Fire) Step(frame uint64, dt float64) {
	if f.w == 0 || f.h == 0 {
		return
	}

	intensity := f.params["intensity"]
	decay := f.params["decay"]
	spread := f.params["spread"]

	// Set bottom row to max heat
	for x := 0; x < f.w; x++ {
		f.heat[(f.h-1)*f.w+x] = intensity
	}

	// Propagate upward: each pixel takes from the one below with random lateral offset
	for y := 0; y < f.h-1; y++ {
		for x := 0; x < f.w; x++ {
			// Random lateral offset (-1, 0, or +1 scaled by spread)
			dx := f.rng.Intn(3) - 1
			srcX := x + int(float64(dx)*spread)
			if srcX < 0 {
				srcX = 0
			}
			if srcX >= f.w {
				srcX = f.w - 1
			}

			// Copy from below with random decay
			src := f.heat[(y+1)*f.w+srcX]
			d := decay * (0.5 + f.rng.Float64())
			val := src - d
			if val < 0 {
				val = 0
			}
			f.heat[y*f.w+x] = val
		}
	}
}

func (f *Fire) Render(c *canvas.Canvas) {
	f.ensureSize(c.Width, c.Height)

	// Fire palette: black → red → orange → yellow → white
	pal := canvas.Palette{
		canvas.Hex("#000000"),
		canvas.Hex("#1A0000"),
		canvas.Hex("#440000"),
		canvas.Hex("#880000"),
		canvas.Hex("#CC2200"),
		canvas.Hex("#FF4400"),
		canvas.Hex("#FF8800"),
		canvas.Hex("#FFCC00"),
		canvas.Hex("#FFEE66"),
		canvas.Hex("#FFFFFF"),
	}

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			h := f.heat[y*f.w+x]
			if h > 1 {
				h = 1
			}
			c.SetPixel(x, y, pal.Sample(h))
		}
	}
}

func (f *Fire) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := f.params[k]; exists {
			f.params[k] = v
		}
	}
}

func (f *Fire) Params() map[string]float64 {
	out := make(map[string]float64, len(f.params))
	for k, v := range f.params {
		out[k] = v
	}
	return out
}
