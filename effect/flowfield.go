package effect

import (
	"math"
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// FlowField visualizes a vector field using particles advected along
// curl noise. Particles trace flow lines with trail decay, creating
// organic streaming patterns. The vector field is derived from the
// curl of a 2D noise potential, guaranteeing divergence-free (swirly) flow.
type FlowField struct {
	time      float64
	particles []flowParticle
	trail     []float64 // brightness buffer for trail decay
	w, h      int
	params    map[string]float64
	rng       *rand.Rand
}

type flowParticle struct {
	x, y float64
	life float64
}

func init() {
	Register("flowfield", func() Effect { return NewFlowField() })
}

func NewFlowField() *FlowField {
	return &FlowField{
		params: map[string]float64{
			"speed":     0.5,
			"particles": 200,
			"scale":     3.0,
			"trail":     0.95,
			"strength":  1.0,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (f *FlowField) Name() string { return "flowfield" }

func (f *FlowField) ensureSize(w, h int) {
	if f.w == w && f.h == h {
		return
	}
	f.w, f.h = w, h
	f.trail = make([]float64, w*h)

	count := int(f.params["particles"])
	f.particles = make([]flowParticle, count)
	for i := range f.particles {
		f.particles[i] = flowParticle{
			x:    f.rng.Float64(),
			y:    f.rng.Float64(),
			life: f.rng.Float64() * 5.0,
		}
	}
}

// curlNoise computes a divergence-free 2D vector field from the curl
// of a scalar noise potential. This guarantees smooth, swirling flow
// without sources or sinks.
func (f *FlowField) curlNoise(x, y, t float64) (vx, vy float64) {
	const eps = 0.01

	// Curl = (dP/dy, -dP/dx) where P is the noise potential.
	// We use noise3 with time as z for smooth animation.
	pRight := noise3(x+eps, y, t)
	pLeft := noise3(x-eps, y, t)
	pUp := noise3(x, y+eps, t)
	pDown := noise3(x, y-eps, t)

	vx = (pUp - pDown) / (2 * eps)
	vy = -(pRight - pLeft) / (2 * eps)
	return
}

func (f *FlowField) Step(frame uint64, dt float64) {
	if f.w == 0 || f.h == 0 {
		return
	}

	speed := f.params["speed"]
	scale := f.params["scale"]
	trailDecay := f.params["trail"]
	strength := f.params["strength"]
	sdt := dt * speed

	f.time += sdt

	// Decay trail buffer.
	for i := range f.trail {
		f.trail[i] *= trailDecay
	}

	// Advect particles along the curl noise field.
	for i := range f.particles {
		p := &f.particles[i]
		p.life -= sdt

		// Respawn dead particles.
		if p.life <= 0 || p.x < 0 || p.x > 1 || p.y < 0 || p.y > 1 {
			p.x = f.rng.Float64()
			p.y = f.rng.Float64()
			p.life = 2.0 + f.rng.Float64()*3.0
			continue
		}

		// Sample velocity field.
		vx, vy := f.curlNoise(p.x*scale, p.y*scale, f.time*0.5)

		// Move particle.
		p.x += vx * sdt * strength * 0.05
		p.y += vy * sdt * strength * 0.05

		// Stamp into trail buffer.
		px := int(p.x * float64(f.w))
		py := int(p.y * float64(f.h))
		if px >= 0 && px < f.w && py >= 0 && py < f.h {
			f.trail[py*f.w+px] = 1.0
		}
	}
}

func (f *FlowField) Render(c *canvas.Canvas) {
	f.ensureSize(c.Width, c.Height)

	bg := canvas.Hex("#04040C")

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			t := f.trail[y*f.w+x]
			if t < 0.01 {
				c.SetPixel(x, y, bg)
				continue
			}

			// Color trails using the velocity direction for hue.
			fx := float64(x) / float64(f.w)
			fy := float64(y) / float64(f.h)
			scale := f.params["scale"]
			vx, vy := f.curlNoise(fx*scale, fy*scale, f.time*0.5)
			angle := math.Atan2(vy, vx)
			hue := (angle/(2*math.Pi) + 0.5) // [0,1]

			col := canvas.Turbo.Sample(hue)
			col.R *= t
			col.G *= t
			col.B *= t

			c.SetPixel(x, y, col)
		}
	}
}

func (f *FlowField) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := f.params[k]; exists {
			f.params[k] = v
		}
	}
}

func (f *FlowField) Params() map[string]float64 {
	out := make(map[string]float64, len(f.params))
	for k, v := range f.params {
		out[k] = v
	}
	return out
}
