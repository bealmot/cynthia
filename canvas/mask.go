package canvas

import "math"

// MaskDirection specifies the direction for gradient masks.
type MaskDirection int

const (
	MaskDirTopToBottom MaskDirection = iota
	MaskDirLeftToRight
	MaskDirRadial
)

// MaskCircle generates a circular alpha mask with soft edges.
// cx, cy, radius are in normalized [0,1] coordinates.
// feather controls the soft edge width (0 = hard edge).
func MaskCircle(w, h int, cx, cy, radius, feather float64) []float64 {
	mask := make([]float64, w*h)
	for y := 0; y < h; y++ {
		ny := float64(y) / float64(h)
		for x := 0; x < w; x++ {
			nx := float64(x) / float64(w)
			dx := nx - cx
			dy := ny - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			mask[y*w+x] = smoothEdge(dist, radius, feather)
		}
	}
	return mask
}

// MaskRect generates a rectangular alpha mask with optional feathered edges.
// x, y, rw, rh are in normalized [0,1] coordinates.
func MaskRect(w, h int, rx, ry, rw, rh, feather float64) []float64 {
	mask := make([]float64, w*h)
	for py := 0; py < h; py++ {
		ny := float64(py) / float64(h)
		for px := 0; px < w; px++ {
			nx := float64(px) / float64(w)

			// Distance from nearest edge (negative = inside).
			dx := math.Max(rx-nx, nx-(rx+rw))
			dy := math.Max(ry-ny, ny-(ry+rh))
			dist := math.Max(dx, dy)

			if feather > 0 && dist > -feather {
				mask[py*w+px] = clamp01((-dist) / feather)
			} else if dist <= 0 {
				mask[py*w+px] = 1
			}
		}
	}
	return mask
}

// MaskGradient generates a linear or radial gradient mask.
// start and end are normalized positions along the gradient direction.
func MaskGradient(w, h int, dir MaskDirection, start, end float64) []float64 {
	mask := make([]float64, w*h)
	span := end - start
	if span == 0 {
		span = 1
	}

	for y := 0; y < h; y++ {
		ny := float64(y) / float64(h)
		for x := 0; x < w; x++ {
			nx := float64(x) / float64(w)

			var t float64
			switch dir {
			case MaskDirTopToBottom:
				t = (ny - start) / span
			case MaskDirLeftToRight:
				t = (nx - start) / span
			case MaskDirRadial:
				dx := nx - 0.5
				dy := ny - 0.5
				dist := math.Sqrt(dx*dx+dy*dy) * 2 // normalize so corner ≈ 1
				t = (dist - start) / span
			}

			mask[y*w+x] = clamp01(1 - t)
		}
	}
	return mask
}

// MaskVignette generates a vignette (darkened edges) mask.
// strength controls how far the darkening extends inward (0–1).
func MaskVignette(w, h int, strength float64) []float64 {
	mask := make([]float64, w*h)
	for y := 0; y < h; y++ {
		ny := float64(y)/float64(h)*2 - 1 // -1 to 1
		for x := 0; x < w; x++ {
			nx := float64(x)/float64(w)*2 - 1
			r2 := nx*nx + ny*ny
			v := 1.0 - r2*strength
			if v < 0 {
				v = 0
			}
			mask[y*w+x] = v
		}
	}
	return mask
}

// smoothEdge returns 1.0 inside radius, 0.0 outside radius+feather,
// and a smooth transition in between.
func smoothEdge(dist, radius, feather float64) float64 {
	if feather <= 0 {
		if dist <= radius {
			return 1
		}
		return 0
	}
	if dist <= radius {
		return 1
	}
	if dist >= radius+feather {
		return 0
	}
	return 1 - (dist-radius)/feather
}
