# YuYuan Pass System 开发维护指南

## 1. 环境要求

| 工具 | 要求 |
| --- | --- |
| Go | `1.24.x`，仓库声明 toolchain `1.24.2` |
| Node.js | 支持 Vite 8 的当前 LTS 版本 |
| npm | 与 Node.js 配套 |
| PostgreSQL | 14-18 |
| Redis | 6+，可按配置关闭 |
| S3 兼容存储 | RustFS 或 MinIO |
| Docker / Compose | 推荐用于集成与交付验证 |

## 2. 本地启动

### 2.1 Server

```bash
cd server
go mod download
go run . -c config.yaml
```

`server/config.yaml` 包含真实连接信息，不得提交。可从无敏感信息的部署模板理解结构，但本地凭据应单独维护。

### 2.2 Web

```bash
cd web
npm install --legacy-peer-deps
npm run dev
```

主要脚本：

| 命令 | 用途 |
| --- | --- |
| `npm run dev` | 启动 Vite 开发服务器 |
| `npm test` | 运行 Node Test 测试集合 |
| `npm run lint` | ESLint |
| `npm run build` | 生产构建 |
| `npm run check` | Lint + Build |

## 3. 目录职责

### 新增业务插件

服务端建议结构：

```text
server/plugin/<domain>/
├─ api/
├─ initialize/
├─ model/
│  ├─ request/
│  └─ response/
├─ router/
├─ service/
└─ plugin.go
```

Web 建议结构：

```text
web/src/plugin/<domain>/
├─ api/
├─ view/
└─ components/        # 需要时
```

不要把业务规则写入 `initialize`、页面组件或 Router；Router 负责路径，API 负责协议，Service 负责规则和事务。

## 4. 新功能开发清单

1. 建立领域模型、请求和响应结构。
2. 在 Service 实现业务规则、权限数据范围和事务。
3. 在 API 层绑定参数并使用统一响应。
4. 注册 read/write 路由；写接口评估挂载 `OperationRecord()`。
5. 增加 GORM 迁移、菜单和 API 权限初始化。
6. 增加前端 API 封装、页面、空/错/加载状态。
7. 增加最窄的单元或集成测试。
8. 更新 Swagger、API、功能、数据和用户文档。

## 5. API 约定

- JSON 字段使用 lowerCamelCase；历史 GVA 字段如 `ID` 保持兼容。
- GET 用于查询，POST 用于创建/动作，PUT 用于更新/状态变更，DELETE 用于删除。
- 分页统一使用 `page`、`pageSize`，返回 `list/total/page/pageSize`。
- 业务失败使用统一响应，不把数据库错误原文直接暴露给用户。
- 文件下载明确设置 Content-Type、文件名和缓存策略。
- 当前用户 ID 从认证上下文读取，不从客户端请求体信任。

## 6. 数据与事务

- 跨多表状态变化必须使用事务。
- 对需要历史追溯的状态保存快照或事件，不只保存当前值。
- 金额口径必须在模型层明确；财务金额优先整数最小单位。
- 唯一业务约束使用数据库唯一索引兜底。
- 数据库与对象存储跨资源操作需要 outbox、清理任务或可重试补偿。

## 7. 权限与安全

- 私有 API 默认放入 JWT + Casbin 路由组。
- “仅登录可用”必须有明确原因，例如用户自己的通知和日程；Service 仍需按当前用户过滤。
- 新菜单必须同时初始化菜单和 API 权限。
- 上传校验扩展名、MIME、大小和目标路径，文件名不能直接作为对象 key。
- 禁止提交 `.env`、`config.yaml`、密钥、证书、数据库备份和远程访问脚本。

## 8. 测试策略

### 服务端

```bash
cd server
go test ./plugin/<domain>/...
go test ./...
```

优先测试：

- 状态机允许/拒绝路径。
- 事务失败是否回滚。
- 权限和数据归属。
- 金额、日期、重复和幂等约束。
- 对象存储失败与清理重试。

### Web

```bash
cd web
npm test
npm run lint
npm run build
```

页面验收至少覆盖桌面端、移动端、深色模式、空数据、请求失败和无权限。

### 部署脚本

```bash
cd deploy/docker-dev
bash tests/release-acceptance-test.sh
./release-acceptance.sh
```

## 9. Swagger

Swagger 生成文件位于 `server/docs/`。API 变更时：

1. 更新 Handler 上的 Swaggo 注释。
2. 重新生成 `docs.go`、`swagger.json`、`swagger.yaml`。
3. 检查鉴权声明、参数位置、请求结构和响应示例。
4. 确保业务插件路径被包含。

当前插件覆盖仍不完整，新增接口不得继续扩大差距。

## 10. M5-M7 开发与验收

### 后端

```bash
cd server
go test ./plugin/smart/...
go test ./plugin/asset/... ./plugin/invoice/... ./plugin/schedule/...
```

重点检查 `server/plugin/smart/service/`：Tool 只能通过注册表进入并按底层 Casbin 权限过滤；权限失败不得创建空会话；会话、日报、订阅、草稿必须带 `user_id`；日报每个口径使用独立查询并通过数据库 Upsert 保证同日幂等；业务草稿确认必须二次校验原业务权限并调用既有 Asset/Schedule Service，不能直接更新资产状态。

### 前端

```bash
cd web
npm run lint -- src/plugin/smart src/plugin/announcement/view/info.vue
npm run build
```

页面验收覆盖 `1440×1000`、`900×900`、`390×844`，检查智能助手会话、日报详情、订阅、草稿候选选择和公告编辑器提取按钮；必须同时验证加载、空数据、API 失败、无权限和移动端无横向溢出。

### 接口安全验收

1. 未携带 JWT 请求 `/smart/copilot/query`、`/smartReport/today`、`/smart/drafts` 返回 `401`。
2. 普通角色缺少 `/smart/*` API 时不能通过动态菜单或直接请求访问。
3. 用户 A 请求用户 B 的会话、日报、投递或草稿 ID 时返回无权/不存在。
4. 关闭 AI Provider 后，业务助手和日报仍返回确定性结果。
5. 同一公告重复提取、同一草稿并发确认、同一日报定时任务重复运行均保持幂等。
6. `/smart/copilot/tools` 不返回角色无权访问的资产、发票、日程或公告 Tool；智能草稿权限不能绕过 `/assetOperation/*` 和 `/workSchedule/create`。

### 数据迁移

智能插件启动时自动迁移 `ai_copilot_sessions`、`ai_copilot_messages`、`smart_daily_reports`、`smart_report_subscriptions`、`smart_report_deliveries` 和 `smart_drafts`。新增 `smart_report_deliveries` 唯一投递索引前必须检查历史重复数据；生产升级前先备份并执行重复清理方案。

### 生产业务验收

```bash
cd deploy/docker-dev
./m5-m7-production-acceptance.sh --execute
```

该脚本是可变更的生产验收，只能在已备份的受控发布窗口执行。它使用临时隔离用户和带前缀的业务数据，临时扩展普通角色验收权限，完成 M5-M7 接口、权限、指标对账、定时投递、降级和并发确认后，无条件恢复原权限并清理测试数据。详细检查项见 [部署运维手册](DEPLOYMENT.md#171-m5-m7-生产业务验收)。

## 11. Git 与发布

- 从 `main` 发布。
- 先检查 `git status -sb` 和完整 diff。
- 只暂存当前任务文件，不带入其他未提交改动。
- 使用 Conventional Commit，例如 `docs: expand product and technical documentation`。
- 推送后以该 commit 构建部署，生产记录完整 commit hash。
- 运行时变更必须通过发布验收；纯文档变更不要求重建运行容器。

## 12. 文档 Definition of Done

| 变更类型 | 必须更新 |
| --- | --- |
| 新用户功能 | README、产品说明、功能规格、用户手册 |
| 新/改 API | API 文档、Swagger、前端 API 封装 |
| 新表/字段/状态 | 数据字典、功能规格、迁移说明 |
| 新环境变量/端口 | README、部署运维手册、`.env.example` |
| 架构或目录变化 | 架构说明、开发指南 |

## 13. 常见维护误区

- 不要直接使用 `server/config.docker.yaml` 作为当前生产配置；当前 Compose 方案从 `deploy/docker-dev/config.init.yaml` 和 `.env` 生成运行配置。
- 不要只给角色菜单权限而遗漏 API 权限。
- 不要把资产草稿当作已生效状态。
- 不要混淆报废业务类型 `scrap`、终态 `retired` 与处置位置类型 `disposal`。
- 不要使用元单位浮点数累计发票金额。
- 不要在删除数据库记录后忽略对象存储残留。
- 不要在已有未提交工作区中使用 `git add .`。
