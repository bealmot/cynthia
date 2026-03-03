// Demo — Bubble Tea app with plasma background and centered text panel.
// Press 1-5 to switch moods and watch spring-animated transitions.
// Press q or Ctrl+C to exit.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/bealmot/cynthia/bubbletea"
)

type model struct {
	fx   bubbletea.Model
	mood string
}

func newModel() model {
	cfg := bubbletea.DefaultConfig()
	cfg.InitialMood = "calm"
	return model{
		fx:   bubbletea.New(cfg),
		mood: "calm",
	}
}

func (m model) Init() tea.Cmd {
	return m.fx.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.mood = "calm"
			return m, bubbletea.ApplyMood("calm")
		case "2":
			m.mood = "thinking"
			return m, bubbletea.ApplyMood("thinking")
		case "3":
			m.mood = "alert"
			return m, bubbletea.ApplyMood("alert")
		case "4":
			m.mood = "celebration"
			return m, bubbletea.ApplyMood("celebration")
		case "5":
			m.mood = "dreaming"
			return m, bubbletea.ApplyMood("dreaming")
		}
	case tea.WindowSizeMsg:
		var cmd tea.Cmd
		m.fx, cmd = m.fx.Update(msg)
		return m, cmd
	}

	// Forward all other messages to effects model
	var cmd tea.Cmd
	m.fx, cmd = m.fx.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	// Render a simple centered text panel
	panel := fmt.Sprintf(
		"\n  ⬡ Hello Cynthia ⬡\n\n  Mood: %s\n\n  1:calm  2:thinking  3:alert\n  4:celebration  5:dreaming\n\n  Press q to quit\n",
		m.mood,
	)

	m.fx.SetContent(panel)
	return m.fx.View()
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
