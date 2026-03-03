package effect

import (
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Starfield is a 3D parallax starfield flying through space.
type Starfield struct {
	stars  []star
	params map[string]float64
	rng    *rand.Rand
}

type star struct {
	x, y, z float64 // 3D position, z = depth [0.1, maxDepth]
}

func init() {
	Register("starfield", func() Effect { return NewStarfield() })
}

func NewStarfield() *Starfield {
	sf := &Starfield{
		params: map[string]float64{
			"speed": 1.0,
			"count": 200,
			"depth": 8.0,
		},
		rng: rand.New(rand.NewSource(42)),
	}
	sf.initStars(200)
	return sf
}

func (sf *Starfield) Name() string { return "starfield" }

func (sf *Starfield) initStars(count int) {
	sf.stars = make([]star, count)
	for i := range sf.stars {
		sf.stars[i] = sf.randomStar()
	}
}

func (sf *Starfield) randomStar() star {
	return star{
		x: sf.rng.Float64()*2 - 1, // [-1, 1]
		y: sf.rng.Float64()*2 - 1,
		z: sf.rng.Float64()*sf.params["depth"] + 0.1,
	}
}

func (sf *Starfield) Step(frame uint64, dt float64) {
	speed := sf.params["speed"] * dt
	depth := sf.params["depth"]

	for i := range sf.stars {
		sf.stars[i].z -= speed
		if sf.stars[i].z <= 0.1 {
			sf.stars[i] = sf.randomStar()
			sf.stars[i].z = depth
		}
	}
}

func (sf *Starfield) Render(c *canvas.Canvas) {
	c.Clear(canvas.Black)

	hw := float64(c.Width) / 2
	hh := float64(c.Height) / 2
	depth := sf.params["depth"]

	for _, s := range sf.stars {
		// Project 3D → 2D
		px := int(s.x/s.z*hw + hw)
		py := int(s.y/s.z*hh + hh)

		if px < 0 || px >= c.Width || py < 0 || py >= c.Height {
			continue
		}

		// Brightness inversely proportional to depth
		brightness := 1.0 - (s.z / depth)
		if brightness < 0 {
			brightness = 0
		}
		c.SetPixel(px, py, canvas.RGB(brightness, brightness, brightness))
	}
}

func (sf *Starfield) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, ok := sf.params[k]; ok {
			sf.params[k] = v
		}
	}
}

func (sf *Starfield) Params() map[string]float64 {
	out := make(map[string]float64, len(sf.params))
	for k, v := range sf.params {
		out[k] = v
	}
	return out
}
