# 资产管理系统任务交接

更新时间：2026-09-09（Asia/Shanghai）

这份文档用于新线程继续当前任务。开始工作前先阅读全文，再检查工作区状态；不要回退或覆盖已有的墙报改动。

## 1. 用户需求与上下文

本轮需求有两部分：

1. 用户查询“行政部的王磊名下都认领了哪些资产”时，系统应该查到实际资产数据，而不是把整句话当普通关键词搜索。
2. 业务助手处于思考/查询中时应有明显但克制的动态效果，并更换机器人头像。

用户曾多次发送“继续”，因此默认目标是把剩余前端工作完成、验证并收口。

## 2. 已完成：后端保管人查询

已完成并推送到 `origin/main` 的提交：

```text
7f51cf3cb2a5596ed1d3ef7fe03810e08fba1be5  fix: support custodian asset queries
```

本次提交涉及：

- `server/plugin/asset/service/asset.go`
  - 增加按 `custodian` 字段查询的服务方法。
  - 部门-姓名（如 `行政部-王磊`）使用精确匹配。
  - 只有姓名时支持匹配不同部门下的同名保管人。
  - 对 `%`、`_`、反斜杠等 LIKE 特殊字符做了转义。
- `server/plugin/smart/service/asset_query.go`
  - 从自然语言中识别“部门的姓名”“姓名名下”“保管人是/为”等表达。
  - 区分部门+姓名精确匹配、裸姓名匹配和汇总类问题。
- `server/plugin/smart/service/planner.go`
  - 识别到保管人后，规划为 `asset.search`，参数放在 `custodian`，不再放进普通 `keyword`。
- `server/plugin/smart/service/smart.go`
  - 读取 `custodian` 参数并调用按保管人查询；回答中带保管人和记录数。
- `server/plugin/smart/service/tool_registry.go`
  - 更新资产查询工具描述和输入 schema。
- `server/plugin/smart/service/asset_custodian_test.go`
  - 增加自然语言解析、部门+姓名隔离、裸姓名匹配和查询结果回归测试。

### 已验证结果

生产数据中 `行政部-王磊` 有 4 条资产记录，数量合计 28。生产 API 已返回以下资产编号：

```text
DEMO-ASSET-086
DEMO-ASSET-075
DEMO-ASSET-068
DEMO-ASSET-015
```

后端局部测试和编译已通过：

```bash
go test ./plugin/smart/... ./plugin/asset/...
```

生产构建、重建和发布验收已完成，`release-acceptance.sh` 结果为 `8/8` 通过。

上一轮生产备份目录：

```text
/data/gin-vue-admin/.deploy/backups/20260909162035-7f51cf3
```

### 已知的全量测试失败（与本次后端改动无关）

`go test ./...` 仍有仓库既有失败：

- MCP 测试依赖本机 `localhost:8888/sse`。
- 自动代码测试缺少数据库环境。
- AST 测试仍引用已经移除的旧 `plugin/gva`。

不要把这些失败误判为保管人查询修复回归；以局部测试和编译结果为准，并在最终报告中保留说明。

## 3. 待完成：业务助手前端

唯一明确的实现文件：

```text
web/src/plugin/smart/view/copilot.vue
```

当前状态：

- 空状态、历史助手消息、等待消息都使用 Element Plus 的 `MagicStick` 图标。
- 头像模板位置约为第 117 行和第 198 行；空状态图标约为第 91 行；导入约为第 267 行。
- `sending` 控制查询中的等待状态。
- 等待状态已有三点动画，但很小且不明显：`.assistant-dots` / `@keyframes assistant-dot`（约第 826-831、900 行）。
- 已存在 `prefers-reduced-motion: reduce` 降级规则，必须保留并覆盖新增动画。
- 可复用头像候选资源：`web/src/assets/icons/ai-gva.svg`。继续前先确认视觉效果是否适合作为机器人头像；若不合适，可在现有项目资源体系内选择更清晰的机器人图标，但不要无必要新增依赖。

建议实现方向：

- 用头像图片（优先检查并复用 `ai-gva.svg`）替换三个 `MagicStick` 头像实例；保持空状态图标是否替换的视觉一致性。
- 在等待状态增加可感知的思考动效，例如头像外围状态环/高光旋转、头像轻微呼吸，以及更清晰的进度点或光晕。
- 动效应服务于“正在思考”的状态，速度克制，不能造成布局抖动或遮挡文字。
- 使用 CSS keyframes 和现有主题变量，不引入新动画库。
- 在移动端保持头像尺寸和布局稳定；确保长文案不会溢出。
- `prefers-reduced-motion` 下禁用或弱化旋转、呼吸和点动画。

## 4. 前端验证清单

在仓库根目录执行：

```powershell
cd web
npm run lint
npm run build
```

可再运行前端测试：

```powershell
npm test
```

UI 完成后必须执行 Impeccable 检测：

```powershell
node C:\Users\wangyufeng\.codex\skills\impeccable\scripts\detect.mjs --json web/src/plugin/smart/view/copilot.vue
```

如启动开发服务，默认使用 Vite；若端口已占用改用其他端口。至少检查桌面和窄屏下：

- 等待状态出现时头像、环形动效、文字和三点不重叠。
- 发送按钮 loading 与等待动画同时出现时布局不跳动。
- 查询完成后动画完全移除，历史消息头像仍正确显示。
- `prefers-reduced-motion` 生效。

## 5. 当前工作区注意事项

`git status` 显示以下已有改动，属于用户/其他任务的墙报工作，不要回退、清理、重新格式化或纳入本任务提交：

```text
M  web/package-lock.json
M  web/package.json
M  web/src/pathInfo.json
M  web/src/view/dashboard/LeadershipWallboard.vue
A  web/src/view/dashboard/wallboard/AnimatedValue.vue
A  web/src/view/dashboard/wallboard/AssetOrbit.vue
A  web/src/view/dashboard/wallboard/WallboardChart.vue
A  web/src/view/dashboard/wallboard/data.js
A  web/src/view/dashboard/wallboard/data.test.js
```

当前分支为 `main`。后端验收基线为 `7f51cf3`；本交接文档已由提交 `a6fab9f` 推送到 `origin/main`，因此当前 `HEAD`/`origin/main` 还包含该文档提交。本交接文档是本轮新增文件；提交时只暂存本文件和前端任务实际修改的文件，切勿使用 `git add .`。

## 6. 生产发布收口注意事项

生产环境事实：

- SSH 别名：`gin-vue-admin-remote`（从 WSL 使用）。
- 服务器：`192.166.20.103`。
- 入口：`http://192.166.20.103:8080/`。
- 生产根目录：`/data/gin-vue-admin`。
- Compose 目录：`/data/gin-vue-admin/deploy/docker-dev`。
- 发布标记：`/data/gin-vue-admin/.deploy/current-commit`。

上一轮记录显示生产 marker 仍是旧版本 `41c322f...`，而已验收的后端版本是 `7f51cf3...`。继续部署前先从服务器只读核对 marker；若仍旧，先把 marker 收口到已验收的后端提交，再根据新提交相对 marker 的变更范围构建受影响服务。前端只改 `web/**` 时通常执行：

```bash
./build.sh web
docker compose --env-file .env -f docker-compose.yml up -d --force-recreate web
./release-acceptance.sh
./ps.sh
```

只有 `release-acceptance.sh` 返回 0 后，才能把完整新提交 hash 写入 `.deploy/current-commit`。发布包必须来自已提交并推送的 Git revision；临时文件放在服务器 `/data/gin-vue-admin/workspace/`，不要在 Windows 仓库或 `%TEMP%` 留发布压缩包。不要打印或提交 `.env`、`config.yaml` 等机密文件。

## 7. 推荐的继续顺序

1. 阅读本文件并运行 `git status -sb`，确认墙报改动仍在。
2. 查看 `copilot.vue` 和 `ai-gva.svg`，确定头像呈现方式。
3. 只修改 `copilot.vue`，完成头像替换和思考动效，保留无障碍属性及 reduced-motion 降级。
4. 运行 lint、build、必要的前端测试和 Impeccable detector。
5. 检查 `git diff --check` 与完整任务 diff；只提交前端文件（以及本交接文档，如需要保留交接记录）。
6. 按生产规则发布前端，执行验收和容器状态检查；最后更新 marker。

新线程可直接使用下面这句话开始：

> 请先阅读 `TASK_HANDOFF.md`，不要回退当前墙报改动；继续完成 `web/src/plugin/smart/view/copilot.vue` 的机器人头像替换和思考动态效果，然后按文档完成验证与发布收口。
