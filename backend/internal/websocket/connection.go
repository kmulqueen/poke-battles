package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ConnectionState represents the state of a WebSocket connection
type ConnectionState int

const (
	// ConnectionStatePending - connection accepted but not authenticated
	ConnectionStatePending ConnectionState = iota
	// ConnectionStateActive - authenticated and ready
	ConnectionStateActive
	// ConnectionStateClosing - about to close
	ConnectionStateClosing
)

// Connection represents a single WebSocket connection
type Connection struct {
	mu sync.RWMutex

	// WebSocket connection
	conn *websocket.Conn

	// Connection state
	state ConnectionState

	// Player identification (set after authentication)
	playerID string
	lobbyCode string

	// Sequence tracking
	outboundSeq    int64 // Next sequence number for outbound messages
	lastReceivedSeq int64 // Last sequence number received from this client

	// Reconnection
	reconnectToken  string
	sessionExpiry   time.Time

	// Heartbeat tracking
	lastHeartbeat time.Time

	// Send channel for outbound messages
	send chan []byte

	// Hub reference for cleanup
	hub *Hub
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192

	// Size of send channel buffer
	sendBufferSize = 256

	// Session duration
	sessionDuration = 24 * time.Hour

	// Reconnect token duration
	reconnectTokenDuration = 5 * time.Minute
)

// NewConnection creates a new connection
func NewConnection(conn *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		conn:          conn,
		state:         ConnectionStatePending,
		outboundSeq:   0,
		lastHeartbeat: time.Now(),
		send:          make(chan []byte, sendBufferSize),
		hub:           hub,
	}
}

// State returns the current connection state
func (c *Connection) State() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// SetState updates the connection state
func (c *Connection) SetState(state ConnectionState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = state
}

// PlayerID returns the player ID
func (c *Connection) PlayerID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.playerID
}

// LobbyCode returns the lobby code
func (c *Connection) LobbyCode() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lobbyCode
}

// Authenticate sets the player credentials after successful authentication
func (c *Connection) Authenticate(playerID, lobbyCode string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	token, err := generateReconnectToken()
	if err != nil {
		return err
	}

	c.playerID = playerID
	c.lobbyCode = lobbyCode
	c.state = ConnectionStateActive
	c.reconnectToken = token
	c.sessionExpiry = time.Now().Add(sessionDuration)

	return nil
}

// GetReconnectToken returns the current reconnect token
func (c *Connection) GetReconnectToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reconnectToken
}

// GetSessionExpiry returns the session expiry time
func (c *Connection) GetSessionExpiry() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionExpiry
}

// ValidateReconnectToken validates a reconnect token
func (c *Connection) ValidateReconnectToken(token string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reconnectToken == token && time.Now().Before(c.sessionExpiry)
}

// RefreshReconnectToken generates a new reconnect token
func (c *Connection) RefreshReconnectToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	token, err := generateReconnectToken()
	if err != nil {
		return "", err
	}
	c.reconnectToken = token
	return token, nil
}

// NextSeq returns and increments the outbound sequence number
func (c *Connection) NextSeq() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.outboundSeq++
	return c.outboundSeq
}

// CurrentSeq returns the current outbound sequence number without incrementing
func (c *Connection) CurrentSeq() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.outboundSeq
}

// UpdateLastReceivedSeq updates the last received sequence number
func (c *Connection) UpdateLastReceivedSeq(seq int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if seq > c.lastReceivedSeq {
		c.lastReceivedSeq = seq
	}
}

// LastReceivedSeq returns the last received sequence number
func (c *Connection) LastReceivedSeq() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastReceivedSeq
}

// UpdateHeartbeat updates the last heartbeat time
func (c *Connection) UpdateHeartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastHeartbeat = time.Now()
}

// LastHeartbeat returns the last heartbeat time
func (c *Connection) LastHeartbeat() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastHeartbeat
}

// SendMessage sends a message to the client with proper envelope
func (c *Connection) SendMessage(msgType MessageType, payload interface{}) error {
	seq := c.NextSeq()
	env, err := NewEnvelopeWithSeq(msgType, seq, payload)
	if err != nil {
		return err
	}
	return c.SendEnvelope(env)
}

// SendMessageWithCorrelation sends a message with correlation ID
func (c *Connection) SendMessageWithCorrelation(msgType MessageType, correlationID string, payload interface{}) error {
	seq := c.NextSeq()
	env, err := NewEnvelopeWithSeq(msgType, seq, payload)
	if err != nil {
		return err
	}
	env.CorrelationID = correlationID
	return c.SendEnvelope(env)
}

// SendEnvelope sends a pre-built envelope
func (c *Connection) SendEnvelope(env *Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return c.SendRaw(data)
}

// SendRaw sends raw bytes to the client
func (c *Connection) SendRaw(data []byte) error {
	select {
	case c.send <- data:
		return nil
	default:
		// Channel full, connection is too slow
		return ErrSendBufferFull
	}
}

// SendError sends an error message
func (c *Connection) SendError(code ErrorCode, message string, correlationID string) error {
	payload := NewErrorPayload(code, message)
	if correlationID != "" {
		return c.SendMessageWithCorrelation(TypeError, correlationID, payload)
	}
	return c.SendMessage(TypeError, payload)
}

// SendErrorWithDetails sends an error message with details
func (c *Connection) SendErrorWithDetails(code ErrorCode, message string, details interface{}, correlationID string) error {
	payload, err := NewErrorPayloadWithDetails(code, message, details)
	if err != nil {
		// Fall back to simple error if details can't be serialized
		return c.SendError(code, message, correlationID)
	}
	if correlationID != "" {
		return c.SendMessageWithCorrelation(TypeError, correlationID, payload)
	}
	return c.SendMessage(TypeError, payload)
}

// Close closes the connection
func (c *Connection) Close() {
	c.mu.Lock()
	if c.state == ConnectionStateClosing {
		c.mu.Unlock()
		return
	}
	c.state = ConnectionStateClosing
	c.mu.Unlock()

	close(c.send)
	c.conn.Close()
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Connection) ReadPump(handler func(*Connection, *Envelope)) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close
			}
			break
		}

		var env Envelope
		if err := json.Unmarshal(message, &env); err != nil {
			c.SendError(ErrCodeMalformedMessage, "Could not parse message envelope", "")
			continue
		}

		// Track sequence number if provided
		if env.Seq > 0 {
			c.UpdateLastReceivedSeq(env.Seq)
		}

		handler(c, &env)
	}
}

// ErrSendBufferFull is returned when the send buffer is full
var ErrSendBufferFull = &SendBufferFullError{}

type SendBufferFullError struct{}

func (e *SendBufferFullError) Error() string {
	return "send buffer full"
}

// generateReconnectToken generates a secure random reconnect token
func generateReconnectToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
