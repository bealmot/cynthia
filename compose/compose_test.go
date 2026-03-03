package compose

import (
	"testing"

	"github.com/bealmot/cynthia/canvas"
)

func TestCompositorBasic(t *testing.T) {
	comp := NewCompositor(10, 5, canvas.ModeHalfBlock)

	// Background layer — solid red
	bg := NewLayer("bg", 0, 10, 5, canvas.ModeHalfBlock)
	bg.Canvas.Clear(canvas.RGB(1, 0, 0))
	comp.AddLayer(bg)

	// Foreground layer — 50% opacity green
	fg := NewLayer("fg", 10, 10, 5, canvas.ModeHalfBlock)
	fg.Canvas.Clear(canvas.RGB(0, 1, 0))
	fg.Opacity = 0.5
	comp.AddLayer(fg)

	comp.Compose()

	// Result should be a blend of red and green
	px := comp.Output().GetPixel(5, 5)
	if px.A < 0.99 {
		t.Errorf("composited alpha = %f, want ~1.0", px.A)
	}
	// Green at 50% over red: R should be reduced, G should be present
	if px.G < 0.3 {
		t.Errorf("composited G = %f, expected significant green component", px.G)
	}
}

func TestCompositorZOrder(t *testing.T) {
	comp := NewCompositor(10, 5, canvas.ModeHalfBlock)

	// Layer at z=10 — blue
	l1 := NewLayer("bottom", 10, 10, 5, canvas.ModeHalfBlock)
	l1.Canvas.Clear(canvas.RGB(0, 0, 1))
	comp.AddLayer(l1)

	// Layer at z=0 — red (should be painted FIRST, so blue goes on top)
	l2 := NewLayer("behind", 0, 10, 5, canvas.ModeHalfBlock)
	l2.Canvas.Clear(canvas.RGB(1, 0, 0))
	comp.AddLayer(l2)

	comp.Compose()

	// Blue (z=10) is on top, so result should be blue
	px := comp.Output().GetPixel(5, 5)
	if px.B < 0.99 {
		t.Errorf("z-order: B = %f, expected blue on top", px.B)
	}
}

func TestCompositorVisibility(t *testing.T) {
	comp := NewCompositor(10, 5, canvas.ModeHalfBlock)

	l := NewLayer("hidden", 0, 10, 5, canvas.ModeHalfBlock)
	l.Canvas.Clear(canvas.RGB(1, 0, 0))
	l.Visible = false
	comp.AddLayer(l)

	comp.Compose()

	// Should be transparent (nothing visible)
	px := comp.Output().GetPixel(5, 5)
	if px.A > 0 {
		t.Errorf("hidden layer should not contribute, got A=%f", px.A)
	}
}

func TestGetRemoveLayer(t *testing.T) {
	comp := NewCompositor(10, 5, canvas.ModeHalfBlock)

	l := NewLayer("test", 0, 10, 5, canvas.ModeHalfBlock)
	comp.AddLayer(l)

	if comp.GetLayer("test") == nil {
		t.Error("GetLayer(test) returned nil")
	}
	if comp.GetLayer("nope") != nil {
		t.Error("GetLayer(nope) should return nil")
	}

	comp.RemoveLayer("test")
	if comp.GetLayer("test") != nil {
		t.Error("after RemoveLayer, GetLayer should return nil")
	}
}

func TestBlendScreen(t *testing.T) {
	// Screen blend: brightens
	src := canvas.RGB(0.5, 0, 0)
	dst := canvas.RGB(0, 0.5, 0)
	result := BlendPixel(src, dst, BlendScreen)
	// screen(0.5, 0) = 0.5, screen(0, 0.5) = 0.5
	if result.R < 0.49 || result.G < 0.49 {
		t.Errorf("screen blend = (%.2f, %.2f, %.2f), expected both ~0.5", result.R, result.G, result.B)
	}
}

func TestBlendAdditive(t *testing.T) {
	src := canvas.RGB(0.5, 0.3, 0)
	dst := canvas.RGB(0.3, 0.5, 0.1)
	result := BlendPixel(src, dst, BlendAdditive)
	if result.R < 0.79 {
		t.Errorf("additive R = %f, expected ~0.8", result.R)
	}
}
