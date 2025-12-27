// Package game defines ECS components and game logic.
package game

// Position component
type Position struct {
	X, Y float64
}

// Velocity component
type Velocity struct {
	X, Y float64
}

// Collider component (AABB bounds relative to position)
type Collider struct {
	OffsetX, OffsetY float64
	Width, Height    float64
}

// Sprite component (for rendering)
type Sprite struct {
	Char  rune   // ASCII representation
	Color uint32 // RGB color
}

// Player component (marks player-controlled entities)
type Player struct {
	ID   int
	Name string
}

// Health component
type Health struct {
	Current int
	Max     int
}

// Damage component (for projectiles, hazards)
type Damage struct {
	Amount int
}

// Gravity component (affected by gravity)
type Gravity struct {
	Scale float64 // Multiplier (1.0 = normal, 0 = no gravity)
}

// Grounded component (touching ground)
type Grounded struct {
	OnGround bool
}
