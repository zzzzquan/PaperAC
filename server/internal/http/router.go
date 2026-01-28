package http

// 路由注册与中间件组装。

import (
  "net/http"

  "aigc-detector/server/internal/auth"
  "aigc-detector/server/internal/config"
  "aigc-detector/server/internal/handlers"
  "aigc-detector/server/internal/http/middleware"

  "github.com/gin-contrib/cors"
  "github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, authService *auth.Service) *gin.Engine {
  gin.SetMode(gin.ReleaseMode)
  router := gin.New()

  router.MaxMultipartMemory = int64(cfg.MaxUploadMB) * 1024 * 1024
  router.Use(middleware.RequestID())
  router.Use(middleware.Logger())
  router.Use(gin.Recovery())
  router.Use(cors.New(corsConfig(cfg)))

  handler := &auth.Handler{Service: authService, Config: cfg}
  taskHandler := &handlers.TaskHandler{Store: authService.Store(), Redis: authService.Redis(), Config: cfg}

  api := router.Group("/api")
  authGroup := api.Group("/auth")
  authGroup.POST("/send-code", handler.SendCode)
  authGroup.POST("/verify", handler.Verify)

  protected := authGroup.Group("")
  protected.Use(middleware.SessionAuth(authService, cfg))
  protected.Use(middleware.CSRF(authService, cfg))
  protected.GET("/me", handler.Me)
  protected.POST("/logout", handler.Logout)

  tasks := api.Group("/tasks")
  tasks.Use(middleware.SessionAuth(authService, cfg))
  tasks.Use(middleware.CSRF(authService, cfg))
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
    AllowMethods:     []string{"GET", "POST", "OPTIONS"},
    AllowHeaders:     []string{"Content-Type", "X-Request-Id", "X-CSRF-Token"},
    ExposeHeaders:    []string{"X-Request-Id"},
    AllowCredentials: true,
  }
}
