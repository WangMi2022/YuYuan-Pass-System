# 品牌重命名生产迁移手册

更新时间：2026-11-25（Asia/Shanghai）

本次改动把仓库内所有 gin-vue-admin 二开痕迹统一重命名为 mit-assets-admin。源码改动已完成，但其中一部分涉及**生产运行时资源**，发布时必须按本文顺序操作，否则会导致服务无法启动、用户登录失效或数据找不到。

## 1. 改动范围总览

| 类别 | 旧值 | 新值 | 是否影响生产运行时 |
| --- | --- | --- | --- |
| Go 全局变量 | `global.GVA_DB` 等 2364 处 | `global.MAA_DB` | 否，编译期 |
| 环境变量键名 | `GVA_PG_HOST`、`GVA_JWT_SIGNING_KEY` 等 28 个 | `MAA_*` | **是**，`.env` 与 `docker-compose.yml` 必须同步 |
| Compose 项目名 / 镜像名 / 容器名 | `gin-vue-admin-dev`、`gva-server-dev` | `mit-assets-admin-dev`、`maa-server-dev` | **是**，容器会被重建 |
| 前端 CSS 类名 | `gva-table-box` 等 25 个 | `maa-*` | 否，随构建产物 |
| 前端 iconfont | `gvaIcon` / `.gvaIcon-*` | `maaIcon` | 否 |
| 路由 name | `GvaLayoutIframe` | `maaLayoutIframe` | 否 |
| localStorage / sessionStorage 键 | `gva_ai_workflow_settings`、`gva:chunk-load-reload-at` | `maa_*` | 轻微：用户本地工作流设置会重置 |
| JWT Audience | `"GVA"` | `"MAA"` | **是**，所有已签发 token 失效，用户需重新登录 |
| JWT Issuer（配置默认值） | `qmPlus` | `mit-assets-admin` | **是**，见 §3.2 |
| MCP 工具名 | `gva_analyze`、`gva_execute`、`gva_review` | `maa_*` | **是**，外部 MCP 客户端配置需同步 |
| MCP 服务名（配置键值） | `GVA_MCP` | `MAA_MCP` | **是**，`config.yaml` 需同步 |
| autocode 模块路径 | `github.com/flipped-aurora/gin-vue-admin/server` | `github.com/WangMi2022/mit-assets-admin/server` | **是**，`config.yaml` 需同步 |
| autocode 请求字段 | `gvaModel` | `maaModel` | 否，前后端已同步 |
| 生产目录 | `/data/gin-vue-admin` | `/data/mit-assets-admin` | **是**，见 §4 |
| SSH 别名 | `gin-vue-admin-remote` | `mit-assets-admin-remote` | **是**，见 §4 |
| 原作者署名 | `piexlmax`、`SliverHorn` 等 | `MiTo` | 否 |
| 默认头像外链 | `qmplusimg.henrongyi.top` | `/logo.png` | 否，仅新用户默认值 |

### 有意保留、不要改的项

| 项 | 位置 | 原因 |
| --- | --- | --- |
| 数据库表名 `gva_announcements_info`、`gva_announcement_reads` | `server/plugin/announcement/`、验收脚本 | 生产库中已存在的真实表，改名需要 `ALTER TABLE`，本次不涉及 |
| 生产数据库名 `gva` | 生产 `.env` 的 `MAA_PG_DB=gva` | 真实数据库，改名需迁移数据库 |
| 生产 S3 桶名 `gva-assets` | 生产 `.env` 的 `MAA_RUSTFS_BUCKET`、`config.yaml` 的 `bucket-name`/`bucket-url` | 真实桶，改名需迁移对象 |
| 弱密钥黑名单中的 `"qmplus"` | `server/utils/jwt_secret.go` | 用于检测沿用旧默认密钥，删除会削弱安全检查 |
| `NOTICE` 文件 | 仓库根目录 | MIT 协议要求的上游版权声明，**不要删除** |
| 安全审计文档中的 `qmPlus` 描述 | `docs/SECURITY_*.md` | 历史审计记录的事实描述 |

## 2. 发布前：在 WSL 仓库完成验证

当前改动是在容器副本上完成的，需先同步回 `D:\资产管理系统\mit-assets-admin`。进入仓库后：

```bash
cd /mnt/d/资产管理系统/mit-assets-admin
git status -sb                       # 确认在 main
gofmt -l server/                     # 字段名变短可能导致结构体 tag 对齐变化，如有输出则 gofmt -w
cd server && go build ./... && go vet ./...
go test ./plugin/systemsetting/... ./utils/... ./mcp/...
cd ../web && npm ci && npm run lint && npm run build
cd .. && git diff --check
```

`go test ./...` 的既有失败（MCP 依赖本机 8888、autocode 缺数据库、AST 测试引用已移除的旧插件目录）与本次无关，以局部测试和编译为准。

## 3. 生产 `.env` 与 `config.yaml` 同步

生产服务器上这两个文件是 Git 忽略的机密文件，**不会随发布包覆盖**，必须手动同步键名。只改键名，**不要改任何值**。

### 3.1 `.env`

```bash
cd /data/mit-assets-admin/deploy/docker-dev
cp .env ../.env.bak-rename-$(date +%Y%m%d%H%M%S)   # 备份到 .deploy 之外的位置亦可
sed -i -E \
  -e 's/^# gin-vue-admin Docker dev runtime/# mit-assets-admin Docker dev runtime/' \
  -e 's/^PROJECT_NAME=gin-vue-admin-dev/PROJECT_NAME=mit-assets-admin-dev/' \
  -e 's/^SERVER_IMAGE=gin-vue-admin-server:dev/SERVER_IMAGE=mit-assets-admin-server:dev/' \
  -e 's/^WEB_IMAGE=gin-vue-admin-web:dev/WEB_IMAGE=mit-assets-admin-web:dev/' \
  -e 's/^SERVER_CONTAINER=gva-server-dev/SERVER_CONTAINER=maa-server-dev/' \
  -e 's/^WEB_CONTAINER=gva-web-dev/WEB_CONTAINER=maa-web-dev/' \
  -e 's/^GVA_/MAA_/' .env
```

改完后核对：`grep -oE '^[A-Z_]+=' .env | sort` 应与仓库 `.env.example` 的键集合一致。`MAA_PG_DB`、`MAA_RUSTFS_BUCKET` 的值保持原样。

### 3.2 `config.yaml`

```bash
cp config.yaml config.yaml.bak-rename-$(date +%Y%m%d%H%M%S)
sed -i -E \
  -e 's|github.com/flipped-aurora/gin-vue-admin/server|github.com/WangMi2022/mit-assets-admin/server|g' \
  -e 's/^(\s*name: )GVA_MCP$/\1MAA_MCP/' \
  -e 's/^(\s*issuer: )qmPlus$/\1mit-assets-admin/' \
  config.yaml
```

`signing-key`、`bucket-name`、`bucket-url`、数据库名**不动**。

> JWT `issuer` 改名后，所有已登录用户的 token 校验会失败，发布后需通知用户重新登录。`audience` 从 `GVA` 改为 `MAA` 的影响相同。

## 4. 生产目录与 SSH 别名迁移

这一步是停机操作，安排在发布窗口内执行。

### 4.1 WSL 侧：SSH 别名

编辑 `~/.ssh/config`，把 `Host gin-vue-admin-remote` 改为 `Host mit-assets-admin-remote`，其余（HostName、User、Key）不变。`AGENTS.md`、`REMOTE.md`、`connect-remote.sh` 已指向新别名。

### 4.2 服务器侧：目录改名

```bash
# 1. 停止应用容器（保留数据卷）
cd /data/gin-vue-admin/deploy/docker-dev
docker compose --env-file .env -f docker-compose.yml down

# 2. 目录改名
cd /data
mv gin-vue-admin mit-assets-admin

# 3. 先完成 §3 的 .env / config.yaml 键名同步，再继续
cd /data/mit-assets-admin/deploy/docker-dev
```

> Compose 项目名从 `gin-vue-admin-dev` 变为 `mit-assets-admin-dev` 后，`docker compose down` 用旧项目名、`up` 用新项目名，旧容器和网络会被清理，镜像名也变化，因此必须完整重建镜像，不能只 `restart`。

## 5. 发布流程

目录改名和配置同步完成后，按 `AGENTS.md` 标准流程发布。本次改动同时涉及 `server/**`、`web/**` 和部署脚本，**必须重建两个服务**：

```bash
# WSL：从已提交的 revision 生成发布包并直传
git archive --format=tar.gz --prefix=mit-assets-admin/ -o - <commit-sha> \
  | ssh mit-assets-admin-remote 'cat > /data/mit-assets-admin/workspace/releases/release-<short-sha>.tar.gz'

# 服务器
cd /data/mit-assets-admin
mkdir -p .deploy/backups/<timestamp>-<short-sha>
tar czf .deploy/backups/<timestamp>-<short-sha>/pre-rename.tar.gz server web deploy/docker-dev --exclude=deploy/docker-dev/.env --exclude=deploy/docker-dev/config.yaml
tar xzf workspace/releases/release-<short-sha>.tar.gz --strip-components=1 -C /data/mit-assets-admin

cd deploy/docker-dev
./build.sh server web
docker compose --env-file .env -f docker-compose.yml up -d --force-recreate server web
./release-acceptance.sh
./ps.sh
```

`release-acceptance.sh` 返回 `0` 后，写入完整 commit hash 到 `/data/mit-assets-admin/.deploy/current-commit`。任一步失败，保留备份、停止并报告失败阶段。

## 6. 发布后核验

| 检查项 | 期望 |
| --- | --- |
| `docker ps` 容器名 | `maa-server-dev`、`maa-web-dev` |
| `http://192.166.20.103:8080/` | 登录页正常加载，品牌显示 mit-assets-admin |
| 重新登录 | 旧 token 失效属预期，重新登录后正常 |
| Swagger（若启用） | 标题为 `mit-assets-admin Swagger API 接口文档` |
| MCP（若启用独立服务） | 外部客户端工具名改为 `maa_analyze`、`maa_execute`、`maa_review` |
| 浏览器 Network 面板 | 无 `gva-` 开头的 class 残留（旧缓存可强刷） |
| `.env` 中 `MAA_PG_DB`、`MAA_RUSTFS_BUCKET` 值 | 与改名前一致，数据正常读写 |

## 7. 回滚

如需回滚，恢复 §3 的 `.env`、`config.yaml` 备份，把目录改回 `/data/gin-vue-admin`，用 `.deploy/backups/<timestamp>-<short-sha>/pre-rename.tar.gz` 恢复源码后重建容器。数据库名、表名、桶名本次均未改动，无需回滚数据。

## 8. 改动正确性审计记录（2026-11-25）

本次重命名在容器副本上完成，无 Go/Node 工具链，因此用 Python 做了静态审计。审计脚本保存在 `docs/rebrand-audit/`，可重复执行。

### 8.1 基线比对

以 `.deploy/backups/20260911145857-d7e1eea`（改名前最后一个完整备份）为基线，对备份文件施加"预期替换规则"后与当前文件逐行比对：

| 指标 | 结果 |
| --- | --- |
| 基线以来有改动的文件 | 416 |
| 被替换规则**完全解释**的文件 | 392 |
| 剩余 24 个文件 | 17 个为审计规则书写误差（当前文件正确）；7 个为 9 月 11 日之后的业务开发（`plugin/smart` 澄清选项、copilot 头像动效），与重命名无关 |

### 8.2 跨层一致性

| 检查项 | 结果 |
| --- | --- |
| 所有 `global.MAA_*` 引用在 `global` 包有声明（14 变量 + 1 类型） | ✓ |
| Go 源码无残留 `GVA_`/`Gva*`/`gva_*` 标识符（DB 表名除外） | ✓ 0 处 |
| 无残留 `GVA_*` 环境变量引用（compose / sh / Go） | ✓ |
| JWT `audience` 签发方与 API Token 签发方均为 `MAA`，`issuer` 读配置 | ✓ |
| MCP 工具注册名 = 提示词引用名 = `maa_analyze/execute/review` | ✓ |
| `maaModel` JSON 字段后端 tag 与前端 7 处使用一致 | ✓ |
| iconfont family `maaIcon` 与类名、5 处使用一致，无 `gvaIcon` 残留 | ✓ |
| 路由 name 只有 `MaaLayoutIframe` 一种写法 | ✓ |
| 前端 `maa-*` 类名"已用未定义"集合与备份中 `gva-*` 完全一致 | ✓ |
| `GvaModel → MaaModel` 等长替换，结构体 tag 对齐不变，**gofmt 无需改动** | ✓ |

### 8.3 机密文件值完整性

与 `.deploy/backups/20260817114601-b10d95e0548e` 的 `.env` 逐键比对（键名对齐后）：

- 26 个键一一对应，无增减
- 值不同的**仅 5 个标识键**：`PROJECT_NAME`、`SERVER_IMAGE`、`WEB_IMAGE`、`SERVER_CONTAINER`、`WEB_CONTAINER`
- 所有 `*_PASSWORD`、`*_SECRET_KEY`、`*_ACCESS_KEY`、`JWT` 密钥、数据库/Redis/S3 地址**完全未动**

### 8.4 仍需在有工具链的环境完成

本文档 §2 列出的 `go build`、`go vet`、`go test`、`npm run lint`、`npm run build` 仍是发布前必要步骤。静态审计能证明引用一致，不能替代编译。

### 8.5 本地 Docker 构建与运行时冒烟（2026-11-25）

在具备 Docker 的宿主上完成了镜像构建与冒烟，不依赖生产环境：

| 项目 | 结果 |
| --- | --- |
| `./build.sh server web` | 两个镜像均 `Successfully tagged`：`mit-assets-admin-server:dev`（113MB）、`mit-assets-admin-web:dev`（61.9MB） |
| server 二进制 | `/app/server`（80MB）正常产出，Dockerfile 中 `maa-server` 构建名生效 |
| server 启动日志 | 前缀 `[mit-assets-admin]`，模块路径 `github.com/WangMi2022/mit-assets-admin/server/...`；在 Redis 连接处按预期失败（本地无 Redis），配置加载、日志、初始化路径均正常 |
| web `index.html` | `<title>mit-assets-admin</title>`，`maa-loading-box` 2 处，`gva` 0 处 |
| web CSS 产物 | `maa-*` 类 121 处，`gva-*` 0 处，`maaIcon` 11 处 |
| web `/api` 反代 | nginx → `server` upstream 链路通（502 为占位 server 预期行为） |
| 全部 assets | 含 `gva-` 的文件数 0 |

该冒烟验证了重命名在真实容器运行时的行为，但不替代生产 `release-acceptance.sh`（需要真实 PG/Redis/S3）。
