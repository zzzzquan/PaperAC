# 情智文检——高情商AIGC（PaperAC）

> 本文件包含项目级别的核心信息。详细的模块文档见 `modules/` 目录。

---

## 1. 项目概述

### 目标与背景
为公开场景提供AIGC检测服务，面向课程论文场景，支持用户上传PDF并生成检测报告；AIGC疑似分级为高度疑似/轻度疑似，比例7:3，合计占比为用户输入x。

### 范围
- **范围内:** PDF上传、正文抽取、句子级AIGC识别、报告生成与下载、账号与配额管理、免责声明展示
- **范围外:** 论文写作助手、全文改写、查重服务

### 干系人
- **负责人:** 待指定

---

## 2. 模块索引

| 模块名称 | 职责 | 状态 | 文档 |
|---------|------|------|------|
| Web前端 | 上传、结果展示、报告下载、账号入口 | 🚧开发中 | [web.md](modules/web.md) |
| API服务 | 统一入口、鉴权、任务编排 | 🚧开发中 | [api.md](modules/api.md) |
| PDF处理 | 解析、清洗、句子拆分 | 📝规划中 | [pdf.md](modules/pdf.md) |
| AIGC判别 | 数学算法判别、阈值控制、结果归一 | 📝规划中 | [aigc.md](modules/aigc.md) |
| 报告生成 | 总览与标红、PDF导出 | 📝规划中 | [report.md](modules/report.md) |
| 账号与计费 | 注册登录、配额、计费策略 | 📝规划中 | [billing.md](modules/billing.md) |
| 合规与风控 | 免责声明、反滥用、数据删除 | 📝规划中 | [compliance.md](modules/compliance.md) |

---

## 3. 快速链接
- [技术约定](../project.md)
- [架构设计](arch.md)
- [API 手册](api.md)
- [数据模型](data.md)
- [变更历史](../history/index.md)
