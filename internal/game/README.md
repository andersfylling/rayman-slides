# game

ECS components and game world. Uses [mlange-42/ark](https://github.com/mlange-42/ark) for entity management.

## Components

| Component | Description |
|-----------|-------------|
| `Position` | X, Y coordinates |
| `Velocity` | Movement speed |
| `Collider` | AABB hitbox |
| `Sprite` | ASCII character and color |
| `Player` | Player ID and name |
| `Health` | Current/max HP |
| `Damage` | Damage amount (for projectiles) |
| `Gravity` | Gravity multiplier |
| `Grounded` | Is touching ground |

## World

```go
world := game.NewWorld()
world.SpawnPlayer(1, "Alice", 100, 200)
world.SpawnEnemy("slime", 300, 200)

// Each tick
world.Update()
```

## Systems (run order)

1. **Input** - Apply player intents to velocity
2. **Physics** - Apply gravity, velocity to position
3. **Collision** - Detect and resolve overlaps
4. **Damage** - Process hits, reduce health
5. **Cleanup** - Remove dead entities

## ECS Library

Using ark for:
- Type-safe component access via generics
- Entity relationships (player → projectile)
- Serialization via ark-serde (for networking)

See `adr/2025-12-27-ecs-library.md`.
