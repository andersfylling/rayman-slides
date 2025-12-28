# Input Handling

**Status:** Accepted (Partially Superseded)
**See also:** [2025-12-28-gio-rendering.md](./2025-12-28-gio-rendering.md)

## Context

Tick-based game with networked server. Need to capture player input and transmit to server. Network latency affects responsiveness.

**Note:** The original design assumed terminal rendering which lacks key-up events. The switch to Gio rendering (see linked ADR) provides real KeyUp/KeyDown events, making the "Simulated Key State" approach obsolete. The intent-based tick buffering design remains valid.

## Terminal Input Constraints (Historical)

> **Note:** These constraints applied to the original terminal-based design.
> With Gio rendering, we have proper KeyDown/KeyUp events from the OS.

Unlike graphical games, terminals typically:
- Only report key-down, not key-up
- Have limited modifier detection
- Vary by terminal emulator and OS
- May buffer input or miss rapid presses

Libraries like tcell/bubbletea abstract some of this, but limitations remain.

## Options Considered

### Raw Key Events

Send every key press directly to server as it happens.

**Workflow:** Key pressed → send to server → server processes

**Pros:**
- Simple implementation
- Low latency for single key actions

**Cons:**
- High network traffic (key per packet)
- No key-up means "hold to run" is hard
- Server must handle out-of-order packets
- No tick alignment

### Intent-Based

Convert key presses to game intents. Send intents, not keys.

**Workflow:** Key pressed → map to intent (Jump, MoveLeft) → send intent

**Pros:**
- Decouples input from action (rebindable keys)
- Smaller vocabulary (fewer message types)
- Server doesn't know/care about key bindings
- Can merge redundant intents

**Cons:**
- Still no key-up for "hold" actions
- Need intent vocabulary design upfront

### Buffered Per-Tick

Collect all input during a tick window. Send as single batch.

**Workflow:** Collect inputs → on tick boundary → batch send → server applies

**Pros:**
- Aligned with server tick rate
- One packet per tick (efficient)
- Deterministic input ordering
- Easier replay recording

**Cons:**
- Adds up to one tick of latency
- Must track tick boundaries on client

### Simulated Key State (No Longer Needed)

> **Superseded:** With Gio rendering, we have real KeyUp events from the OS.

Since terminals lack key-up, simulate held state with timeouts.

**Workflow:** Key pressed → mark as "held" → decay after N ms with no repeat → key-up

**Pros:**
- Enables "hold to run/charge" mechanics
- Works around terminal limitations

**Cons:**
- Feels slightly different from native key-up
- Timeout tuning needed
- Rapid tap vs hold ambiguity

### Client-Side Prediction with Rollback

Client predicts outcome locally, server corrects if wrong.

**Workflow:** Input → predict locally → send to server → receive authoritative state → rollback if mismatch

**Pros:**
- Feels responsive despite latency
- Standard netcode pattern for action games

**Cons:**
- Complex to implement correctly
- Visual "snapping" on misprediction
- Must be able to re-simulate ticks
- Overkill for turn-based or slow-paced

## Decision

**Intent-based with tick buffering:**

1. **Key → Intent mapping** - Client converts keys to intents (configurable bindings)
2. **Real key state** - Gio provides native KeyDown/KeyUp events (simulated hold no longer needed)
3. **Tick-aligned buffer** - Collect intents, send once per tick
4. **Intent frame** - Each tick's input is a set of active intents

```go
type Intent uint8

const (
    IntentNone Intent = 0
    IntentLeft Intent = 1 << iota
    IntentRight
    IntentJump
    IntentAttack
    IntentUse
)

type InputFrame struct {
    Tick    uint64
    Intents Intent  // Bitmask of active intents
}
```

Client-side prediction deferred. Start simple, add if latency is painful.

```
input/
  keys.go       # GameKey enum, KeyEvent types
  gio.go        # Gio key capture (real KeyDown/KeyUp)
  keystate.go   # Key state tracking, Intent conversion
  buffer.go     # Tick-aligned batching
```

## Consequences

- **Rebindable keys** - Users can customize controls
- **Tick alignment** - Inputs arrive in predictable order
- **Real key events** - Gio provides native KeyDown/KeyUp, enabling precise charge-release mechanics
- **Single packet per tick** - Efficient network usage
- **No prediction yet** - May feel laggy on high-latency connections (revisit later)
- **Bitmask intents** - Compact representation, max ~8 simultaneous intents (expandable)
