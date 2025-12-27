# Language: Go

**Status:** Accepted

## Context

Need a language for a terminal-based platformer with multiplayer. Requirements: type safety, good performance, easy concurrency for networking, ability to drop to low-level when needed.

## Decision

Use **Go** as the primary language.

## Consequences

- **Simple concurrency** - Goroutines/channels make multiplayer networking straightforward
- **Fast compilation** - Quick iteration during development
- **Single binary** - Easy distribution, no runtime dependencies
- **Type safe** - Catches errors at compile time, no need for external type checkers
- **CGo/assembly escape hatch** - Can write assembly or call C when needed for performance-critical code
- **Good terminal libraries** - tcell, bubbletea, termbox-go available
- **No generics baggage** - Modern Go has generics if needed for ECS
- **Trade-off:** GC pauses possible, but manageable for a terminal game
