package game

import (
	"testing"

	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// TestAttackChargeRelease tests the charge-release attack mechanic.
// Press to charge, release to fire.
func TestAttackChargeRelease(t *testing.T) {
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

	// Press attack key - should start charging, NOT fire yet
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	if countFists() != 0 {
		t.Fatal("Should not fire while still holding attack key")
	}

	// Continue holding for a few ticks
	for i := 0; i < 10; i++ {
		world.SetPlayerIntent(1, protocol.IntentAttack)
		world.Update()
	}

	if countFists() != 0 {
		t.Fatal("Should still not fire while holding attack key")
	}

	// Release attack key - NOW it should fire
	world.SetPlayerIntent(1, protocol.IntentNone)
	world.Update()

	if countFists() != 1 {
		t.Fatalf("Expected 1 fist after releasing attack, got %d", countFists())
	}
}

// TestAttackQuickTap tests that a quick press-release still fires.
func TestAttackQuickTap(t *testing.T) {
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

	// Quick tap: press and release in consecutive frames
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update() // Press

	world.SetPlayerIntent(1, protocol.IntentNone)
	world.Update() // Release

	if countFists() != 1 {
		t.Fatalf("Expected 1 fist after quick tap, got %d", countFists())
	}
}

// TestAttackChargeDistance tests that longer charge = greater distance.
func TestAttackChargeDistance(t *testing.T) {
	world := NewWorld()
	world.SpawnPlayer(1, "Test", 10, 10)

	getFistDistance := func() float64 {
		query := world.fistFilter.Query()
		defer query.Close()
		for query.Next() {
			_, _, fist := query.Get()
			return fist.MaxDistance
		}
		return 0
	}

	// Quick tap - should get minimum distance
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()
	world.SetPlayerIntent(1, protocol.IntentNone)
	world.Update()

	quickTapDistance := getFistDistance()
	if quickTapDistance < MinFistDistance || quickTapDistance > MinFistDistance+1 {
		t.Fatalf("Quick tap should give ~MinFistDistance, got %.2f", quickTapDistance)
	}

	// Wait for cooldown
	for i := 0; i < AttackCooldown+5; i++ {
		world.SetPlayerIntent(1, protocol.IntentNone)
		world.Update()
	}

	// Clear the first fist by running enough updates for it to despawn
	for i := 0; i < 100; i++ {
		world.SetPlayerIntent(1, protocol.IntentNone)
		world.Update()
	}

	// Long charge - should get more distance
	world.SetPlayerIntent(1, protocol.IntentAttack)
	for i := 0; i < 60; i++ { // Hold for 1 second (60 ticks)
		world.Update()
	}
	world.SetPlayerIntent(1, protocol.IntentNone)
	world.Update()

	chargedDistance := getFistDistance()
	if chargedDistance <= quickTapDistance {
		t.Fatalf("Charged attack should travel further than quick tap: %.2f vs %.2f",
			chargedDistance, quickTapDistance)
	}
}

// TestAttackCooldown verifies that attacks have a cooldown period.
func TestAttackCooldown(t *testing.T) {
	world := NewWorld()
	world.SpawnPlayer(1, "Test", 10, 10)

	// First attack: press and release
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()
	world.SetPlayerIntent(1, protocol.IntentNone)
	world.Update()

	// Check that we're in cooldown (Attacking = true)
	playerQuery := world.attackFilter.Query()
	var attackState *AttackState
	for playerQuery.Next() {
		_, _, _, attack, _, _ := playerQuery.Get()
		attackState = attack
	}

	if attackState == nil {
		t.Fatal("Could not find player attack state")
	}

	if !attackState.Attacking {
		t.Fatal("Should be in cooldown after attack")
	}

	// Try to start another attack during cooldown - should not work
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	playerQuery = world.attackFilter.Query()
	for playerQuery.Next() {
		_, _, _, attack, _, _ := playerQuery.Get()
		attackState = attack
	}

	if attackState.Charging {
		t.Fatal("Should not be able to charge during cooldown")
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

	// Now should be able to charge again
	world.SetPlayerIntent(1, protocol.IntentAttack)
	world.Update()

	playerQuery = world.attackFilter.Query()
	for playerQuery.Next() {
		_, _, _, attack, _, _ := playerQuery.Get()
		attackState = attack
	}

	if !attackState.Charging {
		t.Fatal("Should be able to charge after cooldown expires")
	}
}

// TestAttackNoFireWhileHolding verifies that holding attack doesn't fire multiple times.
func TestAttackNoFireWhileHolding(t *testing.T) {
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

	// Hold attack for many ticks without releasing
	for i := 0; i < 100; i++ {
		world.SetPlayerIntent(1, protocol.IntentAttack)
		world.Update()
	}

	// Should not have fired anything yet
	if countFists() != 0 {
		t.Fatalf("Should not fire while holding, got %d fists", countFists())
	}

	// Release - now it should fire exactly once
	world.SetPlayerIntent(1, protocol.IntentNone)
	world.Update()

	if countFists() != 1 {
		t.Fatalf("Expected exactly 1 fist after release, got %d", countFists())
	}
}
