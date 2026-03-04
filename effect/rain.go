package effect

import (
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Rain renders falling rain streaks with variable speed and density.
// Each column has an independent drop with random speed and brightness,
// creating a natural rainfall appearance.
type Rain struct {
	drops  []raindrop
	w, h   int
	params map[string]float64
	rng    *rand.Rand
}

type raindrop struct {
	y      float64
	speed  float64
	length int
	bright float64
}

func init() {
	Register("rain", func() Effect { return NewRain() })
}

func NewRain() *Rain {
	return &Rain{
		params: map[string]float64{
			"speed":   1.0,
			"density": 0.4,
			"wind":    0.0,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (r *Rain) Name() string { return "rain" }

func (r *Rain) ensureSize(w, h int) {
	if r.w == w && r.h == h {
		return
	}
	r.w, r.h = w, h
	r.drops = make([]raindrop, w)
	for i := range r.drops {
		r.resetDrop(i)
		r.drops[i].y = r.rng.Float64() * float64(h) // stagger initial positions
	}
}

func (r *Rain) resetDrop(i int) {
	r.drops[i] = raindrop{
		y:      -float64(r.rng.Intn(r.h)),
		speed:  0.5 + r.rng.Float64()*1.5,
		length: 2 + r.rng.Intn(4),
		bright: 0.3 + r.rng.Float64()*0.7,
	}
}

func (r *Rain) Step(frame uint64, dt float64) {
	if r.w == 0 {
		return
	}

	speed := r.params["speed"]
	density := r.params["density"]

	for i := range r.drops {
		r.drops[i].y += r.drops[i].speed * speed * dt * float64(r.h)

		// Reset drops that fell off bottom.
		if r.drops[i].y > float64(r.h)+float64(r.drops[i].length) {
			if r.rng.Float64() < density {
				r.resetDrop(i)
			} else {
				// Keep offscreen — not all columns active at once.
				r.drops[i].y = float64(r.h) + 100
			}
		}
	}
}

func (r *Rain) Render(c *canvas.Canvas) {
	r.ensureSize(c.Width, c.Height)

	bg := canvas.Hex("#080810")
	c.Clear(bg)

	dropColor := canvas.Hex("#6688CC")
	splashColor := canvas.Hex("#4466AA")

	for x := 0; x < c.Width; x++ {
		d := r.drops[x]
		headY := int(d.y)

		// Draw drop streak.
		for dy := 0; dy < d.length; dy++ {
			py := headY - dy
			if py < 0 || py >= c.Height {
				continue
			}
			// Fade from bright at head to dim at tail.
			fade := 1.0 - float64(dy)/float64(d.length)
			bright := d.bright * fade
			col := canvas.RGB(
				dropColor.R*bright,
				dropColor.G*bright,
				dropColor.B*bright,
			)
			c.SetPixel(x, py, col)
		}

		// Splash at bottom.
		if headY >= c.Height-1 && headY < c.Height+2 {
			splashY := c.Height - 1
			fade := 1.0 - float64(headY-c.Height+1)*0.5
			if fade > 0 {
				for dx := -1; dx <= 1; dx++ {
					sx := x + dx
					if sx >= 0 && sx < c.Width {
						col := canvas.RGB(
							splashColor.R*fade*0.5,
							splashColor.G*fade*0.5,
							splashColor.B*fade*0.5,
						)
						c.SetPixel(sx, splashY, col)
					}
				}
			}
		}
	}
}

func (r *Rain) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := r.params[k]; exists {
			r.params[k] = v
		}
	}
}

func (r *Rain) Params() map[string]float64 {
	out := make(map[string]float64, len(r.params))
	for k, v := range r.params {
		out[k] = v
	}
	return out
}
