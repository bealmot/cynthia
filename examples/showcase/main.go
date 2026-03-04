// Showcase — Cynthia capabilities demo. Cycles through 4 visual demonstrations,
// each running ~10 seconds: diffused sphere (4 render modes), mirror cube,
// ASCII waves, and animated borders.
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bealmot/cynthia/canvas"
)

// ─── Constants ──────────────────────────────────────────────────────────────

const (
	targetFPS    = 30
	demoDuration = 10.0 // seconds per demo
)

type demoID int

const (
	demoSphere demoID = iota
	demoCube
	demoWaves
	demoBorders
	demoCount
)

var demoNames = [demoCount]string{
	"Diffused Sphere — Four Render Modes",
	"Mirror Cube — Rainbow Reflections",
	"Ocean Waves — ASCII Rendering",
	"Animated Borders — Dotted / Gradient / Colorshift",
}

// ─── Palettes ───────────────────────────────────────────────────────────────

var rainbow = canvas.Palette{
	canvas.Hex("#FF0000"),
	canvas.Hex("#FF8800"),
	canvas.Hex("#FFFF00"),
	canvas.Hex("#00FF00"),
	canvas.Hex("#0088FF"),
	canvas.Hex("#8800FF"),
	canvas.Hex("#FF0088"),
	canvas.Hex("#FF0000"),
}

var ocean = canvas.Palette{
	canvas.Hex("#0A1628"), // deep
	canvas.Hex("#0D2847"),
	canvas.Hex("#1A4A6E"),
	canvas.Hex("#2E7DAF"),
	canvas.Hex("#5BA8D9"),
	canvas.Hex("#A0D4EF"),
	canvas.Hex("#D0EEFF"),
	canvas.Hex("#FFFFFF"), // foam
}

var spherePal = canvas.Palette{
	canvas.Hex("#6E3FA0"),
	canvas.Hex("#A855F7"),
	canvas.Hex("#C084FC"),
	canvas.Hex("#E9D5FF"),
	canvas.Hex("#93E4D4"),
	canvas.Hex("#6E3FA0"),
}

// ─── Tick ───────────────────────────────────────────────────────────────────

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second/targetFPS, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ─── Model ──────────────────────────────────────────────────────────────────

type model struct {
	width, height int
	frame         uint64
	elapsed       float64 // total time
	demo          demoID
	demoTime      float64 // time within current demo
	quitting      bool
}

func initialModel() model {
	return model{width: 80, height: 24}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "n", "right", " ":
			m.demo = (m.demo + 1) % demoCount
			m.demoTime = 0
			return m, nil
		case "p", "left":
			m.demo = (m.demo - 1 + demoCount) % demoCount
			m.demoTime = 0
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		dt := 1.0 / float64(targetFPS)
		m.elapsed += dt
		m.demoTime += dt
		m.frame++
		if m.demoTime >= demoDuration {
			m.demo = (m.demo + 1) % demoCount
			m.demoTime = 0
		}
		return m, tick()
	}
	return m, nil
}

func (m model) View() tea.View {
	if m.quitting || m.width < 10 || m.height < 5 {
		return tea.NewView("")
	}

	var content string
	switch m.demo {
	case demoSphere:
		content = m.renderSphere()
	case demoCube:
		content = m.renderCube()
	case demoWaves:
		content = m.renderWaves()
	case demoBorders:
		content = m.renderBorders()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// ─── Demo 1: Diffused Sphere (4 render modes) ──────────────────────────────

func (m model) renderSphere() string {
	qw := m.width / 2
	qh := (m.height - 2) / 2 // reserve 2 rows for header/footer
	if qw < 4 || qh < 2 {
		return ""
	}

	modes := [4]canvas.RenderMode{
		canvas.ModeHalfBlock, canvas.ModeBraille,
		canvas.ModeASCII, canvas.ModeHalfBlock,
	}
	labels := [4]string{"Half-Block", "Braille", "ASCII", "Normal-Map"}

	// Build 4 canvases and render sphere into each
	canvases := [4]*canvas.Canvas{}
	for i := range canvases {
		canvases[i] = canvas.New(qw, qh, modes[i])
		if i == 3 {
			renderSphereNormalMap(canvases[i], m.demoTime)
		} else {
			renderSphereDiffuse(canvases[i], m.demoTime, modes[i])
		}
		canvases[i].Rasterize()
		drawCellLabel(canvases[i], 1, 0, labels[i], canvas.White)
	}

	// Compose quadrants into a single string
	var buf strings.Builder
	buf.Grow(m.width * m.height * 24)
	writeHeader(&buf, m.width, demoNames[demoSphere], m.demoTime)

	profile := canvas.ProfileTrueColor
	for row := 0; row < qh; row++ {
		writeQuadRow(&buf, canvases[0], canvases[1], row, qw, profile)
		buf.WriteString("\x1b[0m\n")
	}
	for row := 0; row < qh; row++ {
		writeQuadRow(&buf, canvases[2], canvases[3], row, qw, profile)
		if row < qh-1 {
			buf.WriteString("\x1b[0m\n")
		}
	}
	buf.WriteString("\x1b[0m")
	return buf.String()
}

func renderSphereDiffuse(c *canvas.Canvas, t float64, mode canvas.RenderMode) {
	cx := float64(c.Width) / 2
	cy := float64(c.Height) / 2

	// Aspect correction: ASCII pixels are 1:2 (cell shaped), others are ~1:1
	aspect := 1.0
	if mode == canvas.ModeASCII {
		aspect = 0.5
	}

	radius := math.Min(cx, cy*aspect) * 0.8

	// Rotating light
	la := t * 0.8
	light := v3norm(vec3{math.Cos(la), 0.6, math.Sin(la)})

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			nx := (float64(x) - cx) / radius
			ny := (float64(y) - cy) / (radius / aspect)

			d := nx*nx + ny*ny
			if d > 1.0 {
				c.SetPixel(x, y, canvas.RGB(0.02, 0.02, 0.04))
				continue
			}

			nz := math.Sqrt(1.0 - d)
			n := v3norm(vec3{nx, -ny, nz})
			diff := math.Max(0, v3dot(n, light))
			spec := math.Pow(math.Max(0, v3dot(v3reflect(light, n), vec3{0, 0, 1})), 32)

			base := spherePal.Sample(math.Mod(math.Atan2(n.Y, n.X)/math.Pi*0.5+0.5+t*0.1, 1.0))
			sr, sg, sb, _ := base.Straight()
			intensity := 0.08 + diff*0.75 + spec*0.4
			c.SetPixel(x, y, canvas.RGB(
				clamp01(sr*intensity),
				clamp01(sg*intensity),
				clamp01(sb*intensity),
			))
		}
	}
}

func renderSphereNormalMap(c *canvas.Canvas, t float64) {
	cx := float64(c.Width) / 2
	cy := float64(c.Height) / 2
	radius := math.Min(cx, cy) * 0.8

	// Slow rotation offset for the normal visualization
	rot := t * 0.4

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			nx := (float64(x) - cx) / radius
			ny := (float64(y) - cy) / radius

			d := nx*nx + ny*ny
			if d > 1.0 {
				c.SetPixel(x, y, canvas.RGB(0.02, 0.02, 0.04))
				continue
			}

			nz := math.Sqrt(1.0 - d)
			// Rotate normal around Y axis
			rnx := nx*math.Cos(rot) + nz*math.Sin(rot)
			rnz := -nx*math.Sin(rot) + nz*math.Cos(rot)

			// Map normal components to RGB
			c.SetPixel(x, y, canvas.RGB(
				clamp01(rnx*0.5+0.5),
				clamp01(-ny*0.5+0.5),
				clamp01(rnz*0.5+0.5),
			))
		}
	}
}

// ─── Demo 2: Mirror Cube ────────────────────────────────────────────────────

// Face definition: 4 vertex indices, drawn as a filled quad.
type face struct {
	v [4]int
}

var cubeVerts = [8]vec3{
	{-1, -1, -1}, {1, -1, -1}, {1, 1, -1}, {-1, 1, -1},
	{-1, -1, 1}, {1, -1, 1}, {1, 1, 1}, {-1, 1, 1},
}

var cubeFaces = [6]face{
	{[4]int{0, 1, 2, 3}}, // front
	{[4]int{5, 4, 7, 6}}, // back
	{[4]int{4, 0, 3, 7}}, // left
	{[4]int{1, 5, 6, 2}}, // right
	{[4]int{3, 2, 6, 7}}, // top
	{[4]int{4, 5, 1, 0}}, // bottom
}

func (m model) renderCube() string {
	c := canvas.New(m.width, m.height-2, canvas.ModeHalfBlock)

	// Rotation angles
	ay := m.demoTime * 0.7
	ax := m.demoTime * 0.4
	az := m.demoTime * 0.2

	// Transform vertices
	var proj [8]vec2
	var transformed [8]vec3
	cx := float64(c.Width) / 2
	cy := float64(c.Height) / 2
	scale := math.Min(cx, cy) * 0.35

	for i, v := range cubeVerts {
		r := rotateYXZ(v, ax, ay, az)
		transformed[i] = r
		// Perspective projection
		z := r.Z + 4.0 // push back
		proj[i] = vec2{
			X: cx + r.X/z*scale*3,
			Y: cy - r.Y/z*scale*3,
		}
	}

	// Sort faces back-to-front by average Z
	type faceSort struct {
		idx  int
		avgZ float64
	}
	sorted := make([]faceSort, 6)
	for i, f := range cubeFaces {
		z := 0.0
		for _, vi := range f.v {
			z += transformed[vi].Z
		}
		sorted[i] = faceSort{i, z / 4}
	}
	// Simple bubble sort (only 6 items)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].avgZ < sorted[i].avgZ {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Clear to dark background
	c.Clear(canvas.RGB(0.02, 0.01, 0.03))

	// Draw faces back-to-front
	for _, fs := range sorted {
		f := cubeFaces[fs.idx]
		// Face normal (cross product of two edges)
		e1 := v3sub(transformed[f.v[1]], transformed[f.v[0]])
		e2 := v3sub(transformed[f.v[3]], transformed[f.v[0]])
		normal := v3norm(v3cross(e1, e2))

		// Backface cull
		viewDir := v3norm(vec3{0, 0, -1})
		if v3dot(normal, viewDir) > 0 {
			continue
		}

		// Rainbow reflection: map reflected view direction to palette
		reflected := v3reflect(vec3{0, 0, 1}, normal)
		envT := math.Mod((math.Atan2(reflected.Y, reflected.X)+math.Pi)/(2*math.Pi)+m.demoTime*0.15, 1.0)
		baseColor := rainbow.Sample(envT)

		// Diffuse shading from a white light above-right
		light := v3norm(vec3{0.5, 0.8, -0.3})
		diff := math.Max(0, v3dot(normal, light))
		ambient := 0.15
		spec := math.Pow(math.Max(0, v3dot(v3reflect(light, normal), vec3{0, 0, 1})), 64)

		sr, sg, sb, _ := baseColor.Straight()
		intensity := ambient + diff*0.6
		r := clamp01(sr*intensity + spec*0.5)
		g := clamp01(sg*intensity + spec*0.5)
		b := clamp01(sb*intensity + spec*0.5)

		// Get projected quad corners
		corners := [4]vec2{proj[f.v[0]], proj[f.v[1]], proj[f.v[2]], proj[f.v[3]]}
		fillQuad(c, corners, canvas.RGB(r, g, b))
	}

	c.Rasterize()

	var buf strings.Builder
	writeHeader(&buf, m.width, demoNames[demoCube], m.demoTime)
	buf.WriteString(canvas.RenderString(c, canvas.ProfileTrueColor))
	return buf.String()
}

// fillQuad fills a convex quad on the pixel canvas using scanline rasterization.
func fillQuad(c *canvas.Canvas, corners [4]vec2, col canvas.Color) {
	// Find bounding box
	minX, minY := corners[0].X, corners[0].Y
	maxX, maxY := corners[0].X, corners[0].Y
	for _, p := range corners[1:] {
		minX = math.Min(minX, p.X)
		minY = math.Min(minY, p.Y)
		maxX = math.Max(maxX, p.X)
		maxY = math.Max(maxY, p.Y)
	}

	// Clamp to canvas
	y0 := int(math.Max(0, math.Floor(minY)))
	y1 := int(math.Min(float64(c.Height-1), math.Ceil(maxY)))
	x0 := int(math.Max(0, math.Floor(minX)))
	x1 := int(math.Min(float64(c.Width-1), math.Ceil(maxX)))

	// For each pixel, test if inside the quad using cross-product winding
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			p := vec2{float64(x) + 0.5, float64(y) + 0.5}
			if pointInQuad(p, corners) {
				c.SetPixel(x, y, col)
			}
		}
	}
}

func pointInQuad(p vec2, q [4]vec2) bool {
	// Check all cross products have same sign (convex quad winding test)
	sign := false
	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		cross := (q[j].X-q[i].X)*(p.Y-q[i].Y) - (q[j].Y-q[i].Y)*(p.X-q[i].X)
		if i == 0 {
			sign = cross > 0
		} else if (cross > 0) != sign {
			return false
		}
	}
	return true
}

// ─── Demo 3: ASCII Waves ────────────────────────────────────────────────────

func (m model) renderWaves() string {
	rows := m.height - 2
	c := canvas.New(m.width, rows, canvas.ModeASCII)

	t := m.demoTime

	// Sky gradient
	for y := 0; y < c.Height; y++ {
		skyT := float64(y) / float64(c.Height)
		col := canvas.RGB(
			0.01+skyT*0.02,
			0.01+skyT*0.04,
			0.05+skyT*0.08,
		)
		for x := 0; x < c.Width; x++ {
			c.SetPixel(x, y, col)
		}
	}

	// Stars in upper portion
	for i := 0; i < 40; i++ {
		sx := int(math.Mod(float64(i)*137.5+t*2, float64(c.Width)))
		sy := int(math.Mod(float64(i)*97.3, float64(c.Height/3)))
		twinkle := 0.3 + 0.7*math.Abs(math.Sin(t*1.5+float64(i)*0.7))
		c.SetPixel(sx, sy, canvas.RGB(twinkle, twinkle, twinkle*0.9))
	}

	// Render waves: compute wave height for each column
	waterLine := float64(c.Height) * 0.45 // where water starts

	for x := 0; x < c.Width; x++ {
		fx := float64(x)

		// Multiple overlapping sine waves
		h := 0.0
		h += math.Sin(fx*0.04+t*1.2) * 3.0
		h += math.Sin(fx*0.08+t*0.8+1.5) * 1.5
		h += math.Sin(fx*0.15+t*1.8+3.0) * 0.8
		h += math.Sin(fx*0.02+t*0.3) * 5.0

		waveTop := int(waterLine + h)

		// Fill water column
		for y := waveTop; y < c.Height; y++ {
			if y < 0 {
				continue
			}
			// Depth gradient
			depth := float64(y-waveTop) / float64(c.Height-waveTop+1)
			col := ocean.Sample(clamp01(depth))

			// Add horizontal ripple variation
			ripple := math.Sin(fx*0.1+float64(y)*0.3+t*2.0) * 0.08
			sr, sg, sb, _ := col.Straight()
			c.SetPixel(x, y, canvas.RGB(
				clamp01(sr+ripple),
				clamp01(sg+ripple),
				clamp01(sb+ripple*0.5),
			))
		}

		// Foam at wave crest
		if waveTop >= 0 && waveTop < c.Height {
			foam := 0.7 + 0.3*math.Sin(fx*0.3+t*3.0)
			c.SetPixel(x, waveTop, canvas.RGB(foam, foam, foam*0.95))
		}
	}

	c.Rasterize()

	var buf strings.Builder
	writeHeader(&buf, m.width, demoNames[demoWaves], m.demoTime)
	buf.WriteString(canvas.RenderString(c, canvas.ProfileTrueColor))
	return buf.String()
}

// ─── Demo 4: Animated Borders ───────────────────────────────────────────────

func (m model) renderBorders() string {
	c := canvas.New(m.width, m.height-2, canvas.ModeHalfBlock)
	c.Clear(canvas.RGB(0.03, 0.02, 0.05))
	c.Rasterize()

	// Three panels side by side
	panelW := (m.width - 4) / 3
	panelH := m.height - 5
	if panelW < 8 || panelH < 4 {
		return ""
	}
	gap := 1
	startY := 1

	t := m.demoTime

	// Panel 1: Dotted border
	x1 := gap
	drawDottedBorder(c, x1, startY, panelW, panelH, t)
	drawCellLabel(c, x1+2, startY+panelH/2, "Dotted", canvas.Hex("#C4B5F4"))

	// Panel 2: Gradient border
	x2 := gap + panelW + gap
	drawGradientBorder(c, x2, startY, panelW, panelH, t)
	drawCellLabel(c, x2+2, startY+panelH/2, "Gradient", canvas.Hex("#93E4D4"))

	// Panel 3: Colorshift border
	x3 := gap + 2*(panelW+gap)
	drawColorshiftBorder(c, x3, startY, panelW, panelH, t)
	drawCellLabel(c, x3+2, startY+panelH/2, "Colorshift", canvas.Hex("#F4B8D4"))

	var buf strings.Builder
	writeHeader(&buf, m.width, demoNames[demoBorders], m.demoTime)
	buf.WriteString(canvas.RenderString(c, canvas.ProfileTrueColor))
	return buf.String()
}

func drawDottedBorder(c *canvas.Canvas, x, y, w, h int, t float64) {
	dotRunes := []rune{'·', ' ', '·', ' ', '●', ' ', '·', ' '}
	offset := int(t * 8)
	baseCol := canvas.Hex("#8B7FBF")
	brightCol := canvas.Hex("#E0D4FF")

	perim := 2*(w-1) + 2*(h-1)
	for i := 0; i < perim; i++ {
		px, py := perimToXY(i, x, y, w, h)
		runeIdx := (i + offset) % len(dotRunes)
		r := dotRunes[runeIdx]

		// Pulse brightness based on position
		pulse := 0.5 + 0.5*math.Sin(float64(i)*0.3+t*4.0)
		col := baseCol.Lerp(brightCol, pulse)

		c.SetCell(px, py, canvas.Cell{Rune: r, FG: col, BG: canvas.Transparent})
	}
}

func drawGradientBorder(c *canvas.Canvas, x, y, w, h int, t float64) {
	perim := 2*(w-1) + 2*(h-1)
	if perim <= 0 {
		return
	}

	for i := 0; i < perim; i++ {
		px, py := perimToXY(i, x, y, w, h)
		pos := float64(i) / float64(perim)
		ct := math.Mod(pos+t*0.3, 1.0)
		col := rainbow.Sample(ct)

		r := borderRune(px, py, x, y, w, h)
		c.SetCell(px, py, canvas.Cell{Rune: r, FG: col, BG: canvas.Transparent})
	}
}

func drawColorshiftBorder(c *canvas.Canvas, x, y, w, h int, t float64) {
	// Entire border shifts through hue over time
	hue := math.Mod(t*0.4, 1.0)
	col := rainbow.Sample(hue)

	// Add a traveling bright spot
	perim := 2*(w-1) + 2*(h-1)
	if perim <= 0 {
		return
	}
	headPos := math.Mod(t*1.5, 1.0)

	for i := 0; i < perim; i++ {
		px, py := perimToXY(i, x, y, w, h)
		pos := float64(i) / float64(perim)

		// Distance from the traveling highlight
		dist := math.Abs(pos - headPos)
		if dist > 0.5 {
			dist = 1.0 - dist
		}
		highlight := math.Max(0, 1.0-dist*8.0)

		sr, sg, sb, _ := col.Straight()
		final := canvas.RGB(
			clamp01(sr+highlight*0.5),
			clamp01(sg+highlight*0.5),
			clamp01(sb+highlight*0.5),
		)

		r := borderRune(px, py, x, y, w, h)
		c.SetCell(px, py, canvas.Cell{Rune: r, FG: final, BG: canvas.Transparent})
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func writeHeader(buf *strings.Builder, width int, title string, demoTime float64) {
	// Progress bar
	progress := demoTime / demoDuration
	barW := width - 4
	if barW < 10 {
		barW = 10
	}
	filled := int(progress * float64(barW))

	buf.WriteString("\x1b[38;2;120;100;160m")
	// Center the title
	pad := (width - len(title)) / 2
	if pad < 0 {
		pad = 0
	}
	buf.WriteString(strings.Repeat(" ", pad))
	buf.WriteString("\x1b[38;2;200;180;240m")
	buf.WriteString(title)
	buf.WriteString("\x1b[0m\n")

	// Thin progress bar
	buf.WriteString("\x1b[38;2;80;60;120m")
	buf.WriteString("  ")
	for i := 0; i < barW; i++ {
		if i < filled {
			t := float64(i) / float64(barW)
			col := rainbow.Sample(t)
			r, g, b := col.ToRGB8()
			fmt.Fprintf(buf, "\x1b[38;2;%d;%d;%dm", r, g, b)
			buf.WriteRune('━')
		} else {
			buf.WriteString("\x1b[38;2;40;30;60m")
			buf.WriteRune('─')
		}
	}
	buf.WriteString("\x1b[0m\n")
}

func writeQuadRow(buf *strings.Builder, left, right *canvas.Canvas, row, totalW int, profile canvas.ColorProfile) {
	var lastFG, lastBG canvas.Color
	fgSet, bgSet := false, false

	writeCell := func(cell canvas.Cell) {
		fg := canvas.Degrade(cell.FG, profile)
		bg := canvas.Degrade(cell.BG, profile)
		if !fgSet || fg != lastFG {
			writeFG(buf, fg)
			lastFG = fg
			fgSet = true
		}
		if !bgSet || bg != lastBG {
			writeBG(buf, bg)
			lastBG = bg
			bgSet = true
		}
		r := cell.Rune
		if r == 0 {
			r = ' '
		}
		buf.WriteRune(r)
	}

	for col := 0; col < left.CellW; col++ {
		writeCell(left.GetCell(col, row))
	}
	for col := 0; col < right.CellW; col++ {
		writeCell(right.GetCell(col, row))
	}
}

func writeFG(buf *strings.Builder, c canvas.Color) {
	if c.A <= 0 {
		buf.WriteString("\x1b[39m")
		return
	}
	r, g, b := c.ToRGB8()
	fmt.Fprintf(buf, "\x1b[38;2;%d;%d;%dm", r, g, b)
}

func writeBG(buf *strings.Builder, c canvas.Color) {
	if c.A <= 0 {
		buf.WriteString("\x1b[49m")
		return
	}
	r, g, b := c.ToRGB8()
	fmt.Fprintf(buf, "\x1b[48;2;%d;%d;%dm", r, g, b)
}

func drawCellLabel(c *canvas.Canvas, x, y int, text string, fg canvas.Color) {
	for i, r := range text {
		c.SetCell(x+i, y, canvas.Cell{Rune: r, FG: fg, BG: canvas.Transparent})
	}
}

func borderRune(px, py, x, y, w, h int) rune {
	switch {
	case px == x && py == y:
		return '╭'
	case px == x+w-1 && py == y:
		return '╮'
	case px == x+w-1 && py == y+h-1:
		return '╯'
	case px == x && py == y+h-1:
		return '╰'
	case py == y || py == y+h-1:
		return '─'
	default:
		return '│'
	}
}

// perimToXY converts a perimeter index to cell coordinates (clockwise from top-left).
func perimToXY(i, x, y, w, h int) (int, int) {
	top := w - 1
	right := h - 1
	bottom := w - 1

	switch {
	case i < top: // top edge
		return x + i, y
	case i < top+right: // right edge
		return x + w - 1, y + (i - top)
	case i < top+right+bottom: // bottom edge
		return x + w - 1 - (i - top - right), y + h - 1
	default: // left edge
		return x, y + h - 1 - (i - top - right - bottom)
	}
}

// ─── 3D Vector Math ─────────────────────────────────────────────────────────

type vec3 struct{ X, Y, Z float64 }
type vec2 struct{ X, Y float64 }

func v3dot(a, b vec3) float64 { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }

func v3sub(a, b vec3) vec3 { return vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }

func v3cross(a, b vec3) vec3 {
	return vec3{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}

func v3norm(v vec3) vec3 {
	l := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if l < 1e-10 {
		return vec3{}
	}
	return vec3{v.X / l, v.Y / l, v.Z / l}
}

// v3reflect reflects incident vector I about normal N.
func v3reflect(I, N vec3) vec3 {
	d := 2 * v3dot(I, N)
	return vec3{I.X - d*N.X, I.Y - d*N.Y, I.Z - d*N.Z}
}

func rotateYXZ(v vec3, ax, ay, az float64) vec3 {
	// Y rotation
	cosY, sinY := math.Cos(ay), math.Sin(ay)
	x1 := v.X*cosY + v.Z*sinY
	z1 := -v.X*sinY + v.Z*cosY
	// X rotation
	cosX, sinX := math.Cos(ax), math.Sin(ax)
	y1 := v.Y*cosX - z1*sinX
	z2 := v.Y*sinX + z1*cosX
	// Z rotation
	cosZ, sinZ := math.Cos(az), math.Sin(az)
	x2 := x1*cosZ - y1*sinZ
	y2 := x1*sinZ + y1*cosZ
	return vec3{x2, y2, z2}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ─── Main ───────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
