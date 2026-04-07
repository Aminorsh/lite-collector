package routes

import (
	"lite-collector/handlers"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r *gin.RouterGroup, userService *services.UserService) {
	auth := r.Group("/auth")
	{
		auth.POST("/wx-login", handlers.WxLogin(userService))
	}
}
