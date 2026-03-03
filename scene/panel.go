// Package scene provides the scene graph for Cynthia's terminal rendering engine.
// Panels are the primary scene nodes — positioned, z-ordered canvases that can
// hold procedural effects, text content, and animated borders.
package scene

import (
	"github.com/bealmot/cynthia/border"
	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/compose"
	"github.com/bealmot/cynthia/effect"
	"github.com/bealmot/cynthia/widget"
)

// Panel is a scene graph node: a positioned canvas with compositing properties.
type Panel struct {
	ID     string
	Canvas *canvas.Canvas

	// Position in cell coordinates (float for sub-cell precision during animation).
	X, Y float64

	// Z-order: lower values render first (further back).
	Z int

	// Terminal cell dimensions of this panel.
	Width, Height int

	Opacity float64
	Visible bool
	Blend   compose.BlendMode

	// Content sources (applied in order: fill → effect → text → border).
	Fill   canvas.Color   // solid fill color (applied before effect)
	Effect effect.Effect  // procedural pixel content (fire, plasma, etc.)
	Text   string         // ANSI text content stamped onto cell grid
	Border border.Border  // animated border drawn on cell grid

	// Interactive widget (optional). When set, Widget.View() overrides Text each frame.
	Widget  widget.Widget
	Focused bool
}

// NewPanel creates a panel with the given ID and dimensions.
// It starts visible at full opacity with normal blending.
func NewPanel(id string, width, height int, mode canvas.RenderMode) *Panel {
	return &Panel{
		ID:      id,
		Canvas:  canvas.New(width, height, mode),
		Width:   width,
		Height:  height,
		Opacity: 1.0,
		Visible: true,
		Blend:   compose.BlendNormal,
	}
}

// CellX returns the integer cell column for compositing.
func (p *Panel) CellX() int { return int(p.X) }

// CellY returns the integer cell row for compositing.
func (p *Panel) CellY() int { return int(p.Y) }

// compose.ScenePanel interface implementation

func (p *Panel) GetID() string                  { return p.ID }
func (p *Panel) GetCanvas() *canvas.Canvas      { return p.Canvas }
func (p *Panel) GetCellPosition() (x, y int)    { return p.CellX(), p.CellY() }
func (p *Panel) GetZ() int                      { return p.Z }
func (p *Panel) GetOpacity() float64            { return p.Opacity }
func (p *Panel) GetBlend() compose.BlendMode    { return p.Blend }
func (p *Panel) IsVisible() bool                { return p.Visible }
