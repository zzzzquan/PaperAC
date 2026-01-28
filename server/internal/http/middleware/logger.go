package middleware

// 日志中间件：记录路径、状态码、耗时与request_id。

import (
  "log"
  "time"

  "github.com/gin-gonic/gin"

  "aigc-detector/server/internal/util"
)

func Logger() gin.HandlerFunc {
  return func(c *gin.Context) {
    start := time.Now()
    c.Next()
    latency := time.Since(start)
    requestID := util.RequestIDFromContext(c)
    log.Printf("request_id=%s method=%s path=%s status=%d latency=%s", requestID, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
  }
}
