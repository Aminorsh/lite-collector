package routes

import (
	"lite-collector/handlers"
	"lite-collector/services"

	"github.com/gin-gonic/gin"
)

// RegisterJobRoutes registers AI job status routes
func RegisterJobRoutes(r *gin.RouterGroup, aiJobService *services.AIJobService) {
	r.GET("/jobs/pending", handlers.ListPendingJobs(aiJobService))
	r.GET("/jobs/:jobId", handlers.GetJobStatus(aiJobService))
}
