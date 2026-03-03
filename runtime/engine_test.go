package runtime

import (
	"bytes"
	"testing"

	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/scene"
)

func TestEngineNew(t *testing.T) {
	var buf bytes.Buffer
	e := NewEngine(&buf, 80, 24, 30, canvas.ModeHalfBlock)

	if e.Cols() != 80 || e.Rows() != 24 {
		t.Errorf("size = %dx%d, want 80x24", e.Cols(), e.Rows())
	}
	if e.Scene == nil {
		t.Fatal("Scene is nil")
	}
	if e.Compositor == nil {
		t.Fatal("Compositor is nil")
	}
}

func TestEngineStep(t *testing.T) {
	var buf bytes.Buffer
	e := NewEngine(&buf, 40, 12, 30, canvas.ModeHalfBlock)

	// Add a visible panel
	p := scene.NewPanel("test", 10, 5, canvas.ModeHalfBlock)
	p.Canvas.Clear(canvas.RGB(1, 0, 0))
	e.Scene.Add(p)

	// Step should render without panic
	e.Step()

	if buf.Len() == 0 {
		t.Error("expected output after Step, got nothing")
	}
}

func TestEngineStartStop(t *testing.T) {
	var buf bytes.Buffer
	e := NewEngine(&buf, 40, 12, 60, canvas.ModeHalfBlock)

	e.Start()
	// Double start should be safe
	e.Start()

	e.Stop()
	// Double stop should be safe
	e.Stop()
}

func TestEngineResize(t *testing.T) {
	var buf bytes.Buffer
	e := NewEngine(&buf, 40, 12, 30, canvas.ModeHalfBlock)

	e.Lock()
	e.Resize(100, 50)
	e.Unlock()

	if e.Cols() != 100 || e.Rows() != 50 {
		t.Errorf("after resize = %dx%d, want 100x50", e.Cols(), e.Rows())
	}
}

func TestEngineConcurrentAccess(t *testing.T) {
	var buf bytes.Buffer
	e := NewEngine(&buf, 40, 12, 30, canvas.ModeHalfBlock)

	done := make(chan bool)

	// Simulate concurrent MCP handler
	go func() {
		for i := 0; i < 50; i++ {
			e.Lock()
			p := scene.NewPanel(scene.GenerateID(), 5, 3, canvas.ModeHalfBlock)
			p.Canvas.Clear(canvas.RGB(0, 1, 0))
			e.Scene.Add(p)
			e.Unlock()
		}
		done <- true
	}()

	// Concurrent stepping
	go func() {
		for i := 0; i < 50; i++ {
			e.Step()
		}
		done <- true
	}()

	<-done
	<-done
	// No race/panic = success
}
