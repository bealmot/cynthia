package effect

import (
	"github.com/bealmot/cynthia/canvas"
)

// Contour renders animated iso-level curves through a time-varying scalar field.
// Uses marching squares to determine contour line positions, rendered as
// colored lines on a dark background. Multiple iso-levels create a
// topographic map appearance.
type Contour struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("contour", func() Effect { return NewContour() })
}

func NewContour() *Contour {
	return &Contour{
		params: map[string]float64{
			"speed":  0.4,
			"scale":  3.0,
			"levels": 8,
		},
	}
}

func (ct *Contour) Name() string { return "contour" }

func (ct *Contour) Step(frame uint64, dt float64) {
	ct.time += dt * ct.params["speed"]
}

func (ct *Contour) Render(c *canvas.Canvas) {
	scale := ct.params["scale"]
	levels := int(ct.params["levels"])
	t := ct.time

	if levels < 2 {
		levels = 8
	}

	bg := canvas.Hex("#06060F")
	c.Clear(bg)

	fw := float64(c.Width)
	fh := float64(c.Height)

	// Sample the scalar field at each pixel.
	field := make([]float64, c.Width*c.Height)
	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh * scale
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw * scale
			// Animated scalar field using layered noise.
			v := fbm(fx, fy, t, 4, 2.0, 0.5)
			// Normalize from [-1,1] to [0,1].
			v = (v + 1.0) * 0.5
			field[y*c.Width+x] = v
		}
	}

	// Draw contour lines via marching squares.
	for l := 1; l < levels; l++ {
		threshold := float64(l) / float64(levels)
		levelT := float64(l) / float64(levels)
		lineColor := canvas.Viridis.Sample(levelT)

		// Brightness scales with level for depth.
		bright := 0.5 + 0.5*levelT
		lineColor.R *= bright
		lineColor.G *= bright
		lineColor.B *= bright

		for y := 0; y < c.Height-1; y++ {
			for x := 0; x < c.Width-1; x++ {
				// Sample 2x2 cell corners.
				tl := field[y*c.Width+x]
				tr := field[y*c.Width+x+1]
				bl := field[(y+1)*c.Width+x]
				br := field[(y+1)*c.Width+x+1]

				// Marching squares case: 4-bit index from corner states.
				var caseIdx int
				if tl >= threshold {
					caseIdx |= 8
				}
				if tr >= threshold {
					caseIdx |= 4
				}
				if br >= threshold {
					caseIdx |= 2
				}
				if bl >= threshold {
					caseIdx |= 1
				}

				// Skip empty (0) and full (15) cells — no contour.
				if caseIdx == 0 || caseIdx == 15 {
					continue
				}

				// Draw contour pixel at cell position.
				c.SetPixel(x, y, lineColor)
			}
		}
	}
}

func (ct *Contour) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := ct.params[k]; exists {
			ct.params[k] = v
		}
	}
}

func (ct *Contour) Params() map[string]float64 {
	out := make(map[string]float64, len(ct.params))
	for k, v := range ct.params {
		out[k] = v
	}
	return out
}
