package theme

import "github.com/bealmot/cynthia/canvas"

func init() {
	Register(&Theme{
		Name:     "nord",
		BG:       canvas.Hex("#2E3440"),
		BGPanel:  canvas.Hex("#3B4252"),
		BGActive: canvas.Hex("#434C5E"),
		Text:     canvas.Hex("#ECEFF4"),
		TextDim:  canvas.Hex("#D8DEE9"),

		Primary:   canvas.Hex("#88C0D0"), // frost
		Secondary: canvas.Hex("#81A1C1"), // frost 2
		Tertiary:  canvas.Hex("#BF616A"), // aurora red
		Warm:      canvas.Hex("#EBCB8B"), // aurora yellow
		Cool:      canvas.Hex("#A3BE8C"), // aurora green

		EffectPalette: canvas.Palette{
			canvas.Hex("#2E3440"),
			canvas.Hex("#3B4252"),
			canvas.Hex("#88C0D0"),
			canvas.Hex("#81A1C1"),
			canvas.Hex("#5E81AC"),
			canvas.Hex("#88C0D0"),
			canvas.Hex("#3B4252"),
			canvas.Hex("#2E3440"),
		},
	})
}
