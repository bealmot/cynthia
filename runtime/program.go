package runtime

import (
	"io"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/compose"
	"github.com/bealmot/cynthia/scene"
	"github.com/bealmot/cynthia/widget"
)

// Program is a BubbleTea-powered runtime for interactive widget support.
// It replaces Engine's raw goroutine render loop with a tea.Program that
// handles terminal input (keys, mouse) from the display TTY while MCP
// continues on stdin/stdout.
type Program struct {
	Scene      *scene.Scene
	Compositor *compose.Compositor
	Writer     *canvas.Writer
	Events     *widget.EventQueue
	FPS        int
	Mode       canvas.RenderMode

	cols, rows int
	mu         sync.Mutex
	frame      uint64
	tty        io.ReadWriter
	teaProg    *tea.Program
}

// NewProgram creates a BubbleTea-based runtime targeting the given TTY.
func NewProgram(tty io.ReadWriter, cols, rows, fps int, mode canvas.RenderMode) *Program {
	return &Program{
		Scene:      scene.New(),
		Compositor: compose.NewCompositor(cols, rows, mode),
		Writer:     canvas.NewWriter(tty, canvas.ProfileTrueColor),
		Events:     widget.NewEventQueue(),
		FPS:        fps,
		Mode:       mode,
		cols:       cols,
		rows:       rows,
		tty:        tty,
	}
}

// Lock acquires the program mutex. MCP handlers use this to safely mutate the scene.
func (p *Program) Lock() { p.mu.Lock() }

// Unlock releases the program mutex.
func (p *Program) Unlock() { p.mu.Unlock() }

// Cols returns the current terminal column count.
func (p *Program) Cols() int { return p.cols }

// Rows returns the current terminal row count.
func (p *Program) Rows() int { return p.rows }

// Resize changes the output dimensions. Must be called under Lock.
func (p *Program) Resize(cols, rows int) {
	p.cols = cols
	p.rows = rows
	p.Compositor.Resize(cols, rows)
}

// Run starts the BubbleTea program. Blocks until quit.
func (p *Program) Run() error {
	p.teaProg = tea.NewProgram(
		&programModel{prog: p},
		tea.WithInput(p.tty),
		tea.WithOutput(p.tty),
	)
	_, err := p.teaProg.Run()
	return err
}

// Quit signals the BubbleTea program to stop.
func (p *Program) Quit() {
	if p.teaProg != nil {
		p.teaProg.Quit()
	}
}

// frameMsg triggers a render frame.
type frameMsg time.Time

// programModel adapts Program to BubbleTea's Model interface.
type programModel struct {
	prog *Program
}

func (m *programModel) Init() tea.Cmd {
	return tickFrame(m.prog.FPS)
}

func (m *programModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case frameMsg:
		m.renderFrame()
		return m, tickFrame(m.prog.FPS)

	case tea.KeyPressMsg:
		m.handleKey(tea.Key(msg))
		return m, nil

	case tea.MouseClickMsg:
		m.handleMouseClick(tea.Mouse(msg))
		return m, nil

	case tea.WindowSizeMsg:
		m.prog.Lock()
		m.prog.Resize(msg.Width, msg.Height)
		m.prog.Unlock()
		return m, nil
	}

	return m, nil
}

func (m *programModel) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// renderFrame performs one render cycle — same pipeline as Engine.Step().
func (m *programModel) renderFrame() {
	p := m.prog
	dt := 1.0 / float64(p.FPS)

	p.mu.Lock()
	panels := p.Scene.Panels()
	frame := p.frame
	p.frame++

	for _, pan := range panels {
		if !pan.Visible {
			continue
		}

		if pan.Fill.A > 0 {
			pan.Canvas.Clear(pan.Fill)
		} else {
			pan.Canvas.Clear(canvas.Transparent)
		}

		if pan.Effect != nil {
			pan.Effect.Step(frame, dt)
			pan.Effect.Render(pan.Canvas)
		}

		if pan.Border != nil {
			pan.Border.Step(frame, dt)
			pan.Border.Render(pan.Canvas, 0, 0, pan.Width, pan.Height)
		}

		// Widget overrides text each frame
		if pan.Widget != nil {
			pan.Text = pan.Widget.View()
		}

		if pan.Text != "" {
			stampText(pan.Canvas, pan.Text, pan.Border != nil)
		}
	}

	sp := make([]compose.ScenePanel, len(panels))
	for i, pan := range panels {
		sp[i] = pan
	}
	p.mu.Unlock()

	p.Compositor.ComposeScene(sp)
	p.Compositor.Output().Rasterize()
	p.Writer.Render(p.Compositor.Output())
}

// handleKey routes a key event to the focused widget.
func (m *programModel) handleKey(key tea.Key) {
	p := m.prog
	p.mu.Lock()
	defer p.mu.Unlock()

	// Tab cycles focus between widget panels
	if key.Code == tea.KeyTab {
		p.Scene.FocusNext()
		return
	}

	focused := p.Scene.FocusedPanel()
	if focused == nil || focused.Widget == nil {
		return
	}

	// Convert BubbleTea key to widget key string
	widgetKey := keyToString(key)
	if widgetKey != "" {
		focused.Widget.HandleKey(widgetKey)
	}
}

// handleMouseClick routes a click to the hit-tested panel's widget.
func (m *programModel) handleMouseClick(mouse tea.Mouse) {
	p := m.prog
	p.mu.Lock()
	defer p.mu.Unlock()

	hit := p.Scene.HitTest(mouse.X, mouse.Y)
	if hit == nil {
		return
	}

	// If hit panel has a widget, focus it and route the click
	if hit.Widget != nil {
		p.Scene.Focus(hit.ID)

		// Convert to widget-local coordinates
		localX := mouse.X - hit.CellX()
		localY := mouse.Y - hit.CellY()

		// Adjust for border inset
		if hit.Border != nil {
			localX--
			localY--
		}

		btn := "left"
		if mouse.Button == tea.MouseRight {
			btn = "right"
		} else if mouse.Button == tea.MouseMiddle {
			btn = "middle"
		}

		hit.Widget.HandleMouse(localX, localY, btn)
	}
}

// keyToString converts a BubbleTea Key to the string format widgets expect.
func keyToString(key tea.Key) string {
	switch key.Code {
	case tea.KeyBackspace:
		return "backspace"
	case tea.KeyDelete:
		return "delete"
	case tea.KeyLeft:
		return "left"
	case tea.KeyRight:
		return "right"
	case tea.KeyHome:
		return "home"
	case tea.KeyEnd:
		return "end"
	case tea.KeyEnter:
		return "enter"
	case tea.KeySpace:
		return " "
	case tea.KeyTab:
		return "tab"
	case tea.KeyEscape:
		return "escape"
	default:
		// Printable character
		if key.Text != "" {
			return key.Text
		}
		// Single printable rune
		if key.Code >= 32 && key.Code < 127 {
			return string(key.Code)
		}
		return ""
	}
}

func tickFrame(fps int) tea.Cmd {
	d := time.Second / time.Duration(fps)
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return frameMsg(t)
	})
}
