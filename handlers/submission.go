package handlers

import (
	"net/http"

	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// CreateSubmission handles submitting data for a form.
// Request body: flat map of field_key → value
func CreateSubmission(submissionService *services.SubmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uint64)
		formID := c.Param("formId")

		var values map[string]interface{}
		if err := c.ShouldBindJSON(&values); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		submission, err := submissionService.CreateSubmission(formID, userID, values)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create submission"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
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
