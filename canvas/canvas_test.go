package canvas

import (
	"testing"
)

func TestHex(t *testing.T) {
	c := Hex("#FF8000")
	r, g, b := c.ToRGB8()
	if r != 255 || g != 128 || b != 0 {
		t.Errorf("Hex(#FF8000) = (%d,%d,%d), want (255,128,0)", r, g, b)
	}
}

func TestHexShort(t *testing.T) {
	c := Hex("#F80")
	r, g, b := c.ToRGB8()
	if r != 255 || g != 136 || b != 0 {
		t.Errorf("Hex(#F80) = (%d,%d,%d), want (255,136,0)", r, g, b)
	}
}

func TestRGB8(t *testing.T) {
	c := RGB8(100, 200, 50)
	r, g, b := c.ToRGB8()
	if r != 100 || g != 200 || b != 50 {
		t.Errorf("RGB8 roundtrip = (%d,%d,%d), want (100,200,50)", r, g, b)
	}
}

func TestLerp(t *testing.T) {
	a := RGB(0, 0, 0)
	b := RGB(1, 1, 1)
	mid := a.Lerp(b, 0.5)
	if mid.R < 0.49 || mid.R > 0.51 {
		t.Errorf("Lerp midpoint R = %f, want ~0.5", mid.R)
	}
}

func TestPaletteSample(t *testing.T) {
	p := Palette{RGB(1, 0, 0), RGB(0, 1, 0), RGB(0, 0, 1)}

	// t=0 → red
	c := p.Sample(0)
	if c.R < 0.99 {
		t.Errorf("Sample(0) R = %f, want ~1.0", c.R)
	}

	// t=0.5 → green
	c = p.Sample(0.5)
	if c.G < 0.99 {
		t.Errorf("Sample(0.5) G = %f, want ~1.0", c.G)
	}

	// t=1 → blue
	c = p.Sample(1)
	if c.B < 0.99 {
		t.Errorf("Sample(1) B = %f, want ~1.0", c.B)
	}
}

func TestOverComposite(t *testing.T) {
	bg := RGB(0, 0, 1) // blue
	fg := RGBA(1, 0, 0, 0.5) // 50% red
	result := fg.Over(bg)

	// Expected: premultiplied result
	// src = (0.5, 0, 0, 0.5), dst = (0, 0, 1, 1)
	// result = src + dst * (1 - 0.5) = (0.5, 0, 0.5, 1)
	if result.A < 0.99 {
		t.Errorf("Over alpha = %f, want ~1.0", result.A)
	}
	if result.R < 0.49 || result.R > 0.51 {
		t.Errorf("Over R = %f, want ~0.5", result.R)
	}
}

func TestCanvasNewHalfBlock(t *testing.T) {
	c := New(80, 24, ModeHalfBlock)
	if c.Width != 80 || c.Height != 48 {
		t.Errorf("HalfBlock dims = %dx%d, want 80x48", c.Width, c.Height)
	}
	if len(c.Pixels) != 80*48 {
		t.Errorf("pixel count = %d, want %d", len(c.Pixels), 80*48)
	}
}

func TestCanvasRasterizeHalfBlock(t *testing.T) {
	c := New(2, 1, ModeHalfBlock)
	// Set top pixel red, bottom pixel blue
	c.SetPixel(0, 0, RGB(1, 0, 0))
	c.SetPixel(0, 1, RGB(0, 0, 1))
	c.Rasterize()

	cell := c.GetCell(0, 0)
	if cell.Rune != '▀' {
		t.Errorf("rune = %c, want ▀", cell.Rune)
	}
	if cell.FG.R < 0.99 {
		t.Errorf("FG should be red, got R=%f", cell.FG.R)
	}
	if cell.BG.B < 0.99 {
		t.Errorf("BG should be blue, got B=%f", cell.BG.B)
	}
}

func TestCanvasRasterizeBraille(t *testing.T) {
	c := New(1, 1, ModeBraille)
	// Pixel resolution: 2x4
	// Light up all pixels → should be full braille block ⣿ (0x28FF)
	for y := 0; y < 4; y++ {
		for x := 0; x < 2; x++ {
			c.SetPixel(x, y, RGB(1, 1, 1))
		}
	}
	c.Rasterize()
	cell := c.GetCell(0, 0)
	if cell.Rune != '⣿' {
		t.Errorf("full braille = %c (U+%04X), want ⣿ (U+28FF)", cell.Rune, cell.Rune)
	}
}

func TestCanvasRasterizeASCII(t *testing.T) {
	c := New(1, 1, ModeASCII)
	c.SetPixel(0, 0, RGB(1, 1, 1)) // max brightness
	c.Rasterize()
	cell := c.GetCell(0, 0)
	if cell.Rune != '@' {
		t.Errorf("max brightness ASCII = %c, want @", cell.Rune)
	}
}
