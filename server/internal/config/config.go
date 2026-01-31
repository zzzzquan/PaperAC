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
	JWTSecret   string

	UploadDir         string
	ResultDir         string
	MaxUploadMB       int
	TaskTimeout       time.Duration
	WorkerConcurrency int

	DirectMailEndpoint  string
	DirectMailAccessKey string
	DirectMailSecret    string
	DirectMailAccount   string
	DirectMailTemplate  string
	MailerMode          string

	CORSAllowOrigins []string
}

func Load() Config {
	return Config{
		BindAddress:         getEnv("BIND_ADDR", ":8080"),
		Environment:         getEnv("APP_ENV", "development"),
		DatabaseDSN:         getEnv("DATABASE_DSN", ""),
		JWTSecret:           getEnv("JWT_SECRET", "change-me-in-prod"),
		UploadDir:           getEnv("UPLOAD_DIR", "/var/app/uploads"),
		ResultDir:           getEnv("RESULT_DIR", "/var/app/results"),
		MaxUploadMB:         getEnvInt("MAX_UPLOAD_MB", 50),
		TaskTimeout:         getEnvDuration("TASK_TIMEOUT", 5*time.Minute),
		WorkerConcurrency:   getEnvInt("WORKER_CONCURRENCY", 5),
		DirectMailEndpoint:  getEnv("DIRECTMAIL_ENDPOINT", "dm.aliyuncs.com"),
		DirectMailAccessKey: getEnv("ALIYUN_ACCESS_KEY", ""),
		DirectMailSecret:    getEnv("ALIYUN_SECRET_KEY", ""),
		DirectMailAccount:   getEnv("DM_ACCOUNT_NAME", "admin@mail.paperac.com"),
		DirectMailTemplate:  getEnv("DIRECTMAIL_TEMPLATE_NAME", ""),
		MailerMode:          getEnv("MAILER_MODE", "mock"), // mock or directmail
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
