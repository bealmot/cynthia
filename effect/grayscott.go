package effect

import (
	"math"
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// GrayScott implements the Gray-Scott reaction-diffusion system.
// Two chemicals U and V interact: U + 2V → 3V, V → P (inert).
// The feed/kill rates determine the emergent pattern type:
//   - spots:  feed=0.055, kill=0.062
//   - stripes: feed=0.035, kill=0.065
//   - maze:   feed=0.029, kill=0.057
type GrayScott struct {
	u, v       []float64
	uNext, vNext []float64
	w, h       int
	time       float64
	params     map[string]float64
	rng        *rand.Rand
	seeded     bool
}

func init() {
	Register("grayscott", func() Effect { return NewGrayScott() })
}

func NewGrayScott() *GrayScott {
	return &GrayScott{
		params: map[string]float64{
			"speed":     1.0,
			"feed":      0.055,
			"kill":      0.062,
			"diffuse_u": 0.16,
			"diffuse_v": 0.08,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (g *GrayScott) Name() string { return "grayscott" }

func (g *GrayScott) ensureSize(w, h int) {
	if g.w == w && g.h == h {
		return
	}
	g.w, g.h = w, h
	n := w * h
	g.u = make([]float64, n)
	g.v = make([]float64, n)
	g.uNext = make([]float64, n)
	g.vNext = make([]float64, n)
	g.seeded = false
}

func (g *GrayScott) seed() {
	// Initialize: U=1 everywhere, seed V in random patches.
	for i := range g.u {
		g.u[i] = 1.0
		g.v[i] = 0.0
	}

	// Drop several seed patches of V.
	patches := 5 + g.rng.Intn(5)
	for p := 0; p < patches; p++ {
		cx := g.rng.Intn(g.w)
		cy := g.rng.Intn(g.h)
		radius := 2 + g.rng.Intn(4)
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx*dx+dy*dy <= radius*radius {
					x := (cx + dx + g.w) % g.w
					y := (cy + dy + g.h) % g.h
					idx := y*g.w + x
					g.u[idx] = 0.5
					g.v[idx] = 0.25 + g.rng.Float64()*0.01
				}
			}
		}
	}
	g.seeded = true
}

func (g *GrayScott) step() {
	f := g.params["feed"]
	k := g.params["kill"]
	du := g.params["diffuse_u"]
	dv := g.params["diffuse_v"]
	dt := 1.0

	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			idx := y*g.w + x

			// Laplacian with toroidal wrapping.
			xp := ((x + 1) % g.w) + y*g.w
			xm := ((x - 1 + g.w) % g.w) + y*g.w
			yp := x + ((y+1)%g.h)*g.w
			ym := x + ((y-1+g.h)%g.h)*g.w

			lapU := g.u[xp] + g.u[xm] + g.u[yp] + g.u[ym] - 4*g.u[idx]
			lapV := g.v[xp] + g.v[xm] + g.v[yp] + g.v[ym] - 4*g.v[idx]

			uv2 := g.u[idx] * g.v[idx] * g.v[idx]

			g.uNext[idx] = g.u[idx] + dt*(du*lapU-uv2+f*(1-g.u[idx]))
			g.vNext[idx] = g.v[idx] + dt*(dv*lapV+uv2-(f+k)*g.v[idx])

			// Clamp.
			if g.uNext[idx] < 0 {
				g.uNext[idx] = 0
			}
			if g.uNext[idx] > 1 {
				g.uNext[idx] = 1
			}
			if g.vNext[idx] < 0 {
				g.vNext[idx] = 0
			}
			if g.vNext[idx] > 1 {
				g.vNext[idx] = 1
			}
		}
	}

	g.u, g.uNext = g.uNext, g.u
	g.v, g.vNext = g.vNext, g.v
}

func (g *GrayScott) Step(frame uint64, dt float64) {
	if g.w == 0 || g.h == 0 {
		return
	}
	if !g.seeded {
		g.seed()
	}

	speed := g.params["speed"]
	// Run multiple substeps per frame for faster pattern emergence.
	steps := int(math.Max(1, speed*8))
	for i := 0; i < steps; i++ {
		g.step()
	}
	g.time += dt
}

func (g *GrayScott) Render(c *canvas.Canvas) {
	g.ensureSize(c.Width, c.Height)

	pal := canvas.Palette{
		canvas.Hex("#0A0020"),
		canvas.Hex("#1B0A4A"),
		canvas.Hex("#3D2B7F"),
		canvas.Hex("#FF6B6B"),
		canvas.Hex("#FFBE76"),
		canvas.Hex("#FFFFFF"),
	}

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			v := g.v[y*g.w+x]
			v = math.Min(v*4.0, 1.0) // amplify for visibility
			c.SetPixel(x, y, pal.Sample(v))
		}
	}
}

func (g *GrayScott) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := g.params[k]; exists {
			g.params[k] = v
		}
	}
}

func (g *GrayScott) Params() map[string]float64 {
	out := make(map[string]float64, len(g.params))
	for k, v := range g.params {
		out[k] = v
	}
	return out
}
