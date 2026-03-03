// Package canvas provides a terminal pixel buffer with rasterization to Unicode
// half-block, braille, and ASCII density modes. Zero Charm dependencies.
package canvas

import (
	"fmt"
	"math"
)

// Color represents an RGBA color with float64 channels in [0,1].
// Stored in premultiplied alpha form for efficient compositing.
type Color struct {
	R, G, B, A float64
}

// RGBA constructors

func RGB(r, g, b float64) Color {
	return Color{r, g, b, 1}
}

func RGBA(r, g, b, a float64) Color {
	return Color{r * a, g * a, b * a, a}
}

// Hex parses a "#RRGGBB" or "#RGB" hex string into a Color.
func Hex(s string) Color {
	if len(s) == 0 {
		return Color{}
	}
	if s[0] == '#' {
		s = s[1:]
	}
	var r, g, b uint8
	switch len(s) {
	case 6:
		fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
	case 3:
		fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		r, g, b = r*17, g*17, b*17
	default:
		return Color{}
	}
	return RGB(float64(r)/255, float64(g)/255, float64(b)/255)
}

// RGB8 creates a Color from 0-255 integer channels.
func RGB8(r, g, b uint8) Color {
	return RGB(float64(r)/255, float64(g)/255, float64(b)/255)
}

// Straight returns the non-premultiplied RGBA values.
func (c Color) Straight() (r, g, b, a float64) {
	if c.A <= 0 {
		return 0, 0, 0, 0
	}
	return c.R / c.A, c.G / c.A, c.B / c.A, c.A
}

// ToRGB8 converts to 0-255 integer values (un-premultiplies first).
func (c Color) ToRGB8() (r, g, b uint8) {
	sr, sg, sb, _ := c.Straight()
	return uint8(clamp01(sr) * 255), uint8(clamp01(sg) * 255), uint8(clamp01(sb) * 255)
}

// Lerp linearly interpolates between two premultiplied colors.
func (c Color) Lerp(other Color, t float64) Color {
	return Color{
		R: c.R + (other.R-c.R)*t,
		G: c.G + (other.G-c.G)*t,
		B: c.B + (other.B-c.B)*t,
		A: c.A + (other.A-c.A)*t,
	}
}

// Over composites src over dst using premultiplied Porter-Duff.
func (src Color) Over(dst Color) Color {
	invA := 1 - src.A
	return Color{
		R: src.R + dst.R*invA,
		G: src.G + dst.G*invA,
		B: src.B + dst.B*invA,
		A: src.A + dst.A*invA,
	}
}

// Transparent is a fully transparent color.
var Transparent = Color{0, 0, 0, 0}

// Black, White for convenience.
var (
	Black = RGB(0, 0, 0)
	White = RGB(1, 1, 1)
)

// Palette is an ordered list of colors for gradient sampling.
type Palette []Color

// Sample returns the interpolated color at position t ∈ [0,1] along the palette.
func (p Palette) Sample(t float64) Color {
	if len(p) == 0 {
		return Transparent
	}
	if len(p) == 1 {
		return p[0]
	}
	t = clamp01(t)
	scaled := t * float64(len(p)-1)
	idx := int(scaled)
	frac := scaled - float64(idx)
	if idx >= len(p)-1 {
		return p[len(p)-1]
	}
	return p[idx].Lerp(p[idx+1], frac)
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
