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
		Name:        "cynthia.panel.set_params",
		Description: "Smoothly transition effect parameters using spring interpolation. Sets target values that the panel's springs animate toward over multiple frames.",
	}, s.handlePanelSetParams)

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

// MaskInput defines a spatial mask configuration.
type MaskInput struct {
	Type   string             `json:"type" jsonschema:"mask type: circle, rect, gradient, vignette, none"`
	Params map[string]float64 `json:"params,omitempty" jsonschema:"mask parameters (e.g. cx, cy, radius, feather, strength)"`
}

type PanelCreateInput struct {
	ID      string   `json:"id,omitempty" jsonschema:"optional panel ID; auto-generated if omitted"`
	X       float64  `json:"x" jsonschema:"X position in cell columns"`
	Y       float64  `json:"y" jsonschema:"Y position in cell rows"`
	Width   int      `json:"width" jsonschema:"width in terminal columns"`
	Height  int      `json:"height" jsonschema:"height in terminal rows"`
	Z       int      `json:"z,omitempty" jsonschema:"z-order (lower = further back)"`
	Opacity *float64 `json:"opacity,omitempty" jsonschema:"0.0-1.0, default 1.0"`
	Blend   string   `json:"blend,omitempty" jsonschema:"normal, screen, additive, multiply"`
	Color   string   `json:"color,omitempty" jsonschema:"fill color as hex (#RRGGBB or #RGB), default transparent"`
	Effect  string   `json:"effect,omitempty" jsonschema:"initial effect name (fire, plasma, matrix, gradient, particles, starfield)"`
	Border  string   `json:"border,omitempty" jsonschema:"initial border style (cascade, nouveau, pulse)"`
	Text    string   `json:"text,omitempty" jsonschema:"initial text content"`
	Mask    *MaskInput `json:"mask,omitempty" jsonschema:"spatial mask (circle, rect, gradient, vignette)"`
}

type PanelUpdateInput struct {
	ID      string     `json:"id" jsonschema:"panel ID to update"`
	X       *float64   `json:"x,omitempty" jsonschema:"new X position"`
	Y       *float64   `json:"y,omitempty" jsonschema:"new Y position"`
	Width   *int       `json:"width,omitempty" jsonschema:"new width"`
	Height  *int       `json:"height,omitempty" jsonschema:"new height"`
	Z       *int       `json:"z,omitempty" jsonschema:"new z-order"`
	Opacity *float64   `json:"opacity,omitempty" jsonschema:"new opacity"`
	Visible *bool      `json:"visible,omitempty" jsonschema:"visibility toggle"`
	Blend   string     `json:"blend,omitempty" jsonschema:"new blend mode"`
	Color   *string    `json:"color,omitempty" jsonschema:"new fill color (#RRGGBB), empty string for transparent"`
	Mask    *MaskInput `json:"mask,omitempty" jsonschema:"spatial mask (circle, rect, gradient, vignette, none to clear)"`
}

type PanelRemoveInput struct {
	ID string `json:"id" jsonschema:"panel ID to remove"`
}

type PanelListInput struct{}

type PanelSetEffectInput struct {
	ID      string             `json:"id" jsonschema:"panel ID"`
	Effect  string             `json:"effect,omitempty" jsonschema:"single effect name, or empty to clear"`
	Effects []string           `json:"effects,omitempty" jsonschema:"effect pipeline (e.g. [plasma, crt]) — stage 1+ act as post-processors"`
	Params  map[string]float64 `json:"params,omitempty" jsonschema:"effect parameters. For chains, prefix with stage index: 0.speed, 1.scanlines"`
}

type PanelSetParamsInput struct {
	ID              string             `json:"id" jsonschema:"panel ID"`
	Params          map[string]float64 `json:"params" jsonschema:"parameter targets to spring-interpolate toward"`
	TransitionSpeed *float64           `json:"transition_speed,omitempty" jsonschema:"spring frequency multiplier (higher = snappier), default 1.0"`
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
	Mask    string  `json:"mask,omitempty"`
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

	// Initial mask
	if input.Mask != nil {
		applyMask(p, input.Mask)
	}

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
		// Regenerate mask for new canvas dimensions.
		if p.MaskType != "" {
			p.UpdateMask()
		}
	}

	if input.Mask != nil {
		applyMask(p, input.Mask)
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

	// Effect chain: multiple effects pipelined together.
	if len(input.Effects) > 1 {
		stages := make([]effect.Effect, 0, len(input.Effects))
		for _, name := range input.Effects {
			fx := effect.Create(name)
			if fx == nil {
				return errResult("unknown effect: " + name + ". Available: " + strings.Join(sortedEffectNames(), ", "))
			}
			stages = append(stages, fx)
		}
		chain := effect.NewChain(stages...)
		if len(input.Params) > 0 {
			chain.SetParams(input.Params)
		}
		p.Effect = chain
		p.Director = nil
		return nil, SimpleOutput{OK: true, ID: input.ID, Message: "effect chain set: " + strings.Join(input.Effects, " → ")}, nil
	}

	// Single effect (from Effect or Effects[0]).
	effectName := input.Effect
	if effectName == "" && len(input.Effects) == 1 {
		effectName = input.Effects[0]
	}

	if effectName == "" {
		p.Effect = nil
		p.Director = nil
		return nil, SimpleOutput{OK: true, ID: input.ID, Message: "effect cleared"}, nil
	}

	fx := effect.Create(effectName)
	if fx == nil {
		return errResult("unknown effect: " + effectName + ". Available: " + strings.Join(sortedEffectNames(), ", "))
	}
	if len(input.Params) > 0 {
		fx.SetParams(input.Params)
	}
	p.Effect = fx
	// Reset per-panel director on effect change (old springs are stale).
	p.Director = nil

	return nil, SimpleOutput{OK: true, ID: input.ID, Message: "effect set to " + effectName}, nil
}

func (s *Server) handlePanelSetParams(ctx context.Context, req *mcp.CallToolRequest, input PanelSetParamsInput) (*mcp.CallToolResult, SimpleOutput, error) {
	s.lock()
	defer s.unlock()

	p := s.scene().Get(input.ID)
	if p == nil {
		return errResult("panel not found: " + input.ID)
	}
	if p.Effect == nil {
		return errResult("panel has no effect: " + input.ID)
	}

	// Create a PanelDirector if the panel doesn't have one yet.
	if p.Director == nil {
		speed := 1.0
		if input.TransitionSpeed != nil {
			speed = *input.TransitionSpeed
		}
		p.Director = director.NewPanelDirector(speed)

		// Seed springs from current effect params so we interpolate from current values.
		for k, v := range p.Effect.Params() {
			p.Director.SetTarget(k, v)
		}
	} else if input.TransitionSpeed != nil {
		p.Director.SetSpeed(*input.TransitionSpeed)
	}

	// Set spring targets for the requested params.
	p.Director.SetTargets(input.Params)

	return nil, SimpleOutput{OK: true, ID: input.ID, Message: "spring targets set"}, nil
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
		if p.MaskType != "" && p.MaskType != "none" {
			info.Mask = p.MaskType
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

func applyMask(p *scene.Panel, m *MaskInput) {
	p.MaskType = m.Type
	p.MaskParams = m.Params
	p.UpdateMask()
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

type WidgetProgressBarInput struct {
	ID     string  `json:"id,omitempty" jsonschema:"optional panel ID; auto-generated if omitted"`
	X      float64 `json:"x" jsonschema:"X position in cell columns"`
	Y      float64 `json:"y" jsonschema:"Y position in cell rows"`
	Width  int     `json:"width" jsonschema:"width in terminal columns"`
	Label  string  `json:"label,omitempty" jsonschema:"label text shown before the bar"`
	Value  float64 `json:"value,omitempty" jsonschema:"initial value 0.0-1.0"`
	Style  string  `json:"style,omitempty" jsonschema:"bar style: block (default) or ascii"`
	Border string  `json:"border,omitempty" jsonschema:"border style (cascade, nouveau, pulse)"`
	Color  string  `json:"color,omitempty" jsonschema:"fill color as hex (#RRGGBB)"`
}

type WidgetSparklineInput struct {
	ID     string    `json:"id,omitempty" jsonschema:"optional panel ID; auto-generated if omitted"`
	X      float64   `json:"x" jsonschema:"X position in cell columns"`
	Y      float64   `json:"y" jsonschema:"Y position in cell rows"`
	Width  int       `json:"width" jsonschema:"width in terminal columns"`
	Label  string    `json:"label,omitempty" jsonschema:"label text shown before the sparkline"`
	Values []float64 `json:"values,omitempty" jsonschema:"initial data values"`
	Border string    `json:"border,omitempty" jsonschema:"border style (cascade, nouveau, pulse)"`
	Color  string    `json:"color,omitempty" jsonschema:"fill color as hex (#RRGGBB)"`
}

type WidgetSelectListInput struct {
	ID     string   `json:"id,omitempty" jsonschema:"optional panel ID; auto-generated if omitted"`
	X      float64  `json:"x" jsonschema:"X position in cell columns"`
	Y      float64  `json:"y" jsonschema:"Y position in cell rows"`
	Width  int      `json:"width" jsonschema:"width in terminal columns"`
	Items  []string `json:"items" jsonschema:"list of selectable items"`
	Height int      `json:"height,omitempty" jsonschema:"visible rows (default: number of items, max 10)"`
	Border string   `json:"border,omitempty" jsonschema:"border style (cascade, nouveau, pulse)"`
	Color  string   `json:"color,omitempty" jsonschema:"fill color as hex (#RRGGBB)"`
	Focus  *bool    `json:"focus,omitempty" jsonschema:"auto-focus this widget (default true)"`
}

type WidgetTableInput struct {
	ID      string     `json:"id,omitempty" jsonschema:"optional panel ID; auto-generated if omitted"`
	X       float64    `json:"x" jsonschema:"X position in cell columns"`
	Y       float64    `json:"y" jsonschema:"Y position in cell rows"`
	Width   int        `json:"width" jsonschema:"width in terminal columns"`
	Headers []string   `json:"headers" jsonschema:"column header names"`
	Rows    [][]string `json:"rows,omitempty" jsonschema:"initial data rows"`
	Border  string     `json:"border,omitempty" jsonschema:"border style (cascade, nouveau, pulse)"`
	Color   string     `json:"color,omitempty" jsonschema:"fill color as hex (#RRGGBB)"`
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
		Name:        "cynthia.widget.progress_bar",
		Description: "Create a panel with a progress bar widget. Display-only — use get_state to read, set value via update. Shows [████░░░░] style bar.",
	}, s.handleWidgetProgressBar)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.widget.sparkline",
		Description: "Create a panel with a sparkline widget. Display-only — shows values as ▁▂▃▄▅▆▇█ bars. Push values via MCP to update.",
	}, s.handleWidgetSparkline)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.widget.select_list",
		Description: "Create a panel with an interactive selection list. User navigates with up/down arrows and selects with Enter. Use poll_events to detect selections.",
	}, s.handleWidgetSelectList)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cynthia.widget.table",
		Description: "Create a panel with a table widget. Renders tabular data with box-drawing characters (┌─┬─┐). Display-only.",
	}, s.handleWidgetTable)

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

func (s *Server) handleWidgetProgressBar(ctx context.Context, req *mcp.CallToolRequest, input WidgetProgressBarInput) (*mcp.CallToolResult, SimpleOutput, error) {
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
	pb := widget.NewProgressBar(innerWidth, input.Label)
	pb.SetValue(input.Value)
	if input.Style != "" {
		pb.SetStyle(input.Style)
	}
	pb.EventQueue = s.events()
	pb.PanelID = id
	p.Widget = pb

	s.scene().Add(p)

	return nil, SimpleOutput{OK: true, ID: id, Message: "progress bar created"}, nil
}

func (s *Server) handleWidgetSparkline(ctx context.Context, req *mcp.CallToolRequest, input WidgetSparklineInput) (*mcp.CallToolResult, SimpleOutput, error) {
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
	sp := widget.NewSparkline(innerWidth, input.Label)
	if len(input.Values) > 0 {
		sp.SetValues(input.Values)
	}
	sp.EventQueue = s.events()
	sp.PanelID = id
	p.Widget = sp

	s.scene().Add(p)

	return nil, SimpleOutput{OK: true, ID: id, Message: "sparkline created"}, nil
}

func (s *Server) handleWidgetSelectList(ctx context.Context, req *mcp.CallToolRequest, input WidgetSelectListInput) (*mcp.CallToolResult, SimpleOutput, error) {
	id := input.ID
	if id == "" {
		id = scene.GenerateID()
	}

	visibleRows := input.Height
	if visibleRows <= 0 {
		visibleRows = len(input.Items)
		if visibleRows > 10 {
			visibleRows = 10
		}
	}

	panelHeight := visibleRows
	if input.Border != "" {
		panelHeight += 2
	}

	s.lock()
	defer s.unlock()

	p := scene.NewPanel(id, input.Width, panelHeight, s.renderMode())
	p.X = input.X
	p.Y = input.Y

	if input.Color != "" {
		p.Fill = canvas.Hex(input.Color)
	}
	if input.Border != "" {
		p.Border = createBorder(input.Border)
	}

	sl := widget.NewSelectList(input.Items, visibleRows)
	sl.EventQueue = s.events()
	sl.PanelID = id
	p.Widget = sl

	s.scene().Add(p)

	if input.Focus == nil || *input.Focus {
		s.scene().Focus(id)
	}

	return nil, SimpleOutput{OK: true, ID: id, Message: "select list created"}, nil
}

func (s *Server) handleWidgetTable(ctx context.Context, req *mcp.CallToolRequest, input WidgetTableInput) (*mcp.CallToolResult, SimpleOutput, error) {
	id := input.ID
	if id == "" {
		id = scene.GenerateID()
	}

	// Table height: header (1) + separator (1) + rows + top/bottom borders (2) + optional panel border
	panelHeight := 3 + len(input.Rows) // top + header + separator + rows + bottom
	if input.Border != "" {
		panelHeight += 2
	}

	s.lock()
	defer s.unlock()

	p := scene.NewPanel(id, input.Width, panelHeight, s.renderMode())
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
	tbl := widget.NewTable(input.Headers, innerWidth)
	if len(input.Rows) > 0 {
		tbl.SetRows(input.Rows)
	}
	tbl.EventQueue = s.events()
	tbl.PanelID = id
	p.Widget = tbl

	s.scene().Add(p)

	return nil, SimpleOutput{OK: true, ID: id, Message: "table created"}, nil
}
