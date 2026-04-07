package routes

import (
	"lite-collector/handlers"
	"lite-collector/repository"

	"github.com/gin-gonic/gin"
)

// RegisterSubmissionRoutes registers submission-related routes
func RegisterSubmissionRoutes(r *gin.RouterGroup, submissionRepo repository.SubmissionRepository) {
	submissions := r.Group("/submissions")
	{
		submissions.POST("/", handlers.CreateSubmission(submissionRepo))
		submissions.GET("/my", handlers.GetMySubmission(submissionRepo))
	}
}