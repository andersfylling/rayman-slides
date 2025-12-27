package game

import (
	"github.com/andersfylling/rayman-slides/internal/collision"
)

// DemoLevel creates a simple test level with default size
func DemoLevel() *collision.TileMap {
	return DemoLevelForViewport(40, 20)
}

// DemoLevelForViewport creates a test level sized to fit the given viewport
func DemoLevelForViewport(width, height int) *collision.TileMap {
	// Ensure minimum size for playability
	if width < 40 {
		width = 40
	}
	if height < 20 {
		height = 20
	}
	tm := collision.NewTileMap(width, height)

	// Floor
	for x := 0; x < width; x++ {
		tm.Set(x, height-1, collision.TileSolid)
	}

	// Left wall
	for y := 0; y < height; y++ {
		tm.Set(0, y, collision.TileSolid)
	}

	// Right wall
	for y := 0; y < height; y++ {
		tm.Set(width-1, y, collision.TileSolid)
	}

	// Some platforms
	for x := 5; x < 12; x++ {
		tm.Set(x, height-5, collision.TileSolid)
	}

	for x := 15; x < 22; x++ {
		tm.Set(x, height-8, collision.TileSolid)
	}

	for x := 25; x < 32; x++ {
		tm.Set(x, height-5, collision.TileSolid)
	}

	// A small obstacle
	tm.Set(10, height-2, collision.TileSolid)
	tm.Set(10, height-3, collision.TileSolid)

	// Floating platform
	for x := 18; x < 23; x++ {
		tm.Set(x, height-12, collision.TileSolid)
	}

	return tm
}

// RenderTileMap returns ASCII representation of the tilemap
func RenderTileMap(tm *collision.TileMap) [][]rune {
	result := make([][]rune, tm.Height)
	for y := 0; y < tm.Height; y++ {
		result[y] = make([]rune, tm.Width)
		for x := 0; x < tm.Width; x++ {
			tile := tm.Get(x, y)
			switch {
			case tile&collision.TileSolid != 0:
				result[y][x] = '#'
			case tile&collision.TilePlatform != 0:
				result[y][x] = '='
			case tile&collision.TileHazard != 0:
				result[y][x] = '^'
			case tile&collision.TileLadder != 0:
				result[y][x] = 'H'
			case tile&collision.TileWater != 0:
				result[y][x] = '~'
			default:
				result[y][x] = ' '
			}
		}
	}
	return result
}
