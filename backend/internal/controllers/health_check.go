package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthCheckController() *HealthController {
	return &HealthController{}
}

// Get returns a 200 status code if the API is running
func (h *HealthController) Get(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Backend is running",
	})
}
