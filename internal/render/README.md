# render

Terminal rendering backends. Auto-detects capabilities and picks the best renderer.

## Renderers

| Renderer | Resolution | Colors | Compatibility |
|----------|------------|--------|---------------|
| ASCII | 1x1 per char | ANSI 16 | Everything |
| HalfBlock | 1x2 per char | Truecolor/256 | Most modern terminals |
| Braille | 2x4 per char | Truecolor | Terminals with good Unicode |

## Usage

```go
cap := render.Detect()
renderer := render.SelectRenderer(cap, render.ModeAuto)

renderer.Init()
defer renderer.Close()

renderer.Clear()
renderer.SetCell(10, 5, '@', render.ColorWhite, render.ColorBlack)
renderer.Flush()
```

## Auto-Detection

Checks environment variables:
- `COLORTERM=truecolor` → HalfBlock with truecolor
- `TERM=*256color*` → HalfBlock with 256 colors
- Otherwise → ASCII fallback

## Force Mode

```bash
./rayman --ascii    # Force ASCII
./rayman --braille  # Force Braille
```

See `adr/2025-12-27-terminal-rendering.md`.
