# 鉴权系统设计 (JWT Serverless 版)

## 核心目标
- **Stateless**: 移除 Redis，实现无状态鉴权。
- **JWT**: 使用 JSON Web Token 进行会话管理。
- **PostgreSQL**: 替代 Redis 存储临时验证码。

## 1. 验证码登录/注册流程

### Step 1: 发送验证码 (Request OTP)
- **POST** `/api/auth/send-code`
- **Params**: `email`
- **Logic**:
  1. 生成 6 位随机数字。
  2. 存入 PostgreSQL 表 `verification_codes` (email, code, expires_at, created_at)。
     - *索引优化*: 在 `email` 字段建立索引。
  3. 调用阿里云 DirectMail 发送邮件。
- **Response**: `200 OK`

### Step 2: 验证并登录 (Login)
- **POST** `/api/auth/verify`
- **Params**: `email`, `code`
- **Logic**:
  1. 查询 DB: `SELECT * FROM verification_codes WHERE email = ? AND code = ? AND expires_at > NOW()`.
  2. 如不存在或过期 -> 返回错误。
  3. 如有效 ->
     - 删除该验证码记录 (防重放)。
     - 查询或创建 User。
     - 生成 JWT Token。
       - **Payload**: `uid` (UserID), `email`, `exp` (24小时).
       - **Sign**: 使用 `JWT_SECRET` 环境变量签名 (HMAC-SHA256).
- **Response**: `{ token: "eyJhbGciOiJIUz...", user: {...} }`
  - 前端需将 Token 存入 localStorage 或 Cookie (HTTPOnly)。建议使用 Bearer Header。

## 2. API 鉴权 (Middleware)

- **AuthMiddleware**:
  1. 获取请求头 `Authorization: Bearer <token>`。
  2. 使用 `golang-jwt/jwt/v5` 解析并验证签名。
  3. 检查 `exp` 是否过期。
  4. 解析 `uid` 并注入 Gin Context (`c.Set("userID", uid)`).
  5. 如果失败，返回 `401 Unauthorized`.

## 3. 数据库设计 (Schema)

### verification_codes
| Column | Type | Note |
| :--- | :--- | :--- |
| id | SERIAL | PK |
| email | VARCHAR | Index |
| code | VARCHAR(10) | |
| expires_at | TIMESTAMP | |
| created_at | TIMESTAMP | |

## 4. 环境变量配置
- 移除: `REDIS_ADDR`, `SESSION_SECRET` (JWT 自带签名)
- 新增: `JWT_SECRET` (用于签名 Token)
