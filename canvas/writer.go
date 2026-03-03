package canvas

import (
	"fmt"
	"io"
	"strings"
)

// Writer renders a Canvas to an io.Writer using ANSI escape sequences.
// It performs differential updates — only writing cells that changed since
// the last frame.
type Writer struct {
	w       io.Writer
	profile ColorProfile
	prev    []Cell
	prevW   int
	prevH   int
}

// NewWriter creates a Writer targeting the given io.Writer with a color profile.
func NewWriter(w io.Writer, profile ColorProfile) *Writer {
	return &Writer{w: w, profile: profile}
}

// Render writes the canvas cell grid to the output, using differential updates.
func (wr *Writer) Render(c *Canvas) {
	// Detect resize — force full redraw
	if wr.prevW != c.CellW || wr.prevH != c.CellH {
		wr.prev = make([]Cell, c.CellW*c.CellH)
		wr.prevW = c.CellW
		wr.prevH = c.CellH
		// Mark all cells as "different" by leaving them zero-valued
	}

	var buf strings.Builder
	buf.Grow(c.CellW * c.CellH * 20) // rough estimate

	// Hide cursor during render
	buf.WriteString("\x1b[?25l")

	var lastFG, lastBG Color
	fgSet, bgSet := false, false

	for row := 0; row < c.CellH; row++ {
		for col := 0; col < c.CellW; col++ {
			idx := row*c.CellW + col
			cell := c.Cells[idx]

			// Skip unchanged cells
			if wr.prev[idx] == cell {
				fgSet, bgSet = false, false // invalidate color state after skip
				continue
			}
			wr.prev[idx] = cell

			// Move cursor (1-indexed)
			fmt.Fprintf(&buf, "\x1b[%d;%dH", row+1, col+1)

			// Apply colors (with degradation)
			fg := Degrade(cell.FG, wr.profile)
			bg := Degrade(cell.BG, wr.profile)

			if !fgSet || fg != lastFG {
				writeFG(&buf, fg, wr.profile)
				lastFG = fg
				fgSet = true
			}
			if !bgSet || bg != lastBG {
				writeBG(&buf, bg, wr.profile)
				lastBG = bg
				bgSet = true
			}

			// Write the rune
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			buf.WriteRune(r)
		}
	}

	// Reset attributes and show cursor
	buf.WriteString("\x1b[0m")

	io.WriteString(wr.w, buf.String())
}

// FullRender forces a complete redraw (clears diff state).
func (wr *Writer) FullRender(c *Canvas) {
	wr.prev = nil
	wr.prevW = 0
	wr.prevH = 0
	wr.Render(c)
}

// RenderString returns the ANSI string for the full canvas (no diffing).
func RenderString(c *Canvas, profile ColorProfile) string {
	var buf strings.Builder
	buf.Grow(c.CellW * c.CellH * 20)

	var lastFG, lastBG Color
	fgSet, bgSet := false, false

	for row := 0; row < c.CellH; row++ {
		if row > 0 {
			buf.WriteString("\n")
		}
		for col := 0; col < c.CellW; col++ {
			cell := c.Cells[row*c.CellW+col]

			fg := Degrade(cell.FG, profile)
			bg := Degrade(cell.BG, profile)

			if !fgSet || fg != lastFG {
				writeFG(&buf, fg, profile)
				lastFG = fg
				fgSet = true
			}
			if !bgSet || bg != lastBG {
				writeBG(&buf, bg, profile)
				lastBG = bg
				bgSet = true
			}

			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			buf.WriteRune(r)
		}
	}

	buf.WriteString("\x1b[0m")
	return buf.String()
}

func writeFG(buf *strings.Builder, c Color, profile ColorProfile) {
	if c.A <= 0 {
		buf.WriteString("\x1b[39m") // default fg
		return
	}
	r, g, b := c.ToRGB8()
	switch profile {
	case ProfileTrueColor:
		fmt.Fprintf(buf, "\x1b[38;2;%d;%d;%dm", r, g, b)
	case ProfileANSI256:
		fmt.Fprintf(buf, "\x1b[38;5;%dm", toANSI256Index(r, g, b))
	case ProfileANSI16:
		fmt.Fprintf(buf, "\x1b[%dm", toANSI16FG(r, g, b))
	}
}

func writeBG(buf *strings.Builder, c Color, profile ColorProfile) {
	if c.A <= 0 {
		buf.WriteString("\x1b[49m") // default bg
		return
	}
	r, g, b := c.ToRGB8()
	switch profile {
	case ProfileTrueColor:
		fmt.Fprintf(buf, "\x1b[48;2;%d;%d;%dm", r, g, b)
	case ProfileANSI256:
		fmt.Fprintf(buf, "\x1b[48;5;%dm", toANSI256Index(r, g, b))
	case ProfileANSI16:
		fmt.Fprintf(buf, "\x1b[%dm", toANSI16BG(r, g, b))
	}
}

func toANSI256Index(r, g, b uint8) int {
	// Check if it's close to a gray
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return 232 + int(float64(r-8)/247*24)
	}
	ri := int(float64(r) / 255 * 5)
	gi := int(float64(g) / 255 * 5)
	bi := int(float64(b) / 255 * 5)
	return 16 + 36*ri + 6*gi + bi
}

func toANSI16FG(r, g, b uint8) int {
	lum := (int(r) + int(g) + int(b)) / 3
	bright := lum > 128
	base := 30
	if bright {
		base = 90
	}
	// Map to nearest basic color
	ri := boolToInt(r > 128)
	gi := boolToInt(g > 128)
	bi := boolToInt(b > 128)
	return base + ri + 2*gi + 4*bi
}

func toANSI16BG(r, g, b uint8) int {
	return toANSI16FG(r, g, b) + 10
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
