package compose

import (
	"sort"

	"github.com/bealmot/cynthia/canvas"
)

// ScenePanel is the interface a scene graph panel must satisfy for compositing.
// Defined here (in compose) to avoid an import cycle with the scene package.
type ScenePanel interface {
	GetID() string
	GetCanvas() *canvas.Canvas
	GetCellPosition() (x, y int) // top-left corner in cell coordinates
	GetZ() int
	GetOpacity() float64
	GetBlend() BlendMode
	IsVisible() bool
}

// ComposeScene composites a set of panels into the output canvas.
// Panels are positioned at their cell coordinates and clipped to output bounds.
// The output is cleared before compositing.
func (c *Compositor) ComposeScene(panels []ScenePanel) {
	c.output.Clear(canvas.Transparent)

	sorted := make([]ScenePanel, len(panels))
	copy(sorted, panels)
	sort.Slice(sorted, func(i, j int) bool {
		zi, zj := sorted[i].GetZ(), sorted[j].GetZ()
		if zi != zj {
			return zi < zj
		}
		return sorted[i].GetID() < sorted[j].GetID()
	})

	out := c.output

	// Pass 1: pixel blending with position offset
	for _, p := range sorted {
		if !p.IsVisible() || p.GetOpacity() <= 0 {
			continue
		}

		src := p.GetCanvas()
		opacity := p.GetOpacity()
		blend := p.GetBlend()
		cx, cy := p.GetCellPosition()

		// Convert cell offset to pixel offset based on render mode.
		// HalfBlock: 1 col = 1px wide, 1 row = 2px tall
		// Braille:   1 col = 2px wide, 1 row = 4px tall
		// ASCII:     1 col = 1px wide, 1 row = 1px tall
		pxOffX, pxOffY := cellToPixelOffset(cx, cy, out.Mode)

		for y := 0; y < src.Height; y++ {
			outY := y + pxOffY
			if outY < 0 || outY >= out.Height {
				continue
			}
			for x := 0; x < src.Width; x++ {
				outX := x + pxOffX
				if outX < 0 || outX >= out.Width {
					continue
				}

				srcPx := src.GetPixel(x, y)
				if opacity < 1.0 {
					srcPx.R *= opacity
					srcPx.G *= opacity
					srcPx.B *= opacity
					srcPx.A *= opacity
				}
				if srcPx.A <= 0 {
					continue
				}

				dstPx := out.GetPixel(outX, outY)
				out.SetPixel(outX, outY, BlendPixel(srcPx, dstPx, blend))
			}
		}
	}

	// Pass 2: cell grid overlay (text, borders) with position offset
	for _, p := range sorted {
		if !p.IsVisible() || p.GetOpacity() <= 0 {
			continue
		}

		src := p.GetCanvas()
		cx, cy := p.GetCellPosition()

		for row := 0; row < src.CellH; row++ {
			outRow := row + cy
			if outRow < 0 || outRow >= out.CellH {
				continue
			}
			for col := 0; col < src.CellW; col++ {
				outCol := col + cx
				if outCol < 0 || outCol >= out.CellW {
					continue
				}

				cell := src.GetCell(col, row)
				if cell.IsTransparent() {
					continue
				}
				out.SetCell(outCol, outRow, cell)
			}
		}
	}
}

// cellToPixelOffset converts cell coordinates to pixel coordinates for the given mode.
func cellToPixelOffset(cellX, cellY int, mode canvas.RenderMode) (px, py int) {
	switch mode {
	case canvas.ModeHalfBlock:
		return cellX, cellY * 2
	case canvas.ModeBraille:
		return cellX * 2, cellY * 4
	case canvas.ModeASCII:
		return cellX, cellY
	}
	return cellX, cellY * 2 // default to HalfBlock
}
