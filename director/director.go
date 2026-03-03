package director

import (
	"github.com/bealmot/cynthia/effect"
)

const (
	defaultFrequency = 4.0 // snappy but smooth
	defaultDamping   = 1.0 // critically damped
)

// Director manages visual state transitions driven by LLM directives.
// Numeric parameters are spring-interpolated; effect/palette switches are immediate.
type Director struct {
	// Current visual state
	CurrentEffect effect.Effect
	CurrentMood   string
	BorderStyle   string

	// Spring-managed parameters
	springs map[string]*SpringParam

	// Transition speed multiplier (affects spring frequency)
	transitionSpeed float64

	// Overlay tracking
	OverlayEffect   effect.Effect
	overlayRemain   float64
}

// New creates a Director with default calm mood.
func New() *Director {
	d := &Director{
		springs:         make(map[string]*SpringParam),
		transitionSpeed: 1.0,
	}
	d.ApplyMood("calm")
	return d
}

// ApplyMood applies a named mood preset.
func (d *Director) ApplyMood(name string) {
	preset, ok := Moods[name]
	if !ok {
		return
	}
	d.CurrentMood = name

	// Switch effect
	d.switchEffect(preset.Effect)

	// Set border
	d.BorderStyle = preset.BorderStyle

	// Spring-interpolate intensity
	d.setSpringTarget("intensity", preset.Intensity)

	// Spring-interpolate all preset params
	for k, v := range preset.Params {
		d.setSpringTarget(k, v)
	}
}

// Apply processes an LLM directive, switching effects immediately and
// spring-interpolating numeric parameters.
func (d *Director) Apply(dir Directive) {
	// Apply mood first (provides defaults)
	if dir.Mood != "" {
		d.ApplyMood(dir.Mood)
	}

	// Override effect if explicitly set
	if dir.Effect != "" {
		d.switchEffect(dir.Effect)
	}

	// Override intensity
	if dir.Intensity != nil {
		d.setSpringTarget("intensity", *dir.Intensity)
	}

	// Override transition speed
	if dir.TransitionSpeed != nil {
		d.transitionSpeed = *dir.TransitionSpeed
	}

	// Override border
	if dir.BorderStyle != "" {
		d.BorderStyle = dir.BorderStyle
	}

	// Override individual params
	for k, v := range dir.Params {
		d.setSpringTarget(k, v)
	}

	// Handle overlay
	if dir.Overlay != nil {
		d.OverlayEffect = effect.Create(dir.Overlay.Effect)
		d.overlayRemain = dir.Overlay.Duration
	}
}

// Step advances all springs and the overlay timer by dt seconds.
func (d *Director) Step(dt float64) {
	// Advance springs
	params := make(map[string]float64)
	for name, sp := range d.springs {
		params[name] = sp.Update(dt)
	}

	// Push interpolated params to the active effect
	if d.CurrentEffect != nil {
		d.CurrentEffect.SetParams(params)
	}

	// Tick down overlay
	if d.OverlayEffect != nil {
		d.overlayRemain -= dt
		if d.overlayRemain <= 0 {
			d.OverlayEffect = nil
			d.overlayRemain = 0
		}
	}
}

// Intensity returns the current spring-interpolated intensity value.
func (d *Director) Intensity() float64 {
	if sp, ok := d.springs["intensity"]; ok {
		return sp.Value()
	}
	return 0
}

// ParamValue returns the current spring-interpolated value for a named param.
func (d *Director) ParamValue(name string) float64 {
	if sp, ok := d.springs[name]; ok {
		return sp.Value()
	}
	return 0
}

func (d *Director) switchEffect(name string) {
	if d.CurrentEffect != nil && d.CurrentEffect.Name() == name {
		return
	}
	e := effect.Create(name)
	if e != nil {
		d.CurrentEffect = e
	}
}

func (d *Director) setSpringTarget(name string, target float64) {
	sp, ok := d.springs[name]
	if !ok {
		freq := defaultFrequency * d.transitionSpeed
		sp = NewSpringParam(target, freq, defaultDamping)
		d.springs[name] = sp
	}
	sp.SetTarget(target)
}
