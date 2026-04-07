// @title           Lite Collector API
// @version         1.0
// @description     Intelligent data collection platform — WeChat Mini Program backend.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Lite Collector Team

// @license.name  MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT token. Format: "Bearer <token>"

package main

import (
	"lite-collector/config"
	"lite-collector/db"
	_ "lite-collector/docs" // swag generated docs
	"lite-collector/middleware"
	"lite-collector/repository"
	"lite-collector/routes"
	"lite-collector/services"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	dsn := cfg.Database.User + ":" + cfg.Database.Password +
		"@tcp(" + cfg.Database.Host + ":" + cfg.Database.Port + ")/" +
		cfg.Database.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db.Init(dsn)

	// Initialize Gin router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	jwtSecret := []byte(cfg.JWT.Secret)

	// Repositories
	userRepo := repository.NewUserRepository(db.GetDB())
	formRepo := repository.NewFormRepository(db.GetDB())
	submissionRepo := repository.NewSubmissionRepository(db.GetDB())
	aiJobRepo := repository.NewAIJobRepository(db.GetDB())

	// Services
	userService := services.NewUserService(userRepo, jwtSecret)
	formService := services.NewFormService(formRepo)
	submissionService := services.NewSubmissionService(submissionRepo, aiJobRepo)

	// Health check (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := r.Group("/api/v1")
	{
		routes.RegisterAuthRoutes(v1, userService)

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			routes.RegisterFormRoutes(protected, formService, submissionService)
		}
	}

	addr := ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	log.Printf("Swagger UI available at http://localhost%s/swagger/index.html", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
