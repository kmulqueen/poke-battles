package websocket

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"time"

	"poke-battles/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ========================================
// Test Server
// ========================================

// TestServer wraps an httptest.Server with WebSocket infrastructure
type TestServer struct {
	Server       *httptest.Server
	Handler      *Handler
	Hub          *Hub
	LobbyService services.LobbyService

	mu       sync.Mutex
	shutdown bool
}

// NewTestServer creates a new test server with WebSocket support
func NewTestServer() *TestServer {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	lobbyService := services.NewLobbyService()
	handler := NewHandler(hub, lobbyService)

	router := gin.New()
	router.GET("/api/v1/ws/game/:code", handler.HandleConnection)

	server := httptest.NewServer(router)

	ts := &TestServer{
		Server:       server,
		Handler:      handler,
		Hub:          hub,
		LobbyService: lobbyService,
	}

	go hub.Run()

	return ts
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.shutdown {
		return
	}
	ts.shutdown = true

	ts.Hub.Stop()
	ts.Server.Close()
}

// WebSocketURL returns the WebSocket URL for a lobby
func (ts *TestServer) WebSocketURL(lobbyCode string) string {
	return "ws" + strings.TrimPrefix(ts.Server.URL, "http") + "/api/v1/ws/game/" + lobbyCode
}

// CreateLobby creates a lobby and returns its code
func (ts *TestServer) CreateLobby(hostID, username string) (string, error) {
	lobby, err := ts.LobbyService.CreateLobby(hostID, username)
	if err != nil {
		return "", err
	}
	return lobby.Code, nil
}

// JoinLobby adds a player to an existing lobby
func (ts *TestServer) JoinLobby(code, playerID, username string) error {
	_, err := ts.LobbyService.JoinLobby(code, playerID, username)
	return err
}

// WaitForPlayerConnected waits for a player to be connected
func (ts *TestServer) WaitForPlayerConnected(playerID string, timeout time.Duration) bool {
	return waitFor(func() bool {
		return ts.Hub.IsPlayerConnected(playerID)
	}, timeout)
}

// WaitForPlayerDisconnected waits for a player to be disconnected
func (ts *TestServer) WaitForPlayerDisconnected(playerID string, timeout time.Duration) bool {
	return waitFor(func() bool {
		return !ts.Hub.IsPlayerConnected(playerID)
	}, timeout)
}

// ========================================
// Test Client
// ========================================

// TestClient wraps a WebSocket connection for testing
type TestClient struct {
	conn      *websocket.Conn
	PlayerID  string
	LobbyCode string

	mu       sync.Mutex
	received chan *Envelope
	done     chan struct{}
	closed   bool
}

// NewTestClient creates a test client connected to the server
func NewTestClient(serverURL string) (*TestClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	tc := &TestClient{
		conn:     conn,
		received: make(chan *Envelope, 100),
		done:     make(chan struct{}),
	}

	go tc.readLoop()

	return tc, nil
}

// readLoop reads messages from the WebSocket and queues them
func (tc *TestClient) readLoop() {
	defer close(tc.done)

	for {
		_, message, err := tc.conn.ReadMessage()
		if err != nil {
			return
		}

		var env Envelope
		if err := json.Unmarshal(message, &env); err != nil {
			continue
		}

		select {
		case tc.received <- &env:
		default:
			// Buffer full, drop oldest
			select {
			case <-tc.received:
			default:
			}
			tc.received <- &env
		}
	}
}

// Close closes the client connection
func (tc *TestClient) Close() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.closed {
		return nil
	}
	tc.closed = true

	err := tc.conn.Close()
	<-tc.done
	return err
}

// Send sends an envelope to the server
func (tc *TestClient) Send(env *Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return tc.conn.WriteMessage(websocket.TextMessage, data)
}

// SendAuth sends an authentication message
func (tc *TestClient) SendAuth(playerID, lobbyCode string) error {
	tc.PlayerID = playerID
	tc.LobbyCode = lobbyCode

	payload := AuthenticatePayload{
		PlayerID:  playerID,
		LobbyCode: lobbyCode,
	}

	env, err := NewEnvelope(TypeAuthenticate, payload)
	if err != nil {
		return err
	}
	env.CorrelationID = "auth-" + playerID

	return tc.Send(env)
}

// SendReady sends a set_ready message
func (tc *TestClient) SendReady(ready bool) error {
	payload := SetReadyPayload{Ready: ready}
	env, err := NewEnvelope(TypeSetReady, payload)
	if err != nil {
		return err
	}
	env.CorrelationID = fmt.Sprintf("ready-%s-%v", tc.PlayerID, ready)
	return tc.Send(env)
}

// SendHeartbeat sends a heartbeat message
func (tc *TestClient) SendHeartbeat() error {
	env, err := NewEnvelope(TypeHeartbeat, HeartbeatPayload{})
	if err != nil {
		return err
	}
	env.CorrelationID = "heartbeat-" + tc.PlayerID
	return tc.Send(env)
}

// Receive waits for any message with timeout
func (tc *TestClient) Receive(timeout time.Duration) (*Envelope, error) {
	select {
	case env := <-tc.received:
		return env, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("receive timeout after %v", timeout)
	case <-tc.done:
		return nil, fmt.Errorf("connection closed")
	}
}

// ReceiveType waits for a specific message type with timeout
func (tc *TestClient) ReceiveType(msgType MessageType, timeout time.Duration) (*Envelope, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		select {
		case env := <-tc.received:
			if env.Type == msgType {
				return env, nil
			}
			// Not the type we want, continue waiting
		case <-time.After(remaining):
			return nil, fmt.Errorf("timeout waiting for %s after %v", msgType, timeout)
		case <-tc.done:
			return nil, fmt.Errorf("connection closed while waiting for %s", msgType)
		}
	}

	return nil, fmt.Errorf("timeout waiting for %s after %v", msgType, timeout)
}

// ExpectError waits for an error message with the specified code
func (tc *TestClient) ExpectError(code ErrorCode, timeout time.Duration) error {
	env, err := tc.ReceiveType(TypeError, timeout)
	if err != nil {
		return err
	}

	var errPayload ErrorPayload
	if err := env.ParsePayload(&errPayload); err != nil {
		return fmt.Errorf("failed to parse error payload: %w", err)
	}

	if errPayload.Code != code {
		return fmt.Errorf("expected error code %s, got %s: %s", code, errPayload.Code, errPayload.Message)
	}

	return nil
}

// Drain clears all pending messages from the receive buffer
func (tc *TestClient) Drain() {
	for {
		select {
		case <-tc.received:
		default:
			return
		}
	}
}

// PendingCount returns the number of pending messages
func (tc *TestClient) PendingCount() int {
	return len(tc.received)
}

// ========================================
// Synchronization Helpers
// ========================================

// waitFor polls a condition until it returns true or timeout expires
// Uses runtime.Gosched() instead of time.Sleep for efficiency
func waitFor(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		runtime.Gosched()
	}

	return condition()
}

// ========================================
// Assertion Helpers
// ========================================

// AssertAuthSuccess asserts successful authentication and returns the response
func (tc *TestClient) AssertAuthSuccess(timeout time.Duration) (*AuthenticatedPayload, error) {
	env, err := tc.ReceiveType(TypeAuthenticated, timeout)
	if err != nil {
		return nil, err
	}

	var payload AuthenticatedPayload
	if err := env.ParsePayload(&payload); err != nil {
		return nil, fmt.Errorf("failed to parse authenticated payload: %w", err)
	}

	if payload.PlayerID != tc.PlayerID {
		return nil, fmt.Errorf("expected player_id %s, got %s", tc.PlayerID, payload.PlayerID)
	}

	return &payload, nil
}

// AssertLobbyUpdated asserts a lobby_updated message and returns the payload
func (tc *TestClient) AssertLobbyUpdated(timeout time.Duration) (*LobbyUpdatedPayload, error) {
	env, err := tc.ReceiveType(TypeLobbyUpdated, timeout)
	if err != nil {
		return nil, err
	}

	var payload LobbyUpdatedPayload
	if err := env.ParsePayload(&payload); err != nil {
		return nil, fmt.Errorf("failed to parse lobby_updated payload: %w", err)
	}

	return &payload, nil
}
