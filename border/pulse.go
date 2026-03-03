package border

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Pulse is a breathing/pulsing glow border.
type Pulse struct {
	time   float64
	params map[string]float64
	color  canvas.Color
}

func NewPulse() *Pulse {
	return &Pulse{
		params: map[string]float64{
			"speed": 1.0,
			"min":   0.3,
			"max":   1.0,
		},
		color: canvas.Hex("#C4B5F4"), // lavender
	}
}

func (p *Pulse) Name() string { return "pulse" }

func (p *Pulse) Step(frame uint64, dt float64) {
	p.time += dt * p.params["speed"]
}

func (p *Pulse) Render(c *canvas.Canvas, x, y, w, h int) {
	// Breathing intensity: sinusoidal between min and max
	mn := p.params["min"]
	mx := p.params["max"]
	t := (math.Sin(p.time*math.Pi*2) + 1) * 0.5 // [0, 1]
	intensity := mn + t*(mx-mn)

	col := canvas.Color{
		R: p.color.R * intensity,
		G: p.color.G * intensity,
		B: p.color.B * intensity,
		A: p.color.A,
	}

	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			if px > x && px < x+w-1 && py > y && py < y+h-1 {
				continue
			}

			var r rune
			switch {
			case px == x && py == y:
				r = '╭'
			case px == x+w-1 && py == y:
				r = '╮'
			case px == x+w-1 && py == y+h-1:
				r = '╯'
			case px == x && py == y+h-1:
				r = '╰'
			case py == y || py == y+h-1:
				r = '─'
			default:
				r = '│'
			}

			c.SetCell(px, py, canvas.Cell{
				Rune: r,
				FG:   col,
				BG:   canvas.Transparent,
			})
		}
	}
}

func (p *Pulse) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, ok := p.params[k]; ok {
			p.params[k] = v
		}
	}
}
