package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Rotozoom renders a rotating and zooming checkerboard pattern.
// Uses 2D affine transform with time-varying rotation and scale.
type Rotozoom struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("rotozoom", func() Effect { return NewRotozoom() })
}

func NewRotozoom() *Rotozoom {
	return &Rotozoom{
		params: map[string]float64{
			"speed":      0.5,
			"zoom_speed": 0.3,
			"scale":      6.0,
		},
	}
}

func (r *Rotozoom) Name() string { return "rotozoom" }

func (r *Rotozoom) Step(frame uint64, dt float64) {
	r.time += dt * r.params["speed"]
}

func (r *Rotozoom) Render(c *canvas.Canvas) {
	scale := r.params["scale"]
	zoomSpeed := r.params["zoom_speed"]

	angle := r.time
	zoom := 1.0 + 0.5*math.Sin(r.time*zoomSpeed)

	cosA := math.Cos(angle) * zoom
	sinA := math.Sin(angle) * zoom

	pal := canvas.Palette{
		canvas.Hex("#1A0533"),
		canvas.Hex("#4A1C8B"),
		canvas.Hex("#8B2FC9"),
		canvas.Hex("#D35FFF"),
		canvas.Hex("#FFB3FF"),
		canvas.Hex("#D35FFF"),
		canvas.Hex("#8B2FC9"),
		canvas.Hex("#4A1C8B"),
	}

	fw := float64(c.Width)
	fh := float64(c.Height)

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			// Center and normalize.
			fx := (float64(x)/fw - 0.5) * scale
			fy := (float64(y)/fh - 0.5) * scale

			// Affine transform.
			u := fx*cosA - fy*sinA
			v := fx*sinA + fy*cosA

			// Checkerboard pattern using floor parity.
			ix := int(math.Floor(u))
			iy := int(math.Floor(v))
			check := (ix + iy) & 1

			// Smooth color within each cell.
			fu := u - math.Floor(u)
			fv := v - math.Floor(v)
			dist := math.Sqrt((fu-0.5)*(fu-0.5) + (fv-0.5)*(fv-0.5))

			var t float64
			if check == 0 {
				t = 0.2 + dist*0.3
			} else {
				t = 0.7 + dist*0.3
			}
			t = math.Mod(t+r.time*0.1, 1.0)

			c.SetPixel(x, y, pal.Sample(t))
		}
	}
}

func (r *Rotozoom) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := r.params[k]; exists {
			r.params[k] = v
		}
	}
}

func (r *Rotozoom) Params() map[string]float64 {
	out := make(map[string]float64, len(r.params))
	for k, v := range r.params {
		out[k] = v
	}
	return out
}
