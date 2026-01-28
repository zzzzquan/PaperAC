package middleware

// CSRF校验中间件：仅对写操作校验X-CSRF-Token。

import (
  "net/http"
  "strings"

  "aigc-detector/server/internal/auth"
  "aigc-detector/server/internal/config"
  "aigc-detector/server/internal/util"

  "github.com/gin-gonic/gin"
)

func CSRF(service *auth.Service, cfg config.Config) gin.HandlerFunc {
  return func(c *gin.Context) {
    if isSafeMethod(c.Request.Method) {
      c.Next()
      return
    }

    sessionID, _ := c.Cookie(cfg.SessionCookieName)
    if sessionID == "" {
      util.JSON(c, http.StatusUnauthorized, 2001, "未登录", nil)
      c.Abort()
      return
    }

    sessionData, err := service.GetSession(c.Request.Context(), sessionID)
    if err != nil || sessionData == nil {
      util.JSON(c, http.StatusUnauthorized, 2001, "未登录", nil)
      c.Abort()
      return
    }

    token := c.GetHeader("X-CSRF-Token")
    if token == "" || token != sessionData.CSRFToken {
      util.JSON(c, http.StatusForbidden, 2002, "CSRF校验失败", nil)
      c.Abort()
      return
    }

    c.Next()
  }
}

func isSafeMethod(method string) bool {
  return strings.EqualFold(method, http.MethodGet) || strings.EqualFold(method, http.MethodHead) || strings.EqualFold(method, http.MethodOptions)
}
