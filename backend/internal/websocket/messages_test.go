package websocket

import (
	"encoding/json"
	"testing"
)

// ========================================
// Test Helpers (for pure struct tests)
// ========================================

// createTestEnvelope creates a test envelope for payload parsing tests
func createTestEnvelope(msgType MessageType, payload interface{}) *Envelope {
	payloadBytes, _ := json.Marshal(payload)
	return &Envelope{
		Type:          msgType,
		Version:       ProtocolVersion,
		Timestamp:     1234567890,
		CorrelationID: "test-corr-id",
		Payload:       payloadBytes,
	}
}

// ========================================
// Message Type Tests
// ========================================

func TestMessageTypes(t *testing.T) {
	// Verify all expected message types are defined
	clientToServer := []MessageType{
		TypeAuthenticate,
		TypeHeartbeat,
		TypeRequestLobbyState,
		TypeSetReady,
		TypeSubmitAction,
		TypeRequestGameState,
		TypeRequestRematch,
		TypeLeaveGame,
	}

	serverToClient := []MessageType{
		TypeAuthenticated,
		TypeHeartbeatAck,
		TypeLobbyUpdated,
		TypeGameStarting,
		TypeGameStarted,
		TypeGameState,
		TypeActionAcknowledged,
		TypeTurnResult,
		TypeSwitchRequired,
		TypeGameEnded,
		TypeRematchRequested,
		TypeRematchStarting,
		TypeError,
		TypeDisconnectWarning,
	}

	for _, msgType := range clientToServer {
		if msgType == "" {
			t.Errorf("empty message type found in client-to-server types")
		}
	}

	for _, msgType := range serverToClient {
		if msgType == "" {
			t.Errorf("empty message type found in server-to-client types")
		}
	}
}

// ========================================
// Error Code Tests
// ========================================

func TestErrorCodeRecoverability(t *testing.T) {
	recoverableCodes := []ErrorCode{
		ErrCodeInvalidState,
		ErrCodeInvalidAction,
		ErrCodeNotYourTurn,
		ErrCodeTurnMismatch,
		ErrCodeMalformedMessage,
	}

	nonRecoverableCodes := []ErrorCode{
		ErrCodeAuthRequired,
		ErrCodeAuthFailed,
		ErrCodeSessionExpired,
		ErrCodeLobbyNotFound,
		ErrCodeLobbyFull,
		ErrCodeInternalError,
		ErrCodeVersionMismatch,
		ErrCodePlayerNotInLobby,
	}

	for _, code := range recoverableCodes {
		if !IsRecoverable(code) {
			t.Errorf("expected %q to be recoverable", code)
		}
	}

	for _, code := range nonRecoverableCodes {
		if IsRecoverable(code) {
			t.Errorf("expected %q to not be recoverable", code)
		}
	}
}

func TestNewErrorPayload(t *testing.T) {
	payload := NewErrorPayload(ErrCodeAuthRequired, "Test message")

	if payload.Code != ErrCodeAuthRequired {
		t.Errorf("expected code %q, got %q", ErrCodeAuthRequired, payload.Code)
	}
	if payload.Message != "Test message" {
		t.Errorf("expected message %q, got %q", "Test message", payload.Message)
	}
	if payload.Recoverable != IsRecoverable(ErrCodeAuthRequired) {
		t.Error("recoverable mismatch")
	}
}

func TestNewErrorPayloadWithDetails(t *testing.T) {
	details := map[string]string{"key": "value"}
	payload, err := NewErrorPayloadWithDetails(ErrCodeInvalidState, "Test", details)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Details == nil {
		t.Error("expected details to be set")
	}

	var parsed map[string]string
	json.Unmarshal(payload.Details, &parsed)
	if parsed["key"] != "value" {
		t.Errorf("expected details key to be 'value', got %q", parsed["key"])
	}
}

// ========================================
// Envelope Tests
// ========================================

func TestEnvelope_NewEnvelope(t *testing.T) {
	payload := map[string]string{"test": "value"}
	env, err := NewEnvelope(TypeLobbyUpdated, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.Type != TypeLobbyUpdated {
		t.Errorf("expected type %q, got %q", TypeLobbyUpdated, env.Type)
	}
	if env.Version != ProtocolVersion {
		t.Errorf("expected version %d, got %d", ProtocolVersion, env.Version)
	}
	if env.Timestamp == 0 {
		t.Error("expected timestamp to be set")
	}
}

func TestEnvelope_NewEnvelopeWithSeq(t *testing.T) {
	payload := map[string]string{"test": "value"}
	env, err := NewEnvelopeWithSeq(TypeLobbyUpdated, 42, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.Seq != 42 {
		t.Errorf("expected seq 42, got %d", env.Seq)
	}
}

func TestEnvelope_WithCorrelationID(t *testing.T) {
	env, _ := NewEnvelope(TypeLobbyUpdated, nil)
	env = env.WithCorrelationID("corr-123")

	if env.CorrelationID != "corr-123" {
		t.Errorf("expected correlation ID 'corr-123', got %q", env.CorrelationID)
	}
}

func TestEnvelope_ParsePayload(t *testing.T) {
	original := SetReadyPayload{Ready: true}
	env := createTestEnvelope(TypeSetReady, original)

	var parsed SetReadyPayload
	err := env.ParsePayload(&parsed)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Ready != true {
		t.Error("expected parsed Ready to be true")
	}
}

// ========================================
// Lobby Event Types Tests
// ========================================

func TestLobbyEventTypes(t *testing.T) {
	events := []LobbyEvent{
		LobbyEventPlayerJoined,
		LobbyEventPlayerLeft,
		LobbyEventPlayerReadyChanged,
		LobbyEventHostChanged,
		LobbyEventStateChanged,
	}

	for _, event := range events {
		if event == "" {
			t.Error("empty lobby event type found")
		}
	}
}

// ========================================
// Payload Struct Tests
// ========================================

func TestAuthenticatePayload_Marshalling(t *testing.T) {
	payload := AuthenticatePayload{
		PlayerID:       "player-1",
		SessionToken:   "token-123",
		LobbyCode:      "ABC123",
		ReconnectToken: "reconnect-456",
		LastSeq:        10,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed AuthenticatePayload
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if parsed.PlayerID != payload.PlayerID {
		t.Errorf("expected player_id %q, got %q", payload.PlayerID, parsed.PlayerID)
	}
	if parsed.LobbyCode != payload.LobbyCode {
		t.Errorf("expected lobby_code %q, got %q", payload.LobbyCode, parsed.LobbyCode)
	}
}

func TestSetReadyPayload_Marshalling(t *testing.T) {
	payload := SetReadyPayload{Ready: true}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed SetReadyPayload
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if parsed.Ready != payload.Ready {
		t.Errorf("expected ready %v, got %v", payload.Ready, parsed.Ready)
	}
}

func TestLobbyInfo_MarshalEventData(t *testing.T) {
	lobbyInfo := LobbyInfo{
		Code:  "ABC123",
		State: "ready",
	}

	eventData := PlayerJoinedEventData{
		PlayerID: "player-1",
		Username: "Player1",
	}

	data, err := lobbyInfo.MarshalEventData(eventData)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed PlayerJoinedEventData
	json.Unmarshal(data, &parsed)

	if parsed.PlayerID != eventData.PlayerID {
		t.Errorf("expected player_id %q, got %q", eventData.PlayerID, parsed.PlayerID)
	}
}

func TestLobbyInfo_MarshalEventData_Nil(t *testing.T) {
	lobbyInfo := LobbyInfo{}

	data, err := lobbyInfo.MarshalEventData(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data != nil {
		t.Error("expected nil data for nil input")
	}
}

// ========================================
// Hub Edge Case Tests (no time.Sleep)
// ========================================

func TestHub_ConnectionLifecycle(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	if hub.ConnectionCount() != 0 {
		t.Errorf("expected 0 connections initially, got %d", hub.ConnectionCount())
	}

	// Test lobby connection count for non-existent lobby
	if hub.LobbyConnectionCount("NONEXIST") != 0 {
		t.Error("expected 0 connections for non-existent lobby")
	}

	// Test player connected for non-existent player
	if hub.IsPlayerConnected("nonexistent") {
		t.Error("expected player to not be connected")
	}
}

func TestHub_GetConnectionByPlayerID_NotFound(t *testing.T) {
	hub := NewHub()

	conn := hub.GetConnectionByPlayerID("nonexistent")
	if conn != nil {
		t.Error("expected nil connection for non-existent player")
	}
}

func TestHub_GetLobbyConnections_NotFound(t *testing.T) {
	hub := NewHub()

	conns := hub.GetLobbyConnections("NONEXIST")
	if conns != nil {
		t.Error("expected nil connections for non-existent lobby")
	}
}

// ========================================
// Hub Broadcast Edge Cases (no connected players)
// ========================================

func TestHub_BroadcastToLobbyExcept_EmptyLobby(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Should not error or panic when lobby is empty
	err := hub.BroadcastToLobbyExcept("NONEXIST", "player-1", TypeLobbyUpdated, map[string]string{"test": "value"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHub_SendToPlayer_NotConnected(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Should return nil (not an error) when player is not connected
	err := hub.SendToPlayer("nonexistent", TypeLobbyUpdated, map[string]string{"test": "value"})
	if err != nil {
		t.Errorf("expected nil error for disconnected player, got %v", err)
	}
}

func TestHub_SendToPlayerWithCorrelation_NotConnected(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Should return nil (not an error) when player is not connected
	err := hub.SendToPlayerWithCorrelation("nonexistent", TypeLobbyUpdated, "corr-123", map[string]string{"test": "value"})
	if err != nil {
		t.Errorf("expected nil error for disconnected player, got %v", err)
	}
}

func TestHub_SendErrorToPlayer_NotConnected(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Should return nil (not an error) when player is not connected
	err := hub.SendErrorToPlayer("nonexistent", ErrCodeInternalError, "test error", "corr-123")
	if err != nil {
		t.Errorf("expected nil error for disconnected player, got %v", err)
	}
}

func TestHub_DisconnectPlayer_NotConnected(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Should not panic when player is not connected
	hub.DisconnectPlayer("nonexistent")
}

func TestHub_BroadcastToLobby_EmptyLobby(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Should not error or panic when lobby is empty
	err := hub.BroadcastToLobby("NONEXIST", TypeLobbyUpdated, map[string]string{"test": "value"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
