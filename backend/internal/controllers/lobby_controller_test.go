package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"poke-battles/internal/services"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() (*gin.Engine, *LobbyController) {
	svc := services.NewLobbyService()
	ctrl := NewLobbyController(svc)

	router := gin.New()
	api := router.Group("/api/v1")
	{
		api.POST("/lobbies", ctrl.Create)
		api.GET("/lobbies", ctrl.List)
		api.GET("/lobbies/:code", ctrl.Get)
		api.POST("/lobbies/:code/join", ctrl.Join)
		api.POST("/lobbies/:code/leave", ctrl.Leave)
		api.POST("/lobbies/:code/start", ctrl.Start)
	}

	return router, ctrl
}

// ========================================
// Create Lobby Tests
// ========================================

func TestCreate_Success(t *testing.T) {
	router, _ := setupTestRouter()

	body := CreateLobbyRequest{
		PlayerID: "host-1",
		Username: "HostPlayer",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp LobbyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Code) != 6 {
		t.Errorf("expected 6-char code, got %q", resp.Code)
	}
	if resp.State != "waiting" {
		t.Errorf("expected state 'waiting', got %q", resp.State)
	}
	if len(resp.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(resp.Players))
	}
	if resp.HostID != "host-1" {
		t.Errorf("expected host_id 'host-1', got %q", resp.HostID)
	}
}

func TestCreate_MissingPlayerID(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"username": "Player"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreate_MissingUsername(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"player_id": "player-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreate_EmptyBody(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ========================================
// Get Lobby Tests
// ========================================

func TestGet_Success(t *testing.T) {
	router, _ := setupTestRouter()

	// Create a lobby first
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	// Get the lobby
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lobbies/"+createResp.Code, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp LobbyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != createResp.Code {
		t.Errorf("expected code %q, got %q", createResp.Code, resp.Code)
	}
}

func TestGet_NotFound(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/lobbies/NOTFND", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["error"] != errMsgLobbyNotFound {
		t.Errorf("expected error %q, got %q", errMsgLobbyNotFound, resp["error"])
	}
}

// ========================================
// List Lobbies Tests
// ========================================

func TestList_Success(t *testing.T) {
	router, _ := setupTestRouter()

	// Create a lobby first
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	// List all lobbies
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lobbies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp LobbyListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response as array: %v", err)
	}

	if len(resp) != 1 {
		t.Errorf("expected 1 lobby, got %d", len(resp))
	}

	lobby := resp[0]
	if lobby.Code != createResp.Code {
		t.Errorf("expected code %q, got %q", createResp.Code, lobby.Code)
	}
	if lobby.State != "waiting" {
		t.Errorf("expected state 'waiting', got %q", lobby.State)
	}
	if len(lobby.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(lobby.Players))
	}
	if lobby.HostID != "host-1" {
		t.Errorf("expected host_id 'host-1', got %q", lobby.HostID)
	}
	if lobby.MaxPlayers != 2 {
		t.Errorf("expected max_players 2, got %d", lobby.MaxPlayers)
	}
}

func TestList_NoLobbies(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/lobbies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Expected behavior: return 200 with empty array
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (empty list should return 200 with empty array, not 404)", http.StatusOK, w.Code)
	}

	var resp LobbyListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response as array: %v", err)
	}

	if len(resp) != 0 {
		t.Errorf("expected empty array, got %d lobbies", len(resp))
	}
}

func TestList_MultipleLobbies(t *testing.T) {
	router, _ := setupTestRouter()

	// Create three lobbies
	lobbyCodes := make([]string, 3)
	for i := 0; i < 3; i++ {
		createBody := fmt.Sprintf(`{"player_id": "host-%d", "username": "Host%d"}`, i+1, i+1)
		createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		router.ServeHTTP(createW, createReq)

		var createResp LobbyResponse
		json.Unmarshal(createW.Body.Bytes(), &createResp)
		lobbyCodes[i] = createResp.Code
	}

	// List all lobbies
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lobbies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp LobbyListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response as array: %v", err)
	}

	if len(resp) != 3 {
		t.Errorf("expected 3 lobbies, got %d", len(resp))
	}

	// Verify all created lobbies are in the response
	returnedCodes := make(map[string]bool)
	for _, lobby := range resp {
		returnedCodes[lobby.Code] = true

		// Verify response structure
		if lobby.State == "" {
			t.Error("lobby state should not be empty")
		}
		if len(lobby.Players) == 0 {
			t.Error("lobby should have at least one player")
		}
		if lobby.HostID == "" {
			t.Error("lobby host_id should not be empty")
		}
		if lobby.MaxPlayers != 2 {
			t.Errorf("expected max_players 2, got %d", lobby.MaxPlayers)
		}
	}

	for _, code := range lobbyCodes {
		if !returnedCodes[code] {
			t.Errorf("expected lobby %q in response, but it was missing", code)
		}
	}
}

// ========================================
// Join Lobby Tests
// ========================================

func TestJoin_Success(t *testing.T) {
	router, _ := setupTestRouter()

	// Create a lobby first
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	// Join the lobby
	joinBody := `{"player_id": "player-2", "username": "Player2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/join", bytes.NewBufferString(joinBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp LobbyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(resp.Players))
	}
	if resp.State != "ready" {
		t.Errorf("expected state 'ready', got %q", resp.State)
	}
}

func TestJoin_MissingPlayerID(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"username": "Player"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/ABC123/join", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestJoin_LobbyNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"player_id": "player-1", "username": "Player"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/NOTFND/join", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["error"] != errMsgLobbyNotFound {
		t.Errorf("expected error %q, got %q", errMsgLobbyNotFound, resp["error"])
	}
}

func TestJoin_LobbyFull(t *testing.T) {
	router, _ := setupTestRouter()

	// Create and fill lobby
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	joinBody := `{"player_id": "player-2", "username": "Player2"}`
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/join", bytes.NewBufferString(joinBody))
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	router.ServeHTTP(joinW, joinReq)

	// Try to join full lobby - state is Ready, so we get "cannot join in current state"
	body := `{"player_id": "player-3", "username": "Player3"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/join", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	// When lobby has 2 players, state becomes Ready. The state check happens first,
	// so we get "cannot join in current state" instead of "lobby is full"
	if resp["error"] != errMsgLobbyInvalidState {
		t.Errorf("expected error %q, got %q", errMsgLobbyInvalidState, resp["error"])
	}
}

func TestJoin_AlreadyJoined(t *testing.T) {
	router, _ := setupTestRouter()

	// Create lobby
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	// Try to join as host again
	body := `{"player_id": "host-1", "username": "Host"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/join", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["error"] != errMsgPlayerAlreadyInLobby {
		t.Errorf("expected error %q, got %q", errMsgPlayerAlreadyInLobby, resp["error"])
	}
}

// ========================================
// Leave Lobby Tests
// ========================================

func TestLeave_Success(t *testing.T) {
	router, _ := setupTestRouter()

	// Create and fill lobby
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	joinBody := `{"player_id": "player-2", "username": "Player2"}`
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/join", bytes.NewBufferString(joinBody))
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	router.ServeHTTP(joinW, joinReq)

	// Player leaves
	leaveBody := `{"player_id": "player-2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/leave", bytes.NewBufferString(leaveBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["message"] != msgLeftLobby {
		t.Errorf("expected message %q, got %q", msgLeftLobby, resp["message"])
	}
}

func TestLeave_LobbyNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"player_id": "player-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/NOTFND/leave", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestLeave_PlayerNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	// Create lobby
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	// Try to leave as non-existent player
	body := `{"player_id": "nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/leave", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["error"] != errMsgPlayerNotInLobby {
		t.Errorf("expected error %q, got %q", errMsgPlayerNotInLobby, resp["error"])
	}
}

func TestLeave_MissingPlayerID(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/ABC123/leave", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ========================================
// Start Game Tests
// ========================================

func TestStart_Success(t *testing.T) {
	router, _ := setupTestRouter()

	// Create and fill lobby
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	joinBody := `{"player_id": "player-2", "username": "Player2"}`
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/join", bytes.NewBufferString(joinBody))
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	router.ServeHTTP(joinW, joinReq)

	// Start game
	startBody := `{"player_id": "host-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/start", bytes.NewBufferString(startBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp LobbyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.State != "active" {
		t.Errorf("expected state 'active', got %q", resp.State)
	}
}

func TestStart_LobbyNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"player_id": "host-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/NOTFND/start", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestStart_NotHost(t *testing.T) {
	router, _ := setupTestRouter()

	// Create and fill lobby
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	joinBody := `{"player_id": "player-2", "username": "Player2"}`
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/join", bytes.NewBufferString(joinBody))
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	router.ServeHTTP(joinW, joinReq)

	// Non-host tries to start
	body := `{"player_id": "player-2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/start", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["error"] != errMsgOnlyHostCanStart {
		t.Errorf("expected error %q, got %q", errMsgOnlyHostCanStart, resp["error"])
	}
}

func TestStart_NotReady(t *testing.T) {
	router, _ := setupTestRouter()

	// Create lobby without second player
	createBody := `{"player_id": "host-1", "username": "Host"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)

	// Try to start with only 1 player
	body := `{"player_id": "host-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+createResp.Code+"/start", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["error"] != errMsgGameInvalidState {
		t.Errorf("expected error %q, got %q", errMsgGameInvalidState, resp["error"])
	}
}

func TestStart_MissingPlayerID(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/ABC123/start", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ========================================
// Error Mapping Tests
// ========================================

func TestErrorMapping_AllDomainErrorsMapToCorrectHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(router *gin.Engine) string // Returns lobby code if needed
		method         string
		pathBuilder    func(code string) string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "LobbyNotFound on Get",
			setup:          func(r *gin.Engine) string { return "" },
			method:         http.MethodGet,
			pathBuilder:    func(code string) string { return "/api/v1/lobbies/NOTFND" },
			body:           "",
			expectedStatus: http.StatusNotFound,
			expectedError:  errMsgLobbyNotFound,
		},
		{
			name:           "LobbyNotFound on Join",
			setup:          func(r *gin.Engine) string { return "" },
			method:         http.MethodPost,
			pathBuilder:    func(code string) string { return "/api/v1/lobbies/NOTFND/join" },
			body:           `{"player_id": "p1", "username": "P1"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  errMsgLobbyNotFound,
		},
		{
			name: "LobbyInvalidState on Join (full lobby in Ready state)",
			setup: func(r *gin.Engine) string {
				// Create and fill lobby - becomes Ready state
				createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies",
					bytes.NewBufferString(`{"player_id": "h1", "username": "H1"}`))
				createReq.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, createReq)
				var resp LobbyResponse
				json.Unmarshal(w.Body.Bytes(), &resp)

				joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+resp.Code+"/join",
					bytes.NewBufferString(`{"player_id": "p2", "username": "P2"}`))
				joinReq.Header.Set("Content-Type", "application/json")
				w = httptest.NewRecorder()
				r.ServeHTTP(w, joinReq)

				return resp.Code
			},
			method:         http.MethodPost,
			pathBuilder:    func(code string) string { return "/api/v1/lobbies/" + code + "/join" },
			body:           `{"player_id": "p3", "username": "P3"}`,
			expectedStatus: http.StatusConflict,
			expectedError:  errMsgLobbyInvalidState, // State check happens before full check
		},
		{
			name: "PlayerAlreadyJoined on Join",
			setup: func(r *gin.Engine) string {
				createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies",
					bytes.NewBufferString(`{"player_id": "h1", "username": "H1"}`))
				createReq.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, createReq)
				var resp LobbyResponse
				json.Unmarshal(w.Body.Bytes(), &resp)
				return resp.Code
			},
			method:         http.MethodPost,
			pathBuilder:    func(code string) string { return "/api/v1/lobbies/" + code + "/join" },
			body:           `{"player_id": "h1", "username": "H1"}`,
			expectedStatus: http.StatusConflict,
			expectedError:  errMsgPlayerAlreadyInLobby,
		},
		{
			name: "NotHost on Start",
			setup: func(r *gin.Engine) string {
				createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies",
					bytes.NewBufferString(`{"player_id": "h1", "username": "H1"}`))
				createReq.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, createReq)
				var resp LobbyResponse
				json.Unmarshal(w.Body.Bytes(), &resp)

				joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+resp.Code+"/join",
					bytes.NewBufferString(`{"player_id": "p2", "username": "P2"}`))
				joinReq.Header.Set("Content-Type", "application/json")
				w = httptest.NewRecorder()
				r.ServeHTTP(w, joinReq)

				return resp.Code
			},
			method:         http.MethodPost,
			pathBuilder:    func(code string) string { return "/api/v1/lobbies/" + code + "/start" },
			body:           `{"player_id": "p2"}`,
			expectedStatus: http.StatusForbidden,
			expectedError:  errMsgOnlyHostCanStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, _ := setupTestRouter()
			code := tt.setup(router)

			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.pathBuilder(code), bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.pathBuilder(code), nil)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var resp map[string]string
				json.Unmarshal(w.Body.Bytes(), &resp)
				if resp["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, resp["error"])
				}
			}
		})
	}
}

// ========================================
// Full Integration Flow Tests
// ========================================

func TestFullFlow_CreateJoinStartLeave(t *testing.T) {
	router, _ := setupTestRouter()

	// 1. Host creates lobby
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies",
		bytes.NewBufferString(`{"player_id": "host-1", "username": "Host"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("create failed with status %d", createW.Code)
	}

	var lobby LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &lobby)
	code := lobby.Code

	// 2. Player 2 joins
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+code+"/join",
		bytes.NewBufferString(`{"player_id": "player-2", "username": "Player2"}`))
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	router.ServeHTTP(joinW, joinReq)

	if joinW.Code != http.StatusOK {
		t.Fatalf("join failed with status %d", joinW.Code)
	}

	json.Unmarshal(joinW.Body.Bytes(), &lobby)
	if lobby.State != "ready" {
		t.Errorf("expected state 'ready', got %q", lobby.State)
	}

	// 3. Host starts game
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+code+"/start",
		bytes.NewBufferString(`{"player_id": "host-1"}`))
	startReq.Header.Set("Content-Type", "application/json")
	startW := httptest.NewRecorder()
	router.ServeHTTP(startW, startReq)

	if startW.Code != http.StatusOK {
		t.Fatalf("start failed with status %d", startW.Code)
	}

	json.Unmarshal(startW.Body.Bytes(), &lobby)
	if lobby.State != "active" {
		t.Errorf("expected state 'active', got %q", lobby.State)
	}

	// 4. Verify lobby is still accessible
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/lobbies/"+code, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("get failed with status %d", getW.Code)
	}

	json.Unmarshal(getW.Body.Bytes(), &lobby)
	if lobby.State != "active" {
		t.Errorf("expected state 'active', got %q", lobby.State)
	}
}

func TestFullFlow_HostLeaveReassignAndRejoin(t *testing.T) {
	router, _ := setupTestRouter()

	// Create lobby
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies",
		bytes.NewBufferString(`{"player_id": "host-1", "username": "Host"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var lobby LobbyResponse
	json.Unmarshal(createW.Body.Bytes(), &lobby)
	code := lobby.Code

	// Player 2 joins
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+code+"/join",
		bytes.NewBufferString(`{"player_id": "player-2", "username": "Player2"}`))
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	router.ServeHTTP(joinW, joinReq)

	// Host leaves
	leaveReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+code+"/leave",
		bytes.NewBufferString(`{"player_id": "host-1"}`))
	leaveReq.Header.Set("Content-Type", "application/json")
	leaveW := httptest.NewRecorder()
	router.ServeHTTP(leaveW, leaveReq)

	if leaveW.Code != http.StatusOK {
		t.Fatalf("leave failed with status %d", leaveW.Code)
	}

	// Verify player-2 is now host
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/lobbies/"+code, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	json.Unmarshal(getW.Body.Bytes(), &lobby)
	if lobby.HostID != "player-2" {
		t.Errorf("expected host_id 'player-2', got %q", lobby.HostID)
	}
	if lobby.State != "waiting" {
		t.Errorf("expected state 'waiting', got %q", lobby.State)
	}

	// Original host rejoins as regular player
	rejoinReq := httptest.NewRequest(http.MethodPost, "/api/v1/lobbies/"+code+"/join",
		bytes.NewBufferString(`{"player_id": "host-1", "username": "FormerHost"}`))
	rejoinReq.Header.Set("Content-Type", "application/json")
	rejoinW := httptest.NewRecorder()
	router.ServeHTTP(rejoinW, rejoinReq)

	if rejoinW.Code != http.StatusOK {
		t.Fatalf("rejoin failed with status %d", rejoinW.Code)
	}

	json.Unmarshal(rejoinW.Body.Bytes(), &lobby)
	if len(lobby.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(lobby.Players))
	}
	// Player-2 should still be host
	if lobby.HostID != "player-2" {
		t.Errorf("expected host_id to remain 'player-2', got %q", lobby.HostID)
	}
	if lobby.State != "ready" {
		t.Errorf("expected state 'ready', got %q", lobby.State)
	}
}
