package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Lava renders a lava lamp effect — slow-moving metaballs with
// vertical drift and warm color gradients. Blobs rise and fall
// naturally using sine-modulated positions.
type Lava struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("lava", func() Effect { return NewLava() })
}

func NewLava() *Lava {
	return &Lava{
		params: map[string]float64{
			"speed":  0.3,
			"blobs":  5,
			"radius": 0.12,
		},
	}
}

func (l *Lava) Name() string { return "lava" }

func (l *Lava) Step(frame uint64, dt float64) {
	l.time += dt * l.params["speed"]
}

func (l *Lava) Render(c *canvas.Canvas) {
	blobs := int(l.params["blobs"])
	radius := l.params["radius"]
	t := l.time

	if blobs < 1 {
		blobs = 5
	}

	// Warm lava palette: deep red → orange → yellow → white.
	pal := canvas.Palette{
		canvas.Hex("#1A0000"),
		canvas.Hex("#4A0000"),
		canvas.Hex("#8B0000"),
		canvas.Hex("#CC3300"),
		canvas.Hex("#FF6600"),
		canvas.Hex("#FFAA00"),
		canvas.Hex("#FFDD44"),
		canvas.Hex("#FFFF99"),
	}

	// Compute blob positions — vertical drift + horizontal wobble.
	type blob struct{ x, y float64 }
	bs := make([]blob, blobs)
	for i := 0; i < blobs; i++ {
		fi := float64(i)
		phase := fi * 2.4
		// Vertical: slow rise and fall cycle.
		yBase := 0.5 + 0.35*math.Sin(t*0.4+phase)
		// Horizontal: gentle wobble.
		xBase := 0.5 + 0.2*math.Sin(t*0.3+fi*1.7)*math.Cos(t*0.2+fi*0.9)
		bs[i] = blob{x: xBase, y: yBase}
	}

	fw := float64(c.Width)
	fh := float64(c.Height)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Sum metaball fields.
			var field float64
			for _, b := range bs {
				dx := fx - b.x
				dy := fy - b.y
				dist := dx*dx + dy*dy
				if dist < 1e-6 {
					dist = 1e-6
				}
				field += radius * radius / dist
			}

			// Background heat gradient (hotter at bottom).
			bgHeat := (1.0 - fy) * 0.15

			v := math.Min(field+bgHeat, 1.0)
			c.SetPixel(x, y, pal.Sample(v))
		}
	}
}

func (l *Lava) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := l.params[k]; exists {
			l.params[k] = v
		}
	}
}

func (l *Lava) Params() map[string]float64 {
	out := make(map[string]float64, len(l.params))
	for k, v := range l.params {
		out[k] = v
	}
	return out
}
