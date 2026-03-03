// Package director provides LLM art direction with spring-interpolated transitions.
package director

import "github.com/charmbracelet/harmonica"

// SpringParam wraps a harmonica spring for smooth per-parameter interpolation.
type SpringParam struct {
	spring harmonica.Spring
	pos    float64
	vel    float64
	target float64
}

// NewSpringParam creates a spring with the given initial value.
// frequency: oscillations per second (higher = snappier)
// damping: 1.0 = critically damped (no overshoot), <1 = bouncy
func NewSpringParam(initial, frequency, damping float64) *SpringParam {
	return &SpringParam{
		spring: harmonica.NewSpring(harmonica.FPS(60), frequency, damping),
		pos:    initial,
		target: initial,
	}
}

// SetTarget sets the desired value the spring will animate toward.
func (s *SpringParam) SetTarget(target float64) {
	s.target = target
}

// Update advances the spring by dt seconds and returns the current value.
func (s *SpringParam) Update(dt float64) float64 {
	s.pos, s.vel = s.spring.Update(s.pos, s.vel, s.target)
	return s.pos
}

// Value returns the current position without advancing.
func (s *SpringParam) Value() float64 {
	return s.pos
}

// AtRest returns true if the spring is within epsilon of its target.
func (s *SpringParam) AtRest(epsilon float64) bool {
	diff := s.pos - s.target
	if diff < 0 {
		diff = -diff
	}
	velAbs := s.vel
	if velAbs < 0 {
		velAbs = -velAbs
	}
	return diff < epsilon && velAbs < epsilon
}
