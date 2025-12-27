# State Synchronization

**Status:** Accepted

## Context

Server is authoritative and runs game logic. Clients need to receive game state to render. Must balance bandwidth, latency feel, and implementation complexity.

## Options Considered

### Full State Every Tick

Send complete world state to all clients every tick.

**Workflow:** Server tick → serialize all entities → broadcast → clients replace state

**Pros:**
- Dead simple (no diff tracking)
- Client can join mid-game trivially
- No cumulative drift/corruption
- Easy debugging (state is always complete)

**Cons:**
- High bandwidth (scales with world size)
- Redundant data (most things don't change each tick)
- May exceed reasonable packet sizes

### Delta Compression

Track changes, send only what changed since client's last acknowledged state.

**Workflow:** Server tracks per-client baseline → diff against current → send delta → client applies

**Pros:**
- Much smaller packets (only changes)
- Scales to larger worlds
- Industry standard approach

**Cons:**
- Must track per-client baseline
- Complexity in diff/patch logic
- Lost packet = must send full state or accumulate deltas
- Need acknowledgment system

### Entity Interest / Area of Interest

Only send entities near or relevant to each player.

**Workflow:** Determine player's visible area → filter entities → send subset

**Pros:**
- Dramatic bandwidth reduction for large worlds
- Players don't see irrelevant entities anyway
- Scales to MMO-size

**Cons:**
- Edge cases (entity enters view)
- Must track what each client knows about
- Complexity in interest management
- Overkill for single-screen platformer

### Snapshot Interpolation

Client buffers multiple state snapshots, renders interpolated state between them.

**Workflow:** Receive states → buffer 2-3 → render between oldest two → smooth visuals

**Pros:**
- Smooths out network jitter
- Hides minor packet delays
- Visual polish

**Cons:**
- Adds latency (rendering behind real state)
- Buffer management complexity
- Must handle missing snapshots gracefully

### Client-Side Prediction

Client predicts outcome of own inputs immediately, server corrects if wrong.

**Workflow:** Input → apply locally → send to server → receive authoritative → reconcile

**Pros:**
- Feels responsive (immediate feedback)
- Hides round-trip latency
- Standard for FPS/action games

**Cons:**
- Complex reconciliation (rollback and replay)
- Visual "snapping" on misprediction
- Must re-simulate potentially many ticks
- Bug-prone

## Decision

**Delta compression with snapshot interpolation, prediction deferred:**

1. **Delta compression** - Send changes only, reduces bandwidth significantly
2. **Periodic full state** - Every N ticks or on client request (resync safety net)
3. **Snapshot interpolation** - Client buffers 2-3 states, interpolates for smooth render
4. **No prediction initially** - Add later if latency is unacceptable

Rationale: Platformer at 30-60 ticks has small entity counts. Delta keeps packets small. Interpolation smooths jitter without prediction complexity. Full state fallback ensures recovery from corruption.

```go
type StateSnapshot struct {
    Tick     uint64
    Full     bool              // True = complete state, False = delta
    Baseline uint64            // If delta, relative to this tick
    Entities []EntityState     // Full or changed entities
    Removed  []EntityID        // Entities removed since baseline
}

type EntityState struct {
    ID         EntityID
    Components []byte          // Serialized components (ark-serde)
}
```

Interpolation buffer:
```
[tick 100] [tick 101] [tick 102]
     ^           ^
  render      latest
  from        received
```

Client renders between tick 100→101 while 102 arrives.

```
sync/
  snapshot.go     # Snapshot serialization
  delta.go        # Diff and patch logic
  baseline.go     # Per-client baseline tracking
  interpolate.go  # Client-side interpolation buffer
```

## Consequences

- **Reduced bandwidth** - Delta means small packets for typical ticks
- **Smooth rendering** - Interpolation hides jitter without prediction
- **Added latency** - Interpolation buffer adds ~2 ticks of visual delay
- **Resync safety** - Periodic full state prevents drift accumulation
- **No instant response** - Own actions delayed by round trip (acceptable for platformer)
- **Prediction upgrade path** - Can add later for own-player only if needed
