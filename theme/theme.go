// Package theme provides color palette presets for the Cynthia effects engine.
package theme

import "github.com/bealmot/cynthia/canvas"

// Theme defines a named color palette with semantic roles.
type Theme struct {
	Name string

	// Base colors
	BG       canvas.Color
	BGPanel  canvas.Color
	BGActive canvas.Color
	Text     canvas.Color
	TextDim  canvas.Color

	// Accent colors (used by effects and borders)
	Primary   canvas.Color
	Secondary canvas.Color
	Tertiary  canvas.Color
	Warm      canvas.Color
	Cool      canvas.Color

	// Effect palette (gradient used for plasma, fire palette replacement, etc.)
	EffectPalette canvas.Palette
}

var registry = map[string]*Theme{}

// Register adds a theme to the global registry.
func Register(t *Theme) {
	registry[t.Name] = t
}

// Get returns a theme by name, or nil.
func Get(name string) *Theme {
	return registry[name]
}

// Names returns all registered theme names.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
