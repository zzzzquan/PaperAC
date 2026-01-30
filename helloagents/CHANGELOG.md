# Changelog

本文件记录项目所有重要变更。
格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/),
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]
### 新增
- 初始化前后端目录结构与认证基础实现（Go Gin + React CSR）
- 邮箱验证码登录、会话管理与频控机制
- 认证相关文档与部署说明
- 任务上传与编排MVP骨架（任务表、队列、worker与接口）

### 变更
- 文档统一AIGC分级规则：高/低疑似7:3，合计占比为用户输入x
- 移除LLM API依赖描述，改为数学算法占位
- 重写README并补充未实现占位说明
