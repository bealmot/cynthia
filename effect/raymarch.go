package effect

import (
	"math"

	"github.com/bealmot/cynthia/canvas"
)

// RayMarch renders 3D scenes using signed distance function ray marching.
// Includes SDF primitives (sphere, torus, cube) with Phong shading.
type RayMarch struct {
	time   float64
	params map[string]float64
}

func init() {
	Register("raymarch", func() Effect { return NewRayMarch() })
}

func NewRayMarch() *RayMarch {
	return &RayMarch{
		params: map[string]float64{
			"speed":       0.5,
			"fov":         60.0,
			"shape":       0, // 0=sphere, 1=torus, 2=cube
			"light_angle": 0.8,
		},
	}
}

func (r *RayMarch) Name() string { return "raymarch" }

func (r *RayMarch) Step(frame uint64, dt float64) {
	r.time += dt * r.params["speed"]
}

// vec3 is a local 3D vector type.
type vec3 struct{ x, y, z float64 }

func v3add(a, b vec3) vec3     { return vec3{a.x + b.x, a.y + b.y, a.z + b.z} }
func v3sub(a, b vec3) vec3     { return vec3{a.x - b.x, a.y - b.y, a.z - b.z} }
func v3scale(a vec3, s float64) vec3 { return vec3{a.x * s, a.y * s, a.z * s} }
func v3dot(a, b vec3) float64  { return a.x*b.x + a.y*b.y + a.z*b.z }
func v3len(a vec3) float64     { return math.Sqrt(v3dot(a, a)) }

func v3norm(a vec3) vec3 {
	l := v3len(a)
	if l < 1e-10 {
		return vec3{}
	}
	return v3scale(a, 1.0/l)
}

func v3abs(a vec3) vec3 {
	return vec3{math.Abs(a.x), math.Abs(a.y), math.Abs(a.z)}
}

func v3max(a vec3, v float64) vec3 {
	return vec3{math.Max(a.x, v), math.Max(a.y, v), math.Max(a.z, v)}
}

// SDF primitives.

func sdfSphere(p vec3, r float64) float64 {
	return v3len(p) - r
}

func sdfTorus(p vec3, major, minor float64) float64 {
	q := math.Sqrt(p.x*p.x+p.z*p.z) - major
	return math.Sqrt(q*q+p.y*p.y) - minor
}

func sdfCube(p vec3, size float64) float64 {
	d := v3sub(v3abs(p), vec3{size, size, size})
	outside := v3len(v3max(d, 0))
	inside := math.Min(math.Max(d.x, math.Max(d.y, d.z)), 0)
	return outside + inside
}

func (r *RayMarch) sceneSDF(p vec3) float64 {
	shape := int(r.params["shape"])
	switch shape {
	case 1:
		return sdfTorus(p, 0.8, 0.3)
	case 2:
		return sdfCube(p, 0.7)
	default:
		return sdfSphere(p, 1.0)
	}
}

func (r *RayMarch) normal(p vec3) vec3 {
	const eps = 0.001
	d := r.sceneSDF(p)
	return v3norm(vec3{
		r.sceneSDF(vec3{p.x + eps, p.y, p.z}) - d,
		r.sceneSDF(vec3{p.x, p.y + eps, p.z}) - d,
		r.sceneSDF(vec3{p.x, p.y, p.z + eps}) - d,
	})
}

func (r *RayMarch) Render(c *canvas.Canvas) {
	fov := r.params["fov"] * math.Pi / 180.0
	lightAngle := r.params["light_angle"]

	// Orbiting camera.
	camDist := 3.0
	camX := camDist * math.Cos(r.time)
	camZ := camDist * math.Sin(r.time)
	camPos := vec3{camX, 1.0, camZ}
	target := vec3{0, 0, 0}

	// Camera basis.
	forward := v3norm(v3sub(target, camPos))
	right := v3norm(vec3{-forward.z, 0, forward.x})
	up := vec3{0, 1, 0}

	// Light direction (orbiting).
	lightDir := v3norm(vec3{
		math.Cos(r.time * lightAngle),
		0.8,
		math.Sin(r.time * lightAngle),
	})

	fw := float64(c.Width)
	fh := float64(c.Height)
	aspect := fw / fh
	fovScale := math.Tan(fov / 2.0)

	bg := canvas.Hex("#0A0B1A")
	objColor := canvas.RGB(0.8, 0.3, 0.2)

	for py := 0; py < c.Height; py++ {
		for px := 0; px < c.Width; px++ {
			// Normalized screen coordinates [-1, 1].
			u := (2.0*float64(px)/fw - 1.0) * aspect * fovScale
			v := (1.0 - 2.0*float64(py)/fh) * fovScale

			// Ray direction.
			dir := v3norm(v3add(v3add(v3scale(right, u), v3scale(up, v)), forward))

			// March.
			t := 0.0
			hit := false
			for step := 0; step < 64; step++ {
				p := v3add(camPos, v3scale(dir, t))
				d := r.sceneSDF(p)
				if d < 0.001 {
					hit = true
					break
				}
				t += d
				if t > 20.0 {
					break
				}
			}

			if !hit {
				c.SetPixel(px, py, bg)
				continue
			}

			// Shading.
			hitPos := v3add(camPos, v3scale(dir, t))
			n := r.normal(hitPos)

			// Phong: ambient + diffuse + specular.
			ambient := 0.1
			diffuse := math.Max(v3dot(n, lightDir), 0)
			reflect := v3sub(v3scale(n, 2.0*v3dot(n, lightDir)), lightDir)
			specular := math.Pow(math.Max(v3dot(reflect, v3scale(dir, -1)), 0), 32)

			brightness := ambient + diffuse*0.7 + specular*0.4
			if brightness > 1.0 {
				brightness = 1.0
			}

			col := canvas.RGB(
				objColor.R*brightness,
				objColor.G*brightness,
				objColor.B*brightness,
			)
			c.SetPixel(px, py, col)
		}
	}
}

func (r *RayMarch) SetParams(params map[string]float64) {
	for k, v := range params {
		if _, exists := r.params[k]; exists {
			r.params[k] = v
		}
	}
}

func (r *RayMarch) Params() map[string]float64 {
	out := make(map[string]float64, len(r.params))
	for k, v := range r.params {
		out[k] = v
	}
	return out
}
