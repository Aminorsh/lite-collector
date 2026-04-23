package handlers

import (
	"encoding/json"
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

// ListPendingJobs godoc
// @Summary      查询当前用户的待处理 AI 任务
// @Description  返回当前用户所有进行中（status 0/1）的 AI 任务，以及最近 10 分钟内完成/失败（status 2/3）的任务。用于首页“AI 生成中 / 已完成”横幅。
// @Tags         AI任务
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  pendingJobsResponse
// @Failure      401  {object}  errorResponse  "未登录或 token 已过期"
// @Router       /jobs/pending [get]
func ListPendingJobs(aiJobService *services.AIJobService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)

		jobs, err := aiJobService.ListPendingJobs(userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		items := make([]pendingJobItem, 0, len(jobs))
		for _, j := range jobs {
			it := pendingJobItem{
				ID:         j.ID,
				JobType:    j.JobType,
				Status:     j.Status,
				CreatedAt:  j.CreatedAt,
				FinishedAt: j.FinishedAt,
			}
			if fid := extractFormID(j.Input); fid != 0 {
				f := fid
				it.FormID = &f
			}
			items = append(items, it)
		}

		c.JSON(http.StatusOK, pendingJobsResponse{Jobs: items})
	}
}

// extractFormID parses {"form_id":N} out of the Input JSON. Returns 0 when absent.
func extractFormID(input string) uint64 {
	if input == "" {
		return 0
	}
	var payload struct {
		FormID uint64 `json:"form_id"`
	}
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return 0
	}
	return payload.FormID
}
