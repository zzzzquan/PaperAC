# 数据模型

## 概述
以账号、验证码、任务、报告与配额为核心数据。

---

## 数据表/集合

### users
**描述:** 用户账号与基础信息。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uuid | 主键 | 用户ID |
| email | string | 唯一 | 邮箱 |
| created_at | datetime | 非空 | 创建时间 |

### email_verifications
**描述:** 邮箱验证码记录与校验信息。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uuid | 主键 | 记录ID |
| email | string | 非空 | 邮箱 |
| code_hash | string | 非空 | 验证码哈希 |
| expires_at | datetime | 非空 | 过期时间 |
| attempt_count | int | 非空 | 错误次数 |
| consumed_at | datetime | 可空 | 使用时间 |
| request_ip | string | 非空 | 请求IP |
| created_at | datetime | 非空 | 创建时间 |

### user_identities（可选）
**描述:** 第三方身份绑定（预留）。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uuid | 主键 | 记录ID |
| user_id | uuid | 外键 | 用户ID |
| provider | string | 非空 | 身份提供方 |
| identifier | string | 非空 | 唯一标识 |
| created_at | datetime | 非空 | 创建时间 |

### tasks
**描述:** 检测任务与编排状态。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uuid | 主键 | 任务ID |
| user_id | uuid | 外键 | 用户ID |
| status | string | 非空 | pending/running/success/failed/cancelled |
| progress | int | 非空 | 进度0-100 |
| x | numeric | 非空 | 目标占比（0~1） |
| original_filename | string | 非空 | 原文件名 |
| upload_path | string | 非空 | 上传路径 |
| result_path | string | 可空 | 结果路径 |
| error_message | string | 可空 | 错误信息 |
| created_at | datetime | 非空 | 创建时间 |
| updated_at | datetime | 非空 | 更新时间 |
| finished_at | datetime | 可空 | 完成时间 |

### detection_tasks
**描述:** 检测任务主表。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uuid | 主键 | 任务ID |
| user_id | uuid | 外键 | 用户ID |
| status | string | 非空 | pending/running/done/failed |
| threshold | float | 非空 | 目标AIGC占比x（用户输入） |
| created_at | datetime | 非空 | 创建时间 |

### reports
**描述:** 检测报告与汇总结果。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uuid | 主键 | 报告ID |
| task_id | uuid | 外键 | 任务ID |
| summary | json | 非空 | 总览统计（含高/低疑似占比） |
| detail | json | 非空 | 句子级结果 |
| created_at | datetime | 非空 | 创建时间 |

**summary字段建议:**
- aigc_ratio: 总AIGC占比（等于用户输入x）
- high_suspected_ratio: 高度疑似占比（x * 0.7）
- low_suspected_ratio: 轻度疑似占比（x * 0.3）

### usage_quota
**描述:** 配额与计费信息。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uuid | 主键 | 记录ID |
| user_id | uuid | 外键 | 用户ID |
| plan | string | 非空 | 计费方案 |
| remaining | int | 非空 | 剩余次数 |
| updated_at | datetime | 非空 | 更新时间 |
