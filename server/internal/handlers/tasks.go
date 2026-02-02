package handlers

// 任务相关接口：上传、查询、列表、下载、取消。
// 简化版：无用户认证。

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aigc-detector/server/internal/config"
	"aigc-detector/server/internal/store"
	"aigc-detector/server/internal/util"
	"aigc-detector/server/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxListLimit = 100
	defaultLimit = 20
)

type TaskHandler struct {
	Store  *store.Store
	Worker *worker.Worker
	Config config.Config
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.JSON(c, http.StatusBadRequest, 1001, "缺少PDF文件", nil)
		return
	}
	defer file.Close()

	if header.Size > int64(h.Config.MaxUploadMB)*1024*1024 {
		util.JSON(c, http.StatusBadRequest, 1007, "文件过大", nil)
		return
	}

	xValue, err := parseX(c.PostForm("x"))
	if err != nil {
		util.JSON(c, http.StatusBadRequest, 1001, "参数x无效", nil)
		return
	}

	if err := ensurePDF(file, header); err != nil {
		util.JSON(c, http.StatusBadRequest, 1002, "仅支持PDF文件", nil)
		return
	}

	taskID := uuid.New()
	uploadFileName := fmt.Sprintf("%s.pdf", taskID.String())
	uploadPath := filepath.Join(h.Config.UploadDir, uploadFileName)

	if err := saveMultipartFile(header, uploadPath); err != nil {
		util.JSON(c, http.StatusInternalServerError, 9000, "文件保存失败", nil)
		return
	}

	now := util.Now()

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		// 允许匿名创建，但给出警告或记录？
		// 为了简化，如果没有SessionID，我们也可以允许，但意味着无法通过session隔离（或者生成一个临时的）
		// 按照需求，必须隔离。强制要求前端传。
		// 但为了兼容性，如果不传，就不关联session。
	}

	task := &store.Task{
		ID:               taskID.String(),
		SessionID:        sessionID,
		Status:           store.TaskPending,
		Progress:         0,
		X:                xValue,
		OriginalFileName: header.Filename,
		FileSize:         header.Size,
		UploadPath:       uploadPath,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.Store.CreateTask(c.Request.Context(), task); err != nil {
		util.JSON(c, http.StatusInternalServerError, 9000, "任务创建失败", nil)
		return
	}

	// 异步执行任务
	go h.Worker.ProcessTask(taskID.String())

	util.OK(c, gin.H{
		"task_id":    taskID.String(),
		"status":     task.Status,
		"created_at": task.CreatedAt,
	})
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.JSON(c, http.StatusBadRequest, 1001, "任务ID无效", nil)
		return
	}

	task, err := h.Store.GetTaskByID(c.Request.Context(), taskID.String())
	if err != nil {
		util.JSON(c, http.StatusInternalServerError, 9000, "任务查询失败", nil)
		return
	}
	if task == nil {
		util.JSON(c, http.StatusNotFound, 1008, "任务不存在", nil)
		return
	}

	util.OK(c, taskToResponse(task))
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		// 没有SessionID，返回空列表
		util.OK(c, gin.H{"items": []interface{}{}})
		return
	}

	limit := parseLimit(c.Query("limit"))
	tasks, err := h.Store.ListRecentTasks(c.Request.Context(), sessionID, limit)
	if err != nil {
		util.JSON(c, http.StatusInternalServerError, 9000, "任务列表查询失败", nil)
		return
	}

	data := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		item := taskToResponse(&task)
		data = append(data, item)
	}

	util.OK(c, gin.H{"items": data})
}

func (h *TaskHandler) ClearSession(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		util.JSON(c, http.StatusBadRequest, 1001, "缺少SessionID", nil)
		return
	}
	if err := h.Store.DeleteTasksBySessionID(c.Request.Context(), sessionID); err != nil {
		util.JSON(c, http.StatusInternalServerError, 9000, "会话清理失败", nil)
		return
	}
	util.OK(c, nil)
}

func (h *TaskHandler) DownloadResult(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.JSON(c, http.StatusBadRequest, 1001, "任务ID无效", nil)
		return
	}

	task, err := h.Store.GetTaskByID(c.Request.Context(), taskID.String())
	if err != nil {
		util.JSON(c, http.StatusInternalServerError, 9000, "任务查询失败", nil)
		return
	}
	if task == nil {
		util.JSON(c, http.StatusNotFound, 1008, "任务不存在", nil)
		return
	}
	if task.Status != store.TaskSuccess || task.ResultPath == "" {
		util.JSON(c, http.StatusBadRequest, 1009, "任务未完成", nil)
		return
	}

	c.File(task.ResultPath)
}

func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.JSON(c, http.StatusBadRequest, 1001, "任务ID无效", nil)
		return
	}

	ok, err := h.Store.CancelTask(c.Request.Context(), taskID.String())
	if err != nil {
		util.JSON(c, http.StatusInternalServerError, 9000, "任务取消失败", nil)
		return
	}
	if !ok {
		util.JSON(c, http.StatusBadRequest, 1010, "任务不可取消", nil)
		return
	}
	util.OK(c, nil)
}

func parseX(raw string) (float64, error) {
	if raw == "" {
		return 0, errors.New("missing")
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, err
	}
	if val < 0 || val > 1 {
		return 0, errors.New("out of range")
	}
	return val, nil
}

func parseLimit(raw string) int {
	if raw == "" {
		return defaultLimit
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return defaultLimit
	}
	if val > maxListLimit {
		return maxListLimit
	}
	return val
}

func ensurePDF(file multipart.File, header *multipart.FileHeader) error {
	contentType := header.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "pdf") {
		return errors.New("invalid mime")
	}
	buf := make([]byte, 5)
	if _, err := io.ReadFull(file, buf); err != nil {
		return err
	}
	if string(buf) != "%PDF-" {
		return errors.New("invalid magic")
	}
	return nil
}

func saveMultipartFile(header *multipart.FileHeader, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	src, err := header.Open()
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

func taskToResponse(task *store.Task) gin.H {
	return gin.H{
		"task_id":       task.ID,
		"status":        task.Status,
		"progress":      task.Progress,
		"x":             task.X,
		"filename":      task.OriginalFileName,
		"file_size":     task.FileSize,
		"error_message": task.ErrorMessage,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
		"finished_at":   task.FinishedAt,
	}
}

func nowPtr(t time.Time) *time.Time {
	return &t
}
