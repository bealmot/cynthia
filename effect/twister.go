package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Twister renders a rotating 3D bar/column effect.
// Vertical bars are projected onto a rotating cylinder, creating
// a twisted ribbon illusion reminiscent of classic Amiga demos.
type Twister struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("twister", func() Effect { return NewTwister() })
}

func NewTwister() *Twister {
	return &Twister{
		params: map[string]float64{
			"speed":  1.0,
			"bars":   6,
			"twist":  2.0,
			"width":  0.4,
		},
	}
}

func (tw *Twister) Name() string { return "twister" }

func (tw *Twister) Step(frame uint64, dt float64) {
	tw.time += dt * tw.params["speed"]
}

func (tw *Twister) Render(c *canvas.Canvas) {
	bars := int(tw.params["bars"])
	twist := tw.params["twist"]
	width := tw.params["width"]

	if bars < 2 {
		bars = 6
	}

	colors := canvas.Palette{
		canvas.Hex("#FF0044"),
		canvas.Hex("#FF6600"),
		canvas.Hex("#FFCC00"),
		canvas.Hex("#00FF88"),
		canvas.Hex("#0088FF"),
		canvas.Hex("#AA00FF"),
		canvas.Hex("#FF0044"),
	}

	bg := canvas.Hex("#0A0A1A")
	fw := float64(c.Width)
	fh := float64(c.Height)
	cx := fw / 2.0
	halfW := fw * width

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh

		// Rotation angle varies along y-axis for twist effect.
		angle := tw.time + fy*twist*math.Pi*2

		for x := 0; x < c.Width; x++ {
			dx := float64(x) - cx

			// Only render within the cylinder width.
			if math.Abs(dx) > halfW {
				c.SetPixel(x, y, bg)
				continue
			}

			// Map x position to angle on cylinder surface.
			surfaceAngle := math.Asin(dx/halfW) + angle

			// Determine which bar segment we're on.
			barAngle := math.Mod(surfaceAngle, 2*math.Pi)
			if barAngle < 0 {
				barAngle += 2 * math.Pi
			}
			barIdx := int(barAngle / (2 * math.Pi) * float64(bars))
			barFrac := math.Mod(barAngle/(2*math.Pi)*float64(bars), 1.0)

			// Color based on bar index.
			t := float64(barIdx) / float64(bars)
			col := colors.Sample(t)

			// Shading: darken edges of each bar for 3D look.
			shade := 1.0 - math.Abs(barFrac-0.5)*1.5
			if shade < 0.2 {
				shade = 0.2
			}

			// Additional shading from cylinder curvature.
			curvature := math.Cos(math.Asin(dx / halfW))
			shade *= curvature

			col.R *= shade
			col.G *= shade
			col.B *= shade

			c.SetPixel(x, y, col)
		}
	}
}

func (tw *Twister) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := tw.params[k]; exists {
			tw.params[k] = v
		}
	}
}

func (tw *Twister) Params() map[string]float64 {
	out := make(map[string]float64, len(tw.params))
	for k, v := range tw.params {
		out[k] = v
	}
	return out
}
