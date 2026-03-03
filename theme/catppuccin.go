package theme

import "github.com/bealmot/cynthia/canvas"

// Catppuccin Mocha flavor
func init() {
	Register(&Theme{
		Name:     "catppuccin",
		BG:       canvas.Hex("#1E1E2E"),
		BGPanel:  canvas.Hex("#181825"),
		BGActive: canvas.Hex("#313244"),
		Text:     canvas.Hex("#CDD6F4"),
		TextDim:  canvas.Hex("#A6ADC8"),

		Primary:   canvas.Hex("#CBA6F7"), // mauve
		Secondary: canvas.Hex("#94E2D5"), // teal
		Tertiary:  canvas.Hex("#F38BA8"), // red
		Warm:      canvas.Hex("#F9E2AF"), // yellow
		Cool:      canvas.Hex("#89B4FA"), // blue

		EffectPalette: canvas.Palette{
			canvas.Hex("#1E1E2E"),
			canvas.Hex("#313244"),
			canvas.Hex("#CBA6F7"),
			canvas.Hex("#94E2D5"),
			canvas.Hex("#F38BA8"),
			canvas.Hex("#F9E2AF"),
			canvas.Hex("#CBA6F7"),
			canvas.Hex("#1E1E2E"),
		},
	})
}
