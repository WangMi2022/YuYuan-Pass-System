# gin-vue-admin Docker 二开启动脚本

项目内目录：`deploy/docker-dev`

## 文件说明

- `server.Dockerfile`：自定义后端镜像构建文件
- `web.Dockerfile`：自定义前端镜像构建文件
- `docker-compose.yml`：只启动 web/server，数据库和 Redis 使用外部服务
- `.env`：当前环境变量与数据库、Redis、对象存储配置，不提交到 Git
- `config.init.yaml`：首次初始化模板
- `config.yaml`：容器实际挂载配置；初始化成功后后端会写入数据库配置
- `build.sh`：构建镜像
- `up.sh`：构建并启动，随后自动调用初始化接口
- `init-db.sh`：手动初始化数据库
- `down.sh`：停止并删除容器
- `restart.sh`：重启容器
- `logs.sh`：查看日志
- `ps.sh`：查看状态
- `reset-config.sh`：仅重置挂载配置，不删除数据库
- `configure-rustfs.sh`：从 `.env` 写入 Redis 与 RustFS/MinIO S3 配置

## 常用命令

```bash
cd /data/gin-vue-admin/deploy/docker-dev
./build.sh
./up.sh
./logs.sh server
./logs.sh web
./down.sh
```

## RustFS 图片存储

RustFS 的 `9001` 是管理控制台端口，Gin 后端通过同主机 `9000` 的 S3 兼容接口上传图片。
连接信息保存在权限为 `600` 的 `.env` 中，执行 `./up.sh` 时会自动调用
`configure-rustfs.sh` 更新 Redis 和对象存储运行配置。默认桶为 `gva-assets`，对象前缀为 `assets/`。

## 默认访问地址

- 前端：`http://<服务器IP>:8080`
- 后端：`http://<服务器IP>:8888`
- Swagger：`http://<服务器IP>:8888/swagger/index.html`

默认初始化管理员密码见 `.env` 的 `GVA_ADMIN_PASSWORD`。

## 伪造资产演示数据

需要快速填充资产大屏和资产档案列表时，可在部署服务器执行：

```bash
cd /data/gin-vue-admin
./deploy/docker-dev/tools/seed-assets.sh --count 100
```

脚本默认会先清理同前缀 `DEMO-ASSET-*` 的旧演示资产，再生成 100 条新资产；会覆盖当前启用的资产分类，并包含座椅、桌类、电脑、显示、网络、办公、生产和其他资产等模板。可通过参数调整：

```bash
./deploy/docker-dev/tools/seed-assets.sh --count 200 --prefix DEMO-ASSET --reset=true
```
