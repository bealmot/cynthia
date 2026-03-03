package bubbletea

import (
	"strings"

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
	parseANSIToCells(s, m.contentLayer.Canvas)
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

// parseANSIToCells is a simplified ANSI parser that extracts visible characters
// and their colors from an ANSI-formatted string onto the cell grid.
func parseANSIToCells(s string, c *canvas.Canvas) {
	col, row := 0, 0
	fg := canvas.Color{R: 1, G: 1, B: 1, A: 1} // default white
	bg := canvas.Transparent

	i := 0
	for i < len(s) {
		if s[i] == '\n' {
			row++
			col = 0
			i++
			continue
		}

		// Parse ANSI escape sequence
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Find the end of the escape sequence
			j := i + 2
			for j < len(s) && !isCSITerminator(s[j]) {
				j++
			}
			if j < len(s) {
				// Parse SGR (Select Graphic Rendition)
				if s[j] == 'm' {
					params := s[i+2 : j]
					fg, bg = parseSGR(params, fg, bg)
				}
				i = j + 1
			} else {
				i = j
			}
			continue
		}

		// Regular character — place it on the cell grid
		if col < c.CellW && row < c.CellH {
			r := rune(s[i])
			// Handle multi-byte UTF-8
			if r >= 0x80 {
				runes := []rune(s[i:])
				if len(runes) > 0 {
					r = runes[0]
					i += len(string(r)) - 1
				}
			}
			c.SetCell(col, row, canvas.Cell{
				Rune: r,
				FG:   fg,
				BG:   bg,
			})
		}
		col++
		i++
	}
}

func isCSITerminator(b byte) bool {
	return b >= 0x40 && b <= 0x7E
}

// parseSGR handles ANSI SGR (m) sequences to extract fg/bg colors.
func parseSGR(params string, fg, bg canvas.Color) (canvas.Color, canvas.Color) {
	if params == "" || params == "0" {
		return canvas.Color{R: 1, G: 1, B: 1, A: 1}, canvas.Transparent
	}

	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "0": // reset
			fg = canvas.Color{R: 1, G: 1, B: 1, A: 1}
			bg = canvas.Transparent
		case "38": // set fg
			if i+1 < len(parts) && parts[i+1] == "2" && i+4 < len(parts) {
				// 24-bit: 38;2;r;g;b
				r := parseUint8(parts[i+2])
				g := parseUint8(parts[i+3])
				b := parseUint8(parts[i+4])
				fg = canvas.RGB8(r, g, b)
				i += 4
			} else if i+1 < len(parts) && parts[i+1] == "5" && i+2 < len(parts) {
				// 256-color: 38;5;n — just use the value as-is
				i += 2
			}
		case "48": // set bg
			if i+1 < len(parts) && parts[i+1] == "2" && i+4 < len(parts) {
				r := parseUint8(parts[i+2])
				g := parseUint8(parts[i+3])
				b := parseUint8(parts[i+4])
				bg = canvas.RGB8(r, g, b)
				i += 4
			} else if i+1 < len(parts) && parts[i+1] == "5" && i+2 < len(parts) {
				i += 2
			}
		case "39": // default fg
			fg = canvas.Color{R: 1, G: 1, B: 1, A: 1}
		case "49": // default bg
			bg = canvas.Transparent
		}
	}
	return fg, bg
}

func parseUint8(s string) uint8 {
	var v int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + int(c-'0')
		}
	}
	if v > 255 {
		v = 255
	}
	return uint8(v)
}
