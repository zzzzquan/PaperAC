package store

// 数据库访问层：GORM + SQLite 实现。

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store struct {
	DB *gorm.DB
}

// Models
type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex"`
	CreatedAt time.Time
}

type VerificationCode struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"index"`
	Code      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Task struct {
	ID               string `gorm:"primaryKey"`
	UserID           string `gorm:"index"` // Foreign Key to User
	Status           TaskStatus
	Progress         int
	X                float64
	ErrorMessage     string
	OriginalFileName string
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

	// Auto Migrate
	if err := db.AutoMigrate(&User{}, &VerificationCode{}, &Task{}); err != nil {
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

// --- Verification Code ---

func (s *Store) SaveVerificationCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	v := VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: expiresAt,
	}
	// Use WithContext? GORM DB is thread safe, but for timeout we can use WithContext(ctx)
	return s.DB.WithContext(ctx).Create(&v).Error
}

func (s *Store) GetVerificationCode(ctx context.Context, email, code string) (*VerificationCode, error) {
	var v VerificationCode
	err := s.DB.WithContext(ctx).
		Where("email = ? AND code = ? AND expires_at > ?", email, code, time.Now()).
		Order("created_at desc").
		First(&v).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (s *Store) DeleteVerificationCode(ctx context.Context, id uint) error {
	return s.DB.WithContext(ctx).Delete(&VerificationCode{}, id).Error
}

// --- User ---

func (s *Store) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := s.DB.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil for not found (adapter behavior)
		}
		return nil, err
	}
	return &u, nil
}

func (s *Store) CreateUser(ctx context.Context, email string) (*User, error) {
	user := &User{
		ID:        uuid.New().String(),
		Email:     email,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.DB.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) FindUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := s.DB.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
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

func (s *Store) GetTaskForUser(ctx context.Context, taskID, userID string) (*Task, error) {
	var t Task
	err := s.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", taskID, userID).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (s *Store) ListTasksByUser(ctx context.Context, userID string, limit int) ([]Task, error) {
	var tasks []Task
	err := s.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (s *Store) UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus, progress int, resultPath, errorMsg string, finishedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"progress":   progress,
		"updated_at": time.Now().UTC(),
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
			"updated_at": time.Now().UTC(),
		}).Error
}

func (s *Store) CancelTask(ctx context.Context, taskID, userID string) (bool, error) {
	// Optimistic check: only pending tasks can be cancelled
	result := s.DB.WithContext(ctx).Model(&Task{}).
		Where("id = ? AND user_id = ? AND status = ?", taskID, userID, TaskPending).
		Update("status", TaskCancelled)

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
