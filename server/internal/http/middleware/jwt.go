package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"aigc-detector/server/internal/auth"
	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth 鉴权中间件
func JWTAuth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			util.JSON(c, http.StatusUnauthorized, 401, "未通过身份认证", nil)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			log.Printf("JWT Auth Failed: err=%v, valid=%v, token=%s", err, token != nil && token.Valid, tokenString)
			util.JSON(c, http.StatusUnauthorized, 401, "无效的令牌", nil)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			util.JSON(c, http.StatusUnauthorized, 401, "无效的令牌声明", nil)
			c.Abort()
			return
		}

		// 提取用户信息
		userID, _ := claims["uid"].(string)
		email, _ := claims["email"].(string)

		user := auth.UserView{
			UserID: userID,
			Email:  email,
		}

		auth.SetCurrentUser(c, user)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// 1. Try Authorization header: Bearer <token>
	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) > 7 && strings.ToUpper(bearerToken[0:7]) == "BEARER " {
		return bearerToken[7:]
	}
	// 2. Try Cookie (optional, for compatibility)
	cookie, err := c.Cookie("token")
	if err == nil {
		return cookie
	}
	return ""
}
