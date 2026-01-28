package store

// 任务存储与状态管理。

import (
  "context"
  "errors"
  "time"

  "github.com/google/uuid"
  "github.com/jackc/pgx/v5"
)

type TaskStatus string

const (
  TaskPending   TaskStatus = "pending"
  TaskRunning   TaskStatus = "running"
  TaskSuccess   TaskStatus = "success"
  TaskFailed    TaskStatus = "failed"
  TaskCancelled TaskStatus = "cancelled"
)

type Task struct {
  ID              uuid.UUID
  UserID          uuid.UUID
  Status          TaskStatus
  Progress        int
  X               float64
  OriginalFileName string
  UploadPath      string
  ResultPath      string
  ErrorMessage    string
  CreatedAt       time.Time
  UpdatedAt       time.Time
  FinishedAt      *time.Time
}

func (s *Store) CreateTask(ctx context.Context, task *Task) error {
  _, err := s.DB.Exec(ctx, `
    insert into tasks (
      id, user_id, status, progress, x, original_filename, upload_path, result_path, error_message,
      created_at, updated_at, finished_at
    ) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
  `,
    task.ID,
    task.UserID,
    task.Status,
    task.Progress,
    task.X,
    task.OriginalFileName,
    task.UploadPath,
    task.ResultPath,
    task.ErrorMessage,
    task.CreatedAt,
    task.UpdatedAt,
    task.FinishedAt,
  )
  return err
}

func (s *Store) GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error) {
  row := s.DB.QueryRow(ctx, `
    select id, user_id, status, progress, x, original_filename, upload_path, result_path, error_message,
           created_at, updated_at, finished_at
    from tasks
    where id = $1
  `, id)

  var task Task
  err := row.Scan(
    &task.ID,
    &task.UserID,
    &task.Status,
    &task.Progress,
    &task.X,
    &task.OriginalFileName,
    &task.UploadPath,
    &task.ResultPath,
    &task.ErrorMessage,
    &task.CreatedAt,
    &task.UpdatedAt,
    &task.FinishedAt,
  )
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return nil, nil
    }
    return nil, err
  }
  return &task, nil
}

func (s *Store) GetTaskForUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Task, error) {
  row := s.DB.QueryRow(ctx, `
    select id, user_id, status, progress, x, original_filename, upload_path, result_path, error_message,
           created_at, updated_at, finished_at
    from tasks
    where id = $1 and user_id = $2
  `, id, userID)

  var task Task
  err := row.Scan(
    &task.ID,
    &task.UserID,
    &task.Status,
    &task.Progress,
    &task.X,
    &task.OriginalFileName,
    &task.UploadPath,
    &task.ResultPath,
    &task.ErrorMessage,
    &task.CreatedAt,
    &task.UpdatedAt,
    &task.FinishedAt,
  )
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return nil, nil
    }
    return nil, err
  }
  return &task, nil
}

func (s *Store) ListTasksByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Task, error) {
  if limit <= 0 {
    limit = 20
  }
  if limit > 100 {
    limit = 100
  }
  rows, err := s.DB.Query(ctx, `
    select id, user_id, status, progress, x, original_filename, upload_path, result_path, error_message,
           created_at, updated_at, finished_at
    from tasks
    where user_id = $1
    order by created_at desc
    limit $2
  `, userID, limit)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  tasks := make([]Task, 0)
  for rows.Next() {
    var task Task
    if err := rows.Scan(
      &task.ID,
      &task.UserID,
      &task.Status,
      &task.Progress,
      &task.X,
      &task.OriginalFileName,
      &task.UploadPath,
      &task.ResultPath,
      &task.ErrorMessage,
      &task.CreatedAt,
      &task.UpdatedAt,
      &task.FinishedAt,
    ); err != nil {
      return nil, err
    }
    tasks = append(tasks, task)
  }
  return tasks, rows.Err()
}

func (s *Store) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status TaskStatus, progress int, resultPath string, errMsg string, finishedAt *time.Time) error {
  _, err := s.DB.Exec(ctx, `
    update tasks
    set status = $1,
        progress = $2,
        result_path = $3,
        error_message = $4,
        updated_at = $5,
        finished_at = $6
    where id = $7
  `, status, progress, resultPath, errMsg, time.Now().UTC(), finishedAt, id)
  return err
}

func (s *Store) UpdateTaskProgress(ctx context.Context, id uuid.UUID, progress int) error {
  _, err := s.DB.Exec(ctx, `
    update tasks
    set progress = $1,
        updated_at = $2
    where id = $3
  `, progress, time.Now().UTC(), id)
  return err
}

func (s *Store) CancelTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
  tag, err := s.DB.Exec(ctx, `
    update tasks
    set status = $1,
        updated_at = $2,
        finished_at = $3
    where id = $4 and user_id = $5 and status in ('pending','running')
  `, TaskCancelled, time.Now().UTC(), time.Now().UTC(), id, userID)
  if err != nil {
    return false, err
  }
  return tag.RowsAffected() > 0, nil
}
