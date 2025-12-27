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

## Consequences

- **Multiple renderers** - More code, but graceful degradation
- **Runtime detection** - Must query terminal capabilities at startup
- **Asset pipeline** - Sprites need multiple representations or runtime conversion
- **Testing complexity** - Need to test each render mode
- **User override** - Flags allow forcing mode for preference or compatibility
- **No sixel** - Sacrifices max quality for broad compatibility
