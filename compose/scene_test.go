package compose

import (
	"testing"

	"github.com/bealmot/cynthia/canvas"
)

// testPanel is a minimal ScenePanel for testing ComposeScene.
type testPanel struct {
	id      string
	canvas  *canvas.Canvas
	x, y    int
	z       int
	opacity float64
	blend   BlendMode
	visible bool
}

func (p *testPanel) GetID() string                  { return p.id }
func (p *testPanel) GetCanvas() *canvas.Canvas      { return p.canvas }
func (p *testPanel) GetCellPosition() (x, y int)    { return p.x, p.y }
func (p *testPanel) GetZ() int                      { return p.z }
func (p *testPanel) GetOpacity() float64            { return p.opacity }
func (p *testPanel) GetBlend() BlendMode            { return p.blend }
func (p *testPanel) IsVisible() bool                { return p.visible }

func newTestPanel(id string, w, h, x, y, z int) *testPanel {
	c := canvas.New(w, h, canvas.ModeHalfBlock)
	c.Clear(canvas.RGB(1, 0, 0)) // red
	return &testPanel{
		id:      id,
		canvas:  c,
		x:       x,
		y:       y,
		z:       z,
		opacity: 1.0,
		blend:   BlendNormal,
		visible: true,
	}
}

func TestComposeSceneEmpty(t *testing.T) {
	comp := NewCompositor(80, 24, canvas.ModeHalfBlock)
	comp.ComposeScene(nil)

	// Output should be all transparent
	out := comp.Output()
	for i := 0; i < out.Width*out.Height; i++ {
		if out.Pixels[i].A != 0 {
			t.Fatalf("pixel %d should be transparent, got A=%f", i, out.Pixels[i].A)
		}
	}
}

func TestComposeSceneSinglePanel(t *testing.T) {
	comp := NewCompositor(80, 24, canvas.ModeHalfBlock)
	p := newTestPanel("red", 10, 5, 2, 3, 0)

	comp.ComposeScene([]ScenePanel{p})

	out := comp.Output()

	// Pixel at the panel's top-left corner should be red.
	// Panel at cell (2,3) → pixel (2, 6) in HalfBlock mode.
	px := out.GetPixel(2, 6)
	if px.R < 0.9 || px.A < 0.9 {
		t.Errorf("pixel at panel origin = %+v, want red", px)
	}

	// Pixel outside panel should be transparent.
	px = out.GetPixel(0, 0)
	if px.A != 0 {
		t.Errorf("pixel outside panel = %+v, want transparent", px)
	}
}

func TestComposeSceneZOrder(t *testing.T) {
	comp := NewCompositor(80, 24, canvas.ModeHalfBlock)

	// Two overlapping panels at same position
	back := newTestPanel("back", 10, 5, 0, 0, 0)
	back.canvas.Clear(canvas.RGB(0, 0, 1)) // blue

	front := newTestPanel("front", 10, 5, 0, 0, 1)
	front.canvas.Clear(canvas.RGB(1, 0, 0)) // red

	comp.ComposeScene([]ScenePanel{front, back}) // deliberately unsorted

	out := comp.Output()
	px := out.GetPixel(0, 0)

	// Front panel (red, z=1) should be on top
	if px.R < 0.9 {
		t.Errorf("overlapping pixel = %+v, want red (front panel)", px)
	}
}

func TestComposeSceneInvisiblePanel(t *testing.T) {
	comp := NewCompositor(80, 24, canvas.ModeHalfBlock)
	p := newTestPanel("hidden", 10, 5, 0, 0, 0)
	p.visible = false

	comp.ComposeScene([]ScenePanel{p})

	out := comp.Output()
	px := out.GetPixel(0, 0)
	if px.A != 0 {
		t.Errorf("invisible panel should not render, got %+v", px)
	}
}

func TestComposeSceneOpacity(t *testing.T) {
	comp := NewCompositor(80, 24, canvas.ModeHalfBlock)
	p := newTestPanel("dim", 10, 5, 0, 0, 0)
	p.opacity = 0.5

	comp.ComposeScene([]ScenePanel{p})

	out := comp.Output()
	px := out.GetPixel(0, 0)

	// Red channel should be roughly halved by 50% opacity
	if px.R < 0.4 || px.R > 0.6 {
		t.Errorf("50%% opacity pixel R=%f, want ~0.5", px.R)
	}
}

func TestComposeSceneClipping(t *testing.T) {
	comp := NewCompositor(80, 24, canvas.ModeHalfBlock)

	// Panel that extends beyond the right edge
	p := newTestPanel("offscreen", 100, 5, 70, 0, 0)
	comp.ComposeScene([]ScenePanel{p})

	// Should not panic — just clips
	out := comp.Output()
	px := out.GetPixel(75, 0)
	if px.R < 0.9 {
		t.Errorf("visible portion of clipped panel should render, got %+v", px)
	}
}

func TestCellToPixelOffset(t *testing.T) {
	tests := []struct {
		mode       canvas.RenderMode
		cx, cy     int
		wantX, wantY int
	}{
		{canvas.ModeHalfBlock, 5, 3, 5, 6},
		{canvas.ModeBraille, 5, 3, 10, 12},
		{canvas.ModeASCII, 5, 3, 5, 3},
	}

	for _, tt := range tests {
		px, py := cellToPixelOffset(tt.cx, tt.cy, tt.mode)
		if px != tt.wantX || py != tt.wantY {
			t.Errorf("cellToPixelOffset(%d,%d, mode=%d) = %d,%d, want %d,%d",
				tt.cx, tt.cy, tt.mode, px, py, tt.wantX, tt.wantY)
		}
	}
}
