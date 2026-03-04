package director

// MoodPreset defines default visual parameters for a named mood.
type MoodPreset struct {
	Effect      string
	Intensity   float64
	BorderStyle string
	Speed       float64
	Params      map[string]float64
}

// Moods is the built-in set of mood presets.
var Moods = map[string]MoodPreset{
	"calm": {
		Effect:      "gradient",
		Intensity:   0.15,
		BorderStyle: "nouveau",
		Speed:       0.3,
		Params:      map[string]float64{"speed": 0.3, "scale": 2.0},
	},
	"thinking": {
		Effect:      "plasma",
		Intensity:   0.25,
		BorderStyle: "pulse",
		Speed:       0.8,
		Params:      map[string]float64{"speed": 0.8, "scale": 6.0},
	},
	"alert": {
		Effect:      "fire",
		Intensity:   0.4,
		BorderStyle: "cascade",
		Speed:       1.5,
		Params:      map[string]float64{"fire_intensity": 1.0, "decay": 0.03},
	},
	"celebration": {
		Effect:      "particles",
		Intensity:   0.5,
		BorderStyle: "nouveau",
		Speed:       1.0,
		Params:      map[string]float64{"speed": 1.2, "count": 200},
	},
	"dreaming": {
		Effect:      "starfield",
		Intensity:   0.3,
		BorderStyle: "nouveau",
		Speed:       0.2,
		Params:      map[string]float64{"speed": 0.2, "depth": 8.0},
	},
	"organic": {
		Effect:      "noise",
		Intensity:   0.3,
		BorderStyle: "nouveau",
		Speed:       0.4,
		Params:      map[string]float64{"speed": 0.4, "scale": 2.5, "octaves": 5, "persistence": 0.6},
	},
	"cosmic": {
		Effect:      "fractal",
		Intensity:   0.35,
		BorderStyle: "pulse",
		Speed:       0.2,
		Params:      map[string]float64{"speed": 0.2, "max_iter": 80, "zoom_rate": 0.3, "palette_speed": 0.4},
	},
	"digital": {
		Effect:      "life",
		Intensity:   0.2,
		BorderStyle: "cascade",
		Speed:       3.0,
		Params:      map[string]float64{"speed": 3.0, "density": 0.3},
	},
	"hypnotic": {
		Effect:      "tunnel",
		Intensity:   0.35,
		BorderStyle: "pulse",
		Speed:       0.8,
		Params:      map[string]float64{"speed": 0.8, "scale": 5.0, "twist": 0.7},
	},
	"retro": {
		Effect:      "crt",
		Intensity:   0.4,
		BorderStyle: "nouveau",
		Speed:       1.0,
		Params:      map[string]float64{"speed": 1.0, "scanlines": 1.0, "curvature": 0.3},
	},
	"stormy": {
		Effect:      "lightning",
		Intensity:   0.5,
		BorderStyle: "cascade",
		Speed:       1.0,
		Params:      map[string]float64{"speed": 1.0, "rate": 1.5, "branches": 4},
	},
	"serene": {
		Effect:      "aurora",
		Intensity:   0.3,
		BorderStyle: "nouveau",
		Speed:       0.3,
		Params:      map[string]float64{"speed": 0.3, "layers": 5, "wave": 2.5},
	},
	"melancholy": {
		Effect:      "rain",
		Intensity:   0.25,
		BorderStyle: "nouveau",
		Speed:       0.8,
		Params:      map[string]float64{"speed": 0.8, "density": 0.5},
	},
	"molten": {
		Effect:      "lava",
		Intensity:   0.35,
		BorderStyle: "pulse",
		Speed:       0.3,
		Params:      map[string]float64{"speed": 0.3, "blobs": 6, "radius": 0.14},
	},
	"geometric": {
		Effect:      "voronoi",
		Intensity:   0.3,
		BorderStyle: "nouveau",
		Speed:       0.4,
		Params:      map[string]float64{"speed": 0.4, "seeds": 15, "border": 0.025},
	},
	"chaotic": {
		Effect:      "glitch",
		Intensity:   0.45,
		BorderStyle: "cascade",
		Speed:       1.5,
		Params:      map[string]float64{"speed": 1.5, "intensity": 0.6, "rate": 3.0},
	},
	"mechanical": {
		Effect:      "wireframe",
		Intensity:   0.25,
		BorderStyle: "pulse",
		Speed:       0.4,
		Params:      map[string]float64{"speed": 0.4, "shape": 2, "fov": 3.0},
	},
	"biological": {
		Effect:      "grayscott",
		Intensity:   0.3,
		BorderStyle: "nouveau",
		Speed:       1.0,
		Params:      map[string]float64{"speed": 1.0, "feed": 0.055, "kill": 0.062},
	},
	"aquatic": {
		Effect:      "ripple",
		Intensity:   0.3,
		BorderStyle: "nouveau",
		Speed:       0.8,
		Params:      map[string]float64{"speed": 0.8, "damping": 0.97, "rate": 2.5},
	},
	"cartographic": {
		Effect:      "topographic",
		Intensity:   0.3,
		BorderStyle: "nouveau",
		Speed:       0.2,
		Params:      map[string]float64{"speed": 0.2, "scale": 2.0, "levels": 12},
	},
	"flowing": {
		Effect:      "flowfield",
		Intensity:   0.35,
		BorderStyle: "nouveau",
		Speed:       0.5,
		Params:      map[string]float64{"speed": 0.5, "particles": 300, "scale": 2.5, "trail": 0.96},
	},
	"analytical": {
		Effect:      "contour",
		Intensity:   0.25,
		BorderStyle: "pulse",
		Speed:       0.3,
		Params:      map[string]float64{"speed": 0.3, "scale": 2.5, "levels": 10},
	},
}
