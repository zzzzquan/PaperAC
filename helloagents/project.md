# 项目技术约定

---

## 技术栈
- **核心:** Go (Gin) / React CSR / PostgreSQL / Redis

---

## 开发约定
- **代码规范:** 前端 ESLint + Prettier；后端 gofmt + golangci-lint
- **命名约定:** 后端 snake_case（数据库字段）+ Go CamelCase（结构体）；前端 camelCase

---

## 错误与日志
- **策略:** 统一错误码与错误消息结构，响应包含 request_id
- **日志:** 访问日志与认证日志分离，便于追踪邮件发送与登录行为

---

## 测试与流程
- **测试:** 以接口测试与核心流程测试为主（发送验证码、校验登录、会话与频控）
- **提交:** 约定语义化提交信息
