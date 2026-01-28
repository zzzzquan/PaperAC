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

type EmailVerification struct {
  ID           uuid.UUID
  Email        string
  CodeHash     string
  ExpiresAt    time.Time
  AttemptCount int
  ConsumedAt   *time.Time
  RequestIP    string
  CreatedAt    time.Time
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

func (s *Store) CreateEmailVerification(ctx context.Context, verification EmailVerification) error {
  _, err := s.DB.Exec(ctx, `
    insert into email_verifications (
      id, email, code_hash, expires_at, attempt_count, consumed_at, request_ip, created_at
    ) values ($1,$2,$3,$4,$5,$6,$7,$8)
  `,
    verification.ID,
    verification.Email,
    verification.CodeHash,
    verification.ExpiresAt,
    verification.AttemptCount,
    verification.ConsumedAt,
    verification.RequestIP,
    verification.CreatedAt,
  )
  return err
}

func (s *Store) LatestVerificationByEmail(ctx context.Context, email string) (*EmailVerification, error) {
  row := s.DB.QueryRow(ctx, `
    select id, email, code_hash, expires_at, attempt_count, consumed_at, request_ip, created_at
    from email_verifications
    where email = $1
    order by created_at desc
    limit 1
  `, email)

  var v EmailVerification
  err := row.Scan(&v.ID, &v.Email, &v.CodeHash, &v.ExpiresAt, &v.AttemptCount, &v.ConsumedAt, &v.RequestIP, &v.CreatedAt)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return nil, nil
    }
    return nil, err
  }
  return &v, nil
}

func (s *Store) IncrementVerificationAttempt(ctx context.Context, id uuid.UUID) (int, error) {
  var attempt int
  err := s.DB.QueryRow(ctx, `
    update email_verifications
    set attempt_count = attempt_count + 1
    where id = $1
    returning attempt_count
  `, id).Scan(&attempt)
  return attempt, err
}

func (s *Store) ConsumeVerification(ctx context.Context, id uuid.UUID) error {
  _, err := s.DB.Exec(ctx, `
    update email_verifications
    set consumed_at = $1
    where id = $2
  `, time.Now().UTC(), id)
  return err
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (*User, error) {
  row := s.DB.QueryRow(ctx, `
    select id, email, created_at
    from users
    where email = $1
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

func (s *Store) CreateUser(ctx context.Context, email string) (*User, error) {
  user := &User{
    ID:        uuid.New(),
    Email:     email,
    CreatedAt: time.Now().UTC(),
  }
  _, err := s.DB.Exec(ctx, `
    insert into users (id, email, created_at)
    values ($1, $2, $3)
  `, user.ID, user.Email, user.CreatedAt)
  if err != nil {
    return nil, err
  }
  return user, nil
}

func (s *Store) FindUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
  row := s.DB.QueryRow(ctx, `
    select id, email, created_at
    from users
    where id = $1
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
