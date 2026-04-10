package handlers

import (
	"net/http"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// GenerateForm godoc
// @Summary      AI 生成表单
// @Description  根据自然语言描述，使用 AI 生成表单标题、描述和字段结构（schema）。返回结果可直接用于创建表单。
// @Tags         AI任务
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      generateFormRequest  true  "表单描述"
// @Success      200   {object}  generateFormResponse
// @Failure      400   {object}  errorResponse  "请求参数错误"
// @Failure      401   {object}  errorResponse  "未登录或 token 已过期"
// @Failure      502   {object}  errorResponse  "AI 生成失败"
// @Failure      503   {object}  errorResponse  "AI 服务未配置"
// @Router       /forms/generate [post]
func GenerateForm(formGenerator *services.FormGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req generateFormRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: err.Error()}})
			return
		}

		result, err := formGenerator.Generate(req.Description)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, generateFormResponse{
			Title:       result.Title,
			Description: result.Description,
			Schema:      result.Schema,
		})
	}
}

type generateFormRequest struct {
	Description string `json:"description" binding:"required" example:"员工信息登记表，包含姓名、年龄、部门、手机号、月薪"`
}

type generateFormResponse struct {
	Title       string `json:"title"       example:"员工信息登记"`
	Description string `json:"description" example:"请填写个人基本信息"`
	Schema      string `json:"schema"      example:"{\"fields\":[...]}"`
}
