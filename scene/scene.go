package scene

import (
	"sort"
	"sync/atomic"
)

// idCounter generates unique panel IDs when none is provided.
var idCounter atomic.Uint64

// GenerateID returns a unique panel ID.
func GenerateID() string {
	n := idCounter.Add(1)
	return "panel-" + uitoa(n)
}

// uitoa is a minimal uint64 to string without fmt.
func uitoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Scene holds the set of panels that make up the current display.
type Scene struct {
	panels    map[string]*Panel
	FocusedID string // ID of the currently focused widget panel
}

// New creates an empty scene.
func New() *Scene {
	return &Scene{panels: make(map[string]*Panel)}
}

// Add inserts a panel into the scene. If a panel with the same ID exists, it is replaced.
func (s *Scene) Add(p *Panel) {
	s.panels[p.ID] = p
}

// Remove removes a panel by ID. Returns true if it existed.
func (s *Scene) Remove(id string) bool {
	_, ok := s.panels[id]
	delete(s.panels, id)
	return ok
}

// Get returns a panel by ID, or nil.
func (s *Scene) Get(id string) *Panel {
	return s.panels[id]
}

// Panels returns all panels sorted by Z-order (lowest first).
func (s *Scene) Panels() []*Panel {
	out := make([]*Panel, 0, len(s.panels))
	for _, p := range s.panels {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Z != out[j].Z {
			return out[i].Z < out[j].Z
		}
		return out[i].ID < out[j].ID // stable tie-break
	})
	return out
}

// Count returns the number of panels in the scene.
func (s *Scene) Count() int {
	return len(s.panels)
}

// Clear removes all panels.
func (s *Scene) Clear() {
	clear(s.panels)
	s.FocusedID = ""
}

// Focus sets focus to the panel with the given ID. Unfocuses the previous panel.
func (s *Scene) Focus(id string) {
	if old := s.panels[s.FocusedID]; old != nil {
		old.Focused = false
	}
	s.FocusedID = id
	if p := s.panels[id]; p != nil {
		p.Focused = true
	}
}

// FocusedPanel returns the currently focused panel, or nil.
func (s *Scene) FocusedPanel() *Panel {
	if s.FocusedID == "" {
		return nil
	}
	return s.panels[s.FocusedID]
}

// FocusNext cycles focus to the next widget panel (by Z then ID order).
func (s *Scene) FocusNext() {
	panels := s.widgetPanels()
	if len(panels) == 0 {
		return
	}

	// Find current index
	current := -1
	for i, p := range panels {
		if p.ID == s.FocusedID {
			current = i
			break
		}
	}

	next := (current + 1) % len(panels)
	s.Focus(panels[next].ID)
}

// widgetPanels returns visible panels that have widgets, sorted by Z then ID.
func (s *Scene) widgetPanels() []*Panel {
	all := s.Panels() // already sorted by Z, ID
	out := make([]*Panel, 0)
	for _, p := range all {
		if p.Widget != nil && p.Visible {
			out = append(out, p)
		}
	}
	return out
}

// HitTest finds the topmost visible panel containing the cell coordinate (x, y).
// Searches in reverse Z order (highest Z first).
func (s *Scene) HitTest(x, y int) *Panel {
	panels := s.Panels()
	// Reverse iterate: highest Z (last in sorted slice) = topmost
	for i := len(panels) - 1; i >= 0; i-- {
		p := panels[i]
		if !p.Visible {
			continue
		}
		px, py := p.CellX(), p.CellY()
		if x >= px && x < px+p.Width && y >= py && y < py+p.Height {
			return p
		}
	}
	return nil
}
