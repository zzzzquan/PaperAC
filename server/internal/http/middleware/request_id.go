package middleware

// request_id 中间件：优先使用客户端传入的X-Request-Id。

import (
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"

  "aigc-detector/server/internal/util"
)

func RequestID() gin.HandlerFunc {
  return func(c *gin.Context) {
    requestID := c.GetHeader("X-Request-Id")
    if requestID == "" {
      requestID = uuid.NewString()
    }
    util.SetRequestID(c, requestID)
    c.Writer.Header().Set("X-Request-Id", requestID)
    c.Next()
  }
}
