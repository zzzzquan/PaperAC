# 第二阶段：上传与任务编排 MVP

## A. 系统设计说明
### 组件
- **API服务（Gin）**：上传、任务查询、任务列表、结果下载
- **任务Worker**：从Redis队列消费任务并异步处理
- **PostgreSQL**：任务状态与元数据持久化
- **Redis**：任务队列（List + BRPOP）
- **本地存储**：上传文件与结果文件

### 时序
1. 用户上传PDF → 任务落库 → 入队 → 返回task_id
2. Worker消费任务 → pending → running → success/failed/cancelled
3. 用户查询任务状态/进度 → success后下载结果文件

### 状态机
`pending -> running -> success | failed | cancelled`

### 安全与约束
- 仅允许PDF（MIME与文件头魔数校验）
- 最大上传大小可配置（默认50MB）
- 随机文件名避免路径穿越
- 用户只能访问自己的任务

## B. 数据库迁移 SQL
见 `server/internal/store/migrations/002_tasks.sql`。

## C. Redis 队列结构与 key 规范
- 队列Key: `queue:tasks`
- 值: `task_id`
- 消费: `BRPOP` 阻塞获取

## D. Gin 路由与 handler
见 `server/internal/http/router.go` 与 `server/internal/handlers/tasks.go`。

## E. Worker 实现
见 `server/internal/worker/worker.go`。

## F. 本地存储目录与环境变量
- `UPLOAD_DIR` 上传目录（默认 /var/app/uploads）
- `RESULT_DIR` 结果目录（默认 /var/app/results）
- `MAX_UPLOAD_MB` 最大上传（默认 50）
- `TASK_TIMEOUT` 单任务超时（默认 5m）
- `WORKER_CONCURRENCY` worker并发（默认CPU核数）
- `TASK_QUEUE_NAME` 队列key（默认 queue:tasks）

## G. curl 示例
1) 上传创建任务（x范围 0~1）

```bash
curl -i -X POST "http://localhost:8080/api/tasks" \
  -H "X-CSRF-Token: <csrf>" \
  -b "sid=<session_id>" \
  -F "file=@/path/to/demo.pdf" \
  -F "x=0.7"
```

2) 查询任务

```bash
curl -i "http://localhost:8080/api/tasks/<task_id>" \
  -b "sid=<session_id>"
```

3) 下载结果（success后）

```bash
curl -i -L "http://localhost:8080/api/tasks/<task_id>/result" \
  -b "sid=<session_id>" \
  -o result.pdf
```

4) 列表

```bash
curl -i "http://localhost:8080/api/tasks?limit=20" \
  -b "sid=<session_id>"
```

5) 取消任务

```bash
curl -i -X DELETE "http://localhost:8080/api/tasks/<task_id>" \
  -H "X-CSRF-Token: <csrf>" \
  -b "sid=<session_id>"
```
