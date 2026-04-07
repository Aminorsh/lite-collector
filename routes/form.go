package routes

import (
	"lite-collector/handlers"
	"lite-collector/repository"

	"github.com/gin-gonic/gin"
)

// RegisterFormRoutes registers form-related routes
func RegisterFormRoutes(r *gin.RouterGroup, formRepo repository.FormRepository, submissionRepo repository.SubmissionRepository) {
	forms := r.Group("/forms")
	{
		forms.POST("/", handlers.CreateForm(formRepo))
		forms.GET("/", handlers.GetForms(formRepo))
		forms.GET("/:formId", handlers.GetForm(formRepo))
		forms.PUT("/:formId", handlers.UpdateForm(formRepo))
		forms.POST("/:formId/publish", handlers.PublishForm(formRepo))

		// Submission routes nested under forms
		submissions := forms.Group("/:formId/submissions")
		{
			submissions.POST("/", handlers.CreateSubmission(submissionRepo))
			submissions.GET("/my", handlers.GetMySubmission(submissionRepo))
		}
	}
}