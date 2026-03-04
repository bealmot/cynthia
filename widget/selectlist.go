package widget

import (
	"fmt"
	"strings"
)

// SelectList is a scrollable selection list widget.
// Supports up/down navigation and enter to select.
type SelectList struct {
	items    []string
	cursor   int
	offset   int // scroll offset for windowing
	height   int // visible rows
	selected int // last selected index (-1 = none)

	EventQueue *EventQueue
	PanelID    string
}

// NewSelectList creates a selection list widget.
func NewSelectList(items []string, visibleRows int) *SelectList {
	return &SelectList{
		items:    items,
		height:   visibleRows,
		selected: -1,
	}
}

func (s *SelectList) Type() string { return "select_list" }

func (s *SelectList) HandleKey(key string) bool {
	switch key {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
			s.adjustScroll()
		}
		return true
	case "down", "j":
		if s.cursor < len(s.items)-1 {
			s.cursor++
			s.adjustScroll()
		}
		return true
	case "enter":
		if len(s.items) > 0 {
			s.selected = s.cursor
			if s.EventQueue != nil {
				s.EventQueue.Emit(s.PanelID, "select", map[string]any{
					"index": s.cursor,
					"value": s.items[s.cursor],
				})
			}
		}
		return true
	case "home":
		s.cursor = 0
		s.adjustScroll()
		return true
	case "end":
		s.cursor = len(s.items) - 1
		s.adjustScroll()
		return true
	}
	return false
}

func (s *SelectList) HandleMouse(x, y int, button string) bool {
	if button == "left" {
		idx := s.offset + y
		if idx >= 0 && idx < len(s.items) {
			s.cursor = idx
			s.selected = idx
			if s.EventQueue != nil {
				s.EventQueue.Emit(s.PanelID, "select", map[string]any{
					"index": idx,
					"value": s.items[idx],
				})
			}
		}
		return true
	}
	return false
}

func (s *SelectList) adjustScroll() {
	if s.cursor < s.offset {
		s.offset = s.cursor
	}
	if s.cursor >= s.offset+s.height {
		s.offset = s.cursor - s.height + 1
	}
}

func (s *SelectList) View() string {
	if len(s.items) == 0 {
		return "(empty)"
	}

	var b strings.Builder
	end := s.offset + s.height
	if end > len(s.items) {
		end = len(s.items)
	}

	for i := s.offset; i < end; i++ {
		if i > s.offset {
			b.WriteRune('\n')
		}
		if i == s.cursor {
			b.WriteString("> ")
		} else {
			b.WriteString("  ")
		}
		b.WriteString(s.items[i])
	}

	// Scroll indicator.
	if len(s.items) > s.height {
		b.WriteString(fmt.Sprintf(" [%d/%d]", s.cursor+1, len(s.items)))
	}

	return b.String()
}

func (s *SelectList) State() map[string]any {
	selectedValue := ""
	if s.selected >= 0 && s.selected < len(s.items) {
		selectedValue = s.items[s.selected]
	}
	return map[string]any{
		"cursor":         s.cursor,
		"selected_index": s.selected,
		"selected_value": selectedValue,
		"item_count":     len(s.items),
		"items":          s.items,
	}
}

// SetItems replaces the list items and resets cursor.
func (s *SelectList) SetItems(items []string) {
	s.items = items
	s.cursor = 0
	s.offset = 0
	s.selected = -1
}
