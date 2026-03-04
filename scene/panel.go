// Package scene provides the scene graph for Cynthia's terminal rendering engine.
// Panels are the primary scene nodes — positioned, z-ordered canvases that can
// hold procedural effects, text content, and animated borders.
package scene

import (
	"github.com/bealmot/cynthia/border"
	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/compose"
	"github.com/bealmot/cynthia/director"
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

	// Per-panel spring interpolation (optional). When set, params are smoothly
	// transitioned each frame rather than snapping instantly.
	Director *director.PanelDirector

	// Spatial mask (optional). When set, constrains the effect to a region.
	MaskType   string
	MaskParams map[string]float64

	// Scope controls which tab this panel appears on.
	// Empty string or "global" means visible on all tabs.
	// Otherwise must match a tab name: "chat", "journal", "recommendations", "system", "settings".
	Scope string

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

// UpdateMask regenerates the canvas mask buffer from MaskType and MaskParams.
// Should be called when the canvas is resized or mask settings change.
func (p *Panel) UpdateMask() {
	c := p.Canvas
	if p.MaskType == "" || p.MaskType == "none" {
		c.ClearMask()
		return
	}

	mp := p.MaskParams
	get := func(key string, def float64) float64 {
		if v, ok := mp[key]; ok {
			return v
		}
		return def
	}

	var mask []float64
	switch p.MaskType {
	case "circle":
		mask = canvas.MaskCircle(c.Width, c.Height,
			get("cx", 0.5), get("cy", 0.5), get("radius", 0.4), get("feather", 0.1))
	case "rect":
		mask = canvas.MaskRect(c.Width, c.Height,
			get("x", 0.1), get("y", 0.1), get("width", 0.8), get("height", 0.8), get("feather", 0.1))
	case "gradient":
		dir := canvas.MaskDirTopToBottom
		if d, ok := mp["direction"]; ok {
			switch int(d) {
			case 1:
				dir = canvas.MaskDirLeftToRight
			case 2:
				dir = canvas.MaskDirRadial
			}
		}
		mask = canvas.MaskGradient(c.Width, c.Height, dir, get("start", 0), get("end", 1))
	case "vignette":
		mask = canvas.MaskVignette(c.Width, c.Height, get("strength", 0.8))
	default:
		c.ClearMask()
		return
	}
	c.SetMask(mask)
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
