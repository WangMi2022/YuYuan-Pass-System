# YuYuan Pass System API 接口文档

## 1. 基本约定

### 1.1 地址

| 场景 | Base URL |
| --- | --- |
| 浏览器经 Web/Nginx 访问 | `/api` |
| 直接访问 Server | `http://<host>:8888` |
| Swagger UI | `http://<host>:8888/swagger/index.html` |

当前 Nginx 会将 `/api/<path>` 重写为 `/<path>` 后转发给 Server。本文路径均使用后端真实路径，不含 `/api`；浏览器直连 Web 域名时需在前面加 `/api`。

### 1.2 Content-Type

- 普通请求：`application/json`。
- 文件上传：`multipart/form-data`。
- 文件下载：二进制响应，前端使用 `arraybuffer` 或浏览器流。

### 1.3 鉴权头

除公开接口外，请求需要：

```http
x-token: <jwt-token>
x-user-id: <current-user-id>
```

服务端以 JWT 中的身份为准，不能依赖客户端 `x-user-id` 作为授权依据。

### 1.4 统一响应

```json
{
  "code": 0,
  "data": {},
  "msg": "成功"
}
```

| 字段 | 说明 |
| --- | --- |
| `code` | `0` 成功，`7` 业务失败 |
| `data` | 对象、数组、分页对象或空对象 |
| `msg` | 面向用户的结果信息 |

分页响应：

```json
{
  "code": 0,
  "data": {
    "list": [],
    "total": 0,
    "page": 1,
    "pageSize": 10
  },
  "msg": "获取成功"
}
```

JWT 无效或过期通常返回 HTTP `401`。部分文件接口直接返回文件流，不使用统一 JSON 包装。

### 1.5 分页参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `page` | integer | 页码，从 1 开始 |
| `pageSize` | integer | 每页数量 |
| `keyword` | string | 通用关键词，具体支持范围由模块决定 |

## 2. 公共与认证接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/health` | 否 | 服务存活检查，返回 JSON 字符串 `"ok"` |
| POST | `/base/captcha` | 否 | 获取登录验证码 |
| POST | `/base/login` | 否 | 用户登录，返回用户、token 和过期时间 |
| GET | `/appearance/login-logo` | 否 | 当前登录图标 |
| GET | `/appearance/login-background` | 否 | 当前登录背景 |
| GET | `/info/getInfoDataSource` | 否 | 公告数据源 |
| GET | `/info/getInfoPublic` | 否 | 公开公告列表 |

登录请求示例：

```json
{
  "username": "admin",
  "password": "<password>",
  "captcha": "123456",
  "captchaId": "<captcha-id>"
}
```

验证码是否强制取决于服务端防爆配置和登录失败次数。

## 3. 资产接口

### 3.1 资产档案

| 方法 | 路径 | 主要参数 | 说明 |
| --- | --- | --- | --- |
| POST | `/asset/create` | Asset JSON | 新建资产 |
| PUT | `/asset/update` | Asset JSON | 更新资产 |
| DELETE | `/asset/delete` | query `id` | 删除资产 |
| GET | `/asset/detail` | query `id` | 资产详情 |
| GET | `/asset/list` | 分页、`categoryId/status/location/minValue/maxValue` | 资产列表 |
| GET | `/asset/dashboard` | - | 资产统计大屏数据 |
| GET | `/asset/categoryOptions` | - | 分类选项 |
| POST | `/asset/uploadPhoto` | multipart `file` | 上传资产图片 |
| DELETE | `/asset/deletePhoto` | query `key` | 删除未使用图片 |
| GET | `/asset/photo` | query `key` | 后端代理读取资产图片 |

资产保存核心字段示例：

```json
{
  "assetCode": "ASSET-2026-0001",
  "name": "研发笔记本",
  "categoryId": 1,
  "brand": "Example",
  "model": "Pro 14",
  "serialNumber": "SN0001",
  "quantity": 1,
  "unit": "台",
  "unitPrice": 8999,
  "currentValue": 7600,
  "location": "IT 仓库",
  "custodian": "张三",
  "remarks": "2026 年采购"
}
```

`originalValue` 由服务端根据数量和单价计算。新建档案状态为 `pending_inbound`。

### 3.2 分类与位置

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/assetCategory/create` | 新建分类 |
| PUT | `/assetCategory/update` | 更新分类 |
| DELETE | `/assetCategory/delete?id=<id>` | 删除分类 |
| GET | `/assetCategory/list` | 分类分页列表 |
| POST | `/assetLocation/create` | 新建位置 |
| PUT | `/assetLocation/update` | 更新位置 |
| DELETE | `/assetLocation/delete?id=<id>` | 删除位置 |
| GET | `/assetLocation/list` | 位置分页列表 |
| GET | `/assetLocation/options?type=<type>` | 按业务位置类型获取启用位置 |

位置 JSON 示例：

```json
{
  "name": "研发一部",
  "type": "usage",
  "code": "USE-RD-01",
  "description": "研发团队使用位置",
  "sort": 10,
  "enabled": true
}
```

位置类型：`inbound`、`usage`、`transfer`、`return`、`maintenance`、`disposal`。

### 3.3 资产流转

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/assetOperation/list` | 按 `type/status/startDate/endDate` 分页查询 |
| GET | `/assetOperation/detail?id=<id>` | 获取单据和明细 |
| GET | `/assetOperation/assetOptions?type=<type>&keyword=<text>` | 搜索可流转资产 |
| POST | `/assetOperation/create` | 创建草稿或创建并提交 |
| PUT | `/assetOperation/update` | 更新草稿，ID 在 JSON 的 `ID` 字段 |
| PUT | `/assetOperation/submit?id=<id>` | 提交既有草稿 |
| DELETE | `/assetOperation/delete?id=<id>` | 删除草稿 |

创建/更新请求：

```json
{
  "ID": 0,
  "type": "issue",
  "businessDate": "2026-08-13T00:00:00+08:00",
  "targetLocation": "研发一部",
  "targetCustodian": "张三",
  "reason": "项目研发使用",
  "remarks": "",
  "assetIds": [1, 2],
  "submit": false
}
```

业务类型：`inbound`、`issue`、`transfer`、`return`、`maintenance`、`scrap`。单张业务单最多选择 100 项资产。

## 4. 发票接口

### 4.1 发票主流程

| 方法 | 路径 | 主要参数 | 说明 |
| --- | --- | --- | --- |
| POST | `/invoice/upload` | multipart `file` 或多个 `files` | 上传并创建识别任务 |
| GET | `/invoice/list` | `page/pageSize/status/categoryId/direction/startDate/endDate` | 发票台账 |
| GET | `/invoice/detail?id=<id>` | query `id` | 发票、明细和识别信息 |
| GET | `/invoice/capabilities` | - | 当前识别、验真提供方能力 |
| GET | `/invoice/file?id=<id>` | query `id` | 下载原始证据文件 |
| PUT | `/invoice/update` | 发票复核 JSON | 修改待审核数据 |
| PUT | `/invoice/confirm` | `{id,verificationBypass,verificationBypassReason}` | 确认并进入正式统计 |
| PUT | `/invoice/reopen?id=<id>` | query `id` | 管理员重开已确认发票 |
| PUT | `/invoice/retry?id=<id>` | query `id` | 重试识别 |
| POST | `/invoice/recheck?id=<id>` | query `id` | 重新检查识别结果 |
| POST | `/invoice/verify?id=<id>` | query `id` | 发起验真 |
| GET | `/invoice/verificationHistory?id=<id>` | query `id` | 验真历史 |
| POST | `/invoice/provider/test` | provider 配置 JSON | 测试提供方连通性 |
| DELETE | `/invoice/delete?id=<id>` | query `id` | 删除发票并创建文件清理任务 |
| GET | `/invoice/dashboard` | - | 发票与流水统计 |

金额字段均以“分”为单位：

```json
{
  "ID": 12,
  "direction": "expense",
  "invoiceType": "vat_electronic",
  "invoiceCode": "",
  "invoiceNumber": "12345678",
  "issueDate": "2026-08-01T00:00:00+08:00",
  "buyerName": "本公司",
  "sellerName": "供应商公司",
  "amountCents": 100000,
  "taxCents": 13000,
  "totalCents": 113000,
  "categoryId": 2,
  "reviewNotes": "人工复核完成",
  "items": []
}
```

### 4.2 发票分类与规则

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/invoice/categoryOptions` | 启用分类选项 |
| POST | `/invoiceCategory/create` | 新建分类 |
| PUT | `/invoiceCategory/update` | 更新分类 |
| DELETE | `/invoiceCategory/delete?id=<id>` | 删除分类 |
| GET | `/invoiceCategory/list` | 分类列表 |
| POST | `/invoiceRule/create` | 新建分类规则 |
| PUT | `/invoiceRule/update` | 更新分类规则 |
| DELETE | `/invoiceRule/delete?id=<id>` | 删除分类规则 |
| GET | `/invoiceRule/list` | 规则列表 |

## 5. 文档接口

| 方法 | 路径 | 参数 | 说明 |
| --- | --- | --- | --- |
| POST | `/document/upload` | multipart `file` | 上传文档 |
| GET | `/document/list` | `page/pageSize/fileExt` | 文档列表 |
| GET | `/document/detail?id=<id>` | query `id` | 文档详情与在线内容 |
| GET | `/document/file?id=<id>` | query `id` | 原始文件流 |
| PUT | `/document/updateContent` | `ID/title/content/remarks` | 保存在线编辑内容 |
| DELETE | `/document/delete?id=<id>` | query `id` | 删除文档及存储对象 |

在线内容请求示例：

```json
{
  "ID": 8,
  "title": "资产采购规范",
  "content": "# 正文",
  "remarks": "已完成第二次修订"
}
```

## 6. 公告接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | `/info/createInfo` | API 权限 | 新建公告 |
| PUT | `/info/updateInfo` | API 权限 | 更新公告 |
| DELETE | `/info/deleteInfo` | API 权限 | 删除单条公告 |
| DELETE | `/info/deleteInfoByIds` | API 权限 | 批量删除 |
| GET | `/info/findInfo` | API 权限 | 公告详情 |
| GET | `/info/getInfoList` | API 权限 | 管理列表 |
| GET | `/info/notifications` | 登录 | 当前用户通知列表和未读数 |
| POST | `/info/read` | 登录 | 单条已读 |
| POST | `/info/readAll` | 登录 | 全部已读 |
| GET | `/info/stream` | 登录 | SSE 实时通知流 |

SSE 连接应保持代理正确支持流式传输；断线后客户端仍应通过通知列表补偿未读状态。

## 7. 个人日程接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/workSchedule/list` | 当前用户日程 |
| POST | `/workSchedule/create` | 新建日程 |
| PUT | `/workSchedule/update` | 更新日程 |
| DELETE | `/workSchedule/delete?id=<id>` | 删除本人日程 |
| POST | `/workSchedule/import` | 导入旧客户端日程 |
| GET | `/workSchedule/notifications` | 提醒收件箱 |
| POST | `/workSchedule/notifications/read` | 单条提醒已读，请求 `{id}` |
| POST | `/workSchedule/notifications/readAll` | 全部已读 |

请求示例：

```json
{
  "id": 0,
  "clientKey": "browser-migration-key",
  "title": "资产月度盘点",
  "date": "2026-08-20",
  "time": "09:30",
  "type": "asset",
  "note": "核对研发区设备",
  "recurrence": {
    "enabled": true,
    "mode": "monthly",
    "weekdays": [],
    "monthDays": [20],
    "weekday": 0,
    "monthDay": 20
  }
}
```

## 8. 站点接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/site/create` | 新建站点 |
| PUT | `/site/update` | 更新站点 |
| DELETE | `/site/delete?id=<id>` | 删除站点 |
| GET | `/site/list` | 分页、分类、启用状态筛选 |
| GET | `/site/detail?id=<id>` | 站点详情 |
| GET | `/site/categories` | 分类选项 |
| POST | `/site/visit` | 记录访问次数和最近访问时间 |

## 9. 登录外观接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/appearance/login-logo` | 否 | 当前图标 |
| PUT | `/appearance/login-logo` | API 权限 | 保存图标 `{name,url}` |
| DELETE | `/appearance/login-logo` | API 权限 | 恢复默认图标 |
| GET | `/appearance/login-background` | 否 | 当前背景 |
| GET | `/appearance/login-backgrounds` | API 权限 | 背景图库 |
| POST | `/appearance/login-background` | API 权限 | 新增 `{name,url}` |
| PUT | `/appearance/login-background/activate` | API 权限 | 激活 `{id}` |
| DELETE | `/appearance/login-background?id=<id>` | API 权限 | 删除非当前背景 |

## 10. 系统管理接口

系统用户、角色、菜单、API、字典、日志、参数、版本、文件、自动代码等接口数量较多，运行时 Swagger 是权威清单。生成文件位于：

- `server/docs/swagger.yaml`
- `server/docs/swagger.json`
- `server/docs/docs.go`

当前 Swagger 包含 123 条框架与系统路径，但部分二次开发插件的注释覆盖仍需补齐。因此：

- 系统管理接口以 Swagger 为准。
- 本文列出的业务插件接口以插件 Router 源码为准。
- 联调前同时检查浏览器 Network、前端 `api/*.js` 和后端 Router。

## 11. 调用示例

登录并查询资产：

```bash
curl -X POST 'http://localhost:8888/base/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"<password>","captcha":"","captchaId":""}'

curl 'http://localhost:8888/asset/list?page=1&pageSize=20' \
  -H 'x-token: <token>' \
  -H 'x-user-id: <user-id>'
```

经 Web 反向代理调用时，将地址改为 `http://localhost:8080/api/asset/list?...`。

## 12. 错误处理

客户端应同时处理：

1. HTTP 网络错误和超时。
2. HTTP `401`，清理登录状态并返回登录页。
3. HTTP `200` 但业务 `code=7` 的失败响应。
4. 文件接口返回 JSON 错误而不是预期二进制的情况。
5. 长任务的重试、幂等与用户反馈。

## 13. 接口变更规则

1. 不直接破坏已有字段语义；必要时新增字段并提供兼容期。
2. 新增分页列表必须返回 `list/total/page/pageSize`。
3. 新增写接口必须明确是否使用 `OperationRecord()`。
4. 文件接口必须说明认证、Content-Type、大小限制和删除补偿。
5. 修改路由后同步更新 Router、前端 API 封装、权限初始化、Swagger 和本文。
