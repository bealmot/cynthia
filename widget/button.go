package widget

import "strings"

// Button is a clickable button widget.
type Button struct {
	label      string
	width      int
	pressed    bool
	clickCount int

	// EventQueue is set by the runtime to emit click events.
	EventQueue *EventQueue
	PanelID    string
}

// NewButton creates a button widget with the given label.
func NewButton(label string, width int) *Button {
	return &Button{
		label: label,
		width: width,
	}
}

func (b *Button) Type() string { return "button" }

func (b *Button) HandleKey(key string) bool {
	if key == "enter" || key == " " {
		b.press()
		return true
	}
	return false
}

func (b *Button) HandleMouse(x, y int, button string) bool {
	if button == "left" {
		b.press()
		return true
	}
	return false
}

func (b *Button) press() {
	b.pressed = true
	b.clickCount++
	if b.EventQueue != nil {
		b.EventQueue.Emit(b.PanelID, "click", map[string]any{
			"label":       b.label,
			"click_count": b.clickCount,
		})
	}
}

func (b *Button) View() string {
	var sb strings.Builder

	if b.pressed {
		sb.WriteString("[>")
		sb.WriteString(b.label)
		sb.WriteString("<]")
		b.pressed = false // reset visual after one frame
	} else {
		sb.WriteString("[ ")
		sb.WriteString(b.label)
		sb.WriteString(" ]")
	}

	// Pad to width if needed
	result := sb.String()
	runes := []rune(result)
	if b.width > 0 && len(runes) < b.width {
		pad := b.width - len(runes)
		result += strings.Repeat(" ", pad)
	}
	return result
}

func (b *Button) State() map[string]any {
	return map[string]any{
		"label":       b.label,
		"pressed":     b.pressed,
		"click_count": b.clickCount,
	}
}
