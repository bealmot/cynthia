package director

// PanelDirector manages per-panel spring interpolation for smooth parameter transitions.
// Unlike the global Director, it has no mood/effect/overlay logic — it's just a bag of
// springs that interpolates toward target values each frame.
type PanelDirector struct {
	springs map[string]*SpringParam
	speed   float64
}

// NewPanelDirector creates a per-panel spring controller.
// speed is a frequency multiplier (higher = snappier transitions).
func NewPanelDirector(speed float64) *PanelDirector {
	if speed <= 0 {
		speed = 1.0
	}
	return &PanelDirector{
		springs: make(map[string]*SpringParam),
		speed:   speed,
	}
}

// SetTarget sets a spring target for the named parameter.
// Creates a new spring (starting at the target value) if one doesn't exist.
func (pd *PanelDirector) SetTarget(name string, value float64) {
	sp, ok := pd.springs[name]
	if !ok {
		freq := defaultFrequency * pd.speed
		sp = NewSpringParam(value, freq, defaultDamping)
		pd.springs[name] = sp
	}
	sp.SetTarget(value)
}

// SetTargets sets multiple spring targets at once.
func (pd *PanelDirector) SetTargets(params map[string]float64) {
	for k, v := range params {
		pd.SetTarget(k, v)
	}
}

// SetSpeed updates the transition speed multiplier.
func (pd *PanelDirector) SetSpeed(speed float64) {
	if speed > 0 {
		pd.speed = speed
	}
}

// Step advances all springs by dt and returns the interpolated parameter map.
func (pd *PanelDirector) Step(dt float64) map[string]float64 {
	params := make(map[string]float64, len(pd.springs))
	for name, sp := range pd.springs {
		params[name] = sp.Update(dt)
	}
	return params
}

// AtRest returns true if all springs have settled.
func (pd *PanelDirector) AtRest() bool {
	for _, sp := range pd.springs {
		if !sp.AtRest(0.001) {
			return false
		}
	}
	return true
}
