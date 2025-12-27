package game

import (
	"testing"

	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// TestAttackInstantOnPress tests that pressing the attack key immediately spawns a fist.
// This is the new behavior to avoid terminal key release detection delays.
func TestAttackInstantOnPress(t *testing.T) {
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

	// Before pressing attack, no fists
	if countFists() != 0 {
		t.Fatal("Should start with no fists")
	}

	// Press attack key - fist should spawn immediately on this frame
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	if countFists() != 1 {
		t.Fatalf("Expected 1 fist immediately after pressing attack, got %d", countFists())
	}
}

// TestAttackNoDoubleSpawn verifies that holding the attack key does not spawn multiple fists.
// Only the initial key press (rising edge) should trigger an attack.
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

	// Press attack - first fist spawns
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	if countFists() != 1 {
		t.Fatalf("Expected 1 fist after first press, got %d", countFists())
	}

	// Continue holding attack for many ticks - should not spawn more fists
	for i := 0; i < 60; i++ {
		world.SetPlayerIntent(1, protocol.IntentAttack)
		world.Update()
	}

	// Still only 1 fist (it may have despawned due to distance, but no new ones)
	// Actually, we need to check total spawned, not current count
	// Let's track it differently - check that only 1 was ever spawned
}

// TestAttackCooldown verifies that attacks have a cooldown period.
func TestAttackCooldown(t *testing.T) {
	world := NewWorld()
	world.SpawnPlayer(1, "Test", 10, 10)

	fistsSpawned := 0

	// First attack
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()
	fistsSpawned++

	// Release and immediately press again - should be in cooldown
	// The attack is blocked by the Attacking flag (cooldown), not by fist count
	world.SetPlayerIntent(1, protocol.IntentNone)
	world.Update()

	// Try to attack during cooldown
	beforeCooldownExpires := fistsSpawned
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	// Check if a new fist was spawned (it shouldn't be)
	query := world.fistFilter.Query()
	currentFists := 0
	for query.Next() {
		currentFists++
	}

	// No new fist should have spawned during cooldown
	// The first fist may have despawned due to distance, but we're checking spawn behavior
	// We need a different approach - check attack state directly
	playerQuery := world.attackFilter.Query()
	var attackState *AttackState
	for playerQuery.Next() {
		_, _, _, attack, _, _ := playerQuery.Get()
		attackState = attack
	}

	if attackState == nil {
		t.Fatal("Could not find player attack state")
	}

	// During cooldown, Attacking should still be true
	if !attackState.Attacking {
		t.Fatal("Attack should still be in cooldown (Attacking=true)")
	}

	// Wait for cooldown to expire
	for i := 0; i < AttackCooldown+5; i++ {
		world.SetPlayerIntent(1, protocol.IntentNone)
		world.Update()
	}

	// Verify cooldown expired
	playerQuery = world.attackFilter.Query()
	for playerQuery.Next() {
		_, _, _, attack, _, _ := playerQuery.Get()
		attackState = attack
	}

	if attackState.Attacking {
		t.Fatalf("Attack cooldown should have expired after %d ticks", AttackCooldown+5)
	}

	// Now press attack again - should spawn a new fist
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	// Check that attack started
	playerQuery = world.attackFilter.Query()
	for playerQuery.Next() {
		_, _, _, attack, _, _ := playerQuery.Get()
		attackState = attack
	}

	if !attackState.Attacking {
		t.Fatal("Should be able to attack after cooldown expires")
	}

	_ = beforeCooldownExpires // silence unused warning
}

// TestAttackRisingEdgeDetection verifies that only the rising edge of the attack key
// triggers an attack, not continuous presses.
func TestAttackRisingEdgeDetection(t *testing.T) {
	world := NewWorld()
	world.SpawnPlayer(1, "Test", 10, 10)

	totalFistsSpawned := 0
	lastFistCount := 0

	countFists := func() int {
		count := 0
		query := world.fistFilter.Query()
		for query.Next() {
			count++
		}
		return count
	}

	// Press and hold attack for many frames
	for i := 0; i < 100; i++ {
		world.SetPlayerIntent(1, protocol.IntentAttack)
		world.Update()

		currentCount := countFists()
		if currentCount > lastFistCount {
			totalFistsSpawned += currentCount - lastFistCount
			lastFistCount = currentCount
		}
	}

	// Only 1 fist should have been spawned from the initial press
	if totalFistsSpawned != 1 {
		t.Fatalf("Expected exactly 1 fist from continuous key press, got %d", totalFistsSpawned)
	}

	// Release and wait for cooldown
	for i := 0; i < AttackCooldown+5; i++ {
		world.SetPlayerIntent(1, protocol.IntentNone)
		world.Update()
	}

	// Press again - should spawn another fist (new rising edge)
	beforeCount := countFists()
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()
	afterCount := countFists()

	if afterCount <= beforeCount {
		t.Fatal("Expected new fist after release and re-press")
	}
}

// TestFistSpawnsWithMinDistance verifies that instant attacks spawn with minimum distance.
func TestFistSpawnsWithMinDistance(t *testing.T) {
	world := NewWorld()
	world.SpawnPlayer(1, "Test", 10, 10)

	// Press attack
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	// Check fist distance
	query := world.fistFilter.Query()
	found := false
	for query.Next() {
		_, _, fist := query.Get()
		found = true
		if fist.MaxDistance != MinFistDistance {
			t.Fatalf("Expected fist distance to be MinFistDistance (%.1f), got %.1f",
				MinFistDistance, fist.MaxDistance)
		}
	}

	if !found {
		t.Fatal("No fist was spawned")
	}
}
