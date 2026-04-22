package routes

import (
	"lite-collector/handlers"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers user profile routes (requires auth)
func RegisterUserRoutes(r *gin.RouterGroup, userService *services.UserService, storage *services.StorageService) {
	r.PUT("/user/profile", handlers.UpdateProfile(userService))
	r.POST("/user/avatar", handlers.UploadAvatar(userService, storage))
}
