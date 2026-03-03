package canvas

import "math"

// ColorProfile represents the terminal's color capability.
type ColorProfile int

const (
	ProfileTrueColor ColorProfile = iota
	ProfileANSI256
	ProfileANSI16
	ProfileNoColor
)

// Degrade reduces a Color to the nearest representable value for the given profile.
func Degrade(c Color, profile ColorProfile) Color {
	switch profile {
	case ProfileTrueColor:
		return c
	case ProfileANSI256:
		return nearestANSI256(c)
	case ProfileANSI16:
		return nearestANSI16(c)
	case ProfileNoColor:
		return Transparent
	}
	return c
}

// ANSI 256-color cube: 6×6×6 from indices 16-231, then 24 grays at 232-255.
func nearestANSI256(c Color) Color {
	sr, sg, sb, _ := c.Straight()
	r8 := clamp01(sr) * 255
	g8 := clamp01(sg) * 255
	b8 := clamp01(sb) * 255

	// Find nearest in the 6×6×6 cube
	ri := math.Round(r8 / 255 * 5)
	gi := math.Round(g8 / 255 * 5)
	bi := math.Round(b8 / 255 * 5)

	cubeR := ri * 51
	cubeG := gi * 51
	cubeB := bi * 51

	cubeDist := sq(r8-cubeR) + sq(g8-cubeG) + sq(b8-cubeB)

	// Find nearest gray
	gray := (r8 + g8 + b8) / 3
	grayIdx := math.Round((gray - 8) / 10)
	grayIdx = math.Max(0, math.Min(23, grayIdx))
	grayVal := grayIdx*10 + 8
	grayDist := sq(r8-grayVal) + sq(g8-grayVal) + sq(b8-grayVal)

	if grayDist < cubeDist {
		v := grayVal / 255
		return RGBA(v, v, v, c.A)
	}
	return RGBA(cubeR/255, cubeG/255, cubeB/255, c.A)
}

// ANSI 16 basic colors
var ansi16 = [16]Color{
	RGB(0, 0, 0),       // 0 black
	RGB(0.67, 0, 0),    // 1 red
	RGB(0, 0.67, 0),    // 2 green
	RGB(0.67, 0.67, 0), // 3 yellow
	RGB(0, 0, 0.67),    // 4 blue
	RGB(0.67, 0, 0.67), // 5 magenta
	RGB(0, 0.67, 0.67), // 6 cyan
	RGB(0.67, 0.67, 0.67), // 7 white
	RGB(0.33, 0.33, 0.33), // 8 bright black
	RGB(1, 0.33, 0.33),    // 9 bright red
	RGB(0.33, 1, 0.33),    // 10 bright green
	RGB(1, 1, 0.33),       // 11 bright yellow
	RGB(0.33, 0.33, 1),    // 12 bright blue
	RGB(1, 0.33, 1),       // 13 bright magenta
	RGB(0.33, 1, 1),       // 14 bright cyan
	RGB(1, 1, 1),          // 15 bright white
}

func nearestANSI16(c Color) Color {
	sr, sg, sb, _ := c.Straight()
	best := 0
	bestDist := math.MaxFloat64
	for i, ac := range ansi16 {
		d := sq(sr-ac.R) + sq(sg-ac.G) + sq(sb-ac.B)
		if d < bestDist {
			bestDist = d
			best = i
		}
	}
	return RGBA(ansi16[best].R, ansi16[best].G, ansi16[best].B, c.A)
}

func sq(v float64) float64 { return v * v }
