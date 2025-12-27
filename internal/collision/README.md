# collision

Collision detection for tile-based world and entity interactions.

## Two Systems

### 1. Tile Collision

For world geometry (floors, walls, platforms).

```go
tilemap := collision.NewTileMap(100, 50)
tilemap.Set(10, 20, collision.TileSolid)

if tilemap.IsSolid(x, y) {
    // Block movement
}
```

**Tile flags:**
- `TileSolid` - Blocks from all directions
- `TilePlatform` - Pass-through from below
- `TileHazard` - Damages on contact
- `TileLadder` - Allows climbing
- `TileWater` - Slows movement

### 2. AABB Collision

For entity-vs-entity (player vs enemy, projectiles).

```go
a := collision.NewAABB(x1, y1, w1, h1)
b := collision.NewAABB(x2, y2, w2, h2)

if a.Overlaps(b) {
    dx, dy := a.Penetration(b)
    // Push apart by (dx, dy)
}
```

## Performance

For small entity counts (<50), naive O(n²) AABB checks are fine. If needed, `spatial.go` has spatial hashing for optimization.

See `adr/2025-12-27-collision-detection.md`.
