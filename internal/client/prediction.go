package client

import (
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// EntitySnapshot captures the state of an entity at a point in time
type EntitySnapshot struct {
	ID        protocol.EntityID
	PositionX float64
	PositionY float64
	VelocityX float64
	VelocityY float64
	Grounded  bool
}

// WorldSnapshot captures the state of the world at a specific tick
type WorldSnapshot struct {
	Tick     uint64
	Entities []EntitySnapshot
	Checksum uint32 // Fast comparison hash
}

// PredictionBuffer stores recent inputs and predicted states for reconciliation
type PredictionBuffer struct {
	inputs   []protocol.InputFrame
	states   []WorldSnapshot
	capacity int
}

// NewPredictionBuffer creates a prediction buffer with the given capacity
func NewPredictionBuffer(capacity int) *PredictionBuffer {
	return &PredictionBuffer{
		inputs:   make([]protocol.InputFrame, 0, capacity),
		states:   make([]WorldSnapshot, 0, capacity),
		capacity: capacity,
	}
}

// RecordInput stores an input frame for potential replay
func (b *PredictionBuffer) RecordInput(frame protocol.InputFrame) {
	// Remove oldest if at capacity
	if len(b.inputs) >= b.capacity {
		b.inputs = b.inputs[1:]
	}
	b.inputs = append(b.inputs, frame)
}

// RecordState stores a predicted world state
func (b *PredictionBuffer) RecordState(state WorldSnapshot) {
	// Remove oldest if at capacity
	if len(b.states) >= b.capacity {
		b.states = b.states[1:]
	}
	b.states = append(b.states, state)
}

// GetState returns the predicted state for a specific tick, or nil if not found
func (b *PredictionBuffer) GetState(tick uint64) *WorldSnapshot {
	for i := len(b.states) - 1; i >= 0; i-- {
		if b.states[i].Tick == tick {
			return &b.states[i]
		}
	}
	return nil
}

// GetInputsSince returns all inputs after the given tick, sorted by tick
func (b *PredictionBuffer) GetInputsSince(tick uint64) []protocol.InputFrame {
	var result []protocol.InputFrame
	for _, input := range b.inputs {
		if input.Tick > tick {
			result = append(result, input)
		}
	}
	return result
}

// GetInputsInRange returns all inputs in the tick range [startTick, endTick]
func (b *PredictionBuffer) GetInputsInRange(startTick, endTick uint64) []protocol.InputFrame {
	var result []protocol.InputFrame
	for _, input := range b.inputs {
		if input.Tick >= startTick && input.Tick <= endTick {
			result = append(result, input)
		}
	}
	return result
}

// PruneBefore removes all entries older than the given tick
func (b *PredictionBuffer) PruneBefore(tick uint64) {
	// Prune inputs
	i := 0
	for i < len(b.inputs) && b.inputs[i].Tick < tick {
		i++
	}
	if i > 0 {
		b.inputs = b.inputs[i:]
	}

	// Prune states
	j := 0
	for j < len(b.states) && b.states[j].Tick < tick {
		j++
	}
	if j > 0 {
		b.states = b.states[j:]
	}
}

// LatestTick returns the tick of the most recent recorded state, or 0 if empty
func (b *PredictionBuffer) LatestTick() uint64 {
	if len(b.states) == 0 {
		return 0
	}
	return b.states[len(b.states)-1].Tick
}

// InputCount returns the number of stored inputs
func (b *PredictionBuffer) InputCount() int {
	return len(b.inputs)
}

// StateCount returns the number of stored states
func (b *PredictionBuffer) StateCount() int {
	return len(b.states)
}

// Clear removes all stored inputs and states
func (b *PredictionBuffer) Clear() {
	b.inputs = b.inputs[:0]
	b.states = b.states[:0]
}
