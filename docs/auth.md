# 认证与会话说明

## 登录流程
1. 用户输入邮箱请求验证码
2. 系统生成6位验证码并发送邮件
3. 用户提交验证码后完成登录（注册与登录合一）

## 安全约束
- 验证码有效期5分钟
- 同邮箱60秒限频
- 同IP每小时20次
- 最多错误次数5次
- 成功后验证码一次性失效

## 会话策略
- Session Cookie + Redis 会话
- Cookie: HttpOnly + SameSite=Lax
- CSRF: 写操作要求 X-CSRF-Token

## 运行配置
- DirectMail 通过 HTTP API 发送交易邮件
- AccessKey/Secret 从环境变量读取（当前方案）
