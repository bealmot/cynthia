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
	panels map[string]*Panel
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
}
