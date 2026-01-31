package auth

// 认证HTTP处理器。

import (
	"net/http"
	"strings"

	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/util"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
	Config  config.Config
}

type SendCodeRequest struct {
	Email string `json:"email"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (h *Handler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.JSON(c, http.StatusBadRequest, 1001, "参数错误", nil)
		return
	}

	ip := NormalizeIP(c.ClientIP())
	if err := h.Service.SendCode(c.Request.Context(), req.Email, ip); err != nil {
		code, message := ErrorCode(err)
		util.JSON(c, http.StatusBadRequest, code, message, nil)
		return
	}

	util.OK(c, nil)
}

func (h *Handler) Verify(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.JSON(c, http.StatusBadRequest, 1001, "参数错误", nil)
		return
	}

	token, user, err := h.Service.Verify(c.Request.Context(), req.Email, strings.TrimSpace(req.Code))
	if err != nil {
		code, message := ErrorCode(err)
		util.JSON(c, http.StatusBadRequest, code, message, nil)
		return
	}

	// 返回 JWT Token，客户端自行存储 (localStorage / Cookie)
	// 也可以由服务端 Set-Cookie (httpOnly)，但这里返回 JSON 更灵活，适应 Serverless/Mobile
	util.OK(c, gin.H{
		"token":   token,
		"user_id": user.ID.String(),
		"email":   user.Email,
	})
}

func (h *Handler) Me(c *gin.Context) {
	user, err := CurrentUser(c)
	if err != nil {
		util.JSON(c, http.StatusUnauthorized, 2001, "未登录", nil)
		return
	}
	util.OK(c, gin.H{
		"user_id": user.UserID,
		"email":   user.Email,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	// Stateless JWT Logout:
	// Client should discard the token.
	// Server can't force logout without a blacklist (Redis).
	// For MVP Serverless, we do nothing.
	util.OK(c, nil)
}
