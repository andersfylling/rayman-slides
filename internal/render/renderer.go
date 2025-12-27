package render

// Color represents an RGB color
type Color struct {
	R, G, B uint8
}

// Renderer is the interface for all rendering backends
type Renderer interface {
	// Init initializes the terminal for rendering
	Init() error

	// Close restores the terminal to normal state
	Close()

	// Clear clears the screen
	Clear()

	// SetCell sets a character at the given position with colors
	SetCell(x, y int, ch rune, fg, bg Color)

	// Flush writes the buffer to the terminal
	Flush()

	// Size returns the terminal dimensions
	Size() (width, height int)
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
