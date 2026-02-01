package auth

// 认证服务：验证码发送、校验与JWT管理。

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/store"
	"aigc-detector/server/internal/util"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	CodeLength    = 6
	CodeTTL       = 5 * time.Minute
	EmailCooldown = 60 * time.Second // Simple cooldown
)

type Service struct {
	store  *store.Store
	mailer Mailer
	cfg    config.Config
}

func NewService(store *store.Store, mailer Mailer, cfg config.Config) *Service {
	return &Service{store: store, mailer: mailer, cfg: cfg}
}

func (s *Service) Store() *store.Store {
	return s.store
}

func (s *Service) SendCode(ctx context.Context, email string, ip string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return ErrInvalidParams
	}

	if _, err := net.ResolveIPAddr("ip", ip); err != nil {
		return ErrInvalidParams
	}

	// TODO: Rate limiting in DB (Optional for MVP)

	code, err := util.GenerateNumericCode(CodeLength)
	if err != nil {
		return err
	}

	// Save to DB
	expiresAt := time.Now().UTC().Add(CodeTTL)
	if err := s.store.SaveVerificationCode(ctx, email, code, expiresAt); err != nil {
		return err
	}

	if err := s.mailer.SendVerificationCode(ctx, email, code); err != nil {
		return ErrMailSendFailed
	}

	return nil
}

// Verify verifies code and returns (token, user, error)
func (s *Service) Verify(ctx context.Context, email string, code string) (string, *store.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || code == "" {
		return "", nil, ErrInvalidParams
	}

	vCode, err := s.store.GetVerificationCode(ctx, email, code)
	if err != nil {
		return "", nil, err
	}
	if vCode == nil {
		return "", nil, ErrInvalidCode
	}

	// Delete used code
	if err := s.store.DeleteVerificationCode(ctx, vCode.ID); err != nil {
		return "", nil, err
	}

	// Find or Create User
	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		user, err = s.store.CreateUser(ctx, email)
		if err != nil {
			return "", nil, err
		}
	}

	// Generate JWT
	now := time.Now()
	claims := jwt.MapClaims{
		"uid":   user.ID,
		"email": user.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

func (s *Service) FindUser(ctx context.Context, userID string) (*store.User, error) {
	parsed, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidParams
	}
	return s.store.FindUserByID(ctx, parsed.String())
}

// Error definitions
var (
	ErrInvalidParams    = errors.New("参数错误")
	ErrInvalidCode      = errors.New("验证码错误或已失效")
	ErrMailSendFailed   = errors.New("邮件发送失败")
	ErrNotAuthenticated = errors.New("未登录")
)

func ErrorCode(err error) (int, string) {
	switch {
	case errors.Is(err, ErrInvalidParams):
		return 1001, err.Error()
	case errors.Is(err, ErrInvalidCode):
		return 1003, err.Error()
	case errors.Is(err, ErrNotAuthenticated):
		return 2001, err.Error()
	case errors.Is(err, ErrMailSendFailed):
		return 9001, err.Error()
	default:
		return 9000, "系统错误"
	}
}

func NormalizeIP(raw string) string {
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, ",") {
		parts := strings.Split(raw, ",")
		raw = strings.TrimSpace(parts[0])
	}
	if strings.Contains(raw, ":") {
		host, _, err := net.SplitHostPort(raw)
		if err == nil {
			raw = host
		}
	}
	return raw
}
