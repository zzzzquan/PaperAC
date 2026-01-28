package util

// 统一响应结构与输出工具。

import (
  "net/http"

  "github.com/gin-gonic/gin"
)

type Response struct {
  Code      int         `json:"code"`
  Message   string      `json:"message"`
  RequestID string      `json:"request_id"`
  Data      interface{} `json:"data,omitempty"`
}

func JSON(c *gin.Context, status int, code int, message string, data interface{}) {
  c.JSON(status, Response{
    Code:      code,
    Message:   message,
    RequestID: RequestIDFromContext(c),
    Data:      data,
  })
}

func OK(c *gin.Context, data interface{}) {
  JSON(c, http.StatusOK, 0, "ok", data)
}
