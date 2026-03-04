package canvas

import "math"

// RenderMode determines how pixels are mapped to terminal cells.
type RenderMode int

const (
	ModeHalfBlock RenderMode = iota // ▀▄ — 2 vertical pixels per cell
	ModeBraille                     // ⠁-⣿ — 2×4 dot matrix per cell
	ModeASCII                       // density glyphs with color
)

// Canvas is a pixel buffer that rasterizes to terminal cells.
type Canvas struct {
	// Pixel buffer — dimensions are the pixel resolution,
	// which depends on the render mode.
	Width, Height int
	Pixels        []Color

	// Alpha mask — same dimensions as Pixels. nil = no mask.
	// Each value in [0,1] multiplies the corresponding pixel's alpha.
	Mask []float64

	// Cell grid — terminal dimensions
	CellW, CellH int
	Cells         []Cell

	Mode RenderMode
}

// New creates a canvas for the given terminal dimensions and render mode.
// Pixel resolution is derived from the mode:
//   - HalfBlock: cols × rows*2
//   - Braille:   cols*2 × rows*4
//   - ASCII:     cols × rows
func New(cols, rows int, mode RenderMode) *Canvas {
	c := &Canvas{Mode: mode, CellW: cols, CellH: rows}
	switch mode {
	case ModeHalfBlock:
		c.Width, c.Height = cols, rows*2
	case ModeBraille:
		c.Width, c.Height = cols*2, rows*4
	case ModeASCII:
		c.Width, c.Height = cols, rows
	}
	c.Pixels = make([]Color, c.Width*c.Height)
	c.Cells = make([]Cell, cols*rows)
	return c
}

// Resize changes the canvas dimensions, reallocating buffers.
func (c *Canvas) Resize(cols, rows int) {
	*c = *New(cols, rows, c.Mode)
}

// SetMask sets an alpha mask buffer. Must match pixel dimensions (Width*Height).
func (c *Canvas) SetMask(mask []float64) {
	if len(mask) == c.Width*c.Height {
		c.Mask = mask
	}
}

// ClearMask removes the mask.
func (c *Canvas) ClearMask() {
	c.Mask = nil
}

// ApplyMask multiplies each pixel's channels by the corresponding mask value.
func (c *Canvas) ApplyMask() {
	if c.Mask == nil {
		return
	}
	for i, m := range c.Mask {
		if i >= len(c.Pixels) {
			break
		}
		c.Pixels[i].R *= m
		c.Pixels[i].G *= m
		c.Pixels[i].B *= m
		c.Pixels[i].A *= m
	}
}

// SetPixel sets a pixel color at (x, y) in pixel coordinates.
func (c *Canvas) SetPixel(x, y int, col Color) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return
	}
	c.Pixels[y*c.Width+x] = col
}

// GetPixel returns the pixel color at (x, y).
func (c *Canvas) GetPixel(x, y int) Color {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return Transparent
	}
	return c.Pixels[y*c.Width+x]
}

// Clear fills the entire pixel buffer with the given color.
func (c *Canvas) Clear(col Color) {
	for i := range c.Pixels {
		c.Pixels[i] = col
	}
}

// Fill fills a rectangular region of pixels.
func (c *Canvas) Fill(x, y, w, h int, col Color) {
	for py := y; py < y+h && py < c.Height; py++ {
		for px := x; px < x+w && px < c.Width; px++ {
			if px >= 0 && py >= 0 {
				c.Pixels[py*c.Width+px] = col
			}
		}
	}
}

// SetCell directly sets a terminal cell at (col, row) in cell coordinates.
func (c *Canvas) SetCell(col, row int, cell Cell) {
	if col < 0 || col >= c.CellW || row < 0 || row >= c.CellH {
		return
	}
	c.Cells[row*c.CellW+col] = cell
}

// GetCell returns the cell at (col, row).
func (c *Canvas) GetCell(col, row int) Cell {
	if col < 0 || col >= c.CellW || row < 0 || row >= c.CellH {
		return EmptyCell()
	}
	return c.Cells[row*c.CellW+col]
}

// Rasterize converts the pixel buffer into the cell grid based on the render mode.
func (c *Canvas) Rasterize() {
	switch c.Mode {
	case ModeHalfBlock:
		c.rasterizeHalfBlock()
	case ModeBraille:
		c.rasterizeBraille()
	case ModeASCII:
		c.rasterizeASCII()
	}
}

// rasterizeHalfBlock maps pairs of vertical pixels to cells using ▀.
// FG = top pixel color, BG = bottom pixel color.
func (c *Canvas) rasterizeHalfBlock() {
	for row := 0; row < c.CellH; row++ {
		for col := 0; col < c.CellW; col++ {
			topY := row * 2
			botY := topY + 1

			top := c.GetPixel(col, topY)
			bot := c.GetPixel(col, botY)

			c.Cells[row*c.CellW+col] = Cell{
				Rune: '▀',
				FG:   top,
				BG:   bot,
			}
		}
	}
}

// ASCII density ramp (space → densest)
var asciiDensity = []rune{' ', '.', ':', '-', '=', '+', '*', '#', '%', '@'}

func (c *Canvas) rasterizeASCII() {
	for row := 0; row < c.CellH; row++ {
		for col := 0; col < c.CellW; col++ {
			px := c.GetPixel(col, row)
			// luminance
			sr, sg, sb, _ := px.Straight()
			lum := 0.299*sr + 0.587*sg + 0.114*sb
			idx := int(math.Round(clamp01(lum) * float64(len(asciiDensity)-1)))

			c.Cells[row*c.CellW+col] = Cell{
				Rune: asciiDensity[idx],
				FG:   px,
				BG:   Transparent,
			}
		}
	}
}

// Braille: each cell is a 2×4 dot matrix. We threshold luminance.
// Braille base: U+2800, bits map as:
//
//	  col0 col1
//	0: 0x01 0x08
//	1: 0x02 0x10
//	2: 0x04 0x20
//	3: 0x40 0x80
func (c *Canvas) rasterizeBraille() {
	brailleBits := [4][2]rune{
		{0x01, 0x08},
		{0x02, 0x10},
		{0x04, 0x20},
		{0x40, 0x80},
	}

	for row := 0; row < c.CellH; row++ {
		for col := 0; col < c.CellW; col++ {
			var pattern rune
			var rSum, gSum, bSum float64
			var count float64

			for dy := 0; dy < 4; dy++ {
				for dx := 0; dx < 2; dx++ {
					px := c.GetPixel(col*2+dx, row*4+dy)
					sr, sg, sb, _ := px.Straight()
					lum := 0.299*sr + 0.587*sg + 0.114*sb
					if lum > 0.5 {
						pattern |= brailleBits[dy][dx]
					}
					rSum += sr
					gSum += sg
					bSum += sb
					count++
				}
			}

			avg := RGB(rSum/count, gSum/count, bSum/count)
			c.Cells[row*c.CellW+col] = Cell{
				Rune: 0x2800 + pattern,
				FG:   avg,
				BG:   Transparent,
			}
		}
	}
}
