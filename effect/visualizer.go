package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Visualizer renders a simulated audio spectrum analyzer.
// Uses sine superposition to generate fake frequency data,
// with exponential smoothing for fluid bar animation.
type Visualizer struct {
	time    float64
	smooth  []float64
	w, h    int
	params  map[string]float64
}

func init() {
	Register("visualizer", func() Effect { return NewVisualizer() })
}

func NewVisualizer() *Visualizer {
	return &Visualizer{
		params: map[string]float64{
			"speed":     1.0,
			"bars":      16,
			"smoothing": 0.15,
			"amplitude": 1.0,
		},
	}
}

func (v *Visualizer) Name() string { return "visualizer" }

func (v *Visualizer) ensureSize(w, h int) {
	bars := int(v.params["bars"])
	if bars < 1 {
		bars = 16
	}
	if v.w == w && v.h == h && len(v.smooth) == bars {
		return
	}
	v.w, v.h = w, h
	v.smooth = make([]float64, bars)
}

func (v *Visualizer) Step(frame uint64, dt float64) {
	v.time += dt * v.params["speed"]

	if len(v.smooth) == 0 {
		return
	}

	amplitude := v.params["amplitude"]
	smoothing := v.params["smoothing"]
	bars := len(v.smooth)
	t := v.time

	for i := 0; i < bars; i++ {
		freq := float64(i+1) * 0.7
		// Superposition of sine waves at different frequencies.
		val := math.Sin(t*freq*0.5) * 0.4
		val += math.Sin(t*freq*0.3+1.5) * 0.3
		val += math.Sin(t*0.8+float64(i)*0.6) * 0.3
		val = (val + 1.0) * 0.5 * amplitude
		if val > 1.0 {
			val = 1.0
		}
		if val < 0 {
			val = 0
		}

		// Exponential moving average for smooth transitions.
		v.smooth[i] += (val - v.smooth[i]) * smoothing
	}
}

func (v *Visualizer) Render(c *canvas.Canvas) {
	v.ensureSize(c.Width, c.Height)

	bars := len(v.smooth)
	if bars == 0 {
		return
	}

	// Spectrum palette: purple → cyan → green → yellow
	pal := canvas.Palette{
		canvas.Hex("#6C3483"),
		canvas.Hex("#2980B9"),
		canvas.Hex("#1ABC9C"),
		canvas.Hex("#27AE60"),
		canvas.Hex("#F1C40F"),
		canvas.Hex("#E74C3C"),
	}
	bg := canvas.Hex("#0A0A0A")

	barWidth := c.Width / bars
	if barWidth < 1 {
		barWidth = 1
	}

	c.Clear(bg)

	for i := 0; i < bars; i++ {
		height := int(v.smooth[i] * float64(c.Height))
		if height > c.Height {
			height = c.Height
		}

		barX := i * barWidth
		// Color based on bar position in spectrum.
		barColor := pal.Sample(float64(i) / float64(bars))

		for y := c.Height - height; y < c.Height; y++ {
			for dx := 0; dx < barWidth-1 && barX+dx < c.Width; dx++ {
				// Slight gradient: brighter at top.
				t := float64(c.Height-y) / float64(c.Height)
				col := barColor.Lerp(canvas.White, t*0.3)
				c.SetPixel(barX+dx, y, col)
			}
		}
	}
}

func (v *Visualizer) SetParams(params map[string]float64) {
	for k, val := range params {
		if _, exists := v.params[k]; exists {
			v.params[k] = val
		}
	}
}

func (v *Visualizer) Params() map[string]float64 {
	out := make(map[string]float64, len(v.params))
	for k, val := range v.params {
		out[k] = val
	}
	return out
}
