package config

// 配置加载：从环境变量读取运行参数。
// 简化版：移除认证相关配置。

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BindAddress string
	Environment string

	// Database
	DBPath string // SQLite 文件路径

	UploadDir         string
	ResultDir         string
	MaxUploadMB       int
	TaskTimeout       time.Duration
	WorkerConcurrency int

	CORSAllowOrigins []string
}

func Load() Config {
	return Config{
		BindAddress:       getBindAddress(),
		Environment:       getEnv("APP_ENV", "development"),
		DBPath:            getEnv("DB_DSN", "./paperac.db"),      // 默认本地文件路径
		UploadDir:         getEnv("UPLOAD_DIR", "./tmp/uploads"), // 默认本地上传目录
		ResultDir:         getEnv("RESULT_DIR", "./tmp/results"), // 默认本地结果目录
		MaxUploadMB:       getEnvInt("MAX_UPLOAD_MB", 50),
		TaskTimeout:       getEnvDuration("TASK_TIMEOUT", 10*time.Minute),
		WorkerConcurrency: getEnvInt("WORKER_CONCURRENCY", 5),
		CORSAllowOrigins:  splitCSV(getEnv("CORS_ALLOW_ORIGINS", "")),
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
