package director

import (
	"testing"

	_ "github.com/bealmot/cynthia/effect" // register effects
)

func TestNewDirector(t *testing.T) {
	d := New()
	if d.CurrentMood != "calm" {
		t.Errorf("default mood = %q, want calm", d.CurrentMood)
	}
}

func TestApplyMood(t *testing.T) {
	d := New()
	d.ApplyMood("thinking")

	if d.CurrentMood != "thinking" {
		t.Errorf("mood = %q, want thinking", d.CurrentMood)
	}
	if d.CurrentEffect == nil {
		t.Fatal("CurrentEffect is nil after ApplyMood")
	}
	if d.CurrentEffect.Name() != "plasma" {
		t.Errorf("effect = %q, want plasma", d.CurrentEffect.Name())
	}
	if d.BorderStyle != "pulse" {
		t.Errorf("border = %q, want pulse", d.BorderStyle)
	}
}

func TestSpringInterpolation(t *testing.T) {
	d := New()
	d.ApplyMood("calm") // intensity = 0.15

	// Switch to alert — intensity target = 0.4
	d.ApplyMood("alert")

	// Step several frames — spring should converge toward 0.4
	for i := 0; i < 120; i++ {
		d.Step(1.0 / 60.0) // 2 seconds at 60fps
	}

	intensity := d.Intensity()
	if intensity < 0.35 || intensity > 0.45 {
		t.Errorf("after 2s of spring, intensity = %f, expected ~0.4", intensity)
	}
}

func TestDirective(t *testing.T) {
	d := New()

	intens := 0.6
	speed := 1.5
	d.Apply(Directive{
		Effect:          "fire",
		Intensity:       &intens,
		TransitionSpeed: &speed,
		BorderStyle:     "cascade",
		Params:          map[string]float64{"decay": 0.02},
	})

	if d.CurrentEffect == nil || d.CurrentEffect.Name() != "fire" {
		t.Errorf("expected fire effect, got %v", d.CurrentEffect)
	}
	if d.BorderStyle != "cascade" {
		t.Errorf("border = %q, want cascade", d.BorderStyle)
	}

	// Step to converge
	for i := 0; i < 120; i++ {
		d.Step(1.0 / 60.0)
	}

	if d.Intensity() < 0.55 {
		t.Errorf("intensity = %f, expected ~0.6", d.Intensity())
	}
}

func TestOverlayExpiry(t *testing.T) {
	d := New()
	d.Apply(Directive{
		Overlay: &OverlayDirective{Effect: "fire", Duration: 1.0},
	})

	if d.OverlayEffect == nil {
		t.Fatal("overlay should be set")
	}

	// Step past the duration
	for i := 0; i < 90; i++ {
		d.Step(1.0 / 60.0) // 1.5 seconds
	}

	if d.OverlayEffect != nil {
		t.Error("overlay should have expired after 1.5s")
	}
}

func TestUnknownMoodIgnored(t *testing.T) {
	d := New()
	d.ApplyMood("calm")
	d.ApplyMood("nonexistent")
	if d.CurrentMood != "calm" {
		t.Errorf("unknown mood should not change state, got %q", d.CurrentMood)
	}
}
