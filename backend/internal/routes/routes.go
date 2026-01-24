package routes

import (
	"poke-battles/internal/config"
	"poke-battles/internal/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers API routes
func RegisterRoutes(server *gin.Engine) {
	//* Version Groups
	v1 := server.Group(config.V1ApiBasePath)

	//* V1 Routes
	{
		healthCheckRoute := v1.Group("/health")
		health := controllers.NewHealthCheckController()
		healthCheckRoute.GET("/", health.Get)
	}

}
