package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Spirograph traces hypotrochoid curves — a point on a circle rolling
// inside another circle. The pen draws continuously with trail decay,
// creating intricate geometric patterns.
type Spirograph struct {
	time  float64
	trail []trailPoint
	params map[string]float64
}

type trailPoint struct {
	x, y float64
	age  float64
}

func init() {
	Register("spirograph", func() Effect { return NewSpirograph() })
}

func NewSpirograph() *Spirograph {
	return &Spirograph{
		params: map[string]float64{
			"speed":     1.0,
			"r_outer":   5.0,
			"r_inner":   3.0,
			"offset":    2.0,
			"trail_len": 500,
		},
	}
}

func (s *Spirograph) Name() string { return "spirograph" }

func (s *Spirograph) Step(frame uint64, dt float64) {
	speed := s.params["speed"]
	R := s.params["r_outer"]
	r := s.params["r_inner"]
	d := s.params["offset"]
	maxTrail := int(s.params["trail_len"])

	s.time += dt * speed

	// Hypotrochoid parametric equations.
	t := s.time * 2.0
	diff := R - r
	ratio := diff / r
	x := diff*math.Cos(t) + d*math.Cos(ratio*t)
	y := diff*math.Sin(t) - d*math.Sin(ratio*t)

	// Normalize to [0,1] range.
	maxExtent := R + d
	x = x/(2*maxExtent) + 0.5
	y = y/(2*maxExtent) + 0.5

	s.trail = append(s.trail, trailPoint{x, y, 0})

	// Age existing points and trim.
	for i := range s.trail {
		s.trail[i].age += dt * speed
	}
	if len(s.trail) > maxTrail {
		s.trail = s.trail[len(s.trail)-maxTrail:]
	}
}

func (s *Spirograph) Render(c *canvas.Canvas) {
	bg := canvas.Hex("#080810")
	c.Clear(bg)

	pal := canvas.Palette{
		canvas.Hex("#FF0066"),
		canvas.Hex("#FF6600"),
		canvas.Hex("#FFCC00"),
		canvas.Hex("#00FF88"),
		canvas.Hex("#0088FF"),
		canvas.Hex("#AA00FF"),
		canvas.Hex("#FF0066"),
	}

	fw := float64(c.Width)
	fh := float64(c.Height)
	maxTrail := s.params["trail_len"]

	for i, pt := range s.trail {
		px := int(pt.x * fw)
		py := int(pt.y * fh)

		if px < 0 || px >= c.Width || py < 0 || py >= c.Height {
			continue
		}

		// Color cycles along the trail; brightness fades with age.
		progress := float64(i) / float64(len(s.trail))
		colorT := math.Mod(progress*3+s.time*0.2, 1.0)
		fade := 1.0 - float64(len(s.trail)-1-i)/maxTrail

		col := pal.Sample(colorT)
		col.R *= fade
		col.G *= fade
		col.B *= fade

		c.SetPixel(px, py, col)

		// Slight glow on neighboring pixels.
		for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			gx, gy := px+d[0], py+d[1]
			if gx >= 0 && gx < c.Width && gy >= 0 && gy < c.Height {
				glow := canvas.RGB(col.R*0.3, col.G*0.3, col.B*0.3)
				existing := c.GetPixel(gx, gy)
				blended := canvas.RGB(
					math.Min(existing.R+glow.R, 1),
					math.Min(existing.G+glow.G, 1),
					math.Min(existing.B+glow.B, 1),
				)
				c.SetPixel(gx, gy, blended)
			}
		}
	}
}

func (s *Spirograph) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := s.params[k]; exists {
			s.params[k] = v
		}
	}
}

func (s *Spirograph) Params() map[string]float64 {
	out := make(map[string]float64, len(s.params))
	for k, v := range s.params {
		out[k] = v
	}
	return out
}
