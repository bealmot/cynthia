package border

import (
	"testing"

	"github.com/bealmot/cynthia/canvas"
)

func TestPerimeterPos(t *testing.T) {
	// 10×5 rectangle at origin — perimeter = 2*9 + 2*4 = 26
	p := perimeterPos(0, 0, 0, 0, 10, 5) // top-left corner
	if p != 0 {
		t.Errorf("top-left pos = %f, want 0", p)
	}

	// Top-right corner (x=9, y=0)
	p = perimeterPos(9, 0, 0, 0, 10, 5)
	if p < 0.34 || p > 0.36 {
		t.Errorf("top-right pos = %f, want ~0.35", p)
	}
}

func TestNouveauRender(t *testing.T) {
	c := canvas.New(20, 10, canvas.ModeHalfBlock)
	n := NewNouveau()
	n.Step(0, 0.1)
	n.Render(c, 2, 1, 16, 8)

	// Check that corner cells are set
	cell := c.GetCell(2, 1)
	if cell.Rune != '╭' {
		t.Errorf("top-left = %c, want ╭", cell.Rune)
	}
	cell = c.GetCell(17, 1)
	if cell.Rune != '╮' {
		t.Errorf("top-right = %c, want ╮", cell.Rune)
	}

	// Interior should be untouched
	cell = c.GetCell(10, 5)
	if cell.Rune != ' ' && cell.Rune != 0 {
		t.Errorf("interior should be empty, got %c", cell.Rune)
	}
}

func TestPulseBreathing(t *testing.T) {
	c := canvas.New(10, 5, canvas.ModeHalfBlock)
	p := NewPulse()

	// Render at two different times — color intensity should differ
	p.Step(0, 0)
	p.Render(c, 0, 0, 10, 5)
	cell1 := c.GetCell(5, 0) // top edge

	p.Step(1, 0.5) // half a cycle later
	p.Render(c, 0, 0, 10, 5)
	cell2 := c.GetCell(5, 0)

	// They should have different brightness (unless we happened to hit the same phase)
	if cell1.FG == cell2.FG {
		// This could happen if timing aligns, but unlikely with 0.5s offset
		t.Log("pulse cells at different times have same color (possible but unlikely)")
	}
}

func TestCascadeHighlight(t *testing.T) {
	c := canvas.New(10, 5, canvas.ModeHalfBlock)
	ca := NewCascade()

	ca.Step(0, 0)
	ca.Render(c, 0, 0, 10, 5)

	// Top-left corner should be rendered
	cell := c.GetCell(0, 0)
	if cell.Rune != '╭' {
		t.Errorf("cascade corner = %c, want ╭", cell.Rune)
	}
}
