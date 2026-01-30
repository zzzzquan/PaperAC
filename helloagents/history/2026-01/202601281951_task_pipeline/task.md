# 任务清单: 上传与任务编排（MVP）

目录: `helloagents/plan/202601281951_task_pipeline/`

---

## 1. 数据模型与迁移
- [√] 1.1 在 `server/internal/store/migrations/002_tasks.sql` 添加 tasks 表与索引，验证 why.md#需求-上传创建任务-场景-创建任务

## 2. 配置与存储
- [√] 2.1 在 `server/internal/config/config.go` 增加上传目录、结果目录、任务超时与并发配置，验证 why.md#需求-任务异步处理-场景-处理任务

## 3. 任务存储与查询
- [√] 3.1 在 `server/internal/store/tasks.go` 实现任务创建/查询/列表/状态更新，验证 why.md#需求-上传创建任务-场景-创建任务

## 4. API与路由
- [√] 4.1 在 `server/internal/handlers/tasks.go` 实现上传/查询/列表/下载/取消接口，验证 why.md#需求-查询与下载-场景-查询任务
- [√] 4.2 在 `server/internal/http/router.go` 注册 /api/tasks 路由与权限校验

## 5. Worker处理
- [√] 5.1 在 `server/internal/worker/worker.go` 实现队列消费、并发控制、超时处理与结果写入，验证 why.md#需求-任务异步处理-场景-处理任务
- [√] 5.2 在 `server/cmd/api/main.go` 启动worker

## 6. 安全检查
- [√] 6.1 执行安全检查（按G9: 输入验证、敏感信息处理、权限控制、EHRB风险规避）

## 7. 文档更新
- [√] 7.1 更新 `helloagents/wiki/api.md`
- [√] 7.2 更新 `helloagents/wiki/data.md`
- [√] 7.3 更新 `helloagents/wiki/modules/api.md`

## 8. 测试
- [√] 8.1 在 `server/tests` 补充任务流程集成测试占位
