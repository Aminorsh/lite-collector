package handlers

import (
	"net/http"
	"time"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// CreateSubmission godoc
// @Summary      Submit a form
// @Description  Submit field values for a published form. Each user can submit once per form. AI anomaly detection is triggered asynchronously after submission.
// @Tags         submissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int                      true  "Form ID"
// @Param        body    body      map[string]interface{}   true  "Field key→value map"
// @Success      201     {object}  submissionResponse
// @Failure      400     {object}  errorResponse
// @Failure      401     {object}  errorResponse
// @Failure      500     {object}  errorResponse
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
// @Summary      Get my submission
// @Description  Returns the authenticated user's submission for a given form, including all field values.
// @Tags         submissions
// @Produce      json
// @Security     BearerAuth
// @Param        formId  path      int  true  "Form ID"
// @Success      200     {object}  submissionWithValuesResponse
// @Failure      401     {object}  errorResponse
// @Failure      404     {object}  errorResponse
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
