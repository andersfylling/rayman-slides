# rayman-slides

An open-source 2D side-scrolling platformer inspired by the classic Rayman 1 (1995), developed using Claude Code.

## Quick Start

```bash
# Clone and build
git clone https://github.com/andersfylling/rayman-slides.git
cd rayman-slides
make build

# Run the game
./bin/rayman

# Controls: WASD or arrow keys to move, Q or Esc to quit
```

### Requirements

- Go 1.22+
- A terminal with color support

## About

This project recreates the gameplay and feel of Rayman 1 using entirely custom, AI-generated assets. No original game data is used—all sprites, music, and level designs are original creations that evoke the whimsical fairy-tale aesthetic of the source material.

### Key Features

- Limbless character with progressive ability unlocks (telescopic fist, grappling, helicopter hair, running)
- 6 thematic worlds mixing natural and imaginary landscapes
- Collectible-based progression (free all caged creatures to unlock the final world)
- Boss battles at the end of each world
- Swappable "Profiles" (asset packs) for different visual themes
- **Terminal-based rendering** - runs in any terminal

### Why "Slides"?

The project uses a "profile" system allowing different skins/asset packs to slide in and out, keeping the core mechanics while enabling visual customization.

## Project Structure

```
cmd/
  rayman/       # Game client
  rayserver/    # Dedicated multiplayer server
  lookup/       # Room code discovery service

internal/
  game/         # ECS world, components, systems
  collision/    # Tile-based + AABB collision
  render/       # Terminal rendering (ASCII, half-block, braille)
  input/        # Keyboard capture, intent mapping
  network/      # Client-server transport
  sync/         # State synchronization
  lobby/        # Room codes for multiplayer
  protocol/     # Wire format, messages

adr/            # Architectural Decision Records
docs/           # Reference material
```

## Development

```bash
# Build all binaries
make build

# Run tests
make test

# Run the client
make run

# Run the dedicated server
make server

# Format code
make fmt

# Clean build artifacts
make clean
```

## Architecture

See the [adr/](adr/) folder for architectural decisions:

- [Language: Go](adr/2025-12-27-language-go.md)
- [Game Loop: Tick-based](adr/2025-12-27-game-loop-tick-based.md)
- [ECS: mlange-42/ark](adr/2025-12-27-ecs-library.md)
- [Network: Client-server TCP](adr/2025-12-27-network-architecture.md)
- [Rendering: Tiered terminal](adr/2025-12-27-terminal-rendering.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Tech Stack

- **Language:** Go
- **ECS:** [mlange-42/ark](https://github.com/mlange-42/ark)
- **Terminal:** [gdamore/tcell](https://github.com/gdamore/tcell)

## Status

Early development - basic movement and collision working.

## License

See [LICENSE](LICENSE)
