package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Copper renders horizontal color gradient bars with sine-wave displacement.
// Inspired by the Amiga copper/raster bar effect — smooth color cycling
// across horizontal bands that wave back and forth.
type Copper struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("copper", func() Effect { return NewCopper() })
}

func NewCopper() *Copper {
	return &Copper{
		params: map[string]float64{
			"speed":  1.0,
			"bars":   8,
			"wave":   0.3,
		},
	}
}

func (co *Copper) Name() string { return "copper" }

func (co *Copper) Step(frame uint64, dt float64) {
	co.time += dt * co.params["speed"]
}

func (co *Copper) Render(c *canvas.Canvas) {
	bars := co.params["bars"]
	wave := co.params["wave"]
	t := co.time

	pal := canvas.Palette{
		canvas.Hex("#FF0000"),
		canvas.Hex("#FF8800"),
		canvas.Hex("#FFFF00"),
		canvas.Hex("#00FF00"),
		canvas.Hex("#0088FF"),
		canvas.Hex("#8800FF"),
		canvas.Hex("#FF0088"),
		canvas.Hex("#FF0000"),
	}

	fh := float64(c.Height)
	fw := float64(c.Width)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh

		// Sine-wave horizontal displacement per scanline.
		offset := math.Sin(fy*math.Pi*4+t*2) * wave

		// Bar color based on y position + time scroll.
		barPos := math.Mod(fy*bars+t*0.5+offset, 1.0)
		if barPos < 0 {
			barPos += 1.0
		}
		barColor := pal.Sample(barPos)

		// Brightness modulation: brighter at center of each bar.
		barPhase := math.Mod(fy*bars+t*0.5, 1.0)
		brightness := 0.5 + 0.5*math.Cos(barPhase*math.Pi*2)
		brightness = 0.3 + brightness*0.7

		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Subtle horizontal gradient for depth.
			hGrad := 0.8 + 0.2*math.Sin(fx*math.Pi)

			col := canvas.RGB(
				barColor.R*brightness*hGrad,
				barColor.G*brightness*hGrad,
				barColor.B*brightness*hGrad,
			)
			c.SetPixel(x, y, col)
		}
	}
}

func (co *Copper) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := co.params[k]; exists {
			co.params[k] = v
		}
	}
}

func (co *Copper) Params() map[string]float64 {
	out := make(map[string]float64, len(co.params))
	for k, v := range co.params {
		out[k] = v
	}
	return out
}
