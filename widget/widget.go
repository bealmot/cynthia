// Package widget provides interactive terminal widgets for Cynthia panels.
// Widgets handle keyboard and mouse input, render text representations,
// and expose state for MCP polling.
package widget

import (
	"sync"
	"time"
)

// Widget is the interface for interactive panel content.
// Widgets receive input events, produce text views for rendering,
// and expose state for MCP clients to read.
type Widget interface {
	// HandleKey processes a key event. Returns true if consumed.
	HandleKey(key string) bool
	// HandleMouse processes a mouse click within widget-local coordinates.
	// Returns true if consumed.
	HandleMouse(x, y int, button string) bool
	// View returns the current text representation for stamping into the panel.
	View() string
	// State returns the widget's current value for MCP reading.
	State() map[string]any
	// Type returns the widget type name ("text_input", "button").
	Type() string
}

// Event represents something that happened in a widget (button click, text submit).
// MCP clients poll for these since there's no server-push mechanism.
type Event struct {
	PanelID   string         `json:"panel_id"`
	Type      string         `json:"type"`       // "click", "submit", "change"
	Data      map[string]any `json:"data"`       // widget-specific payload
	Timestamp int64          `json:"timestamp"`  // unix millis
}

// EventQueue collects widget events for MCP polling.
type EventQueue struct {
	mu     sync.Mutex
	events []Event
}

// NewEventQueue creates an empty event queue.
func NewEventQueue() *EventQueue {
	return &EventQueue{}
}

// Push adds an event to the queue.
func (q *EventQueue) Push(e Event) {
	q.mu.Lock()
	q.events = append(q.events, e)
	q.mu.Unlock()
}

// Drain returns all queued events and clears the queue.
func (q *EventQueue) Drain() []Event {
	q.mu.Lock()
	out := q.events
	q.events = nil
	q.mu.Unlock()
	if out == nil {
		out = []Event{}
	}
	return out
}

// Emit is a convenience for pushing an event with the current timestamp.
func (q *EventQueue) Emit(panelID, eventType string, data map[string]any) {
	q.Push(Event{
		PanelID:   panelID,
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}
