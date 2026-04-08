package handlers

import (
	"net/http"
	"strconv"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// GetJobStatus godoc
// @Summary      查询 AI 任务状态
// @Description  根据任务 ID 查询异步 AI 任务的当前状态及结果。状态值：0=排队中 1=处理中 2=已完成 3=失败。
// @Tags         AI任务
// @Produce      json
// @Security     BearerAuth
// @Param        jobId  path      int  true  "AI 任务 ID"
// @Success      200    {object}  jobStatusResponse
// @Failure      401    {object}  errorResponse  "未登录或 token 已过期"
// @Failure      404    {object}  errorResponse  "任务不存在"
// @Router       /jobs/{jobId} [get]
func GetJobStatus(aiJobService *services.AIJobService) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID, err := strconv.ParseUint(c.Param("jobId"), 10, 64)
		if err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: "invalid job id"}})
			return
		}

		job, err := aiJobService.GetJobStatus(jobID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, jobStatusResponse{
			ID:         job.ID,
			JobType:    job.JobType,
			Status:     job.Status,
			Input:      job.Input,
			Output:     job.Output,
			CreatedAt:  job.CreatedAt,
			FinishedAt: job.FinishedAt,
		})
	}
}
