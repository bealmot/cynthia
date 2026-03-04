package canvas

import "strings"

// ParseANSIToCells parses an ANSI-formatted string and places visible
// characters with their colors onto the canvas cell grid.
// Cells with no explicit BG color get Transparent, allowing composited
// layers underneath to show through.
func ParseANSIToCells(s string, c *Canvas) {
	col, row := 0, 0
	fg := Color{R: 1, G: 1, B: 1, A: 1} // default white
	bg := Transparent

	i := 0
	for i < len(s) {
		if s[i] == '\n' {
			row++
			col = 0
			i++
			continue
		}

		// Parse ANSI escape sequence
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Find the end of the escape sequence
			j := i + 2
			for j < len(s) && !isCSITerminator(s[j]) {
				j++
			}
			if j < len(s) {
				// Parse SGR (Select Graphic Rendition)
				if s[j] == 'm' {
					params := s[i+2 : j]
					fg, bg = parseSGR(params, fg, bg)
				}
				i = j + 1
			} else {
				i = j
			}
			continue
		}

		// Regular character — place it on the cell grid
		if col < c.CellW && row < c.CellH {
			r := rune(s[i])
			// Handle multi-byte UTF-8
			if r >= 0x80 {
				runes := []rune(s[i:])
				if len(runes) > 0 {
					r = runes[0]
					i += len(string(r)) - 1
				}
			}
			c.SetCell(col, row, Cell{
				Rune: r,
				FG:   fg,
				BG:   bg,
			})
		}
		col++
		i++
	}
}

func isCSITerminator(b byte) bool {
	return b >= 0x40 && b <= 0x7E
}

// parseSGR handles ANSI SGR (m) sequences to extract fg/bg colors.
func parseSGR(params string, fg, bg Color) (Color, Color) {
	if params == "" || params == "0" {
		return Color{R: 1, G: 1, B: 1, A: 1}, Transparent
	}

	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "0": // reset
			fg = Color{R: 1, G: 1, B: 1, A: 1}
			bg = Transparent
		case "38": // set fg
			if i+1 < len(parts) && parts[i+1] == "2" && i+4 < len(parts) {
				// 24-bit: 38;2;r;g;b
				r := parseUint8(parts[i+2])
				g := parseUint8(parts[i+3])
				b := parseUint8(parts[i+4])
				fg = RGB8(r, g, b)
				i += 4
			} else if i+1 < len(parts) && parts[i+1] == "5" && i+2 < len(parts) {
				// 256-color: 38;5;n — just use the value as-is
				i += 2
			}
		case "48": // set bg
			if i+1 < len(parts) && parts[i+1] == "2" && i+4 < len(parts) {
				r := parseUint8(parts[i+2])
				g := parseUint8(parts[i+3])
				b := parseUint8(parts[i+4])
				bg = RGB8(r, g, b)
				i += 4
			} else if i+1 < len(parts) && parts[i+1] == "5" && i+2 < len(parts) {
				i += 2
			}
		case "39": // default fg
			fg = Color{R: 1, G: 1, B: 1, A: 1}
		case "49": // default bg
			bg = Transparent
		}
	}
	return fg, bg
}

func parseUint8(s string) uint8 {
	var v int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + int(c-'0')
		}
	}
	if v > 255 {
		v = 255
	}
	return uint8(v)
}
