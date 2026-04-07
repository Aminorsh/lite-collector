package main

import (
	"lite-collector/config"
	"lite-collector/db"
	"lite-collector/middleware"
	"lite-collector/repository"
	"lite-collector/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	dataSourceName := cfg.Database.User + ":" + cfg.Database.Password + "@tcp(" + cfg.Database.Host + ":" + cfg.Database.Port + ")/" + cfg.Database.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db.Init(dataSourceName)

	// Initialize Gin router
	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.GetDB())
	formRepo := repository.NewFormRepository(db.GetDB())
	submissionRepo := repository.NewSubmissionRepository(db.GetDB())

	// Health check endpoint (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (no auth required)
		routes.RegisterAuthRoutes(v1, userRepo)

		// Protected routes (require auth)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware()) // Apply auth middleware
		{
			routes.RegisterFormRoutes(protected, formRepo)
			// Submission routes are now handled within form routes as nested routes
			// routes.RegisterSubmissionRoutes(protected) // Removed to avoid conflicts
		}
	}

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}