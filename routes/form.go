package routes

import (
	"lite-collector/handlers"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// RegisterFormRoutes registers form and nested submission routes
func RegisterFormRoutes(r *gin.RouterGroup, formService *services.FormService, submissionService *services.SubmissionService) {
	forms := r.Group("/forms")
	{
		forms.POST("/", handlers.CreateForm(formService))
		forms.GET("/", handlers.GetForms(formService))
		forms.GET("/:formId", handlers.GetForm(formService))
		forms.PUT("/:formId", handlers.UpdateForm(formService))
		forms.POST("/:formId/publish", handlers.PublishForm(formService))

		submissions := forms.Group("/:formId/submissions")
		{
			submissions.POST("/", handlers.CreateSubmission(submissionService))
			submissions.GET("/my", handlers.GetMySubmission(submissionService))
		}
	}
}
