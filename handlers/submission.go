package handlers

import (
	"net/http"

	"lite-collector/services"
	"lite-collector/utils"

	"github.com/gin-gonic/gin"
)

// CreateSubmission handles submitting data for a form.
// Request body: flat map of field_key → value
func CreateSubmission(submissionService *services.SubmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		var values map[string]any
		if err := c.ShouldBindJSON(&values); err != nil {
			e := utils.ErrBadRequest
			c.JSON(e.HTTPStatus, gin.H{"error": gin.H{"code": e.Code, "message": err.Error()}})
			return
		}

		submission, err := submissionService.CreateSubmission(formID, userID, values)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, gin.H{"error": gin.H{"code": e.Code, "message": e.Message}})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":           submission.ID,
			"status":       submission.Status,
			"submitted_at": submission.SubmittedAt,
		})
	}
}

// GetMySubmission returns the current user's submission for a form, including field values.
func GetMySubmission(submissionService *services.SubmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		result, err := submissionService.GetMySubmissionWithValues(formID, userID)
		if err != nil {
			e := utils.AsAppError(err)
			c.JSON(e.HTTPStatus, gin.H{"error": gin.H{"code": e.Code, "message": e.Message}})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":           result.Submission.ID,
			"status":       result.Submission.Status,
			"submitted_at": result.Submission.SubmittedAt,
			"values":       result.Values,
		})
	}
}
