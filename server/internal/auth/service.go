package auth

// 认证服务：验证码发送、校验与会话管理。

import (
  "context"
  "errors"
  "net"
  "strings"
  "time"

  "aigc-detector/server/internal/config"
  "aigc-detector/server/internal/session"
  "aigc-detector/server/internal/store"
  "aigc-detector/server/internal/util"

  "github.com/google/uuid"
)

const (
  CodeLength              = 6
  CodeTTL                 = 5 * time.Minute
  EmailCooldown           = 60 * time.Second
  IPMaxPerHour            = 20
  MaxVerificationAttempts = 5
)

type Service struct {
  store   *store.Store
  redis   *store.RedisStore
  mailer  Mailer
  cfg     config.Config
}

func NewService(store *store.Store, redis *store.RedisStore, mailer Mailer, cfg config.Config) *Service {
  return &Service{store: store, redis: redis, mailer: mailer, cfg: cfg}
}

func (s *Service) Store() *store.Store {
  return s.store
}

func (s *Service) Redis() *store.RedisStore {
  return s.redis
}

type VerificationResult struct {
  UserID string
  Email  string
}

func (s *Service) SendCode(ctx context.Context, email string, ip string) error {
  email = strings.TrimSpace(strings.ToLower(email))
  if email == "" || !strings.Contains(email, "@") {
    return ErrInvalidParams
  }

  if _, err := net.ResolveIPAddr("ip", ip); err != nil {
    return ErrInvalidParams
  }

  ok, err := s.redis.EmailCooldown(ctx, email, EmailCooldown)
  if err != nil {
    return err
  }
  if !ok {
    return ErrRateLimited
  }

  hourKey := time.Now().UTC().Format("2006010215")
  count, err := s.redis.IncrementIPRate(ctx, ip, hourKey, 2*time.Hour)
  if err != nil {
    return err
  }
  if count > IPMaxPerHour {
    return ErrRateLimited
  }

  code, err := util.GenerateNumericCode(CodeLength)
  if err != nil {
    return err
  }

  verification := store.EmailVerification{
    ID:           uuid.New(),
    Email:        email,
    CodeHash:     util.HashCode(code, email),
    ExpiresAt:    time.Now().UTC().Add(CodeTTL),
    AttemptCount: 0,
    ConsumedAt:   nil,
    RequestIP:    ip,
    CreatedAt:    time.Now().UTC(),
  }

  if err := s.store.CreateEmailVerification(ctx, verification); err != nil {
    return err
  }

  if err := s.mailer.SendVerificationCode(ctx, email, code); err != nil {
    s.redis.ClearEmailCooldown(ctx, email)
    return ErrMailSendFailed
  }

  return nil
}

func (s *Service) Verify(ctx context.Context, email string, code string) (*VerificationResult, string, string, error) {
  email = strings.TrimSpace(strings.ToLower(email))
  if email == "" || code == "" {
    return nil, "", "", ErrInvalidParams
  }

  verification, err := s.store.LatestVerificationByEmail(ctx, email)
  if err != nil {
    return nil, "", "", err
  }
  if verification == nil {
    return nil, "", "", ErrCodeExpired
  }

  if verification.ConsumedAt != nil {
    return nil, "", "", ErrCodeExpired
  }
  if time.Now().UTC().After(verification.ExpiresAt) {
    return nil, "", "", ErrCodeExpired
  }
  if verification.AttemptCount >= MaxVerificationAttempts {
    return nil, "", "", ErrTooManyAttempts
  }

  expected := util.HashCode(code, email)
  if expected != verification.CodeHash {
    attempt, err := s.store.IncrementVerificationAttempt(ctx, verification.ID)
    if err != nil {
      return nil, "", "", err
    }
    if attempt >= MaxVerificationAttempts {
      return nil, "", "", ErrTooManyAttempts
    }
    return nil, "", "", ErrInvalidCode
  }

  if err := s.store.ConsumeVerification(ctx, verification.ID); err != nil {
    return nil, "", "", err
  }

  user, err := s.store.FindUserByEmail(ctx, email)
  if err != nil {
    return nil, "", "", err
  }
  if user == nil {
    user, err = s.store.CreateUser(ctx, email)
    if err != nil {
      return nil, "", "", err
    }
  }

  sessionID := uuid.NewString()
  csrfToken := uuid.NewString()
  sessionData := session.Data{
    UserID:    user.ID.String(),
    CSRFToken: csrfToken,
    CreatedAt: time.Now().UTC(),
    LastSeen:  time.Now().UTC(),
  }
  if err := s.redis.SetSession(ctx, sessionID, sessionData, s.cfg.SessionTTL); err != nil {
    return nil, "", "", err
  }

  return &VerificationResult{UserID: user.ID.String(), Email: user.Email}, sessionID, csrfToken, nil
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (*session.Data, error) {
  if sessionID == "" {
    return nil, nil
  }
  return s.redis.GetSession(ctx, sessionID)
}

func (s *Service) FindUser(ctx context.Context, userID string) (*store.User, error) {
  parsed, err := uuid.Parse(userID)
  if err != nil {
    return nil, ErrInvalidParams
  }
  return s.store.FindUserByID(ctx, parsed)
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
  if sessionID == "" {
    return nil
  }
  return s.redis.DeleteSession(ctx, sessionID)
}

var (
  ErrInvalidParams    = errors.New("参数错误")
  ErrInvalidCode      = errors.New("验证码错误")
  ErrCodeExpired      = errors.New("验证码已失效")
  ErrRateLimited      = errors.New("发送过于频繁")
  ErrMailSendFailed   = errors.New("邮件发送失败")
  ErrNotAuthenticated = errors.New("未登录")
  ErrTooManyAttempts  = errors.New("验证码错误次数过多")
)

func ErrorCode(err error) (int, string) {
  switch {
  case errors.Is(err, ErrInvalidParams):
    return 1001, err.Error()
  case errors.Is(err, ErrInvalidCode):
    return 1003, err.Error()
  case errors.Is(err, ErrRateLimited):
    return 1004, err.Error()
  case errors.Is(err, ErrCodeExpired):
    return 1005, err.Error()
  case errors.Is(err, ErrTooManyAttempts):
    return 1006, err.Error()
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
