package websocket

import (
	"encoding/json"
	"time"
)

// Protocol version
const ProtocolVersion = 1

// MessageType represents the type of WebSocket message
type MessageType string

// Client -> Server message types
const (
	// Connection & Authentication
	TypeAuthenticate     MessageType = "authenticate"
	TypeHeartbeat        MessageType = "heartbeat"

	// Lobby Lifecycle
	TypeRequestLobbyState MessageType = "request_lobby_state"
	TypeSetReady          MessageType = "set_ready"

	// Battle Lifecycle
	TypeSubmitAction     MessageType = "submit_action"
	TypeRequestGameState MessageType = "request_game_state"

	// Post-Battle
	TypeRequestRematch MessageType = "request_rematch"
	TypeLeaveGame      MessageType = "leave_game"
)

// Server -> Client message types
const (
	// Connection & Authentication
	TypeAuthenticated MessageType = "authenticated"
	TypeHeartbeatAck  MessageType = "heartbeat_ack"

	// Lobby Lifecycle
	TypeLobbyUpdated  MessageType = "lobby_updated"
	TypeGameStarting  MessageType = "game_starting"
	TypeGameStarted   MessageType = "game_started"

	// Battle Lifecycle
	TypeGameState          MessageType = "game_state"
	TypeActionAcknowledged MessageType = "action_acknowledged"
	TypeTurnResult         MessageType = "turn_result"
	TypeSwitchRequired     MessageType = "switch_required"
	TypeGameEnded          MessageType = "game_ended"

	// Rematch Flow
	TypeRematchRequested MessageType = "rematch_requested"
	TypeRematchStarting  MessageType = "rematch_starting"

	// Errors
	TypeError            MessageType = "error"
	TypeDisconnectWarning MessageType = "disconnect_warning"
)

// Envelope is the standard message wrapper for all WebSocket messages
type Envelope struct {
	Type          MessageType     `json:"type"`
	Version       int             `json:"version"`
	Timestamp     int64           `json:"timestamp"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	Seq           int64           `json:"seq,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

// NewEnvelope creates a new envelope with current timestamp and protocol version
func NewEnvelope(msgType MessageType, payload interface{}) (*Envelope, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Envelope{
		Type:      msgType,
		Version:   ProtocolVersion,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payloadBytes,
	}, nil
}

// NewEnvelopeWithSeq creates a new envelope with sequence number
func NewEnvelopeWithSeq(msgType MessageType, seq int64, payload interface{}) (*Envelope, error) {
	env, err := NewEnvelope(msgType, payload)
	if err != nil {
		return nil, err
	}
	env.Seq = seq
	return env, nil
}

// WithCorrelationID adds a correlation ID to the envelope
func (e *Envelope) WithCorrelationID(id string) *Envelope {
	e.CorrelationID = id
	return e
}

// ParsePayload unmarshals the payload into the provided struct
func (e *Envelope) ParsePayload(v interface{}) error {
	return json.Unmarshal(e.Payload, v)
}

// ========================================
// Client -> Server Payloads
// ========================================

// AuthenticatePayload is sent by clients to establish identity
type AuthenticatePayload struct {
	PlayerID       string `json:"player_id"`
	SessionToken   string `json:"session_token"`
	LobbyCode      string `json:"lobby_code"`
	ReconnectToken string `json:"reconnect_token,omitempty"`
	LastSeq        int64  `json:"last_seq,omitempty"`
}

// HeartbeatPayload is sent by clients to keep connection alive
type HeartbeatPayload struct{}

// RequestLobbyStatePayload is sent to get current lobby state
type RequestLobbyStatePayload struct{}

// SetReadyPayload is sent to signal ready status
type SetReadyPayload struct {
	Ready bool `json:"ready"`
}

// ActionType represents the type of battle action
type ActionType string

const (
	ActionTypeAttack  ActionType = "attack"
	ActionTypeSwitch  ActionType = "switch"
	ActionTypeItem    ActionType = "item"
	ActionTypeForfeit ActionType = "forfeit"
)

// SubmitActionPayload is sent during battle
type SubmitActionPayload struct {
	TurnNumber int             `json:"turn_number"`
	ActionType ActionType      `json:"action_type"`
	ActionData json.RawMessage `json:"action_data"`
}

// AttackActionData contains data for an attack action
type AttackActionData struct {
	MoveID     string `json:"move_id"`
	TargetSlot int    `json:"target_slot"`
}

// SwitchActionData contains data for a switch action
type SwitchActionData struct {
	CreatureSlot int `json:"creature_slot"`
}

// ItemActionData contains data for an item action
type ItemActionData struct {
	ItemID     string `json:"item_id"`
	TargetSlot int    `json:"target_slot"`
}

// RequestGameStatePayload is sent to request full game snapshot
type RequestGameStatePayload struct {
	IncludeHistory bool `json:"include_history"`
}

// RequestRematchPayload is sent after game ends
type RequestRematchPayload struct{}

// LeaveGamePayload is sent to exit game/lobby
type LeaveGamePayload struct{}

// ========================================
// Server -> Client Payloads
// ========================================

// AuthenticatedPayload confirms authentication
type AuthenticatedPayload struct {
	PlayerID         string `json:"player_id"`
	ReconnectToken   string `json:"reconnect_token"`
	SessionExpiresAt int64  `json:"session_expires_at"`
}

// HeartbeatAckPayload acknowledges heartbeat
type HeartbeatAckPayload struct {
	ServerTime int64 `json:"server_time"`
}

// LobbyEvent represents types of lobby updates
type LobbyEvent string

const (
	LobbyEventPlayerJoined      LobbyEvent = "player_joined"
	LobbyEventPlayerLeft        LobbyEvent = "player_left"
	LobbyEventPlayerReadyChanged LobbyEvent = "player_ready_changed"
	LobbyEventHostChanged       LobbyEvent = "host_changed"
	LobbyEventStateChanged      LobbyEvent = "state_changed"
)

// LobbyPlayerInfo represents a player in the lobby
type LobbyPlayerInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsHost   bool   `json:"is_host"`
	IsReady  bool   `json:"is_ready"`
}

// LobbyInfo represents the lobby state
type LobbyInfo struct {
	Code    string            `json:"code"`
	State   string            `json:"state"`
	Players []LobbyPlayerInfo `json:"players"`
}

// LobbyUpdatedPayload notifies of lobby state changes
type LobbyUpdatedPayload struct {
	Lobby     LobbyInfo       `json:"lobby"`
	Event     LobbyEvent      `json:"event"`
	EventData json.RawMessage `json:"event_data,omitempty"`
}

// PlayerJoinedEventData is event data for player_joined
type PlayerJoinedEventData struct {
	PlayerID string `json:"player_id"`
	Username string `json:"username"`
}

// PlayerLeftEventData is event data for player_left
type PlayerLeftEventData struct {
	PlayerID string `json:"player_id"`
}

// PlayerReadyChangedEventData is event data for player_ready_changed
type PlayerReadyChangedEventData struct {
	PlayerID string `json:"player_id"`
	Ready    bool   `json:"ready"`
}

// HostChangedEventData is event data for host_changed
type HostChangedEventData struct {
	NewHostID string `json:"new_host_id"`
}

// StateChangedEventData is event data for state_changed
type StateChangedEventData struct {
	OldState string `json:"old_state"`
	NewState string `json:"new_state"`
}

// GameStartingPayload notifies that game countdown begins
type GameStartingPayload struct {
	StartsAt     int64 `json:"starts_at"`
	CountdownSec int   `json:"countdown_sec"`
}

// GameStartedPayload notifies that the game has started
type GameStartedPayload struct {
	GameID string `json:"game_id,omitempty"`
}

// CreatureInfo represents a creature in battle
type CreatureInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CurrentHP   int    `json:"current_hp"`
	MaxHP       int    `json:"max_hp"`
	Status      string `json:"status,omitempty"`
	IsActive    bool   `json:"is_active"`
}

// MoveInfo represents a move (only sent for player's own creatures)
type MoveInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	PP       int    `json:"pp"`
	MaxPP    int    `json:"max_pp"`
	Power    int    `json:"power,omitempty"`
	Accuracy int    `json:"accuracy,omitempty"`
}

// DetailedCreatureInfo includes full details (for player's own team)
type DetailedCreatureInfo struct {
	CreatureInfo
	Moves []MoveInfo `json:"moves,omitempty"`
}

// PlayerBattleState represents a player's battle state
type PlayerBattleState struct {
	PlayerID     string                 `json:"player_id"`
	Username     string                 `json:"username"`
	Team         []DetailedCreatureInfo `json:"team,omitempty"`         // Only for own team
	ActiveSlot   int                    `json:"active_slot"`
	BenchCount   int                    `json:"bench_count,omitempty"` // For opponent
	ActiveHP     int                    `json:"active_hp,omitempty"`   // For opponent's active
	ActiveMaxHP  int                    `json:"active_max_hp,omitempty"`
	ActiveStatus string                 `json:"active_status,omitempty"`
}

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseActionSelection GamePhase = "action_selection"
	GamePhaseTurnResolution  GamePhase = "turn_resolution"
	GamePhaseEnded           GamePhase = "ended"
)

// GameStatePayload contains full game snapshot
type GameStatePayload struct {
	TurnNumber    int               `json:"turn_number"`
	Phase         GamePhase         `json:"phase"`
	PlayerState   PlayerBattleState `json:"player_state"`
	OpponentState PlayerBattleState `json:"opponent_state"`
	TurnTimer     *TurnTimerInfo    `json:"turn_timer,omitempty"`
}

// TurnTimerInfo contains timer information
type TurnTimerInfo struct {
	ExpiresAt int64 `json:"expires_at"`
	Duration  int   `json:"duration_sec"`
}

// ActionAcknowledgedPayload confirms action received
type ActionAcknowledgedPayload struct {
	TurnNumber int `json:"turn_number"`
}

// TurnEventType represents types of turn events
type TurnEventType string

const (
	TurnEventMoveUsed        TurnEventType = "move_used"
	TurnEventDamageDealt     TurnEventType = "damage_dealt"
	TurnEventStatusApplied   TurnEventType = "status_applied"
	TurnEventCreatureFainted TurnEventType = "creature_fainted"
	TurnEventCreatureSwitched TurnEventType = "creature_switched"
	TurnEventStatChanged     TurnEventType = "stat_changed"
	TurnEventMoveFailed      TurnEventType = "move_failed"
	TurnEventActionTimeout   TurnEventType = "action_timeout"
)

// TurnEvent represents a single event in turn resolution
type TurnEvent struct {
	Order int             `json:"order"`
	Type  TurnEventType   `json:"type"`
	Actor string          `json:"actor,omitempty"`
	Data  json.RawMessage `json:"data"`
}

// TurnResultPayload contains turn resolution with events
type TurnResultPayload struct {
	TurnNumber     int              `json:"turn_number"`
	Events         []TurnEvent      `json:"events"`
	ResultingState GameStatePayload `json:"resulting_state"`
}

// MoveUsedEventData for move_used event
type MoveUsedEventData struct {
	MoveID string `json:"move_id"`
}

// DamageDealtEventData for damage_dealt event
type DamageDealtEventData struct {
	Target        string `json:"target"`
	Damage        int    `json:"damage"`
	Effectiveness string `json:"effectiveness"` // super_effective, not_very_effective, normal, no_effect
	Critical      bool   `json:"critical,omitempty"`
}

// StatusAppliedEventData for status_applied event
type StatusAppliedEventData struct {
	Target string `json:"target"`
	Status string `json:"status"`
}

// CreatureFaintedEventData for creature_fainted event
type CreatureFaintedEventData struct {
	CreatureID string `json:"creature_id"`
	Owner      string `json:"owner"`
}

// CreatureSwitchedEventData for creature_switched event
type CreatureSwitchedEventData struct {
	FromSlot int `json:"from_slot"`
	ToSlot   int `json:"to_slot"`
}

// StatChangedEventData for stat_changed event
type StatChangedEventData struct {
	Target string `json:"target"`
	Stat   string `json:"stat"`
	Stages int    `json:"stages"` // positive or negative
}

// MoveFailedEventData for move_failed event
type MoveFailedEventData struct {
	MoveID string `json:"move_id"`
	Reason string `json:"reason"`
}

// SwitchRequiredPayload prompts forced switch
type SwitchRequiredPayload struct {
	Reason           string `json:"reason"` // fainted, move_effect
	AvailableSlots   []int  `json:"available_slots"`
	TimeoutAt        int64  `json:"timeout_at"`
}

// GameEndReason represents why the game ended
type GameEndReason string

const (
	GameEndReasonVictory            GameEndReason = "victory"
	GameEndReasonForfeit            GameEndReason = "forfeit"
	GameEndReasonOpponentDisconnect GameEndReason = "opponent_disconnect"
	GameEndReasonTimeout            GameEndReason = "timeout"
)

// GameEndedPayload announces game conclusion
type GameEndedPayload struct {
	WinnerID    string            `json:"winner_id"`
	LoserID     string            `json:"loser_id"`
	Reason      GameEndReason     `json:"reason"`
	FinalState  *GameStatePayload `json:"final_state,omitempty"`
}

// RematchRequestedPayload notifies of rematch request
type RematchRequestedPayload struct {
	PlayerID string `json:"player_id"`
}

// RematchStartingPayload announces rematch countdown
type RematchStartingPayload struct {
	StartsAt     int64 `json:"starts_at"`
	CountdownSec int   `json:"countdown_sec"`
}

// DisconnectWarningPayload warns of impending disconnect
type DisconnectWarningPayload struct {
	Reason   string `json:"reason"`
	TimeoutAt int64 `json:"timeout_at"`
}
