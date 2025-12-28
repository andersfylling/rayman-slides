//go:build gio

package input

import (
	"gioui.org/io/key"
)

// GioInput implements System for Gio.
// Events are pushed from the Gio event loop.
type GioInput struct {
	prevState [KeyCount]bool
	currState [KeyCount]bool
	quitFlag  bool
	events    []KeyEvent
}

// NewGioInput creates a new Gio input system.
func NewGioInput() *GioInput {
	return &GioInput{}
}

// HandleKeyEvent processes a Gio key event.
// Call this from the Gio event loop.
func (g *GioInput) HandleKeyEvent(e key.Event) {
	gk := gioKeyToGameKey(e.Name)
	if gk >= KeyCount {
		return
	}

	if e.State == key.Press {
		if !g.currState[gk] {
			g.currState[gk] = true
			g.events = append(g.events, KeyEvent{Type: KeyDown, Key: gk})
			if gk == KeyQuit {
				g.quitFlag = true
			}
		}
	} else if e.State == key.Release {
		if g.currState[gk] {
			g.currState[gk] = false
			g.events = append(g.events, KeyEvent{Type: KeyUp, Key: gk})
		}
	}
}

// Poll returns pending key events.
func (g *GioInput) Poll() []KeyEvent {
	events := g.events
	g.events = nil
	return events
}

// ShouldQuit returns true if quit was requested.
func (g *GioInput) ShouldQuit() bool {
	return g.quitFlag
}

// gioKeyToGameKey maps Gio key names to GameKey.
func gioKeyToGameKey(name key.Name) GameKey {
	switch name {
	case key.NameLeftArrow, "A":
		return KeyLeft
	case key.NameRightArrow, "D":
		return KeyRight
	case key.NameUpArrow, "W", key.NameSpace:
		return KeyJump
	case "J":
		return KeyAttack
	case "K":
		return KeyUse
	case key.NameEscape, "Q":
		return KeyQuit
	default:
		return KeyCount // Invalid
	}
}
