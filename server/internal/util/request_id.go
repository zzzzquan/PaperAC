package util

// request_id 在中间件中注入，统一从上下文读取。

import "github.com/gin-gonic/gin"

const requestIDKey = "request_id"

func SetRequestID(c *gin.Context, value string) {
  c.Set(requestIDKey, value)
}

func RequestIDFromContext(c *gin.Context) string {
  if val, ok := c.Get(requestIDKey); ok {
    if str, ok := val.(string); ok {
      return str
    }
  }
  return ""
}
