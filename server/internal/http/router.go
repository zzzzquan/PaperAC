package http

// 路由注册与中间件组装。

import (
	"net/http"
	"strings"

	"aigc-detector/server/internal/auth"
	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/handlers"
	"aigc-detector/server/internal/http/middleware"
	"aigc-detector/server/internal/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
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

	// Static File Serving (Monolithic)
	// Serve "./dist" directory at root "/"
	// Make sure this is AFTER CORS if needed, or BEFORE if public
	router.Use(static.Serve("/", static.LocalFile("./dist", true)))

	// SPA Fallback: If not API, serve index.html
	router.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.File("./dist/index.html")
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
		}
	})

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
	// In monolithic mode, same-origin is default, but for dev we keep localhost
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}

	return cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "X-Request-Id", "Authorization"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: true,
	}
}
