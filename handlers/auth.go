package handlers

import (
	"net/http"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// WxLogin godoc
// @Summary      微信登录
// @Description  使用微信 wx.login() 返回的临时 code 换取 JWT token。首次登录时自动创建用户。
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body      wxLoginRequest  true  "微信登录 code"
// @Success      200   {object}  wxLoginResponse
// @Failure      400   {object}  errorResponse   "请求参数错误"
// @Failure      500   {object}  errorResponse   "服务器内部错误"
// @Router       /auth/wx-login [post]
func WxLogin(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req wxLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: "code is required"}})
			return
		}

		token, user, err := userService.Login(req.Code)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, wxLoginResponse{
			Token: token,
			User: userInfo{
				ID:        user.ID,
				OpenID:    user.OpenID,
				Nickname:  user.Nickname,
				AvatarURL: user.AvatarURL,
			},
		})
	}
}

// Request / response types used by swag — also serve as readable contracts.

type wxLoginRequest struct {
	Code string `json:"code" binding:"required" example:"wx_login_code_abc123"`
}

type wxLoginResponse struct {
	Token string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  userInfo `json:"user"`
}

type userInfo struct {
	ID        uint64 `json:"id"         example:"1"`
	OpenID    string `json:"openid"     example:"oXxxx_abc123"`
	Nickname  string `json:"nickname"   example:"WeChat User"`
	AvatarURL string `json:"avatar_url" example:""`
}
