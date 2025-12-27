package render

import (
	"github.com/andersfylling/rayman-slides/internal/game"
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// Color represents an RGB color
type Color struct {
	R, G, B uint8
}

// Common colors
var (
	ColorBlack   = Color{0, 0, 0}
	ColorWhite   = Color{255, 255, 255}
	ColorRed     = Color{255, 0, 0}
	ColorGreen   = Color{0, 255, 0}
	ColorBlue    = Color{0, 0, 255}
	ColorYellow  = Color{255, 255, 0}
	ColorCyan    = Color{0, 255, 255}
	ColorMagenta = Color{255, 0, 255}
)

// Camera represents the viewport into the game world
type Camera struct {
	X, Y          float64 // Center position in world coordinates
	Width, Height float64 // Viewport size in world units
}

// InputEvent represents a user input event
type InputEvent struct {
	Type   InputEventType
	Intent protocol.Intent // For key events
	Quit   bool            // For quit events
}

// InputEventType identifies the kind of input event
type InputEventType int

const (
	InputNone InputEventType = iota
	InputKey
	InputQuit
	InputResize
)

// GameRenderer is the high-level interface for all rendering backends.
// Both terminal and graphical renderers implement this interface.
type GameRenderer interface {
	// Lifecycle
	Init() error
	Close()

	// Frame management
	BeginFrame()
	EndFrame()

	// World rendering - renderer handles its own sprite atlas internally
	RenderWorld(world *game.World, camera Camera)

	// UI rendering
	RenderText(x, y float64, text string, color Color)

	// Input - each renderer handles its own input mechanism
	PollInput() (InputEvent, bool)

	// Viewport info (in world units for terminal, pixels for SDL)
	ViewportSize() (width, height float64)
}

// TileRenderer is an optional interface for renderers that support tile-based rendering
type TileRenderer interface {
	GameRenderer
	RenderTileMap(tiles [][]rune, camera Camera)
}

// Atlas is the interface for sprite lookup - each renderer has its own implementation
type Atlas interface {
	Get(spriteID string) interface{} // Returns renderer-specific sprite data
	Has(spriteID string) bool
}
