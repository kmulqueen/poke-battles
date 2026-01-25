package controllers

// Error messages for API responses
const (
	errMsgCreateLobby          = "failed to create lobby"
	errMsgLobbyNotFound        = "lobby not found"
	errMsgGetLobby             = "failed to get lobby"
	errMsgJoinLobby            = "failed to join lobby"
	errMsgLobbyFull            = "lobby is full"
	errMsgLeaveLobby           = "failed to leave lobby"
	errMsgPlayerAlreadyInLobby = "player already in lobby"
	errMsgPlayerNotInLobby     = "player not found in lobby"
	errMsgLobbyInvalidState    = "cannot join lobby in current state"
	errMsgStartGame            = "failed to start game"
	errMsgOnlyHostCanStart     = "only host can start the game"
	errMsgGameInvalidState     = "cannot start game in current state"
	errMsgNotEnoughPlayers     = "not enough players to start"
	errMsgGameStartLobbyState  = "game started but failed to get lobby state"
)

// Success messages for API responses
const (
	msgLeftLobby = "left lobby successfully"
)
