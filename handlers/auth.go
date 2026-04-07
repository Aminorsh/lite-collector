package handlers

import (
	"net/http"

	"lite-collector/services"
	"lite-collector/utils"

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
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, gin.H{"error": gin.H{"code": e.Code, "message": "code is required"}})
			return
		}

		token, user, err := userService.Login(req.Code)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, gin.H{"error": gin.H{"code": e.Code, "message": e.Message}})
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
