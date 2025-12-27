# rayserver

Dedicated game server. Runs the authoritative game simulation and accepts client connections.

## Usage

```bash
# Start server on default port
./rayserver

# Custom port and player limit
./rayserver --port 7777 --max-players 4

# With room code registration
./rayserver --register --lookup https://lookup.example.com

# Load specific map
./rayserver --map maps/campaign/level1.tmj
```

## Flags

| Flag | Description |
|------|-------------|
| `--port` | Port to listen on (default: 7777) |
| `--max-players` | Maximum players (default: 4) |
| `--map` | Map file to load |
| `--register` | Register with lookup service |
| `--lookup` | Lookup service URL |
| `--name` | Server name (shown in room listing) |
| `--tick-rate` | Ticks per second (default: 60) |

## Architecture

The server:
1. Loads the map and initializes the ECS world
2. Listens for TCP connections
3. Runs a fixed-rate tick loop (default 60/sec)
4. Receives player inputs each tick
5. Simulates game state
6. Broadcasts state snapshots to clients

See `adr/2025-12-27-game-loop-tick-based.md` for details.
