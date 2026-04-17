package routes

import (
	"lite-collector/handlers"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// RegisterJobRoutes registers AI job status routes
func RegisterJobRoutes(r *gin.RouterGroup, formService *services.FormService, aiJobService *services.AIJobService, pdfService *services.PDFService) {
	r.GET("/jobs/:jobId", handlers.GetJobStatus(aiJobService))
	r.GET("/jobs/:jobId/pdf", handlers.GetReportPDF(formService, aiJobService, pdfService))
}
