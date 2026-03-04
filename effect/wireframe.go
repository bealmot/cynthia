package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Wireframe renders rotating 3D wireframe shapes using perspective projection.
// Shapes: 0=cube, 1=octahedron, 2=icosahedron.
type Wireframe struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("wireframe", func() Effect { return NewWireframe() })
}

func NewWireframe() *Wireframe {
	return &Wireframe{
		params: map[string]float64{
			"speed": 0.5,
			"shape": 0,
			"fov":   3.0,
		},
	}
}

func (w *Wireframe) Name() string { return "wireframe" }

func (w *Wireframe) Step(frame uint64, dt float64) {
	w.time += dt * w.params["speed"]
}

type wfVert struct{ x, y, z float64 }
type wfEdge struct{ a, b int }

func wfCube() ([]wfVert, []wfEdge) {
	v := []wfVert{
		{-1, -1, -1}, {1, -1, -1}, {1, 1, -1}, {-1, 1, -1},
		{-1, -1, 1}, {1, -1, 1}, {1, 1, 1}, {-1, 1, 1},
	}
	e := []wfEdge{
		{0, 1}, {1, 2}, {2, 3}, {3, 0},
		{4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}
	return v, e
}

func wfOctahedron() ([]wfVert, []wfEdge) {
	v := []wfVert{
		{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
	}
	e := []wfEdge{
		{0, 2}, {0, 3}, {0, 4}, {0, 5},
		{1, 2}, {1, 3}, {1, 4}, {1, 5},
		{2, 4}, {2, 5}, {3, 4}, {3, 5},
	}
	return v, e
}

func wfIcosahedron() ([]wfVert, []wfEdge) {
	phi := (1 + math.Sqrt(5)) / 2
	v := []wfVert{
		{-1, phi, 0}, {1, phi, 0}, {-1, -phi, 0}, {1, -phi, 0},
		{0, -1, phi}, {0, 1, phi}, {0, -1, -phi}, {0, 1, -phi},
		{phi, 0, -1}, {phi, 0, 1}, {-phi, 0, -1}, {-phi, 0, 1},
	}
	e := []wfEdge{
		{0, 11}, {0, 5}, {0, 1}, {0, 7}, {0, 10},
		{1, 5}, {1, 9}, {1, 8}, {1, 7},
		{2, 4}, {2, 3}, {2, 6}, {2, 10}, {2, 11},
		{3, 4}, {3, 9}, {3, 8}, {3, 6},
		{4, 5}, {4, 9}, {4, 11},
		{5, 9}, {5, 11},
		{6, 7}, {6, 8}, {6, 10},
		{7, 8}, {7, 10},
		{8, 9}, {10, 11},
	}
	return v, e
}

func (w *Wireframe) rotateY(v wfVert, angle float64) wfVert {
	c, s := math.Cos(angle), math.Sin(angle)
	return wfVert{v.x*c + v.z*s, v.y, -v.x*s + v.z*c}
}

func (w *Wireframe) rotateX(v wfVert, angle float64) wfVert {
	c, s := math.Cos(angle), math.Sin(angle)
	return wfVert{v.x, v.y*c - v.z*s, v.y*s + v.z*c}
}

func (w *Wireframe) Render(c *canvas.Canvas) {
	shape := int(w.params["shape"])
	fov := w.params["fov"]
	t := w.time

	var verts []wfVert
	var edges []wfEdge
	switch shape {
	case 1:
		verts, edges = wfOctahedron()
	case 2:
		verts, edges = wfIcosahedron()
	default:
		verts, edges = wfCube()
	}

	bg := canvas.Hex("#06060F")
	lineColor := canvas.Hex("#00FFAA")
	vertColor := canvas.Hex("#FFFFFF")

	c.Clear(bg)

	fw := float64(c.Width)
	fh := float64(c.Height)

	// Transform and project vertices.
	type proj struct{ x, y int }
	projected := make([]proj, len(verts))
	for i, v := range verts {
		// Normalize icosahedron vertices.
		if shape == 2 {
			l := math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
			v.x /= l
			v.y /= l
			v.z /= l
		}

		v = w.rotateY(v, t)
		v = w.rotateX(v, t*0.7)

		// Perspective projection.
		z := v.z + fov
		if z < 0.1 {
			z = 0.1
		}
		px := int(v.x/z*fw*0.3 + fw/2)
		py := int(v.y/z*fh*0.3 + fh/2)
		projected[i] = proj{px, py}
	}

	// Draw edges.
	for _, e := range edges {
		a := projected[e.a]
		b := projected[e.b]
		w.drawLine(c, a.x, a.y, b.x, b.y, lineColor)
	}

	// Draw vertices as bright dots.
	for _, p := range projected {
		if p.x >= 0 && p.x < c.Width && p.y >= 0 && p.y < c.Height {
			c.SetPixel(p.x, p.y, vertColor)
		}
	}
}

func (w *Wireframe) drawLine(c *canvas.Canvas, x1, y1, x2, y2 int, col canvas.Color) {
	dx := x2 - x1
	dy := y2 - y1
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		if x1 >= 0 && x1 < c.Width && y1 >= 0 && y1 < c.Height {
			c.SetPixel(x1, y1, col)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (w *Wireframe) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := w.params[k]; exists {
			w.params[k] = v
		}
	}
}

func (w *Wireframe) Params() map[string]float64 {
	out := make(map[string]float64, len(w.params))
	for k, v := range w.params {
		out[k] = v
	}
	return out
}
