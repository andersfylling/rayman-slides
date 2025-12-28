# Gio Rendering

**Status:** Accepted
**Supersedes:** [2025-12-27-terminal-rendering.md](./2025-12-27-terminal-rendering.md)

## Context

The original design used terminal rendering via tcell. This had a fundamental limitation: terminals don't provide key release events, only key presses. This made charge-release mechanics (hold to charge, release to fire) unreliable.

After implementing workarounds (timing-based key release detection), we found the approach fundamentally flawed for action game input.

## Decision

**Switch to Gio for rendering and input.**

[Gio](https://gioui.org) is a pure Go immediate-mode UI toolkit with:
- Native Wayland support (no XWayland needed)
- Proper KeyDown/KeyUp events from the OS
- No CGO required
- GPU-accelerated rendering

### Architecture

```
┌──────────────────┐
│   Input System   │  internal/input/gio.go
│  (Gio key events)│  - Receives key.Event from Gio
│                  │  - Proper Press/Release states
└────────┬─────────┘
         │ KeyDown/KeyUp
         ▼
┌──────────────────┐
│    Key State     │  Tracks which keys are held
│                  │  Converts to Intent bitmask
└────────┬─────────┘
         │ Intents
         ▼
┌──────────────────┐
│   Game World     │  internal/game/world.go
│                  │  - Processes intents
│                  │  - Updates game state
└────────┬─────────┘
         │ Renderables
         ▼
┌──────────────────┐
│   Gio Renderer   │  internal/render/gio.go
│                  │  - Pure display, no input
│                  │  - Draws colored rectangles
└──────────────────┘
```

### Key Differences from Terminal

| Aspect | Terminal (tcell) | Gio |
|--------|------------------|-----|
| Key Release | Simulated via timing | Real OS events |
| Rendering | ASCII/Unicode chars | GPU rectangles |
| Platform | Any terminal | Wayland/X11/Windows |
| CGO | No | No |
| Focus | Automatic (terminal) | Click to focus |

### Build Tags

Gio code uses `//go:build gio` tag:
- `internal/input/gio.go`
- `internal/render/gio.go`
- `cmd/rayman-gui/main.go`

Build with: `go build -tags gio`

## Consequences

- **Proper input handling** - Charge-release mechanics work correctly
- **No terminal support** - Game requires a graphical environment
- **Click to focus** - User must click window before keyboard works
- **Native Wayland** - No XWayland dependency
- **Simpler input code** - No timing hacks or key repeat detection
- **Camera clamping** - Viewport clamps to map edges (no dead space)

## Future Work

- Replace colored rectangles with actual sprites
- Add sprite atlas for Gio renderer
- Consider keeping terminal renderer as optional (with limited mechanics)
