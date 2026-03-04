package widget

import (
	"strings"
)

// Table renders tabular data with box-drawing characters.
// Auto-sizes columns from content and truncates with "..." when needed.
type Table struct {
	headers []string
	rows    [][]string
	maxCols []int // computed column widths
	width   int

	EventQueue *EventQueue
	PanelID    string
}

// NewTable creates a table widget with the given headers.
func NewTable(headers []string, width int) *Table {
	t := &Table{
		headers: headers,
		width:   width,
	}
	t.computeWidths()
	return t
}

func (t *Table) Type() string { return "table" }

func (t *Table) HandleKey(key string) bool   { return false }
func (t *Table) HandleMouse(x, y int, button string) bool { return false }

func (t *Table) computeWidths() {
	cols := len(t.headers)
	if cols == 0 {
		return
	}

	t.maxCols = make([]int, cols)
	for i, h := range t.headers {
		if len(h) > t.maxCols[i] {
			t.maxCols[i] = len(h)
		}
	}
	for _, row := range t.rows {
		for i := 0; i < cols && i < len(row); i++ {
			if len(row[i]) > t.maxCols[i] {
				t.maxCols[i] = len(row[i])
			}
		}
	}

	// Clamp total width. If columns exceed available width, shrink proportionally.
	totalContent := 0
	for _, w := range t.maxCols {
		totalContent += w
	}
	// Account for borders: │ + content + │ between each col + │
	borderOverhead := cols + 1
	available := t.width - borderOverhead
	if available < cols {
		available = cols
	}

	if totalContent > available {
		scale := float64(available) / float64(totalContent)
		for i := range t.maxCols {
			t.maxCols[i] = int(float64(t.maxCols[i]) * scale)
			if t.maxCols[i] < 1 {
				t.maxCols[i] = 1
			}
		}
	}
}

func (t *Table) truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

func (t *Table) padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return t.truncate(s, width)
	}
	return s + strings.Repeat(" ", width-len(runes))
}

func (t *Table) View() string {
	if len(t.headers) == 0 {
		return "(empty table)"
	}

	t.computeWidths()
	var b strings.Builder
	cols := len(t.headers)

	// Top border: ┌─┬─┐
	b.WriteRune('┌')
	for i := 0; i < cols; i++ {
		b.WriteString(strings.Repeat("─", t.maxCols[i]))
		if i < cols-1 {
			b.WriteRune('┬')
		}
	}
	b.WriteRune('┐')
	b.WriteRune('\n')

	// Header row: │ header │
	b.WriteRune('│')
	for i, h := range t.headers {
		b.WriteString(t.padRight(h, t.maxCols[i]))
		b.WriteRune('│')
	}
	b.WriteRune('\n')

	// Header separator: ├─┼─┤
	b.WriteRune('├')
	for i := 0; i < cols; i++ {
		b.WriteString(strings.Repeat("─", t.maxCols[i]))
		if i < cols-1 {
			b.WriteRune('┼')
		}
	}
	b.WriteRune('┤')
	b.WriteRune('\n')

	// Data rows.
	for _, row := range t.rows {
		b.WriteRune('│')
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			b.WriteString(t.padRight(cell, t.maxCols[i]))
			b.WriteRune('│')
		}
		b.WriteRune('\n')
	}

	// Bottom border: └─┴─┘
	b.WriteRune('└')
	for i := 0; i < cols; i++ {
		b.WriteString(strings.Repeat("─", t.maxCols[i]))
		if i < cols-1 {
			b.WriteRune('┴')
		}
	}
	b.WriteRune('┘')

	return b.String()
}

func (t *Table) State() map[string]any {
	return map[string]any{
		"headers":   t.headers,
		"rows":      t.rows,
		"row_count": len(t.rows),
		"col_count": len(t.headers),
	}
}

// AddRow appends a data row to the table.
func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

// SetRows replaces all data rows.
func (t *Table) SetRows(rows [][]string) {
	t.rows = rows
}

// ClearRows removes all data rows (headers remain).
func (t *Table) ClearRows() {
	t.rows = nil
}
