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

	// Use the BubbleTea-powered runtime for interactive widget support.
	// BubbleTea reads input from the TTY and handles raw mode, mouse events,
	// and cleanup automatically. MCP continues on stdin/stdout.
	prog := runtime.NewProgram(tty, cols, rows, *fps, mode)

	// Ensure cleanup on exit.
	defer prog.Quit()

	// Start the BubbleTea render loop in the background.
	progErr := make(chan error, 1)
	go func() {
		progErr <- prog.Run()
	}()

	// Handle SIGWINCH for terminal resize.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		for range sigCh {
			if c, r, err := term.GetSize(int(tty.Fd())); err == nil {
				prog.Lock()
				prog.Resize(c, r)
				prog.Unlock()
			}
		}
	}()

	// Handle SIGINT/SIGTERM for clean shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run MCP server over stdio — blocks until client disconnects or signal.
	srv := mcpserver.NewWithProgram(prog)
	mcpErr := srv.Run(ctx)

	// Stop the BubbleTea program when MCP disconnects.
	prog.Quit()

	// If MCP exited cleanly, check if BubbleTea had an error.
	if mcpErr == nil {
		select {
		case err := <-progErr:
			return err
		default:
		}
	}

	return mcpErr
}
