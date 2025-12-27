# Contributing to rayman-slides

Thanks for your interest in contributing! This document explains how to get involved.

## Getting Started

1. **Fork and clone** the repository
2. **Install Go 1.22+** if you haven't already
3. **Build and test** to verify your setup:
   ```bash
   make deps
   make build
   make test
   ```
4. **Run the game** to see it in action:
   ```bash
   ./bin/rayman
   ```

## Finding Work

### Good First Issues

Look for issues labeled `good first issue` - these are suitable for newcomers and have clear scope.

### Areas Needing Help

- **Gameplay:** New abilities, enemy behaviors, level design
- **Rendering:** Improve terminal graphics, add effects
- **Networking:** Multiplayer features, latency handling
- **Testing:** Add test coverage (see test personas below)
- **Documentation:** Tutorials, examples, API docs

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

Use prefixes:
- `feature/` - New functionality
- `fix/` - Bug fixes
- `docs/` - Documentation
- `refactor/` - Code improvements

### 2. Make Changes

- Follow the code style in [AGENTS.md](AGENTS.md)
- Add tests for new functionality
- Update documentation if needed

### 3. Test Your Changes

```bash
make test    # Run all tests
make fmt     # Format code
make lint    # Check for issues (requires golangci-lint)
```

### 4. Commit

Write clear commit messages:

```
Add slime enemy with basic AI

- Slime bounces left/right until hitting a wall
- Takes 1 hit to defeat
- Drops 5 tings on death
```

### 5. Open a Pull Request

- Fill out the PR template
- Link any related issues
- Wait for CI to pass
- Respond to review feedback

## Code Architecture

### Key Concepts

| Concept | Location | Description |
|---------|----------|-------------|
| ECS World | `internal/game/world.go` | Game state and systems |
| Components | `internal/game/components.go` | Entity data |
| Collision | `internal/collision/` | Tile + AABB detection |
| Rendering | `internal/render/` | Terminal output |
| Input | `internal/input/` | Keyboard handling |
| Protocol | `internal/protocol/` | Network messages |

### Adding a New Enemy

See [docs/examples/new-enemy.md](docs/examples/new-enemy.md) for a walkthrough.

### Adding a New Ability

1. Add intent to `internal/protocol/messages.go`
2. Add input binding in `internal/input/capture.go`
3. Handle intent in `internal/game/world.go` input system
4. Add any new components needed

## Testing Guidelines

We use test personas to ensure comprehensive coverage. When writing tests, consider:

| Persona | Focus |
|---------|-------|
| Happy Path | Normal usage, expected inputs |
| Edge Case | Boundaries, zero, max, empty |
| Saboteur | Invalid inputs, errors, nil |
| Concurrency | Race conditions, parallel access |
| Performance | Benchmarks, allocations |
| Replayer | Determinism, reproducibility |

See [adr/2025-12-27-ci-testing.md](adr/2025-12-27-ci-testing.md) for details.

### Example Test Structure

```go
func TestCollision_PlayerOnGround(t *testing.T) {
    // Arrange
    tm := collision.NewTileMap(10, 10)
    tm.Set(5, 9, collision.TileSolid)

    // Act
    isSolid := tm.IsSolid(5, 9)

    // Assert
    if !isSolid {
        t.Error("expected tile to be solid")
    }
}
```

## Questions?

- Open an issue for bugs or feature requests
- Check existing issues and ADRs before proposing major changes
- Be respectful and constructive in discussions

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
