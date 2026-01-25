package main

import (
	"os"

	"poke-battles/internal/middleware"
	"poke-battles/internal/routes"
	"poke-battles/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	// Middleware
	server.Use(middleware.CORS())

	// Services
	lobbyService := services.NewLobbyService()

	// Routes
	routes.RegisterRoutes(server, lobbyService)

	// Run server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := server.Run(":" + port); err != nil {
		panic(err)
	}
}
