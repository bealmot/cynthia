package compose

import "github.com/bealmot/cynthia/canvas"

// Layer is a canvas with compositing attributes.
type Layer struct {
	Canvas  *canvas.Canvas
	Z       int       // z-order (lower = further back)
	Opacity float64   // 0 = invisible, 1 = fully opaque
	Blend   BlendMode // how this layer composites over the one below
	Visible bool
	Name    string
}

// NewLayer creates a visible layer at the given z-order.
func NewLayer(name string, z int, cols, rows int, mode canvas.RenderMode) *Layer {
	return &Layer{
		Canvas:  canvas.New(cols, rows, mode),
		Z:       z,
		Opacity: 1.0,
		Blend:   BlendNormal,
		Visible: true,
		Name:    name,
	}
}
