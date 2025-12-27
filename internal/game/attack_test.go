package game

import (
	"testing"

	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// TestAttackChargingNoFalseRelease tests that holding the attack key does not
// cause a false fist launch due to terminal key repeat timing gaps.
// This simulates the problematic scenario where the terminal's initial key
// repeat delay (~500ms) would cause a gap that previously triggered a fist launch.
func TestAttackChargingNoFalseRelease(t *testing.T) {
	world := NewWorld()

	// Spawn a player
	world.SpawnPlayer(1, "Test", 10, 10)

	// Helper to count fists in the world
	countFists := func() int {
		count := 0
		query := world.fistFilter.Query()
		for query.Next() {
			count++
		}
		return count
	}

	// Phase 1: Press attack key (simulating initial key press)
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update() // Tick 1 - start charging

	if countFists() != 0 {
		t.Fatal("Fist should not be spawned immediately when attack key is pressed")
	}

	// Phase 2: Simulate terminal initial key repeat delay by keeping attack pressed
	// In real scenario, the intentHoldDuration (600ms) keeps the intent active
	// even during the gap before terminal key repeat starts
	for i := 0; i < 30; i++ { // ~500ms worth of ticks at 60TPS
		world.SetPlayerIntent(1, protocol.IntentAttack)
		world.Update()
	}

	if countFists() != 0 {
		t.Fatal("Fist should not be spawned while attack key is held down")
	}

	// Phase 3: Release attack key - now fist should spawn after debounce
	for i := 0; i < ReleaseDebounceThreshold; i++ {
		world.SetPlayerIntent(1, protocol.IntentNone)
		world.Update()
	}

	if countFists() != 1 {
		t.Fatalf("Expected exactly 1 fist after release, got %d", countFists())
	}
}

// TestAttackNoDoubleSpawn verifies that charging and releasing only spawns one fist,
// not two (which was the original bug).
func TestAttackNoDoubleSpawn(t *testing.T) {
	world := NewWorld()

	world.SpawnPlayer(1, "Test", 10, 10)

	countFists := func() int {
		count := 0
		query := world.fistFilter.Query()
		for query.Next() {
			count++
		}
		return count
	}

	// Press attack
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	// Hold for 1 second (60 ticks)
	for i := 0; i < 60; i++ {
		world.SetPlayerIntent(1, protocol.IntentAttack)
		world.Update()
	}

	// Release
	for i := 0; i < ReleaseDebounceThreshold+5; i++ {
		world.SetPlayerIntent(1, protocol.IntentNone)
		world.Update()
	}

	if countFists() != 1 {
		t.Fatalf("Expected exactly 1 fist after charge and release, got %d", countFists())
	}
}

// TestFistDistanceBasedOnCharge verifies that charge duration affects fist distance.
func TestFistDistanceBasedOnCharge(t *testing.T) {
	testCases := []struct {
		name          string
		chargeTicks   int
		minExpected   float64
		maxExpected   float64
	}{
		// Minimal charge - need at least a few ticks for stable charging state
		{"minimal charge", 10, MinFistDistance, MinFistDistance + 2},
		{"half charge", MaxChargeTicks / 2, (MinFistDistance + MaxFistDistance) / 2 - 2, (MinFistDistance + MaxFistDistance) / 2 + 2},
		{"full charge", MaxChargeTicks, MaxFistDistance - 1, MaxFistDistance + 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			world := NewWorld()
			world.SpawnPlayer(1, "Test", 10, 10)

			// Hold attack for specified ticks (charging)
			for i := 0; i < tc.chargeTicks; i++ {
				world.SetPlayerIntent(1, protocol.IntentAttack)
				world.Update()
			}

			// Release - need exactly enough frames to pass debounce threshold
			// Don't run extra updates or short-distance fists will be removed
			var spawnedDistance float64
			for i := 0; i < ReleaseDebounceThreshold+1; i++ {
				world.SetPlayerIntent(1, protocol.IntentNone)
				world.Update()

				// Check if fist was just spawned (on the debounce threshold frame)
				query := world.fistFilter.Query()
				for query.Next() {
					_, _, fist := query.Get()
					spawnedDistance = fist.MaxDistance
				}
				if spawnedDistance > 0 {
					break
				}
			}

			if spawnedDistance == 0 {
				t.Fatalf("No fist spawned after %d charge ticks", tc.chargeTicks)
			}

			if spawnedDistance < tc.minExpected || spawnedDistance > tc.maxExpected {
				t.Fatalf("Expected fist distance in range [%.1f, %.1f], got %.1f",
					tc.minExpected, tc.maxExpected, spawnedDistance)
			}
		})
	}
}
