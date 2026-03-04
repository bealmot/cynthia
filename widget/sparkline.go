package widget

import "strings"

// Sparkline renders a series of values as a compact bar chart using
// Unicode block elements: ▁▂▃▄▅▆▇█
type Sparkline struct {
	values []float64
	width  int
	label  string

	EventQueue *EventQueue
	PanelID    string
}

// Unicode bar characters from lowest to highest.
var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// NewSparkline creates a sparkline widget with the given display width.
func NewSparkline(width int, label string) *Sparkline {
	return &Sparkline{
		width: width,
		label: label,
	}
}

func (s *Sparkline) Type() string { return "sparkline" }

func (s *Sparkline) HandleKey(key string) bool   { return false }
func (s *Sparkline) HandleMouse(x, y int, button string) bool { return false }

func (s *Sparkline) View() string {
	var b strings.Builder

	if s.label != "" {
		b.WriteString(s.label)
		b.WriteString(" ")
	}

	if len(s.values) == 0 {
		b.WriteString(strings.Repeat("▁", s.width))
		return b.String()
	}

	// Find min/max for normalization.
	minV, maxV := s.values[0], s.values[0]
	for _, v := range s.values[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	// Show the last `width` values (or all if fewer).
	barWidth := s.width
	if s.label != "" {
		barWidth -= len(s.label) + 1
	}
	if barWidth < 1 {
		barWidth = 1
	}

	start := 0
	if len(s.values) > barWidth {
		start = len(s.values) - barWidth
	}

	rng := maxV - minV
	for i := start; i < len(s.values); i++ {
		var idx int
		if rng > 0 {
			norm := (s.values[i] - minV) / rng
			idx = int(norm * float64(len(sparkBlocks)-1))
			if idx >= len(sparkBlocks) {
				idx = len(sparkBlocks) - 1
			}
			if idx < 0 {
				idx = 0
			}
		}
		b.WriteRune(sparkBlocks[idx])
	}

	// Pad remaining width with low bars.
	rendered := len(s.values) - start
	for i := rendered; i < barWidth; i++ {
		b.WriteRune(sparkBlocks[0])
	}

	return b.String()
}

func (s *Sparkline) State() map[string]any {
	return map[string]any{
		"values": s.values,
		"count":  len(s.values),
		"label":  s.label,
	}
}

// Push appends a value to the sparkline data.
func (s *Sparkline) Push(v float64) {
	s.values = append(s.values, v)
	// Keep at most 1000 values to prevent unbounded growth.
	if len(s.values) > 1000 {
		s.values = s.values[len(s.values)-1000:]
	}
}

// SetValues replaces all sparkline data.
func (s *Sparkline) SetValues(vals []float64) {
	s.values = make([]float64, len(vals))
	copy(s.values, vals)
}
