package client

import (
	"github.com/andersfylling/rayman-slides/internal/game"
)

// Reconciler handles comparing server state to client predictions
// and performing rollback + replay when mismatches occur
type Reconciler struct {
	predictions *PredictionBuffer
	tolerance   float64 // Position difference tolerance for matching
}

// NewReconciler creates a reconciler with the given prediction buffer
func NewReconciler(predictions *PredictionBuffer) *Reconciler {
	return &Reconciler{
		predictions: predictions,
		tolerance:   0.01, // Small tolerance for floating point comparison
	}
}

// SetTolerance sets the position mismatch tolerance
func (r *Reconciler) SetTolerance(tolerance float64) {
	r.tolerance = tolerance
}

// ReconcileResult contains information about a reconciliation attempt
type ReconcileResult struct {
	Reconciled     bool   // Whether reconciliation was performed
	RolledBack     bool   // Whether a rollback occurred
	ReplayedTicks  int    // Number of ticks replayed after rollback
	ServerTick     uint64 // The tick that was reconciled
	MismatchReason string // If rollback, why the mismatch occurred
}

// Reconcile compares the server's authoritative state to our predicted state
// and performs rollback + replay if there's a mismatch
//
// Flow:
// 1. Find our prediction for the server's tick
// 2. Compare states
// 3. If mismatch: rollback world to server state, replay all inputs since that tick
// 4. Prune old predictions
func (r *Reconciler) Reconcile(
	world *game.World,
	serverState *game.WorldState,
	currentTick uint64,
) ReconcileResult {
	result := ReconcileResult{
		ServerTick: serverState.Tick,
	}

	// Find our prediction for the same tick as the server state
	predicted := r.predictions.GetState(serverState.Tick)

	if predicted == nil {
		// No prediction to compare - this happens at start or after long gap
		// Just accept the server state and move on
		world.Restore(*serverState)
		r.predictions.PruneBefore(serverState.Tick)
		result.Reconciled = true
		return result
	}

	// Compare our prediction to server state
	if r.statesMatch(predicted, serverState) {
		// Prediction was correct! Just prune old data
		r.predictions.PruneBefore(serverState.Tick)
		result.Reconciled = true
		return result
	}

	// Mismatch detected - need to rollback and replay
	result.RolledBack = true
	result.MismatchReason = r.describeMismatch(predicted, serverState)

	// Step 1: Rollback to server state
	world.Restore(*serverState)

	// Step 2: Get all inputs we applied since the server tick
	inputs := r.predictions.GetInputsSince(serverState.Tick)
	result.ReplayedTicks = len(inputs)

	// Step 3: Replay each input
	for _, input := range inputs {
		world.SetPlayerIntent(1, input.Intents) // TODO: proper player ID
		world.Update()
	}

	// Step 4: Prune old predictions
	r.predictions.PruneBefore(serverState.Tick)

	result.Reconciled = true
	return result
}

// statesMatch compares a predicted WorldSnapshot to the server's WorldState
func (r *Reconciler) statesMatch(predicted *WorldSnapshot, server *game.WorldState) bool {
	// Quick checksum comparison if both have it
	if predicted.Checksum != 0 && server.Checksum != 0 {
		if predicted.Checksum == server.Checksum {
			return true
		}
	}

	// Detailed comparison
	if len(predicted.Entities) != len(server.Entities) {
		return false
	}

	for i := range predicted.Entities {
		pe := &predicted.Entities[i]
		se := &server.Entities[i]

		// Compare positions within tolerance
		dx := pe.PositionX - se.Position.X
		dy := pe.PositionY - se.Position.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}

		if dx > r.tolerance || dy > r.tolerance {
			return false
		}

		// Compare grounded state
		if pe.Grounded != se.Grounded.OnGround {
			return false
		}
	}

	return true
}

// describeMismatch returns a human-readable description of why states don't match
func (r *Reconciler) describeMismatch(predicted *WorldSnapshot, server *game.WorldState) string {
	if len(predicted.Entities) != len(server.Entities) {
		return "entity count mismatch"
	}

	for i := range predicted.Entities {
		pe := &predicted.Entities[i]
		se := &server.Entities[i]

		dx := pe.PositionX - se.Position.X
		dy := pe.PositionY - se.Position.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}

		if dx > r.tolerance || dy > r.tolerance {
			return "position mismatch"
		}

		if pe.Grounded != se.Grounded.OnGround {
			return "grounded state mismatch"
		}
	}

	return "checksum mismatch (detailed comparison passed)"
}

// ConvertToWorldSnapshot converts a game.WorldState to a client WorldSnapshot
// for storing in the prediction buffer
func ConvertToWorldSnapshot(state *game.WorldState) WorldSnapshot {
	ws := WorldSnapshot{
		Tick:     state.Tick,
		Checksum: state.Checksum,
		Entities: make([]EntitySnapshot, 0, len(state.Entities)),
	}

	for _, es := range state.Entities {
		ws.Entities = append(ws.Entities, EntitySnapshot{
			PositionX: es.Position.X,
			PositionY: es.Position.Y,
			VelocityX: es.Velocity.X,
			VelocityY: es.Velocity.Y,
			Grounded:  es.Grounded.OnGround,
		})
	}

	return ws
}
