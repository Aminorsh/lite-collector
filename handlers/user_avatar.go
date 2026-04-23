package handlers

import (
	"net/http"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

const maxAvatarBytes = 2 * 1024 * 1024 // 2 MB

// UploadAvatar godoc
// @Summary      上传用户头像
// @Description  接收 multipart/form-data 格式的头像图片文件（字段名 file），保存到服务器并更新当前用户的 avatar_url。最大 2MB，仅支持 jpeg/png/webp。
// @Tags         认证
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "头像图片 (jpeg/png/webp, ≤2MB)"
// @Success      200   {object}  avatarResponse
// @Failure      400   {object}  errorResponse  "未提供文件"
// @Failure      401   {object}  errorResponse  "未登录或 token 已过期"
// @Failure      413   {object}  errorResponse  "文件过大"
// @Failure      415   {object}  errorResponse  "不支持的文件类型"
// @Router       /user/avatar [post]
func UploadAvatar(userService *services.UserService, storage *services.StorageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)

		fileHeader, err := c.FormFile("file")
		if err != nil {
			e := utils.ErrAvatarMissing
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		if fileHeader.Size > maxAvatarBytes {
			e := utils.ErrAvatarTooLarge
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		f, err := fileHeader.Open()
		if err != nil {
			e := utils.ErrInternal
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}
		defer f.Close()

		contentType := fileHeader.Header.Get("Content-Type")
		url, err := storage.UploadAvatar(userID, f, contentType)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		user, err := userService.UpdateProfile(userID, "", url)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, avatarResponse{
			AvatarURL: user.AvatarURL,
		})
	}
}

type avatarResponse struct {
	AvatarURL string `json:"avatar_url" example:"/static/avatars/1-1704067200000000000.jpg"`
}
