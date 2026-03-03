package theme

import "github.com/bealmot/cynthia/canvas"

func init() {
	Register(&Theme{
		Name:     "metis",
		BG:       canvas.Hex("#12101A"),
		BGPanel:  canvas.Hex("#1A1726"),
		BGActive: canvas.Hex("#221F30"),
		Text:     canvas.Hex("#F5F0E8"),
		TextDim:  canvas.Hex("#8A8494"),

		Primary:   canvas.Hex("#C4B5F4"), // lavender
		Secondary: canvas.Hex("#93E4D4"), // aqua
		Tertiary:  canvas.Hex("#F4B8D4"), // rose
		Warm:      canvas.Hex("#E8D5A3"), // gold
		Cool:      canvas.Hex("#B8D4F4"), // celeste

		EffectPalette: canvas.Palette{
			canvas.Hex("#12101A"),
			canvas.Hex("#4A1C6C"),
			canvas.Hex("#C4B5F4"),
			canvas.Hex("#93E4D4"),
			canvas.Hex("#F4B8D4"),
			canvas.Hex("#E8D5A3"),
			canvas.Hex("#C4B5F4"),
			canvas.Hex("#12101A"),
		},
	})
}
