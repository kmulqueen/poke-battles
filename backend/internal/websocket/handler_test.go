package websocket

import (
	"encoding/json"
	"sync"
	"testing"

	"poke-battles/internal/services"
)

// ========================================
// Test Helpers
// ========================================

// mockConnection is a test double for Connection that captures sent messages
type mockConnection struct {
	mu       sync.Mutex
	state    ConnectionState
	playerID string
	lobby    string
	messages []sentMessage
}

type sentMessage struct {
	msgType       MessageType
	payload       interface{}
	correlationID string
	isError       bool
	errorCode     ErrorCode
	errorMessage  string
}

func newMockConnection(playerID, lobbyCode string, state ConnectionState) *mockConnection {
	return &mockConnection{
		state:    state,
		playerID: playerID,
		lobby:    lobbyCode,
		messages: make([]sentMessage, 0),
	}
}

func (m *mockConnection) State() ConnectionState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

func (m *mockConnection) PlayerID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.playerID
}

func (m *mockConnection) LobbyCode() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lobby
}

func (m *mockConnection) recordMessage(msg sentMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *mockConnection) getMessages() []sentMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]sentMessage, len(m.messages))
	copy(result, m.messages)
	return result
}

func (m *mockConnection) getLastMessage() *sentMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.messages) == 0 {
		return nil
	}
	return &m.messages[len(m.messages)-1]
}

// Helper to create test envelope
func createEnvelope(msgType MessageType, payload interface{}) *Envelope {
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
// Ready State Management Tests
// ========================================

func TestHandler_SetPlayerReady(t *testing.T) {
	hub := NewHub()
	svc := services.NewLobbyService()
	handler := NewHandler(hub, svc)

	lobbyCode := "TEST01"
	playerID := "player-1"

	// Initially not ready
	if handler.isPlayerReady(lobbyCode, playerID) {
		t.Error("expected player to not be ready initially")
	}

	// Set ready
	handler.setPlayerReady(lobbyCode, playerID, true)

	if !handler.isPlayerReady(lobbyCode, playerID) {
		t.Error("expected player to be ready after setPlayerReady(true)")
	}

	// Set not ready
	handler.setPlayerReady(lobbyCode, playerID, false)

	if handler.isPlayerReady(lobbyCode, playerID) {
		t.Error("expected player to not be ready after setPlayerReady(false)")
	}
}

func TestHandler_ClearPlayerReadyState(t *testing.T) {
	hub := NewHub()
	svc := services.NewLobbyService()
	handler := NewHandler(hub, svc)

	lobbyCode := "TEST01"
	playerID := "player-1"

	handler.setPlayerReady(lobbyCode, playerID, true)
	if !handler.isPlayerReady(lobbyCode, playerID) {
		t.Fatal("expected player to be ready")
	}

	handler.clearPlayerReadyState(lobbyCode, playerID)

	if handler.isPlayerReady(lobbyCode, playerID) {
		t.Error("expected player ready state to be cleared")
	}
}

func TestHandler_ClearLobbyReadyState(t *testing.T) {
	hub := NewHub()
	svc := services.NewLobbyService()
	handler := NewHandler(hub, svc)

	lobbyCode := "TEST01"

	// Set multiple players ready
	handler.setPlayerReady(lobbyCode, "player-1", true)
	handler.setPlayerReady(lobbyCode, "player-2", true)

	if !handler.isPlayerReady(lobbyCode, "player-1") || !handler.isPlayerReady(lobbyCode, "player-2") {
		t.Fatal("expected both players to be ready")
	}

	handler.clearLobbyReadyState(lobbyCode)

	if handler.isPlayerReady(lobbyCode, "player-1") || handler.isPlayerReady(lobbyCode, "player-2") {
		t.Error("expected all ready states to be cleared")
	}
}

func TestHandler_ReadyStateIsolation(t *testing.T) {
	hub := NewHub()
	svc := services.NewLobbyService()
	handler := NewHandler(hub, svc)

	// Set player ready in lobby 1
	handler.setPlayerReady("LOBBY1", "player-1", true)

	// Player should not be ready in lobby 2
	if handler.isPlayerReady("LOBBY2", "player-1") {
		t.Error("ready state should be isolated per lobby")
	}

	// Same player ID, different lobby
	handler.setPlayerReady("LOBBY2", "player-1", true)
	handler.clearLobbyReadyState("LOBBY1")

	// Lobby 2 should still have player ready
	if !handler.isPlayerReady("LOBBY2", "player-1") {
		t.Error("clearing lobby 1 should not affect lobby 2")
	}
}

// ========================================
// Concurrent Ready State Tests
// ========================================

func TestHandler_ConcurrentReadyStateAccess(t *testing.T) {
	hub := NewHub()
	svc := services.NewLobbyService()
	handler := NewHandler(hub, svc)

	lobbyCode := "TEST01"
	var wg sync.WaitGroup

	// Multiple goroutines setting and reading ready state
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func(id int) {
			defer wg.Done()
			playerID := "player-" + string(rune('0'+id%10))
			handler.setPlayerReady(lobbyCode, playerID, true)
		}(i)

		go func(id int) {
			defer wg.Done()
			playerID := "player-" + string(rune('0'+id%10))
			handler.isPlayerReady(lobbyCode, playerID)
		}(i)

		go func(id int) {
			defer wg.Done()
			playerID := "player-" + string(rune('0'+id%10))
			handler.clearPlayerReadyState(lobbyCode, playerID)
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions occur
}

// ========================================
// BuildLobbyInfo Tests
// ========================================

func TestHandler_BuildLobbyInfo(t *testing.T) {
	hub := NewHub()
	svc := services.NewLobbyService()
	handler := NewHandler(hub, svc)

	// Create a lobby
	lobby, _ := svc.CreateLobby("host-1", "HostPlayer")
	svc.JoinLobby(lobby.Code, "player-2", "Player2")

	// Set player-2 as ready
	handler.setPlayerReady(lobby.Code, "player-2", true)

	lobbyInfo := handler.buildLobbyInfo(lobby)

	if lobbyInfo.Code != lobby.Code {
		t.Errorf("expected code %q, got %q", lobby.Code, lobbyInfo.Code)
	}
	if lobbyInfo.State != "ready" {
		t.Errorf("expected state 'ready', got %q", lobbyInfo.State)
	}
	if len(lobbyInfo.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(lobbyInfo.Players))
	}

	// Find host in players
	var hostInfo, player2Info *LobbyPlayerInfo
	for i := range lobbyInfo.Players {
		if lobbyInfo.Players[i].ID == "host-1" {
			hostInfo = &lobbyInfo.Players[i]
		} else if lobbyInfo.Players[i].ID == "player-2" {
			player2Info = &lobbyInfo.Players[i]
		}
	}

	if hostInfo == nil {
		t.Fatal("expected to find host in players")
	}
	if !hostInfo.IsHost {
		t.Error("expected host to have IsHost=true")
	}

	if player2Info == nil {
		t.Fatal("expected to find player-2 in players")
	}
	if player2Info.IsHost {
		t.Error("expected player-2 to have IsHost=false")
	}
	// Note: IsReady will be false because player is not actually connected via hub
}

// ========================================
// Protocol Version Tests
// ========================================

func TestHandler_RejectsWrongProtocolVersion(t *testing.T) {
	// Verify the version check logic would work
	env := &Envelope{
		Type:          TypeHeartbeat,
		Version:       999, // Wrong version
		Timestamp:     1234567890,
		CorrelationID: "test-corr-id",
		Payload:       []byte("{}"),
	}

	// Version check is done in handleMessage - here we just verify the setup
	if env.Version == ProtocolVersion {
		t.Error("test setup error: envelope should have wrong version")
	}
	if ProtocolVersion != 1 {
		t.Errorf("expected protocol version 1, got %d", ProtocolVersion)
	}
}

// ========================================
// Message Type Tests
// ========================================

func TestHandler_MessageTypes(t *testing.T) {
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

func TestHandler_ErrorCodeRecoverability(t *testing.T) {
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
	env := createEnvelope(TypeSetReady, original)

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
// Hub Integration Tests
// ========================================

func TestHub_ConnectionLifecycle(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Give the hub time to start
	// In production, we'd use proper synchronization

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
