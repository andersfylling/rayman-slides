// Package input handles keyboard capture and intent mapping.
package input

import (
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// Handler captures terminal input and converts to intents
type Handler struct {
	mapping  map[rune]protocol.Intent
	state    protocol.Intent // Currently held intents
	holdTime map[protocol.Intent]int64
}

// NewHandler creates an input handler with default key bindings
func NewHandler() *Handler {
	h := &Handler{
		mapping:  make(map[rune]protocol.Intent),
		holdTime: make(map[protocol.Intent]int64),
	}
	h.SetDefaultBindings()
	return h
}

// SetDefaultBindings configures WASD + arrow keys
func (h *Handler) SetDefaultBindings() {
	// Arrow keys (these are multi-byte, simplified here)
	// In practice, use tcell/bubbletea key constants

	// WASD
	h.mapping['a'] = protocol.IntentLeft
	h.mapping['A'] = protocol.IntentLeft
	h.mapping['d'] = protocol.IntentRight
	h.mapping['D'] = protocol.IntentRight
	h.mapping['w'] = protocol.IntentJump
	h.mapping['W'] = protocol.IntentJump
	h.mapping[' '] = protocol.IntentJump // Space

	// Attack and use
	h.mapping['j'] = protocol.IntentAttack
	h.mapping['J'] = protocol.IntentAttack
	h.mapping['k'] = protocol.IntentUse
	h.mapping['K'] = protocol.IntentUse
}

// Bind sets a custom key binding
func (h *Handler) Bind(key rune, intent protocol.Intent) {
	h.mapping[key] = intent
}

// OnKeyPress handles a key press event
func (h *Handler) OnKeyPress(key rune) {
	if intent, ok := h.mapping[key]; ok {
		h.state |= intent
		// TODO: Record timestamp for hold detection
	}
}

// OnKeyRelease handles a key release (if terminal supports it)
func (h *Handler) OnKeyRelease(key rune) {
	if intent, ok := h.mapping[key]; ok {
		h.state &^= intent
	}
}

// State returns current intent state
func (h *Handler) State() protocol.Intent {
	return h.state
}

// Clear resets the intent state
func (h *Handler) Clear() {
	h.state = protocol.IntentNone
}
