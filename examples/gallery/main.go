// Gallery — browse all Cynthia effects interactively.
// Left/Right or h/l to navigate, q to quit.
// Each effect renders fullscreen with a minimal HUD overlay.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/effect"

	// Register all built-in effects.
	_ "github.com/bealmot/cynthia/effect"
)

const targetFPS = 30

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second/targetFPS, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type model struct {
	width, height int
	frame         uint64
	index         int
	effects       []string
	instances     map[string]effect.Effect
	mode          canvas.RenderMode
	paused        bool
}

func initialModel() model {
	names := effect.Names()
	sort.Strings(names)

	instances := make(map[string]effect.Effect, len(names))
	for _, name := range names {
		instances[name] = effect.Create(name)
	}

	return model{
		width:     80,
		height:    24,
		effects:   names,
		instances: instances,
		mode:      canvas.ModeHalfBlock,
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "right", "l", "n":
			m.index = (m.index + 1) % len(m.effects)
			return m, nil
		case "left", "h", "p":
			m.index = (m.index - 1 + len(m.effects)) % len(m.effects)
			return m, nil
		case "home":
			m.index = 0
			return m, nil
		case "end":
			m.index = len(m.effects) - 1
			return m, nil
		case " ":
			m.paused = !m.paused
			return m, nil
		case "1":
			m.mode = canvas.ModeHalfBlock
			return m, nil
		case "2":
			m.mode = canvas.ModeBraille
			return m, nil
		case "3":
			m.mode = canvas.ModeASCII
			return m, nil
		case "r":
			// Reset current effect.
			name := m.effects[m.index]
			m.instances[name] = effect.Create(name)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		if !m.paused {
			name := m.effects[m.index]
			fx := m.instances[name]
			dt := 1.0 / float64(targetFPS)
			fx.Step(m.frame, dt)
			m.frame++
		}
		return m, tick()
	}
	return m, nil
}

func (m model) View() tea.View {
	if m.width < 10 || m.height < 5 || len(m.effects) == 0 {
		return tea.NewView("")
	}

	name := m.effects[m.index]
	fx := m.instances[name]

	// Reserve 2 rows for HUD.
	canvasRows := m.height - 2

	c := canvas.New(m.width, canvasRows, m.mode)
	c.Clear(canvas.Black)
	fx.Render(c)
	c.Rasterize()

	// Build HUD.
	var buf strings.Builder
	buf.Grow(m.width * m.height * 24)

	// HUD line 1: effect name + index.
	modeName := [3]string{"half-block", "braille", "ascii"}[m.mode]
	title := fmt.Sprintf(" %s  [%d/%d]  (%s)", name, m.index+1, len(m.effects), modeName)
	if m.paused {
		title += "  PAUSED"
	}

	// Pad/truncate title.
	if len(title) > m.width {
		title = title[:m.width]
	}
	titlePad := m.width - len(title)
	if titlePad < 0 {
		titlePad = 0
	}

	// Title bar: white on dark.
	buf.WriteString("\x1b[48;2;20;15;30m\x1b[38;2;200;180;240m")
	buf.WriteString(title)
	buf.WriteString(strings.Repeat(" ", titlePad))
	buf.WriteString("\x1b[0m\n")

	// HUD line 2: params + controls.
	params := fx.Params()
	var paramParts []string
	// Sort param keys for stability.
	paramKeys := make([]string, 0, len(params))
	for k := range params {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)
	for _, k := range paramKeys {
		paramParts = append(paramParts, fmt.Sprintf("%s=%.2g", k, params[k]))
	}
	paramStr := strings.Join(paramParts, " ")

	controls := "  ←/→:nav  space:pause  1/2/3:mode  r:reset  q:quit"

	line2 := " " + paramStr
	remaining := m.width - len(line2) - len(controls)
	if remaining < 1 {
		// Truncate params if needed.
		maxParams := m.width - len(controls) - 2
		if maxParams > 0 && len(paramStr) > maxParams {
			paramStr = paramStr[:maxParams] + "…"
		}
		line2 = " " + paramStr
		remaining = m.width - len(line2) - len(controls)
		if remaining < 0 {
			remaining = 0
			controls = ""
		}
	}

	buf.WriteString("\x1b[48;2;15;12;25m\x1b[38;2;120;100;150m")
	buf.WriteString(line2)
	buf.WriteString(strings.Repeat(" ", remaining))
	buf.WriteString("\x1b[38;2;80;70;110m")
	buf.WriteString(controls)
	buf.WriteString("\x1b[0m\n")

	// Canvas content.
	buf.WriteString(canvas.RenderString(c, canvas.ProfileTrueColor))

	v := tea.NewView(buf.String())
	v.AltScreen = true
	return v
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
