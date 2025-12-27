// Package collision implements collision detection.
// Tile-based for world geometry, AABB for entity interactions.
package collision

// TileFlag represents collision properties of a tile
type TileFlag uint8

const (
	TileEmpty    TileFlag = 0
	TileSolid    TileFlag = 1 << iota // Blocks movement from all directions
	TilePlatform                      // Blocks from below only (pass-through)
	TileHazard                        // Damages on contact
	TileLadder                        // Allows climbing
	TileWater                         // Slows movement, allows swimming
)

// TileMap holds collision data for the world
type TileMap struct {
	Width  int
	Height int
	Tiles  []TileFlag
}

// NewTileMap creates a tile map with given dimensions
func NewTileMap(width, height int) *TileMap {
	return &TileMap{
		Width:  width,
		Height: height,
		Tiles:  make([]TileFlag, width*height),
	}
}

// Get returns the tile flag at the given position
func (m *TileMap) Get(x, y int) TileFlag {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return TileSolid // Out of bounds = solid
	}
	return m.Tiles[y*m.Width+x]
}

// Set sets the tile flag at the given position
func (m *TileMap) Set(x, y int, flag TileFlag) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}
	m.Tiles[y*m.Width+x] = flag
}

// IsSolid checks if the tile blocks movement
func (m *TileMap) IsSolid(x, y int) bool {
	return m.Get(x, y)&TileSolid != 0
}

// IsPlatform checks if the tile is a pass-through platform
func (m *TileMap) IsPlatform(x, y int) bool {
	return m.Get(x, y)&TilePlatform != 0
}
