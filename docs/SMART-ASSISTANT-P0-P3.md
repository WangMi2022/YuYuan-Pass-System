# 业务助手 P0-P3 实施说明

日期：2026-08-17

## 资产条件查询

更新：2026-09-11。资产实时查询使用现有数据库，无需新增向量库。

- 常见明确问法由规则完整解析；剩余条件交给已有 AI Gateway 的模型解析。模型关闭、超时、输出非法字段或无法完整表达条件时返回澄清消息，不执行部分条件查询。
- `asset.search` 接受结构化 `query`，保留旧 `keyword/custodian` 参数；`asset.warranty.expiring` 复用同一执行模块，兼容 `days`。
- 条件支持 `all/any` 嵌套、比较、文本包含、日期空值判断。支持金额、数量、名称/编号/品牌/型号/序列号、分类、保管人、保管部门、位置、供应商、状态、购置日、生产日和质保到期日。
- 质保过期表示到期日早于上海时区的今天；今天到期仍属于未过保。未填写日期不参与日期比较。未指定即将到期范围时使用未来 30 天，并展示实际日期。
- 部门按当前 `custodian=部门-姓名` 格式筛选，不代表租户权限。资产模块尚未完成 Tenant/Department 字段迁移，本次不引入新的归属模型或声称完成行级租户隔离；所有查询继续检查 `/asset/list`。
- 维修次数按已完成维修单去重，时间范围使用业务日期，排除已删除单据与记录；额外检查 `/assetOperation/list` 权限。
- 支持原生字段排序，以及状态、保管人、分类、位置分组。分组按记录数降序；单页默认 20，最多 100，汇总始终基于全部匹配记录。记录数、实物数量、原值和当前估值分别计算。
- `query` 最多 32 个条件节点、嵌套深度 4、3 个排序字段；字段和运算符白名单校验后绑定 SQL 参数。模型不能提供 SQL、表名、租户 ID 或权限条件。
- 资产回答由后端直接生成，不再次交给模型改写数字。规划条件写入 CopilotRun；澄清记录状态为 `clarification`，不含工具执行。
- 同一会话中对“过期”的澄清回复“质保”会保留原部门和保管人条件；“下一页”“上一页”“第 N 页”复用最近一次成功资产查询的条件并重新鉴权。使用年限、借用截止日和维修费用尚无完整数据口径，系统会提示不支持。

示例：“有哪些已经过保的资产”“行政部王磊名下价值超过5000元的电脑，按价值降序”“2026年维修至少3次的设备”“今年购入的闲置设备，按分类汇总”。复杂问法需启用已配置的模型。

验证：`go test ./plugin/smart/... ./plugin/asset/...`。`asset_structured_query_test.go` 覆盖组合条件、日期边界、未知日期、全量统计、分页、分组、维修权限、非法输入、模型失败和澄清追问。

可选集成核验：显式设置 `SMART_QUERY_PROVIDER_CONFIG` 后运行 `TestAssetQueryConfiguredProvider`，使用实际 Provider 解析固定问题，业务数据及调用审计留在内存测试库；显式设置 `SMART_QUERY_POSTGRES_CONFIG` 后运行 `TestAssetQueryPostgresReadOnly`，以数据库强制只读事务核对日期、组合、分组和维修查询。两个变量只接收运行配置文件路径，不在测试输出中打印凭据或模型原文。

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
