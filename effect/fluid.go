package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Fluid implements Jos Stam's stable fluid simulation.
// Steps: diffuse → advect → project (pressure solve).
// Dye injection creates visible swirling patterns.
type Fluid struct {
	// Velocity fields (with ghost cells: size = (w+2)*(h+2))
	vx, vy     []float64
	vx0, vy0   []float64
	// Dye density (3 channels for color)
	dR, dG, dB    []float64
	dR0, dG0, dB0 []float64
	w, h           int // inner grid size (excludes ghost cells)
	time           float64
	params         map[string]float64
}

func init() {
	Register("fluid", func() Effect { return NewFluid() })
}

func NewFluid() *Fluid {
	return &Fluid{
		params: map[string]float64{
			"viscosity": 0.0001,
			"diffusion": 0.0001,
			"speed":     1.0,
			"dye_rate":  0.8,
		},
	}
}

func (f *Fluid) Name() string { return "fluid" }

func (f *Fluid) ensureSize(w, h int) {
	if f.w == w && f.h == h {
		return
	}
	f.w, f.h = w, h
	n := (w + 2) * (h + 2)
	f.vx = make([]float64, n)
	f.vy = make([]float64, n)
	f.vx0 = make([]float64, n)
	f.vy0 = make([]float64, n)
	f.dR = make([]float64, n)
	f.dG = make([]float64, n)
	f.dB = make([]float64, n)
	f.dR0 = make([]float64, n)
	f.dG0 = make([]float64, n)
	f.dB0 = make([]float64, n)
}

// idx converts (x,y) in 1-based inner coords to linear index.
func (f *Fluid) idx(x, y int) int {
	return y*(f.w+2) + x
}

// setBnd sets boundary conditions. b=1 for x-velocity, b=2 for y-velocity, b=0 for scalar.
func (f *Fluid) setBnd(b int, x []float64) {
	w, h := f.w, f.h
	s := w + 2

	for i := 1; i <= w; i++ {
		if b == 2 {
			x[i] = -x[i+s]
			x[i+(h+1)*s] = -x[i+h*s]
		} else {
			x[i] = x[i+s]
			x[i+(h+1)*s] = x[i+h*s]
		}
	}
	for j := 1; j <= h; j++ {
		if b == 1 {
			x[j*s] = -x[1+j*s]
			x[(w+1)+j*s] = -x[w+j*s]
		} else {
			x[j*s] = x[1+j*s]
			x[(w+1)+j*s] = x[w+j*s]
		}
	}

	x[0] = 0.5 * (x[1] + x[s])
	x[(w+1)] = 0.5 * (x[w] + x[(w+1)+s])
	x[(h+1)*s] = 0.5 * (x[1+(h+1)*s] + x[h*s])
	x[(w+1)+(h+1)*s] = 0.5 * (x[w+(h+1)*s] + x[(w+1)+h*s])
}

// diffuse performs implicit diffusion via Gauss-Seidel.
func (f *Fluid) diffuse(b int, x, x0 []float64, diff, dt float64) {
	a := dt * diff * float64(f.w*f.h)
	s := f.w + 2

	for k := 0; k < 20; k++ {
		for j := 1; j <= f.h; j++ {
			for i := 1; i <= f.w; i++ {
				idx := j*s + i
				x[idx] = (x0[idx] + a*(x[idx-1]+x[idx+1]+x[idx-s]+x[idx+s])) / (1 + 4*a)
			}
		}
		f.setBnd(b, x)
	}
}

// advect moves quantities along the velocity field using bilinear interpolation.
func (f *Fluid) advect(b int, d, d0, u, v []float64, dt float64) {
	w, h := f.w, f.h
	s := w + 2
	dt0x := dt * float64(w)
	dt0y := dt * float64(h)

	for j := 1; j <= h; j++ {
		for i := 1; i <= w; i++ {
			idx := j*s + i
			x := float64(i) - dt0x*u[idx]
			y := float64(j) - dt0y*v[idx]

			if x < 0.5 {
				x = 0.5
			}
			if x > float64(w)+0.5 {
				x = float64(w) + 0.5
			}
			if y < 0.5 {
				y = 0.5
			}
			if y > float64(h)+0.5 {
				y = float64(h) + 0.5
			}

			i0 := int(x)
			j0 := int(y)
			i1 := i0 + 1
			j1 := j0 + 1
			sx := x - float64(i0)
			sy := y - float64(j0)

			d[idx] = (1-sx)*((1-sy)*d0[j0*s+i0]+sy*d0[j1*s+i0]) +
				sx*((1-sy)*d0[j0*s+i1]+sy*d0[j1*s+i1])
		}
	}
	f.setBnd(b, d)
}

// project enforces incompressibility via pressure solve.
func (f *Fluid) project(u, v, p, div []float64) {
	w, h := f.w, f.h
	s := w + 2

	for j := 1; j <= h; j++ {
		for i := 1; i <= w; i++ {
			idx := j*s + i
			div[idx] = -0.5 * (u[idx+1] - u[idx-1] + v[idx+s] - v[idx-s]) / float64(w)
			p[idx] = 0
		}
	}
	f.setBnd(0, div)
	f.setBnd(0, p)

	for k := 0; k < 20; k++ {
		for j := 1; j <= h; j++ {
			for i := 1; i <= w; i++ {
				idx := j*s + i
				p[idx] = (div[idx] + p[idx-1] + p[idx+1] + p[idx-s] + p[idx+s]) / 4
			}
		}
		f.setBnd(0, p)
	}

	for j := 1; j <= h; j++ {
		for i := 1; i <= w; i++ {
			idx := j*s + i
			u[idx] -= 0.5 * float64(w) * (p[idx+1] - p[idx-1])
			v[idx] -= 0.5 * float64(h) * (p[idx+s] - p[idx-s])
		}
	}
	f.setBnd(1, u)
	f.setBnd(2, v)
}

func (f *Fluid) Step(frame uint64, dt float64) {
	if f.w == 0 || f.h == 0 {
		return
	}

	speed := f.params["speed"]
	visc := f.params["viscosity"]
	diff := f.params["diffusion"]
	dyeRate := f.params["dye_rate"]
	sdt := dt * speed

	// Inject dye and velocity in a swirling pattern using noise.
	f.time += sdt
	cx := float64(f.w) / 2.0
	cy := float64(f.h) / 2.0
	r := float64(f.w) * 0.3

	// Two swirling injection points.
	for _, angle := range []float64{f.time, f.time + math.Pi} {
		ix := int(cx + r*math.Cos(angle))
		iy := int(cy + r*math.Sin(angle))
		if ix >= 1 && ix <= f.w && iy >= 1 && iy <= f.h {
			idx := f.idx(ix, iy)
			// Velocity tangent to circle.
			f.vx[idx] += -math.Sin(angle) * 5.0 * sdt
			f.vy[idx] += math.Cos(angle) * 5.0 * sdt

			// Colored dye using noise for organic variation.
			n := noise2(float64(ix)*0.1, f.time*0.5)
			hue := math.Mod(f.time*0.3+n*0.5, 1.0)
			if hue < 0 {
				hue += 1.0
			}
			dr, dg, db := hueToRGB(hue)
			f.dR[idx] += dr * dyeRate * sdt * 10
			f.dG[idx] += dg * dyeRate * sdt * 10
			f.dB[idx] += db * dyeRate * sdt * 10
		}
	}

	// Velocity step: diffuse → project → advect → project.
	copy(f.vx0, f.vx)
	copy(f.vy0, f.vy)
	f.diffuse(1, f.vx, f.vx0, visc, sdt)
	f.diffuse(2, f.vy, f.vy0, visc, sdt)
	f.project(f.vx, f.vy, f.vx0, f.vy0)

	copy(f.vx0, f.vx)
	copy(f.vy0, f.vy)
	f.advect(1, f.vx, f.vx0, f.vx0, f.vy0, sdt)
	f.advect(2, f.vy, f.vy0, f.vx0, f.vy0, sdt)
	f.project(f.vx, f.vy, f.vx0, f.vy0)

	// Dye step: diffuse → advect (per channel).
	copy(f.dR0, f.dR)
	copy(f.dG0, f.dG)
	copy(f.dB0, f.dB)
	f.diffuse(0, f.dR, f.dR0, diff, sdt)
	f.diffuse(0, f.dG, f.dG0, diff, sdt)
	f.diffuse(0, f.dB, f.dB0, diff, sdt)

	copy(f.dR0, f.dR)
	copy(f.dG0, f.dG)
	copy(f.dB0, f.dB)
	f.advect(0, f.dR, f.dR0, f.vx, f.vy, sdt)
	f.advect(0, f.dG, f.dG0, f.vx, f.vy, sdt)
	f.advect(0, f.dB, f.dB0, f.vx, f.vy, sdt)
}

func (f *Fluid) Render(c *canvas.Canvas) {
	f.ensureSize(c.Width, c.Height)

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			idx := f.idx(x+1, y+1)
			r := math.Min(f.dR[idx], 1.0)
			g := math.Min(f.dG[idx], 1.0)
			b := math.Min(f.dB[idx], 1.0)
			if r < 0 {
				r = 0
			}
			if g < 0 {
				g = 0
			}
			if b < 0 {
				b = 0
			}
			c.SetPixel(x, y, canvas.RGB(r, g, b))
		}
	}
}

// hueToRGB converts a hue [0,1] to RGB using HSV with S=1, V=1.
func hueToRGB(h float64) (r, g, b float64) {
	h = h * 6.0
	i := int(h)
	f := h - float64(i)
	switch i % 6 {
	case 0:
		return 1, f, 0
	case 1:
		return 1 - f, 1, 0
	case 2:
		return 0, 1, f
	case 3:
		return 0, 1 - f, 1
	case 4:
		return f, 0, 1
	default:
		return 1, 0, 1 - f
	}
}

func (f *Fluid) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := f.params[k]; exists {
			f.params[k] = v
		}
	}
}

func (f *Fluid) Params() map[string]float64 {
	out := make(map[string]float64, len(f.params))
	for k, v := range f.params {
		out[k] = v
	}
	return out
}
