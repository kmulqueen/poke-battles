package services

import (
	"errors"
	"fmt"
	"sync"

	"poke-battles/internal/game"
)

// Sentinel errors for error type checking with errors.Is()
var (
	ErrLobbyNotFound = errors.New("lobby not found")
	ErrNotHost       = errors.New("only host can start the game")
)

// LobbyService defines the interface for lobby operations
type LobbyService interface {
	CreateLobby(hostID, hostUsername string) (*game.Lobby, error)
	JoinLobby(code, playerID, playerUsername string) (*game.Lobby, error)
	LeaveLobby(code, playerID string) error
	GetLobby(code string) (*game.Lobby, error)
	StartGame(code, playerID string) error
	ListLobbies() ([]*game.Lobby, error)
}

// lobbyService implements LobbyService with in-memory storage
type lobbyService struct {
	mu      sync.RWMutex
	lobbies map[string]*game.Lobby
}

// NewLobbyService creates a new lobby service instance
func NewLobbyService() LobbyService {
	return &lobbyService{
		lobbies: make(map[string]*game.Lobby),
	}
}

// CreateLobby creates a new lobby with the given host
func (s *lobbyService) CreateLobby(hostID, hostUsername string) (*game.Lobby, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a unique room code
	var code string
	for {
		code = game.GenerateRoomCode()
		if _, exists := s.lobbies[code]; !exists {
			break
		}
	}

	lobby := game.NewLobby(code, hostID, hostUsername)
	s.lobbies[code] = lobby

	return lobby, nil
}

// JoinLobby adds a player to an existing lobby
func (s *lobbyService) JoinLobby(code, playerID, playerUsername string) (*game.Lobby, error) {
	s.mu.RLock()
	lobby, exists := s.lobbies[code]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("lobby %q: %w", code, ErrLobbyNotFound)
	}

	if err := lobby.AddPlayer(playerID, playerUsername); err != nil {
		return nil, fmt.Errorf("lobby %q, player %q: %w", code, playerID, err)
	}

	return lobby, nil
}

// LeaveLobby removes a player from a lobby and cleans up empty lobbies
func (s *lobbyService) LeaveLobby(code, playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lobby, exists := s.lobbies[code]
	if !exists {
		return fmt.Errorf("lobby %q: %w", code, ErrLobbyNotFound)
	}

	if err := lobby.RemovePlayer(playerID); err != nil {
		return fmt.Errorf("lobby %q, player %q: %w", code, playerID, err)
	}

	// Clean up empty lobbies
	if lobby.PlayerCount() == 0 {
		delete(s.lobbies, code)
	}

	return nil
}

// GetLobby retrieves a lobby by its code
func (s *lobbyService) GetLobby(code string) (*game.Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lobby, exists := s.lobbies[code]
	if !exists {
		return nil, fmt.Errorf("lobby %q: %w", code, ErrLobbyNotFound)
	}

	return lobby, nil
}

// ListLobbies retrieves a list of all lobbies
func (s *lobbyService) ListLobbies() ([]*game.Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lobbies := make([]*game.Lobby, 0, len(s.lobbies))
	for _, lobby := range s.lobbies {
		lobbies = append(lobbies, lobby)
	}
	return lobbies, nil
}

// StartGame starts the game for a lobby (host only)
func (s *lobbyService) StartGame(code, playerID string) error {
	s.mu.RLock()
	lobby, exists := s.lobbies[code]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("lobby %q: %w", code, ErrLobbyNotFound)
	}

	if !lobby.IsHost(playerID) {
		return fmt.Errorf("lobby %q, player %q: %w", code, playerID, ErrNotHost)
	}

	if err := lobby.Start(); err != nil {
		return fmt.Errorf("lobby %q: %w", code, err)
	}

	return nil
}
