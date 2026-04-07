package handlers

import (
	"net/http"
	"time"

	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// WxLogin handles WeChat login request
// Expected request body: { "code": "wechat_login_code" }
func WxLogin(userRepo repository.UserRepository, jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Code string `json:"code" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// In a real implementation, we would exchange the code with WeChat servers
		// For now, we'll simulate a response
		openid := "simulated_openid_" + req.Code // This is just for demo
		nickname := "WeChat User"
		avatarURL := ""

		// Find or create user
		userService := services.NewUserService(userRepo)
		user, err := userService.FindByOpenID(openid)
		if err != nil {
			// User doesn't exist, create new one
			user = &models.User{
				OpenID:    openid,
				Nickname:  nickname,
				AvatarURL: avatarURL,
			}
			if err := userService.Create(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user"})
				return
			}
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"openid":  user.OpenID,
			"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours expiry
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Return user info and token
		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":       user.ID,
				"openid":   user.OpenID,
				"nickname": user.Nickname,
				"avatar_url": user.AvatarURL,
			},
			"token": tokenString,
		})
	}
}