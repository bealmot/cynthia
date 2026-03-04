package effect

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bealmot/cynthia/canvas"
)

// PostProcessor is an optional interface for effects that can transform
// existing pixel content rather than generating from scratch. Effects
// that implement this (e.g. CRT, glitch) are used as post-process stages
// in an EffectChain.
type PostProcessor interface {
	RenderPost(c *canvas.Canvas)
}

// EffectChain pipelines multiple effects within a single panel.
// Stage 0 renders normally; stages 1+ use RenderPost if available,
// otherwise fall back to Render (which may overwrite previous output).
type EffectChain struct {
	stages []Effect
}

// NewChain creates an effect chain from the given effects.
func NewChain(effects ...Effect) *EffectChain {
	return &EffectChain{stages: effects}
}

func (ec *EffectChain) Name() string {
	names := make([]string, len(ec.stages))
	for i, s := range ec.stages {
		names[i] = s.Name()
	}
	return fmt.Sprintf("chain(%s)", strings.Join(names, ","))
}

func (ec *EffectChain) Step(frame uint64, dt float64) {
	for _, s := range ec.stages {
		s.Step(frame, dt)
	}
}

func (ec *EffectChain) Render(c *canvas.Canvas) {
	for i, s := range ec.stages {
		if i == 0 {
			s.Render(c)
		} else if pp, ok := s.(PostProcessor); ok {
			pp.RenderPost(c)
		} else {
			s.Render(c)
		}
	}
}

// SetParams routes parameters by numeric prefix to individual stages.
// "0.speed" → stage 0's "speed", "1.scanlines" → stage 1's "scanlines".
// Unprefixed params are forwarded to all stages.
func (ec *EffectChain) SetParams(params map[string]float64) {
	for k, v := range params {
		if dot := strings.IndexByte(k, '.'); dot > 0 {
			idx, err := strconv.Atoi(k[:dot])
			if err == nil && idx >= 0 && idx < len(ec.stages) {
				ec.stages[idx].SetParams(map[string]float64{k[dot+1:]: v})
			}
		} else {
			// Unprefixed: forward to all stages.
			for _, s := range ec.stages {
				s.SetParams(map[string]float64{k: v})
			}
		}
	}
}

// Params merges all stage params with numeric prefixes.
func (ec *EffectChain) Params() map[string]float64 {
	out := make(map[string]float64)
	for i, s := range ec.stages {
		prefix := strconv.Itoa(i) + "."
		for k, v := range s.Params() {
			out[prefix+k] = v
		}
	}
	return out
}
