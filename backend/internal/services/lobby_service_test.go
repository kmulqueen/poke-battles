package services

import (
	"errors"
	"sync"
	"testing"

	"poke-battles/internal/game"
)

// ========================================
// Happy Path Tests
// ========================================

func TestCreateLobby_Success(t *testing.T) {
	svc := NewLobbyService()

	lobby, err := svc.CreateLobby("host-1", "HostPlayer")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lobby == nil {
		t.Fatal("expected lobby to be non-nil")
	}
	if len(lobby.Code) != 6 {
		t.Errorf("expected 6-char code, got %q", lobby.Code)
	}
	if !lobby.HasPlayer("host-1") {
		t.Error("expected host to be in lobby")
	}
	if lobby.GetState() != game.LobbyStateWaiting {
		t.Errorf("expected state Waiting, got %v", lobby.GetState())
	}
}

func TestCreateLobby_UniqueRoomCodes(t *testing.T) {
	svc := NewLobbyService()
	codes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		lobby, err := svc.CreateLobby("host-"+string(rune('0'+i)), "Host")
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		if codes[lobby.Code] {
			t.Errorf("duplicate room code: %q", lobby.Code)
		}
		codes[lobby.Code] = true
	}
}

func TestJoinLobby_Success(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")

	lobby, err := svc.JoinLobby(created.Code, "player-2", "Player2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lobby.PlayerCount() != 2 {
		t.Errorf("expected 2 players, got %d", lobby.PlayerCount())
	}
	if !lobby.HasPlayer("player-2") {
		t.Error("expected player-2 in lobby")
	}
	if lobby.GetState() != game.LobbyStateReady {
		t.Errorf("expected state Ready, got %v", lobby.GetState())
	}
}

func TestLeaveLobby_Success(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")
	svc.JoinLobby(created.Code, "player-2", "Player2")

	err := svc.LeaveLobby(created.Code, "player-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	lobby, _ := svc.GetLobby(created.Code)
	if lobby.PlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", lobby.PlayerCount())
	}
	if lobby.HasPlayer("player-2") {
		t.Error("expected player-2 to be removed")
	}
}

func TestLeaveLobby_DeletesEmptyLobby(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")
	code := created.Code

	err := svc.LeaveLobby(code, "host-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = svc.GetLobby(code)
	if !errors.Is(err, ErrLobbyNotFound) {
		t.Errorf("expected ErrLobbyNotFound after last player leaves, got %v", err)
	}
}

func TestGetLobby_Success(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")

	lobby, err := svc.GetLobby(created.Code)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lobby.Code != created.Code {
		t.Errorf("expected code %q, got %q", created.Code, lobby.Code)
	}
}

func TestStartGame_Success(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")
	svc.JoinLobby(created.Code, "player-2", "Player2")

	err := svc.StartGame(created.Code, "host-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	lobby, _ := svc.GetLobby(created.Code)
	if lobby.GetState() != game.LobbyStateActive {
		t.Errorf("expected state Active, got %v", lobby.GetState())
	}
}

// ========================================
// Validation Error Tests
// ========================================

func TestJoinLobby_NotFound(t *testing.T) {
	svc := NewLobbyService()

	_, err := svc.JoinLobby("NOTFOUND", "player-1", "Player")
	if !errors.Is(err, ErrLobbyNotFound) {
		t.Errorf("expected ErrLobbyNotFound, got %v", err)
	}
}

func TestJoinLobby_LobbyFull(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")
	svc.JoinLobby(created.Code, "player-2", "Player2")

	// When lobby has 2 players, state becomes Ready.
	// The state check happens before "full" check, so we get ErrInvalidStateForJoin.
	_, err := svc.JoinLobby(created.Code, "player-3", "Player3")
	if !errors.Is(err, game.ErrInvalidStateForJoin) {
		t.Errorf("expected ErrInvalidStateForJoin (state=Ready prevents join), got %v", err)
	}
}

func TestJoinLobby_PlayerAlreadyJoined(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")

	_, err := svc.JoinLobby(created.Code, "host-1", "HostAgain")
	if !errors.Is(err, game.ErrPlayerAlreadyJoined) {
		t.Errorf("expected ErrPlayerAlreadyJoined, got %v", err)
	}
}

func TestGetLobby_NotFound(t *testing.T) {
	svc := NewLobbyService()

	_, err := svc.GetLobby("NOTFOUND")
	if !errors.Is(err, ErrLobbyNotFound) {
		t.Errorf("expected ErrLobbyNotFound, got %v", err)
	}
}

func TestLeaveLobby_NotFound(t *testing.T) {
	svc := NewLobbyService()

	err := svc.LeaveLobby("NOTFOUND", "player-1")
	if !errors.Is(err, ErrLobbyNotFound) {
		t.Errorf("expected ErrLobbyNotFound, got %v", err)
	}
}

func TestLeaveLobby_PlayerNotFound(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")

	err := svc.LeaveLobby(created.Code, "nonexistent")
	if !errors.Is(err, game.ErrPlayerNotFound) {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestStartGame_NotFound(t *testing.T) {
	svc := NewLobbyService()

	err := svc.StartGame("NOTFOUND", "player-1")
	if !errors.Is(err, ErrLobbyNotFound) {
		t.Errorf("expected ErrLobbyNotFound, got %v", err)
	}
}

func TestStartGame_NotHost(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")
	svc.JoinLobby(created.Code, "player-2", "Player2")

	err := svc.StartGame(created.Code, "player-2")
	if !errors.Is(err, ErrNotHost) {
		t.Errorf("expected ErrNotHost, got %v", err)
	}
}

func TestStartGame_InvalidState(t *testing.T) {
	svc := NewLobbyService()

	created, _ := svc.CreateLobby("host-1", "Host")
	// Only 1 player, state is Waiting

	err := svc.StartGame(created.Code, "host-1")
	if !errors.Is(err, game.ErrInvalidStateForStart) {
		t.Errorf("expected ErrInvalidStateForStart, got %v", err)
	}
}

// ========================================
// Full Flow Integration Tests
// ========================================

func TestFullFlow_CreateJoinReady(t *testing.T) {
	svc := NewLobbyService()

	// Host creates lobby
	lobby, err := svc.CreateLobby("host-1", "Host")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if lobby.GetState() != game.LobbyStateWaiting {
		t.Errorf("expected Waiting state after create, got %v", lobby.GetState())
	}

	// Player joins
	lobby, err = svc.JoinLobby(lobby.Code, "player-2", "Player2")
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}
	if lobby.GetState() != game.LobbyStateReady {
		t.Errorf("expected Ready state after join, got %v", lobby.GetState())
	}
	if lobby.PlayerCount() != 2 {
		t.Errorf("expected 2 players, got %d", lobby.PlayerCount())
	}
}

func TestFullFlow_CreateJoinLeaveRejoin(t *testing.T) {
	svc := NewLobbyService()

	// Host creates lobby
	lobby, _ := svc.CreateLobby("host-1", "Host")
	code := lobby.Code

	// Player joins
	svc.JoinLobby(code, "player-2", "Player2")

	// Player leaves
	err := svc.LeaveLobby(code, "player-2")
	if err != nil {
		t.Fatalf("leave failed: %v", err)
	}

	lobby, _ = svc.GetLobby(code)
	if lobby.GetState() != game.LobbyStateWaiting {
		t.Errorf("expected Waiting state after leave, got %v", lobby.GetState())
	}

	// Player rejoins
	lobby, err = svc.JoinLobby(code, "player-2", "Player2")
	if err != nil {
		t.Fatalf("rejoin failed: %v", err)
	}
	if lobby.GetState() != game.LobbyStateReady {
		t.Errorf("expected Ready state after rejoin, got %v", lobby.GetState())
	}
}

func TestFullFlow_HostLeaveReassign(t *testing.T) {
	svc := NewLobbyService()

	// Host creates lobby
	lobby, _ := svc.CreateLobby("host-1", "Host")
	code := lobby.Code

	// Player joins
	svc.JoinLobby(code, "player-2", "Player2")

	// Host leaves
	err := svc.LeaveLobby(code, "host-1")
	if err != nil {
		t.Fatalf("host leave failed: %v", err)
	}

	// Verify player-2 is now host
	lobby, _ = svc.GetLobby(code)
	if !lobby.IsHost("player-2") {
		t.Error("expected player-2 to be new host")
	}
	if lobby.PlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", lobby.PlayerCount())
	}
}

func TestFullFlow_StartAndVerifyActive(t *testing.T) {
	svc := NewLobbyService()

	// Create and fill lobby
	lobby, _ := svc.CreateLobby("host-1", "Host")
	svc.JoinLobby(lobby.Code, "player-2", "Player2")

	// Start game
	err := svc.StartGame(lobby.Code, "host-1")
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	// Verify active state
	lobby, _ = svc.GetLobby(lobby.Code)
	if lobby.GetState() != game.LobbyStateActive {
		t.Errorf("expected Active state, got %v", lobby.GetState())
	}
}

// ========================================
// Concurrency Tests
// ========================================

func TestConcurrent_CreateLobbies(t *testing.T) {
	svc := NewLobbyService()

	var wg sync.WaitGroup
	lobbyCodes := make(chan string, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			lobby, err := svc.CreateLobby("host-"+string(rune(id)), "Host")
			if err != nil {
				t.Errorf("create failed: %v", err)
				return
			}
			lobbyCodes <- lobby.Code
		}(i)
	}
	wg.Wait()
	close(lobbyCodes)

	// Verify all codes are unique
	seen := make(map[string]bool)
	for code := range lobbyCodes {
		if seen[code] {
			t.Errorf("duplicate lobby code: %q", code)
		}
		seen[code] = true
	}
}

func TestConcurrent_JoinSameLobby(t *testing.T) {
	svc := NewLobbyService()

	lobby, _ := svc.CreateLobby("host-1", "Host")
	code := lobby.Code

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// 10 players try to join - only 1 should succeed
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := svc.JoinLobby(code, "player-"+string(rune('0'+id)), "Player")
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if successCount != 1 {
		t.Errorf("expected 1 successful join, got %d", successCount)
	}

	lobby, _ = svc.GetLobby(code)
	if lobby.PlayerCount() != 2 {
		t.Errorf("expected 2 players, got %d", lobby.PlayerCount())
	}
}

func TestConcurrent_GetAndModify(t *testing.T) {
	svc := NewLobbyService()

	lobby, _ := svc.CreateLobby("host-1", "Host")
	code := lobby.Code
	svc.JoinLobby(code, "player-2", "Player2")

	var wg sync.WaitGroup

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.GetLobby(code)
		}()
	}

	// Concurrent modifications
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.LeaveLobby(code, "player-2")
			svc.JoinLobby(code, "player-2", "Player2")
		}()
	}

	wg.Wait()
	// No race conditions should occur
}
