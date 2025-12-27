# rayman

The game client. Renders the game in your terminal and handles input.

## Usage

```bash
# Local/singleplayer (starts embedded server)
./rayman

# Connect to a server
./rayman --connect 192.168.1.100:7777

# Join via room code
./rayman --room ABCD-1234

# Force ASCII rendering
./rayman --ascii

# Force braille rendering
./rayman --braille
```

## Flags

| Flag | Description |
|------|-------------|
| `--connect` | Server address (IP:port) |
| `--room` | Room code to join |
| `--name` | Player name |
| `--ascii` | Force ASCII rendering mode |
| `--braille` | Force braille rendering mode |
| `--config` | Path to config file |

## Local Play

When no `--connect` or `--room` is provided, the client starts an embedded server for local/singleplayer gameplay. This uses the same server code as the dedicated server.
