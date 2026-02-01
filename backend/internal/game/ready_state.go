package game

import (
	"sync"
)

// ReadyTracker manages player ready state across lobbies.
// This is ephemeral state - not persisted to the domain model.
type ReadyTracker struct {
	mu    sync.RWMutex
	state map[string]map[string]bool // lobbyCode -> playerID -> ready
}

// NewReadyTracker creates a new ReadyTracker
func NewReadyTracker() *ReadyTracker {
	return &ReadyTracker{
		state: make(map[string]map[string]bool),
	}
}

// SetReady sets a player's ready state in a lobby
func (r *ReadyTracker) SetReady(lobbyCode, playerID string, ready bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.state[lobbyCode]; !ok {
		r.state[lobbyCode] = make(map[string]bool)
	}
	r.state[lobbyCode][playerID] = ready
}

// IsReady checks if a player has set ready in a lobby
func (r *ReadyTracker) IsReady(lobbyCode, playerID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if lobbyReady, ok := r.state[lobbyCode]; ok {
		return lobbyReady[playerID]
	}
	return false
}

// ClearPlayer removes a player's ready state from a lobby
func (r *ReadyTracker) ClearPlayer(lobbyCode, playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if lobbyReady, ok := r.state[lobbyCode]; ok {
		delete(lobbyReady, playerID)
		if len(lobbyReady) == 0 {
			delete(r.state, lobbyCode)
		}
	}
}

// ClearLobby removes all ready state for a lobby
func (r *ReadyTracker) ClearLobby(lobbyCode string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.state, lobbyCode)
}

// AllReady checks if all specified players are ready in a lobby
func (r *ReadyTracker) AllReady(lobbyCode string, playerIDs []string) bool {
	// Empty player list is vacuously true
	if len(playerIDs) == 0 {
		return true
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	lobbyReady, ok := r.state[lobbyCode]
	if !ok {
		return false
	}

	for _, playerID := range playerIDs {
		if !lobbyReady[playerID] {
			return false
		}
	}
	return true
}
