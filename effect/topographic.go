package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Topographic renders an animated topographic/elevation map with filled
// contour bands. Each elevation band gets a distinct color from a
// perceptually uniform palette, with thin contour lines between bands.
type Topographic struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("topographic", func() Effect { return NewTopographic() })
}

func NewTopographic() *Topographic {
	return &Topographic{
		params: map[string]float64{
			"speed":  0.3,
			"scale":  2.5,
			"levels": 10,
		},
	}
}

func (t *Topographic) Name() string { return "topographic" }

func (t *Topographic) Step(frame uint64, dt float64) {
	t.time += dt * t.params["speed"]
}

func (t *Topographic) Render(c *canvas.Canvas) {
	scale := t.params["scale"]
	levels := int(t.params["levels"])
	tm := t.time

	if levels < 2 {
		levels = 10
	}

	// Earth-tone palette for terrain.
	terrainPal := canvas.Palette{
		canvas.Hex("#1A4D2E"), // deep water / low valley
		canvas.Hex("#2D6A4F"),
		canvas.Hex("#40916C"),
		canvas.Hex("#52B788"),
		canvas.Hex("#74C69D"),
		canvas.Hex("#95D5B2"),
		canvas.Hex("#B7E4C7"),
		canvas.Hex("#D8F3DC"),
		canvas.Hex("#E8DCC8"),
		canvas.Hex("#C4A882"),
		canvas.Hex("#A67B5B"),
		canvas.Hex("#8B5E3C"),
		canvas.Hex("#F5F5F0"), // peak / snow
	}

	contourColor := canvas.Hex("#1A1A2E")

	fw := float64(c.Width)
	fh := float64(c.Height)
	fLevels := float64(levels)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh * scale
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw * scale

			// Animated terrain field using fbm.
			v := fbm(fx, fy, tm, 5, 2.0, 0.5)
			// Normalize from [-1,1] to [0,1].
			v = (v + 1.0) * 0.5

			// Quantize to elevation band.
			band := math.Floor(v * fLevels)
			bandFrac := v*fLevels - band
			bandNorm := band / fLevels

			// Detect contour line: thin region at band boundaries.
			isContour := bandFrac < 0.08 || bandFrac > 0.92

			if isContour {
				c.SetPixel(x, y, contourColor)
			} else {
				col := terrainPal.Sample(bandNorm)
				c.SetPixel(x, y, col)
			}
		}
	}
}

func (t *Topographic) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := t.params[k]; exists {
			t.params[k] = v
		}
	}
}

func (t *Topographic) Params() map[string]float64 {
	out := make(map[string]float64, len(t.params))
	for k, v := range t.params {
		out[k] = v
	}
	return out
}
