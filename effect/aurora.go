package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// Aurora renders shimmering curtains of northern lights.
// Layered sine waves with noise perturbation create flowing vertical
// bands of light with characteristic green-to-purple color gradients.
type Aurora struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("aurora", func() Effect { return NewAurora() })
}

func NewAurora() *Aurora {
	return &Aurora{
		params: map[string]float64{
			"speed":    0.5,
			"layers":   4,
			"wave":     3.0,
			"altitude": 0.4,
		},
	}
}

func (a *Aurora) Name() string { return "aurora" }

func (a *Aurora) Step(frame uint64, dt float64) {
	a.time += dt * a.params["speed"]
}

func (a *Aurora) Render(c *canvas.Canvas) {
	layers := int(a.params["layers"])
	wave := a.params["wave"]
	altitude := a.params["altitude"]
	t := a.time

	if layers < 1 {
		layers = 4
	}

	// Sky gradient: dark blue at top, darker at bottom.
	skyTop := canvas.Hex("#020820")
	skyBot := canvas.Hex("#010410")

	fw := float64(c.Width)
	fh := float64(c.Height)

	for y := 0; y < c.Height; y++ {
		fy := float64(y) / fh
		skyColor := skyTop.Lerp(skyBot, fy)

		for x := 0; x < c.Width; x++ {
			fx := float64(x) / fw

			// Accumulate aurora light from multiple curtain layers.
			var intensity float64
			var hue float64

			for l := 0; l < layers; l++ {
				fl := float64(l)
				// Each layer has a different wave pattern.
				curtainX := fx*wave + math.Sin(fy*3.0+t*0.8+fl*1.5)*0.3
				curtainVal := math.Sin(curtainX*math.Pi*2 + t*(0.5+fl*0.2))

				// Noise modulation for organic movement.
				n := noise2(fx*3.0+fl*10, t*0.3+fl*5)
				curtainVal += n * 0.3

				// Vertical profile: aurora appears at a certain altitude band.
				vertCenter := altitude - fl*0.08
				vertWidth := 0.15 + fl*0.03
				vertProfile := math.Exp(-math.Pow((fy-vertCenter)/vertWidth, 2))

				layerIntensity := math.Max(0, curtainVal) * vertProfile * 0.4
				intensity += layerIntensity
				hue += layerIntensity * (0.3 + fl*0.15) // green → purple shift per layer
			}

			if intensity < 0.01 {
				c.SetPixel(x, y, skyColor)
				continue
			}

			if intensity > 1.0 {
				intensity = 1.0
			}

			// Aurora color: green at base, transitioning to purple/pink at top.
			avgHue := hue / math.Max(intensity, 0.01)
			auroraColor := auroraHue(avgHue)

			// Blend aurora over sky.
			col := canvas.RGB(
				skyColor.R+auroraColor.R*intensity,
				skyColor.G+auroraColor.G*intensity,
				skyColor.B+auroraColor.B*intensity,
			)
			// Clamp.
			if col.R > 1 {
				col.R = 1
			}
			if col.G > 1 {
				col.G = 1
			}
			if col.B > 1 {
				col.B = 1
			}
			c.SetPixel(x, y, col)
		}
	}
}

// auroraHue maps a value to aurora-characteristic colors.
func auroraHue(t float64) canvas.Color {
	pal := canvas.Palette{
		canvas.Hex("#00FF66"), // green (most common)
		canvas.Hex("#00FFAA"),
		canvas.Hex("#33CCFF"), // cyan
		canvas.Hex("#8844FF"), // purple
		canvas.Hex("#FF44AA"), // pink (rare, high energy)
	}
	t = math.Max(0, math.Min(1, t))
	return pal.Sample(t)
}

func (a *Aurora) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := a.params[k]; exists {
			a.params[k] = v
		}
	}
}

func (a *Aurora) Params() map[string]float64 {
	out := make(map[string]float64, len(a.params))
	for k, v := range a.params {
		out[k] = v
	}
	return out
}
