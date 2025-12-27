package game

import (
	"github.com/andersfylling/rayman-slides/internal/collision"
	"github.com/andersfylling/rayman-slides/internal/protocol"
	"github.com/mlange-42/ark/ecs"
)

// World holds all game state using ark ECS
type World struct {
	Tick     uint64
	ECS      *ecs.World
	TileMap  *collision.TileMap
	TileSize float64 // Size of each tile in world units

	// Mappers for entity creation
	playerMapper *ecs.Map9[Position, Velocity, Collider, Sprite, Player, Health, Gravity, Grounded, Controller]
	enemyMapper  *ecs.Map7[Position, Velocity, Collider, Sprite, Health, Gravity, Grounded]

	// Filters for queries
	playerFilter  *ecs.Filter2[Position, Player]
	physicsFilter *ecs.Filter4[Position, Velocity, Gravity, Grounded]
	renderFilter  *ecs.Filter2[Position, Sprite]
	controlFilter *ecs.Filter3[Velocity, Grounded, Controller]
}

// Controller tracks which intents are active for an entity
type Controller struct {
	Intents protocol.Intent
}

// NewWorld creates a new game world
func NewWorld() *World {
	w := &World{
		TileSize: 1.0,
	}
	w.ECS = ecs.NewWorld()

	// Initialize mappers
	w.playerMapper = ecs.NewMap9[Position, Velocity, Collider, Sprite, Player, Health, Gravity, Grounded, Controller](w.ECS)
	w.enemyMapper = ecs.NewMap7[Position, Velocity, Collider, Sprite, Health, Gravity, Grounded](w.ECS)

	// Initialize filters
	w.playerFilter = ecs.NewFilter2[Position, Player](w.ECS)
	w.physicsFilter = ecs.NewFilter4[Position, Velocity, Gravity, Grounded](w.ECS)
	w.renderFilter = ecs.NewFilter2[Position, Sprite](w.ECS)
	w.controlFilter = ecs.NewFilter3[Velocity, Grounded, Controller](w.ECS)

	return w
}

// SetTileMap sets the collision tile map
func (w *World) SetTileMap(tm *collision.TileMap) {
	w.TileMap = tm
}

// Update advances the world by one tick
func (w *World) Update() {
	w.Tick++
	w.runInputSystem()
	w.runPhysicsSystem()
	w.runCollisionSystem()
}

// runInputSystem applies player intents to velocity
func (w *World) runInputSystem() {
	const moveSpeed = 0.5
	const jumpSpeed = 1.0

	query := w.controlFilter.Query()
	defer query.Close()
	for query.Next() {
		vel, grounded, ctrl := query.Get()

		// Reset horizontal velocity
		vel.X = 0

		if ctrl.Intents&protocol.IntentLeft != 0 {
			vel.X = -moveSpeed
		}
		if ctrl.Intents&protocol.IntentRight != 0 {
			vel.X = moveSpeed
		}

		// Jump only if grounded
		if ctrl.Intents&protocol.IntentJump != 0 && grounded.OnGround {
			vel.Y = -jumpSpeed
			grounded.OnGround = false
		}
	}
}

// runPhysicsSystem applies gravity and velocity
func (w *World) runPhysicsSystem() {
	const gravityAccel = 0.08

	query := w.physicsFilter.Query()
	defer query.Close()
	for query.Next() {
		pos, vel, grav, grounded := query.Get()

		// Apply gravity
		vel.Y += gravityAccel * grav.Scale

		// Cap fall speed
		if vel.Y > 1.0 {
			vel.Y = 1.0
		}

		// Apply velocity
		pos.X += vel.X
		pos.Y += vel.Y

		// Mark as not grounded (collision system will set true if on ground)
		grounded.OnGround = false
	}
}

// runCollisionSystem resolves collisions with tilemap
func (w *World) runCollisionSystem() {
	if w.TileMap == nil {
		return
	}

	query := w.physicsFilter.Query()
	defer query.Close()
	for query.Next() {
		pos, vel, _, grounded := query.Get()

		// Default collider size
		colW, colH := 0.8, 0.9

		// Check tile collision at new position
		// Check feet position
		tileX := int(pos.X)
		tileY := int(pos.Y + colH)

		// Ground collision
		if w.TileMap.IsSolid(tileX, tileY) {
			if vel.Y > 0 {
				// Landing on ground
				pos.Y = float64(tileY) - colH
				vel.Y = 0
				grounded.OnGround = true
			}
		}

		// Ceiling collision
		headTileY := int(pos.Y)
		if w.TileMap.IsSolid(tileX, headTileY) && vel.Y < 0 {
			pos.Y = float64(headTileY + 1)
			vel.Y = 0
		}

		// Wall collision (left)
		wallTileX := int(pos.X - colW/2)
		wallTileY := int(pos.Y + colH/2)
		if w.TileMap.IsSolid(wallTileX, wallTileY) {
			pos.X = float64(wallTileX+1) + colW/2
			vel.X = 0
		}

		// Wall collision (right)
		wallTileX = int(pos.X + colW/2)
		if w.TileMap.IsSolid(wallTileX, wallTileY) {
			pos.X = float64(wallTileX) - colW/2
			vel.X = 0
		}

		// Keep in bounds
		if pos.X < colW/2 {
			pos.X = colW / 2
		}
		if pos.X > float64(w.TileMap.Width)-colW/2 {
			pos.X = float64(w.TileMap.Width) - colW/2
		}
		if pos.Y < 0 {
			pos.Y = 0
		}
		if pos.Y > float64(w.TileMap.Height)-colH {
			pos.Y = float64(w.TileMap.Height) - colH
			vel.Y = 0
			grounded.OnGround = true
		}
	}
}

// SpawnPlayer creates a player entity
func (w *World) SpawnPlayer(id int, name string, x, y float64) ecs.Entity {
	return w.playerMapper.NewEntity(
		&Position{X: x, Y: y},
		&Velocity{X: 0, Y: 0},
		&Collider{Width: 0.8, Height: 0.9},
		&Sprite{Char: '@', Color: 0x00FF00},
		&Player{ID: id, Name: name},
		&Health{Current: 3, Max: 3},
		&Gravity{Scale: 1.0},
		&Grounded{OnGround: false},
		&Controller{Intents: protocol.IntentNone},
	)
}

// SpawnEnemy creates an enemy entity
func (w *World) SpawnEnemy(enemyType string, x, y float64) ecs.Entity {
	ch := 'E'
	color := uint32(0xFF0000)

	switch enemyType {
	case "slime":
		ch = 's'
		color = 0x00FF00
	case "bat":
		ch = 'b'
		color = 0x800080
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

// SetPlayerIntent sets the input intent for all players
func (w *World) SetPlayerIntent(playerID int, intents protocol.Intent) {
	query := w.controlFilter.Query()
	defer query.Close()
	for query.Next() {
		_, _, ctrl := query.Get()
		ctrl.Intents = intents
	}
}

// GetRenderables returns all entities with position and sprite for rendering
func (w *World) GetRenderables() []struct {
	X, Y  float64
	Char  rune
	Color uint32
} {
	var result []struct {
		X, Y  float64
		Char  rune
		Color uint32
	}

	query := w.renderFilter.Query()
	defer query.Close()
	for query.Next() {
		pos, sprite := query.Get()
		result = append(result, struct {
			X, Y  float64
			Char  rune
			Color uint32
		}{
			X:     pos.X,
			Y:     pos.Y,
			Char:  sprite.Char,
			Color: sprite.Color,
		})
	}

	return result
}

// GetPlayerPosition returns the first player's position
func (w *World) GetPlayerPosition() (float64, float64, bool) {
	query := w.playerFilter.Query()
	defer query.Close()
	if query.Next() {
		pos, _ := query.Get()
		return pos.X, pos.Y, true
	}
	return 0, 0, false
}
