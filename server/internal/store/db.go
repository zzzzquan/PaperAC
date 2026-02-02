package store

// 数据库访问层：GORM + SQLite 实现。
// 简化版：仅保留 Task 模型用于日志记录。

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"aigc-detector/server/internal/util"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store struct {
	DB *gorm.DB
}

// Task 任务记录（用于日志追踪）
type Task struct {
	ID               string `gorm:"primaryKey"`
	Status           TaskStatus
	Progress         int
	X                float64
	ErrorMessage     string
	OriginalFileName string
	FileSize         int64 // 文件大小（字节）
	UploadPath       string
	ResultPath       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	FinishedAt       *time.Time
}

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskSuccess   TaskStatus = "success"
	TaskFailed    TaskStatus = "failed"
	TaskCancelled TaskStatus = "cancelled"
)

func NewSqlite(dbPath string) (*Store, error) {
	// Ensure db directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto Migrate - 仅 Task 表
	if err := db.AutoMigrate(&Task{}); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

func (s *Store) Close() {
	// GORM sql.DB generic close
	sqlDB, err := s.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// --- Task ---

func (s *Store) CreateTask(ctx context.Context, task *Task) error {
	// Ensure UUID string format
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	return s.DB.WithContext(ctx).Create(task).Error
}

func (s *Store) GetTaskByID(ctx context.Context, id string) (*Task, error) {
	var t Task
	err := s.DB.WithContext(ctx).First(&t, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// ListRecentTasks 获取最近的任务列表（不按用户过滤）
func (s *Store) ListRecentTasks(ctx context.Context, limit int) ([]Task, error) {
	var tasks []Task
	err := s.DB.WithContext(ctx).
		Order("created_at desc").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (s *Store) UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus, progress int, resultPath, errorMsg string, finishedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"progress":   progress,
		"updated_at": util.Now(),
	}
	// GORM ignores empty strings by default in struct updates, map is safer
	if resultPath != "" {
		updates["result_path"] = resultPath
	}
	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}
	if finishedAt != nil {
		updates["finished_at"] = finishedAt
	}

	return s.DB.WithContext(ctx).Model(&Task{}).Where("id = ?", taskID).Updates(updates).Error
}

func (s *Store) UpdateTaskProgress(ctx context.Context, taskID string, progress int) error {
	return s.DB.WithContext(ctx).Model(&Task{}).Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"progress":   progress,
			"updated_at": util.Now(),
		}).Error
}

// CancelTask 取消任务（仅 pending 状态可取消）
func (s *Store) CancelTask(ctx context.Context, taskID string) (bool, error) {
	result := s.DB.WithContext(ctx).Model(&Task{}).
		Where("id = ? AND status = ?", taskID, TaskPending).
		Update("status", TaskCancelled)

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
