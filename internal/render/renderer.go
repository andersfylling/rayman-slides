// Package render provides game rendering functionality.
package render

// Camera represents the viewport into the game world
type Camera struct {
	X, Y          float64 // Center position in world coordinates
	Width, Height float64 // Viewport size in world units
}
