# 架构设计

## 总体架构
```mermaid
flowchart TD
    U[用户] --> F[Web前端]
    F --> A[API服务]
    A --> P[PDF处理]
    A --> D[AIGC判别]
    A --> R[报告生成]
    A --> B[账号与计费]
    A --> S[临时存储/缓存]
```

## 技术栈
- **后端:** Go (Gin)
- **前端:** React CSR
- **数据:** PostgreSQL + Redis + 临时对象存储
- **邮件通道:** 阿里云 DirectMail（HTTP API）
- **AIGC算法:** 数学算法（占位，后续设计）

## 核心流程
```mermaid
sequenceDiagram
    participant User as 用户
    participant Web as Web前端
    participant API as API服务
    participant PDF as PDF处理
    participant Algo as AIGC算法
    participant Report as 报告生成
    User->>Web: 上传PDF并设置阈值
    Web->>API: 创建检测任务
    API->>PDF: 解析并拆分句子
    API->>Algo: 句子级AIGC判别（数学算法）
    API->>Report: 汇总并生成报告
    API->>Web: 返回总览与报告下载链接
```

## 重大架构决策
完整的ADR存储在各变更的how.md中，本章节提供索引。

| adr_id | title | date | status | affected_modules | details |
|--------|-------|------|--------|------------------|---------|
| ADR-001 | 认证会话方案（Session Cookie + Redis） | 2026-01-28 | ✅已采纳 | API服务/Web前端 | helloagents/history/2026-01/202601281838_auth_foundation/how.md |
| ADR-002 | 任务队列方案（Redis List + BRPOP） | 2026-01-28 | ✅已采纳 | API服务/任务编排 | helloagents/history/2026-01/202601281951_task_pipeline/how.md |
