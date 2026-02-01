package websocket

import (
	"testing"
	"time"
)

const testTimeout = 2 * time.Second

// ========================================
// Harness Smoke Test
// ========================================

func TestHarness_Smoke(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create a lobby
	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Connect a client
	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Authenticate
	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to send auth: %v", err)
	}

	// Should receive authenticated response
	authPayload, err := client.AssertAuthSuccess(testTimeout)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	if authPayload.ReconnectToken == "" {
		t.Error("expected reconnect token to be set")
	}

	// Should receive lobby state
	lobbyPayload, err := client.AssertLobbyUpdated(testTimeout)
	if err != nil {
		t.Fatalf("failed to receive lobby update: %v", err)
	}

	if lobbyPayload.Lobby.Code != lobbyCode {
		t.Errorf("expected lobby code %s, got %s", lobbyCode, lobbyPayload.Lobby.Code)
	}

	// Verify player is connected via hub
	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Error("expected player to be connected")
	}
}

// ========================================
// Authentication Tests
// ========================================

func TestWS_Auth_Success(t *testing.T) {
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

	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to send auth: %v", err)
	}

	authPayload, err := client.AssertAuthSuccess(testTimeout)
	if err != nil {
		t.Fatalf("auth assertion failed: %v", err)
	}

	if authPayload.PlayerID != "player-1" {
		t.Errorf("expected player_id player-1, got %s", authPayload.PlayerID)
	}
	if authPayload.ReconnectToken == "" {
		t.Error("expected reconnect token")
	}
	if authPayload.SessionExpiresAt == 0 {
		t.Error("expected session expiry")
	}
}

func TestWS_Auth_PlayerNotInLobby(t *testing.T) {
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

	// Try to auth as player-2 who hasn't joined
	if err := client.SendAuth("player-2", lobbyCode); err != nil {
		t.Fatalf("failed to send auth: %v", err)
	}

	if err := client.ExpectError(ErrCodePlayerNotInLobby, testTimeout); err != nil {
		t.Fatalf("expected PLAYER_NOT_IN_LOBBY error: %v", err)
	}
}

func TestWS_Auth_VersionMismatch(t *testing.T) {
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

	// Send with wrong version
	payload := AuthenticatePayload{
		PlayerID:  "player-1",
		LobbyCode: lobbyCode,
	}
	env, _ := NewEnvelope(TypeAuthenticate, payload)
	env.Version = 999 // Wrong version

	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeVersionMismatch, testTimeout); err != nil {
		t.Fatalf("expected VERSION_MISMATCH error: %v", err)
	}
}

func TestWS_Auth_RequiresAuthForActions(t *testing.T) {
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

	// Try to send heartbeat without auth
	if err := client.SendHeartbeat(); err != nil {
		t.Fatalf("failed to send heartbeat: %v", err)
	}

	if err := client.ExpectError(ErrCodeAuthRequired, testTimeout); err != nil {
		t.Fatalf("expected AUTH_REQUIRED error: %v", err)
	}
}

// ========================================
// Message Ordering Tests
// ========================================

func TestWS_Ordering_SeqIncrement(t *testing.T) {
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

	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to send auth: %v", err)
	}

	// First message (authenticated)
	env1, err := client.Receive(testTimeout)
	if err != nil {
		t.Fatalf("failed to receive first message: %v", err)
	}
	seq1 := env1.Seq

	// Second message (lobby_updated)
	env2, err := client.Receive(testTimeout)
	if err != nil {
		t.Fatalf("failed to receive second message: %v", err)
	}
	seq2 := env2.Seq

	if seq2 <= seq1 {
		t.Errorf("expected seq to increment: seq1=%d, seq2=%d", seq1, seq2)
	}
}

func TestWS_Ordering_CorrelationID(t *testing.T) {
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
		t.Fatalf("failed to send auth: %v", err)
	}

	// Get authenticated response
	authEnv, err := client.ReceiveType(TypeAuthenticated, testTimeout)
	if err != nil {
		t.Fatalf("failed to receive authenticated: %v", err)
	}

	expectedCorr := "auth-player-1"
	if authEnv.CorrelationID != expectedCorr {
		t.Errorf("expected correlation_id %q, got %q", expectedCorr, authEnv.CorrelationID)
	}
}

func TestWS_Ordering_HeartbeatAck(t *testing.T) {
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
		t.Fatalf("failed to send auth: %v", err)
	}
	client.Drain()

	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player not connected")
	}

	// Send heartbeat
	if err := client.SendHeartbeat(); err != nil {
		t.Fatalf("failed to send heartbeat: %v", err)
	}

	// Should receive heartbeat_ack
	env, err := client.ReceiveType(TypeHeartbeatAck, testTimeout)
	if err != nil {
		t.Fatalf("failed to receive heartbeat_ack: %v", err)
	}

	expectedCorr := "heartbeat-player-1"
	if env.CorrelationID != expectedCorr {
		t.Errorf("expected correlation_id %q, got %q", expectedCorr, env.CorrelationID)
	}

	var payload HeartbeatAckPayload
	if err := env.ParsePayload(&payload); err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}

	if payload.ServerTime == 0 {
		t.Error("expected server_time to be set")
	}
}

// ========================================
// Broadcast Tests
// ========================================

func TestWS_Broadcast_PlayerReady(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create lobby with host
	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Join second player
	if err := ts.JoinLobby(lobbyCode, "player-2", "Player2"); err != nil {
		t.Fatalf("failed to join lobby: %v", err)
	}

	// Connect both clients
	client1, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect client1: %v", err)
	}
	defer client1.Close()

	client2, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect client2: %v", err)
	}
	defer client2.Close()

	// Authenticate both
	if err := client1.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth client1: %v", err)
	}
	if err := client2.SendAuth("player-2", lobbyCode); err != nil {
		t.Fatalf("failed to auth client2: %v", err)
	}

	// Wait for both to be connected and consume initial messages
	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player-1 not connected")
	}
	if !ts.WaitForPlayerConnected("player-2", testTimeout) {
		t.Fatal("player-2 not connected")
	}

	// Consume authenticated + lobby_updated for each client
	if _, err := client1.ReceiveType(TypeAuthenticated, testTimeout); err != nil {
		t.Fatalf("client1 auth: %v", err)
	}
	if _, err := client1.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("client1 lobby: %v", err)
	}
	if _, err := client2.ReceiveType(TypeAuthenticated, testTimeout); err != nil {
		t.Fatalf("client2 auth: %v", err)
	}
	if _, err := client2.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("client2 lobby: %v", err)
	}

	// Player 1 sets ready
	if err := client1.SendReady(true); err != nil {
		t.Fatalf("failed to send ready: %v", err)
	}

	// Both should receive lobby_updated with player_ready_changed event
	update1, err := client1.AssertLobbyUpdated(testTimeout)
	if err != nil {
		t.Fatalf("client1 failed to receive update: %v", err)
	}
	if update1.Event != LobbyEventPlayerReadyChanged {
		t.Errorf("expected event %s, got %s", LobbyEventPlayerReadyChanged, update1.Event)
	}

	update2, err := client2.AssertLobbyUpdated(testTimeout)
	if err != nil {
		t.Fatalf("client2 failed to receive update: %v", err)
	}
	if update2.Event != LobbyEventPlayerReadyChanged {
		t.Errorf("expected event %s, got %s", LobbyEventPlayerReadyChanged, update2.Event)
	}
}

func TestWS_Broadcast_BothReady_GameStarts(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create lobby with host
	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Join second player
	if err := ts.JoinLobby(lobbyCode, "player-2", "Player2"); err != nil {
		t.Fatalf("failed to join lobby: %v", err)
	}

	// Connect both clients
	client1, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect client1: %v", err)
	}
	defer client1.Close()

	client2, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect client2: %v", err)
	}
	defer client2.Close()

	// Authenticate both
	if err := client1.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth client1: %v", err)
	}
	if err := client2.SendAuth("player-2", lobbyCode); err != nil {
		t.Fatalf("failed to auth client2: %v", err)
	}

	// Wait for both connected
	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player-1 not connected")
	}
	if !ts.WaitForPlayerConnected("player-2", testTimeout) {
		t.Fatal("player-2 not connected")
	}

	// Drain initial messages
	client1.Drain()
	client2.Drain()

	// Both players set ready
	if err := client1.SendReady(true); err != nil {
		t.Fatalf("failed to send ready for client1: %v", err)
	}
	client1.Drain()
	client2.Drain()

	if err := client2.SendReady(true); err != nil {
		t.Fatalf("failed to send ready for client2: %v", err)
	}

	// Both should receive game_starting
	_, err = client1.ReceiveType(TypeGameStarting, testTimeout)
	if err != nil {
		t.Fatalf("client1 failed to receive game_starting: %v", err)
	}

	_, err = client2.ReceiveType(TypeGameStarting, testTimeout)
	if err != nil {
		t.Fatalf("client2 failed to receive game_starting: %v", err)
	}

	// Both should receive game_started
	_, err = client1.ReceiveType(TypeGameStarted, testTimeout)
	if err != nil {
		t.Fatalf("client1 failed to receive game_started: %v", err)
	}

	_, err = client2.ReceiveType(TypeGameStarted, testTimeout)
	if err != nil {
		t.Fatalf("client2 failed to receive game_started: %v", err)
	}
}

// ========================================
// Disconnect Tests
// ========================================

func TestWS_Disconnect_ClearsReadyState(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create lobby with host
	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Connect client
	client, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	// Authenticate
	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player not connected")
	}

	// Consume initial messages
	if _, err := client.ReceiveType(TypeAuthenticated, testTimeout); err != nil {
		t.Fatalf("auth response: %v", err)
	}
	if _, err := client.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("lobby update: %v", err)
	}

	// Set ready
	if err := client.SendReady(true); err != nil {
		t.Fatalf("failed to send ready: %v", err)
	}

	// Wait for ready state broadcast to confirm it was processed
	if _, err := client.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("ready broadcast: %v", err)
	}

	// Verify ready state is set
	if !ts.Handler.isPlayerReady(lobbyCode, "player-1") {
		t.Fatal("expected player to be ready")
	}

	// Disconnect
	client.Close()

	// Wait for disconnect to be processed
	if !ts.WaitForPlayerDisconnected("player-1", testTimeout) {
		t.Fatal("player still connected after close")
	}

	// Ready state should be cleared
	if ts.Handler.isPlayerReady(lobbyCode, "player-1") {
		t.Error("expected ready state to be cleared after disconnect")
	}
}

// ========================================
// Additional Auth Tests
// ========================================

func TestWS_Auth_InvalidPayload(t *testing.T) {
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

	// Send auth with missing required fields
	payload := AuthenticatePayload{
		PlayerID:  "", // Missing player_id
		LobbyCode: "", // Missing lobby_code
	}
	env, _ := NewEnvelope(TypeAuthenticate, payload)

	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeAuthFailed, testTimeout); err != nil {
		t.Fatalf("expected AUTH_FAILED error: %v", err)
	}
}

func TestWS_Auth_LobbyNotFound(t *testing.T) {
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

	// Try to auth with wrong lobby code
	if err := client.SendAuth("player-1", "WRONGCODE"); err != nil {
		t.Fatalf("failed to send auth: %v", err)
	}

	if err := client.ExpectError(ErrCodeLobbyNotFound, testTimeout); err != nil {
		t.Fatalf("expected LOBBY_NOT_FOUND error: %v", err)
	}
}

// ========================================
// Ready State Tests
// ========================================

func TestWS_Ready_Toggle(t *testing.T) {
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

	if err := client.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player not connected")
	}

	// Consume initial messages
	if _, err := client.ReceiveType(TypeAuthenticated, testTimeout); err != nil {
		t.Fatalf("auth: %v", err)
	}
	if _, err := client.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("lobby: %v", err)
	}

	// Set ready = true
	if err := client.SendReady(true); err != nil {
		t.Fatalf("failed to send ready true: %v", err)
	}

	update1, err := client.AssertLobbyUpdated(testTimeout)
	if err != nil {
		t.Fatalf("failed to receive update after ready true: %v", err)
	}
	if update1.Event != LobbyEventPlayerReadyChanged {
		t.Errorf("expected event %s, got %s", LobbyEventPlayerReadyChanged, update1.Event)
	}

	// Find player in lobby and verify ready state
	var player1Ready bool
	for _, p := range update1.Lobby.Players {
		if p.ID == "player-1" {
			player1Ready = p.IsReady
			break
		}
	}
	if !player1Ready {
		t.Error("expected player to be ready in lobby update")
	}

	// Set ready = false
	if err := client.SendReady(false); err != nil {
		t.Fatalf("failed to send ready false: %v", err)
	}

	update2, err := client.AssertLobbyUpdated(testTimeout)
	if err != nil {
		t.Fatalf("failed to receive update after ready false: %v", err)
	}
	if update2.Event != LobbyEventPlayerReadyChanged {
		t.Errorf("expected event %s, got %s", LobbyEventPlayerReadyChanged, update2.Event)
	}

	// Verify player is no longer ready
	for _, p := range update2.Lobby.Players {
		if p.ID == "player-1" {
			if p.IsReady {
				t.Error("expected player to not be ready after toggle")
			}
			break
		}
	}
}

func TestWS_Ready_RequiresAuth(t *testing.T) {
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

	// Try to set ready without authenticating
	if err := client.SendReady(true); err != nil {
		t.Fatalf("failed to send ready: %v", err)
	}

	if err := client.ExpectError(ErrCodeAuthRequired, testTimeout); err != nil {
		t.Fatalf("expected AUTH_REQUIRED error: %v", err)
	}
}

// ========================================
// Error Handling Tests
// ========================================

func TestWS_Error_UnknownMessageType(t *testing.T) {
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

	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player not connected")
	}
	client.Drain()

	// Send unknown message type
	env := &Envelope{
		Type:      MessageType("unknown_type"),
		Version:   ProtocolVersion,
		Timestamp: 1234567890,
		Payload:   []byte("{}"),
	}

	if err := client.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if err := client.ExpectError(ErrCodeMalformedMessage, testTimeout); err != nil {
		t.Fatalf("expected MALFORMED_MESSAGE error: %v", err)
	}
}

// ========================================
// Hub Integration Tests
// ========================================

func TestWS_BroadcastToLobbyExcept_ExcludedPlayerDoesNotReceive(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create lobby with one player initially
	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Connect client1 first, before adding player-2
	client1, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect client1: %v", err)
	}
	defer client1.Close()

	if err := client1.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth client1: %v", err)
	}
	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player-1 not connected")
	}

	// Receive client1's auth messages
	if _, err := client1.ReceiveType(TypeAuthenticated, testTimeout); err != nil {
		t.Fatalf("client1 failed to receive authenticated: %v", err)
	}
	if _, err := client1.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("client1 failed to receive lobby_state: %v", err)
	}

	// Now add and connect player-2
	if err := ts.JoinLobby(lobbyCode, "player-2", "Player2"); err != nil {
		t.Fatalf("failed to join lobby: %v", err)
	}

	client2, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect client2: %v", err)
	}
	defer client2.Close()

	if err := client2.SendAuth("player-2", lobbyCode); err != nil {
		t.Fatalf("failed to auth client2: %v", err)
	}
	if !ts.WaitForPlayerConnected("player-2", testTimeout) {
		t.Fatal("player-2 not connected")
	}

	// Receive client2's auth messages
	if _, err := client2.ReceiveType(TypeAuthenticated, testTimeout); err != nil {
		t.Fatalf("client2 failed to receive authenticated: %v", err)
	}
	if _, err := client2.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("client2 failed to receive lobby_state: %v", err)
	}

	// Client1 may have received notifications about player-2 connecting - drain them
	client1.Drain()
	time.Sleep(50 * time.Millisecond)
	client1.Drain()

	// Clear client2's buffer too
	client2.Drain()

	// Send a unique message type to verify routing - use game_starting since it's distinctive
	ts.Hub.BroadcastToLobbyExcept(lobbyCode, "player-1", TypeGameStarting, GameStartingPayload{
		StartsAt:     12345,
		CountdownSec: 3,
	})

	// Client2 should receive the broadcast
	_, err = client2.ReceiveType(TypeGameStarting, testTimeout)
	if err != nil {
		t.Fatalf("client2 should receive broadcast: %v", err)
	}

	// Client1 should NOT receive the game_starting message
	env, err := client1.Receive(200 * time.Millisecond)
	if err == nil && env.Type == TypeGameStarting {
		t.Error("client1 should NOT receive broadcast (was excluded)")
	}
}

func TestWS_SendToPlayer_OnlyTargetReceives(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create lobby with one player initially
	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Connect client1 first
	client1, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect client1: %v", err)
	}
	defer client1.Close()

	if err := client1.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth client1: %v", err)
	}
	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player-1 not connected")
	}

	// Receive client1's auth messages
	if _, err := client1.ReceiveType(TypeAuthenticated, testTimeout); err != nil {
		t.Fatalf("client1 failed to receive authenticated: %v", err)
	}
	if _, err := client1.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("client1 failed to receive lobby_state: %v", err)
	}

	// Now add and connect player-2
	if err := ts.JoinLobby(lobbyCode, "player-2", "Player2"); err != nil {
		t.Fatalf("failed to join lobby: %v", err)
	}

	client2, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect client2: %v", err)
	}
	defer client2.Close()

	if err := client2.SendAuth("player-2", lobbyCode); err != nil {
		t.Fatalf("failed to auth client2: %v", err)
	}
	if !ts.WaitForPlayerConnected("player-2", testTimeout) {
		t.Fatal("player-2 not connected")
	}

	// Receive client2's auth messages
	if _, err := client2.ReceiveType(TypeAuthenticated, testTimeout); err != nil {
		t.Fatalf("client2 failed to receive authenticated: %v", err)
	}
	if _, err := client2.ReceiveType(TypeLobbyUpdated, testTimeout); err != nil {
		t.Fatalf("client2 failed to receive lobby_state: %v", err)
	}

	// Drain any cross-player notifications
	client1.Drain()
	client2.Drain()
	time.Sleep(50 * time.Millisecond)
	client1.Drain()
	client2.Drain()

	// Send game_starting only to player-1 (distinctive message type)
	ts.Hub.SendToPlayer("player-1", TypeGameStarting, GameStartingPayload{
		StartsAt:     12345,
		CountdownSec: 3,
	})

	// Client1 should receive the message
	_, err = client1.ReceiveType(TypeGameStarting, testTimeout)
	if err != nil {
		t.Fatalf("client1 should receive message: %v", err)
	}

	// Client2 should NOT receive game_starting
	env, err := client2.Receive(200 * time.Millisecond)
	if err == nil && env.Type == TypeGameStarting {
		t.Error("client2 should NOT receive message (not targeted)")
	}
}

func TestWS_DisconnectPlayer_PlayerDisconnected(t *testing.T) {
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

	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player not connected")
	}
	client.Drain()

	// Force disconnect via hub
	ts.Hub.DisconnectPlayer("player-1")

	// Player should be disconnected
	if !ts.WaitForPlayerDisconnected("player-1", testTimeout) {
		t.Error("expected player to be disconnected")
	}
}

// ========================================
// Reconnection Flow Tests
// ========================================

func TestWS_Reconnect_ValidToken(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Connect and authenticate first time
	client1, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if err := client1.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	authPayload, err := client1.AssertAuthSuccess(testTimeout)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	reconnectToken := authPayload.ReconnectToken
	if reconnectToken == "" {
		t.Fatal("expected reconnect token")
	}

	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player not connected")
	}

	// Close first connection
	client1.Close()
	if !ts.WaitForPlayerDisconnected("player-1", testTimeout) {
		t.Fatal("player still connected after close")
	}

	// Reconnect with token
	client2, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to reconnect: %v", err)
	}
	defer client2.Close()

	// Set PlayerID on client2 before sending auth
	client2.PlayerID = "player-1"
	client2.LobbyCode = lobbyCode

	// Send auth with reconnect token
	payload := AuthenticatePayload{
		PlayerID:       "player-1",
		LobbyCode:      lobbyCode,
		ReconnectToken: reconnectToken,
	}
	env, _ := NewEnvelope(TypeAuthenticate, payload)
	env.CorrelationID = "reconnect-auth"
	if err := client2.Send(env); err != nil {
		t.Fatalf("failed to send reconnect auth: %v", err)
	}

	// Should succeed
	_, err = client2.AssertAuthSuccess(testTimeout)
	if err != nil {
		t.Fatalf("reconnect auth failed: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Error("player should be connected after reconnect")
	}
}

func TestWS_Reconnect_InvalidToken(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	lobbyCode, err := ts.CreateLobby("player-1", "Player1")
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	// Connect and authenticate first time to establish session
	client1, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if err := client1.SendAuth("player-1", lobbyCode); err != nil {
		t.Fatalf("failed to auth: %v", err)
	}

	_, err = client1.AssertAuthSuccess(testTimeout)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	if !ts.WaitForPlayerConnected("player-1", testTimeout) {
		t.Fatal("player not connected")
	}

	// Try to reconnect with invalid token while still connected
	// This tests the reconnection validation path
	client2, err := NewTestClient(ts.WebSocketURL(lobbyCode))
	if err != nil {
		t.Fatalf("failed to connect second client: %v", err)
	}
	defer client2.Close()

	// Set PlayerID on client2 before sending auth
	client2.PlayerID = "player-1"
	client2.LobbyCode = lobbyCode

	// Send auth with invalid reconnect token - should still work as new auth
	// (the token is just ignored if invalid, and we proceed with regular auth)
	payload := AuthenticatePayload{
		PlayerID:       "player-1",
		LobbyCode:      lobbyCode,
		ReconnectToken: "invalid-token-that-does-not-exist",
	}
	env, _ := NewEnvelope(TypeAuthenticate, payload)
	env.CorrelationID = "invalid-reconnect"
	if err := client2.Send(env); err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	// Should still succeed (new session replaces old)
	_, err = client2.AssertAuthSuccess(testTimeout)
	if err != nil {
		t.Fatalf("auth should succeed even with invalid reconnect token: %v", err)
	}

	// Clean up first client
	client1.Close()
}
