# Sprite and Asset Format

**Status:** Accepted

## Context

Need a format for sprites and visual assets. Requirements:
- Single source of truth (not duplicated per render mode)
- Standard tooling (image editors, AI generators)
- Converts to ASCII/half-block/braille at build or runtime

## Options Considered

### YAML ASCII Art

Define sprites directly as ASCII text in YAML.

**Pros:**
- Human-readable
- Git-friendly diffs
- AI can generate text

**Cons:**
- Not a real image (can't preview easily)
- Must maintain ASCII manually
- Different "source" per render mode
- Artists can't use standard tools

### PNG as Source of Truth

Standard PNG images, converted to terminal representation.

**Workflow:** Create/generate PNG → converter produces ASCII/half-block/braille

**Pros:**
- Standard format, any image editor works
- AI image generators output PNG directly
- Single source, multiple outputs
- Preview sprites visually
- Existing sprite art can be imported

**Cons:**
- Need conversion tooling
- Conversion quality depends on algorithm
- Larger files than text

### SVG as Source of Truth

Vector graphics, rasterize then convert.

**Pros:**
- Scales to any resolution
- Small file size
- Text-based (git-friendly)

**Cons:**
- Overkill for small pixel art
- Extra rasterization step
- Less tooling than PNG

### Aseprite Format (.ase/.aseprite)

Native format from popular pixel art tool.

**Pros:**
- Layers, animation frames built-in
- Industry standard for pixel art

**Cons:**
- Proprietary format
- Requires Aseprite or compatible tools
- AI tools don't output this format

## Decision

**PNG as source of truth with build-time conversion:**

```
assets/
  sprites/
    player/
      idle.png       # Source image (e.g., 16x24 pixels)
      walk.png       # Can be sprite sheet with multiple frames
      walk.json      # Optional metadata (frame count, timing)
    enemies/
      slime.png
```

### Conversion Pipeline

```
PNG → Converter → Multiple outputs

                  ┌─► ascii.go    (plain ASCII: @#%*+-.  )
idle.png ────────►├─► halfblock.go (▀▄█ with colors)
                  └─► braille.go  (⠁⠂⠃...⣿ patterns)
```

Converter runs at:
- **Build time** - Pre-generate all representations, embed in binary
- **Runtime** - Convert on load (slower startup, but more flexible)

Recommend build-time for release, runtime for development.

### Conversion Algorithm

For each render mode:

**ASCII:**
```
1. Convert to grayscale
2. Map brightness to character: " .:-=+*#%@"
3. Optionally map hue to ANSI color
```

**Half-block:**
```
1. For each 1x2 pixel pair (top, bottom):
   - Top pixel → foreground color
   - Bottom pixel → background color
   - Output: ▀ with fg/bg colors
```

**Braille:**
```
1. For each 2x4 pixel block:
   - Threshold each pixel to on/off
   - Map to braille dot pattern (8 dots = 256 patterns)
   - Output: braille character + color
```

### Metadata File (Optional)

For animations and sprite sheets:

```json
{
  "frames": 4,
  "frame_width": 16,
  "frame_height": 24,
  "fps": 8,
  "anchor": [8, 24]
}
```

Or infer from filename: `walk_4x1_8fps.png` (4 frames, 1 row, 8 FPS)

### Converter Tool

```bash
# Convert single sprite
./tools/sprite-convert idle.png --mode ascii

# Convert all sprites, output Go code
./tools/sprite-convert assets/sprites/ --output internal/sprites/generated.go

# Preview in terminal
./tools/sprite-convert idle.png --preview
```

### AI Image Generation

Can use any image AI:
- DALL-E, Midjourney, Stable Diffusion for concepts
- Pixel art specific models for cleaner results
- Google's models, etc.

Prompt example:
```
16x24 pixel art sprite of a cartoon character running,
side view, transparent background, retro game style
```

Then convert the output PNG.

### Directory Structure

```
assets/
  sprites/
    player/
      idle.png
      walk.png
      walk.json       # Animation metadata
      jump.png
      attack.png
    enemies/
      slime.png
      bat.png
    items/
      coin.png
      heart.png
    tiles/
      ground.png
      platform.png

tools/
  sprite-convert/     # Conversion tool
    main.go
    ascii.go
    halfblock.go
    braille.go
```

## Integration with Game Logic

**Critical: Game logic is decoupled from sprite visuals.**

The game world uses abstract sprite IDs, not visual data:

```go
// Game component - knows nothing about how sprites look
type Sprite struct {
    ID    string  // "player", "slime", "platform"
    Color uint32  // Color hint (optional)
}

// Spawning an entity
world.SpawnEnemy("slime", x, y)  // Uses sprite ID "slime"
```

### Sprite Atlases

Each renderer maintains a `SpriteAtlas` that maps IDs to its native format:

```go
// ASCII atlas
atlas.Set("player", SpriteData{Char: '@', FG: ColorGreen})
atlas.Set("slime",  SpriteData{Char: 's', FG: ColorGreen})

// Half-block atlas (different representation, same IDs)
atlas.Set("player", SpriteData{Char: '█', FG: ColorGreen})
atlas.Set("slime",  SpriteData{Char: '▄', FG: ColorGreen})
```

### Converter Output

The sprite converter produces atlas entries, not raw images:

```bash
# Convert PNGs to Go atlas code
./tools/sprite-convert assets/sprites/ --output internal/render/sprites_gen.go
```

Output:
```go
func GeneratedASCIIAtlas() *SpriteAtlas {
    atlas := NewSpriteAtlas()
    atlas.Set("player_idle", SpriteData{Char: '@', FG: Color{0, 255, 0}})
    atlas.Set("player_walk", SpriteData{Char: '@', FG: Color{0, 255, 0}})
    atlas.Set("slime", SpriteData{Char: 's', FG: Color{0, 200, 0}})
    // ...
    return atlas
}
```

### Flow

```
PNG Source → Converter → Sprite Atlas (per renderer)
                              ↓
Game World (sprite IDs) → Renderer (looks up atlas) → Screen
```

This allows:
- Multiple renderers showing the same game state
- Easy theming (swap atlas, keep game logic)
- Future graphical renderers without changing game code

## Consequences

- **Standard tooling** - Any image editor or AI tool works
- **Single source** - PNG is truth, conversions are derived
- **Visual preview** - Can see sprites without running game
- **Import existing art** - Can use existing pixel art sprites
- **Build step** - Need to run converter (can automate in Makefile)
- **Conversion quality** - Results depend on algorithm tuning
- **Larger assets** - PNGs bigger than text, but still tiny for pixel art
- **Decoupled rendering** - Game logic uses IDs, renderers handle visuals
