// @title           Lite Collector 接口文档
// @version         1.0
// @description     轻量级智能数据收集平台后端接口。所有接口（除登录外）需在请求头携带 JWT token，格式：Bearer <token>。
// @description
// @description     **状态码说明（提交记录 status）：** 0=待检测 1=正常 2=有异常
// @description     **状态码说明（表单 status）：** 0=草稿 1=已发布 2=已归档

// @contact.name   Lite Collector Team

// @license.name  MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 登录后获取的 JWT token。在 Swagger UI 中直接粘贴 token 即可（无需加 Bearer 前缀）；实际客户端请求格式为：Bearer <token>

package main

import (
	"lite-collector/config"
	"lite-collector/db"
	_ "lite-collector/docs" // swag generated docs
	"lite-collector/jobs"
	"lite-collector/middleware"
	"lite-collector/repository"
	"lite-collector/routes"
	"lite-collector/services"
	"log"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Build DSN using the driver's config struct so special characters in
	// the password (e.g. @, #) are handled correctly.
	mysqlCfg := mysqlDriver.NewConfig()
	mysqlCfg.User = cfg.Database.User
	mysqlCfg.Passwd = cfg.Database.Password
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = cfg.Database.Host + ":" + cfg.Database.Port
	mysqlCfg.DBName = cfg.Database.DBName
	mysqlCfg.ParseTime = true
	mysqlCfg.Params = map[string]string{"charset": "utf8mb4", "loc": "Local"}
	db.Init(mysqlCfg.FormatDSN())

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
	baseDataRepo := repository.NewBaseDataRepository(db.GetDB())

	// Services
	userService := services.NewUserService(userRepo, jwtSecret, cfg.Wechat.AppID, cfg.Wechat.AppSecret)
	formService := services.NewFormService(formRepo)
	submissionService := services.NewSubmissionService(submissionRepo, aiJobRepo)
	aiJobService := services.NewAIJobService(aiJobRepo)
	baseDataService := services.NewBaseDataService(baseDataRepo)

	// Start anomaly detection worker if DeepSeek API key is configured
	if cfg.DeepSeek.APIKey != "" {
		deepseekClient := services.NewDeepSeekClient(cfg.DeepSeek.APIKey)
		anomalyWorker := jobs.NewAnomalyWorker(aiJobRepo, submissionRepo, formRepo, deepseekClient)
		anomalyWorker.Start()
	} else {
		log.Println("DEEPSEEK_API_KEY not set — anomaly detection worker disabled")
	}

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
			routes.RegisterUserRoutes(protected, userService)
			routes.RegisterFormRoutes(protected, formService, submissionService, baseDataService)
			routes.RegisterJobRoutes(protected, aiJobService)
		}
	}

	addr := ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	log.Printf("Swagger UI available at http://localhost%s/swagger/index.html", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
