package compose

import (
	"sort"

	"github.com/bealmot/cynthia/canvas"
)

// Compositor manages a stack of layers and composites them into a final output.
type Compositor struct {
	layers []*Layer
	output *canvas.Canvas
}

// NewCompositor creates a compositor for the given terminal dimensions.
func NewCompositor(cols, rows int, mode canvas.RenderMode) *Compositor {
	return &Compositor{
		output: canvas.New(cols, rows, mode),
	}
}

// AddLayer adds a layer to the compositor.
func (c *Compositor) AddLayer(l *Layer) {
	c.layers = append(c.layers, l)
}

// GetLayer returns a layer by name, or nil.
func (c *Compositor) GetLayer(name string) *Layer {
	for _, l := range c.layers {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// RemoveLayer removes a layer by name.
func (c *Compositor) RemoveLayer(name string) {
	for i, l := range c.layers {
		if l.Name == name {
			c.layers = append(c.layers[:i], c.layers[i+1:]...)
			return
		}
	}
}

// Output returns the composited canvas.
func (c *Compositor) Output() *canvas.Canvas { return c.output }

// Resize changes the output dimensions.
func (c *Compositor) Resize(cols, rows int) {
	c.output.Resize(cols, rows)
}

// Compose flattens all visible layers (sorted by z-order) into the output canvas.
func (c *Compositor) Compose() {
	c.output.Clear(canvas.Transparent)

	// Sort layers by z-order (lowest first = painted first)
	sorted := make([]*Layer, len(c.layers))
	copy(sorted, c.layers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Z < sorted[j].Z
	})

	for _, layer := range sorted {
		if !layer.Visible || layer.Opacity <= 0 {
			continue
		}

		src := layer.Canvas
		// Blend each pixel from this layer onto the output
		// Layers may have different pixel dimensions; use the smaller extent
		maxX := min(c.output.Width, src.Width)
		maxY := min(c.output.Height, src.Height)

		for y := 0; y < maxY; y++ {
			for x := 0; x < maxX; x++ {
				srcPx := src.GetPixel(x, y)

				// Apply layer opacity
				if layer.Opacity < 1.0 {
					srcPx.R *= layer.Opacity
					srcPx.G *= layer.Opacity
					srcPx.B *= layer.Opacity
					srcPx.A *= layer.Opacity
				}

				// Skip fully transparent pixels
				if srcPx.A <= 0 {
					continue
				}

				dstPx := c.output.GetPixel(x, y)
				c.output.SetPixel(x, y, BlendPixel(srcPx, dstPx, layer.Blend))
			}
		}
	}

	// Also composite cell grids for layers that have cell-level content
	// (borders, text overlays). Cells with transparency are skipped.
	for _, layer := range sorted {
		if !layer.Visible || layer.Opacity <= 0 {
			continue
		}

		src := layer.Canvas
		maxCol := min(c.output.CellW, src.CellW)
		maxRow := min(c.output.CellH, src.CellH)

		for row := 0; row < maxRow; row++ {
			for col := 0; col < maxCol; col++ {
				cell := src.GetCell(col, row)
				if cell.IsTransparent() {
					continue
				}
				c.output.SetCell(col, row, cell)
			}
		}
	}
}
