package canvas

// Attrs holds terminal cell attributes as a bitmask.
type Attrs uint8

const (
	AttrBold      Attrs = 1 << iota
	AttrItalic
	AttrUnderline
	AttrBlink
	AttrReverse
)

// Cell represents a single terminal character cell.
type Cell struct {
	Rune rune
	FG   Color
	BG   Color
	Attr Attrs
}

// IsTransparent returns true if both fg and bg are fully transparent.
func (c Cell) IsTransparent() bool {
	return c.FG.A <= 0 && c.BG.A <= 0 && c.Rune == 0
}

// EmptyCell returns a transparent cell.
func EmptyCell() Cell {
	return Cell{Rune: ' ', FG: Transparent, BG: Transparent}
}
