package auth

// 邮件发送客户端：封装DirectMail调用，便于替换与测试。

import (
  "bytes"
  "context"
  "encoding/json"
  "errors"
  "log"
  "net/http"
  "time"

  "aigc-detector/server/internal/config"
)

type Mailer interface {
  SendVerificationCode(ctx context.Context, email string, code string) error
}

type DirectMailClient struct {
  endpoint   string
  accessKey  string
  secret     string
  account    string
  template   string
  httpClient *http.Client
}

func NewDirectMailClient(cfg config.Config) Mailer {
  if cfg.MailerMode == "mock" {
    return &MockMailer{}
  }
  return &DirectMailClient{
    endpoint:   cfg.DirectMailEndpoint,
    accessKey:  cfg.DirectMailAccessKey,
    secret:     cfg.DirectMailSecret,
    account:    cfg.DirectMailAccount,
    template:   cfg.DirectMailTemplate,
    httpClient: &http.Client{Timeout: 10 * time.Second},
  }
}

func (c *DirectMailClient) SendVerificationCode(ctx context.Context, email string, code string) error {
  if c.endpoint == "" || c.accessKey == "" || c.secret == "" || c.account == "" || c.template == "" {
    return errors.New("DirectMail 配置不完整")
  }

  payload := map[string]interface{}{
    "account_name":  c.account,
    "template_name": c.template,
    "to_address":    email,
    "template_params": map[string]string{
      "code": code,
    },
  }

  body, err := json.Marshal(payload)
  if err != nil {
    return err
  }

  req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
  if err != nil {
    return err
  }
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("X-Access-Key", c.accessKey)
  req.Header.Set("X-Access-Secret", c.secret)

  resp, err := c.httpClient.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    return errors.New("DirectMail 发送失败")
  }

  return nil
}

type MockMailer struct{}

func (m *MockMailer) SendVerificationCode(ctx context.Context, email string, code string) error {
  log.Printf("模拟发送验证码: email=%s code=%s", email, code)
  return nil
}
