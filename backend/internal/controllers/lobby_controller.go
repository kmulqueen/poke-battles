package controllers

import (
	"errors"
	"net/http"

	"poke-battles/internal/game"
	"poke-battles/internal/services"

	"github.com/gin-gonic/gin"
)

// Request types

type CreateLobbyRequest struct {
	PlayerID string `json:"player_id" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type JoinLobbyRequest struct {
	PlayerID string `json:"player_id" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type LeaveLobbyRequest struct {
	PlayerID string `json:"player_id" binding:"required"`
}

type StartGameRequest struct {
	PlayerID string `json:"player_id" binding:"required"`
}

// Response types

type PlayerResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type LobbyResponse struct {
	Code       string           `json:"code"`
	State      string           `json:"state"`
	Players    []PlayerResponse `json:"players"`
	HostID     string           `json:"host_id"`
	MaxPlayers int              `json:"max_players"`
}

type LobbyListResponse []LobbyResponse

// LobbyController handles HTTP requests for lobby operations
type LobbyController struct {
	lobbyService services.LobbyService
}

// NewLobbyController creates a new lobby controller
func NewLobbyController(ls services.LobbyService) *LobbyController {
	return &LobbyController{
		lobbyService: ls,
	}
}

// toLobbyResponse converts a domain Lobby to a response DTO
func toLobbyResponse(lobby *game.Lobby) LobbyResponse {
	players := lobby.GetPlayers()
	playerResponses := make([]PlayerResponse, len(players))
	for i, p := range players {
		playerResponses[i] = PlayerResponse{
			ID:       p.ID,
			Username: p.Username,
		}
	}

	return LobbyResponse{
		Code:       lobby.Code,
		State:      lobby.GetState().String(),
		Players:    playerResponses,
		HostID:     lobby.GetHostID(),
		MaxPlayers: lobby.MaxPlayers,
	}
}

// Create handles POST /api/v1/lobbies
func (c *LobbyController) Create(ctx *gin.Context) {
	var req CreateLobbyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lobby, err := c.lobbyService.CreateLobby(req.PlayerID, req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errMsgCreateLobby})
		return
	}

	ctx.JSON(http.StatusCreated, toLobbyResponse(lobby))
}

// Get handles GET /api/v1/lobbies/:code
func (c *LobbyController) Get(ctx *gin.Context) {
	code := ctx.Param("code")

	lobby, err := c.lobbyService.GetLobby(code)
	if err != nil {
		if errors.Is(err, services.ErrLobbyNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": errMsgLobbyNotFound})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errMsgGetLobby})
		return
	}

	ctx.JSON(http.StatusOK, toLobbyResponse(lobby))
}

// List handles GET /api/v1/lobbies
func (c *LobbyController) List(ctx *gin.Context) {
	lobbies, err := c.lobbyService.ListLobbies()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errMsgGetLobbies})
		return
	}

	response := make(LobbyListResponse, len(lobbies))
	for i, lobby := range lobbies {
		response[i] = toLobbyResponse(lobby)
	}

	ctx.JSON(http.StatusOK, response)
}

// Join handles POST /api/v1/lobbies/:code/join
func (c *LobbyController) Join(ctx *gin.Context) {
	code := ctx.Param("code")

	var req JoinLobbyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lobby, err := c.lobbyService.JoinLobby(code, req.PlayerID, req.Username)
	if err != nil {
		status := http.StatusInternalServerError
		message := errMsgJoinLobby

		switch {
		case errors.Is(err, services.ErrLobbyNotFound):
			status = http.StatusNotFound
			message = errMsgLobbyNotFound
		case errors.Is(err, game.ErrLobbyFull):
			status = http.StatusConflict
			message = errMsgLobbyFull
		case errors.Is(err, game.ErrPlayerAlreadyJoined):
			status = http.StatusConflict
			message = errMsgPlayerAlreadyInLobby
		case errors.Is(err, game.ErrInvalidStateForJoin):
			status = http.StatusConflict
			message = errMsgLobbyInvalidState
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusOK, toLobbyResponse(lobby))
}

// Leave handles POST /api/v1/lobbies/:code/leave
func (c *LobbyController) Leave(ctx *gin.Context) {
	code := ctx.Param("code")

	var req LeaveLobbyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.lobbyService.LeaveLobby(code, req.PlayerID)
	if err != nil {
		status := http.StatusInternalServerError
		message := errMsgLeaveLobby

		switch {
		case errors.Is(err, services.ErrLobbyNotFound):
			status = http.StatusNotFound
			message = errMsgLobbyNotFound
		case errors.Is(err, game.ErrPlayerNotFound):
			status = http.StatusNotFound
			message = errMsgPlayerNotInLobby
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": msgLeftLobby})
}

// Start handles POST /api/v1/lobbies/:code/start
func (c *LobbyController) Start(ctx *gin.Context) {
	code := ctx.Param("code")

	var req StartGameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.lobbyService.StartGame(code, req.PlayerID)
	if err != nil {
		status := http.StatusInternalServerError
		message := errMsgStartGame

		switch {
		case errors.Is(err, services.ErrLobbyNotFound):
			status = http.StatusNotFound
			message = errMsgLobbyNotFound
		case errors.Is(err, services.ErrNotHost):
			status = http.StatusForbidden
			message = errMsgOnlyHostCanStart
		case errors.Is(err, game.ErrInvalidStateForStart):
			status = http.StatusConflict
			message = errMsgGameInvalidState
		case errors.Is(err, game.ErrNotEnoughPlayers):
			status = http.StatusConflict
			message = errMsgNotEnoughPlayers
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	// Get the updated lobby to return
	lobby, err := c.lobbyService.GetLobby(code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errMsgGameStartLobbyState})
		return
	}

	ctx.JSON(http.StatusOK, toLobbyResponse(lobby))
}
