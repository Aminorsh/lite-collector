package routes

import (
	"lite-collector/handlers"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers user profile routes (requires auth)
func RegisterUserRoutes(r *gin.RouterGroup, userService *services.UserService) {
	r.PUT("/user/profile", handlers.UpdateProfile(userService))
}
