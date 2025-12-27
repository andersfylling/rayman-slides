# Collision Detection

**Status:** Accepted

## Context

Need collision detection for a tile-based platformer. Must handle world geometry (tiles) and entity interactions (player, enemies, projectiles). Performance matters for multiplayer with many entities.

## Options Considered

### Grid/Tile-Based

Check which tile(s) an entity occupies. Tile has solid/passable flag.

**Workflow:** Entity moves → check destination tiles → block or allow

**Pros:**
- Very fast (O(1) tile lookup)
- Simple to implement and debug
- Natural fit for tile-based maps
- Easy to add tile types (spikes, water, ladder)

**Cons:**
- Coarse granularity (tile-sized resolution)
- Doesn't handle entity-vs-entity
- Slopes/angles need special handling

### AABB (Axis-Aligned Bounding Box)

Rectangle collision. Check if two rectangles overlap.

**Workflow:** Each entity has bounds → check overlap → resolve penetration

**Pros:**
- Simple math (min/max comparisons)
- Sub-tile precision
- Works for entity-vs-entity
- Easy collision response (push out of overlap)

**Cons:**
- O(n²) naive comparison for n entities
- Poor fit for rotated or circular shapes
- Need broadphase optimization at scale

### Pixel-Perfect

Check actual sprite pixels for overlap.

**Workflow:** Bitmask per sprite → AND masks on overlap → any bits set = collision

**Pros:**
- Precise to the pixel
- "Fair" feeling (no invisible walls)

**Cons:**
- Expensive (per-pixel checks)
- Overkill for most cases
- ASCII rendering doesn't have pixel precision anyway

### Spatial Hashing

Divide world into cells. Only check entities in same/adjacent cells.

**Workflow:** Hash entity position → bucket → check only bucket neighbors

**Pros:**
- Reduces O(n²) to O(n) average case
- Good for evenly distributed entities
- Simple to implement

**Cons:**
- Overhead not worth it for few entities (<50)
- Cell size tuning needed
- Entities spanning cells checked multiple times

### Quadtree

Hierarchical space partitioning. Subdivide regions with many entities.

**Workflow:** Insert entities into tree → query region → get candidates

**Pros:**
- Adapts to entity distribution
- Efficient for clustered entities
- Good for large sparse worlds

**Cons:**
- More complex than spatial hashing
- Tree rebalancing on movement
- Overkill for small entity counts

## Decision

**Layered approach:**

1. **Tile-based for world geometry** - Floors, walls, platforms, hazards
2. **AABB for entity-vs-entity** - Player, enemies, projectiles, pickups
3. **Spatial hashing if needed** - Add later if entity count causes slowdown

Tile collision handles 90% of cases cheaply. AABB handles interactions. No need for pixel-perfect since we're rendering to ASCII anyway.

```
collision/
  tiles.go      # Tile flag checks (solid, hazard, platform)
  aabb.go       # Rectangle overlap + resolution
  spatial.go    # Optional broadphase (add when needed)
```

## Consequences

- **Two systems** - Tile checks and AABB checks run separately
- **Tile types** - Need flags: solid, platform (pass-through from below), hazard, ladder
- **AABB resolution** - Must handle "which direction to push" for platformer feel
- **Performance headroom** - Spatial hashing ready to add if multiplayer scales up
- **No pixel precision** - Acceptable since ASCII rendering is coarse anyway
