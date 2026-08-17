# 租户与部门级数据隔离实施规格

- **决策状态**：已接受
- **决策日期**：2026-08-17
- **关联 ADR**：[0001 Tenant-rooted department data isolation](adr/0001-tenant-department-data-isolation.md)
- **当前状态**：模型与迁移方案已明确，生产数据尚未完成字段迁移和强制过滤

## 1. 目标

本方案建立统一的业务数据边界，避免把菜单权限、API 权限、角色关系或对象 key 当作行级授权依据。

必须满足：

1. 不同租户的用户不能读取、修改、导出、统计或通过 AI 间接获取对方数据。
2. 同一租户内按本人、当前部门、部门子树或全租户范围授权。
3. 列表、详情、写操作、文件代理、后台任务和智能工具使用同一范围解析结果。
4. 超级管理员跨租户行为显式、可审计，租户管理员不能跨租户。
5. 历史数据可以无损迁入默认租户和根部门，并支持分阶段上线与回滚。

本方案不把审批流、岗位体系、成本中心或法人核算关系合并进 Department；这些概念需要时独立建模。

## 2. 当前差距

| 模块 | 当前边界 | 缺口 |
| --- | --- | --- |
| 用户与角色 | 用户只有主角色；角色可配置 `DataAuthorityId` | 没有租户、部门和标准数据范围 |
| 资产、位置、流转、风险 | 主要 Service 直接查询全表 | 已获 API 权限的用户可能看到全部组织数据 |
| 智能建档 | 普通用户仅看自己，超级管理员看全部 | 没有部门和租户范围 |
| 发票 | 本人、关联角色或超级管理员全量 | 角色范围不是组织边界，无法阻止跨租户角色关系 |
| 文档、公告、站点 | 主要数据为全局列表 | 没有组织归属和跨范围校验 |
| 日程 | 按用户隔离 | 个人边界有效，但缺少租户归属用于关联和审计 |
| Copilot、日报、草稿 | 用户范围与发票角色范围混合 | Tool 查询和聚合没有统一 Data Scope |
| 对象存储 | 私有 bucket，业务 API 代理 | 数据库记录授权仍需租户和部门过滤 |

## 3. 领域模型

### 3.1 Tenant

Tenant 是最高业务数据边界。当前内部部署也必须存在一个默认 Tenant，不能用 `tenant_id = 0` 表示“单租户模式”。

建议表 `sys_tenants`：

| 字段 | 约束 | 说明 |
| --- | --- | --- |
| `id` | 主键 | 租户标识 |
| `code` | 全局唯一、不可变 | 稳定业务编码 |
| `name` | 非空 | 租户名称 |
| `status` | `active/disabled` | 禁用后拒绝新会话和业务写入 |
| `is_default` | 最多一条为 true | 历史数据回填目标 |

### 3.2 Department

Department 只存在于一个 Tenant 内，使用邻接表保存父子关系，并维护可查询的祖先路径或闭包表支持子树范围。

建议表 `sys_departments`：

| 字段 | 约束 | 说明 |
| --- | --- | --- |
| `id` | 主键 | 部门标识 |
| `tenant_id` | 非空、索引 | 所属租户 |
| `parent_id` | 同租户外键，可空 | 根部门没有父部门 |
| `code` | 租户内唯一 | 部门编码 |
| `name` | 非空 | 部门名称 |
| `path` | 租户内唯一 | 稳定祖先路径，用于子树查询 |
| `status` | `active/disabled` | 停用不删除历史归属 |

任何移动部门操作必须验证新父节点属于同一 Tenant，并在事务中更新整棵子树路径。

### 3.3 Department Membership

建议表 `sys_user_departments`：

| 字段 | 约束 | 说明 |
| --- | --- | --- |
| `tenant_id` | 非空 | 用户和部门必须属于该租户 |
| `user_id` | 非空 | 用户 |
| `department_id` | 非空 | 部门 |
| `is_primary` | 每用户最多一条为 true | 创建业务数据时的默认部门 |

`sys_users` 增加非空 `tenant_id` 和 `primary_department_id`。用户可以有多个部门成员关系，但一次请求必须有明确的当前部门；切换当前部门只能在自己的成员关系中进行。

### 3.4 Role Data Scope

`sys_authorities` 增加 `data_scope`：

| 值 | 可见范围 |
| --- | --- |
| `self` | 本人创建或本人负责的数据 |
| `department` | 当前部门数据 |
| `department_tree` | 当前部门及全部下级部门数据 |
| `tenant` | 当前租户全部数据 |
| `all` | 全平台，仅平台超级管理员 |

用户拥有多个角色时，在同一 Tenant 内取范围并集。任何角色关系、菜单授权或 `DataAuthorityId` 都不能扩大到其他 Tenant。角色 `888` 在当前系统中映射为 Platform Administrator；新增租户管理员角色应使用 `tenant`，不能复用 `all`。

## 4. 统一访问上下文

认证后解析不可变的 `DataScope`：

```text
DataScope
  ActorUserID
  TenantID
  CurrentDepartmentID
  VisibleDepartmentIDs
  Level: self | department | department_tree | tenant | all
  PlatformAdministrator
```

解析规则：

1. JWT 只提供身份线索；中间件继续实时校验用户启用状态和 `security_version`。
2. Tenant、部门成员关系和角色范围从数据库或可失效缓存解析，不信任请求体中的归属字段。
3. `tenant_id = 0`、用户无主部门、租户禁用或部门不属于租户时拒绝业务请求。
4. 平台超级管理员跨租户操作必须显式选择目标 Tenant，并写安全审计；不能默认为全租户聚合。

建议新增独立深模块 `server/datascope`，对外只暴露 Actor 解析、GORM Scope 和归属校验。业务 Service 的租户型方法必须接收 `datascope.Scope`，零值 Scope 直接失败。

## 5. 查询与写入规则

### 5.1 查询

- `all`：仅平台管理员显式指定目标 Tenant 时查询；跨租户报表使用单独受审计入口。
- `tenant`：强制 `tenant_id = scope.TenantID`。
- `department_tree`：同时限制 `tenant_id` 和 `department_id IN scope.VisibleDepartmentIDs`。
- `department`：同时限制 `tenant_id` 和当前 `department_id`。
- `self`：同时限制 `tenant_id`，再按领域使用 `created_by`、`owner_user_id` 或 `custodian_user_id`。

禁止只在 Controller 给列表追加条件。详情、更新和删除必须从已过滤的查询开始，并将越权与不存在统一返回 `404` 或稳定业务错误，避免泄露记录存在性。

### 5.2 创建与更新

- `tenant_id`、`department_id`、`created_by` 由服务端写入，客户端同名字段忽略或拒绝。
- 跨部门转移使用独立领域动作，验证操作者同时拥有源和目标范围。
- 关联记录必须校验 Tenant 一致，例如资产和流转单、发票和分类、公告和接收人。
- 批量操作先在范围内加载全部 ID；命中数量不一致时整体失败，不能部分越权成功。

### 5.3 聚合、导出和 AI

- Dashboard、日报、导出和统计从已过滤查询构建，不能先全表聚合再在内存过滤。
- AI Gateway 的 Tool 输入不接受任意 Tenant/Department；Tool 使用当前 Actor 的 DataScope。
- 引用、文件和对象下载再次执行数据授权，不能只依赖 Copilot 已返回的 ID。
- 后台任务按 Tenant 分片执行，并把 `tenant_id` 写入任务、运行记录和审计记录。

## 6. 表归属矩阵

| 类型 | 表或模块 | 归属规则 |
| --- | --- | --- |
| 平台全局 | API、菜单模板、平台 Provider 定义 | 不加部门字段；密钥仍仅平台管理员可见 |
| 租户配置 | 资产分类、位置、发票分类、分类规则、公告模板、AI 配额 | `tenant_id` 非空，`department_id` 可选 |
| 部门业务 | 资产、流转单、风险事件、发票、文档、公告 | `tenant_id`、`department_id` 非空 |
| 用户私有 | 日程、通知、Copilot 会话、个人订阅 | `tenant_id`、`user_id` 非空；部门用于审计或业务关联 |
| 派生记录 | 识别任务、发票明细、查验、清理任务、风险日志、智能草稿 | 复制根记录的 `tenant_id`，必要时复制 `department_id` |

新对象 key 使用 `tenant/{tenant_id}/{domain}/{yyyy}/{mm}/{uuid}.{ext}`。历史对象不要求立即搬迁；数据库授权仍是唯一访问入口，新写入先切换前缀，后台任务再按内容 hash 校验后迁移旧 key。

## 7. 数据库约束与索引

- 所有租户型查询索引以 `tenant_id` 为首列，部门型高频查询使用 `(tenant_id, department_id, status)`。
- 资产编号、部门编码、分类编码等唯一约束改为租户内唯一，例如 `(tenant_id, asset_code)`。
- 子表同时保存 `tenant_id`，并通过复合外键或事务校验阻止跨租户关联。
- 删除 Tenant 和 Department 使用停用或受控归档，不级联删除业务历史。
- 迁移完成后为业务表设置 `tenant_id NOT NULL`；需要部门归属的表设置 `department_id NOT NULL`。

PostgreSQL RLS 作为第二阶段防御：只有在应用过滤、迁移、连接池会话变量清理和后台任务测试全部通过后启用。RLS 不能替代 Service 契约，也不能让 SQLite 单元测试失去同等业务校验。

## 8. 迁移与发布顺序

### 阶段 A：结构准备

1. 备份 PostgreSQL，记录各业务表行数和孤儿关联。
2. 创建默认 Tenant、根 Department 和用户成员关系。
3. 以可空方式增加归属字段和新索引，不立即改变线上查询。
4. 回填历史数据：Tenant 为默认租户；部门按用户主部门，无法推断时进入根部门并生成待治理清单。

### 阶段 B：双写和观察

1. 创建流程开始写入归属字段。
2. 引入 `DataScope`，在 observe 模式记录“旧查询结果数与新范围结果数”差异，不向用户暴露跨范围数据明细。
3. 修复跨租户关联、空归属和唯一键冲突。

### 阶段 C：分模块强制

1. 资产、流转、风险和智能建档。
2. 发票、查验、质量看板和对象文件。
3. 文档、公告、站点和日程关联。
4. Copilot、日报、智能草稿、AI 配额和审计。

每个模块同时切换列表、详情、写操作、统计、导出、文件和后台任务；不能只上线列表过滤。

### 阶段 D：收紧约束

1. 将归属字段改为非空。
2. 替换全局唯一索引为租户内唯一索引。
3. 删除 observe 旁路，生产只保留 enforce。
4. 评估并启用 PostgreSQL RLS、跨租户告警和季度权限复审。

## 9. 验收矩阵

固定准备以下主体：

- A：租户 T1、部门 D1、`self`。
- B：租户 T1、部门 D1、`department`。
- C：租户 T1、父部门 D0、`department_tree`。
- D：租户 T1、`tenant` 管理员。
- E：租户 T2 管理员。
- P：Platform Administrator。

每个业务模块必须验证：

| 场景 | 预期 |
| --- | --- |
| A 访问自己记录 | 允许 |
| A 访问同部门其他人的记录 | 拒绝 |
| B 访问 D1 记录 | 允许 |
| B 访问同租户兄弟部门 | 拒绝 |
| C 访问子部门记录 | 允许 |
| C 访问父部门或兄弟部门 | 拒绝 |
| D 访问 T1 任意部门 | 允许 |
| D 或任意 T1 用户访问 T2 | `404/403`，审计记录跨租户尝试 |
| P 未选择目标 Tenant 直接查询 | 拒绝 |
| P 显式选择目标 Tenant | 允许并写高风险审计 |

同一矩阵必须覆盖列表、详情、创建、更新、删除、批量、Dashboard、导出、文件代理、AI Tool 和后台任务。仅通过 Casbin API 权限测试不能算数据隔离验收完成。

## 10. 完成定义

部门/租户级数据隔离只有同时满足以下条件才算落地：

1. 所有目标表完成归属字段回填，无 `tenant_id = 0` 和跨租户外键。
2. 所有业务 Service 强制接收并应用 DataScope；代码扫描不存在绕过的全表入口。
3. 资产、发票、文档、公告、日程、风险和智能模块通过完整跨范围测试矩阵。
4. 对象代理、导出、日报和 AI Tool 不泄露跨范围数据。
5. Platform Administrator 跨租户操作显式选择目标并进入安全审计。
6. 生产 observe 差异归零，切换 enforce 后负向合成测试通过。
7. 数据库约束、索引、备份、回滚和历史数据治理记录齐全。

## 11. 回滚原则

- 结构迁移采用向前兼容方式，回滚应用版本时不删除归属字段和已回填数据。
- 双写阶段可回滚到旧读取，但进入 enforce 后不得通过关闭过滤恢复跨租户全表访问。
- 模块切换失败时回滚该模块发布，并保留 Tenant 过滤；不能以恢复公开数据作为可用性措施。
- 对象 key 迁移保留旧 key 到校验和切换完成，失败通过清理任务重试，不执行无备份批量删除。
