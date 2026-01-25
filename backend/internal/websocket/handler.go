package websocket

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"poke-battles/internal/game"
	"poke-battles/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Configure allowed origins for production
		return true
	},
}

// Handler handles WebSocket connections and messages
type Handler struct {
	hub          *Hub
	lobbyService services.LobbyService
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, lobbyService services.LobbyService) *Handler {
	return &Handler{
		hub:          hub,
		lobbyService: lobbyService,
	}
}

// HandleConnection handles a new WebSocket connection
func (h *Handler) HandleConnection(c *gin.Context) {
	lobbyCode := c.Param("code")
	if lobbyCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lobby code required"})
		return
	}

	// Verify lobby exists before upgrading
	_, err := h.lobbyService.GetLobby(lobbyCode)
	if err != nil {
		if errors.Is(err, services.ErrLobbyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lobby not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return // Upgrade already writes error response
	}

	// Create connection and register with hub
	conn := NewConnection(wsConn, h.hub)
	h.hub.Register(conn)

	// Start read/write pumps
	go conn.WritePump()
	conn.ReadPump(h.handleMessage)
}

// handleMessage routes incoming messages to appropriate handlers
func (h *Handler) handleMessage(conn *Connection, env *Envelope) {
	// Version check
	if env.Version != ProtocolVersion {
		conn.SendError(ErrCodeVersionMismatch, "Protocol version not supported", env.CorrelationID)
		return
	}

	// Route based on message type
	switch env.Type {
	// Connection & Authentication
	case TypeAuthenticate:
		h.handleAuthenticate(conn, env)
	case TypeHeartbeat:
		h.handleHeartbeat(conn, env)

	// Lobby Lifecycle
	case TypeRequestLobbyState:
		h.handleRequestLobbyState(conn, env)
	case TypeSetReady:
		h.handleSetReady(conn, env)

	// Battle Lifecycle (placeholders for future implementation)
	case TypeSubmitAction:
		h.handleSubmitAction(conn, env)
	case TypeRequestGameState:
		h.handleRequestGameState(conn, env)

	// Post-Battle
	case TypeRequestRematch:
		h.handleRequestRematch(conn, env)
	case TypeLeaveGame:
		h.handleLeaveGame(conn, env)

	default:
		conn.SendError(ErrCodeMalformedMessage, "Unknown message type", env.CorrelationID)
	}
}

// handleAuthenticate handles authentication requests
func (h *Handler) handleAuthenticate(conn *Connection, env *Envelope) {
	var payload AuthenticatePayload
	if err := env.ParsePayload(&payload); err != nil {
		conn.SendError(ErrCodeMalformedMessage, "Invalid authenticate payload", env.CorrelationID)
		return
	}

	// Validate required fields
	if payload.PlayerID == "" || payload.LobbyCode == "" {
		conn.SendError(ErrCodeAuthFailed, "player_id and lobby_code are required", env.CorrelationID)
		return
	}

	// Get lobby
	lobby, err := h.lobbyService.GetLobby(payload.LobbyCode)
	if err != nil {
		if errors.Is(err, services.ErrLobbyNotFound) {
			conn.SendError(ErrCodeLobbyNotFound, "Lobby not found", env.CorrelationID)
			return
		}
		conn.SendError(ErrCodeInternalError, "Internal error", env.CorrelationID)
		return
	}

	// Verify player is in lobby
	if !lobby.HasPlayer(payload.PlayerID) {
		conn.SendError(ErrCodePlayerNotInLobby, "Player not in lobby", env.CorrelationID)
		return
	}

	// Verify lobby state allows connection
	state := lobby.GetState()
	if state != game.LobbyStateWaiting && state != game.LobbyStateReady && state != game.LobbyStateActive {
		conn.SendError(ErrCodeInvalidState, "Lobby not in valid state for connection", env.CorrelationID)
		return
	}

	// TODO: Validate session_token against auth service
	// For now, we trust the player_id if they're in the lobby

	// Handle reconnection if token provided
	if payload.ReconnectToken != "" {
		existingConn := h.hub.GetConnectionByPlayerID(payload.PlayerID)
		if existingConn != nil && existingConn.ValidateReconnectToken(payload.ReconnectToken) {
			// Valid reconnection - disconnect old connection
			h.hub.Unregister(existingConn)
		}
	}

	// Authenticate the connection
	if err := conn.Authenticate(payload.PlayerID, payload.LobbyCode); err != nil {
		conn.SendError(ErrCodeInternalError, "Authentication failed", env.CorrelationID)
		return
	}

	// Associate with lobby in hub
	h.hub.AssociateWithLobby(conn)

	// Send authenticated response
	authPayload := AuthenticatedPayload{
		PlayerID:         payload.PlayerID,
		ReconnectToken:   conn.GetReconnectToken(),
		SessionExpiresAt: conn.GetSessionExpiry().UnixMilli(),
	}
	conn.SendMessageWithCorrelation(TypeAuthenticated, env.CorrelationID, authPayload)

	// Send current lobby state
	h.sendLobbyState(conn, lobby)
}

// handleHeartbeat handles heartbeat messages
func (h *Handler) handleHeartbeat(conn *Connection, env *Envelope) {
	if conn.State() != ConnectionStateActive {
		conn.SendError(ErrCodeAuthRequired, "Authentication required", env.CorrelationID)
		return
	}

	conn.UpdateHeartbeat()

	ackPayload := HeartbeatAckPayload{
		ServerTime: time.Now().UnixMilli(),
	}
	conn.SendMessageWithCorrelation(TypeHeartbeatAck, env.CorrelationID, ackPayload)
}

// handleRequestLobbyState handles requests for current lobby state
func (h *Handler) handleRequestLobbyState(conn *Connection, env *Envelope) {
	if conn.State() != ConnectionStateActive {
		conn.SendError(ErrCodeAuthRequired, "Authentication required", env.CorrelationID)
		return
	}

	lobby, err := h.lobbyService.GetLobby(conn.LobbyCode())
	if err != nil {
		conn.SendError(ErrCodeLobbyNotFound, "Lobby not found", env.CorrelationID)
		return
	}

	h.sendLobbyState(conn, lobby)
}

// handleSetReady handles ready status changes
func (h *Handler) handleSetReady(conn *Connection, env *Envelope) {
	if conn.State() != ConnectionStateActive {
		conn.SendError(ErrCodeAuthRequired, "Authentication required", env.CorrelationID)
		return
	}

	var payload SetReadyPayload
	if err := env.ParsePayload(&payload); err != nil {
		conn.SendError(ErrCodeMalformedMessage, "Invalid set_ready payload", env.CorrelationID)
		return
	}

	// TODO: Implement ready state tracking when player ready state is added to domain
	// For now, just acknowledge the message
	// The lobby already tracks state (Waiting -> Ready based on player count)
	// This would be used for an explicit "ready to start" button

	lobby, err := h.lobbyService.GetLobby(conn.LobbyCode())
	if err != nil {
		conn.SendError(ErrCodeLobbyNotFound, "Lobby not found", env.CorrelationID)
		return
	}

	// Broadcast updated state to all players
	h.broadcastLobbyUpdate(lobby, LobbyEventPlayerReadyChanged, PlayerReadyChangedEventData{
		PlayerID: conn.PlayerID(),
		Ready:    payload.Ready,
	})
}

// handleSubmitAction handles battle action submissions
func (h *Handler) handleSubmitAction(conn *Connection, env *Envelope) {
	if conn.State() != ConnectionStateActive {
		conn.SendError(ErrCodeAuthRequired, "Authentication required", env.CorrelationID)
		return
	}

	// TODO: Implement when battle system is added
	// For now, return invalid state error
	conn.SendError(ErrCodeInvalidState, "No active battle", env.CorrelationID)
}

// handleRequestGameState handles requests for game state
func (h *Handler) handleRequestGameState(conn *Connection, env *Envelope) {
	if conn.State() != ConnectionStateActive {
		conn.SendError(ErrCodeAuthRequired, "Authentication required", env.CorrelationID)
		return
	}

	// TODO: Implement when battle system is added
	conn.SendError(ErrCodeInvalidState, "No active battle", env.CorrelationID)
}

// handleRequestRematch handles rematch requests
func (h *Handler) handleRequestRematch(conn *Connection, env *Envelope) {
	if conn.State() != ConnectionStateActive {
		conn.SendError(ErrCodeAuthRequired, "Authentication required", env.CorrelationID)
		return
	}

	// TODO: Implement when battle system is added
	conn.SendError(ErrCodeInvalidState, "No game to rematch", env.CorrelationID)
}

// handleLeaveGame handles leave game requests
func (h *Handler) handleLeaveGame(conn *Connection, env *Envelope) {
	if conn.State() != ConnectionStateActive {
		conn.SendError(ErrCodeAuthRequired, "Authentication required", env.CorrelationID)
		return
	}

	lobbyCode := conn.LobbyCode()
	playerID := conn.PlayerID()

	// Remove player from lobby
	err := h.lobbyService.LeaveLobby(lobbyCode, playerID)
	if err != nil {
		// Player may already be removed, that's okay
		if !errors.Is(err, game.ErrPlayerNotFound) && !errors.Is(err, services.ErrLobbyNotFound) {
			conn.SendError(ErrCodeInternalError, "Failed to leave lobby", env.CorrelationID)
			return
		}
	}

	// Notify remaining players
	lobby, err := h.lobbyService.GetLobby(lobbyCode)
	if err == nil {
		h.broadcastLobbyUpdate(lobby, LobbyEventPlayerLeft, PlayerLeftEventData{
			PlayerID: playerID,
		})
	}

	// Close connection
	h.hub.Unregister(conn)
}

// sendLobbyState sends the current lobby state to a connection
func (h *Handler) sendLobbyState(conn *Connection, lobby *game.Lobby) {
	lobbyInfo := h.buildLobbyInfo(lobby)
	payload := LobbyUpdatedPayload{
		Lobby: lobbyInfo,
		Event: LobbyEventStateChanged,
	}
	conn.SendMessage(TypeLobbyUpdated, payload)
}

// broadcastLobbyUpdate broadcasts a lobby update to all players in the lobby
func (h *Handler) broadcastLobbyUpdate(lobby *game.Lobby, event LobbyEvent, eventData interface{}) {
	lobbyInfo := h.buildLobbyInfo(lobby)
	payload := LobbyUpdatedPayload{
		Lobby: lobbyInfo,
		Event: event,
	}

	if eventData != nil {
		data, _ := lobbyInfo.MarshalEventData(eventData)
		payload.EventData = data
	}

	h.hub.BroadcastToLobby(lobby.Code, TypeLobbyUpdated, payload)
}

// buildLobbyInfo creates a LobbyInfo from a game.Lobby
func (h *Handler) buildLobbyInfo(lobby *game.Lobby) LobbyInfo {
	players := lobby.GetPlayers()
	hostID := lobby.GetHostID()

	playerInfos := make([]LobbyPlayerInfo, len(players))
	for i, p := range players {
		playerInfos[i] = LobbyPlayerInfo{
			ID:       p.ID,
			Username: p.Username,
			IsHost:   p.ID == hostID,
			IsReady:  false, // TODO: Track ready state per player
		}
	}

	return LobbyInfo{
		Code:    lobby.Code,
		State:   lobby.GetState().String(),
		Players: playerInfos,
	}
}

// MarshalEventData marshals event data to JSON
func (l *LobbyInfo) MarshalEventData(data interface{}) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	return json.Marshal(data)
}

// BroadcastPlayerJoined broadcasts a player joined event
func (h *Handler) BroadcastPlayerJoined(lobbyCode string, playerID, username string) {
	lobby, err := h.lobbyService.GetLobby(lobbyCode)
	if err != nil {
		return
	}
	h.broadcastLobbyUpdate(lobby, LobbyEventPlayerJoined, PlayerJoinedEventData{
		PlayerID: playerID,
		Username: username,
	})
}

// BroadcastPlayerLeft broadcasts a player left event
func (h *Handler) BroadcastPlayerLeft(lobbyCode string, playerID string) {
	lobby, err := h.lobbyService.GetLobby(lobbyCode)
	if err != nil {
		return
	}
	h.broadcastLobbyUpdate(lobby, LobbyEventPlayerLeft, PlayerLeftEventData{
		PlayerID: playerID,
	})
}

// BroadcastGameStarting broadcasts a game starting event
func (h *Handler) BroadcastGameStarting(lobbyCode string, countdownSec int) {
	startsAt := time.Now().Add(time.Duration(countdownSec) * time.Second).UnixMilli()
	payload := GameStartingPayload{
		StartsAt:     startsAt,
		CountdownSec: countdownSec,
	}
	h.hub.BroadcastToLobby(lobbyCode, TypeGameStarting, payload)
}
