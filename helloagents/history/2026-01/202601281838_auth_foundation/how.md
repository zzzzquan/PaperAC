# 技术设计: 认证基础能力（邮箱验证码登录）

## 技术方案
### 核心技术
- 前端: React CSR + Cookie 会话
- 后端: Go + Gin
- 数据: PostgreSQL
- 缓存/会话: Redis
- 邮件: 阿里云 DirectMail HTTP API（模板参数注入）

### 实现要点
- 使用 Session Cookie + Redis 会话存储
- 验证码与频控写入 PostgreSQL，热点频控计数使用 Redis
- DirectMail 通过模板参数发送验证码，密钥从环境变量读取

## 架构决策 ADR
### ADR-001: 认证会话方案
**上下文:** 仅Web端，登录态使用Cookie，需支持快速吊销与风控。
**决策:** 采用 Session Cookie + Redis 会话。
**理由:** 易控失效、支持服务端强制下线、与Redis统一管理。
**替代方案:** JWT + 刷新机制 → 拒绝原因: 吊销与风控复杂度高。
**影响:** 服务端需维护会话存储与过期清理。

## 项目结构与阶段规划
### 目录结构（单仓）
```
/web      前端工程（React CSR）
/server   后端工程（Go Gin）
/docs     接口与部署说明
```

### 阶段规划（MVP → 增强）
- **MVP:** 邮箱验证码登录、基础鉴权、任务创建与查询、报告下载
- **增强:** 风控策略细化、配额/计费、审计日志、监控告警

## 数据模型
```sql
-- users
create table users (
  id uuid primary key,
  email varchar(255) not null unique,
  created_at timestamp not null default now()
);

-- email_verifications
create table email_verifications (
  id uuid primary key,
  email varchar(255) not null,
  code_hash varchar(255) not null,
  expires_at timestamp not null,
  attempt_count int not null default 0,
  consumed_at timestamp null,
  request_ip inet not null,
  created_at timestamp not null default now()
);
create index idx_email_verifications_email on email_verifications(email);
create index idx_email_verifications_expires_at on email_verifications(expires_at);

-- user_identities (可选，预留第三方登录)
create table user_identities (
  id uuid primary key,
  user_id uuid not null references users(id),
  provider varchar(50) not null,
  identifier varchar(255) not null,
  created_at timestamp not null default now(),
  unique(provider, identifier)
);
```

## Redis 设计
- **频控（邮箱）**: `rl:email:{email}` → 发送冷却，TTL=60s
- **频控（IP）**: `rl:ip:{ip}:{yyyyMMddHH}` → 计数器，TTL=2h
- **验证码错误次数**: `vc:err:{email}:{code_id}` → 计数器，TTL=5m
- **会话**: `sess:{session_id}` → {user_id, created_at, last_seen}, TTL=30d

## API设计
### POST /api/auth/send-code
- **请求:**
```json
{
  "email": "user@example.com"
}
```
- **响应:**
```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req_123"
}
```
- **错误码:** 1001 参数错误，1004 频控限制，9001 邮件发送失败

### POST /api/auth/verify
- **请求:**
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```
- **响应:**
```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req_123",
  "data": {
    "user_id": "uuid",
    "email": "user@example.com"
  }
}
```
- **错误码:** 1001 参数错误，1003 验证码错误，1005 验证码已失效，9002 登录失败

### POST /api/auth/logout
- **请求:** 无
- **响应:**
```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req_123"
}
```

### GET /api/auth/me
- **请求:** 无
- **响应:**
```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req_123",
  "data": {
    "user_id": "uuid",
    "email": "user@example.com"
  }
}
```
- **错误码:** 2001 未登录/会话失效

## 错误码规范
- 0: 成功
- 1xxx: 参数/校验类错误
- 2xxx: 认证/登录态错误
- 9xxx: 系统内部错误

## 核心流程时序
### send-code
用户输入邮箱 → 校验参数 → 频控检查 → 生成验证码 → 记录数据库 → 调用DirectMail → 返回 request_id

### verify
用户提交验证码 → 校验有效期/错误次数 → 自动创建用户（如不存在） → 创建会话 → 写入Cookie → 使验证码失效

### logout
校验会话 → 删除Redis会话 → 清除Cookie → 返回成功

### me
读取会话Cookie → 查询Redis会话 → 返回用户摘要

## 中间件设计
- **会话校验:** 读取Cookie中的session_id，校验Redis会话，失败返回2001
- **request_id:** 生成并透传到日志与响应体，便于追踪
- **日志:** 结构化日志，记录请求路径、状态码、耗时、request_id
- **CORS:** 仅允许Web端域名，支持带Cookie请求
- **CSRF:** 仅Web端+SameSite=Lax，敏感写操作使用双重Cookie或自定义Header校验

## 安全与性能
- **安全:** 密钥仅通过环境变量注入；验证码哈希存储；HttpOnly/SameSite Cookie；频控与错误次数限制
- **性能:** 频控计数使用Redis；会话Redis存储；数据库索引覆盖高频查询

## 测试与部署
- **测试:** 发送验证码、校验登录、频控、错误次数、会话过期等核心流程集成测试
- **部署:** 配置环境变量（DirectMail AccessKey/Secret、Redis、DB）并校验发信域名SPF/DKIM
