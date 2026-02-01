package worker

// 任务worker：无状态异步处理（Serverless Goroutine模式）。

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"aigc-detector/server/internal/algo"
	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/parser"
	"aigc-detector/server/internal/report"
	"aigc-detector/server/internal/splitter"
	"aigc-detector/server/internal/store"

	"github.com/google/uuid"
)

type Worker struct {
	Store  *store.Store
	Config config.Config
}

type ResultPayload struct {
	TaskID  string                `json:"task_id"`
	Status  string                `json:"status"`
	X       float64               `json:"x"`
	Message string                `json:"message"`
	Results []algo.SentenceResult `json:"results,omitempty"`
}

// ProcessTask 异步处理任务 (Goroutine compatible)
func (w *Worker) ProcessTask(taskID string) {
	// 创建一个独立的上下文，避免请求取消影响任务
	ctx := context.Background()

	// 设置最大执行时间
	ctx, cancel := context.WithTimeout(ctx, w.Config.TaskTimeout)
	defer cancel()

	w.process(ctx, taskID)
}

func (w *Worker) process(ctx context.Context, taskID string) {
	// Validate UUID format but keep string for store
	if _, err := uuid.Parse(taskID); err != nil {
		log.Printf("[Worker] Invalid TaskID: %s", taskID)
		return
	}

	task, err := w.Store.GetTaskByID(ctx, taskID)
	if err != nil || task == nil {
		log.Printf("[Worker] Task not found: %s", taskID)
		return
	}
	if task.Status == store.TaskCancelled {
		return
	}

	_ = w.Store.UpdateTaskStatus(ctx, taskID, store.TaskRunning, 10, "", "", nil)

	// 执行核心逻辑
	err = w.execute(ctx, task)

	if err != nil {
		log.Printf("[Worker] Task failed: %s, error: %v", taskID, err)
		_ = w.Store.UpdateTaskStatus(ctx, taskID, store.TaskFailed, 0, "", err.Error(), nowPtr(time.Now().UTC()))
	} else {
		log.Printf("[Worker] Task success: %s", taskID)
	}
}

func (w *Worker) execute(ctx context.Context, task *store.Task) error {
	if err := w.Store.UpdateTaskProgress(ctx, task.ID, 30); err != nil {
		return err
	}

	var sentences []string
	var results []algo.SentenceResult

	// 提取PDF文本
	text, err := parser.ExtractText(task.UploadPath)
	if err != nil {
		log.Printf("PDF解析失败: %v", err)
	} else {
		log.Printf("PDF解析成功，长度: %d", len(text))
		// 分句
		sentences = splitter.Split(text)
		log.Printf("分句完成，共 %d 句", len(sentences))

		// 算法处理
		processor := algo.NewProcessor()
		results = processor.Process(sentences, task.X)
	}

	// 模拟耗时
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

	// 生成 HTML 报告
	htmlData, err := report.GenerateHTML(task.OriginalFileName, results, task.X)
	if err != nil {
		return err
	}

	resultFileName := filepath.Join(w.Config.ResultDir, task.ID+".html")
	if err := os.WriteFile(resultFileName, htmlData, 0o644); err != nil {
		return err
	}

	// 保存原始 JSON 数据
	resultJSON := filepath.Join(w.Config.ResultDir, task.ID+"_result.json")
	payload := ResultPayload{
		TaskID:  task.ID,
		Status:  string(store.TaskSuccess),
		X:       task.X,
		Message: "Analysis completed successfully",
		Results: results,
	}
	_ = writeJSON(resultJSON, payload)

	finishedAt := time.Now().UTC()
	if err := w.Store.UpdateTaskStatus(ctx, task.ID, store.TaskSuccess, 100, resultFileName, "", &finishedAt); err != nil {
		return err
	}

	return nil
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
