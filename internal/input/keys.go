package input

import (
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// GameKey represents a logical game key (backend-agnostic)
type GameKey uint8

const (
	KeyLeft GameKey = iota
	KeyRight
	KeyJump
	KeyAttack
	KeyUse
	KeyQuit
	KeyCount // Sentinel for array sizing
)

// KeyEventType indicates press or release
type KeyEventType uint8

const (
	KeyDown KeyEventType = iota
	KeyUp
)

// KeyEvent represents a key state transition
type KeyEvent struct {
	Type KeyEventType
	Key  GameKey
}

// KeyState tracks pressed state of all keys using fixed-size arrays
type KeyState struct {
	pressed [KeyCount]bool
}

// NewKeyState creates a new key state tracker
func NewKeyState() *KeyState {
	return &KeyState{}
}

// IsPressed returns whether a key is currently pressed
func (s *KeyState) IsPressed(key GameKey) bool {
	if key >= KeyCount {
		return false
	}
	return s.pressed[key]
}

// SetPressed updates a key's pressed state
func (s *KeyState) SetPressed(key GameKey, pressed bool) {
	if key >= KeyCount {
		return
	}
	s.pressed[key] = pressed
}

// ToIntents converts key state to protocol.Intent bitmask
func (s *KeyState) ToIntents() protocol.Intent {
	var intents protocol.Intent
	if s.pressed[KeyLeft] {
		intents |= protocol.IntentLeft
	}
	if s.pressed[KeyRight] {
		intents |= protocol.IntentRight
	}
	if s.pressed[KeyJump] {
		intents |= protocol.IntentJump
	}
	if s.pressed[KeyAttack] {
		intents |= protocol.IntentAttack
	}
	if s.pressed[KeyUse] {
		intents |= protocol.IntentUse
	}
	return intents
}

// Clone returns a copy of the key state
func (s *KeyState) Clone() KeyState {
	return KeyState{
		pressed: s.pressed,
	}
}

// Reset clears all key states
func (s *KeyState) Reset() {
	for i := range s.pressed {
		s.pressed[i] = false
	}
}
