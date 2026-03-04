package widget

import (
	"fmt"
	"strings"
)

// ProgressBar is a display-only progress indicator widget.
// Rendered as: [████████░░░░] 75%
type ProgressBar struct {
	value float64 // 0.0 to 1.0
	label string
	width int
	style string // "block" (default), "ascii"

	EventQueue *EventQueue
	PanelID    string
}

// NewProgressBar creates a progress bar widget.
func NewProgressBar(width int, label string) *ProgressBar {
	return &ProgressBar{
		width: width,
		label: label,
		style: "block",
	}
}

func (p *ProgressBar) Type() string { return "progress_bar" }

func (p *ProgressBar) HandleKey(key string) bool   { return false }
func (p *ProgressBar) HandleMouse(x, y int, button string) bool { return false }

func (p *ProgressBar) View() string {
	var b strings.Builder

	if p.label != "" {
		b.WriteString(p.label)
		b.WriteString(" ")
	}

	// Bar area: width minus brackets, space, and percentage (e.g. " 100%")
	barWidth := p.width - 2 - 5 // [] + " 100%"
	if p.label != "" {
		barWidth -= len(p.label) + 1
	}
	if barWidth < 1 {
		barWidth = 1
	}

	filled := int(p.value * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	b.WriteRune('[')
	if p.style == "ascii" {
		b.WriteString(strings.Repeat("=", filled))
		b.WriteString(strings.Repeat("-", barWidth-filled))
	} else {
		b.WriteString(strings.Repeat("█", filled))
		b.WriteString(strings.Repeat("░", barWidth-filled))
	}
	b.WriteRune(']')

	pct := int(p.value * 100)
	if pct > 100 {
		pct = 100
	}
	b.WriteString(fmt.Sprintf(" %3d%%", pct))

	return b.String()
}

func (p *ProgressBar) State() map[string]any {
	return map[string]any{
		"value": p.value,
		"label": p.label,
		"style": p.style,
	}
}

// SetValue sets the progress value (clamped to [0, 1]).
func (p *ProgressBar) SetValue(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.value = v
}

// SetLabel updates the display label.
func (p *ProgressBar) SetLabel(label string) {
	p.label = label
}

// SetStyle sets the bar style ("block" or "ascii").
func (p *ProgressBar) SetStyle(style string) {
	if style == "ascii" || style == "block" {
		p.style = style
	}
}
