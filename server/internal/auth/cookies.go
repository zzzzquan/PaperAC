package auth

// 会话与CSRF Cookie 管理。

import (
  "net/http"

  "aigc-detector/server/internal/config"

  "github.com/gin-gonic/gin"
)

const csrfCookieName = "csrf_token"

func setSessionCookie(c *gin.Context, cfg config.Config, sessionID string) {
  http.SetCookie(c.Writer, &http.Cookie{
    Name:     cfg.SessionCookieName,
    Value:    sessionID,
    MaxAge:   int(cfg.SessionTTL.Seconds()),
    Path:     "/",
    Domain:   cfg.SessionCookieDomain,
    Secure:   cfg.CookieSecure,
    HttpOnly: true,
    SameSite: http.SameSiteLaxMode,
  })
}

func clearSessionCookie(c *gin.Context, cfg config.Config) {
  http.SetCookie(c.Writer, &http.Cookie{
    Name:     cfg.SessionCookieName,
    Value:    "",
    MaxAge:   -1,
    Path:     "/",
    Domain:   cfg.SessionCookieDomain,
    Secure:   cfg.CookieSecure,
    HttpOnly: true,
    SameSite: http.SameSiteLaxMode,
  })
}

func setCSRFCookie(c *gin.Context, cfg config.Config, token string) {
  http.SetCookie(c.Writer, &http.Cookie{
    Name:     csrfCookieName,
    Value:    token,
    MaxAge:   int(cfg.SessionTTL.Seconds()),
    Path:     "/",
    Domain:   cfg.SessionCookieDomain,
    Secure:   cfg.CookieSecure,
    HttpOnly: false,
    SameSite: http.SameSiteLaxMode,
  })
}

func clearCSRFCookie(c *gin.Context, cfg config.Config) {
  http.SetCookie(c.Writer, &http.Cookie{
    Name:     csrfCookieName,
    Value:    "",
    MaxAge:   -1,
    Path:     "/",
    Domain:   cfg.SessionCookieDomain,
    Secure:   cfg.CookieSecure,
    HttpOnly: false,
    SameSite: http.SameSiteLaxMode,
  })
}
