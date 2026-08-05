# Docker Compose 部署脚本

本目录用于单机 Docker Compose 部署。Compose 只运行 Web 和 Server 两个容器，PostgreSQL、Redis、MinIO/RustFS 使用外部服务。

完整的环境准备、首次部署、HTTPS、备份、升级、回滚和故障处理说明见：[部署运维手册](../../docs/DEPLOYMENT.md)。

## 首次启动

```bash
cd deploy/docker-dev
cp .env.example .env
chmod 600 .env
# 编辑 .env，替换所有 change-me 值

chmod +x ./*.sh tools/*.sh
./up.sh
```

启动完成后执行：

```bash
./release-acceptance.sh
./ps.sh
```

默认访问地址：

- Web：`http://<服务器IP>:8080`
- API：`http://<服务器IP>:8888`
- Swagger：`http://<服务器IP>:8888/swagger/index.html`

初始管理员用户名为 `admin`，密码由 `.env` 中的 `GVA_ADMIN_PASSWORD` 决定。

## 脚本说明

| 脚本 | 用途 |
| --- | --- |
| `up.sh` | 校验配置、构建并启动服务、初始化数据库并执行完整上线验收 |
| `build.sh [web|server]` | 构建全部或指定服务镜像 |
| `init-db.sh` | 手动执行首次数据库初始化，重复执行会自动跳过 |
| `health-check.sh` | 日常快速检查 Web、API 和数据库初始化状态 |
| `release-acceptance.sh` | 代码更新后的上线门禁，全部检查通过才可标记上线成功 |
| `configure-rustfs.sh` | 将 `.env` 中的 Redis 和 S3 配置写入运行配置 |
| `restart.sh [web|server]` | 仅重启容器，不重新构建镜像 |
| `logs.sh [web|server]` | 持续查看最近 200 行日志 |
| `ps.sh` | 查看 Compose 服务状态 |
| `down.sh` | 停止并删除应用容器与网络，不删除外部数据 |
| `reset-config.sh` | 备份并重建 `config.yaml`，不清理数据库 |
| `tools/seed-assets.sh` | 生成资产演示数据 |

## 常用命令

```bash
# 代码更新后的完整上线验收
./release-acceptance.sh

# 查看状态和快速健康检查
./ps.sh
./health-check.sh

# 查看日志
./logs.sh server
./logs.sh web

# 只更新前端
./build.sh web
docker compose --env-file .env -f docker-compose.yml up -d --force-recreate web

# 只更新后端
./build.sh server
docker compose --env-file .env -f docker-compose.yml up -d --force-recreate server

# 同时包含前后端和数据模型变更的版本（例如登录图标配置）
./build.sh server web
docker compose --env-file .env -f docker-compose.yml up -d --force-recreate server web

# 停止服务
./down.sh
```

## 上线验收门禁

每次重新构建并创建容器后执行：

```bash
./release-acceptance.sh
```

脚本只执行无副作用的只读检查，不登录、不写业务数据，也不会调用 OCR、发票验真或大模型等付费接口。它会验证：

- `server`、`web` 均已声明并处于运行状态，且新容器没有异常重启。
- Server `/health` 与数据库初始化状态正常。
- Web 首页是当前资产管理系统，而不是 Nginx 错误页或兜底页。
- 浏览器实际使用的 `/api/health` 反代链路正常。
- 首页引用的当前版本 JavaScript 入口可访问、非空且 MIME 类型正确。

退出码 `0` 表示全部通过，可以标记上线成功；`1` 表示验收失败；`2` 表示参数、文件或依赖命令配置错误。失败时禁止更新部署版本标记。
后端首次就绪默认最多等待约 90 秒，可通过 `STARTUP_MAX_ATTEMPTS` 和 `RETRY_INTERVAL_SECONDS` 调整；其他检查使用 `MAX_ATTEMPTS`，避免持续故障时长时间阻塞。

脚本自身的回归测试为：

```bash
bash tests/release-acceptance-test.sh
```

## 重要文件

- `.env`：部署环境变量和外部服务凭据，权限应为 `600`，禁止提交。
- `config.init.yaml`：无敏感信息的首次初始化模板，可提交。
- `config.yaml`：后端实际运行配置，包含连接信息，禁止提交。
- `server.Dockerfile`、`web.Dockerfile`：二次开发使用的多阶段镜像构建文件。
- `nginx.conf`：SPA、静态资源缓存、API 代理和旧资源 404 策略。

后端容器启动时会执行 GORM 自动迁移。升级到包含登录图标配置的版本时会自动创建 `system_login_logos` 表，更新前应先备份 PostgreSQL。

## 演示资产

```bash
./tools/seed-assets.sh --count 100
```

脚本会清理相同前缀的旧演示资产并重新生成家具、电脑、显示、网络、办公和生产设备等数据。不要在正式数据环境中使用与业务资产相同的前缀。
