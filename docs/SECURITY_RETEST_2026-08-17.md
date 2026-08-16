# 安全整改复测报告（不含 HTTPS/TLS）

- **复测日期**：2026-08-17
- **对应初审**：`SECURITY_AUDIT_2026-08-16.md`
- **整改计划**：`SECURITY_REMEDIATION_PLAN_2026-08-16.md`
- **复测范围**：应用代码、路由与鉴权、AI/MCP、上传与对象访问、Go/Node 依赖、Docker 部署基线
- **明确排除**：HTTPS/TLS、证书、反向代理 TLS 配置、云账号和主机防火墙

## 1. 结论

初审列出的 13 项应用层风险均已完成代码整改。Go 官方漏洞扫描在锁定 Go `1.25.13` 后显示 **0 个可达漏洞**；前端生产依赖扫描显示 **0 critical、0 high、2 moderate**。生产提交 `a9d7e89516bba9e4d4614358979910433c453caf` 已于 2026-08-17 发布，远程上线门禁 8 项全部通过。

剩余的 2 个中危告警来自 `exceljs@4.4.0` 的传递依赖 `uuid@8.3.2`，对应 `GHSA-w5hq-g745-h8pq`。npm 给出的唯一自动修复路径是降级到 `exceljs@3.4.0`，属于破坏性回退，不能未经导出/导入回归直接执行。该问题没有被隐瞒或标为已关闭，已列入 P2 依赖治理项。

## 2. 整改状态

| 编号 | 初审风险 | 状态 | 已落地控制 |
|---|---|---|---|
| SEC-01 | 初始化抢先接管 | 已修复 | 必须显式 `GVA_INSTALL_MODE=true`、仅 loopback、固定时间比较安装令牌、串行化、成功后令牌失效、初始化管理员密码至少 12 字符、弱 JWT 密钥拒绝启动/初始化。 |
| SEC-02 | 任意对象 key 读取资产图片 | 已修复 | 图片读取改为私有 API；需要已认证身份，并验证资产归属或短期、按用户绑定的图片令牌；限制对象路径、MIME、大小并启用 `nosniff`。 |
| SEC-03 | SVG 持久化 XSS | 已修复 | 通用上传不再允许 SVG 扩展名或 SVG 内容；默认关闭本地上传目录静态公开。 |
| SEC-04 | 对象同名覆盖 | 已修复 | 本地和 MinIO 对象名加入 UUID，避免同日同名文件覆盖。 |
| SEC-05 | 前端高危/严重依赖 | 部分关闭 | 已移除动态 SFC、表格和表单设计器风险链，升级 Axios/ECharts，生产扫描无 critical/high；ExcelJS 传递依赖仍待治理。 |
| SEC-06 | Go 可达依赖/标准库漏洞 | 已修复 | 工具链固定 `go1.25.13`，升级 Excelize、S3 SDK、MongoDB Driver、Redis、`x/*`、XZ、PGX；复扫 0 个可达漏洞。 |
| SEC-07 | Swagger 与 API 端口暴露 | 已修复 | Swagger 仅在 Debug 或 `GVA_ENABLE_SWAGGER=true` 下作为私有路由注册；服务端口默认绑定 `127.0.0.1`。 |
| SEC-08 | 禁用/改权后旧 JWT 有效 | 已修复 | JWT 带 `security_version`；中间件实时核验用户启用状态与版本，禁用、删除、改角色和改密码会使旧令牌失效。 |
| SEC-09 | Copilot 正文写入通用日志 | 已修复 | AI 路由改用 `AIOperationRecord`，只记录元数据占位，不保存提示词和模型响应正文。 |
| SEC-10 | AI 缺少限流和 SSE 并发保护 | 已修复 | Copilot、草稿、公告、日报接入请求体上限、每用户窗口限流、90 秒超时，SSE 每用户最多 2 路。 |
| SEC-11 | 匿名错误日志灌库 | 已修复 | 错误上报移入 JWT/Casbin 私有路由，增加 64KB 上限、按用户限流、字段清理/截断并丢弃客户端提交的处置状态。 |
| SEC-12 | 匿名 Casbin 刷新 | 已修复 | `freshCasbin` 移入私有路由，删除历史公开忽略规则，只为超级管理员补齐权限。 |
| SEC-13 | 通用媒体上传无边界 | 已修复 | 请求体上限 11MB、单文件上限 10MB，使用 `http.MaxBytesReader`，并保留类型签名校验。 |

## 3. 执行证据

| 检查 | 命令/方式 | 结果 |
|---|---|---|
| 后端全包编译 | `go test ./... -run '^$' -count=1` | 通过。 |
| 初始化防护 | `go test ./api/v1/system -run TestAllowDatabaseInitialization -count=1` | 通过，覆盖显式安装模式、loopback、有效令牌和已消费令牌。 |
| MCP 鉴权 | `go test ./mcp -count=1` | 通过。 |
| 资产图片 MIME | `go test ./plugin/asset/api -run TestIsAllowedAssetPhotoContentType -count=1` | 通过。 |
| JWT/中间件/智能模块 | `go test ./middleware ./utils ./plugin/smart/... -count=1` | 通过。 |
| Go 漏洞扫描 | `govulncheck ./...`，Go `1.25.13` | 0 个可达漏洞；另有 4 个仅导入和 11 个仅依赖、未调用的提示。 |
| 前端测试 | `npm test` | 31 项通过。 |
| 前端生产构建 | `npm run build` | 通过。 |
| 前端依赖扫描 | `npm audit --omit=dev --json` | `critical=0`、`high=0`、`moderate=2`。 |
| Git 差异检查 | `git diff --check` | 通过。 |

## 4. 部署基线

生产 Compose 默认采用以下安全值；发布时必须保留并由验收确认：

- `SERVER_BIND=127.0.0.1`，不再把 8888 直接暴露到所有网卡。
- `GVA_ENABLE_SWAGGER=false`，Swagger 仅由具备 Casbin 权限的管理员显式启用。
- `GVA_ENABLE_PUBLIC_UPLOADS=false`，禁止将本地上传目录匿名静态公开。
- 已初始化环境必须注入至少 32 字节的 `GVA_JWT_SIGNING_KEY`。
- 全新安装必须同时设置 `GVA_INSTALL_MODE=true` 和一次性 `GVA_INSTALL_TOKEN`，完成后应删除或轮换该令牌，并将安装模式改回 `false`。
- 独立 MCP 默认监听 `127.0.0.1`，要求 `GVA_MCP_AUTH_TOKEN` 或 `mcp.auth_token`，请求体限制为 64KB。
- Server/Web 构建镜像固定到具体版本；前端构建使用 `npm ci`。

## 5. 剩余风险与处置

### 5.1 ExcelJS / UUID 中危依赖

`exceljs@4.4.0` 通过 `uuid@8.3.2` 引入 `uuid <11.1.1` 的边界检查漏洞。当前项目不能依赖 npm 建议的 `exceljs@3.4.0` 回退，因为这会改变导入/导出行为，且不能证明回退后的整条依赖链更安全。

处置顺序：

1. 用实际业务模板验证 Excel 导入、导出、图片、公式与大文件场景。
2. 评估 ExcelJS 上游修复版本，或替换为维护活跃的导入/导出实现。
3. 在隔离分支验证可兼容的 `uuid@11.1.1` override；只有 Node/browser 构建与业务回归均通过才能采用。
4. 在 CI 保留 `npm audit --omit=dev` 门禁；在该项关闭前不把安全状态表述为“零依赖告警”。

### 5.2 运行环境复测结果

本机没有 Docker，因此 Compose 解析和镜像构建在生产 Docker 主机完成。生产执行 `./build.sh server web`、容器强制重建和 `./release-acceptance.sh`，结果为 `passed=8 failed=0`；`gva-server-dev`、`gva-web-dev` 均为 `running`，重启次数均为 `0`。本次发布备份位于 `/data/gin-vue-admin/.deploy/backups/20260817073213-a9d7e89/`。还必须确认 MinIO/RustFS bucket 仍为私有策略；代码不再生成匿名通用对象 URL，但 bucket 策略属于运行时配置，不能从仓库推断。

### 5.3 未纳入本轮的边界

本报告不涉及 HTTPS/TLS。对象级部门/租户数据隔离需要基于业务组织模型进一步设计；当前整改保证图片不能凭任意对象 key 匿名读取，不替代尚未定义的行级权限模型。

## 6. 发布验收记录

本次发布已执行以下只读核验，验收通过后才写入部署版本标记：

1. `docker compose` 构建和 `up -d --force-recreate server web` 成功。
2. `./release-acceptance.sh` 返回 `0`，`./ps.sh` 显示 Server 和 Web 正常运行。
3. `/health`、`/api/health`、Web 首页与当前 Vite 资源均通过验收。
4. 端口映射显示 Server 仅绑定 `127.0.0.1:8888`。
5. 部署标记已写入完整提交 SHA；匿名接口探测和 bucket 私有策略仍按运维窗口执行专项复核。

## 7. 复审计划

- 发布当日：完成上述运行时验收并保存容器状态和响应码。
- 7 天内：完成 ExcelJS/UUID 兼容替换或形成带到期日的风险接受记录。
- 每月：重跑 `govulncheck ./...`、`npm audit --omit=dev` 和镜像基线检查。
- 每季度：复审对象存储权限、MCP 暴露面、AI 配额和审计保留周期。
