package routes

import (
	"lite-collector/handlers"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// RegisterSubmissionRoutes registers standalone submission routes.
// NOTE: submissions are also available as nested routes under /forms/:formId/submissions
// via RegisterFormRoutes. This file is kept for potential future use.
func RegisterSubmissionRoutes(r *gin.RouterGroup, submissionService *services.SubmissionService) {
	submissions := r.Group("/submissions")
	{
		submissions.POST("/", handlers.CreateSubmission(submissionService))
		submissions.GET("/my", handlers.GetMySubmission(submissionService))
	}
}
