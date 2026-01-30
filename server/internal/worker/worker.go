package worker

// 任务worker：消费队列并执行异步处理。

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/parser"
	"aigc-detector/server/internal/store"

	"github.com/google/uuid"
)

type Worker struct {
	Store  *store.Store
	Redis  *store.RedisStore
	Config config.Config
}

type ResultPayload struct {
	TaskID  string  `json:"task_id"`
	Status  string  `json:"status"`
	X       float64 `json:"x"`
	Message string  `json:"message"`
}

func (w *Worker) Start(ctx context.Context) {
	concurrency := w.Config.WorkerConcurrency
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.loop(ctx)
		}()
	}

	<-ctx.Done()
	wg.Wait()
}

func (w *Worker) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		taskID, err := w.Redis.PopTask(ctx, w.Config.TaskQueueName, 5*time.Second)
		if err != nil {
			continue
		}
		if taskID == "" {
			continue
		}

		w.processTask(ctx, taskID)
	}
}

func (w *Worker) processTask(ctx context.Context, taskID string) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return
	}

	task, err := w.Store.GetTaskByID(ctx, id)
	if err != nil || task == nil {
		return
	}
	if task.Status == store.TaskCancelled {
		return
	}

	_ = w.Store.UpdateTaskStatus(ctx, id, store.TaskRunning, 10, "", "", nil)

	taskCtx, cancel := context.WithTimeout(ctx, w.Config.TaskTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- w.executeStub(taskCtx, task)
	}()

	select {
	case <-taskCtx.Done():
		_ = w.Store.UpdateTaskStatus(ctx, id, store.TaskFailed, 0, "", "任务超时", nowPtr(time.Now().UTC()))
	case err := <-done:
		if err != nil {
			_ = w.Store.UpdateTaskStatus(ctx, id, store.TaskFailed, 0, "", err.Error(), nowPtr(time.Now().UTC()))
			return
		}
	}
}

func (w *Worker) executeStub(ctx context.Context, task *store.Task) error {
	if err := w.Store.UpdateTaskProgress(ctx, task.ID, 30); err != nil {
		return err
	}

	// 提取PDF文本
	text, err := parser.ExtractText(task.UploadPath)
	if err != nil {
		log.Printf("PDF解析失败: %v", err)
	} else {
		preview := text
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		log.Printf("PDF解析成功，预览: %s", preview)
	}

	// 模拟耗时，以便前端能看到进度条
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return errors.New("任务超时")
	case <-timer.C:
	}

	if err := w.Store.UpdateTaskProgress(ctx, task.ID, 60); err != nil {
		return err
	}

	if err := os.MkdirAll(w.Config.ResultDir, 0o755); err != nil {
		return err
	}

	resultFileName := filepath.Join(w.Config.ResultDir, task.ID.String()+".pdf")
	if err := copyFile(task.UploadPath, resultFileName); err != nil {
		return err
	}

	resultJSON := filepath.Join(w.Config.ResultDir, task.ID.String()+"_result.json")
	payload := ResultPayload{
		TaskID:  task.ID.String(),
		Status:  string(store.TaskSuccess),
		X:       task.X,
		Message: "stub result generated",
	}
	if err := writeJSON(resultJSON, payload); err != nil {
		return err
	}

	finishedAt := time.Now().UTC()
	if err := w.Store.UpdateTaskStatus(ctx, task.ID, store.TaskSuccess, 100, resultFileName, "", &finishedAt); err != nil {
		return err
	}

	return nil
}

func copyFile(source string, dest string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func writeJSON(path string, payload ResultPayload) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func nowPtr(t time.Time) *time.Time {
	return &t
}
