package websocket

import (
	"testing"
	"time"
)

const handlerTestTimeout = 2 * time.Second

// ========================================
// handleSubmitAction Tests
// ========================================

func TestHandler_SubmitAction_RequiresAuth(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Send submit_action without authenticating
	env, _ := NewEnvelope(TypeSubmitAction, map[string]interface{}{
		"action_type": "attack",
	})
	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeAuthRequired, handlerTestTimeout); err != nil {
		t.Fatalf("expected AUTH_REQUIRED error: %v", err)
	}
}

func TestHandler_SubmitAction_NoActiveBattle(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Authenticate first
	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", handlerTestTimeout) {
		t.Fatal("player not connected")
	}
	client.Drain()

	// Send submit_action when there is no active battle
	env, _ := NewEnvelope(TypeSubmitAction, map[string]interface{}{
		"action_type": "attack",
	})
	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeInvalidState, handlerTestTimeout); err != nil {
		t.Fatalf("expected INVALID_STATE error: %v", err)
	}
}

// ========================================
// handleRequestGameState Tests
// ========================================

func TestHandler_RequestGameState_RequiresAuth(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Send request_game_state without authenticating
	env, _ := NewEnvelope(TypeRequestGameState, map[string]interface{}{})
	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeAuthRequired, handlerTestTimeout); err != nil {
		t.Fatalf("expected AUTH_REQUIRED error: %v", err)
	}
}

func TestHandler_RequestGameState_NoActiveBattle(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Authenticate first
	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", handlerTestTimeout) {
		t.Fatal("player not connected")
	}
	client.Drain()

	// Send request_game_state when there is no active battle
	env, _ := NewEnvelope(TypeRequestGameState, map[string]interface{}{})
	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeInvalidState, handlerTestTimeout); err != nil {
		t.Fatalf("expected INVALID_STATE error: %v", err)
	}
}

// ========================================
// handleRequestRematch Tests
// ========================================

func TestHandler_RequestRematch_RequiresAuth(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Send request_rematch without authenticating
	env, _ := NewEnvelope(TypeRequestRematch, map[string]interface{}{})
	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeAuthRequired, handlerTestTimeout); err != nil {
		t.Fatalf("expected AUTH_REQUIRED error: %v", err)
	}
}

func TestHandler_RequestRematch_NoGame(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Authenticate first
	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", handlerTestTimeout) {
		t.Fatal("player not connected")
	}
	client.Drain()

	// Send request_rematch when there is no game
	env, _ := NewEnvelope(TypeRequestRematch, map[string]interface{}{})
	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeInvalidState, handlerTestTimeout); err != nil {
		t.Fatalf("expected INVALID_STATE error: %v", err)
	}
}

// ========================================
// handleLeaveGame Tests
// ========================================

func TestHandler_LeaveGame_RequiresAuth(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Send leave_game without authenticating
	env, _ := NewEnvelope(TypeLeaveGame, map[string]interface{}{})
	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeAuthRequired, handlerTestTimeout); err != nil {
		t.Fatalf("expected AUTH_REQUIRED error: %v", err)
	}
}

func TestHandler_LeaveGame_Success(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Add a second player so the lobby persists when player-1 leaves
	if err := ts.JoinLobby(lobbyCode, "player-2", "Player2"); err != nil {
		t.Fatalf("failed to join lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Authenticate
	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", handlerTestTimeout) {
		t.Fatal("player not connected")
	}
	client.Drain()

	// Send leave_game
	env, _ := NewEnvelope(TypeLeaveGame, map[string]interface{}{})
	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	// Player should be disconnected
	if !ts.WaitForPlayerDisconnected("player-1", handlerTestTimeout) {
		t.Error("expected player to be disconnected after leave_game")
	}
}

// ========================================
// BroadcastPlayerJoined Tests
// ========================================

func TestHandler_BroadcastPlayerJoined(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Authenticate
	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", handlerTestTimeout) {
		t.Fatal("player not connected")
	}

	// Explicitly receive expected auth messages
	if _, err := client.ReceiveType(TypeAuthenticated, handlerTestTimeout); err != nil {
		t.Fatalf("failed to receive authenticated: %v", err)
	}
	if _, err := client.ReceiveType(TypeLobbyUpdated, handlerTestTimeout); err != nil {
		t.Fatalf("failed to receive lobby_state: %v", err)
	}

	// Broadcast player joined
	ts.Handler.BroadcastPlayerJoined(lobbyCode, "player-2", "Player2")

	// Client should receive lobby_updated with player_joined event
	update, err := client.AssertLobbyUpdated(handlerTestTimeout)
	if err != nil {
		t.Fatalf("failed to receive lobby update: %v", err)
	}

	if update.Event != LobbyEventPlayerJoined {
		t.Errorf("expected event %s, got %s", LobbyEventPlayerJoined, update.Event)
	}
}

// ========================================
// BroadcastPlayerLeft Tests
// ========================================

func TestHandler_BroadcastPlayerLeft(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Add second player so lobby has 2 players
	if err := ts.JoinLobby(lobbyCode, "player-2", "Player2"); err != nil {
		t.Fatalf("failed to join lobby: %v", err)
	}

	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Authenticate as player-1
	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", handlerTestTimeout) {
		t.Fatal("player not connected")
	}

	// Explicitly receive expected auth messages
	if _, err := client.ReceiveType(TypeAuthenticated, handlerTestTimeout); err != nil {
		t.Fatalf("failed to receive authenticated: %v", err)
	}
	if _, err := client.ReceiveType(TypeLobbyUpdated, handlerTestTimeout); err != nil {
		t.Fatalf("failed to receive lobby_state: %v", err)
	}

	// Broadcast that player left (don't actually remove - just test the broadcast)
	ts.Handler.BroadcastPlayerLeft(lobbyCode, "player-2")

	// Client should receive lobby_updated with player_left event
	update, err := client.AssertLobbyUpdated(handlerTestTimeout)
	if err != nil {
		t.Fatalf("failed to receive lobby update: %v", err)
	}

	if update.Event != LobbyEventPlayerLeft {
		t.Errorf("expected event %s, got %s", LobbyEventPlayerLeft, update.Event)
	}
}

// ========================================
// BroadcastPlayerJoined / BroadcastPlayerLeft Edge Cases
// ========================================

func TestHandler_BroadcastPlayerJoined_NonExistentLobby(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Should not panic when lobby doesn't exist
	ts.Handler.BroadcastPlayerJoined("NONEXISTENT", "player-1", "Player1")
}

func TestHandler_BroadcastPlayerLeft_NonExistentLobby(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Should not panic when lobby doesn't exist
	ts.Handler.BroadcastPlayerLeft("NONEXISTENT", "player-1")
}
