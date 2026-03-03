package theme

import "github.com/bealmot/cynthia/canvas"

func init() {
	Register(&Theme{
		Name:     "dracula",
		BG:       canvas.Hex("#282A36"),
		BGPanel:  canvas.Hex("#21222C"),
		BGActive: canvas.Hex("#44475A"),
		Text:     canvas.Hex("#F8F8F2"),
		TextDim:  canvas.Hex("#6272A4"),

		Primary:   canvas.Hex("#BD93F9"), // purple
		Secondary: canvas.Hex("#8BE9FD"), // cyan
		Tertiary:  canvas.Hex("#FF79C6"), // pink
		Warm:      canvas.Hex("#F1FA8C"), // yellow
		Cool:      canvas.Hex("#50FA7B"), // green

		EffectPalette: canvas.Palette{
			canvas.Hex("#282A36"),
			canvas.Hex("#44475A"),
			canvas.Hex("#BD93F9"),
			canvas.Hex("#8BE9FD"),
			canvas.Hex("#FF79C6"),
			canvas.Hex("#F1FA8C"),
			canvas.Hex("#BD93F9"),
			canvas.Hex("#282A36"),
		},
	})
}
