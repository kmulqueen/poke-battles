package websocket

import (
	"sync"
)

// Hub maintains the set of active connections and broadcasts messages to lobbies
type Hub struct {
	mu sync.RWMutex

	// All active connections indexed by connection pointer
	connections map[*Connection]bool

	// Connections grouped by lobby code
	lobbies map[string]map[*Connection]bool

	// Player ID to connection mapping (for targeted messages)
	players map[string]*Connection

	// Channels for connection lifecycle
	register   chan *Connection
	unregister chan *Connection

	// Stop channel for graceful shutdown
	stop chan struct{}

	// Callback invoked when an authenticated player disconnects
	onDisconnect func(playerID, lobbyCode string)
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		connections: make(map[*Connection]bool),
		lobbies:     make(map[string]map[*Connection]bool),
		players:     make(map[string]*Connection),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		stop:        make(chan struct{}),
	}
}

// SetOnDisconnect sets the callback invoked when an authenticated player disconnects
func (h *Hub) SetOnDisconnect(callback func(playerID, lobbyCode string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onDisconnect = callback
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case <-h.stop:
			return
		case conn := <-h.register:
			h.handleRegister(conn)
		case conn := <-h.unregister:
			h.handleUnregister(conn)
		}
	}
}

// Stop gracefully shuts down the hub's main loop
func (h *Hub) Stop() {
	close(h.stop)
}

// Register adds a connection to the hub
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// Unregister removes a connection from the hub
func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

func (h *Hub) handleRegister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[conn] = true
}

func (h *Hub) handleUnregister(conn *Connection) {
	h.mu.Lock()

	if _, ok := h.connections[conn]; !ok {
		h.mu.Unlock()
		return
	}

	delete(h.connections, conn)

	// Remove from lobby
	lobbyCode := conn.LobbyCode()
	if lobbyCode != "" {
		if lobby, ok := h.lobbies[lobbyCode]; ok {
			delete(lobby, conn)
			if len(lobby) == 0 {
				delete(h.lobbies, lobbyCode)
			}
		}
	}

	// Remove from players map
	playerID := conn.PlayerID()
	if playerID != "" {
		if h.players[playerID] == conn {
			delete(h.players, playerID)
		}
	}

	// Capture callback before releasing lock
	callback := h.onDisconnect
	h.mu.Unlock()

	// Invoke callback outside lock to prevent deadlock
	if callback != nil && playerID != "" && lobbyCode != "" {
		callback(playerID, lobbyCode)
	}

	conn.Close()
}

// AssociateWithLobby associates a connection with a lobby after authentication
func (h *Hub) AssociateWithLobby(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	lobbyCode := conn.LobbyCode()
	playerID := conn.PlayerID()

	if lobbyCode == "" || playerID == "" {
		return
	}

	// Add to lobby map
	if _, ok := h.lobbies[lobbyCode]; !ok {
		h.lobbies[lobbyCode] = make(map[*Connection]bool)
	}
	h.lobbies[lobbyCode][conn] = true

	// Add to players map
	h.players[playerID] = conn
}

// GetConnectionByPlayerID returns the connection for a player
func (h *Hub) GetConnectionByPlayerID(playerID string) *Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.players[playerID]
}

// GetLobbyConnections returns all connections in a lobby
func (h *Hub) GetLobbyConnections(lobbyCode string) []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	lobby, ok := h.lobbies[lobbyCode]
	if !ok {
		return nil
	}

	conns := make([]*Connection, 0, len(lobby))
	for conn := range lobby {
		conns = append(conns, conn)
	}
	return conns
}

// BroadcastToLobby sends a message to all connections in a lobby
func (h *Hub) BroadcastToLobby(lobbyCode string, msgType MessageType, payload interface{}) error {
	conns := h.GetLobbyConnections(lobbyCode)
	if len(conns) == 0 {
		return nil
	}

	// Each connection must receive its own sequence number.
	// Do not optimize by reusing a single marshaled message.
	for _, conn := range conns {
		if conn.State() == ConnectionStateActive {
			conn.SendMessage(msgType, payload)
		}
	}

	return nil
}

// BroadcastToLobbyExcept sends a message to all connections in a lobby except one
func (h *Hub) BroadcastToLobbyExcept(lobbyCode string, exceptPlayerID string, msgType MessageType, payload interface{}) error {
	conns := h.GetLobbyConnections(lobbyCode)
	if len(conns) == 0 {
		return nil
	}

	// Each connection must receive its own sequence number.
	// Do not optimize by reusing a single marshaled message.
	for _, conn := range conns {
		if conn.State() == ConnectionStateActive && conn.PlayerID() != exceptPlayerID {
			conn.SendMessage(msgType, payload)
		}
	}

	return nil
}

// SendToPlayer sends a message to a specific player
func (h *Hub) SendToPlayer(playerID string, msgType MessageType, payload interface{}) error {
	conn := h.GetConnectionByPlayerID(playerID)
	if conn == nil {
		return nil // Player not connected
	}
	return conn.SendMessage(msgType, payload)
}

// SendToPlayerWithCorrelation sends a message to a specific player with correlation ID
func (h *Hub) SendToPlayerWithCorrelation(playerID string, msgType MessageType, correlationID string, payload interface{}) error {
	conn := h.GetConnectionByPlayerID(playerID)
	if conn == nil {
		return nil
	}
	return conn.SendMessageWithCorrelation(msgType, correlationID, payload)
}

// SendErrorToPlayer sends an error to a specific player
func (h *Hub) SendErrorToPlayer(playerID string, code ErrorCode, message string, correlationID string) error {
	conn := h.GetConnectionByPlayerID(playerID)
	if conn == nil {
		return nil
	}
	return conn.SendError(code, message, correlationID)
}

// ConnectionCount returns the total number of connections
func (h *Hub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// LobbyConnectionCount returns the number of connections in a lobby
func (h *Hub) LobbyConnectionCount(lobbyCode string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if lobby, ok := h.lobbies[lobbyCode]; ok {
		return len(lobby)
	}
	return 0
}

// IsPlayerConnected checks if a player is connected
func (h *Hub) IsPlayerConnected(playerID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.players[playerID]
	return ok
}

// DisconnectPlayer forcefully disconnects a player
func (h *Hub) DisconnectPlayer(playerID string) {
	conn := h.GetConnectionByPlayerID(playerID)
	if conn != nil {
		h.Unregister(conn)
	}
}
