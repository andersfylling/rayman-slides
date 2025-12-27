# Game Loop: Tick-Based with Server Authority

**Status:** Accepted

## Context

Multiplayer platformer needs deterministic state. Time-based loops cause drift and race conditions between clients. Need single-player and multiplayer to use the same code path.

## Decision

- **Tick-based loop** - Server advances game state in discrete ticks
- **Server is authoritative** - All game logic runs on server
- **Server as library** - Client embeds server code for localhost/single-player
- **Standalone server binary** - Same code, separate build for dedicated hosting
- **Semver handshake** - Client/server exchange versions on connect, reject incompatible

## Architecture

```
pkg/
  server/       # Game logic, tick loop, state
  client/       # Rendering, input capture, network
  protocol/     # Shared types, version const

cmd/
  rayman/       # Client binary (embeds server for local play)
  rayserver/    # Dedicated server binary
```

## Consequences

- **Deterministic** - Same inputs = same outputs, no time drift
- **Replay support** - Can record/replay tick inputs
- **Easier netcode** - Clients send inputs, receive state snapshots
- **Latency visible** - Input delay noticeable, need client-side prediction later
- **Single code path** - Local and networked play use identical logic
