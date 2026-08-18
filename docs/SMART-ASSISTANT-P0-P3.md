# 业务助手 P0-P3 实施说明

日期：2026-08-17

## 已落地

- P0：组合问题可规划多个只读 Tool；OpenAI Compatible 与 Anthropic Provider 会发送已脱敏业务 Payload；日程严格按目标日期过滤；公告列表严格只返回未读项；模型失败或漏答时返回确定性答案。
- P1：新增 Assistant Orchestrator、Rule Planner、LLM Planner adapter、Tool Registry、Asia/Shanghai 时间解析、最多三个只读 Tool 并发执行、CopilotRun 观测记录和 JSON 评测集。
- P2：新增私有 Knowledge Source 切片、替换、检索和权限隔离；PostgreSQL 使用全文检索表达式索引，SQLite 测试使用 LIKE adapter。
- P3：新增 LangGraph Planner adapter seam。默认链路不依赖 Python/Node，也不会因外部图运行时不可用而影响业务助手。

## 响应兼容

原有 `tool` 字段保留；组合问题新增：

- `tools`：按执行顺序返回 Tool 名称。
- `planner`：当前 Planner adapter 名称。
- `partial`：部分 Tool 因权限或执行错误未完成时为 true。

单 Tool 的 `data` 结构保持不变；多 Tool 时 `data` 为以 Tool 名称为 key 的对象。

## 知识库安全默认

当前文档模块还没有 Tenant、Department、Owner User 和 Role 行级归属。因此历史 `document_files` 不会自动进入知识索引。知识索引只接受调用方显式传入并绑定 Actor 的内容。后续在文档模型完成 ownership/Data Scope 迁移后，再接上传、更新、删除事件同步。

## LangGraph 启用门槛

只有同时满足以下条件才考虑把 LangGraph adapter 接入灰度流量：

1. committed planner evaluation set 的 Tool recall 和 exact match 高于 Rule Planner；
2. P95 规划延迟满足目标；
3. 外部运行时不可用时可立即回退 Rule Planner；
4. Tool Registry、Casbin、Data Scope 和确定性 synthesis 不被绕过；
5. 部署、日志、追踪和版本回滚均已自动化。
