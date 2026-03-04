package effect

import (
	"math"
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Glitch renders digital corruption artifacts — random block displacement,
// color channel separation, and scan line corruption. Effects trigger in
// bursts rather than uniformly for a more convincing "broken signal" look.
type Glitch struct {
	time     float64
	burstT   float64
	bursting bool
	w, h     int
	base     []canvas.Color // base image to corrupt
	params   map[string]float64
	rng      *rand.Rand
}

func init() {
	Register("glitch", func() Effect { return NewGlitch() })
}

func NewGlitch() *Glitch {
	return &Glitch{
		params: map[string]float64{
			"speed":     1.0,
			"intensity": 0.5,
			"rate":      2.0,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (g *Glitch) Name() string { return "glitch" }

func (g *Glitch) ensureSize(w, h int) {
	if g.w == w && g.h == h {
		return
	}
	g.w, g.h = w, h
	g.base = make([]canvas.Color, w*h)
}

func (g *Glitch) Step(frame uint64, dt float64) {
	g.time += dt * g.params["speed"]

	// Burst control: glitches come in waves.
	rate := g.params["rate"]
	g.burstT -= dt * g.params["speed"]
	if g.burstT <= 0 {
		g.bursting = !g.bursting
		if g.bursting {
			g.burstT = 0.1 + g.rng.Float64()*0.3 // burst duration
		} else {
			g.burstT = 1.0/rate + g.rng.Float64()/rate // quiet period
		}
	}
}

func (g *Glitch) Render(c *canvas.Canvas) {
	g.ensureSize(c.Width, c.Height)

	intensity := g.params["intensity"]

	// Generate base pattern (animated gradient + noise as corruption source).
	for y := 0; y < c.Height; y++ {
		fy := float64(y) / float64(c.Height)
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / float64(c.Width)
			r := math.Sin(fx*math.Pi*2+g.time)*0.3 + 0.4
			gr := math.Sin(fy*math.Pi*3+g.time*1.3)*0.3 + 0.3
			b := math.Sin((fx+fy)*math.Pi*4+g.time*0.7)*0.3 + 0.5
			g.base[y*g.w+x] = canvas.RGB(r, gr, b)
		}
	}

	// Copy base to canvas.
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			c.SetPixel(x, y, g.base[y*g.w+x])
		}
	}

	if !g.bursting {
		return
	}

	// === Glitch effects during burst ===

	// 1. Block displacement: shift random horizontal strips.
	numStrips := 2 + g.rng.Intn(int(intensity*6)+1)
	for s := 0; s < numStrips; s++ {
		y0 := g.rng.Intn(c.Height)
		h := 1 + g.rng.Intn(4)
		dx := g.rng.Intn(int(intensity*20)+1) - int(intensity*10)

		for dy := 0; dy < h; dy++ {
			y := y0 + dy
			if y >= c.Height {
				break
			}
			for x := 0; x < c.Width; x++ {
				srcX := x - dx
				if srcX < 0 || srcX >= c.Width {
					c.SetPixel(x, y, canvas.Hex("#000000"))
				} else {
					c.SetPixel(x, y, g.base[y*g.w+srcX])
				}
			}
		}
	}

	// 2. Color channel separation on random rows.
	if intensity > 0.3 {
		for i := 0; i < int(intensity*5); i++ {
			y := g.rng.Intn(c.Height)
			shift := 1 + g.rng.Intn(3)
			for x := 0; x < c.Width; x++ {
				existing := c.GetPixel(x, y)
				// Shift red channel right.
				redSrc := c.GetPixel(x+shift, y)
				// Shift blue channel left.
				blueSrc := c.GetPixel(x-shift, y)
				c.SetPixel(x, y, canvas.RGB(redSrc.R, existing.G, blueSrc.B))
			}
		}
	}

	// 3. White noise scanlines.
	if intensity > 0.5 {
		for i := 0; i < int(intensity*3); i++ {
			y := g.rng.Intn(c.Height)
			for x := 0; x < c.Width; x++ {
				if g.rng.Float64() < 0.3 {
					v := g.rng.Float64()
					c.SetPixel(x, y, canvas.RGB(v, v, v))
				}
			}
		}
	}
}

// RenderPost applies glitch corruption to existing canvas pixels.
// Unlike Render, it reads the existing pixel buffer instead of generating content.
func (g *Glitch) RenderPost(c *canvas.Canvas) {
	g.ensureSize(c.Width, c.Height)
	intensity := g.params["intensity"]

	// Snapshot existing pixels as the base to corrupt.
	copy(g.base, c.Pixels)

	if !g.bursting {
		return // no glitch during quiet periods
	}

	// 1. Block displacement.
	numStrips := 2 + g.rng.Intn(int(intensity*6)+1)
	for s := 0; s < numStrips; s++ {
		y0 := g.rng.Intn(c.Height)
		h := 1 + g.rng.Intn(4)
		dx := g.rng.Intn(int(intensity*20)+1) - int(intensity*10)

		for dy := 0; dy < h; dy++ {
			y := y0 + dy
			if y >= c.Height {
				break
			}
			for x := 0; x < c.Width; x++ {
				srcX := x - dx
				if srcX < 0 || srcX >= c.Width {
					c.SetPixel(x, y, canvas.Black)
				} else {
					c.SetPixel(x, y, g.base[y*g.w+srcX])
				}
			}
		}
	}

	// 2. Color channel separation.
	if intensity > 0.3 {
		for i := 0; i < int(intensity*5); i++ {
			y := g.rng.Intn(c.Height)
			shift := 1 + g.rng.Intn(3)
			for x := 0; x < c.Width; x++ {
				existing := c.GetPixel(x, y)
				redSrc := c.GetPixel(x+shift, y)
				blueSrc := c.GetPixel(x-shift, y)
				c.SetPixel(x, y, canvas.RGB(redSrc.R, existing.G, blueSrc.B))
			}
		}
	}

	// 3. White noise scanlines.
	if intensity > 0.5 {
		for i := 0; i < int(intensity*3); i++ {
			y := g.rng.Intn(c.Height)
			for x := 0; x < c.Width; x++ {
				if g.rng.Float64() < 0.3 {
					v := g.rng.Float64()
					c.SetPixel(x, y, canvas.RGB(v, v, v))
				}
			}
		}
	}
}

func (g *Glitch) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := g.params[k]; exists {
			g.params[k] = v
		}
	}
}

func (g *Glitch) Params() map[string]float64 {
	out := make(map[string]float64, len(g.params))
	for k, v := range g.params {
		out[k] = v
	}
	return out
}
