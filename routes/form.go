package routes

import (
	"lite-collector/handlers"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// RegisterFormRoutes registers form and nested submission/base-data routes
func RegisterFormRoutes(r *gin.RouterGroup, formService *services.FormService, submissionService *services.SubmissionService, baseDataService *services.BaseDataService, aiJobService *services.AIJobService, formGenerator *services.FormGenerator) {
	forms := r.Group("/forms")
	{
		forms.POST("/", handlers.CreateForm(formService))
		forms.POST("/generate", handlers.GenerateForm(formGenerator)) // must be before /:formId
		forms.GET("/", handlers.GetForms(formService))
		forms.GET("/:formId", handlers.GetForm(formService))
		forms.GET("/:formId/schema", handlers.GetPublishedForm(formService)) // any auth'd user; only published forms
		forms.PUT("/:formId", handlers.UpdateForm(formService))
		forms.POST("/:formId/publish", handlers.PublishForm(formService))
		forms.POST("/:formId/archive", handlers.ArchiveForm(formService))
		forms.POST("/:formId/report", handlers.GenerateReport(formService, aiJobService))
		forms.GET("/:formId/report/latest", handlers.GetLatestReport(formService, aiJobService))

		baseData := forms.Group("/:formId/base-data")
		{
			baseData.POST("/", handlers.BatchImportBaseData(formService, baseDataService))
			baseData.GET("/", handlers.ListBaseData(formService, baseDataService))
			baseData.GET("/lookup", handlers.LookupBaseData(formService, baseDataService))
			baseData.DELETE("/", handlers.DeleteBaseData(formService, baseDataService))
		}

		submissions := forms.Group("/:formId/submissions")
		{
			submissions.POST("/", handlers.CreateSubmission(submissionService))
			submissions.GET("/", handlers.ListSubmissions(formService, submissionService))   // owner: all submissions
			submissions.GET("/my", handlers.GetMySubmission(submissionService))             // submitter: own submission
			submissions.GET("/overview", handlers.GetSubmissionsOverview(formService, submissionService)) // owner: table view with values + anomaly reasons
			submissions.GET("/:submissionId", handlers.GetSubmission(formService, submissionService)) // owner: one submission detail
		}
	}
}
