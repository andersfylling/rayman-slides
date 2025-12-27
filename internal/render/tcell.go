package render

import (
	"github.com/gdamore/tcell/v2"
)

// TcellRenderer renders using tcell for cross-platform terminal support
type TcellRenderer struct {
	screen tcell.Screen
}

// NewTcellRenderer creates a new tcell-based renderer
func NewTcellRenderer() *TcellRenderer {
	return &TcellRenderer{}
}

func (r *TcellRenderer) Init() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := screen.Init(); err != nil {
		return err
	}
	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	screen.EnableMouse()
	screen.Clear()
	r.screen = screen
	return nil
}

func (r *TcellRenderer) Close() {
	if r.screen != nil {
		r.screen.Fini()
	}
}

func (r *TcellRenderer) Clear() {
	if r.screen != nil {
		r.screen.Clear()
	}
}

func (r *TcellRenderer) SetCell(x, y int, ch rune, fg, bg Color) {
	if r.screen == nil {
		return
	}
	fgColor := tcell.NewRGBColor(int32(fg.R), int32(fg.G), int32(fg.B))
	bgColor := tcell.NewRGBColor(int32(bg.R), int32(bg.G), int32(bg.B))
	style := tcell.StyleDefault.Foreground(fgColor).Background(bgColor)
	r.screen.SetContent(x, y, ch, nil, style)
}

func (r *TcellRenderer) Flush() {
	if r.screen != nil {
		r.screen.Show()
	}
}

func (r *TcellRenderer) Size() (int, int) {
	if r.screen == nil {
		return 80, 24
	}
	return r.screen.Size()
}

// Screen returns the underlying tcell screen for input handling
func (r *TcellRenderer) Screen() tcell.Screen {
	return r.screen
}

// DrawString draws a string at the given position
func (r *TcellRenderer) DrawString(x, y int, s string, fg, bg Color) {
	for i, ch := range s {
		r.SetCell(x+i, y, ch, fg, bg)
	}
}

// DrawBox draws a box border
func (r *TcellRenderer) DrawBox(x, y, width, height int, fg, bg Color) {
	// Corners
	r.SetCell(x, y, '+', fg, bg)
	r.SetCell(x+width-1, y, '+', fg, bg)
	r.SetCell(x, y+height-1, '+', fg, bg)
	r.SetCell(x+width-1, y+height-1, '+', fg, bg)

	// Top and bottom edges
	for i := 1; i < width-1; i++ {
		r.SetCell(x+i, y, '-', fg, bg)
		r.SetCell(x+i, y+height-1, '-', fg, bg)
	}

	// Left and right edges
	for i := 1; i < height-1; i++ {
		r.SetCell(x, y+i, '|', fg, bg)
		r.SetCell(x+width-1, y+i, '|', fg, bg)
	}
}
