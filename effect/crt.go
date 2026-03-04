package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// CRT simulates a retro cathode-ray tube display — scanlines, phosphor glow,
// slight barrel distortion, and color fringing. Designed as a standalone
// effect that generates its own animated content with CRT post-processing.
type CRT struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("crt", func() Effect { return NewCRT() })
}

func NewCRT() *CRT {
	return &CRT{
		params: map[string]float64{
			"speed":      1.0,
			"scanlines":  1.0,
			"curvature":  0.3,
			"flicker":    0.05,
			"chromatic":  1.0,
		},
	}
}

func (cr *CRT) Name() string { return "crt" }

func (cr *CRT) Step(frame uint64, dt float64) {
	cr.time += dt * cr.params["speed"]
}

func (cr *CRT) Render(c *canvas.Canvas) {
	scanlineStrength := cr.params["scanlines"]
	curvature := cr.params["curvature"]
	flicker := cr.params["flicker"]
	chromatic := cr.params["chromatic"]

	fw := float64(c.Width)
	fh := float64(c.Height)

	// Global brightness flicker.
	globalBright := 1.0 - flicker*math.Sin(cr.time*60)*math.Sin(cr.time*7.3)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Barrel distortion: warp UVs.
			uvx := fx*2 - 1
			uvy := fy*2 - 1
			r2 := uvx*uvx + uvy*uvy
			distort := 1.0 + curvature*r2

			sx := (uvx*distort + 1) / 2
			sy := (uvy*distort + 1) / 2

			// Out of bounds = black (CRT edge).
			if sx < 0 || sx > 1 || sy < 0 || sy > 1 {
				c.SetPixel(x, y, canvas.Black)
				continue
			}

			// Generate base content — animated color pattern.
			t := cr.time
			baseR := math.Sin(sx*math.Pi*4+t)*0.3 + math.Sin(sy*math.Pi*3+t*1.3)*0.2 + 0.5
			baseG := math.Sin(sx*math.Pi*3+t*0.7)*0.3 + math.Sin(sy*math.Pi*5+t*0.9)*0.2 + 0.4
			baseB := math.Sin(sx*math.Pi*5+t*1.1)*0.3 + math.Sin(sy*math.Pi*2+t*1.5)*0.2 + 0.5

			// Chromatic aberration: offset R and B channels slightly.
			if chromatic > 0 {
				offset := 0.003 * chromatic
				rSx := sx + offset
				bSx := sx - offset
				baseR = math.Sin(rSx*math.Pi*4+t)*0.3 + math.Sin(sy*math.Pi*3+t*1.3)*0.2 + 0.5
				baseB = math.Sin(bSx*math.Pi*5+t*1.1)*0.3 + math.Sin(sy*math.Pi*2+t*1.5)*0.2 + 0.5
			}

			// Scanline darkening.
			scanline := 1.0
			if scanlineStrength > 0 {
				scanPhase := math.Mod(fy*fh, 2.0)
				if scanPhase < 1.0 {
					scanline = 1.0 - scanlineStrength*0.4
				}
			}

			// Phosphor mask: RGB sub-pixel pattern.
			subPixel := 1.0
			xMod := x % 3
			switch xMod {
			case 0:
				baseG *= 0.85
				baseB *= 0.7
			case 1:
				baseR *= 0.85
				baseB *= 0.7
			case 2:
				baseR *= 0.7
				baseG *= 0.85
			}

			// Vignette: darken edges.
			vignette := 1.0 - r2*0.5
			if vignette < 0 {
				vignette = 0
			}

			brightness := globalBright * scanline * subPixel * vignette

			col := canvas.RGB(
				clampF(baseR*brightness),
				clampF(baseG*brightness),
				clampF(baseB*brightness),
			)
			c.SetPixel(x, y, col)
		}
	}
}

// RenderPost applies CRT post-processing to existing canvas pixels.
// Unlike Render, it reads the existing pixel buffer instead of generating content.
func (cr *CRT) RenderPost(c *canvas.Canvas) {
	scanlineStrength := cr.params["scanlines"]
	curvature := cr.params["curvature"]
	flicker := cr.params["flicker"]
	chromatic := cr.params["chromatic"]

	fw := float64(c.Width)
	fh := float64(c.Height)

	// Snapshot existing pixels so we can read from them while writing.
	src := make([]canvas.Color, len(c.Pixels))
	copy(src, c.Pixels)

	globalBright := 1.0 - flicker*math.Sin(cr.time*60)*math.Sin(cr.time*7.3)

	readSrc := func(x, y int) canvas.Color {
		if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
			return canvas.Black
		}
		return src[y*c.Width+x]
	}

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh
		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Barrel distortion.
			uvx := fx*2 - 1
			uvy := fy*2 - 1
			r2 := uvx*uvx + uvy*uvy
			distort := 1.0 + curvature*r2

			sx := (uvx*distort + 1) / 2
			sy := (uvy*distort + 1) / 2

			if sx < 0 || sx > 1 || sy < 0 || sy > 1 {
				c.SetPixel(x, y, canvas.Black)
				continue
			}

			// Sample from source pixels.
			srcX := int(sx * fw)
			srcY := int(sy * fh)
			px := readSrc(srcX, srcY)
			baseR, baseG, baseB, _ := px.Straight()

			// Chromatic aberration.
			if chromatic > 0 {
				offset := int(chromatic * 2)
				rPx := readSrc(srcX+offset, srcY)
				bPx := readSrc(srcX-offset, srcY)
				baseR, _, _, _ = rPx.Straight()
				_, _, baseB, _ = bPx.Straight()
			}

			// Scanline darkening.
			scanline := 1.0
			if scanlineStrength > 0 {
				scanPhase := math.Mod(fy*fh, 2.0)
				if scanPhase < 1.0 {
					scanline = 1.0 - scanlineStrength*0.4
				}
			}

			// Phosphor mask.
			xMod := x % 3
			switch xMod {
			case 0:
				baseG *= 0.85
				baseB *= 0.7
			case 1:
				baseR *= 0.85
				baseB *= 0.7
			case 2:
				baseR *= 0.7
				baseG *= 0.85
			}

			// Vignette.
			vignette := 1.0 - r2*0.5
			if vignette < 0 {
				vignette = 0
			}

			brightness := globalBright * scanline * vignette
			c.SetPixel(x, y, canvas.RGB(
				clampF(baseR*brightness),
				clampF(baseG*brightness),
				clampF(baseB*brightness),
			))
		}
	}
}

func clampF(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (cr *CRT) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := cr.params[k]; exists {
			cr.params[k] = v
		}
	}
}

func (cr *CRT) Params() map[string]float64 {
	out := make(map[string]float64, len(cr.params))
	for k, v := range cr.params {
		out[k] = v
	}
	return out
}
