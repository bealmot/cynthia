# Cynthia — Capabilities & Possibilities

## Architecture

Cynthia is a layered rendering pipeline for terminals:

```
Effect (procedural math) → Canvas (pixel buffer) → Rasterize (pixels→cells) → ANSI output
```

Each layer is independent and composable. The key insight: Cynthia treats the terminal as a **pixel display** rendered with Unicode characters, not a text grid. Anything expressible as `f(x, y, t) → color` can be rendered.

## Current Capabilities

### Three Render Modes
The same pixel data visualized differently:
- **HalfBlock** — `▀` with FG/BG encoding 2 vertical pixels per cell (smooth color)
- **Braille** — 2x4 dot matrix per cell, 8 pixels per cell (highest resolution)
- **ASCII** — luminance-to-density glyph mapping (`. : - = + * # % @`)

### Compositing Engine
`compose.Compositor` supports z-ordered layers with 4 blend modes:
- **Normal** — standard Porter-Duff src-over
- **Screen** — additive brightening: `1 - (1-src)(1-dst)`
- **Additive** — pure addition (clamped)
- **Multiply** — darkening via pixel multiplication

Layers have independent opacity, visibility, and z-order.

### Spring-Animated State Machine
The Director uses harmonica springs to smoothly interpolate between parameter values. No abrupt jumps — everything eases in/out with configurable frequency and damping.

### LLM-Controllable via JSON Directives
```json
{"mood": "alert", "intensity": 0.8, "effect": "fire", "borderStyle": "cascade"}
```
The `Directive` struct lets an AI control the visual atmosphere organically.

### Built-in Effects
| Effect | Algorithm | Vibe |
|--------|-----------|------|
| gradient | Angle-projected color sweep + cycling | Calm, smooth |
| plasma | 4x overlapping sine waves | Thinking, trippy |
| fire | DOOM PSX heat propagation | Alert, intense |
| particles | 2D fountain emitter with gravity | Celebration |
| starfield | 3D parallax tunnel | Dreaming, spacey |
| matrix | Falling digital rain columns | Hacker |

### Border Animations
- **Pulse** — breathing glow (sinusoidal intensity)
- **Cascade** — traveling highlight racing around the frame
- **Nouveau** — art nouveau palette cycling along the perimeter

### Color Pipeline
- Premultiplied alpha for efficient compositing
- Palette with interpolated sampling
- Graceful degradation: TrueColor → ANSI256 → ANSI16 → monochrome
- Differential ANSI rendering (only writes changed cells)

---

## Possible Extensions

### Visual Effects
- **Fluid simulation** — 2D Navier-Stokes on the pixel grid
- **Metaballs** — organic blobby shapes that merge and split
- **Voronoi/Delaunay patterns** — procedural cell structures
- **Perlin/simplex noise** — terrain, fog, clouds
- **Conway's Game of Life** — cellular automata as background
- **Lissajous curves / spirograph** — animated mathematical art
- **Fractal zoom** — Mandelbrot/Julia set with animated zoom
- **Ray marching** — SDF rendering for complex 3D (torus, boolean ops)
- **Bloom/glow post-processing** — blur bright pixels, add back via compositor

### Architectural Extensions
- **Shader-like per-pixel functions** — `PixelShader(x, y, t) → Color` for one-liner effects
- **Easing library** — cubic-bezier, elastic, bounce alongside springs
- **Scene graph** — nested layers with transforms (translate, scale, rotate)
- **Particle system v2** — attractors, repulsors, force fields
- **Pixel-level text rendering** — anti-aliased/stylized text on the pixel grid
- **Input regions** — clickable/hoverable canvas zones
- **Recording/export** — capture frames to GIF or ANSI-art files

### Compositing Techniques
- **Multiply** noise textures over effects for gritty/vintage looks
- **Screen** blend two effects for additive light mixing
- **Animated masks** — one layer's luminance masks another
- **Picture-in-picture** — small effect previews overlaid on main view
- **Transition effects** — wipes, dissolves, crossfades via overlay layer

### Border Extensions
- **Typing/typewriter** — characters appear one by one
- **Glitch** — random segments shift or corrupt
- **Fire border** — fire effect constrained to perimeter
- **Loading border** — progress indicator built into the frame
- **Nested borders** — double/triple frames with different animations

### Integration Patterns
- **Status-aware CLI tools** — background mood driven by build/deploy status
- **Music visualizer** — audio FFT data piped into effect parameters
- **Chat UI** — mood driven by sentiment analysis
- **System dashboard** — metrics mapped to visual intensity
- **Terminal game engine** — canvas + compositor + input = 2D game engine
- **Generative art tool** — expose params as TUI sliders, screenshot compositions

---

## Showcase Demo

`examples/showcase/` demonstrates 4 capabilities:

1. **Diffused Sphere** — Phong-shaded sphere in 4 render modes (HalfBlock, Braille, ASCII, Normal-Map) with rotating light
2. **Mirror Cube** — 3D cube with perspective projection, painter's algorithm, rainbow environment reflection
3. **ASCII Waves** — layered sine-wave ocean with depth gradients, foam, and stars
4. **Animated Borders** — dotted (scrolling pattern), gradient (rainbow perimeter), colorshift (hue rotation + traveling highlight)

Run: `go run ./examples/showcase/`
Controls: `n`/`right`/`space` next, `p`/`left` prev, `q` quit.

---

## Comparable Projects

Most terminal UI libraries (bubbletea, tview, termui) treat the terminal as a text grid. The closest comparable is [tcell](https://github.com/gdamore/tcell) for low-level access, but Cynthia adds the pixel abstraction, compositing, and effect orchestration layers. The procedural effect system draws from demoscene traditions (DOOM fire, plasma, starfield) adapted for the terminal medium.
