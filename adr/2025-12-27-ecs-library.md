# ECS Library Selection

**Status:** Accepted

## Context

Need an Entity Component System for game entities. Rolling our own adds maintenance burden. Two Go libraries considered: andygeiss/ecs and mlange-42/ark.

## Options Considered

### andygeiss/ecs

- 167 stars, last release Sept 2024
- Zero dependencies
- Interface-based API (Component, System, Entity managers)
- Bitmask filtering for entity queries
- Simple, traditional ECS pattern

**Pros:** Minimal, easy to understand, stable
**Cons:** No recent activity, no serialization, no entity relationships

### mlange-42/ark

- 194 stars, last release Dec 2025 (v0.7.0)
- Zero dependencies
- Generics-based API with query/filter pattern
- Entity relationships as first-class feature
- Built-in event system with filtering
- Batch operations for spawn/despawn
- Addon: ark-serde for JSON serialization

**Pros:** Active, modern Go, serialization for netcode, relationships for hierarchies
**Cons:** Slightly more complex API, younger project

## Decision

Use **mlange-42/ark**.

Entity relationships simplify player→projectile ownership. Serialization addon directly supports network state snapshots. Active maintenance reduces risk of abandonment.

## Consequences

- **Serialization** - ark-serde enables state snapshots for multiplayer
- **Relationships** - Clean parent-child for projectiles, effects, UI elements
- **Events** - Decouple damage/death/spawn without tight coupling
- **Learning curve** - Query-based API differs from traditional ECS tutorials
- **Dependency risk** - If abandoned, migration needed (but same risk with any lib)
