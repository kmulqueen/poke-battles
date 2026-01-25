package game

import (
	"errors"
	"sync"
	"time"
)

// Domain errors
var (
	ErrLobbyFull            = errors.New("lobby is full")
	ErrPlayerAlreadyJoined  = errors.New("player already in lobby")
	ErrPlayerNotFound       = errors.New("player not found in lobby")
	ErrInvalidStateForJoin  = errors.New("cannot join lobby in current state")
	ErrInvalidStateForStart = errors.New("cannot start lobby in current state")
	ErrNotEnoughPlayers     = errors.New("not enough players to start")
)

// LobbyState represents the current state of a lobby
type LobbyState int

const (
	LobbyStateWaiting LobbyState = iota // Waiting for players
	LobbyStateReady                     // Both players joined, ready to start
	LobbyStateActive                    // Game in progress
)

// String returns a human-readable representation of the lobby state
func (s LobbyState) String() string {
	switch s {
	case LobbyStateWaiting:
		return "waiting"
	case LobbyStateReady:
		return "ready"
	case LobbyStateActive:
		return "active"
	default:
		return "unknown"
	}
}

// Player represents a player in a lobby
type Player struct {
	ID       string
	Username string
}

// Lobby represents a game lobby
type Lobby struct {
	mu         sync.RWMutex
	Code       string
	State      LobbyState
	Players    []*Player
	HostID     string
	MaxPlayers int
	CreatedAt  time.Time
}

// NewLobby creates a new lobby with the given host as the first player
func NewLobby(code, hostID, hostUsername string) *Lobby {
	host := &Player{
		ID:       hostID,
		Username: hostUsername,
	}
	return &Lobby{
		Code:       code,
		State:      LobbyStateWaiting,
		Players:    []*Player{host},
		HostID:     hostID,
		MaxPlayers: 2,
		CreatedAt:  time.Now(),
	}
}

// AddPlayer adds a player to the lobby with validation
func (l *Lobby) AddPlayer(id, username string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check state - can only join while waiting
	if l.State != LobbyStateWaiting {
		return ErrInvalidStateForJoin
	}

	// Check if player already in lobby
	for _, p := range l.Players {
		if p.ID == id {
			return ErrPlayerAlreadyJoined
		}
	}

	// Check if lobby is full
	if len(l.Players) >= l.MaxPlayers {
		return ErrLobbyFull
	}

	// Add player
	l.Players = append(l.Players, &Player{
		ID:       id,
		Username: username,
	})

	// Transition to Ready if we now have max players
	if len(l.Players) == l.MaxPlayers {
		l.State = LobbyStateReady
	}

	return nil
}

// RemovePlayer removes a player from the lobby
func (l *Lobby) RemovePlayer(id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Find and remove the player
	found := false
	for i, p := range l.Players {
		if p.ID == id {
			l.Players = append(l.Players[:i], l.Players[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrPlayerNotFound
	}

	// If we were Ready and now have fewer players, go back to Waiting
	if l.State == LobbyStateReady && len(l.Players) < l.MaxPlayers {
		l.State = LobbyStateWaiting
	}

	// If host left and there are remaining players, assign new host
	if id == l.HostID && len(l.Players) > 0 {
		l.HostID = l.Players[0].ID
	}

	return nil
}

// GetState returns the current lobby state (thread-safe)
func (l *Lobby) GetState() LobbyState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.State
}

// CanStart returns true if the lobby can start a game
func (l *Lobby) CanStart() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.State == LobbyStateReady && len(l.Players) == l.MaxPlayers
}

// Start transitions the lobby from Ready to Active
func (l *Lobby) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != LobbyStateReady {
		return ErrInvalidStateForStart
	}

	if len(l.Players) < l.MaxPlayers {
		return ErrNotEnoughPlayers
	}

	l.State = LobbyStateActive
	return nil
}

// PlayerCount returns the number of players in the lobby (thread-safe)
func (l *Lobby) PlayerCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.Players)
}

// HasPlayer checks if a player is already in the lobby
func (l *Lobby) HasPlayer(id string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, p := range l.Players {
		if p.ID == id {
			return true
		}
	}
	return false
}

// IsHost checks if the given player ID is the host
func (l *Lobby) IsHost(id string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.HostID == id
}

// GetPlayers returns a copy of the players slice (thread-safe)
func (l *Lobby) GetPlayers() []*Player {
	l.mu.RLock()
	defer l.mu.RUnlock()
	players := make([]*Player, len(l.Players))
	for i, p := range l.Players {
		players[i] = &Player{
			ID:       p.ID,
			Username: p.Username,
		}
	}
	return players
}

// GetHostID returns the host player ID (thread-safe)
func (l *Lobby) GetHostID() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.HostID
}
