// Command cynthia-mcp is a distributable MCP server that exposes Cynthia's
// scene graph engine to AI agents. It communicates via stdio (JSON-RPC) and
// renders to a separate terminal.
//
// Usage:
//
//	cynthia-mcp --tty /dev/ttys003
//
// The --tty flag specifies which terminal to render to. Find it by running
// `tty` in the target terminal window. If omitted, falls back to /dev/tty
// (the controlling terminal — only works when not spawned by another process).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/mcpserver"
	"github.com/bealmot/cynthia/runtime"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "cynthia-mcp: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ttyPath := flag.String("tty", "/dev/tty", "TTY device to render to (run `tty` in the target terminal)")
	fps := flag.Int("fps", 30, "render frames per second")
	flag.Parse()

	// Open the target tty for rendering — stdio is reserved for MCP JSON-RPC.
	tty, err := os.OpenFile(*ttyPath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w (run `tty` in the target terminal to get the path)", *ttyPath, err)
	}
	defer tty.Close()

	// Get terminal size from the tty fd.
	cols, rows, err := term.GetSize(int(tty.Fd()))
	if err != nil {
		cols, rows = 80, 24
	}

	mode := canvas.ModeHalfBlock
	engine := runtime.NewEngine(tty, cols, rows, *fps, mode)

	// Enter alt screen and hide cursor on the tty.
	fmt.Fprint(tty, "\x1b[?1049h\x1b[?25l")

	// Ensure cleanup on exit.
	defer func() {
		engine.Stop()
		fmt.Fprint(tty, "\x1b[?25h\x1b[?1049l")
	}()

	// Start the render loop.
	engine.Start()

	// Handle SIGWINCH for terminal resize.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		for range sigCh {
			if c, r, err := term.GetSize(int(tty.Fd())); err == nil {
				engine.Lock()
				engine.Resize(c, r)
				engine.Unlock()
			}
		}
	}()

	// Handle SIGINT/SIGTERM for clean shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run MCP server over stdio — blocks until client disconnects or signal.
	srv := mcpserver.New(engine)
	return srv.Run(ctx)
}
