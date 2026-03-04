package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Voronoi renders animated Voronoi cells — colored regions based on
// nearest seed point. Seeds drift smoothly, creating a stained-glass look.
type Voronoi struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("voronoi", func() Effect { return NewVoronoi() })
}

func NewVoronoi() *Voronoi {
	return &Voronoi{
		params: map[string]float64{
			"speed":  0.5,
			"seeds":  12,
			"border": 0.03,
		},
	}
}

func (v *Voronoi) Name() string { return "voronoi" }

func (v *Voronoi) Step(frame uint64, dt float64) {
	v.time += dt * v.params["speed"]
}

func (v *Voronoi) Render(c *canvas.Canvas) {
	seeds := int(v.params["seeds"])
	borderWidth := v.params["border"]
	t := v.time

	if seeds < 2 {
		seeds = 12
	}

	// Compute seed positions — each drifts on its own trajectory.
	type seed struct{ x, y float64 }
	pts := make([]seed, seeds)
	for i := 0; i < seeds; i++ {
		fi := float64(i)
		pts[i] = seed{
			x: 0.5 + 0.4*math.Sin(t*0.4+fi*2.1)*math.Cos(t*0.3+fi*0.7),
			y: 0.5 + 0.4*math.Cos(t*0.35+fi*1.8)*math.Sin(t*0.25+fi*1.3),
		}
	}

	// Cell colors — one per seed.
	pal := canvas.Palette{
		canvas.Hex("#E74C3C"),
		canvas.Hex("#8E44AD"),
		canvas.Hex("#3498DB"),
		canvas.Hex("#1ABC9C"),
		canvas.Hex("#2ECC71"),
		canvas.Hex("#F1C40F"),
		canvas.Hex("#E67E22"),
		canvas.Hex("#E91E63"),
		canvas.Hex("#00BCD4"),
		canvas.Hex("#9C27B0"),
		canvas.Hex("#4CAF50"),
		canvas.Hex("#FF9800"),
	}

	borderColor := canvas.Hex("#1A1A2E")

	fw := float64(c.Width)
	fh := float64(c.Height)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Find nearest and second-nearest seed.
			minDist := math.MaxFloat64
			minDist2 := math.MaxFloat64
			nearest := 0

			for i, p := range pts {
				dx := fx - p.x
				dy := fy - p.y
				d := dx*dx + dy*dy
				if d < minDist {
					minDist2 = minDist
					minDist = d
					nearest = i
				} else if d < minDist2 {
					minDist2 = d
				}
			}

			// Border detection: thin line where two cells are equidistant.
			diff := math.Sqrt(minDist2) - math.Sqrt(minDist)
			if diff < borderWidth {
				c.SetPixel(x, y, borderColor)
				continue
			}

			// Cell color based on seed index.
			cellT := float64(nearest) / float64(seeds)
			col := pal.Sample(cellT)

			// Slight shading by distance from seed center.
			shade := 0.7 + 0.3*(1.0-math.Min(math.Sqrt(minDist)*3, 1.0))
			col.R *= shade
			col.G *= shade
			col.B *= shade

			c.SetPixel(x, y, col)
		}
	}
}

func (v *Voronoi) SetParams(params map[string]float64) {
	for k, val := range params {
		if _, exists := v.params[k]; exists {
			v.params[k] = val
		}
	}
}

func (v *Voronoi) Params() map[string]float64 {
	out := make(map[string]float64, len(v.params))
	for k, val := range v.params {
		out[k] = val
	}
	return out
}
