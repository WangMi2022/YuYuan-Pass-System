# YuYuan Pass System 文档中心

本目录是项目文档的统一入口。文档以当前仓库代码、部署脚本和运行配置为依据，最近一次整体审计日期为 **2026-08-13**。

## 文档地图

| 文档 | 适用对象 | 主要内容 |
| --- | --- | --- |
| [项目总览](../README.md) | 所有人 | 产品定位、核心能力、技术栈、快速启动、项目结构 |
| [项目审计报告](PROJECT-AUDIT.md) | 负责人、技术负责人 | 当前实现、技术债、风险、改进优先级 |
| [智能资产运营中心开发实施文档](SMART-ASSET-OPERATIONS-DEVELOPMENT-PLAN.md) | 产品、研发、测试、项目负责人 | 智能建档、风险中心、业务助手、智能日报的里程碑、任务、接口、数据与验收 |
| [产品说明书](PRODUCT-MANUAL.md) | 产品、管理、交付 | 用户角色、产品目标、功能范围、业务流程、验收口径 |
| [功能规格说明](FUNCTIONAL-SPECIFICATION.md) | 产品、研发、测试 | 模块能力、业务规则、状态机、权限和验收标准 |
| [用户使用手册](USER-GUIDE.md) | 系统用户、管理员 | 登录、资产、发票、文档、日程、公告、权限等操作方法 |
| [API 接口文档](API.md) | 前后端、集成方、测试 | 鉴权、统一响应、业务接口清单、调用示例、错误处理 |
| [系统架构说明](ARCHITECTURE.md) | 架构师、研发、运维 | 分层结构、插件机制、鉴权链路、部署拓扑、数据流 |
| [数据字典](DATA-DICTIONARY.md) | 研发、数据、运维 | 核心表、字段口径、状态枚举、关联关系、存储约定 |
| [开发维护指南](DEVELOPMENT.md) | 研发、维护人员 | 本地环境、目录职责、开发流程、测试、Swagger 与发布 |
| [部署运维手册](DEPLOYMENT.md) | 运维、交付 | 环境准备、Compose 部署、升级、备份、回滚、故障处理 |
| [Docker 部署说明](../deploy/docker-dev/README.md) | 运维、开发 | 脚本速查和上线验收门禁 |

## 阅读路径

### 新用户

1. 阅读 [项目总览](../README.md)。
2. 按 [用户使用手册](USER-GUIDE.md) 熟悉业务页面。
3. 遇到权限、图片、文档或通知问题时查看使用手册的常见问题。

### 产品与项目管理

1. 阅读 [产品说明书](PRODUCT-MANUAL.md) 确认目标和产品边界。
2. 阅读 [功能规格说明](FUNCTIONAL-SPECIFICATION.md) 核对业务规则和验收条件。
3. 阅读 [项目审计报告](PROJECT-AUDIT.md) 安排后续优化优先级。
4. 按 [智能资产运营中心开发实施文档](SMART-ASSET-OPERATIONS-DEVELOPMENT-PLAN.md) 拆解智能化里程碑和开发看板。

### 开发与集成

1. 阅读 [系统架构说明](ARCHITECTURE.md)。
2. 按 [开发维护指南](DEVELOPMENT.md) 启动开发环境。
3. 通过 [API 接口文档](API.md) 和运行时 Swagger 联调。
4. 涉及数据迁移时查阅 [数据字典](DATA-DICTIONARY.md)。

### 运维与交付

1. 阅读 [部署运维手册](DEPLOYMENT.md)。
2. 使用 `deploy/docker-dev/` 内脚本部署和验收。
3. 每次发布后执行 `./release-acceptance.sh`，通过后再标记上线成功。

## 文档维护规则

- 路由、请求字段或响应字段变化时，同步更新 `API.md` 和 Swagger 注释。
- 数据表、状态枚举或业务口径变化时，同步更新 `DATA-DICTIONARY.md` 和 `FUNCTIONAL-SPECIFICATION.md`。
- 新增用户可见功能时，同步更新 `PRODUCT-MANUAL.md`、`USER-GUIDE.md` 和根 `README.md`。
- 部署变量、端口、依赖服务或发布脚本变化时，同步更新 `DEPLOYMENT.md`。
- 文档中的版本、日期和验证结论必须来自实际 Git 提交或本次验证结果，不写未经验证的“已通过”。
