package effect

import (
	"math"
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Particles is a particle system with gravity and emitter.
type Particles struct {
	particles []particle
	params    map[string]float64
	rng       *rand.Rand
	time      float64
}

type particle struct {
	x, y   float64 // position in [0,1] normalized coords
	vx, vy float64 // velocity
	life   float64 // remaining life [0,1]
	color  canvas.Color
}

func init() {
	Register("particles", func() Effect { return NewParticles() })
}

func NewParticles() *Particles {
	return &Particles{
		params: map[string]float64{
			"speed":   1.0,
			"count":   100,
			"gravity": 0.5,
			"spread":  0.3,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (p *Particles) Name() string { return "particles" }

func (p *Particles) Step(frame uint64, dt float64) {
	speed := p.params["speed"]
	gravity := p.params["gravity"]
	count := int(p.params["count"])
	spread := p.params["spread"]
	p.time += dt

	// Emit new particles
	for len(p.particles) < count {
		angle := p.rng.Float64() * math.Pi * 2
		vel := (0.5 + p.rng.Float64()*0.5) * spread
		p.particles = append(p.particles, particle{
			x:     0.5,
			y:     0.7, // emit from lower center
			vx:    math.Cos(angle) * vel,
			vy:    math.Sin(angle)*vel - 0.3, // upward bias
			life:  1.0,
			color: sparkleColor(p.rng),
		})
	}

	// Update existing particles
	alive := p.particles[:0]
	for _, pt := range p.particles {
		pt.x += pt.vx * dt * speed
		pt.y += pt.vy * dt * speed
		pt.vy += gravity * dt * speed // gravity pulls down
		pt.life -= dt * 0.5

		if pt.life > 0 {
			alive = append(alive, pt)
		}
	}
	p.particles = alive
}

func (p *Particles) Render(c *canvas.Canvas) {
	for _, pt := range p.particles {
		px := int(pt.x * float64(c.Width))
		py := int(pt.y * float64(c.Height))

		if px < 0 || px >= c.Width || py < 0 || py >= c.Height {
			continue
		}

		// Fade alpha with life
		col := pt.color
		col.R *= pt.life
		col.G *= pt.life
		col.B *= pt.life
		col.A *= pt.life

		c.SetPixel(px, py, col)
	}
}

func sparkleColor(rng *rand.Rand) canvas.Color {
	colors := []canvas.Color{
		canvas.Hex("#C4B5F4"), // lavender
		canvas.Hex("#93E4D4"), // aqua
		canvas.Hex("#F4B8D4"), // rose
		canvas.Hex("#E8D5A3"), // gold
		canvas.Hex("#F5F0E8"), // ivory
	}
	return colors[rng.Intn(len(colors))]
}

func (p *Particles) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, ok := p.params[k]; ok {
			p.params[k] = v
		}
	}
}

func (p *Particles) Params() map[string]float64 {
	out := make(map[string]float64, len(p.params))
	for k, v := range p.params {
		out[k] = v
	}
	return out
}
