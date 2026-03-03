package border

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Cascade is a traveling highlight that races around the border frame.
type Cascade struct {
	time   float64
	params map[string]float64
	base   canvas.Color
	bright canvas.Color
}

func NewCascade() *Cascade {
	return &Cascade{
		params: map[string]float64{
			"speed":    2.0,
			"width":    0.15, // highlight width as fraction of perimeter
			"tail_len": 0.1,
		},
		base:   canvas.Hex("#5A5464"), // dim
		bright: canvas.Hex("#F4B8D4"), // rose
	}
}

func (ca *Cascade) Name() string { return "cascade" }

func (ca *Cascade) Step(frame uint64, dt float64) {
	ca.time += dt * ca.params["speed"]
}

func (ca *Cascade) Render(c *canvas.Canvas, x, y, w, h int) {
	headPos := math.Mod(ca.time, 1.0)
	width := ca.params["width"]

	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			if px > x && px < x+w-1 && py > y && py < y+h-1 {
				continue
			}

			t := perimeterPos(px, py, x, y, w, h)

			// Distance from the highlight head (wrapping around)
			dist := t - headPos
			if dist < 0 {
				dist += 1.0
			}
			if dist > 0.5 {
				dist = 1.0 - dist
			}

			// Intensity based on distance from head
			var intensity float64
			if dist < width {
				intensity = 1.0 - (dist / width)
			}

			col := ca.base.Lerp(ca.bright, intensity)

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

func (ca *Cascade) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, ok := ca.params[k]; ok {
			ca.params[k] = v
		}
	}
}
