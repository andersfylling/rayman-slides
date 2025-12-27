# input

Keyboard capture and intent mapping.

## Concepts

**Key** → **Intent** → **InputFrame**

- Keys are raw keyboard input (a, d, space, etc.)
- Intents are game actions (MoveLeft, Jump, Attack)
- InputFrames bundle intents for a single tick

## Usage

```go
handler := input.NewHandler()

// Customize bindings
handler.Bind('z', protocol.IntentJump)

// On key event
handler.OnKeyPress('a')

// Get current state
intents := handler.State()  // IntentLeft

// Buffer for network
buffer := input.NewBuffer()
buffer.Add(intents)
buffer.Tick()

// Send to server
frames := buffer.Flush()
```

## Default Bindings

| Key | Intent |
|-----|--------|
| A / ← | Move left |
| D / → | Move right |
| W / Space | Jump |
| J | Attack |
| K | Use |

## Terminal Limitations

Terminals don't reliably report key-up events. We simulate "held" state by detecting repeated key presses within a threshold.

See `adr/2025-12-27-input-handling.md`.
