package game

import (
	"sync"
	"testing"
)

// ========================================
// Happy Path Tests
// ========================================

func TestNewLobby_CreatesValidLobby(t *testing.T) {
	code := "ABC123"
	hostID := "host-1"
	hostUsername := "HostPlayer"

	lobby := NewLobby(code, hostID, hostUsername)

	if lobby.Code != code {
		t.Errorf("expected code %q, got %q", code, lobby.Code)
	}
	if lobby.GetState() != LobbyStateWaiting {
		t.Errorf("expected state Waiting, got %v", lobby.GetState())
	}
	if lobby.PlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", lobby.PlayerCount())
	}
	if !lobby.IsHost(hostID) {
		t.Error("expected host to be identified as host")
	}
	if !lobby.HasPlayer(hostID) {
		t.Error("expected host to be in lobby")
	}
	if lobby.MaxPlayers != 2 {
		t.Errorf("expected max players 2, got %d", lobby.MaxPlayers)
	}
}

func TestAddPlayer_Success(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	err := lobby.AddPlayer("player-2", "Player2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lobby.PlayerCount() != 2 {
		t.Errorf("expected 2 players, got %d", lobby.PlayerCount())
	}
	if !lobby.HasPlayer("player-2") {
		t.Error("expected player-2 to be in lobby")
	}
	if lobby.GetState() != LobbyStateReady {
		t.Errorf("expected state Ready after 2nd player, got %v", lobby.GetState())
	}
}

func TestRemovePlayer_Success(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	err := lobby.RemovePlayer("player-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lobby.PlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", lobby.PlayerCount())
	}
	if lobby.HasPlayer("player-2") {
		t.Error("expected player-2 to be removed")
	}
	if lobby.GetState() != LobbyStateWaiting {
		t.Errorf("expected state Waiting after player left, got %v", lobby.GetState())
	}
}

func TestRemovePlayer_ReassignsHost(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	err := lobby.RemovePlayer("host-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lobby.IsHost("host-1") {
		t.Error("expected host-1 to no longer be host")
	}
	if !lobby.IsHost("player-2") {
		t.Error("expected player-2 to become new host")
	}
	if lobby.GetHostID() != "player-2" {
		t.Errorf("expected host ID to be player-2, got %q", lobby.GetHostID())
	}
}

func TestStart_Success(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	err := lobby.Start()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lobby.GetState() != LobbyStateActive {
		t.Errorf("expected state Active, got %v", lobby.GetState())
	}
}

func TestCanStart_ReturnsTrueWhenReady(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	if lobby.CanStart() {
		t.Error("expected CanStart to be false with 1 player")
	}

	lobby.AddPlayer("player-2", "Player2")

	if !lobby.CanStart() {
		t.Error("expected CanStart to be true with 2 players in Ready state")
	}
}

func TestHasPlayer_ReturnsCorrectly(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	if !lobby.HasPlayer("host-1") {
		t.Error("expected HasPlayer to return true for host")
	}
	if lobby.HasPlayer("nonexistent") {
		t.Error("expected HasPlayer to return false for non-existent player")
	}

	lobby.AddPlayer("player-2", "Player2")
	if !lobby.HasPlayer("player-2") {
		t.Error("expected HasPlayer to return true for joined player")
	}
}

func TestIsHost_ReturnsCorrectly(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	if !lobby.IsHost("host-1") {
		t.Error("expected IsHost to return true for host")
	}
	if lobby.IsHost("player-2") {
		t.Error("expected IsHost to return false for non-host player")
	}
	if lobby.IsHost("nonexistent") {
		t.Error("expected IsHost to return false for non-existent player")
	}
}

func TestGetPlayers_ReturnsSnapshot(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	players := lobby.GetPlayers()
	if len(players) != 2 {
		t.Errorf("expected 2 players, got %d", len(players))
	}

	// Verify it's a copy (modifying returned slice shouldn't affect lobby)
	players[0].Username = "Modified"
	originalPlayers := lobby.GetPlayers()
	if originalPlayers[0].Username == "Modified" {
		t.Error("GetPlayers should return a copy, not the original slice")
	}
}

func TestLobbyState_String(t *testing.T) {
	tests := []struct {
		state    LobbyState
		expected string
	}{
		{LobbyStateWaiting, "waiting"},
		{LobbyStateReady, "ready"},
		{LobbyStateActive, "active"},
		{LobbyState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("LobbyState(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

// ========================================
// Validation Error Tests
// ========================================

func TestAddPlayer_LobbyFull(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	// When lobby has 2 players, state transitions to Ready.
	// The state check happens before the "full" check in AddPlayer,
	// so we get ErrInvalidStateForJoin instead of ErrLobbyFull.
	err := lobby.AddPlayer("player-3", "Player3")
	if err != ErrInvalidStateForJoin {
		t.Errorf("expected ErrInvalidStateForJoin (state=Ready prevents join), got %v", err)
	}
}

func TestAddPlayer_LobbyFullInWaitingState(t *testing.T) {
	// Test the actual "lobby full" error by manually setting MaxPlayers=1
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.mu.Lock()
	lobby.MaxPlayers = 1 // Force max to 1 for this test
	lobby.mu.Unlock()

	err := lobby.AddPlayer("player-2", "Player2")
	if err != ErrLobbyFull {
		t.Errorf("expected ErrLobbyFull, got %v", err)
	}
}

func TestAddPlayer_DuplicatePlayer(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	err := lobby.AddPlayer("host-1", "DuplicateHost")
	if err != ErrPlayerAlreadyJoined {
		t.Errorf("expected ErrPlayerAlreadyJoined, got %v", err)
	}
}

func TestAddPlayer_InvalidState(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")
	lobby.Start() // Now in Active state

	err := lobby.AddPlayer("player-3", "Player3")
	if err != ErrInvalidStateForJoin {
		t.Errorf("expected ErrInvalidStateForJoin, got %v", err)
	}
}

func TestRemovePlayer_NotFound(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	err := lobby.RemovePlayer("nonexistent")
	if err != ErrPlayerNotFound {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestStart_InvalidState(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	err := lobby.Start()
	if err != ErrInvalidStateForStart {
		t.Errorf("expected ErrInvalidStateForStart, got %v", err)
	}
}

func TestStart_NotEnoughPlayers(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	// Manually set state to Ready without 2 players (edge case testing)
	lobby.mu.Lock()
	lobby.State = LobbyStateReady
	lobby.mu.Unlock()

	err := lobby.Start()
	if err != ErrNotEnoughPlayers {
		t.Errorf("expected ErrNotEnoughPlayers, got %v", err)
	}
}

// ========================================
// State Transition Tests
// ========================================

func TestStateTransition_WaitingToReady(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	if lobby.GetState() != LobbyStateWaiting {
		t.Errorf("expected initial state Waiting, got %v", lobby.GetState())
	}

	lobby.AddPlayer("player-2", "Player2")

	if lobby.GetState() != LobbyStateReady {
		t.Errorf("expected state Ready after 2nd player, got %v", lobby.GetState())
	}
}

func TestStateTransition_ReadyToWaiting(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	if lobby.GetState() != LobbyStateReady {
		t.Fatalf("expected state Ready, got %v", lobby.GetState())
	}

	lobby.RemovePlayer("player-2")

	if lobby.GetState() != LobbyStateWaiting {
		t.Errorf("expected state Waiting after player left, got %v", lobby.GetState())
	}
}

func TestStateTransition_ReadyToActive(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	if lobby.GetState() != LobbyStateReady {
		t.Fatalf("expected state Ready, got %v", lobby.GetState())
	}

	lobby.Start()

	if lobby.GetState() != LobbyStateActive {
		t.Errorf("expected state Active, got %v", lobby.GetState())
	}
}

func TestStateTransition_NoTransitionOnFirstAdd(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	// Initial state should remain Waiting (host is already added in NewLobby)
	if lobby.GetState() != LobbyStateWaiting {
		t.Errorf("expected state Waiting with 1 player, got %v", lobby.GetState())
	}
	if lobby.PlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", lobby.PlayerCount())
	}
}

func TestStateTransition_PreservedOnGet(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	// Multiple GetState calls shouldn't change state
	for i := 0; i < 10; i++ {
		if lobby.GetState() != LobbyStateReady {
			t.Errorf("expected state Ready to be preserved, got %v", lobby.GetState())
		}
	}
}

// ========================================
// Edge Case Tests
// ========================================

func TestAddPlayer_EmptyStrings(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	// Empty ID but has username - currently allowed by implementation
	err := lobby.AddPlayer("", "EmptyID")
	if err != nil {
		t.Logf("Adding player with empty ID returned: %v", err)
	}
}

func TestRemovePlayer_HostOfTwo(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	err := lobby.RemovePlayer("host-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Player2 should be new host
	if !lobby.IsHost("player-2") {
		t.Error("expected player-2 to be new host")
	}
	if lobby.PlayerCount() != 1 {
		t.Errorf("expected 1 player remaining, got %d", lobby.PlayerCount())
	}
	// State should be Waiting since we dropped below max
	if lobby.GetState() != LobbyStateWaiting {
		t.Errorf("expected state Waiting, got %v", lobby.GetState())
	}
}

func TestRemovePlayer_OnlyPlayer(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	err := lobby.RemovePlayer("host-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lobby.PlayerCount() != 0 {
		t.Errorf("expected 0 players, got %d", lobby.PlayerCount())
	}
	// Note: The service layer is responsible for deleting empty lobbies
}

func TestStart_CalledTwice(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	err := lobby.Start()
	if err != nil {
		t.Fatalf("first Start() should succeed, got %v", err)
	}

	err = lobby.Start()
	if err != ErrInvalidStateForStart {
		t.Errorf("expected ErrInvalidStateForStart on second call, got %v", err)
	}
}

// ========================================
// Concurrency Tests
// ========================================

func TestGetState_ThreadSafe(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = lobby.GetState()
		}()
	}
	wg.Wait()
	// No race condition should occur
}

func TestGetPlayers_ThreadSafe(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			players := lobby.GetPlayers()
			if len(players) < 1 {
				t.Error("expected at least 1 player")
			}
		}()
	}
	wg.Wait()
}

func TestConcurrent_MultipleJoins(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// 10 goroutines try to join simultaneously - only 1 should succeed
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := lobby.AddPlayer(
				"player-"+string(rune('0'+id)),
				"Player",
			)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful join, got %d", successCount)
	}
	if lobby.PlayerCount() != 2 {
		t.Errorf("expected 2 players total, got %d", lobby.PlayerCount())
	}
}

func TestConcurrent_JoinAndLeave(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")
	lobby.AddPlayer("player-2", "Player2")

	var wg sync.WaitGroup

	// Concurrent operations
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			lobby.RemovePlayer("player-2")
		}()
		go func() {
			defer wg.Done()
			lobby.AddPlayer("player-2", "Player2")
		}()
	}
	wg.Wait()

	// State should be consistent (either 1 or 2 players)
	count := lobby.PlayerCount()
	if count < 1 || count > 2 {
		t.Errorf("expected player count 1 or 2, got %d", count)
	}
}

func TestConcurrent_ReadWhileModify(t *testing.T) {
	lobby := NewLobby("ABC123", "host-1", "Host")

	var wg sync.WaitGroup

	// Start readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = lobby.GetState()
				_ = lobby.PlayerCount()
				_ = lobby.GetPlayers()
				_ = lobby.GetHostID()
				_ = lobby.HasPlayer("host-1")
				_ = lobby.IsHost("host-1")
				_ = lobby.CanStart()
			}
		}()
	}

	// Start writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			playerID := "player-" + string(rune('a'+id))
			for j := 0; j < 10; j++ {
				lobby.AddPlayer(playerID, "Player")
				lobby.RemovePlayer(playerID)
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions occur
}
