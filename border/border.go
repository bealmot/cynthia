// Package border provides animated border/chrome effects for terminal panels.
package border

import "github.com/bealmot/cynthia/canvas"

// Border is an animated border effect that draws around a rectangle.
type Border interface {
	Step(frame uint64, dt float64)
	Render(c *canvas.Canvas, x, y, w, h int) // draws border in cell coords
	SetParams(params map[string]float64)
	Name() string
}

// perimeterPos maps a cell position on the border perimeter to a [0,1] value.
// Goes clockwise: top → right → bottom → left.
func perimeterPos(px, py, x, y, w, h int) float64 {
	perimeter := 2*(w-1) + 2*(h-1)
	if perimeter <= 0 {
		return 0
	}

	var pos int
	switch {
	case py == y: // top edge
		pos = px - x
	case px == x+w-1: // right edge
		pos = (w - 1) + (py - y)
	case py == y+h-1: // bottom edge (reversed)
		pos = (w - 1) + (h - 1) + (x + w - 1 - px)
	case px == x: // left edge (reversed)
		pos = 2*(w-1) + (h - 1) + (y + h - 1 - py)
	}

	return float64(pos) / float64(perimeter)
}
