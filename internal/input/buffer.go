package input

import (
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// Buffer collects input frames for tick-aligned sending
type Buffer struct {
	frames []protocol.InputFrame
	tick   uint64
}

// NewBuffer creates an input buffer
func NewBuffer() *Buffer {
	return &Buffer{
		frames: make([]protocol.InputFrame, 0, 16),
	}
}

// Add records an input frame for the current tick
func (b *Buffer) Add(intents protocol.Intent) {
	b.frames = append(b.frames, protocol.InputFrame{
		Tick:    b.tick,
		Intents: intents,
	})
}

// Tick advances to the next tick
func (b *Buffer) Tick() {
	b.tick++
}

// Flush returns all buffered frames and clears the buffer
func (b *Buffer) Flush() []protocol.InputFrame {
	frames := b.frames
	b.frames = make([]protocol.InputFrame, 0, 16)
	return frames
}

// CurrentTick returns the current tick number
func (b *Buffer) CurrentTick() uint64 {
	return b.tick
}
