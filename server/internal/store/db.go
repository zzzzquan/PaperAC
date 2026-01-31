package store

// 数据库访问层：负责PostgreSQL连接与核心查询。

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

type User struct {
	ID        uuid.UUID
	Email     string
	CreatedAt time.Time
}

type VerificationCode struct {
	ID        int
	Email     string
	Code      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewPostgres(dsn string) (*Store, error) {
	if dsn == "" {
		return nil, errors.New("DATABASE_DSN 未配置")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &Store{DB: pool}, nil
}

func (s *Store) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}

// SaveVerificationCode 保存验证码
func (s *Store) SaveVerificationCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO verification_codes (email, code, expires_at)
		VALUES ($1, $2, $3)
	`, email, code, expiresAt)
	return err
}

// GetVerificationCode 获取有效的验证码
func (s *Store) GetVerificationCode(ctx context.Context, email, code string) (*VerificationCode, error) {
	row := s.DB.QueryRow(ctx, `
		SELECT id, email, code, expires_at, created_at
		FROM verification_codes
		WHERE email = $1 AND code = $2 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`, email, code)

	var v VerificationCode
	err := row.Scan(&v.ID, &v.Email, &v.Code, &v.ExpiresAt, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

// DeleteVerificationCode 删除验证码（防重放）
func (s *Store) DeleteVerificationCode(ctx context.Context, id int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM verification_codes WHERE id = $1`, id)
	return err
}

// FindUserByEmail 查找用户
func (s *Store) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	row := s.DB.QueryRow(ctx, `
		SELECT id, email, created_at
		FROM users
		WHERE email = $1
	`, email)

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// CreateUser 创建用户
func (s *Store) CreateUser(ctx context.Context, email string) (*User, error) {
	user := &User{
		ID:        uuid.New(),
		Email:     email,
		CreatedAt: time.Now().UTC(),
	}
	_, err := s.DB.Exec(ctx, `
		INSERT INTO users (id, email, created_at)
		VALUES ($1, $2, $3)
	`, user.ID, user.Email, user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindUserByID 根据ID查找用户
func (s *Store) FindUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := s.DB.QueryRow(ctx, `
		SELECT id, email, created_at
		FROM users
		WHERE id = $1
	`, id)

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
