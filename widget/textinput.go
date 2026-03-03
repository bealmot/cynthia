package widget

import (
	"strings"
)

// TextInput is a single-line text input widget.
type TextInput struct {
	value       []rune
	cursor      int
	placeholder string
	width       int

	// EventQueue is set by the runtime to emit submit events.
	EventQueue *EventQueue
	PanelID    string
}

// NewTextInput creates a text input widget.
func NewTextInput(width int, placeholder string) *TextInput {
	return &TextInput{
		width:       width,
		placeholder: placeholder,
	}
}

func (t *TextInput) Type() string { return "text_input" }

func (t *TextInput) HandleKey(key string) bool {
	switch key {
	case "backspace":
		if t.cursor > 0 {
			t.value = append(t.value[:t.cursor-1], t.value[t.cursor:]...)
			t.cursor--
		}
		return true
	case "delete":
		if t.cursor < len(t.value) {
			t.value = append(t.value[:t.cursor], t.value[t.cursor+1:]...)
		}
		return true
	case "left":
		if t.cursor > 0 {
			t.cursor--
		}
		return true
	case "right":
		if t.cursor < len(t.value) {
			t.cursor++
		}
		return true
	case "home":
		t.cursor = 0
		return true
	case "end":
		t.cursor = len(t.value)
		return true
	case "enter":
		if t.EventQueue != nil {
			t.EventQueue.Emit(t.PanelID, "submit", map[string]any{
				"value": string(t.value),
			})
		}
		return true
	default:
		// Single printable character
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			t.value = append(t.value, 0)
			copy(t.value[t.cursor+1:], t.value[t.cursor:])
			t.value[t.cursor] = rune(key[0])
			t.cursor++
			return true
		}
		// Multi-byte rune (e.g. from paste)
		runes := []rune(key)
		if len(runes) == 1 && runes[0] >= 32 {
			t.value = append(t.value, 0)
			copy(t.value[t.cursor+1:], t.value[t.cursor:])
			t.value[t.cursor] = runes[0]
			t.cursor++
			return true
		}
		return false
	}
}

func (t *TextInput) HandleMouse(x, y int, button string) bool {
	// Click to position cursor
	if button == "left" {
		pos := x
		if pos > len(t.value) {
			pos = len(t.value)
		}
		if pos < 0 {
			pos = 0
		}
		t.cursor = pos
		return true
	}
	return false
}

func (t *TextInput) View() string {
	if len(t.value) == 0 && t.placeholder != "" {
		// Show placeholder with cursor at start
		var b strings.Builder
		b.WriteRune('_')
		maxW := t.width - 1
		if maxW > len(t.placeholder) {
			maxW = len(t.placeholder)
		}
		if maxW > 0 {
			b.WriteString(t.placeholder[:maxW])
		}
		return b.String()
	}

	// Determine visible window of text
	visW := t.width
	if visW <= 0 {
		visW = 40
	}

	// Scroll so cursor is visible
	start := 0
	if t.cursor >= visW {
		start = t.cursor - visW + 1
	}
	end := start + visW
	if end > len(t.value) {
		end = len(t.value)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		if i == t.cursor {
			b.WriteRune('_') // cursor indicator
		}
		b.WriteRune(t.value[i])
	}
	// If cursor is at the end
	if t.cursor >= end {
		b.WriteRune('_')
	}

	return b.String()
}

func (t *TextInput) State() map[string]any {
	return map[string]any{
		"value":  string(t.value),
		"cursor": t.cursor,
	}
}

// SetValue sets the text input value programmatically.
func (t *TextInput) SetValue(s string) {
	t.value = []rune(s)
	if t.cursor > len(t.value) {
		t.cursor = len(t.value)
	}
}

// Value returns the current text.
func (t *TextInput) Value() string {
	return string(t.value)
}
