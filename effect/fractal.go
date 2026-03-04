package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Fractal renders an animated Mandelbrot zoom with smooth coloring.
// Uses log2(log2(|z|)) for continuous iteration counts to eliminate banding.
type Fractal struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("fractal", func() Effect { return NewFractal() })
}

func NewFractal() *Fractal {
	return &Fractal{
		params: map[string]float64{
			"speed":         0.3,
			"max_iter":      64,
			"zoom_rate":     0.5,
			"palette_speed": 0.5,
			"center_x":     -0.7435,
			"center_y":      0.1314,
		},
	}
}

func (f *Fractal) Name() string { return "fractal" }

func (f *Fractal) Step(frame uint64, dt float64) {
	f.time += dt * f.params["speed"]
}

func (f *Fractal) Render(c *canvas.Canvas) {
	maxIter := int(f.params["max_iter"])
	zoomRate := f.params["zoom_rate"]
	palSpeed := f.params["palette_speed"]
	cx := f.params["center_x"]
	cy := f.params["center_y"]

	// Continuous zoom: exponential scale decrease over time.
	zoom := math.Exp(-f.time * zoomRate)
	if zoom < 1e-14 {
		zoom = 1e-14 // prevent underflow
	}

	// Fractal palette: deep space to neon.
	pal := canvas.Palette{
		canvas.Hex("#000764"),
		canvas.Hex("#206BCB"),
		canvas.Hex("#EDFFFF"),
		canvas.Hex("#FFAA00"),
		canvas.Hex("#000200"),
		canvas.Hex("#000764"),
	}

	fw := float64(c.Width)
	fh := float64(c.Height)
	aspect := fw / fh

	for py := 0; py < c.Height; py++ {
		for px := 0; px < c.Width; px++ {
			// Map pixel to complex plane.
			x0 := cx + (float64(px)/fw-0.5)*zoom*aspect*4.0
			y0 := cy + (float64(py)/fh-0.5)*zoom*4.0

			var x, y float64
			var i int
			for i = 0; i < maxIter; i++ {
				x2 := x * x
				y2 := y * y
				if x2+y2 > 256 { // escape radius² = 256 for smooth coloring
					break
				}
				y = 2*x*y + y0
				x = x2 - y2 + x0
			}

			var col canvas.Color
			if i == maxIter {
				col = canvas.Black
			} else {
				// Smooth iteration count.
				log_zn := math.Log(x*x+y*y) / 2.0
				nu := math.Log(log_zn/math.Ln2) / math.Ln2
				smooth := float64(i) + 1.0 - nu

				// Map to palette with time-based cycling.
				t := math.Mod(smooth/float64(maxIter)*4.0+f.time*palSpeed, 1.0)
				if t < 0 {
					t += 1.0
				}
				col = pal.Sample(t)
			}

			c.SetPixel(px, py, col)
		}
	}
}

func (f *Fractal) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := f.params[k]; exists {
			f.params[k] = v
		}
	}
}

func (f *Fractal) Params() map[string]float64 {
	out := make(map[string]float64, len(f.params))
	for k, v := range f.params {
		out[k] = v
	}
	return out
}
