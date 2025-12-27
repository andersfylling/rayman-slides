package game

// World holds all game state
// TODO: Replace with ark ECS world
type World struct {
	Tick uint64
	// TODO: ark.World
	// TODO: Component mappers
	// TODO: Filters/queries
}

// NewWorld creates a new game world
func NewWorld() *World {
	return &World{}
}

// Update advances the world by one tick
func (w *World) Update() {
	w.Tick++
	// TODO: Run systems:
	// 1. Input system - apply player intents
	// 2. Physics system - apply velocity, gravity
	// 3. Collision system - detect and resolve
	// 4. Damage system - handle hits
	// 5. Cleanup system - remove dead entities
}

// SpawnPlayer creates a player entity
func (w *World) SpawnPlayer(id int, name string, x, y float64) {
	// TODO: Use ark to create entity with components:
	// Position, Velocity, Collider, Sprite, Player, Health, Gravity, Grounded
}

// SpawnEnemy creates an enemy entity
func (w *World) SpawnEnemy(enemyType string, x, y float64) {
	// TODO: Create enemy entity based on type
}
