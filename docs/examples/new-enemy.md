# How to Add a New Enemy

This guide walks through adding a new enemy type to the game.

## Overview

Enemies in rayman-slides are ECS entities with specific components. To add a new enemy:

1. Define its visual appearance and behavior
2. Add a spawn function or extend the existing one
3. (Optional) Add AI behavior via a new system

## Example: Adding a "Bat" Enemy

Let's add a bat that flies back and forth.

### Step 1: Define the Enemy Type

The `SpawnEnemy` function in `internal/game/world.go` already supports different enemy types. Let's see how the bat is defined:

```go
func (w *World) SpawnEnemy(enemyType string, x, y float64) ecs.Entity {
    ch := 'E'
    color := uint32(0xFF0000)

    switch enemyType {
    case "slime":
        ch = 's'
        color = 0x00FF00
    case "bat":
        ch = 'b'           // ASCII character
        color = 0x800080   // Purple color
    }

    return w.enemyMapper.NewEntity(
        &Position{X: x, Y: y},
        &Velocity{X: 0, Y: 0},
        &Collider{Width: 0.8, Height: 0.8},
        &Sprite{Char: ch, Color: color},
        &Health{Current: 1, Max: 1},
        &Gravity{Scale: 1.0},
        &Grounded{OnGround: false},
    )
}
```

### Step 2: Spawn the Enemy in a Level

In `cmd/rayman/main.go` or your level loader:

```go
world.SpawnEnemy("bat", 20, 5)  // x=20, y=5
```

### Step 3: Add Unique Behavior (Optional)

If your enemy needs special behavior (like flying), you need to:

1. **Add a component** to mark entities with this behavior
2. **Add a filter** to query these entities
3. **Add a system** to update them each tick

#### 3a. Add Component

In `internal/game/components.go`:

```go
// FlyingAI marks an entity that flies back and forth
type FlyingAI struct {
    Direction float64 // 1.0 = right, -1.0 = left
    Speed     float64
    MinX      float64 // Turn around point
    MaxX      float64
}
```

#### 3b. Add Mapper and Filter

In `internal/game/world.go`, update the World struct:

```go
type World struct {
    // ... existing fields ...

    // New mapper for flying enemies
    flyingEnemyMapper *ecs.Map8[Position, Velocity, Collider, Sprite, Health, Gravity, Grounded, FlyingAI]

    // New filter
    flyingAIFilter *ecs.Filter2[Position, FlyingAI]
}
```

Initialize in `NewWorld()`:

```go
w.flyingEnemyMapper = ecs.NewMap8[Position, Velocity, Collider, Sprite, Health, Gravity, Grounded, FlyingAI](w.ECS)
w.flyingAIFilter = ecs.NewFilter2[Position, FlyingAI](w.ECS)
```

#### 3c. Add Spawn Function

```go
func (w *World) SpawnFlyingEnemy(enemyType string, x, y, minX, maxX float64) ecs.Entity {
    return w.flyingEnemyMapper.NewEntity(
        &Position{X: x, Y: y},
        &Velocity{X: 0, Y: 0},
        &Collider{Width: 0.8, Height: 0.8},
        &Sprite{Char: 'b', Color: 0x800080},
        &Health{Current: 1, Max: 1},
        &Gravity{Scale: 0.0},  // No gravity for flying
        &Grounded{OnGround: false},
        &FlyingAI{
            Direction: 1.0,
            Speed:     0.1,
            MinX:      minX,
            MaxX:      maxX,
        },
    )
}
```

#### 3d. Add AI System

In `internal/game/world.go`:

```go
func (w *World) runFlyingAISystem() {
    query := w.flyingAIFilter.Query()
    for query.Next() {
        pos, ai := query.Get()

        // Move in current direction
        pos.X += ai.Speed * ai.Direction

        // Reverse at boundaries
        if pos.X <= ai.MinX {
            ai.Direction = 1.0
        } else if pos.X >= ai.MaxX {
            ai.Direction = -1.0
        }
    }
}
```

Call it from `Update()`:

```go
func (w *World) Update() {
    w.Tick++
    w.runInputSystem()
    w.runFlyingAISystem()  // Add this line
    w.runPhysicsSystem()
    w.runCollisionSystem()
}
```

### Step 4: Test It

Add to your level:

```go
world.SpawnFlyingEnemy("bat", 15, 8, 10, 25)  // Flies between x=10 and x=25
```

Run the game:

```bash
make build && ./bin/rayman
```

## Tips

- **Start simple:** Get the enemy visible first, then add behavior
- **Use existing patterns:** Look at how slimes work before adding complexity
- **Test incrementally:** Verify each step works before moving on
- **Check collisions:** Make sure enemies interact with tiles correctly

## Component Reference

| Component | Purpose |
|-----------|---------|
| `Position` | Where the entity is |
| `Velocity` | How it moves each tick |
| `Collider` | Hitbox size |
| `Sprite` | Visual appearance |
| `Health` | How much damage it can take |
| `Gravity` | Scale of gravity (0 = flying) |
| `Grounded` | Is it on the ground |

## See Also

- [internal/game/world.go](../../internal/game/world.go) - ECS world and systems
- [internal/game/components.go](../../internal/game/components.go) - Component definitions
- [adr/2025-12-27-ecs-library.md](../../adr/2025-12-27-ecs-library.md) - Why we use ark
