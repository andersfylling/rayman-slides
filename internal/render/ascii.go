package render

// ASCIIRenderer renders using plain ASCII characters
type ASCIIRenderer struct {
	width, height int
	// TODO: Screen buffer
}

// NewASCIIRenderer creates an ASCII renderer
func NewASCIIRenderer() *ASCIIRenderer {
	return &ASCIIRenderer{}
}

func (r *ASCIIRenderer) Init() error {
	// TODO: Initialize terminal (raw mode, hide cursor)
	r.width, r.height = 80, 24 // Default, should query terminal
	return nil
}

func (r *ASCIIRenderer) Close() {
	// TODO: Restore terminal
}

func (r *ASCIIRenderer) Clear() {
	// TODO: Clear buffer
}

func (r *ASCIIRenderer) SetCell(x, y int, ch rune, fg, bg Color) {
	// TODO: Set cell in buffer with ANSI 16-color approximation
}

func (r *ASCIIRenderer) Flush() {
	// TODO: Write buffer to stdout
}

func (r *ASCIIRenderer) Size() (int, int) {
	return r.width, r.height
}
