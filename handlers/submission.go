package handlers

import (
	"net/http"
	"time"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// CreateSubmission godoc
// @Summary      提交表单数据
// @Description  向已发布的表单提交字段数据。每位用户对每个表单只能提交一次。提交后后端自动触发 AI 异常检测（异步），status 会在后台更新。
// @Tags         提交记录
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int                    true  "表单 ID"
// @Param        body    body      map[string]interface{} true  "字段 key→value 映射，key 对应表单 schema 中的 field_key"
// @Success      201     {object}  submissionResponse
// @Failure      400     {object}  errorResponse  "请求参数错误"
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      500     {object}  errorResponse  "服务器内部错误"
// @Router       /forms/{formId}/submissions [post]
func CreateSubmission(submissionService *services.SubmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		var values map[string]any
		if err := c.ShouldBindJSON(&values); err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: err.Error()}})
			return
		}

		submission, err := submissionService.CreateSubmission(formID, userID, values)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusCreated, submissionResponse{
			ID:          submission.ID,
			Status:      submission.Status,
			SubmittedAt: submission.SubmittedAt,
		})
	}
}

// GetMySubmission godoc
// @Summary      获取我的提交记录
// @Description  获取当前登录用户在指定表单中的提交记录，包含所有填写的字段值。每人每表单限提交一次。
// @Tags         提交记录
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "表单 ID"
// @Success      200     {object}  submissionWithValuesResponse
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      404     {object}  errorResponse  "提交记录不存在"
// @Router       /forms/{formId}/submissions/my [get]
func GetMySubmission(submissionService *services.SubmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		result, err := submissionService.GetMySubmissionWithValues(formID, userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, submissionWithValuesResponse{
			ID:          result.Submission.ID,
			Status:      result.Submission.Status,
			SubmittedAt: result.Submission.SubmittedAt,
			Values:      result.Values,
		})
	}
}

// Request / response types

type submissionResponse struct {
	ID          uint64    `json:"id"           example:"7"`
	Status      int8      `json:"status"       example:"0"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type submissionWithValuesResponse struct {
	ID          uint64                 `json:"id"           example:"7"`
	Status      int8                   `json:"status"       example:"1"`
	SubmittedAt time.Time              `json:"submitted_at"`
	Values      map[string]interface{} `json:"values"`
}
