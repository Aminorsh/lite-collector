package handlers

import (
	"net/http"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// GenerateForm godoc
// @Summary      AI 生成表单（异步）
// @Description  根据自然语言描述异步生成表单结构。返回任务 ID，前端轮询 GET /jobs/:jobId 获取结果。完成时 output 字段为 JSON 字符串，包含 title/description/schema。
// @Tags         AI任务
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      generateFormRequest  true  "表单描述"
// @Success      202   {object}  generateFormResponse
// @Failure      400   {object}  errorResponse  "请求参数错误"
// @Failure      401   {object}  errorResponse  "未登录或 token 已过期"
// @Failure      500   {object}  errorResponse  "服务器内部错误"
// @Router       /forms/generate [post]
func GenerateForm(aiJobService *services.AIJobService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req generateFormRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: err.Error()}})
			return
		}

		userID := c.MustGet("user_id").(uint64)
		jobID, err := aiJobService.EnqueueFormGeneration(userID, req.Description)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusAccepted, generateFormResponse{
			JobID:   jobID,
			Message: "AI 生成任务已排队，请通过 GET /api/v1/jobs/:jobId 查询结果",
		})
	}
}

type generateFormRequest struct {
	Description string `json:"description" binding:"required" example:"员工信息登记表，包含姓名、年龄、部门、手机号、月薪"`
}

type generateFormResponse struct {
	JobID   uint64 `json:"job_id"  example:"42"`
	Message string `json:"message" example:"AI 生成任务已排队，请通过 GET /api/v1/jobs/:jobId 查询结果"`
}
