# 技术设计: 上传与任务编排（MVP）

## 技术方案
### 核心技术
- 后端: Go + Gin
- 数据: PostgreSQL
- 队列/缓存: Redis List + BRPOP
- 存储: 本地目录（/var/app/uploads, /var/app/results）

### 实现要点
- 上传接口只负责校验与落库，不执行处理逻辑
- worker独立协程池执行任务，支持并发与超时控制
- 任务状态机：pending → running → success | failed | cancelled

## 架构决策 ADR
### ADR-002: 任务队列方案
**上下文:** 需要轻量可靠的任务队列，避免引入重型框架。
**决策:** 采用 Redis List + BRPOP 作为队列。
**理由:** 实现简单、性能稳定、满足MVP需求。
**替代方案:** Redis Streams → 拒绝原因: 复杂度高、维护成本高。
**影响:** 需要自行处理重试与超时标记。

## API设计
### POST /api/tasks
- **请求:** multipart/form-data，file=PDF，x 为浮点数（0~1）
- **响应:**
```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req_123",
  "data": {
    "task_id": "uuid",
    "status": "pending",
    "created_at": "2026-01-28T12:00:00Z"
  }
}
```

### GET /api/tasks/:id
- **响应:**
```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req_123",
  "data": {
    "task_id": "uuid",
    "status": "running",
    "progress": 40,
    "x": 0.7,
    "filename": "origin.pdf",
    "error_message": "",
    "created_at": "2026-01-28T12:00:00Z",
    "updated_at": "2026-01-28T12:00:03Z",
    "finished_at": null
  }
}
```

### GET /api/tasks/:id/result
- **响应:** 仅 success 可下载结果文件

### GET /api/tasks
- **响应:** 返回最近20条任务

### DELETE /api/tasks/:id
- **响应:** 先占位，成功则状态改为cancelled

## 数据模型
```sql
create table tasks (
  id uuid primary key,
  user_id uuid not null references users(id),
  status varchar(20) not null,
  progress int not null default 0,
  x numeric(5,4) not null,
  original_filename varchar(255) not null,
  upload_path text not null,
  result_path text null,
  error_message text null,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  finished_at timestamp null
);
create index idx_tasks_user_created on tasks(user_id, created_at desc);
```

## 队列设计
- **队列Key:** `queue:tasks`
- **元素:** task_id（uuid）
- **Worker:** BRPOP阻塞拉取，失败时标记任务 failed

## 安全与性能
- **安全:** MIME/魔数校验PDF；最大上传50MB；随机文件名；权限校验 user_id
- **性能:** 并发=CPU核数或 WORKER_CONCURRENCY

## 测试与部署
- **测试:** 上传创建任务、队列消费、任务状态流转、下载结果
- **部署:** 配置上传/结果目录、Redis/DB、worker并发与超时
