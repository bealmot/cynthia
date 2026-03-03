// Gradient demo — renders a horizontal rainbow gradient to stdout.
// Verifies Phase 1 canvas foundation works end-to-end.
package main

import (
	"fmt"
	"os"

	"github.com/bealmot/cynthia/canvas"
)

func main() {
	cols := 80
	rows := 12

	c := canvas.New(cols, rows, canvas.ModeHalfBlock)

	// Rainbow palette
	rainbow := canvas.Palette{
		canvas.Hex("#FF0000"), // red
		canvas.Hex("#FF8000"), // orange
		canvas.Hex("#FFFF00"), // yellow
		canvas.Hex("#00FF00"), // green
		canvas.Hex("#0080FF"), // blue
		canvas.Hex("#8000FF"), // purple
		canvas.Hex("#FF0080"), // pink
	}

	// Fill pixel buffer with horizontal gradient
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			t := float64(x) / float64(c.Width-1)
			c.SetPixel(x, y, rainbow.Sample(t))
		}
	}

	c.Rasterize()

	// Render to string
	out := canvas.RenderString(c, canvas.ProfileTrueColor)
	fmt.Fprint(os.Stdout, out)
	fmt.Println()
}
