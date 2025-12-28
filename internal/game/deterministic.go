package game

import (
	"hash/fnv"

	"github.com/andersfylling/rayman-slides/internal/protocol"
	"github.com/mlange-42/ark/ecs"
)

// EntityState captures the full state of an entity for snapshot/restore
type EntityState struct {
	Entity    ecs.Entity
	Position  Position
	Velocity  Velocity
	Grounded  Grounded
	HasPlayer bool
	Player    Player
	HasAttack bool
	Attack    AttackState
}

// WorldState is a complete snapshot of the game world for rollback
type WorldState struct {
	Tick     uint64
	Entities []EntityState
	Checksum uint32
}

// Snapshot creates a complete snapshot of the current world state
// This captures all entity states needed for rollback and replay
func (w *World) Snapshot() WorldState {
	state := WorldState{
		Tick:     w.Tick,
		Entities: make([]EntityState, 0),
	}

	// Capture all physics entities (players and enemies)
	query := w.physicsFilter.Query()
	for query.Next() {
		entity := query.Entity()
		pos, vel, _, grounded := query.Get()

		es := EntityState{
			Entity:   entity,
			Position: *pos,
			Velocity: *vel,
			Grounded: *grounded,
		}

		// Check if this entity has Player component
		playerQuery := w.playerFilter.Query()
		for playerQuery.Next() {
			if playerQuery.Entity() == entity {
				_, player := playerQuery.Get()
				es.HasPlayer = true
				es.Player = *player
				break
			}
		}
		playerQuery.Close()

		// Check if this entity has AttackState component
		attackQuery := w.attackFilter.Query()
		for attackQuery.Next() {
			if attackQuery.Entity() == entity {
				_, _, _, attack, _, _ := attackQuery.Get()
				es.HasAttack = true
				es.Attack = *attack
				break
			}
		}
		attackQuery.Close()

		state.Entities = append(state.Entities, es)
	}

	// Calculate checksum for fast comparison
	state.Checksum = state.computeChecksum()

	return state
}

// Restore applies a saved world state, rolling back to that point in time
func (w *World) Restore(state WorldState) {
	w.Tick = state.Tick

	for _, es := range state.Entities {
		// Find and update the entity
		// We use the physics filter since all relevant entities have physics
		query := w.physicsFilter.Query()
		for query.Next() {
			if query.Entity() == es.Entity {
				pos, vel, _, grounded := query.Get()
				*pos = es.Position
				*vel = es.Velocity
				*grounded = es.Grounded
				break
			}
		}
		query.Close()

		// Restore attack state if present
		if es.HasAttack {
			attackQuery := w.attackFilter.Query()
			for attackQuery.Next() {
				if attackQuery.Entity() == es.Entity {
					_, _, _, attack, _, _ := attackQuery.Get()
					*attack = es.Attack
					break
				}
			}
			attackQuery.Close()
		}
	}
}

// computeChecksum calculates a fast hash for comparing world states
func (state *WorldState) computeChecksum() uint32 {
	h := fnv.New32a()

	// Hash tick
	tickBytes := make([]byte, 8)
	tickBytes[0] = byte(state.Tick)
	tickBytes[1] = byte(state.Tick >> 8)
	tickBytes[2] = byte(state.Tick >> 16)
	tickBytes[3] = byte(state.Tick >> 24)
	tickBytes[4] = byte(state.Tick >> 32)
	tickBytes[5] = byte(state.Tick >> 40)
	tickBytes[6] = byte(state.Tick >> 48)
	tickBytes[7] = byte(state.Tick >> 56)
	h.Write(tickBytes)

	// Hash each entity's position (most important for mismatch detection)
	for _, es := range state.Entities {
		// Convert float64 to bytes for hashing
		// Using a simple representation - position * 1000 to preserve some precision
		posX := int64(es.Position.X * 1000)
		posY := int64(es.Position.Y * 1000)

		posBytes := make([]byte, 16)
		posBytes[0] = byte(posX)
		posBytes[1] = byte(posX >> 8)
		posBytes[2] = byte(posX >> 16)
		posBytes[3] = byte(posX >> 24)
		posBytes[4] = byte(posX >> 32)
		posBytes[5] = byte(posX >> 40)
		posBytes[6] = byte(posX >> 48)
		posBytes[7] = byte(posX >> 56)
		posBytes[8] = byte(posY)
		posBytes[9] = byte(posY >> 8)
		posBytes[10] = byte(posY >> 16)
		posBytes[11] = byte(posY >> 24)
		posBytes[12] = byte(posY >> 32)
		posBytes[13] = byte(posY >> 40)
		posBytes[14] = byte(posY >> 48)
		posBytes[15] = byte(posY >> 56)
		h.Write(posBytes)
	}

	return h.Sum32()
}

// StatesMatch compares two world states for equivalence within tolerance
func StatesMatch(a, b *WorldState, tolerance float64) bool {
	// Quick checksum comparison
	if a.Checksum == b.Checksum {
		return true
	}

	// If checksums differ, do detailed comparison
	if len(a.Entities) != len(b.Entities) {
		return false
	}

	for i := range a.Entities {
		ea := &a.Entities[i]
		eb := &b.Entities[i]

		// Compare positions within tolerance
		dx := ea.Position.X - eb.Position.X
		dy := ea.Position.Y - eb.Position.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}

		if dx > tolerance || dy > tolerance {
			return false
		}

		// Compare grounded state
		if ea.Grounded.OnGround != eb.Grounded.OnGround {
			return false
		}
	}

	return true
}

// ToProtocolSnapshot converts a WorldState to a protocol.StateSnapshot for network transmission
func (state *WorldState) ToProtocolSnapshot() protocol.StateSnapshot {
	snapshot := protocol.StateSnapshot{
		Tick:     state.Tick,
		Full:     true,
		Entities: make([]protocol.EntityState, 0, len(state.Entities)),
	}

	for _, es := range state.Entities {
		// Serialize entity state to bytes
		// Format: [posX:8][posY:8][velX:8][velY:8][grounded:1][hasPlayer:1][playerID:4]
		data := make([]byte, 0, 38)

		// Position (as int64 * 1000 for precision)
		posX := int64(es.Position.X * 1000)
		posY := int64(es.Position.Y * 1000)
		data = appendInt64(data, posX)
		data = appendInt64(data, posY)

		// Velocity
		velX := int64(es.Velocity.X * 1000)
		velY := int64(es.Velocity.Y * 1000)
		data = appendInt64(data, velX)
		data = appendInt64(data, velY)

		// Grounded
		if es.Grounded.OnGround {
			data = append(data, 1)
		} else {
			data = append(data, 0)
		}

		// Player info
		if es.HasPlayer {
			data = append(data, 1)
			data = appendInt32(data, int32(es.Player.ID))
		} else {
			data = append(data, 0)
		}

		snapshot.Entities = append(snapshot.Entities, protocol.EntityState{
			ID:         protocol.EntityID(es.Entity.ID()),
			Components: data,
		})
	}

	return snapshot
}

// Helper functions for byte serialization
func appendInt64(data []byte, v int64) []byte {
	return append(data,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56),
	)
}

func appendInt32(data []byte, v int32) []byte {
	return append(data,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
	)
}
