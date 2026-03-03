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
}
