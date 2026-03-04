package effect

import "math"

// OpenSimplex2S noise implementation (public domain).
// Provides 2D and 3D noise plus fractional Brownian motion (fbm).

const (
	// Skew/unskew constants for 2D simplex grid.
	skew2D   = 0.366025403784439   // (sqrt(3)-1)/2
	unskew2D = 0.211324865405187   // (3-sqrt(3))/6
	// Skew/unskew constants for 3D.
	skew3D   = 1.0 / 3.0
	unskew3D = 1.0 / 6.0
)

// grad2 and grad3 are gradient vectors for noise hashing.
var grad2 = [8][2]float64{
	{1, 1}, {-1, 1}, {1, -1}, {-1, -1},
	{1, 0}, {-1, 0}, {0, 1}, {0, -1},
}

var grad3 = [12][3]float64{
	{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
	{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
	{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
}

// perm is a permutation table for hashing coordinates to gradient indices.
var perm [512]int

func init() {
	// Populate permutation table with a fixed seed.
	var p [256]int
	for i := range p {
		p[i] = i
	}
	// Fisher-Yates shuffle with deterministic seed.
	seed := int64(0)
	for i := 255; i > 0; i-- {
		seed = (seed*6364136223846793005 + 1442695040888963407) & 0x7fffffffffffffff
		j := int(seed % int64(i+1))
		p[i], p[j] = p[j], p[i]
	}
	for i := 0; i < 512; i++ {
		perm[i] = p[i&255]
	}
}

// noise2 evaluates 2D simplex noise at (x, y). Returns value in [-1, 1].
func noise2(x, y float64) float64 {
	// Skew input to simplex cell.
	s := (x + y) * skew2D
	i := math.Floor(x + s)
	j := math.Floor(y + s)

	// Unskew back to (x,y) space.
	t := (i + j) * unskew2D
	x0 := x - (i - t)
	y0 := y - (j - t)

	// Determine which simplex triangle we're in.
	var i1, j1 int
	if x0 > y0 {
		i1, j1 = 1, 0
	} else {
		i1, j1 = 0, 1
	}

	// Offsets for middle and far corners.
	x1 := x0 - float64(i1) + unskew2D
	y1 := y0 - float64(j1) + unskew2D
	x2 := x0 - 1.0 + 2.0*unskew2D
	y2 := y0 - 1.0 + 2.0*unskew2D

	// Hash coordinates to gradient indices.
	ii := int(i) & 255
	jj := int(j) & 255

	// Calculate contributions from the three corners.
	var n0, n1, n2 float64

	t0 := 0.5 - x0*x0 - y0*y0
	if t0 > 0 {
		t0 *= t0
		gi := perm[ii+perm[jj]] & 7
		n0 = t0 * t0 * (grad2[gi][0]*x0 + grad2[gi][1]*y0)
	}

	t1 := 0.5 - x1*x1 - y1*y1
	if t1 > 0 {
		t1 *= t1
		gi := perm[ii+i1+perm[jj+j1]] & 7
		n1 = t1 * t1 * (grad2[gi][0]*x1 + grad2[gi][1]*y1)
	}

	t2 := 0.5 - x2*x2 - y2*y2
	if t2 > 0 {
		t2 *= t2
		gi := perm[ii+1+perm[jj+1]] & 7
		n2 = t2 * t2 * (grad2[gi][0]*x2 + grad2[gi][1]*y2)
	}

	// Scale to [-1, 1].
	return 70.0 * (n0 + n1 + n2)
}

// noise3 evaluates 3D simplex noise at (x, y, z). Returns value in [-1, 1].
func noise3(x, y, z float64) float64 {
	// Skew to simplex cell.
	s := (x + y + z) * skew3D
	i := math.Floor(x + s)
	j := math.Floor(y + s)
	k := math.Floor(z + s)

	t := (i + j + k) * unskew3D
	x0 := x - (i - t)
	y0 := y - (j - t)
	z0 := z - (k - t)

	// Determine simplex traversal order.
	var i1, j1, k1, i2, j2, k2 int
	if x0 >= y0 {
		if y0 >= z0 {
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 1, 0
		} else if x0 >= z0 {
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 0, 1
		} else {
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 1, 0, 1
		}
	} else {
		if y0 < z0 {
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 0, 1, 1
		} else if x0 < z0 {
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 0, 1, 1
		} else {
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 1, 1, 0
		}
	}

	x1 := x0 - float64(i1) + unskew3D
	y1 := y0 - float64(j1) + unskew3D
	z1 := z0 - float64(k1) + unskew3D
	x2 := x0 - float64(i2) + 2.0*unskew3D
	y2 := y0 - float64(j2) + 2.0*unskew3D
	z2 := z0 - float64(k2) + 2.0*unskew3D
	x3 := x0 - 1.0 + 3.0*unskew3D
	y3 := y0 - 1.0 + 3.0*unskew3D
	z3 := z0 - 1.0 + 3.0*unskew3D

	ii := int(i) & 255
	jj := int(j) & 255
	kk := int(k) & 255

	var n0, n1, n2, n3 float64

	t0 := 0.6 - x0*x0 - y0*y0 - z0*z0
	if t0 > 0 {
		t0 *= t0
		gi := perm[ii+perm[jj+perm[kk]]] % 12
		n0 = t0 * t0 * (grad3[gi][0]*x0 + grad3[gi][1]*y0 + grad3[gi][2]*z0)
	}

	t1 := 0.6 - x1*x1 - y1*y1 - z1*z1
	if t1 > 0 {
		t1 *= t1
		gi := perm[ii+i1+perm[jj+j1+perm[kk+k1]]] % 12
		n1 = t1 * t1 * (grad3[gi][0]*x1 + grad3[gi][1]*y1 + grad3[gi][2]*z1)
	}

	t2 := 0.6 - x2*x2 - y2*y2 - z2*z2
	if t2 > 0 {
		t2 *= t2
		gi := perm[ii+i2+perm[jj+j2+perm[kk+k2]]] % 12
		n2 = t2 * t2 * (grad3[gi][0]*x2 + grad3[gi][1]*y2 + grad3[gi][2]*z2)
	}

	t3 := 0.6 - x3*x3 - y3*y3 - z3*z3
	if t3 > 0 {
		t3 *= t3
		gi := perm[ii+1+perm[jj+1+perm[kk+1]]] % 12
		n3 = t3 * t3 * (grad3[gi][0]*x3 + grad3[gi][1]*y3 + grad3[gi][2]*z3)
	}

	return 32.0 * (n0 + n1 + n2 + n3)
}

// fbm computes fractional Brownian motion by layering noise octaves.
// Higher octaves add progressively finer detail.
func fbm(x, y, z float64, octaves int, lacunarity, persistence float64) float64 {
	var sum float64
	amplitude := 1.0
	frequency := 1.0
	maxAmp := 0.0

	for i := 0; i < octaves; i++ {
		sum += amplitude * noise3(x*frequency, y*frequency, z*frequency)
		maxAmp += amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}

	return sum / maxAmp
}

// fbm2 computes 2D fractional Brownian motion.
func fbm2(x, y float64, octaves int, lacunarity, persistence float64) float64 {
	var sum float64
	amplitude := 1.0
	frequency := 1.0
	maxAmp := 0.0

	for i := 0; i < octaves; i++ {
		sum += amplitude * noise2(x*frequency, y*frequency)
		maxAmp += amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}

	return sum / maxAmp
}
