package game

import (
	"fmt"

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
	attackMapper *ecs.Map1[AttackState] // Separate mapper for attack state
	fistMapper   *ecs.Map4[Position, Velocity, Sprite, Fist]
	fistChecker  *ecs.Map1[Fist] // For checking if entity has Fist component

	// Filters for queries
	playerFilter  *ecs.Filter2[Position, Player]
	physicsFilter *ecs.Filter4[Position, Velocity, Gravity, Grounded]
	renderFilter  *ecs.Filter2[Position, Sprite]
	controlFilter *ecs.Filter3[Velocity, Grounded, Controller]
	attackFilter  *ecs.Filter6[Position, Sprite, Controller, AttackState, Velocity, Player]
	fistFilter    *ecs.Filter3[Position, Velocity, Fist]
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
	w.attackMapper = ecs.NewMap1[AttackState](w.ECS)
	w.fistMapper = ecs.NewMap4[Position, Velocity, Sprite, Fist](w.ECS)
	w.fistChecker = ecs.NewMap1[Fist](w.ECS)

	// Initialize filters
	w.playerFilter = ecs.NewFilter2[Position, Player](w.ECS)
	w.physicsFilter = ecs.NewFilter4[Position, Velocity, Gravity, Grounded](w.ECS)
	w.renderFilter = ecs.NewFilter2[Position, Sprite](w.ECS)
	w.controlFilter = ecs.NewFilter3[Velocity, Grounded, Controller](w.ECS)
	w.attackFilter = ecs.NewFilter6[Position, Sprite, Controller, AttackState, Velocity, Player](w.ECS)
	w.fistFilter = ecs.NewFilter3[Position, Velocity, Fist](w.ECS)

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
	w.runAttackSystem()
	w.runFistSystem()
	w.runPhysicsSystem()
	w.runCollisionSystem()
}

// runInputSystem applies player intents to velocity
func (w *World) runInputSystem() {
	const moveSpeed = 0.5
	const jumpSpeed = 1.0

	query := w.controlFilter.Query()
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

// runAttackSystem handles charge-release attack mechanics.
// Press attack key to start charging, release to fire.
// Longer charge = greater fist travel distance.
func (w *World) runAttackSystem() {
	// Collect fists to spawn (can't spawn during query iteration)
	type fistSpawn struct {
		x, y        float64
		facingRight bool
		distance    float64
		ownerID     int
	}
	var fistsToSpawn []fistSpawn

	query := w.attackFilter.Query()
	for query.Next() {
		pos, sprite, ctrl, attack, vel, player := query.Get()

		// Update facing direction from velocity
		if vel.X > 0 {
			attack.FacingRight = true
		} else if vel.X < 0 {
			attack.FacingRight = false
		}

		attackPressed := ctrl.Intents&protocol.IntentAttack != 0

		// Detect edges
		attackJustPressed := attackPressed && !attack.AttackWasPressed
		attackJustReleased := !attackPressed && attack.AttackWasPressed

		// Update state for next frame's edge detection
		attack.AttackWasPressed = attackPressed

		// Start charging on key press (if not in cooldown)
		if attackJustPressed && !attack.Attacking && !attack.Charging {
			attack.Charging = true
			attack.ChargeTicks = 0
		}

		// Continue charging while key held
		if attack.Charging && attackPressed {
			attack.ChargeTicks++
			if attack.ChargeTicks > MaxChargeTicks {
				attack.ChargeTicks = MaxChargeTicks
			}
		}

		// Fire on key release
		if attackJustReleased && attack.Charging {
			// Calculate distance based on charge time
			chargeRatio := float64(attack.ChargeTicks) / float64(MaxChargeTicks)
			distance := MinFistDistance + chargeRatio*(MaxFistDistance-MinFistDistance)

			fistsToSpawn = append(fistsToSpawn, fistSpawn{
				x:           pos.X,
				y:           pos.Y,
				facingRight: attack.FacingRight,
				distance:    distance,
				ownerID:     player.ID,
			})

			// End charging, start punch animation
			attack.Charging = false
			attack.ChargeTicks = 0
			attack.Attacking = true
			attack.TicksLeft = AttackCooldown
		}

		// Update sprite based on state
		if attack.Charging {
			// Charging animation - 3 levels based on charge progress
			chargeLevel := 1
			if attack.ChargeTicks > MaxChargeTicks/3 {
				chargeLevel = 2
			}
			if attack.ChargeTicks > MaxChargeTicks*2/3 {
				chargeLevel = 3
			}

			if attack.FacingRight {
				sprite.ID = fmt.Sprintf("player_charge_right_%d", chargeLevel)
			} else {
				sprite.ID = fmt.Sprintf("player_charge_left_%d", chargeLevel)
			}
		} else if attack.Attacking {
			if attack.TicksLeft > 0 {
				attack.TicksLeft--
				// Punch animation (arm extended)
				if attack.FacingRight {
					sprite.ID = "player_punch_right"
				} else {
					sprite.ID = "player_punch_left"
				}
			} else {
				attack.Attacking = false
				sprite.ID = "player"
			}
		} else {
			sprite.ID = "player"
		}
	}

	// Spawn fists after query completes
	for _, f := range fistsToSpawn {
		w.SpawnFist(f.x, f.y, f.facingRight, f.distance, f.ownerID)
	}
}

// runFistSystem updates flying fist projectiles
func (w *World) runFistSystem() {
	// Collect entities to remove (can't remove during query)
	var toRemove []ecs.Entity

	query := w.fistFilter.Query()
	for query.Next() {
		pos, vel, fist := query.Get()
		entity := query.Entity()

		// Move the fist
		pos.X += vel.X

		// Check if fist has traveled max distance
		traveled := pos.X - fist.StartX
		if !fist.FacingRight {
			traveled = -traveled
		}

		if traveled >= fist.MaxDistance {
			toRemove = append(toRemove, entity)
		}
	}

	// Remove fists that have traveled their distance
	for _, e := range toRemove {
		w.ECS.RemoveEntity(e)
	}
}

// SpawnFist creates a flying fist projectile
// The fist spawns at chest height (0.5 units above the character's foot position)
func (w *World) SpawnFist(x, y float64, facingRight bool, maxDistance float64, ownerID int) ecs.Entity {
	velX := FistSpeed
	spriteID := "fist_right"
	if !facingRight {
		velX = -FistSpeed
		spriteID = "fist_left"
	}

	// Offset Y to chest level (character position is at feet, chest is about 0.5 units up)
	chestY := y - 0.5

	return w.fistMapper.NewEntity(
		&Position{X: x, Y: chestY},
		&Velocity{X: velX, Y: 0},
		&Sprite{ID: spriteID, Color: 0xFFFF00},
		&Fist{
			StartX:      x,
			MaxDistance: maxDistance,
			FacingRight: facingRight,
			OwnerID:     ownerID,
		},
	)
}

// runPhysicsSystem applies gravity and velocity
func (w *World) runPhysicsSystem() {
	const gravityAccel = 0.08

	query := w.physicsFilter.Query()
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
	entity := w.playerMapper.NewEntity(
		&Position{X: x, Y: y},
		&Velocity{X: 0, Y: 0},
		&Collider{Width: 0.8, Height: 0.9},
		&Sprite{ID: "player", Color: 0x00FF00},
		&Player{ID: id, Name: name},
		&Health{Current: 3, Max: 3},
		&Gravity{Scale: 1.0},
		&Grounded{OnGround: false},
		&Controller{Intents: protocol.IntentNone},
	)
	// Add attack state component
	w.attackMapper.Add(entity, &AttackState{FacingRight: true})
	return entity
}

// SpawnEnemy creates an enemy entity
func (w *World) SpawnEnemy(enemyType string, x, y float64) ecs.Entity {
	spriteID := enemyType // Use enemy type as sprite ID
	color := uint32(0xFF0000)

	switch enemyType {
	case "slime":
		color = 0x00FF00
	case "bat":
		color = 0x800080
	default:
		spriteID = "enemy"
	}

	return w.enemyMapper.NewEntity(
		&Position{X: x, Y: y},
		&Velocity{X: 0, Y: 0},
		&Collider{Width: 0.8, Height: 0.8},
		&Sprite{ID: spriteID, Color: color},
		&Health{Current: 1, Max: 1},
		&Gravity{Scale: 1.0},
		&Grounded{OnGround: false},
	)
}

// SetPlayerIntent sets the input intent for all players
func (w *World) SetPlayerIntent(playerID int, intents protocol.Intent) {
	query := w.controlFilter.Query()
	for query.Next() {
		_, _, ctrl := query.Get()
		ctrl.Intents = intents
	}
}

// Renderable represents an entity that can be drawn
type Renderable struct {
	X, Y     float64
	SpriteID string
	Color    uint32 // Color hint (renderers may use their atlas colors instead)
	FlipX    bool   // Flip sprite horizontally (facing left)
}

// GetRenderables returns all entities with position and sprite for rendering
func (w *World) GetRenderables() []Renderable {
	var result []Renderable

	query := w.renderFilter.Query()
	for query.Next() {
		pos, sprite := query.Get()
		entity := query.Entity()

		// Determine facing direction
		flipX := false

		// Check if entity has AttackState (player) - FacingRight=false means facing left
		if w.attackMapper.HasAll(entity) {
			attack := w.attackMapper.Get(entity)
			flipX = !attack.FacingRight
		}

		// Check if entity is a Fist - FacingRight=false means facing left
		if w.fistChecker.HasAll(entity) {
			fist := w.fistChecker.Get(entity)
			flipX = !fist.FacingRight
		}

		result = append(result, Renderable{
			X:        pos.X,
			Y:        pos.Y,
			SpriteID: sprite.ID,
			Color:    sprite.Color,
			FlipX:    flipX,
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
