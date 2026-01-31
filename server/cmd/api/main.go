package main

// 应用入口：加载配置、初始化依赖并启动HTTP服务。

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"aigc-detector/server/internal/auth"
	"aigc-detector/server/internal/config"
	httpserver "aigc-detector/server/internal/http"
	"aigc-detector/server/internal/store"
	"aigc-detector/server/internal/worker"
)

func main() {
	cfg := config.Load()

	db, err := store.NewPostgres(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	// 阿里云 DirectMail
	mailer := auth.NewDirectMailClient(cfg)
	authService := auth.NewService(db, mailer, cfg)

	if err := ensureDirs(cfg); err != nil {
		log.Fatalf("初始化存储目录失败: %v", err)
	}

	workerService := &worker.Worker{Store: db, Config: cfg}
	router := httpserver.NewRouter(cfg, authService, workerService)

	server := &http.Server{
		Addr:         cfg.BindAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("HTTP服务启动: %s", cfg.BindAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务启动失败: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP服务关闭失败: %v", err)
	}
}

func ensureDirs(cfg config.Config) error {
	if err := os.MkdirAll(filepath.Clean(cfg.UploadDir), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Clean(cfg.ResultDir), 0o755); err != nil {
		return err
	}
	return nil
}
