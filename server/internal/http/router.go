package http

// 路由注册与中间件组装。

import (
	"net/http"

	"aigc-detector/server/internal/auth"
	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/handlers"
	"aigc-detector/server/internal/http/middleware"
	"aigc-detector/server/internal/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, authService *auth.Service, worker *worker.Worker) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.MaxMultipartMemory = int64(cfg.MaxUploadMB) * 1024 * 1024
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(corsConfig(cfg)))

	handler := &auth.Handler{Service: authService, Config: cfg}
	taskHandler := &handlers.TaskHandler{Store: authService.Store(), Worker: worker, Config: cfg}

	api := router.Group("/api")
	authGroup := api.Group("/auth")
	authGroup.POST("/send-code", handler.SendCode)
	authGroup.POST("/verify", handler.Verify)

	protected := authGroup.Group("")
	protected.Use(middleware.JWTAuth(cfg))
	protected.GET("/me", handler.Me)
	protected.POST("/logout", handler.Logout)

	tasks := api.Group("/tasks")
	tasks.Use(middleware.JWTAuth(cfg))
	tasks.POST("", taskHandler.CreateTask)
	tasks.GET("", taskHandler.ListTasks)
	tasks.GET("/:id", taskHandler.GetTask)
	tasks.GET("/:id/result", taskHandler.DownloadResult)
	tasks.DELETE("/:id", taskHandler.CancelTask)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func corsConfig(cfg config.Config) cors.Config {
	origins := cfg.CORSAllowOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}

	return cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "X-Request-Id", "Authorization"}, // Added Authorization
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: true,
	}
}
