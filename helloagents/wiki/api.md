# API 手册

## 概述
对外提供认证、上传、检测、报告与账号管理接口。

## 认证方式
邮箱验证码登录，使用 Session Cookie（Redis 会话）；写操作需携带 X-CSRF-Token。

---

## 接口列表

### Auth

#### POST /api/auth/send-code
**描述:** 发送邮箱验证码。

#### POST /api/auth/verify
**描述:** 校验验证码并创建会话。

#### POST /api/auth/logout
**描述:** 退出登录并清除会话。

#### GET /api/auth/me
**描述:** 获取当前登录用户。

**错误码:**
| 错误码 | 说明 |
|--------|------|
| 1001 | 参数错误 |
| 1003 | 验证码错误 |
| 1004 | 频控限制 |
| 1005 | 验证码失效 |
| 1006 | 错误次数过多 |
| 2001 | 未登录 |
| 2002 | CSRF校验失败 |
| 9001 | 邮件发送失败 |
| 9000 | 系统错误 |

### Detection

#### POST /api/detect
**描述:** 上传PDF并创建检测任务（AIGC占比x由用户输入）。

#### GET /api/detect/{taskId}
**描述:** 获取检测任务状态与总览结果（包含高/低疑似占比，占比合计为x）。

### Report

#### GET /api/report/{taskId}
**描述:** 获取网页报告数据。

#### GET /api/report/{taskId}/download
**描述:** 下载PDF报告。

### Billing

#### GET /api/quota
**描述:** 查询剩余配额与使用情况。

### Tasks

#### POST /api/tasks
**描述:** 上传PDF并创建任务（x范围 0~1）。

#### GET /api/tasks
**描述:** 获取当前用户最近任务列表。

#### GET /api/tasks/{id}
**描述:** 获取任务详情与进度。

#### GET /api/tasks/{id}/result
**描述:** 下载结果文件（仅success可用）。

#### DELETE /api/tasks/{id}
**描述:** 取消任务（pending/running）。
