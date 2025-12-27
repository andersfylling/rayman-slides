# Command Binaries

This directory contains the main entry points for the project.

| Binary | Description |
|--------|-------------|
| `rayman` | Game client - play the game |
| `rayserver` | Dedicated server - host multiplayer games |
| `lookup` | Room code service - translates room codes to server addresses |

## Building

```bash
make build
```

Binaries are output to `bin/`.

## Running

```bash
# Play locally (embeds server)
./bin/rayman

# Host a dedicated server
./bin/rayserver --port 7777

# Run lookup service (for room codes)
./bin/lookup --port 8080
```
