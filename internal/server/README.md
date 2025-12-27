# server

Authoritative game server. Runs the tick loop and game simulation.

## Responsibilities

1. Accept client connections
2. Receive player inputs
3. Run fixed-rate tick loop (60/sec default)
4. Simulate game state (physics, collision, damage)
5. Broadcast state snapshots to clients

## Usage

```go
cfg := server.DefaultConfig()
cfg.Port = 7777
cfg.MaxPlayers = 4

srv := server.New(cfg)
if err := srv.Start(); err != nil {
    log.Fatal(err)
}
```

## Embedding

The client embeds the server for local/singleplayer:

```go
// In client code
srv := server.New(server.DefaultConfig())
go srv.Start()
// Connect to localhost
```

## Tick Loop

```
while running:
    receive_inputs()      // From all clients
    world.Update()        // Advance simulation
    broadcast_state()     // Send snapshots
    sleep_until_next_tick()
```

See `adr/2025-12-27-game-loop-tick-based.md`.
