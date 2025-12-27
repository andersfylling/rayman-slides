# Terminal Rendering

**Status:** Accepted

## Context

Game runs in terminal. Need to render sprites, tiles, and UI as text characters. Must balance visual quality with terminal compatibility.

## Options Considered

### Plain ASCII

Standard printable characters only: `#.@*^o=|-`

**Workflow:** Map tile/entity types to single characters

**Pros:**
- Works in every terminal ever made
- Classic roguelike aesthetic
- Tiny memory footprint
- Easy to reason about (1 char = 1 tile)

**Cons:**
- Very low resolution
- Limited expressiveness (hard to distinguish entities)
- Color limited to ANSI 16 for compatibility

**Example:**
```
#######
#..@..#    @ = player, . = floor, # = wall
#..*..#    * = enemy
#######
```

### Block Elements

Unicode box drawing and block characters: `█▀▄░▒▓│─┌┐└┘`

**Workflow:** Combine blocks for shapes, use shading for depth

**Pros:**
- More visual variety than plain ASCII
- Can represent 2 vertical "pixels" per character (▀▄)
- Good font support (most monospace fonts include these)

**Cons:**
- Still coarse resolution
- Some older terminals/fonts missing glyphs
- Limited to rectangular shapes

**Example:**
```
████████████
█░░░░░▓▓░░░█
█░░█▀▀█░░░░█
████████████
```

### Braille Patterns

Unicode braille: `⠁⠂⠃⠄...⣿` (256 patterns, 2x4 dot grid per char)

**Workflow:** Treat each char as 2x4 pixel grid, set dots as needed

**Pros:**
- Highest resolution (8 "pixels" per character cell)
- Can render actual sprite shapes
- Full 256 patterns available

**Cons:**
- Thin/dotty appearance (not solid fills)
- Many fonts render braille poorly or not at all
- Accessibility concerns (screen readers)
- Complex encoding logic

**Example:**
```
⣿⣿⣿⣿⣿⣿⣿⣿
⣿⠀⠀⣴⣿⣦⠀⣿
⣿⠀⠀⣿⣿⣿⠀⣿
⣿⣿⣿⣿⣿⣿⣿⣿
```

### Half-Block with Color

Use `▀▄█` with foreground/background colors for 2 pixels per char.

**Workflow:** Each character cell = 2 vertical pixels, fg color = top, bg color = bottom

**Pros:**
- 2x vertical resolution over plain text
- Solid fills (not dotty like braille)
- Truecolor support (16 million colors) in modern terminals
- Widely supported block characters

**Cons:**
- Requires truecolor terminal for best results
- Falls back poorly on 16-color terminals
- Horizontal resolution still 1:1

**Example:**
```
▄▄▄▄▄▄▄▄▄▄
█  ▄██▄  █   (with colors: brown walls, blue sky,
█  ████  █    red character)
▀▀▀▀▀▀▀▀▀▀
```

### Sixel / Kitty Graphics Protocol

Actual inline images in terminal.

**Workflow:** Encode sprite as sixel or kitty protocol → terminal renders image

**Pros:**
- Full graphical fidelity
- Can use actual sprite artwork
- Looks like a "real" game

**Cons:**
- Very limited terminal support (kitty, iTerm2, some others)
- No Windows Terminal support (as of 2024)
- Complex protocol handling
- Falls back to nothing on unsupported terminals
- Defeats the "terminal game" aesthetic

## Decision

**Tiered rendering with runtime detection:**

1. **Default: Half-block with ANSI 256** - Good balance of quality and compatibility
2. **Fallback: Plain ASCII with ANSI 16** - For limited terminals
3. **Enhanced: Half-block with truecolor** - When detected/configured
4. **Optional: Braille mode** - User-selectable for high-res preference

Detection order:
```
COLORTERM=truecolor → truecolor half-block
TERM contains "256color" → ANSI 256 half-block
Otherwise → Plain ASCII fallback
--ascii flag → Force ASCII mode
--braille flag → Force braille mode
```

```
render/
  detect.go     # Terminal capability detection
  ascii.go      # Plain ASCII renderer
  halfblock.go  # Half-block renderer (default)
  braille.go    # Braille renderer (optional)
  color.go      # Color palette management
```

## Rendering Architecture

**Critical principle: Game logic is decoupled from rendering.**

The game world uses abstract sprite IDs, not visual representations:

```
Game World (abstract)              Renderers (concrete)
┌─────────────────────┐           ┌──────────────────┐
│ Position{X, Y}      │           │ ASCII Renderer   │
│ Sprite{ID: "player"}│──────────►│ atlas["player"]  │
│ Collider{W, H}      │           │   → '@' green    │
└─────────────────────┘           └──────────────────┘
                                  ┌──────────────────┐
                      ───────────►│ Half-block       │
                                  │ atlas["player"]  │
                                  │   → '█' green    │
                                  └──────────────────┘
                                  ┌──────────────────┐
                      ───────────►│ SDL/Vulkan       │
                                  │ atlas["player"]  │
                                  │   → player.png   │
                                  └──────────────────┘
```

### Sprite Atlas

Each renderer maintains a `SpriteAtlas` that maps sprite IDs to its native format:

```go
// Game component - abstract
type Sprite struct {
    ID    string  // "player", "slime", "platform"
    Color uint32  // Color hint (renderer may ignore)
}

// Renderer-specific lookup
type SpriteAtlas struct {
    sprites map[string]SpriteData
}

atlas := render.DefaultASCIIAtlas()
sprite := atlas.Get("player")  // → {Char: '@', FG: green}
```

### Benefits

- **Same game state, multiple views** - Run ASCII and graphical renderers simultaneously
- **Easy theming** - Swap atlas without touching game logic
- **Testable** - Game logic doesn't depend on rendering
- **Future-proof** - Add new renderers (Vulkan, SDL) without changing game code

### GameRenderer Interface

The high-level interface that all renderers implement:

```go
type GameRenderer interface {
    // Lifecycle
    Init() error
    Close()

    // Frame management
    BeginFrame()
    EndFrame()

    // World rendering - renderer handles its own sprite atlas
    RenderWorld(world *game.World, camera Camera)

    // UI rendering
    RenderText(x, y float64, text string, color Color)

    // Input - each renderer handles its own input mechanism
    PollInput() (InputEvent, bool)

    // Viewport info
    ViewportSize() (width, height float64)
}
```

This interface works for both terminal and graphical renderers:
- Terminal: positions are character cells, input via stdin
- SDL/Vulkan: positions are pixels, input via SDL events

### File Structure

```
render/
  renderer.go    # GameRenderer interface, Camera, InputEvent types
  atlas.go       # SpriteAtlas type, default atlases
  detect.go      # Terminal capability detection, renderer selection
  tcell.go       # tcell-based terminal renderer (implements GameRenderer)
```

Future graphical renderers would be added as separate packages:

```
render/
  sdl/           # SDL2 renderer (build tag: sdl)
    renderer.go
    atlas.go
  vulkan/        # Vulkan renderer (build tag: vulkan)
    renderer.go
    atlas.go
```

## Consequences

- **Multiple renderers** - More code, but graceful degradation
- **Runtime detection** - Must query terminal capabilities at startup
- **Sprite atlases** - Each renderer maintains its own sprite mappings
- **Decoupled architecture** - Game logic never knows how it's being rendered
- **Testing complexity** - Need to test each render mode
- **User override** - Flags allow forcing mode for preference or compatibility
- **No sixel** - Sacrifices max quality for broad compatibility
