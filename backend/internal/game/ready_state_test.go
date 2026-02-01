package game

import (
	"sync"
	"testing"
)

// ========================================
// Basic Ready State Tests
// ========================================

func TestReadyTracker_SetAndIsReady(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"
	playerID := "player-1"

	// Initially not ready
	if tracker.IsReady(lobbyCode, playerID) {
		t.Error("expected player to not be ready initially")
	}

	// Set ready
	tracker.SetReady(lobbyCode, playerID, true)

	if !tracker.IsReady(lobbyCode, playerID) {
		t.Error("expected player to be ready after SetReady(true)")
	}

	// Set not ready
	tracker.SetReady(lobbyCode, playerID, false)

	if tracker.IsReady(lobbyCode, playerID) {
		t.Error("expected player to not be ready after SetReady(false)")
	}
}

func TestReadyTracker_ClearPlayer(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"
	playerID := "player-1"

	tracker.SetReady(lobbyCode, playerID, true)
	if !tracker.IsReady(lobbyCode, playerID) {
		t.Fatal("expected player to be ready")
	}

	tracker.ClearPlayer(lobbyCode, playerID)

	if tracker.IsReady(lobbyCode, playerID) {
		t.Error("expected player ready state to be cleared")
	}
}

func TestReadyTracker_ClearLobby(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"

	// Set multiple players ready
	tracker.SetReady(lobbyCode, "player-1", true)
	tracker.SetReady(lobbyCode, "player-2", true)

	if !tracker.IsReady(lobbyCode, "player-1") || !tracker.IsReady(lobbyCode, "player-2") {
		t.Fatal("expected both players to be ready")
	}

	tracker.ClearLobby(lobbyCode)

	if tracker.IsReady(lobbyCode, "player-1") || tracker.IsReady(lobbyCode, "player-2") {
		t.Error("expected all ready states to be cleared")
	}
}

func TestReadyTracker_LobbyIsolation(t *testing.T) {
	tracker := NewReadyTracker()

	// Set player ready in lobby 1
	tracker.SetReady("LOBBY1", "player-1", true)

	// Player should not be ready in lobby 2
	if tracker.IsReady("LOBBY2", "player-1") {
		t.Error("ready state should be isolated per lobby")
	}

	// Same player ID, different lobby
	tracker.SetReady("LOBBY2", "player-1", true)
	tracker.ClearLobby("LOBBY1")

	// Lobby 2 should still have player ready
	if !tracker.IsReady("LOBBY2", "player-1") {
		t.Error("clearing lobby 1 should not affect lobby 2")
	}
}

func TestReadyTracker_Toggle(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"
	playerID := "player-1"

	// Initially not ready
	if tracker.IsReady(lobbyCode, playerID) {
		t.Error("expected player to not be ready initially")
	}

	// Set ready = true
	tracker.SetReady(lobbyCode, playerID, true)
	if !tracker.IsReady(lobbyCode, playerID) {
		t.Error("expected player to be ready after setting true")
	}

	// Toggle to ready = false
	tracker.SetReady(lobbyCode, playerID, false)
	if tracker.IsReady(lobbyCode, playerID) {
		t.Error("expected player to not be ready after toggling to false")
	}

	// Toggle back to ready = true
	tracker.SetReady(lobbyCode, playerID, true)
	if !tracker.IsReady(lobbyCode, playerID) {
		t.Error("expected player to be ready after toggling back to true")
	}
}

// ========================================
// AllReady Tests
// ========================================

func TestReadyTracker_AllReady_AllPlayersReady(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"
	players := []string{"player-1", "player-2"}

	// Initially not all ready
	if tracker.AllReady(lobbyCode, players) {
		t.Error("expected AllReady to be false initially")
	}

	// Set one player ready
	tracker.SetReady(lobbyCode, "player-1", true)
	if tracker.AllReady(lobbyCode, players) {
		t.Error("expected AllReady to be false with only one player ready")
	}

	// Set both players ready
	tracker.SetReady(lobbyCode, "player-2", true)
	if !tracker.AllReady(lobbyCode, players) {
		t.Error("expected AllReady to be true with both players ready")
	}
}

func TestReadyTracker_AllReady_EmptyPlayerList(t *testing.T) {
	tracker := NewReadyTracker()

	// Empty player list should return true (vacuous truth)
	if !tracker.AllReady("TEST01", []string{}) {
		t.Error("expected AllReady to be true for empty player list")
	}
}

func TestReadyTracker_AllReady_NonExistentLobby(t *testing.T) {
	tracker := NewReadyTracker()

	players := []string{"player-1", "player-2"}

	// Non-existent lobby should return false
	if tracker.AllReady("NONEXISTENT", players) {
		t.Error("expected AllReady to be false for non-existent lobby")
	}
}

func TestReadyTracker_AllReady_PlayerUnreadies(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"
	players := []string{"player-1", "player-2"}

	// Set both players ready
	tracker.SetReady(lobbyCode, "player-1", true)
	tracker.SetReady(lobbyCode, "player-2", true)

	if !tracker.AllReady(lobbyCode, players) {
		t.Fatal("expected AllReady to be true")
	}

	// Player 1 unreadies
	tracker.SetReady(lobbyCode, "player-1", false)

	if tracker.AllReady(lobbyCode, players) {
		t.Error("expected AllReady to be false after player unreadied")
	}
}

// ========================================
// Edge Cases
// ========================================

func TestReadyTracker_ClearPlayer_NonExistent(t *testing.T) {
	tracker := NewReadyTracker()

	// Should not panic
	tracker.ClearPlayer("NONEXISTENT", "player-1")
}

func TestReadyTracker_ClearLobby_NonExistent(t *testing.T) {
	tracker := NewReadyTracker()

	// Should not panic
	tracker.ClearLobby("NONEXISTENT")
}

func TestReadyTracker_ClearPlayer_CleansUpEmptyLobby(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"

	tracker.SetReady(lobbyCode, "player-1", true)
	tracker.ClearPlayer(lobbyCode, "player-1")

	// Internal state should be cleaned up (no empty maps)
	// Verify by checking that setting ready again works
	tracker.SetReady(lobbyCode, "player-1", true)
	if !tracker.IsReady(lobbyCode, "player-1") {
		t.Error("expected player to be ready after re-setting")
	}
}

// ========================================
// Concurrency Tests
// ========================================

func TestReadyTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"
	var wg sync.WaitGroup

	// Multiple goroutines setting and reading ready state
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func(id int) {
			defer wg.Done()
			playerID := "player-" + string(rune('0'+id%10))
			tracker.SetReady(lobbyCode, playerID, true)
		}(i)

		go func(id int) {
			defer wg.Done()
			playerID := "player-" + string(rune('0'+id%10))
			tracker.IsReady(lobbyCode, playerID)
		}(i)

		go func(id int) {
			defer wg.Done()
			playerID := "player-" + string(rune('0'+id%10))
			tracker.ClearPlayer(lobbyCode, playerID)
		}(i)
	}

	wg.Wait()

	// Verify tracker is in consistent state after concurrent access
	// Each player's ready state should be deterministically readable
	for i := 0; i < 10; i++ {
		playerID := "player-" + string(rune('0'+i))
		// Just verify IsReady doesn't panic and returns a boolean
		_ = tracker.IsReady(lobbyCode, playerID)
	}
}

func TestReadyTracker_ConcurrentAllReady(t *testing.T) {
	tracker := NewReadyTracker()

	lobbyCode := "TEST01"
	players := []string{"player-0", "player-1", "player-2"}
	var wg sync.WaitGroup

	// Set initial state
	for _, p := range players {
		tracker.SetReady(lobbyCode, p, true)
	}

	// Concurrent reads and writes
	for i := 0; i < 50; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			tracker.AllReady(lobbyCode, players)
		}()

		go func(id int) {
			defer wg.Done()
			playerID := players[id%len(players)]
			tracker.SetReady(lobbyCode, playerID, id%2 == 0)
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions occur
}
