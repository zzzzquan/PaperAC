package config

// 配置加载：从环境变量读取运行参数。

import (
  "os"
  "strconv"
  "strings"
  "time"
)

type Config struct {
  BindAddress string
  Environment string

  DatabaseDSN string
  RedisAddr   string
  RedisPass   string
  RedisDB     int

  SessionTTL        time.Duration
  SessionCookieName string
  SessionCookieDomain string
  CookieSecure      bool

  UploadDir         string
  ResultDir         string
  MaxUploadMB       int
  TaskTimeout       time.Duration
  WorkerConcurrency int
  TaskQueueName     string

  DirectMailEndpoint   string
  DirectMailAccessKey  string
  DirectMailSecret     string
  DirectMailAccount    string
  DirectMailTemplate   string
  MailerMode           string

  CORSAllowOrigins []string
}

func Load() Config {
  return Config{
    BindAddress:         getEnv("BIND_ADDR", ":8080"),
    Environment:         getEnv("APP_ENV", "development"),
    DatabaseDSN:         getEnv("DATABASE_DSN", ""),
    RedisAddr:           getEnv("REDIS_ADDR", "127.0.0.1:6379"),
    RedisPass:           getEnv("REDIS_PASSWORD", ""),
    RedisDB:             getEnvInt("REDIS_DB", 0),
    SessionTTL:          getEnvDuration("SESSION_TTL", 30*24*time.Hour),
    SessionCookieName:   getEnv("SESSION_COOKIE", "sid"),
    SessionCookieDomain: getEnv("COOKIE_DOMAIN", ""),
    CookieSecure:        getEnvBool("COOKIE_SECURE", false),
    UploadDir:           getEnv("UPLOAD_DIR", "/var/app/uploads"),
    ResultDir:           getEnv("RESULT_DIR", "/var/app/results"),
    MaxUploadMB:         getEnvInt("MAX_UPLOAD_MB", 50),
    TaskTimeout:         getEnvDuration("TASK_TIMEOUT", 5*time.Minute),
    WorkerConcurrency:   getEnvInt("WORKER_CONCURRENCY", 0),
    TaskQueueName:       getEnv("TASK_QUEUE_NAME", "queue:tasks"),
    DirectMailEndpoint:  getEnv("DIRECTMAIL_ENDPOINT", ""),
    DirectMailAccessKey: getEnv("DIRECTMAIL_ACCESS_KEY", ""),
    DirectMailSecret:    getEnv("DIRECTMAIL_ACCESS_SECRET", ""),
    DirectMailAccount:   getEnv("DIRECTMAIL_ACCOUNT_NAME", ""),
    DirectMailTemplate:  getEnv("DIRECTMAIL_TEMPLATE_NAME", ""),
    MailerMode:          getEnv("MAILER_MODE", "directmail"),
    CORSAllowOrigins:    splitCSV(getEnv("CORS_ALLOW_ORIGINS", "")),
  }
}

func getEnv(key, fallback string) string {
  if val := os.Getenv(key); val != "" {
    return val
  }
  return fallback
}

func getEnvInt(key string, fallback int) int {
  val := os.Getenv(key)
  if val == "" {
    return fallback
  }
  parsed, err := strconv.Atoi(val)
  if err != nil {
    return fallback
  }
  return parsed
}

func getEnvBool(key string, fallback bool) bool {
  val := os.Getenv(key)
  if val == "" {
    return fallback
  }
  return val == "1" || strings.EqualFold(val, "true") || strings.EqualFold(val, "yes")
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
  val := os.Getenv(key)
  if val == "" {
    return fallback
  }
  parsed, err := time.ParseDuration(val)
  if err != nil {
    return fallback
  }
  return parsed
}

func splitCSV(value string) []string {
  if value == "" {
    return []string{}
  }
  parts := strings.Split(value, ",")
  for i, part := range parts {
    parts[i] = strings.TrimSpace(part)
  }
  return parts
}
