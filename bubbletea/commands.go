package bubbletea

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bealmot/cynthia/director"
)

// FrameMsg is sent on each animation tick.
type FrameMsg time.Time

// DirectiveMsg wraps an LLM directive for the effects director.
type DirectiveMsg struct {
	Directive director.Directive
}

// ApplyMood returns a command that sends a mood change to the model.
func ApplyMood(mood string) tea.Cmd {
	return func() tea.Msg {
		return DirectiveMsg{Directive: director.Directive{Mood: mood}}
	}
}

// ApplyDirective returns a command that sends a full directive to the model.
func ApplyDirective(d director.Directive) tea.Cmd {
	return func() tea.Msg {
		return DirectiveMsg{Directive: d}
	}
}

// tick returns a command that fires a FrameMsg after the given duration.
func tick(fps int) tea.Cmd {
	d := time.Second / time.Duration(fps)
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return FrameMsg(t)
	})
}
