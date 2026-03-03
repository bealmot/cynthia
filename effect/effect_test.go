package effect

import (
	"testing"

	"github.com/bealmot/cynthia/canvas"
)

func TestRegistry(t *testing.T) {
	names := Names()
	if len(names) < 2 {
		t.Errorf("expected at least 2 registered effects, got %d", len(names))
	}

	// Check plasma exists
	p := Create("plasma")
	if p == nil {
		t.Fatal("Create(plasma) returned nil")
	}
	if p.Name() != "plasma" {
		t.Errorf("plasma.Name() = %q, want plasma", p.Name())
	}

	// Check fire exists
	f := Create("fire")
	if f == nil {
		t.Fatal("Create(fire) returned nil")
	}

	// Unknown returns nil
	if Create("nonexistent") != nil {
		t.Error("Create(nonexistent) should return nil")
	}
}

func TestPlasmaRender(t *testing.T) {
	c := canvas.New(20, 10, canvas.ModeHalfBlock)
	p := NewPlasma()

	p.Step(0, 1.0/30.0)
	p.Render(c)

	// Verify pixels were set (not all transparent)
	hasColor := false
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if c.GetPixel(x, y).A > 0 {
				hasColor = true
				break
			}
		}
		if hasColor {
			break
		}
	}
	if !hasColor {
		t.Error("plasma render produced no visible pixels")
	}
}

func TestPlasmaParams(t *testing.T) {
	p := NewPlasma()
	params := p.Params()
	if params["speed"] != 1.0 {
		t.Errorf("default speed = %f, want 1.0", params["speed"])
	}

	p.SetParams(map[string]float64{"speed": 2.0})
	if p.Params()["speed"] != 2.0 {
		t.Errorf("after SetParams speed = %f, want 2.0", p.Params()["speed"])
	}

	// Unknown params are ignored
	p.SetParams(map[string]float64{"bogus": 99})
	if _, exists := p.Params()["bogus"]; exists {
		t.Error("unknown param 'bogus' should not be stored")
	}
}

func TestFireRender(t *testing.T) {
	c := canvas.New(20, 10, canvas.ModeHalfBlock)
	f := NewFire()

	// Step a few frames to let fire propagate
	for i := 0; i < 30; i++ {
		f.Step(uint64(i), 1.0/30.0)
		f.Render(c) // ensures size is set
	}

	// Bottom pixels should be hot (near-white/yellow)
	bottomPx := c.GetPixel(10, c.Height-1)
	if bottomPx.R < 0.5 {
		t.Errorf("bottom pixel R = %f, expected hot (>0.5)", bottomPx.R)
	}
}
