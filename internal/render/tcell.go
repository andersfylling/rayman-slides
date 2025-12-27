package render

import (
	"github.com/andersfylling/rayman-slides/internal/game"
	"github.com/andersfylling/rayman-slides/internal/protocol"
	"github.com/gdamore/tcell/v2"
)

// TcellRenderer renders using tcell for cross-platform terminal support
type TcellRenderer struct {
	screen   tcell.Screen
	atlas    *SpriteAtlas
	tileMap  [][]rune // Cached tile map for rendering
	eventCh  chan tcell.Event
	quitCh   chan struct{}
}

// NewTcellRenderer creates a new tcell-based renderer
func NewTcellRenderer() *TcellRenderer {
	return &TcellRenderer{
		atlas:   DefaultASCIIAtlas(),
		eventCh: make(chan tcell.Event, 32),
		quitCh:  make(chan struct{}),
	}
}

// SetAtlas allows overriding the default sprite atlas
func (r *TcellRenderer) SetAtlas(atlas *SpriteAtlas) {
	r.atlas = atlas
}

// SetTileMap sets the tile map to render
func (r *TcellRenderer) SetTileMap(tiles [][]rune) {
	r.tileMap = tiles
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
	screen.Clear()
	r.screen = screen

	// Start event polling goroutine
	go r.pollEvents()

	return nil
}

func (r *TcellRenderer) pollEvents() {
	for {
		select {
		case <-r.quitCh:
			return
		default:
			ev := r.screen.PollEvent()
			if ev == nil {
				return
			}
			select {
			case r.eventCh <- ev:
			default:
				// Drop event if channel full
			}
		}
	}
}

func (r *TcellRenderer) Close() {
	close(r.quitCh)
	if r.screen != nil {
		r.screen.Fini()
	}
}

func (r *TcellRenderer) BeginFrame() {
	if r.screen != nil {
		r.screen.Clear()
	}
}

func (r *TcellRenderer) EndFrame() {
	if r.screen != nil {
		r.screen.Show()
	}
}

func (r *TcellRenderer) ViewportSize() (float64, float64) {
	if r.screen == nil {
		return 80, 24
	}
	w, h := r.screen.Size()
	return float64(w), float64(h)
}

func (r *TcellRenderer) RenderWorld(world *game.World, camera Camera) {
	if r.screen == nil {
		return
	}

	screenW, screenH := r.screen.Size()

	// Calculate camera offset
	cameraX := int(camera.X) - screenW/2
	cameraY := int(camera.Y) - screenH/2

	// Clamp camera
	if cameraX < 0 {
		cameraX = 0
	}
	if cameraY < 0 {
		cameraY = 0
	}
	if r.tileMap != nil && len(r.tileMap) > 0 && len(r.tileMap[0]) > 0 {
		maxCamX := len(r.tileMap[0]) - screenW
		maxCamY := len(r.tileMap) - screenH
		if cameraX > maxCamX && maxCamX >= 0 {
			cameraX = maxCamX
		}
		if cameraY > maxCamY && maxCamY >= 0 {
			cameraY = maxCamY
		}
	}

	// Render tiles
	if r.tileMap != nil {
		for y := 0; y < screenH && y+cameraY < len(r.tileMap); y++ {
			for x := 0; x < screenW && x+cameraX < len(r.tileMap[0]); x++ {
				tileY := y + cameraY
				tileX := x + cameraX
				if tileY >= 0 && tileY < len(r.tileMap) && tileX >= 0 && tileX < len(r.tileMap[0]) {
					ch := r.tileMap[tileY][tileX]
					if ch != ' ' {
						r.setCell(x, y, ch, ColorWhite, ColorBlack)
					}
				}
			}
		}
	}

	// Render entities using sprite atlas
	for _, e := range world.GetRenderables() {
		screenX := int(e.X) - cameraX
		screenY := int(e.Y) - cameraY

		if screenX >= 0 && screenX < screenW && screenY >= 0 && screenY < screenH {
			sprite := r.atlas.Get(e.SpriteID)
			r.setCell(screenX, screenY, sprite.Char, sprite.FG, sprite.BG)
		}
	}
}

func (r *TcellRenderer) RenderText(x, y float64, text string, color Color) {
	if r.screen == nil {
		return
	}
	ix, iy := int(x), int(y)
	for i, ch := range text {
		r.setCell(ix+i, iy, ch, color, ColorBlack)
	}
}

// RenderTileMap implements TileRenderer interface
func (r *TcellRenderer) RenderTileMap(tiles [][]rune, camera Camera) {
	r.tileMap = tiles
}

func (r *TcellRenderer) PollInput() (InputEvent, bool) {
	select {
	case ev := <-r.eventCh:
		return r.translateEvent(ev), true
	default:
		return InputEvent{Type: InputNone}, false
	}
}

func (r *TcellRenderer) translateEvent(ev tcell.Event) InputEvent {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		intent := protocol.IntentNone

		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			return InputEvent{Type: InputQuit, Quit: true}
		case tcell.KeyLeft:
			intent = protocol.IntentLeft
		case tcell.KeyRight:
			intent = protocol.IntentRight
		case tcell.KeyUp:
			intent = protocol.IntentJump
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q', 'Q':
				return InputEvent{Type: InputQuit, Quit: true}
			case 'a', 'A':
				intent = protocol.IntentLeft
			case 'd', 'D':
				intent = protocol.IntentRight
			case 'w', 'W', ' ':
				intent = protocol.IntentJump
			case 'j', 'J':
				intent = protocol.IntentAttack
			case 'k', 'K':
				intent = protocol.IntentUse
			}
		}

		if intent != protocol.IntentNone {
			return InputEvent{Type: InputKey, Intent: intent}
		}

	case *tcell.EventResize:
		if r.screen != nil {
			r.screen.Sync()
		}
		return InputEvent{Type: InputResize}
	}

	return InputEvent{Type: InputNone}
}

// setCell is a helper to set a cell with colors
func (r *TcellRenderer) setCell(x, y int, ch rune, fg, bg Color) {
	if r.screen == nil {
		return
	}
	fgColor := tcell.NewRGBColor(int32(fg.R), int32(fg.G), int32(fg.B))
	bgColor := tcell.NewRGBColor(int32(bg.R), int32(bg.G), int32(bg.B))
	style := tcell.StyleDefault.Foreground(fgColor).Background(bgColor)
	r.screen.SetContent(x, y, ch, nil, style)
}

// DrawHUD draws the heads-up display (convenience method for terminal)
func (r *TcellRenderer) DrawHUD(text string) {
	if r.screen == nil {
		return
	}
	_, h := r.screen.Size()
	r.RenderText(0, float64(h-1), text, ColorYellow)
}
