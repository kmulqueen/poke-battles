package main

import (
	"poke-battles/internal/config"
	"poke-battles/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	// CORS middleware
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// Routes
	routes.RegisterRoutes(server)

	// Run server
	if err := server.Run(config.ServerPORT); err != nil {
		panic(err)
	}
}
