package effect

import (
	"math"
	"math/rand"

	"github.com/bealmot/cynthia/canvas"
)

// Lightning renders branching electrical arcs using midpoint displacement.
// Bolts fire periodically from top to bottom with random branching and
// a bloom/glow effect around the main arc.
type Lightning struct {
	time    float64
	bolts   []bolt
	cooldown float64
	w, h    int
	glow    []float64 // glow buffer for bloom effect
	params  map[string]float64
	rng     *rand.Rand
}

type bolt struct {
	segments []boltSeg
	life     float64
	maxLife  float64
}

type boltSeg struct {
	x1, y1, x2, y2 float64
	brightness      float64
}

func init() {
	Register("lightning", func() Effect { return NewLightning() })
}

func NewLightning() *Lightning {
	return &Lightning{
		params: map[string]float64{
			"speed":     1.0,
			"rate":      0.5,
			"branches":  3,
			"jitter":    0.15,
		},
		rng: rand.New(rand.NewSource(42)),
	}
}

func (l *Lightning) Name() string { return "lightning" }

func (l *Lightning) ensureSize(w, h int) {
	if l.w == w && l.h == h {
		return
	}
	l.w, l.h = w, h
	l.glow = make([]float64, w*h)
}

func (l *Lightning) generateBolt() bolt {
	// Start position near top center with some randomness.
	startX := 0.3 + l.rng.Float64()*0.4
	endX := 0.2 + l.rng.Float64()*0.6
	jitter := l.params["jitter"]
	maxBranches := int(l.params["branches"])

	var segs []boltSeg
	// Midpoint displacement for main bolt.
	segs = l.subdivide(startX, 0, endX, 1.0, jitter, 5, 1.0, segs)

	// Add branches.
	for b := 0; b < maxBranches; b++ {
		if len(segs) == 0 {
			break
		}
		// Pick a random segment to branch from.
		idx := l.rng.Intn(len(segs))
		seg := segs[idx]
		midX := (seg.x1 + seg.x2) / 2
		midY := (seg.y1 + seg.y2) / 2
		branchEndX := midX + (l.rng.Float64()-0.5)*0.4
		branchEndY := midY + 0.1 + l.rng.Float64()*0.3
		if branchEndY > 1.0 {
			branchEndY = 1.0
		}
		segs = l.subdivide(midX, midY, branchEndX, branchEndY, jitter*0.7, 3, 0.5, segs)
	}

	return bolt{
		segments: segs,
		life:     0,
		maxLife:  0.3 + l.rng.Float64()*0.2,
	}
}

func (l *Lightning) subdivide(x1, y1, x2, y2, jitter float64, depth int, brightness float64, segs []boltSeg) []boltSeg {
	if depth <= 0 {
		segs = append(segs, boltSeg{x1, y1, x2, y2, brightness})
		return segs
	}
	midX := (x1+x2)/2 + (l.rng.Float64()-0.5)*jitter
	midY := (y1 + y2) / 2
	segs = l.subdivide(x1, y1, midX, midY, jitter*0.65, depth-1, brightness, segs)
	segs = l.subdivide(midX, midY, x2, y2, jitter*0.65, depth-1, brightness, segs)
	return segs
}

func (l *Lightning) Step(frame uint64, dt float64) {
	speed := l.params["speed"]
	rate := l.params["rate"]
	sdt := dt * speed

	// Decay glow buffer.
	for i := range l.glow {
		l.glow[i] *= 0.85
	}

	// Update existing bolts.
	alive := l.bolts[:0]
	for i := range l.bolts {
		l.bolts[i].life += sdt
		if l.bolts[i].life < l.bolts[i].maxLife {
			alive = append(alive, l.bolts[i])
		}
	}
	l.bolts = alive

	// Maybe spawn new bolt.
	l.cooldown -= sdt
	if l.cooldown <= 0 && l.w > 0 {
		l.bolts = append(l.bolts, l.generateBolt())
		l.cooldown = 1.0/rate + l.rng.Float64()*0.5
	}

	l.time += sdt
}

func (l *Lightning) Render(c *canvas.Canvas) {
	l.ensureSize(c.Width, c.Height)

	bg := canvas.Hex("#050510")
	boltColor := canvas.RGB(0.7, 0.8, 1.0)

	fw := float64(c.Width)
	fh := float64(c.Height)

	// Rasterize bolts into glow buffer.
	for _, b := range l.bolts {
		fade := 1.0 - b.life/b.maxLife
		for _, seg := range b.segments {
			l.drawLine(
				int(seg.x1*fw), int(seg.y1*fh),
				int(seg.x2*fw), int(seg.y2*fh),
				seg.brightness*fade,
			)
		}
	}

	// Render glow buffer to canvas.
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			g := l.glow[y*l.w+x]
			if g < 0.01 {
				c.SetPixel(x, y, bg)
				continue
			}
			if g > 1.0 {
				g = 1.0
			}
			col := canvas.RGB(
				bg.R+boltColor.R*g,
				bg.G+boltColor.G*g,
				bg.B+boltColor.B*g,
			)
			c.SetPixel(x, y, col)
		}
	}
}

// drawLine rasterizes a line into the glow buffer using Bresenham's.
func (l *Lightning) drawLine(x1, y1, x2, y2 int, brightness float64) {
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
		// Draw pixel with bloom (3x3 kernel).
		for by := -1; by <= 1; by++ {
			for bx := -1; bx <= 1; bx++ {
				px, py := x1+bx, y1+by
				if px >= 0 && px < l.w && py >= 0 && py < l.h {
					dist := math.Sqrt(float64(bx*bx + by*by))
					falloff := brightness * math.Max(0, 1.0-dist*0.5)
					idx := py*l.w + px
					l.glow[idx] += falloff
				}
			}
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

func (l *Lightning) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := l.params[k]; exists {
			l.params[k] = v
		}
	}
}

func (l *Lightning) Params() map[string]float64 {
	out := make(map[string]float64, len(l.params))
	for k, v := range l.params {
		out[k] = v
	}
	return out
}
