package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Plasma generates sine interference patterns with palette cycling.
type Plasma struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("plasma", func() Effect { return NewPlasma() })
}

// NewPlasma creates a plasma effect with default parameters.
func NewPlasma() *Plasma {
	return &Plasma{
		params: map[string]float64{
			"speed":     1.0,
			"scale":     4.0,
			"intensity": 1.0,
			"offset":    0.0,
		},
	}
}

func (p *Plasma) Name() string { return "plasma" }

func (p *Plasma) Step(frame uint64, dt float64) {
	p.time += dt * p.params["speed"]
}

func (p *Plasma) Render(c *canvas.Canvas) {
	scale := p.params["scale"]
	intensity := p.params["intensity"]
	offset := p.params["offset"]
	t := p.time

	// Default palette if none configured — purple/teal demoscene classic
	pal := canvas.Palette{
		canvas.Hex("#0D0221"),
		canvas.Hex("#4A1C6C"),
		canvas.Hex("#9B59B6"),
		canvas.Hex("#3498DB"),
		canvas.Hex("#1ABC9C"),
		canvas.Hex("#F39C12"),
		canvas.Hex("#9B59B6"),
		canvas.Hex("#0D0221"),
	}

	fw := float64(c.Width)
	fh := float64(c.Height)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Four overlapping sine waves at different scales/speeds
			v := math.Sin(fx*scale*math.Pi + t)
			v += math.Sin(fy*scale*math.Pi + t*0.7)
			v += math.Sin((fx+fy)*scale*0.5*math.Pi + t*1.3)
			dist := math.Sqrt((fx-0.5)*(fx-0.5) + (fy-0.5)*(fy-0.5))
			v += math.Sin(dist*scale*math.Pi + t*0.5)

			// Normalize to [0,1] — sum range is [-4,4]
			v = (v/4.0 + 1.0) * 0.5
			v = v*intensity + offset
			v = math.Mod(v, 1.0)
			if v < 0 {
				v += 1.0
			}

			c.SetPixel(x, y, pal.Sample(v))
		}
	}
}

func (p *Plasma) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := p.params[k]; exists {
			p.params[k] = v
		}
	}
}

func (p *Plasma) Params() map[string]float64 {
	out := make(map[string]float64, len(p.params))
	for k, v := range p.params {
		out[k] = v
	}
	return out
}
