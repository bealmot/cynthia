// Package compose provides layer compositing with z-ordering and blend modes.
package compose

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// BlendMode determines how two layers are combined.
type BlendMode int

const (
	BlendNormal   BlendMode = iota // Porter-Duff src-over
	BlendScreen                    // 1 - (1-src)(1-dst) — brightens
	BlendAdditive                  // src + dst, clamped
	BlendMultiply                  // src * dst — darkens
)

// BlendPixel blends src over dst using the given mode.
// Both colors are premultiplied alpha.
func BlendPixel(src, dst canvas.Color, mode BlendMode) canvas.Color {
	switch mode {
	case BlendNormal:
		return src.Over(dst)
	case BlendScreen:
		return screenBlend(src, dst)
	case BlendAdditive:
		return additiveBlend(src, dst)
	case BlendMultiply:
		return multiplyBlend(src, dst)
	}
	return src.Over(dst)
}

func screenBlend(src, dst canvas.Color) canvas.Color {
	// Screen in premultiplied: result = src + dst - src*dst
	return canvas.Color{
		R: clamp01(src.R + dst.R - src.R*dst.R),
		G: clamp01(src.G + dst.G - src.G*dst.G),
		B: clamp01(src.B + dst.B - src.B*dst.B),
		A: clamp01(src.A + dst.A - src.A*dst.A),
	}
}

func additiveBlend(src, dst canvas.Color) canvas.Color {
	return canvas.Color{
		R: clamp01(src.R + dst.R),
		G: clamp01(src.G + dst.G),
		B: clamp01(src.B + dst.B),
		A: clamp01(src.A + dst.A),
	}
}

func multiplyBlend(src, dst canvas.Color) canvas.Color {
	// Multiply: result = src * dst (for premultiplied, this just works)
	return canvas.Color{
		R: src.R * dst.R,
		G: src.G * dst.G,
		B: src.B * dst.B,
		A: clamp01(src.A + dst.A*(1-src.A)),
	}
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
