package handlers

import (
	"net/http"

	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// WxLogin handles WeChat login.
// Request body: { "code": "<wechat_login_code>" }
func WxLogin(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Code string `json:"code" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
			return
		}

		token, user, err := userService.Login(req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":         user.ID,
				"openid":     user.OpenID,
				"nickname":   user.Nickname,
				"avatar_url": user.AvatarURL,
			},
		})
	}
}
