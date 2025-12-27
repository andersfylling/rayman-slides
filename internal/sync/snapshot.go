// Package sync handles state synchronization between server and clients.
package sync

import (
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// SnapshotBuffer holds recent snapshots for interpolation
type SnapshotBuffer struct {
	snapshots []protocol.StateSnapshot
	capacity  int
}

// NewSnapshotBuffer creates a buffer with the given capacity
func NewSnapshotBuffer(capacity int) *SnapshotBuffer {
	return &SnapshotBuffer{
		snapshots: make([]protocol.StateSnapshot, 0, capacity),
		capacity:  capacity,
	}
}

// Add adds a snapshot to the buffer
func (b *SnapshotBuffer) Add(snap protocol.StateSnapshot) {
	if len(b.snapshots) >= b.capacity {
		// Remove oldest
		copy(b.snapshots, b.snapshots[1:])
		b.snapshots = b.snapshots[:len(b.snapshots)-1]
	}
	b.snapshots = append(b.snapshots, snap)
}

// Get returns the two snapshots to interpolate between
// Returns nil if not enough snapshots
func (b *SnapshotBuffer) Get() (*protocol.StateSnapshot, *protocol.StateSnapshot) {
	if len(b.snapshots) < 2 {
		return nil, nil
	}
	return &b.snapshots[0], &b.snapshots[1]
}

// Advance removes the oldest snapshot (after interpolation complete)
func (b *SnapshotBuffer) Advance() {
	if len(b.snapshots) > 0 {
		copy(b.snapshots, b.snapshots[1:])
		b.snapshots = b.snapshots[:len(b.snapshots)-1]
	}
}

// Latest returns the most recent snapshot
func (b *SnapshotBuffer) Latest() *protocol.StateSnapshot {
	if len(b.snapshots) == 0 {
		return nil
	}
	return &b.snapshots[len(b.snapshots)-1]
}

// Len returns the number of buffered snapshots
func (b *SnapshotBuffer) Len() int {
	return len(b.snapshots)
}
