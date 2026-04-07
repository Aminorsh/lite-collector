package routes

import (
	"lite-collector/handlers"
	"lite-collector/repository"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r *gin.RouterGroup, userRepo repository.UserRepository, jwtSecret []byte) {
	auth := r.Group("/auth")
	{
		auth.POST("/wx-login", handlers.WxLogin(userRepo, jwtSecret))
	}
}