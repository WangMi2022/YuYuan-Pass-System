# 资产管理系统 MVP

## 已实现

- 资产分类：预置座椅/板凳、桌类家具、电脑整机、显示设备、网络设备、办公设备、生产设备、其他资产，可继续增删改。
- 资产档案：编号、名称、分类、品牌、型号、序列号、数量、单位、采购单价、资产原值、当前估值、状态、位置、保管人、供应商、日期、照片和备注。
- 计价统计：资产原值按“数量 × 采购单价”自动计算，大屏汇总原值、当前估值和价值减少额。
- 分类统计：按分类统计档案数、实物数量、资产原值和当前估值。
- 可视化大屏：核心指标、分类价值、资产状态、位置排行、分类概览和最近登记。
- 图片存储：图片写入 RustFS 的 `gva-assets` 桶、`assets/` 前缀，浏览器通过后端代理读取私有对象。
- 权限：菜单与 API 首次自动授予默认管理员角色，其他角色可在角色管理中分配。

## 代码位置

```text
server/plugin/asset/             后端模型、服务、API、路由、迁移和初始化
web/src/plugin/asset/            前端 API、资产档案、资产分类和大屏
deploy/docker-dev/               二开 Dockerfile、Compose 和运维脚本
deploy/docker-dev/configure-rustfs.sh
```

## 开发部署

```bash
cd deploy/docker-dev
cp .env.example .env
./up.sh
./logs.sh server
./logs.sh web
```

RustFS 管理控制台使用 `9001`，S3 兼容 API 使用 `9000`。连接参数保存在
`deploy/docker-dev/.env`，`up.sh` 会自动执行 `configure-rustfs.sh`。

## 页面

- 资产中心 / 资产大屏
- 资产中心 / 资产档案
- 资产中心 / 资产分类
