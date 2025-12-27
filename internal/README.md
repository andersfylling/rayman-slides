# Internal Packages

Private packages for the game. Not intended for import by external projects.

| Package | Description |
|---------|-------------|
| `protocol` | Shared types, messages, version constants |
| `server` | Authoritative game server and tick loop |
| `client` | Terminal client, rendering, input |
| `game` | ECS components and game world |
| `collision` | Tile and AABB collision detection |
| `render` | Terminal renderers (ASCII, half-block, braille) |
| `input` | Keyboard capture and intent mapping |
| `network` | TCP/QUIC transport layer |
| `sync` | State snapshots and delta compression |
| `lobby` | Room codes and server discovery |

## Package Dependencies

```
cmd/rayman ──► client ──► render
                │         input
                │         network
                │         sync
                ▼
              server ──► game ──► collision
                │
                ▼
              protocol

cmd/rayserver ──► server

cmd/lookup ──► lobby
```
