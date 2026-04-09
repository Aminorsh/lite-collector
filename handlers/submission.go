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

// ListSubmissions godoc
// @Summary      获取表单的所有提交记录
// @Description  获取指定表单下所有用户的提交记录（不含字段值）。仅表单创建者可访问。
// @Tags         提交记录
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "表单 ID"
// @Success      200     {object}  submissionListResponse
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权访问该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
// @Router       /forms/{formId}/submissions [get]
func ListSubmissions(formService *services.FormService, submissionService *services.SubmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		// Verify ownership via form service
		if _, err := formService.GetFormByID(formID, userID); err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		submissions, err := submissionService.GetSubmissionsByFormID(formID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		items := make([]submissionResponse, 0, len(submissions))
		for _, s := range submissions {
			items = append(items, submissionResponse{
				ID:          s.ID,
				Status:      s.Status,
				SubmittedAt: s.SubmittedAt,
			})
		}
		c.JSON(http.StatusOK, submissionListResponse{Submissions: items})
	}
}

// GetSubmission godoc
// @Summary      获取单条提交记录详情
// @Description  获取指定提交记录及其所有字段值。仅表单创建者可访问。
// @Tags         提交记录
// @Produce      json
// @Security     BearerAuth
// @Param        formId        path      int  true  "表单 ID"
// @Param        submissionId  path      int  true  "提交记录 ID"
// @Success      200           {object}  submissionWithValuesResponse
// @Failure      401           {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403           {object}  errorResponse  "无权访问该表单"
// @Failure      404           {object}  errorResponse  "表单或提交记录不存在"
// @Router       /forms/{formId}/submissions/{submissionId} [get]
func GetSubmission(formService *services.FormService, submissionService *services.SubmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")
		submissionID := c.Param("submissionId")

		// Verify ownership via form service
		if _, err := formService.GetFormByID(formID, userID); err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		result, err := submissionService.GetSubmissionByIDWithValues(submissionID)
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

// GetSubmissionsOverview godoc
// @Summary      获取表单提交总览（含字段值和异常原因）
// @Description  返回指定表单下所有提交记录的完整信息，包括每条提交的字段值和 AI 异常检测原因。适用于表格展示。仅表单创建者可访问。
// @Tags         提交记录
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "表单 ID"
// @Success      200     {object}  submissionOverviewResponse
// @Failure      401     {object}  errorResponse  "未登录或 token 已过期"
// @Failure      403     {object}  errorResponse  "无权访问该表单"
// @Failure      404     {object}  errorResponse  "表单不存在"
// @Router       /forms/{formId}/submissions/overview [get]
func GetSubmissionsOverview(formService *services.FormService, submissionService *services.SubmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		// Verify ownership and get form details
		form, err := formService.GetFormByID(formID, userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		items, err := submissionService.GetSubmissionsOverview(formID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, errorResponse{Error: errorDetail{Code: e.Code, Message: e.Message}})
			return
		}

		c.JSON(http.StatusOK, submissionOverviewResponse{
			FormID: form.ID,
			Title:  form.Title,
			Schema: string(form.Schema),
			Submissions: items,
		})
	}
}

// Request / response types

type submissionResponse struct {
	ID          uint64    `json:"id"           example:"7"`
	Status      int8      `json:"status"       example:"0"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type submissionListResponse struct {
	Submissions []submissionResponse `json:"submissions"`
}

type submissionOverviewResponse struct {
	FormID      uint64                            `json:"form_id"      example:"3"`
	Title       string                            `json:"title"        example:"员工信息登记"`
	Schema      string                            `json:"schema"`
	Submissions []services.SubmissionOverviewItem  `json:"submissions"`
}

type submissionWithValuesResponse struct {
	ID          uint64                 `json:"id"           example:"7"`
	Status      int8                   `json:"status"       example:"1"`
	SubmittedAt time.Time              `json:"submitted_at"`
	Values      map[string]interface{} `json:"values"`
}
