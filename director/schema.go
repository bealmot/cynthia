package director

// Directive is the JSON structure the LLM outputs to control visual state.
type Directive struct {
	Mood            string             `json:"mood,omitempty"`
	Effect          string             `json:"effect,omitempty"`
	Intensity       *float64           `json:"intensity,omitempty"`
	Params          map[string]float64 `json:"params,omitempty"`
	BorderStyle     string             `json:"border_style,omitempty"`
	TransitionSpeed *float64           `json:"transition_speed,omitempty"`
	Overlay         *OverlayDirective  `json:"overlay,omitempty"`
}

// OverlayDirective controls transient overlay effects.
type OverlayDirective struct {
	Effect   string  `json:"effect"`
	Duration float64 `json:"duration"`
}
