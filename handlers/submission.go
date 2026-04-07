package handlers

import (
	"net/http"

	"lite-collector/repository"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// CreateSubmission handles creating a new form submission
func CreateSubmission(submissionRepo repository.SubmissionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (submitter)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get form ID from URL
		formID := c.Param("formId")
		if formID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID required"})
			return
		}

		// Parse request body - map of field_key to value
		var values map[string]interface{}
		if err := c.ShouldBindJSON(&values); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Create submission
		submissionService := services.NewSubmissionService(submissionRepo)
		submission, err := submissionService.CreateSubmission(formID, userID.(uint64), values)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission"})
			return
		}

		// Enqueue AI anomaly detection job (async)
		go func() {
			jobService := services.NewAIJobService()
			jobService.EnqueueAnomalyDetection(submission.ID)
		}()

		c.JSON(http.StatusCreated, gin.H{
			"id":         submission.ID,
			"status":     submission.Status,
			"submitted_at": submission.SubmittedAt,
		})
	}
}

// GetMySubmission handles getting the current user's submission for a form
func GetMySubmission(submissionRepo repository.SubmissionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get form ID from URL
		formID := c.Param("formId")
		if formID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID required"})
			return
		}

		// Get submission
		submissionService := services.NewSubmissionService(submissionRepo)
		submission, err := submissionService.GetMySubmission(formID, userID.(uint64))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
			return
		}

		// Get submission values
		submissionValues, err := submissionService.GetSubmissionValues(submission.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get submission values"})
			return
		}

		// Build response
		response := gin.H{
			"id":         submission.ID,
			"status":     submission.Status,
			"submitted_at": submission.SubmittedAt,
			"values":     make(map[string]interface{}),
		}

		// Add field values
		for _, sv := range submissionValues {
			response["values"].(map[string]interface{})[sv.FieldKey] = sv.Value
		}

		c.JSON(http.StatusOK, response)
	}
}