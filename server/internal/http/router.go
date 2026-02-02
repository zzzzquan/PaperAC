package http

// 路由注册与中间件组装（无登录认证版本）。

import (
	"net/http"
	"strings"

	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/handlers"
	"aigc-detector/server/internal/http/middleware"
	"aigc-detector/server/internal/store"
	"aigc-detector/server/internal/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, db *store.Store, workerService *worker.Worker) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.MaxMultipartMemory = int64(cfg.MaxUploadMB) * 1024 * 1024
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(corsConfig(cfg)))

	// Static File Serving (Monolithic)
	// Serve "./dist" directory at root "/"
	router.Use(static.Serve("/", static.LocalFile("./dist", true)))

	// SPA Fallback: If not API, serve index.html
	router.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.File("./dist/index.html")
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
		}
	})

	taskHandler := &handlers.TaskHandler{Store: db, Worker: workerService, Config: cfg}

	api := router.Group("/api")

	// 任务路由（无需认证）
	tasks := api.Group("/tasks")
	tasks.POST("", taskHandler.CreateTask)
	tasks.GET("", taskHandler.ListTasks)
	tasks.GET("/:id", taskHandler.GetTask)
	tasks.GET("/:id/result", taskHandler.DownloadResult)
	tasks.DELETE("/:id", taskHandler.CancelTask)

	// 会话清理
	api.DELETE("/session", taskHandler.ClearSession)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func corsConfig(cfg config.Config) cors.Config {
	origins := cfg.CORSAllowOrigins
	// In monolithic mode, same-origin is default, but for dev we keep localhost
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}

	return cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "X-Request-Id", "X-Session-ID"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: true,
	}
}
