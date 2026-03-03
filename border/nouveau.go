package border

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Nouveau is an art nouveau border with color cycling along the perimeter.
type Nouveau struct {
	time   float64
	params map[string]float64
	pal    canvas.Palette
}

func NewNouveau() *Nouveau {
	return &Nouveau{
		params: map[string]float64{
			"speed": 0.5,
		},
		pal: canvas.Palette{
			canvas.Hex("#C4B5F4"), // lavender
			canvas.Hex("#93E4D4"), // aqua
			canvas.Hex("#F4B8D4"), // rose
			canvas.Hex("#E8D5A3"), // gold
			canvas.Hex("#B8D4F4"), // celeste
			canvas.Hex("#C4B5F4"), // back to lavender
		},
	}
}

func (n *Nouveau) Name() string { return "nouveau" }

func (n *Nouveau) Step(frame uint64, dt float64) {
	n.time += dt * n.params["speed"]
}

func (n *Nouveau) Render(c *canvas.Canvas, x, y, w, h int) {
	// Border runes
	corners := [4]rune{'╭', '╮', '╯', '╰'}
	horiz := '─'
	vert := '│'

	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			// Only draw border cells (edges of the rectangle)
			if px > x && px < x+w-1 && py > y && py < y+h-1 {
				continue
			}

			t := perimeterPos(px, py, x, y, w, h)
			// Offset by time for cycling
			ct := math.Mod(t+n.time, 1.0)
			if ct < 0 {
				ct += 1.0
			}
			col := n.pal.Sample(ct)

			var r rune
			switch {
			case px == x && py == y:
				r = corners[0] // top-left
			case px == x+w-1 && py == y:
				r = corners[1] // top-right
			case px == x+w-1 && py == y+h-1:
				r = corners[2] // bottom-right
			case px == x && py == y+h-1:
				r = corners[3] // bottom-left
			case py == y || py == y+h-1:
				r = horiz
			default:
				r = vert
			}

			c.SetCell(px, py, canvas.Cell{
				Rune: r,
				FG:   col,
				BG:   canvas.Transparent,
			})
		}
	}
}

func (n *Nouveau) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, ok := n.params[k]; ok {
			n.params[k] = v
		}
	}
}
