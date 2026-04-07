package routes

import (
	"lite-collector/db"
	"lite-collector/handlers"
	"lite-collector/repository"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r *gin.RouterGroup) {
	// Initialize user repository
	userRepo := repository.NewUserRepository(db.DB)
	auth := r.Group("/auth")
	{
		auth.POST("/wx-login", handlers.WxLogin(userRepo))
	}
}