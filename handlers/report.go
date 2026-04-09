package handlers

import (
	"fmt"
	"net/http"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// GenerateReport godoc
// @Summary      生成表单数据报告
// @Description  触发 AI 对指定表单的所有提交数据进行汇总分析，生成报告。异步处理，返回任务 ID，前端轮询 GET /jobs/:jobId 获取结果。仅表单创建者可操作。
// @Tags         AI任务
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "表单 ID"
// @Success      202     {object}  generateReportResponse
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权操作该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
// @Failure      500     {object}  errorResponse  "服务器内部错误"
// @Router       /forms/{formId}/report [post]
func GenerateReport(formService *services.FormService, aiJobService *services.AIJobService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		form, err := formService.GetFormByID(formID, userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		jobID, err := aiJobService.EnqueueReport(userID, form.ID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusAccepted, generateReportResponse{
			JobID:   jobID,
			Message: fmt.Sprintf("报告生成已排队，请通过 GET /api/v1/jobs/%d 查询进度", jobID),
		})
	}
}

type generateReportResponse struct {
	JobID   uint64 `json:"job_id"  example:"6"`
	Message string `json:"message" example:"报告生成已排队，请通过 GET /api/v1/jobs/6 查询进度"`
}
