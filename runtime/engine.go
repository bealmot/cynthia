// Package runtime provides the headless tick-loop engine for Cynthia's scene graph.
// The Engine composes panels from the scene graph and renders them to a writer,
// with thread-safe access for concurrent MCP tool handlers.
package runtime

import (
	"io"
	"sync"
	"time"

	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/compose"
	"github.com/bealmot/cynthia/scene"
)

// Engine drives the render loop: scene → compositor → writer.
type Engine struct {
	Scene      *scene.Scene
	Compositor *compose.Compositor
	Writer     *canvas.Writer
	FPS        int
	Mode       canvas.RenderMode

	cols, rows int
	mu         sync.Mutex
	running    bool
	stopCh     chan struct{}
	frame      uint64
}

// NewEngine creates an engine targeting the given writer.
func NewEngine(w io.Writer, cols, rows, fps int, mode canvas.RenderMode) *Engine {
	return &Engine{
		Scene:      scene.New(),
		Compositor: compose.NewCompositor(cols, rows, mode),
		Writer:     canvas.NewWriter(w, canvas.ProfileTrueColor),
		FPS:        fps,
		Mode:       mode,
		cols:       cols,
		rows:       rows,
	}
}

// Lock acquires the engine mutex. MCP handlers use this to safely mutate the scene.
func (e *Engine) Lock() { e.mu.Lock() }

// Unlock releases the engine mutex.
func (e *Engine) Unlock() { e.mu.Unlock() }

// Cols returns the current terminal column count.
func (e *Engine) Cols() int { return e.cols }

// Rows returns the current terminal row count.
func (e *Engine) Rows() int { return e.rows }

// Resize changes the output dimensions. Must be called under Lock.
func (e *Engine) Resize(cols, rows int) {
	e.cols = cols
	e.rows = rows
	e.Compositor.Resize(cols, rows)
}

// Step performs one render cycle: tick panel content, composite, rasterize, write.
func (e *Engine) Step() {
	dt := 1.0 / float64(e.FPS)

	e.mu.Lock()
	panels := e.Scene.Panels()
	frame := e.frame
	e.frame++

	// Tick and render each panel's content while holding the lock
	// (effects/borders mutate panel canvases).
	for _, p := range panels {
		if !p.Visible {
			continue
		}

		// Clear canvas for this frame
		if p.Fill.A > 0 {
			p.Canvas.Clear(p.Fill)
		} else {
			p.Canvas.Clear(canvas.Transparent)
		}

		// Procedural effect → pixel buffer
		if p.Effect != nil {
			p.Effect.Step(frame, dt)
			p.Effect.Render(p.Canvas)
		}

		// Border → cell grid
		if p.Border != nil {
			p.Border.Step(frame, dt)
			p.Border.Render(p.Canvas, 0, 0, p.Width, p.Height)
		}

		// Widget → dynamically update text content each frame
		if p.Widget != nil {
			p.Text = p.Widget.View()
		}

		// Text → cell grid (stamped over border interior)
		if p.Text != "" {
			stampText(p.Canvas, p.Text, p.Border != nil)
		}
	}

	// Convert to ScenePanel interface slice
	sp := make([]compose.ScenePanel, len(panels))
	for i, p := range panels {
		sp[i] = p
	}
	e.mu.Unlock()

	e.Compositor.ComposeScene(sp)
	e.Compositor.Output().Rasterize()
	e.Writer.Render(e.Compositor.Output())
}

// stampText writes an ANSI text string onto a panel's cell grid.
// If hasBorder is true, text is inset by 1 cell on each side.
func stampText(c *canvas.Canvas, text string, hasBorder bool) {
	startCol, startRow := 0, 0
	maxCol, maxRow := c.CellW, c.CellH
	if hasBorder {
		startCol, startRow = 1, 1
		maxCol--
		maxRow--
	}

	col, row := startCol, startRow
	for _, r := range text {
		if row >= maxRow {
			break
		}
		if r == '\n' {
			col = startCol
			row++
			continue
		}
		if col >= maxCol {
			continue
		}
		cell := c.GetCell(col, row)
		cell.Rune = r
		if cell.FG.A == 0 {
			cell.FG = canvas.White
		}
		c.SetCell(col, row, cell)
		col++
	}
}

// Start begins the render loop in a goroutine. It runs until Stop is called.
func (e *Engine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.stopCh = make(chan struct{})
	e.mu.Unlock()

	go e.loop()
}

// Stop halts the render loop and waits for it to finish.
func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	close(e.stopCh)
	e.mu.Unlock()
}

func (e *Engine) loop() {
	interval := time.Duration(float64(time.Second) / float64(e.FPS))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.Step()
		}
	}
}
