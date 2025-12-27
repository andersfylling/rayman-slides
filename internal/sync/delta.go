package sync

import (
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// Baseline tracks the last acknowledged state per client
type Baseline struct {
	tick     uint64
	entities map[protocol.EntityID][]byte // Last sent component data
}

// NewBaseline creates a new baseline tracker
func NewBaseline() *Baseline {
	return &Baseline{
		entities: make(map[protocol.EntityID][]byte),
	}
}

// Update sets the baseline to the given snapshot
func (b *Baseline) Update(snap *protocol.StateSnapshot) {
	b.tick = snap.Tick
	for _, e := range snap.Entities {
		b.entities[e.ID] = e.Components
	}
	for _, id := range snap.Removed {
		delete(b.entities, id)
	}
}

// Tick returns the baseline tick
func (b *Baseline) Tick() uint64 {
	return b.tick
}

// Diff computes the delta between baseline and current state
func Diff(baseline *Baseline, current []protocol.EntityState) protocol.StateSnapshot {
	snap := protocol.StateSnapshot{
		Full:     false,
		Baseline: baseline.tick,
		Entities: make([]protocol.EntityState, 0),
		Removed:  make([]protocol.EntityID, 0),
	}

	currentIDs := make(map[protocol.EntityID]bool)

	for _, e := range current {
		currentIDs[e.ID] = true

		// Check if changed from baseline
		if old, exists := baseline.entities[e.ID]; exists {
			if !bytesEqual(old, e.Components) {
				snap.Entities = append(snap.Entities, e)
			}
		} else {
			// New entity
			snap.Entities = append(snap.Entities, e)
		}
	}

	// Find removed entities
	for id := range baseline.entities {
		if !currentIDs[id] {
			snap.Removed = append(snap.Removed, id)
		}
	}

	return snap
}

// Apply applies a delta snapshot to a state
func Apply(state map[protocol.EntityID]protocol.EntityState, snap *protocol.StateSnapshot) {
	if snap.Full {
		// Full snapshot - replace everything
		for k := range state {
			delete(state, k)
		}
	}

	for _, e := range snap.Entities {
		state[e.ID] = e
	}

	for _, id := range snap.Removed {
		delete(state, id)
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
