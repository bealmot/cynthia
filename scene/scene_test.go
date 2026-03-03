package scene

import (
	"testing"

	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/compose"
)

func TestNewPanel(t *testing.T) {
	p := NewPanel("test", 20, 10, canvas.ModeHalfBlock)

	if p.ID != "test" {
		t.Errorf("ID = %q, want %q", p.ID, "test")
	}
	if p.Width != 20 || p.Height != 10 {
		t.Errorf("size = %dx%d, want 20x10", p.Width, p.Height)
	}
	if p.Canvas.CellW != 20 || p.Canvas.CellH != 10 {
		t.Errorf("canvas cells = %dx%d, want 20x10", p.Canvas.CellW, p.Canvas.CellH)
	}
	if !p.Visible {
		t.Error("panel should be visible by default")
	}
	if p.Opacity != 1.0 {
		t.Errorf("opacity = %f, want 1.0", p.Opacity)
	}
	if p.Blend != compose.BlendNormal {
		t.Errorf("blend = %d, want BlendNormal", p.Blend)
	}
}

func TestPanelCellPosition(t *testing.T) {
	p := NewPanel("p", 10, 5, canvas.ModeHalfBlock)
	p.X = 3.7
	p.Y = 2.9

	if p.CellX() != 3 {
		t.Errorf("CellX = %d, want 3", p.CellX())
	}
	if p.CellY() != 2 {
		t.Errorf("CellY = %d, want 2", p.CellY())
	}
}

func TestPanelSatisfiesScenePanel(t *testing.T) {
	p := NewPanel("iface-test", 10, 5, canvas.ModeHalfBlock)
	p.X = 5
	p.Y = 3
	p.Z = 2
	p.Opacity = 0.8

	// Verify it satisfies the interface
	var sp compose.ScenePanel = p

	if sp.GetID() != "iface-test" {
		t.Errorf("GetID = %q", sp.GetID())
	}
	x, y := sp.GetCellPosition()
	if x != 5 || y != 3 {
		t.Errorf("GetCellPosition = %d,%d, want 5,3", x, y)
	}
	if sp.GetZ() != 2 {
		t.Errorf("GetZ = %d, want 2", sp.GetZ())
	}
	if sp.GetOpacity() != 0.8 {
		t.Errorf("GetOpacity = %f, want 0.8", sp.GetOpacity())
	}
}

func TestSceneAddGetRemove(t *testing.T) {
	s := New()
	p := NewPanel("a", 10, 5, canvas.ModeHalfBlock)

	s.Add(p)

	if s.Count() != 1 {
		t.Fatalf("Count = %d, want 1", s.Count())
	}
	if got := s.Get("a"); got != p {
		t.Error("Get returned wrong panel")
	}
	if s.Get("nonexistent") != nil {
		t.Error("Get should return nil for missing ID")
	}

	if !s.Remove("a") {
		t.Error("Remove should return true for existing panel")
	}
	if s.Count() != 0 {
		t.Error("scene should be empty after remove")
	}
	if s.Remove("a") {
		t.Error("Remove should return false for missing panel")
	}
}

func TestScenePanelsZOrder(t *testing.T) {
	s := New()
	p1 := NewPanel("back", 10, 5, canvas.ModeHalfBlock)
	p1.Z = 10
	p2 := NewPanel("front", 10, 5, canvas.ModeHalfBlock)
	p2.Z = 1
	p3 := NewPanel("mid", 10, 5, canvas.ModeHalfBlock)
	p3.Z = 5

	s.Add(p1)
	s.Add(p2)
	s.Add(p3)

	panels := s.Panels()
	if len(panels) != 3 {
		t.Fatalf("len = %d, want 3", len(panels))
	}
	if panels[0].ID != "front" || panels[1].ID != "mid" || panels[2].ID != "back" {
		t.Errorf("z-order = [%s, %s, %s], want [front, mid, back]",
			panels[0].ID, panels[1].ID, panels[2].ID)
	}
}

func TestScenePanelsStableTieBreak(t *testing.T) {
	s := New()
	// Same Z, should sort by ID
	s.Add(&Panel{ID: "charlie", Z: 0, Canvas: canvas.New(1, 1, canvas.ModeHalfBlock), Width: 1, Height: 1, Visible: true, Opacity: 1})
	s.Add(&Panel{ID: "alpha", Z: 0, Canvas: canvas.New(1, 1, canvas.ModeHalfBlock), Width: 1, Height: 1, Visible: true, Opacity: 1})
	s.Add(&Panel{ID: "bravo", Z: 0, Canvas: canvas.New(1, 1, canvas.ModeHalfBlock), Width: 1, Height: 1, Visible: true, Opacity: 1})

	panels := s.Panels()
	if panels[0].ID != "alpha" || panels[1].ID != "bravo" || panels[2].ID != "charlie" {
		t.Errorf("tie-break order = [%s, %s, %s], want [alpha, bravo, charlie]",
			panels[0].ID, panels[1].ID, panels[2].ID)
	}
}

func TestSceneClear(t *testing.T) {
	s := New()
	s.Add(NewPanel("a", 1, 1, canvas.ModeHalfBlock))
	s.Add(NewPanel("b", 1, 1, canvas.ModeHalfBlock))
	s.Clear()

	if s.Count() != 0 {
		t.Errorf("Count after Clear = %d, want 0", s.Count())
	}
}

func TestGenerateIDUnique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateID()
		if ids[id] {
			t.Fatalf("duplicate ID: %s", id)
		}
		ids[id] = true
	}
}
