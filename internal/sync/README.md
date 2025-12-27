# sync

State synchronization between server and clients.

## Components

### Delta Compression

Only send what changed since the client's last acknowledged state.

```go
baseline := sync.NewBaseline()
baseline.Update(lastAckedSnapshot)

delta := sync.Diff(baseline, currentEntities)
// delta contains only changed/new/removed entities
```

### Snapshot Buffer

Client buffers multiple snapshots for smooth interpolation.

```go
buffer := sync.NewSnapshotBuffer(3)
buffer.Add(snapshot)

// Get two snapshots to interpolate between
from, to := buffer.Get()
// Render at position between from and to
```

## Flow

```
Server                              Client
   │                                   │
   │──── Full Snapshot (tick 0) ──────►│ (baseline)
   │                                   │
   │◄──── Ack tick 0 ─────────────────│
   │                                   │
   │──── Delta (tick 1, base 0) ──────►│
   │──── Delta (tick 2, base 0) ──────►│
   │                                   │
   │◄──── Ack tick 2 ─────────────────│
   │                                   │
   │──── Delta (tick 3, base 2) ──────►│
```

## Interpolation

Client renders between two buffered states for smooth visuals, even with network jitter.

See `adr/2025-12-27-state-synchronization.md`.
