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
// Uses abstract sprite IDs - renderers map these to their native format
type Sprite struct {
	ID    string // Sprite identifier (e.g., "player", "slime", "platform")
	Color uint32 // RGB color hint (renderers may use or ignore)
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

// AttackState tracks attack animation state
type AttackState struct {
	Attacking   bool // Currently in attack animation
	TicksLeft   int  // Animation ticks remaining
	FacingRight bool // Direction of attack

	// Charging state
	IsCharging  bool   // Currently holding attack key
	ChargeStart uint64 // Tick when charge started

	// Release debounce - prevents false releases from terminal key repeat gaps
	ReleaseDebounce int // Frames since attack key was last detected as released
}

// ReleaseDebounceThreshold is how many consecutive frames the key must be
// released before we actually launch the fist. With intentHoldDuration at 600ms
// (exceeding terminal initial key repeat delay), we can use a small value for
// responsive release.
const ReleaseDebounceThreshold = 2

// AttackDuration is how many ticks the punch animation lasts
const AttackDuration = 8

// Charge constants
const (
	MaxChargeTicks  = 180  // 3 seconds at 60 TPS = maximum charge
	MinFistDistance = 1.0  // Minimum distance (no charge)
	MaxFistDistance = 20.0 // Maximum distance (full charge) - 20x character width
	FistSpeed       = 0.8  // Speed of the flying fist per tick
)

// Fist component marks a flying fist projectile
type Fist struct {
	StartX       float64 // Starting X position
	MaxDistance  float64 // Maximum distance to travel
	FacingRight  bool    // Direction of travel
	OwnerID      int     // Player who threw the fist
}
