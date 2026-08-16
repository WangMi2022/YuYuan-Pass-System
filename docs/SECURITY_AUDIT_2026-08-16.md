# 系统安全审计报告（不含 HTTPS/TLS）

- **审计日期**：2026-08-16
- **审计对象**：`gin-vue-admin`（`D:\资产管理系统\gin-vue-admin`）
- **审计类型**：代码、配置、依赖与已运行实例的防守性安全审计
- **报告状态**：初审报告；代码复测结果见 `SECURITY_RETEST_2026-08-17.md`

> 本报告明确**不评估 HTTPS/TLS**，包括证书、协议版本、密码套件、HSTS、明文 HTTP 到 HTTPS 跳转等问题。报告只覆盖应用、API、鉴权、文件、AI、依赖、容器暴露面和数据完整性。

## 1. 执行摘要

本次审计发现 13 项需要处理的风险，其中：

| 定性等级 | 数量 | 说明 |
|---|---:|---|
| 严重（条件性） | 1 | 新部署或数据库未初始化时，初始化接口可能被抢先接管 |
| 高 | 5 | 匿名对象读取、持久化 XSS、对象覆盖、前端生产依赖、Go 依赖/标准库 |
| 中 | 6 | Swagger/API 暴露、旧 JWT 不失效、Copilot 审计泄露、AI 无资源保护、匿名错误灌库、无界上传 |
| 低至中 | 1 | 公开 Casbin 刷新接口可被资源消耗 |

初审时最需要立刻处理的是：

1. **收紧资产图片代理**：不能让调用者仅凭同一对象前缀和自带 `key` 读取任意对象。
2. **禁用 SVG 上传或服务端安全转换**：当前本地存储路径与同源静态访问组合后，存在持久化 XSS 条件。
3. **封闭初始化、Swagger 和 API 端口**：初始化接口必须一次性、受控；Swagger 不应匿名公开；生产服务不应直接发布到所有网卡。
4. **为通用上传和 AI 接口增加配额/限流/体积上限**：避免认证用户消耗存储、数据库、模型预算和连接资源。
5. **升级前端与 Go 依赖**：扫描结果显示前端生产依赖存在 3 个 critical、10 个 high；Go 官方扫描发现 33 个可达漏洞。

生产核验显示：当前实例已完成初始化，生产 JWT 不是默认 `qmPlus`，独立 MCP 未启用；因此部分风险是**部署基线或故障恢复场景风险**，不是当前线上可直接利用的结果。代码路径仍需修复，避免新环境和故障切换时重新暴露。

## 2. 范围、假设与排除项

### 2.1 覆盖范围

- 后端公开/私有路由、JWT、Casbin、登录与会话失效。
- 文件上传、对象存储、文件代理、文档、发票和资产图片链路。
- AI Gateway、Copilot、MCP、AI 审计和资源保护。
- Swagger、API 端口、CORS 配置、Docker Compose/生产 Dockerfile。
- 前端生产依赖和 Go 模块/标准库依赖。
- 已运行实例的只读 HTTP/端口核验。

### 2.2 明确排除

- HTTPS/TLS 证书与私钥管理。
- TLS 版本、密码套件、HSTS、HTTP 明文传输和反向代理 TLS 终止。
- 未提供源代码或凭据的第三方云控制台配置。
- 不对数据库、MinIO 中业务数据做破坏性测试。

### 2.3 审计假设

- 以仓库代码为主，生产核验只发起无副作用请求。
- 发现“可利用”表示代码路径满足条件，不代表已对真实业务数据执行读取、覆盖或删除。
- 未使用默认凭据登录，也未尝试修改生产数据。

## 3. 方法与证据

### 3.1 使用的检查方法

1. 静态阅读路由、中间件、服务、上传和配置代码。
2. 追踪从公开入口到数据/对象/模型调用的完整调用链。
3. 检查 Docker Compose、Dockerfile、生产监听地址和默认配置。
4. 运行 `npm audit --omit=dev --json`。
5. 运行 Go 官方 `govulncheck ./...`。
6. 尝试运行 `gosec`；因 2026 年模块发布路径与模块声明不匹配未成功，详见限制项。
7. 对已运行实例执行端口和 Swagger 的只读核验。

### 3.2 生产核验快照

| 项目 | 结果 |
|---|---|
| 初始化状态 | `/init/initdb` 返回 `{"code":7,"msg":"已存在数据库配置"}`，当前已初始化 |
| JWT 默认值 | 生产 `JWT_DEFAULT=NO`，未使用示例 `qmPlus` |
| Server 监听 | `0.0.0.0:8888` |
| Swagger | `/swagger/index.html` 匿名返回 HTTP `200` |
| OSS | `minio` |
| 独立 MCP | `separate: false`，Compose 未发布 8889 |
| CORS | 配置有 `allow-all`，但 Gin CORS 中间件在 `server/initialize/router.go:56-58` 被注释，未作为当前启用漏洞 |

## 4. 风险分级规则

- **严重**：满足条件后可接管系统初始化、控制管理员面或造成大范围数据/权限失控。
- **高**：可跨业务边界读取/修改数据、执行持久化脚本，或影响生产依赖的关键安全属性。
- **中**：需要认证或特定部署条件，可能导致敏感信息泄露、资源耗尽或会话控制失效。
- **低至中**：影响较窄，但可被自动化滥用或作为其他攻击的放大器。

等级为定性判断，**未伪造 CVSS 分数**。最终等级应结合生产网络边界、账号权限、数据敏感度和监控能力复评。

## 5. 风险总览

| 编号 | 等级 | 风险 | 影响面 | 当前状态 |
|---|---|---|---|---|
| SEC-01 | 严重（条件性） | 未初始化实例可被抢先接管 | 新部署、DB 故障恢复 | 当前生产已初始化 |
| SEC-02 | 高 | 公开资产图片代理可读取同前缀任意对象 | 发票、文档、资产图片等 OSS 对象 | 代码路径成立 |
| SEC-03 | 高 | SVG 上传与同源静态访问形成持久化 XSS | 本地存储/同源代理部署 | 当前生产 OSS，仍需修复代码基线 |
| SEC-04 | 高 | MinIO 对象键碰撞导致同名文件覆盖 | 文档、资产图片 | 代码路径成立 |
| SEC-05 | 高 | 前端生产依赖含已知漏洞 | 浏览器端、导入/渲染链路 | 20 项生产依赖告警 |
| SEC-06 | 高 | Go 依赖与标准库存在可达漏洞 | 服务端请求、解析、压缩、导入 | 33 项可达漏洞 |
| SEC-07 | 中 | API 端口和 Swagger 直接暴露 | 网络攻击面、接口枚举 | 线上已核验 |
| SEC-08 | 中 | 禁用/删除/变更角色后旧 JWT 继续有效 | 会话、权限变更 | 代码路径成立 |
| SEC-09 | 中 | Copilot 问答进入通用操作日志 | 业务数据、问题内容、回答 | 代码路径成立 |
| SEC-10 | 中 | 智能接口未接入 AI 限流与 SSE 并发保护 | 数据库、模型预算、连接数 | 代码路径成立 |
| SEC-11 | 中 | 匿名错误日志接口可灌库 | 数据库容量、日志噪声 | 公开路由成立 |
| SEC-12 | 低至中 | 公开 Casbin 刷新可被滥用 | CPU、数据库/策略加载 | 公开路由成立 |
| SEC-13 | 中 | 通用媒体上传无大小上限 | 存储、带宽、进程内存 | 认证用户可触发 |

## 6. 详细发现

### SEC-01：未初始化实例可被抢先接管（严重，条件性）

**证据**：`server/initialize/router.go:78` 注册公开初始化路由；`server/router/system/sys_initdb.go:12` 无鉴权；`server/api/v1/system/sys_initdb.go:22` 进入初始化 API；`server/service/system/sys_initdb.go:100` 仅以 `global.GVA_DB != nil` 判断是否允许初始化。

**攻击前提**：新部署、数据库连接失败、初始化状态丢失或服务启动时 `global.GVA_DB == nil`，且攻击者能访问初始化路由。

**影响**：攻击者可抢先提交数据库连接和管理员初始化参数，建立自己的首个管理员或控制数据库配置，形成系统接管。

**当前缓解**：当前生产返回“已存在数据库配置”，所以现网初始化接口暂不可利用；但这是环境状态，不是代码层防护。

**修复建议**：

- 首次安装前由部署系统生成一次性高熵安装令牌，令牌通过环境变量或受保护文件注入，不写入日志。
- 初始化接口只在明确的 `INSTALL_MODE=true` 且未完成安装时开放；完成后持久化不可逆安装标记。
- 默认仅允许 loopback 或受控管理网段访问；生产反向代理不转发公开初始化路由。
- 初始化请求增加 CSRF/Origin 校验、失败次数限制、审计和自动过期。
- 数据库配置提交前做连接测试，成功后原子写入；失败不改变现有配置。

**验收**：无令牌、过期令牌、重复使用令牌、非允许来源均返回统一拒绝；令牌成功使用一次后立即失效；普通生产启动不再出现可公开初始化窗口。

### SEC-02：公开资产图片代理可读取同前缀任意对象（高）

**证据**：公开路由 `server/plugin/asset/router/asset.go:28`；`server/plugin/asset/api/asset.go:145` 只检查统一 `BasePath`；`server/plugin/asset/api/asset.go:186` 按调用者传入的 `key` 读取对象；`server/utils/upload/minio_oss.go:75-78` 让多个业务共享 `BasePath/date/hash.ext` 前缀；文档模型 `server/plugin/document/model/document.go:14` 返回 `fileKey`。

**攻击前提**：攻击者知道或猜到任一对象 key。对象 key 可能来自文档响应、日志、浏览器缓存、错误信息或同一租户协作信息。

**影响**：攻击者可绕过发票/文档业务权限，通过匿名 `/asset/photo?key=...` 读取其他业务对象；响应未按业务记录校验归属，也未限制真实图片 MIME，可能扩大到非图片对象读取或内容探测。

**修复建议**：

- 删除“按任意 key 代理”的公开语义，改为 `GET /asset/:id/photo/:photoId`，服务端从资产照片记录取得对象 key。
- 先执行 JWT、租户/部门和 Casbin 校验，再读取对象；不要信任请求中的对象 key。
- 对返回对象做业务类型校验、允许的 MIME 白名单和 `Content-Disposition: inline`/`attachment` 明确设置。
- MinIO 使用私有 bucket；通过短时、单对象、最小权限 presigned URL 或后端流式代理返回。
- 增加越权回归测试：用户 A 的资产照片、发票和文档 key 不能被用户 B 或匿名读取。

**验收**：只传 `id/photoId` 可访问本人有权照片；任意伪造 `key`、跨业务 key、目录变体和 URL 编码变体均返回 403/404，且不泄露对象是否存在。

### SEC-03：SVG 上传与同源静态访问形成持久化 XSS（高）

**证据**：`server/api/v1/example/exa_file_upload_download.go:94-99` 允许 `.svg`；`server/api/v1/example/exa_file_upload_download.go:79-81` 只检查前 4096 字节出现 `<svg`，没有清理脚本、事件属性或外链；`server/utils/upload/local.go:31-69` 保留扩展名并写入本地；`server/initialize/router.go:54` 匿名静态公开；前端 token 存在 `web/src/pinia/modules/user.js:23`。

**攻击前提**：攻击者具备通用上传权限，且部署使用本地静态目录或将对象存储通过同源可执行内容类型代理。

**影响**：上传恶意 SVG 后，受害者访问同源 URL 时脚本执行；可能窃取 localStorage token、冒用当前会话执行业务操作或钓鱼。当前生产使用 MinIO，实际 MIME/同源代理仍需在部署侧复测。

**修复建议**：

- 最小改动是完全禁止 SVG；扩展名、Content-Type 和文件签名三处同时校验。
- 若业务必须支持 SVG，使用经过审计的 sanitizer 白名单清理脚本、事件属性、`foreignObject`、外链和危险 URL，或服务端栅格化为 PNG/WebP。
- 上传内容使用不可执行域名/独立 origin；静态响应设置 `Content-Disposition: attachment` 或安全 `Content-Type`，禁止浏览器 MIME 猜测。
- 不把长期高权限 token 放在可被同源脚本读取的 localStorage；逐步迁移 HttpOnly、Secure、SameSite Cookie 或短期内存 token。

**验收**：带脚本和事件处理器的恶意 SVG fixture 被拒绝或转为安全位图；响应不会以 `image/svg+xml` 在业务同源下执行；Playwright 回归断言 `alert`、网络外发和 token 读取均不发生。

### SEC-04：MinIO 对象键碰撞导致同名文件覆盖（高）

**证据**：`server/utils/upload/minio_oss.go:73-78` 的 key 主要由 `BasePath/日期/MD5(原文件名).扩展名` 构成；文档服务 `server/plugin/document/service/document.go:137` 和资产 API `server/plugin/asset/api/asset.go:133` 直接传原文件名。发票 `server/plugin/invoice/service/invoice.go:214-216` 和智能建档 `server/plugin/asset/service/recognition.go:144-148` 已使用 hash+UUID，属于不同情况。

**攻击前提**：同一日期、同一存储前缀下出现同名文件，或攻击者能预测并重复使用文件名。

**影响**：后上传对象覆盖先上传对象，造成跨用户证据替换、文档内容篡改、资产照片被替换或审计证据失真。

**修复建议**：

- 对所有业务统一使用 `业务/租户/年/月/日/UUID.ext`；原文件名仅存数据库展示字段。
- 对象创建使用 `If-None-Match: *`/存在性检查，发现冲突时重试随机 key，禁止覆盖写。
- 数据库记录 `objectKey`、大小、MIME、内容 hash、上传人、上传时间和版本号。
- 对历史对象做碰撞盘点，冲突对象进入隔离区，由业务确认后恢复。

**验收**：同名文件连续上传 100 次产生 100 个不同 key；重复上传不能改变既有对象内容；数据库和对象存储记录可互相校验。

### SEC-05：前端生产依赖存在已知漏洞（高）

**扫描证据**：`npm audit --omit=dev --json` 共 20 项生产依赖告警：`critical=3`、`high=10`、`moderate=5`、`low=2`。

**重点依赖**：

| 依赖 | 当前声明/版本 | 主要问题 | 处理建议 |
|---|---|---|---|
| `@form-create/designer` | `^3.2.6`，`web/package.json:21` | 间接引入旧编辑器链，存在 XSS 告警 | 清理 HTML、限制粘贴协议，升级/替换编辑器 |
| `@wangeditor/editor` / `@wangeditor/editor-for-vue` | `web/package.json:29-30` | 编辑器链存在 XSS 告警 | 清理 HTML、限制粘贴协议，升级/替换编辑器 |
| `axios` | `1.8.2`，`web/package.json:32` | DoS、原型污染、请求劫持类告警 | 按审计建议升级到 `1.19.0` 或当前兼容安全版并回归 API |
| `echarts` | `5.5.1`，`web/package.json:33` | XSS | 评估升级到 `6.1.0`；若暂不能升主版本，限制富文本/HTML 渲染输入 |
| `vue3-sfc-loader` 链 | `^0.9.5`，`web/package.json:53` | 旧 Vue compiler/PostCSS 链风险 | 禁止加载不可信 SFC，升级链路并做 CSP/沙箱隔离 |
| `x-data-spreadsheet` | `^1.1.9`，`web/package.json:55` | XSS，无直接修复版本 | 隔离到受控页面、禁用 HTML 公式/渲染，规划替换 |
| `xlsx` | `^0.18.5`，`web/package.json:56` | 原型污染、ReDoS，无 npm 修复版本 | 上传前大小/复杂度限制，Web Worker 隔离，评估替换或锁定安全 fork |

**影响**：漏洞是否可达取决于页面是否渲染不可信 HTML、导入不可信表格或加载远程 SFC；资产、发票和文档场景均存在用户上传内容，不能仅以“需要登录”降级。

**修复建议**：先升级有明确修复版本的直接依赖，再替换无修复组件；锁定 lockfile，使用 `npm ci`；将 `npm audit --omit=dev` 纳入 CI 门禁；为 HTML、表格和图表输入建立统一 sanitizer 和大小/复杂度限制。

### SEC-06：Go 依赖与标准库存在可达漏洞（高）

**扫描证据**：Go 官方 `govulncheck ./...` 报告 33 个可达漏洞，涉及 7 个模块及标准库。扫描环境为 Go `1.25.5`；多项标准库问题在 `1.25.13` 修复。仓库 `server/go.mod:3` 声明 `go 1.24.0`，生产 Docker 使用浮动 `golang:1.24-alpine`（`deploy/docker-dev/server.Dockerfile:3`）。

**重点升级方向**：

| 模块 | 当前 | 建议基线 | 备注 |
|---|---|---|---|
| Go toolchain | `1.24.x`/扫描机 `1.25.5` | 固定到包含修复的最新受支持 patch 版本 | 不使用浮动基础镜像 |
| `excelize` | `v2.9.0` | `v2.11.0` | 恶意工作表可能 OOM/Panic |
| `aws-sdk-go-v2/service/s3` | `v1.96.1` | `v1.97.3` | 对象存储请求链 |
| `golang.org/x/text` | `v0.32.0` | `v0.39.0` | 文本解析链 |
| `golang.org/x/net` | `v0.48.0` | 至少 `v0.55.0` | 网络/HTTP 链 |
| `ulikunitz/xz` | `v0.5.12` | `v0.5.15` | 压缩解码链 |
| `go-redis/v9` | `v9.7.0` | `v9.7.3` | Redis 链 |
| `pgx/v5` | `v5.8.0` | `v5.9.2` | 扫描调用链主要落在 seed 工具 |

**修复建议**：先升级 Go patch 版本和直接可达模块，重新生成 `go.sum`；对 Excel、压缩包、远程响应设置大小/耗时上限；在 CI 同时运行 `govulncheck ./...` 和锁定版本构建；若 `gosec` 模块路径问题未解决，使用可复现容器或官方发布二进制执行补充扫描。

### SEC-07：API 端口和 Swagger 直接暴露（中）

**证据**：`server/initialize/router.go:60-62` 匿名注册 Swagger；`deploy/docker-dev/docker-compose.yml:14-15` 发布端口；生产核验发现 `0.0.0.0:8888` 且 Swagger 返回 HTTP `200`。

**影响**：攻击者可匿名枚举接口、参数、模型和调试信息，降低后续攻击成本；服务端口直接暴露增加绕过反向代理安全策略的机会。

**修复建议**：生产关闭 Swagger，或仅在管理网段/管理员 JWT 下开放；Compose 默认绑定 `127.0.0.1`，由反向代理按需转发；将健康检查与业务 API 分离；部署防火墙仅开放必要入口。

**验收**：公网请求 Swagger 返回 404/403；外部无法直连 8888；内部反向代理和健康检查保持正常。

### SEC-08：禁用/删除/角色变更后旧 JWT 不即时失效（中）

**证据**：用户状态回查在 `server/middleware/jwt.go:47-54` 被注释；登录时仍在 `server/api/v1/system/sys_user.go:82` 检查状态；示例配置 `server/config.docker.yaml:6` 的 token 有效期为 7 天。

**影响**：已登录用户被禁用、删除、降权或移出部门后，旧 token 可能继续按旧 claims 调用接口，直到自然过期或命中黑名单。

**修复建议**：

- 在用户表增加 `tokenVersion`/`securityVersion`，JWT 携带版本；禁用、删除、改角色、重置密码时递增版本。
- 中间件每次请求校验版本；高风险操作可叠加短期 access token 与 refresh token 轮换。
- 保留明确的 token 黑名单/撤销接口，避免把所有会话状态放进高延迟查询。

**验收**：禁用/删除/降权后旧 token 在下一次请求即返回 401/403；普通请求延迟和数据库负载在基线范围内。

### SEC-09：Copilot 问答进入通用操作日志（中）

**证据**：`server/plugin/smart/router/smart.go:24-27` 使用 `OperationRecord()`；`server/middleware/operation.go:37-42,67-90,165-189` 记录请求和响应前 1024 字节；`server/middleware/operation.go:30` 脱敏字段未覆盖 `question`、回答和业务字段。已有 `server/middleware/ai_security.go:88-107` 可隐藏正文，但 Copilot 路由没有使用该专用记录链。

**影响**：资产编号、发票信息、人员信息、自然语言问题和模型回答可能被复制到通用审计表，扩大敏感数据留存面、备份面和管理员可见范围。

**修复建议**：Copilot 改用 `AIOperationRecord`，只保留请求 hash、模型、租户、耗时、token、结果码和风险标签；正文放入短期加密存储并按最小权限访问，默认不写通用 operation record；完善字段级脱敏和保留期限。

**验收**：通用操作日志不出现问题正文、回答正文或业务字段；AI 审计可按 request hash 关联，但无法从普通审计页面还原原文。

### SEC-10：智能接口未接入 AI 限流与 SSE 并发保护（中）

**证据**：`server/plugin/smart/initialize/router.go:11-13` 仅初始化 JWT/Casbin；`server/plugin/smart/router/smart.go:11-35` 未使用 `AISecurity`/`AISSEConcurrency`；保护中间件已存在于 `server/middleware/ai_security.go:43-85`。

**影响**：认证用户可高频调用 Copilot、日报或 SSE，消耗数据库、模型预算、并发连接和队列；多个普通账号可放大为低成本资源耗尽。

**修复建议**：统一接入 AI Gateway 和 AI 安全中间件；按用户、租户、IP、模型和接口配置 QPS、并发、每日 token/费用配额；SSE 设置连接超时、最大输出 token、断开清理和全局并发上限；超限返回可观测的 429/430 业务码。

**验收**：压测超过阈值时稳定限流；断开 SSE 后连接、goroutine、数据库游标和模型请求均释放；审计能够核对配额消耗。

### SEC-11：匿名错误日志接口可灌库（中）

**证据**：公开路由 `server/router/system/sys_error.go:14,26`；`server/api/v1/system/sys_error.go:30-42` 无身份、限流或体积限制；`server/service/system/sys_error.go:18-23` 直接写 DB；模型包含多个 text 字段（`server/model/system/sys_error.go:11-15`）。

**影响**：匿名攻击者可批量写入大文本，制造数据库膨胀、慢查询、告警噪声和审计污染；若错误内容在后台展示，还可能形成存储型 XSS。

**修复建议**：改为认证后上报或仅允许受控前端错误收集令牌；限制单条和单 IP/设备频率、字段长度、每日配额；服务端统一截断、清理控制字符和 HTML；异步写入有界队列，失败不阻塞主请求。

**验收**：匿名批量请求被拒绝/限速；单字段超过上限被截断或拒绝；压力测试不会让数据库增长失控。

### SEC-12：公开 Casbin 刷新可被滥用（低至中）

**证据**：`server/router/system/sys_api.go:14,33` 公开路由；`server/api/v1/system/sys_api.go:315-322` 每次请求执行策略刷新。

**影响**：匿名高频刷新可能造成 CPU、数据库和策略加载资源消耗；与其他接口滥用叠加时扩大拒绝服务窗口。

**修复建议**：移入管理员私有路由；增加最小刷新间隔、分布式锁和幂等缓存；配置变更成功后由服务端内部事件触发刷新，不接受匿名手工刷新。

**验收**：匿名返回 401/403；管理员连续刷新在间隔内只产生一次实际加载；多实例不会重复风暴。

### SEC-13：通用媒体上传无文件大小上限（中）

**证据**：上传接口 `server/api/v1/example/exa_file_upload_download.go:33-48` 只校验类型；没有 `file.Size`/`MaxBytesReader`；本地写入 `server/utils/upload/local.go:64` 使用无界 `io.Copy`；MinIO 路径 `server/utils/upload/minio_oss.go:90` 按完整大小写入。默认 Casbin 中普通认证用户可调用该接口（`server/model/system/request/sys_casbin.go:24`）。

**影响**：认证用户可灌满对象存储和带宽，上传大文件时增加进程内存、临时磁盘和请求占用；多账号并发可造成服务降级。

**修复建议**：统一请求体、单文件、批次和用户/租户日配额；使用流式限制读取，不信任客户端 `Content-Length`；先传临时对象，病毒/类型/大小检查通过后再提交；失败和超时清理临时对象。

**验收**：超过 10 MB（或业务配置值）的文件在读取前被拒绝；并发压测下存储和连接有硬上限；异常中断不会遗留可访问临时对象。

## 7. 条件性风险与未列为当前漏洞的项目

### 7.1 独立 MCP 未启用，但代码基线需加固

`server/cmd/mcp/main.go:22-32` 监听全接口；`server/mcp/server.go:35-49` 未显式拒绝未认证请求；token 仅放入 context（`server/mcp/context.go:15-18`）。当前配置 `separate:false`，Compose 未发布 8889，因此本次不把它列为当前线上风险。若未来启用独立 MCP，必须在监听、反向代理和 handler 三层做认证、来源限制、工具白名单和审计。

### 7.2 默认 JWT `qmPlus` 是部署基线风险

示例/初始化配置仍存在默认值（`deploy/docker-dev/config.init.yaml:6`、`server/config.docker.yaml:5`），但生产已核验不是默认值。应在启动时拒绝默认 secret，而不是只依赖运维记忆。

### 7.3 CORS `allow-all` 当前未启用

配置存在宽松值，但 Gin CORS 中间件被注释；本报告不将其误报为当前已启用漏洞。若重新启用，必须改为明确 allowlist，且不能与凭据跨域同时使用通配来源。

### 7.4 `downloadOnlineSkill` 不构成当前 SSRF/RCE

公开路由 `server/router/system/sys_skills.go:33` 对应服务层固定返回禁用错误（`server/service/system/sys_skills.go:386-387`），没有实际下载执行链，因此不列为 SSRF/RCE。

## 8. 已确认的安全控制

- AI Provider 默认拒绝私网地址，解析后逐 IP 检查并禁用跳转（`server/ai/transport.go`）。
- AI Gateway 强制身份和 Casbin 二次校验（`server/ai/gateway.go`）。
- AI 专用审计默认只保存 hash/元数据，不直接保存正文。
- 发票和智能建档上传已有约 10 MB 级别限制和随机对象 key。
- Zip Slip 有 `..` 与目标目录校验。
- 未发现 pprof 暴露。
- 登录流程会检查用户状态（`server/api/v1/system/sys_user.go:82`）。

这些控制降低了部分攻击面，但不能替代本报告列出的业务对象授权、会话撤销、上传隔离和依赖升级。

## 9. 限制与复测要求

1. 未执行破坏性越权读取、对象覆盖或大文件压测；SEC-02、SEC-04、SEC-13 应在隔离环境用合成数据复测。
2. 生产 MinIO 的最终响应 MIME、bucket policy、反向代理同源关系需由部署侧补充核验。
3. `gosec` 未成功运行：2026 年当前模块发布路径与模块声明不匹配，`latest` 和固定 `v2.22.8` 均失败；需使用可复现二进制/容器重新执行。
4. `npm audit` 和 `govulncheck` 反映扫描时依赖图；合并升级后必须重新生成报告。
5. 未审计外部 IdP、云控制台、主机防火墙、备份权限和员工终端安全。
6. 生产核验为单时点快照，不能替代持续监控。

## 10. 结论

当前系统并非“无鉴权裸奔”：JWT、Casbin、AI Gateway SSRF 防护和部分上传限制已经存在。初审识别出的对象代理授权、SVG 内容安全、初始化窗口、对象 key、会话撤销、AI/上传资源边界与高危依赖缺口已进入整改；代码复测、剩余依赖例外与生产验收结论以 `SECURITY_RETEST_2026-08-17.md` 为准。HTTPS/TLS 不在本报告结论内。
