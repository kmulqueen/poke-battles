package routes

import (
	"poke-battles/internal/controllers"
	"poke-battles/internal/services"
	"poke-battles/internal/websocket"

	"github.com/gin-gonic/gin"
)

const v1BasePath = "/api/v1"

// RegisterRoutes registers API routes with injected dependencies
func RegisterRoutes(server *gin.Engine, lobbyService services.LobbyService, wsHandler *websocket.Handler) {
	v1 := server.Group(v1BasePath)

	// Health check
	healthCheckRoute := v1.Group("/health")
	health := controllers.NewHealthCheckController()
	healthCheckRoute.GET("/", health.Get)

	// Lobbies
	lobbiesRoute := v1.Group("/lobbies")
	lobby := controllers.NewLobbyController(lobbyService)
	lobbiesRoute.POST("", lobby.Create)
	lobbiesRoute.GET("/:code", lobby.Get)
	lobbiesRoute.POST("/:code/join", lobby.Join)
	lobbiesRoute.POST("/:code/leave", lobby.Leave)
	lobbiesRoute.POST("/:code/start", lobby.Start)

	// WebSocket
	wsRoute := v1.Group("/ws")
	wsRoute.GET("/game/:code", wsHandler.HandleConnection)
}
