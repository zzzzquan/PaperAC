package middleware

// 会话校验中间件：验证Cookie并注入当前用户。

import (
  "net/http"

  "aigc-detector/server/internal/auth"
  "aigc-detector/server/internal/config"
  "aigc-detector/server/internal/util"

  "github.com/gin-gonic/gin"
)

func SessionAuth(service *auth.Service, cfg config.Config) gin.HandlerFunc {
  return func(c *gin.Context) {
    sessionID, err := c.Cookie(cfg.SessionCookieName)
    if err != nil || sessionID == "" {
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

    user, err := service.FindUser(c.Request.Context(), sessionData.UserID)
    if err != nil || user == nil {
      util.JSON(c, http.StatusUnauthorized, 2001, "未登录", nil)
      c.Abort()
      return
    }

    auth.SetCurrentUser(c, auth.UserView{UserID: user.ID.String(), Email: user.Email})
    c.Next()
  }
}
