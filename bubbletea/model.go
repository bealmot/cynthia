package bubbletea

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/compose"
	"github.com/bealmot/cynthia/director"
	_ "github.com/bealmot/cynthia/effect" // register built-in effects
)

// Model is a Bubble Tea v2 model that wraps the Cynthia effects engine.
// Embed it in your parent model and call SetContent() with your rendered UI.
type Model struct {
	config     Config
	compositor *compose.Compositor
	director   *director.Director
	frame      uint64
	cols, rows int

	// Layer references
	bgLayer      *compose.Layer
	contentLayer *compose.Layer
	overlayLayer *compose.Layer
}

// New creates a Cynthia Model with the given config.
func New(cfg Config) Model {
	if cfg.FPS <= 0 {
		cfg.FPS = 60
	}
	if cfg.Cols <= 0 {
		cfg.Cols = 80
	}
	if cfg.Rows <= 0 {
		cfg.Rows = 24
	}

	comp := compose.NewCompositor(cfg.Cols, cfg.Rows, cfg.RenderMode)

	bg := compose.NewLayer("background", 0, cfg.Cols, cfg.Rows, cfg.RenderMode)
	content := compose.NewLayer("content", 20, cfg.Cols, cfg.Rows, cfg.RenderMode)
	overlay := compose.NewLayer("overlay", 30, cfg.Cols, cfg.Rows, cfg.RenderMode)
	overlay.Blend = compose.BlendAdditive

	comp.AddLayer(bg)
	comp.AddLayer(content)
	comp.AddLayer(overlay)

	dir := director.New()
	if cfg.InitialMood != "" {
		dir.ApplyMood(cfg.InitialMood)
	}

	return Model{
		config:       cfg,
		compositor:   comp,
		director:     dir,
		cols:         cfg.Cols,
		rows:         cfg.Rows,
		bgLayer:      bg,
		contentLayer: content,
		overlayLayer: overlay,
	}
}

// Init starts the animation tick loop.
func (m Model) Init() tea.Cmd {
	return tick(m.config.FPS)
}

// Update handles frame ticks, directives, and window resizes.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FrameMsg:
		dt := 1.0 / float64(m.config.FPS)

		// Advance the director (spring interpolation)
		m.director.Step(dt)

		// Step and render the main effect
		if m.director.CurrentEffect != nil {
			m.director.CurrentEffect.Step(m.frame, dt)
			m.bgLayer.Canvas.Clear(canvas.Transparent)
			m.director.CurrentEffect.Render(m.bgLayer.Canvas)
			m.bgLayer.Canvas.Rasterize()
		}

		// Step and render overlay
		if m.director.OverlayEffect != nil {
			m.director.OverlayEffect.Step(m.frame, dt)
			m.overlayLayer.Canvas.Clear(canvas.Transparent)
			m.director.OverlayEffect.Render(m.overlayLayer.Canvas)
			m.overlayLayer.Canvas.Rasterize()
			m.overlayLayer.Visible = true
		} else {
			m.overlayLayer.Visible = false
		}

		// Apply director intensity as bg layer opacity
		m.bgLayer.Opacity = m.director.Intensity()

		m.frame++
		return m, tick(m.config.FPS)

	case DirectiveMsg:
		m.director.Apply(msg.Directive)
		return m, nil

	case tea.WindowSizeMsg:
		m.cols = msg.Width
		m.rows = msg.Height
		m.compositor.Resize(m.cols, m.rows)
		m.bgLayer.Canvas.Resize(m.cols, m.rows)
		m.contentLayer.Canvas.Resize(m.cols, m.rows)
		m.overlayLayer.Canvas.Resize(m.cols, m.rows)
		return m, nil
	}

	return m, nil
}

// SetContent places a rendered ANSI string on the content layer.
// This is how the parent TUI passes its normal UI rendering into the compositor.
func (m *Model) SetContent(s string) {
	m.contentLayer.Canvas.Clear(canvas.Transparent)
	canvas.ParseANSIToCells(s, m.contentLayer.Canvas)
}

// View composites all layers and returns the result as a tea.View.
func (m Model) View() tea.View {
	m.compositor.Compose()
	out := canvas.RenderString(m.compositor.Output(), m.config.ColorProfile)
	v := tea.NewView(out)
	v.AltScreen = true
	return v
}

// Director returns the underlying director for direct access.
func (m *Model) Director() *director.Director {
	return m.director
}

