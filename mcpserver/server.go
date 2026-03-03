// Package mcpserver provides an MCP (Model Context Protocol) server that exposes
// Cynthia's scene graph engine to AI agents. Tools allow creating, positioning,
// styling, and removing panels with procedural effects, animated borders, and text.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/bealmot/cynthia/border"
	"github.com/bealmot/cynthia/canvas"
	"github.com/bealmot/cynthia/compose"
	"github.com/bealmot/cynthia/director"
	"github.com/bealmot/cynthia/effect"
	"github.com/bealmot/cynthia/runtime"
	"github.com/bealmot/cynthia/scene"
	"github.com/bealmot/cynthia/widget"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	// Blank import to register built-in effects.
	_ "github.com/bealmot/cynthia/effect"
)

// Server wraps the MCP server and the Cynthia runtime engine.
type Server struct {
	engine    *runtime.Engine
	program   *runtime.Program
	mcpServer *mcp.Server
}

// New creates an MCP server wired to the given engine (non-interactive mode).
func New(engine *runtime.Engine) *Server {
	impl := &mcp.Implementation{
		Name:    "cynthia",
		Version: "0.1.0",
	}
	mcpSrv := mcp.NewServer(impl, nil)

	s := &Server{
		engine:    engine,
		mcpServer: mcpSrv,
	}
	s.registerTools()
	return s
}

// NewWithProgram creates an MCP server wired to the interactive BubbleTea runtime.
func NewWithProgram(prog *runtime.Program) *Server {
	impl := &mcp.Implementation{
		Name:    "cynthia",
		Version: "0.1.0",
	}
	mcpSrv := mcp.NewServer(impl, nil)

	s := &Server{
		program:   prog,
		mcpServer: mcpSrv,
	}
	s.registerTools()
	s.registerWidgetTools()
	return s
}

// lock acquires the underlying runtime's mutex.
func (s *Server) lock() {
	if s.program != nil {
		s.program.Lock()
	} else {
		s.engine.Lock()
	}
}

// unlock releases the underlying runtime's mutex.
func (s *Server) unlock() {
	if s.program != nil {
		s.program.Unlock()
	} else {
		s.engine.Unlock()
	}
}

// scene returns the underlying scene.
func (s *Server) scene() *scene.Scene {
	if s.program != nil {
		return s.program.Scene
	}
	return s.engine.Scene
}

// renderMode returns the active render mode.
func (s *Server) renderMode() canvas.RenderMode {
	if s.program != nil {
		return s.program.Mode
	}
	return s.engine.Mode
}

// termCols returns the terminal column count.
func (s *Server) termCols() int {
	if s.program != nil {
		return s.program.Cols()
	}
	return s.engine.Cols()
}

// termRows returns the terminal row count.
func (s *Server) termRows() int {
	if s.program != nil {
		return s.program.Rows()
	}
	return s.engine.Rows()
}

// events returns the event queue (only available in program mode).
func (s *Server) events() *widget.EventQueue {
	if s.program != nil {
		return s.program.Events
	}
	return nil
}

// Run starts the MCP server on stdio. Blocks until the client disconnects.
func (s *Server) Run(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) registerTools() {
	// Panel CRUD
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.panel.create",
		Description: "Create a new panel. Returns the panel ID. Panels start as solid fill; use set_effect/set_text/set_border to add content.",
	}, s.handlePanelCreate)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.panel.update",
		Description: "Update panel properties: position, size, opacity, z-order, visibility, blend mode, fill color.",
	}, s.handlePanelUpdate)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.panel.remove",
		Description: "Instantly remove a panel from the scene.",
	}, s.handlePanelRemove)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.panel.list",
		Description: "List all panels with their state (position, size, effect, border, text, etc.).",
	}, s.handlePanelList)

	// Panel content
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.panel.set_effect",
		Description: fmt.Sprintf("Set a procedural effect on a panel. Available effects: %s. Pass params to tune (e.g. intensity, speed, decay). Set effect to empty string to clear.", strings.Join(sortedEffectNames(), ", ")),
	}, s.handlePanelSetEffect)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.panel.set_border",
		Description: "Set an animated border on a panel. Styles: cascade, nouveau, pulse. Set style to empty string to clear.",
	}, s.handlePanelSetBorder)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.panel.set_text",
		Description: "Set text content on a panel. Text is stamped onto the cell grid (inset if panel has a border). Set text to empty string to clear.",
	}, s.handlePanelSetText)

	// Scene-level
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.scene.mood",
		Description: fmt.Sprintf("Apply a mood preset that creates/updates a fullscreen background panel with themed effect and border. Moods: %s.", strings.Join(sortedMoodNames(), ", ")),
	}, s.handleSceneMood)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.scene.clear",
		Description: "Remove all panels from the scene.",
	}, s.handleSceneClear)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.scene.snapshot",
		Description: "Return the full scene state as JSON (all panels with properties).",
	}, s.handleSceneSnapshot)
}

// ============================================================
// Tool input types
// ============================================================

type PanelCreateInput struct {
	ID      string  `json:"id,omitempty" jsonschema:"optional panel ID; auto-generated if omitted"`
	X       float64 `json:"x" jsonschema:"X position in cell columns"`
	Y       float64 `json:"y" jsonschema:"Y position in cell rows"`
	Width   int     `json:"width" jsonschema:"width in terminal columns"`
	Height  int     `json:"height" jsonschema:"height in terminal rows"`
	Z       int     `json:"z,omitempty" jsonschema:"z-order (lower = further back)"`
	Opacity *float64 `json:"opacity,omitempty" jsonschema:"0.0-1.0, default 1.0"`
	Blend   string  `json:"blend,omitempty" jsonschema:"normal, screen, additive, multiply"`
	Color   string  `json:"color,omitempty" jsonschema:"fill color as hex (#RRGGBB or #RGB), default transparent"`
	Effect  string  `json:"effect,omitempty" jsonschema:"initial effect name (fire, plasma, matrix, gradient, particles, starfield)"`
	Border  string  `json:"border,omitempty" jsonschema:"initial border style (cascade, nouveau, pulse)"`
	Text    string  `json:"text,omitempty" jsonschema:"initial text content"`
}

type PanelUpdateInput struct {
	ID      string   `json:"id" jsonschema:"panel ID to update"`
	X       *float64 `json:"x,omitempty" jsonschema:"new X position"`
	Y       *float64 `json:"y,omitempty" jsonschema:"new Y position"`
	Width   *int     `json:"width,omitempty" jsonschema:"new width"`
	Height  *int     `json:"height,omitempty" jsonschema:"new height"`
	Z       *int     `json:"z,omitempty" jsonschema:"new z-order"`
	Opacity *float64 `json:"opacity,omitempty" jsonschema:"new opacity"`
	Visible *bool    `json:"visible,omitempty" jsonschema:"visibility toggle"`
	Blend   string   `json:"blend,omitempty" jsonschema:"new blend mode"`
	Color   *string  `json:"color,omitempty" jsonschema:"new fill color (#RRGGBB), empty string for transparent"`
}

type PanelRemoveInput struct {
	ID string `json:"id" jsonschema:"panel ID to remove"`
}

type PanelListInput struct{}

type PanelSetEffectInput struct {
	ID     string             `json:"id" jsonschema:"panel ID"`
	Effect string             `json:"effect" jsonschema:"effect name (fire, plasma, matrix, gradient, particles, starfield) or empty to clear"`
	Params map[string]float64 `json:"params,omitempty" jsonschema:"effect parameters (e.g. intensity, speed, decay)"`
}

type PanelSetBorderInput struct {
	ID     string             `json:"id" jsonschema:"panel ID"`
	Style  string             `json:"style" jsonschema:"border style (cascade, nouveau, pulse) or empty to clear"`
	Params map[string]float64 `json:"params,omitempty" jsonschema:"border parameters (e.g. speed)"`
}

type PanelSetTextInput struct {
	ID   string `json:"id" jsonschema:"panel ID"`
	Text string `json:"text" jsonschema:"text content (supports newlines)"`
}

type SceneMoodInput struct {
	Mood string `json:"mood" jsonschema:"mood preset name"`
}

type SceneClearInput struct{}

type SceneSnapshotInput struct{}

// ============================================================
// Tool output types
// ============================================================

type PanelInfo struct {
	ID      string  `json:"id"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Z       int     `json:"z"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Opacity float64 `json:"opacity"`
	Visible bool    `json:"visible"`
	Blend   string  `json:"blend"`
	Effect  string  `json:"effect,omitempty"`
	Border  string  `json:"border,omitempty"`
	HasText bool    `json:"has_text,omitempty"`
	Widget  string  `json:"widget,omitempty"`
	Focused bool    `json:"focused,omitempty"`
}

type SimpleOutput struct {
	OK      bool   `json:"ok"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

type PanelListOutput struct {
	Panels []PanelInfo `json:"panels"`
	Count  int         `json:"count"`
}

type SceneSnapshotOutput struct {
	Panels []PanelInfo `json:"panels"`
	Count  int         `json:"count"`
	Cols   int         `json:"cols"`
	Rows   int         `json:"rows"`
}

// ============================================================
// Tool handlers
// ============================================================

func (s *Server) handlePanelCreate(ctx context.Context, req *mcp.CallToolRequest, input PanelCreateInput) (*mcp.CallToolResult, SimpleOutput, error) {
	id := input.ID
	if id == "" {
		id = scene.GenerateID()
	}

	s.lock()
	defer s.unlock()

	p := scene.NewPanel(id, input.Width, input.Height, s.renderMode())
	p.X = input.X
	p.Y = input.Y
	p.Z = input.Z
	if input.Opacity != nil {
		p.Opacity = *input.Opacity
	}
	p.Blend = parseBlendMode(input.Blend)

	// Fill color
	if input.Color != "" {
		p.Fill = canvas.Hex(input.Color)
	}

	// Initial effect
	if input.Effect != "" {
		if fx := effect.Create(input.Effect); fx != nil {
			p.Effect = fx
		}
	}

	// Initial border
	if input.Border != "" {
		p.Border = createBorder(input.Border)
	}

	// Initial text
	p.Text = input.Text

	s.scene().Add(p)

	return nil, SimpleOutput{OK: true, ID: id}, nil
}

func (s *Server) handlePanelUpdate(ctx context.Context, req *mcp.CallToolRequest, input PanelUpdateInput) (*mcp.CallToolResult, SimpleOutput, error) {
	s.lock()
	defer s.unlock()

	p := s.scene().Get(input.ID)
	if p == nil {
		return errResult("panel not found: " + input.ID)
	}

	if input.X != nil {
		p.X = *input.X
	}
	if input.Y != nil {
		p.Y = *input.Y
	}
	if input.Z != nil {
		p.Z = *input.Z
	}
	if input.Opacity != nil {
		p.Opacity = *input.Opacity
	}
	if input.Visible != nil {
		p.Visible = *input.Visible
	}
	if input.Blend != "" {
		p.Blend = parseBlendMode(input.Blend)
	}
	if input.Color != nil {
		if *input.Color == "" {
			p.Fill = canvas.Transparent
		} else {
			p.Fill = canvas.Hex(*input.Color)
		}
	}
	if input.Width != nil || input.Height != nil {
		w, h := p.Width, p.Height
		if input.Width != nil {
			w = *input.Width
		}
		if input.Height != nil {
			h = *input.Height
		}
		p.Width = w
		p.Height = h
		p.Canvas = canvas.New(w, h, s.renderMode())
	}

	return nil, SimpleOutput{OK: true, ID: input.ID}, nil
}

func (s *Server) handlePanelRemove(ctx context.Context, req *mcp.CallToolRequest, input PanelRemoveInput) (*mcp.CallToolResult, SimpleOutput, error) {
	s.lock()
	defer s.unlock()

	existed := s.scene().Remove(input.ID)
	if !existed {
		return errResult("panel not found: " + input.ID)
	}
	return nil, SimpleOutput{OK: true, ID: input.ID}, nil
}

func (s *Server) handlePanelList(ctx context.Context, req *mcp.CallToolRequest, input PanelListInput) (*mcp.CallToolResult, PanelListOutput, error) {
	s.lock()
	panels := s.scene().Panels()
	s.unlock()

	return nil, PanelListOutput{Panels: panelInfos(panels), Count: len(panels)}, nil
}

func (s *Server) handlePanelSetEffect(ctx context.Context, req *mcp.CallToolRequest, input PanelSetEffectInput) (*mcp.CallToolResult, SimpleOutput, error) {
	s.lock()
	defer s.unlock()

	p := s.scene().Get(input.ID)
	if p == nil {
		return errResult("panel not found: " + input.ID)
	}

	if input.Effect == "" {
		p.Effect = nil
		return nil, SimpleOutput{OK: true, ID: input.ID, Message: "effect cleared"}, nil
	}

	fx := effect.Create(input.Effect)
	if fx == nil {
		return errResult("unknown effect: " + input.Effect + ". Available: " + strings.Join(sortedEffectNames(), ", "))
	}
	if len(input.Params) > 0 {
		fx.SetParams(input.Params)
	}
	p.Effect = fx

	return nil, SimpleOutput{OK: true, ID: input.ID, Message: "effect set to " + input.Effect}, nil
}

func (s *Server) handlePanelSetBorder(ctx context.Context, req *mcp.CallToolRequest, input PanelSetBorderInput) (*mcp.CallToolResult, SimpleOutput, error) {
	s.lock()
	defer s.unlock()

	p := s.scene().Get(input.ID)
	if p == nil {
		return errResult("panel not found: " + input.ID)
	}

	if input.Style == "" {
		p.Border = nil
		return nil, SimpleOutput{OK: true, ID: input.ID, Message: "border cleared"}, nil
	}

	b := createBorder(input.Style)
	if b == nil {
		return errResult("unknown border style: " + input.Style + ". Available: cascade, nouveau, pulse")
	}
	if len(input.Params) > 0 {
		b.SetParams(input.Params)
	}
	p.Border = b

	return nil, SimpleOutput{OK: true, ID: input.ID, Message: "border set to " + input.Style}, nil
}

func (s *Server) handlePanelSetText(ctx context.Context, req *mcp.CallToolRequest, input PanelSetTextInput) (*mcp.CallToolResult, SimpleOutput, error) {
	s.lock()
	defer s.unlock()

	p := s.scene().Get(input.ID)
	if p == nil {
		return errResult("panel not found: " + input.ID)
	}

	p.Text = input.Text
	return nil, SimpleOutput{OK: true, ID: input.ID}, nil
}

func (s *Server) handleSceneMood(ctx context.Context, req *mcp.CallToolRequest, input SceneMoodInput) (*mcp.CallToolResult, SimpleOutput, error) {
	mood, ok := director.Moods[input.Mood]
	if !ok {
		return errResult("unknown mood: " + input.Mood + ". Available: " + strings.Join(sortedMoodNames(), ", "))
	}

	s.lock()
	defer s.unlock()

	// Create or update the fullscreen mood background panel.
	const moodID = "_mood_bg"
	sc := s.scene()
	cols, rows := s.termCols(), s.termRows()
	mode := s.renderMode()

	p := sc.Get(moodID)
	if p == nil {
		p = scene.NewPanel(moodID, cols, rows, mode)
		p.Z = -100 // well behind everything
		sc.Add(p)
	}

	// Resize to current terminal size
	if p.Width != cols || p.Height != rows {
		p.Width = cols
		p.Height = rows
		p.Canvas = canvas.New(p.Width, p.Height, mode)
	}

	// Apply mood preset
	p.Opacity = mood.Intensity
	if fx := effect.Create(mood.Effect); fx != nil {
		if mood.Params != nil {
			fx.SetParams(mood.Params)
		}
		p.Effect = fx
	}
	if mood.BorderStyle != "" {
		p.Border = createBorder(mood.BorderStyle)
	}

	return nil, SimpleOutput{OK: true, Message: "mood set to " + input.Mood}, nil
}

func (s *Server) handleSceneClear(ctx context.Context, req *mcp.CallToolRequest, input SceneClearInput) (*mcp.CallToolResult, SimpleOutput, error) {
	s.lock()
	s.scene().Clear()
	s.unlock()

	return nil, SimpleOutput{OK: true, Message: "scene cleared"}, nil
}

func (s *Server) handleSceneSnapshot(ctx context.Context, req *mcp.CallToolRequest, input SceneSnapshotInput) (*mcp.CallToolResult, SceneSnapshotOutput, error) {
	s.lock()
	panels := s.scene().Panels()
	cols, rows := s.termCols(), s.termRows()
	s.unlock()

	return nil, SceneSnapshotOutput{
		Panels: panelInfos(panels),
		Count:  len(panels),
		Cols:   cols,
		Rows:   rows,
	}, nil
}

// ============================================================
// Helpers
// ============================================================

func errResult(msg string) (*mcp.CallToolResult, SimpleOutput, error) {
	return &mcp.CallToolResult{IsError: true}, SimpleOutput{OK: false, Message: msg}, nil
}

func panelInfos(panels []*scene.Panel) []PanelInfo {
	infos := make([]PanelInfo, len(panels))
	for i, p := range panels {
		info := PanelInfo{
			ID:      p.ID,
			X:       p.X,
			Y:       p.Y,
			Z:       p.Z,
			Width:   p.Width,
			Height:  p.Height,
			Opacity: p.Opacity,
			Visible: p.Visible,
			Blend:   blendModeString(p.Blend),
			HasText: p.Text != "",
		}
		if p.Effect != nil {
			info.Effect = p.Effect.Name()
		}
		if p.Border != nil {
			info.Border = p.Border.Name()
		}
		if p.Widget != nil {
			info.Widget = p.Widget.Type()
			info.Focused = p.Focused
		}
		infos[i] = info
	}
	return infos
}

func parseBlendMode(s string) compose.BlendMode {
	switch s {
	case "screen":
		return compose.BlendScreen
	case "additive":
		return compose.BlendAdditive
	case "multiply":
		return compose.BlendMultiply
	default:
		return compose.BlendNormal
	}
}

func blendModeString(m compose.BlendMode) string {
	switch m {
	case compose.BlendScreen:
		return "screen"
	case compose.BlendAdditive:
		return "additive"
	case compose.BlendMultiply:
		return "multiply"
	default:
		return "normal"
	}
}

func createBorder(style string) border.Border {
	switch style {
	case "cascade":
		return border.NewCascade()
	case "nouveau":
		return border.NewNouveau()
	case "pulse":
		return border.NewPulse()
	default:
		return nil
	}
}

func sortedEffectNames() []string {
	names := effect.Names()
	sort.Strings(names)
	return names
}

func sortedMoodNames() []string {
	names := make([]string, 0, len(director.Moods))
	for k := range director.Moods {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func marshalJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ============================================================
// Widget tools (only registered in interactive/program mode)
// ============================================================

type WidgetTextInputInput struct {
	ID          string  `json:"id,omitempty" jsonschema:"optional panel ID; auto-generated if omitted"`
	X           float64 `json:"x" jsonschema:"X position in cell columns"`
	Y           float64 `json:"y" jsonschema:"Y position in cell rows"`
	Width       int     `json:"width" jsonschema:"width in terminal columns"`
	Placeholder string  `json:"placeholder,omitempty" jsonschema:"placeholder text shown when empty"`
	Border      string  `json:"border,omitempty" jsonschema:"border style (cascade, nouveau, pulse)"`
	Color       string  `json:"color,omitempty" jsonschema:"fill color as hex (#RRGGBB)"`
	Focus       *bool   `json:"focus,omitempty" jsonschema:"auto-focus this widget (default true)"`
}

type WidgetButtonInput struct {
	ID     string  `json:"id,omitempty" jsonschema:"optional panel ID; auto-generated if omitted"`
	X      float64 `json:"x" jsonschema:"X position in cell columns"`
	Y      float64 `json:"y" jsonschema:"Y position in cell rows"`
	Width  int     `json:"width" jsonschema:"width in terminal columns"`
	Label  string  `json:"label" jsonschema:"button label text"`
	Border string  `json:"border,omitempty" jsonschema:"border style (cascade, nouveau, pulse)"`
	Color  string  `json:"color,omitempty" jsonschema:"fill color as hex (#RRGGBB)"`
}

type WidgetGetStateInput struct {
	ID string `json:"id" jsonschema:"panel ID of the widget"`
}

type WidgetSetFocusInput struct {
	ID string `json:"id" jsonschema:"panel ID to focus"`
}

type WidgetPollEventsInput struct{}

type WidgetStateOutput struct {
	OK       bool           `json:"ok"`
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	State    map[string]any `json:"state"`
	Focused  bool           `json:"focused"`
}

type WidgetEventsOutput struct {
	Events []widget.Event `json:"events"`
	Count  int            `json:"count"`
}

func (s *Server) registerWidgetTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.widget.text_input",
		Description: "Create a panel with an interactive text input widget. The terminal user can type into it. Use get_state to read the current value.",
	}, s.handleWidgetTextInput)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.widget.button",
		Description: "Create a panel with a clickable button widget. The terminal user can click it or press Enter when focused. Use poll_events to detect clicks.",
	}, s.handleWidgetButton)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.widget.get_state",
		Description: "Read a widget panel's current state (text value, button clicks, etc.).",
	}, s.handleWidgetGetState)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.widget.set_focus",
		Description: "Focus a specific widget panel so it receives keyboard input.",
	}, s.handleWidgetSetFocus)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.widget.poll_events",
		Description: "Get widget events (button clicks, text submissions) since last poll. Events are drained from the queue.",
	}, s.handleWidgetPollEvents)
}

func (s *Server) handleWidgetTextInput(ctx context.Context, req *mcp.CallToolRequest, input WidgetTextInputInput) (*mcp.CallToolResult, SimpleOutput, error) {
	id := input.ID
	if id == "" {
		id = scene.GenerateID()
	}

	// Text input panels are 3 rows tall (border top + input + border bottom)
	// or 1 row if no border
	height := 1
	if input.Border != "" {
		height = 3
	}

	s.lock()
	defer s.unlock()

	p := scene.NewPanel(id, input.Width, height, s.renderMode())
	p.X = input.X
	p.Y = input.Y

	if input.Color != "" {
		p.Fill = canvas.Hex(input.Color)
	}
	if input.Border != "" {
		p.Border = createBorder(input.Border)
	}

	// Create the text input widget
	innerWidth := input.Width
	if input.Border != "" {
		innerWidth -= 2 // inset for border
	}
	ti := widget.NewTextInput(innerWidth, input.Placeholder)
	ti.EventQueue = s.events()
	ti.PanelID = id
	p.Widget = ti

	s.scene().Add(p)

	// Auto-focus unless explicitly disabled
	if input.Focus == nil || *input.Focus {
		s.scene().Focus(id)
	}

	return nil, SimpleOutput{OK: true, ID: id, Message: "text input created"}, nil
}

func (s *Server) handleWidgetButton(ctx context.Context, req *mcp.CallToolRequest, input WidgetButtonInput) (*mcp.CallToolResult, SimpleOutput, error) {
	id := input.ID
	if id == "" {
		id = scene.GenerateID()
	}

	height := 1
	if input.Border != "" {
		height = 3
	}

	s.lock()
	defer s.unlock()

	p := scene.NewPanel(id, input.Width, height, s.renderMode())
	p.X = input.X
	p.Y = input.Y

	if input.Color != "" {
		p.Fill = canvas.Hex(input.Color)
	}
	if input.Border != "" {
		p.Border = createBorder(input.Border)
	}

	innerWidth := input.Width
	if input.Border != "" {
		innerWidth -= 2
	}
	btn := widget.NewButton(input.Label, innerWidth)
	btn.EventQueue = s.events()
	btn.PanelID = id
	p.Widget = btn

	s.scene().Add(p)

	return nil, SimpleOutput{OK: true, ID: id, Message: "button created"}, nil
}

func (s *Server) handleWidgetGetState(ctx context.Context, req *mcp.CallToolRequest, input WidgetGetStateInput) (*mcp.CallToolResult, WidgetStateOutput, error) {
	s.lock()
	defer s.unlock()

	p := s.scene().Get(input.ID)
	if p == nil {
		return &mcp.CallToolResult{IsError: true}, WidgetStateOutput{OK: false, ID: input.ID}, nil
	}
	if p.Widget == nil {
		return &mcp.CallToolResult{IsError: true}, WidgetStateOutput{OK: false, ID: input.ID}, nil
	}

	return nil, WidgetStateOutput{
		OK:      true,
		ID:      input.ID,
		Type:    p.Widget.Type(),
		State:   p.Widget.State(),
		Focused: p.Focused,
	}, nil
}

func (s *Server) handleWidgetSetFocus(ctx context.Context, req *mcp.CallToolRequest, input WidgetSetFocusInput) (*mcp.CallToolResult, SimpleOutput, error) {
	s.lock()
	defer s.unlock()

	p := s.scene().Get(input.ID)
	if p == nil {
		return errResult("panel not found: " + input.ID)
	}
	if p.Widget == nil {
		return errResult("panel has no widget: " + input.ID)
	}

	s.scene().Focus(input.ID)
	return nil, SimpleOutput{OK: true, ID: input.ID, Message: "focused"}, nil
}

func (s *Server) handleWidgetPollEvents(ctx context.Context, req *mcp.CallToolRequest, input WidgetPollEventsInput) (*mcp.CallToolResult, WidgetEventsOutput, error) {
	eq := s.events()
	if eq == nil {
		return nil, WidgetEventsOutput{Events: []widget.Event{}, Count: 0}, nil
	}

	events := eq.Drain()
	return nil, WidgetEventsOutput{Events: events, Count: len(events)}, nil
}
