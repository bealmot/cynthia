package theme

import "testing"

func TestRegisteredThemes(t *testing.T) {
	expected := []string{"metis", "dracula", "nord", "catppuccin"}
	for _, name := range expected {
		th := Get(name)
		if th == nil {
			t.Errorf("theme %q not registered", name)
			continue
		}
		if th.Name != name {
			t.Errorf("theme.Name = %q, want %q", th.Name, name)
		}
		if th.BG.A <= 0 {
			t.Errorf("theme %q has transparent BG", name)
		}
		if len(th.EffectPalette) < 4 {
			t.Errorf("theme %q has too few palette entries: %d", name, len(th.EffectPalette))
		}
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) < 4 {
		t.Errorf("expected at least 4 themes, got %d", len(names))
	}
}

func TestUnknownTheme(t *testing.T) {
	if Get("nonexistent") != nil {
		t.Error("Get(nonexistent) should return nil")
	}
}
