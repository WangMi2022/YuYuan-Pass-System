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
  "specifications": "32GB RAM / 1TB SSD",
  "productionDate": "2026-06-01T00:00:00+08:00",
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

### 3.4 资产风险中心

所有风险接口均位于 JWT + Casbin 私有路由；写接口额外挂载 `OperationRecord()`。

| 方法 | 路径 | 主要参数 | 说明 |
| --- | --- | --- | --- |
| GET | `/assetRisk/dashboard` | - | 风险指标、分布、30 天趋势、最近事件和最近扫描 |
| GET | `/assetRisk/list` | 分页、`status/severity/category/ruleCode/assetId/assignedTo/keyword` | 风险事件列表 |
| GET | `/assetRisk/detail` | query `id` | 风险事件、证据、关联资产和处理日志 |
| GET | `/assetRisk/rules` | - | 17 条风险规则及当前版本 |
| PUT | `/assetRisk/rules` | 规则更新 JSON | 修改等级、阈值和启用状态，版本自动递增 |
| POST | `/assetRisk/scan` | 可选 `{"runId": 12}` | 启动新扫描，或从失败任务游标续扫；允许空请求体 |
| GET | `/assetRisk/scans` | 分页、`status/triggerType` | 扫描运行记录 |
| PUT | `/assetRisk/acknowledge` | 风险动作 JSON | 确认已接手风险 |
| PUT | `/assetRisk/resolve` | 风险动作 JSON | 标记解决，必须填写处理说明 |
| PUT | `/assetRisk/ignore` | 风险动作 JSON | 忽略风险，必须填写原因且只能单条操作 |
| PUT | `/assetRisk/reopen` | 风险动作 JSON | 重新打开已解决或已忽略风险，必须填写说明 |
| PUT | `/assetRisk/assign` | 分配 JSON | 批量分配处理人，`assignedTo=0` 表示取消分配 |

风险动作请求：

```json
{
  "ids": [21, 22],
  "note": "已核对实物和流转记录"
}
```

确认、解决、重新打开最多可处理 100 条；忽略操作必须逐条提交。分配请求为：

```json
{
  "ids": [21, 22],
  "assignedTo": 8
}
```

规则更新请求：

```json
{
  "ID": 3,
  "severity": "high",
  "parameters": {
    "days": 30,
    "minValue": 10000
  },
  "enabled": true
}
```

不同规则允许的参数不同，未知参数和不安全阈值会被服务端拒绝。风险等级为 `low/medium/high/critical`；事件状态为 `open/acknowledged/resolved/ignored`；扫描状态为 `running/success/failed`，触发方式为 `manual/scheduled`。

总览响应的主要字段为：

- `totalOpen`：待处理和已确认风险总数。
- `highOpen`：其中高风险和严重风险数量。
- `todayNew`：当天首次命中的事件数。
- `overdue`：首次命中已超过 7 天且仍未关闭的事件数。
- `byCategory/bySeverity/byStatus/byCustodian`：分组指标。
- `trend`：最近 30 天新增和解决数量。
- `recentEvents/latestScan/generatedAt`：最近风险、最近扫描和统计时间。

扫描按每批 200 项资产执行，同一进程和数据库中仍有新鲜心跳的扫描会阻止并发启动。失败任务保存游标、心跳、计数和错误，可通过 `runId` 续扫；同一风险使用稳定指纹幂等更新。

### 3.5 智能资产建档

所有接口均位于 JWT + Casbin 私有路由。默认管理员可查看全部任务，普通用户只能访问自己创建的任务；读取权限从资产列表权限继承，写权限从资产新增权限继承。

| 方法 | 路径 | 主要参数 | 说明 |
| --- | --- | --- | --- |
| POST | `/assetRecognition/create` | multipart `files` 或 `file` | 上传 1-6 张图片并创建异步识别任务 |
| GET | `/assetRecognition/list` | `page/pageSize/status` | 识别任务分页列表 |
| GET | `/assetRecognition/detail?id=<id>` | query `id` | 任务、草稿、置信度、警告和重复候选 |
| POST | `/assetRecognition/retry` | `{"id": 12}` | 仅将未确认的失败任务重新排队 |
| PUT | `/assetRecognition/draft` | `id` 和完整 `draft` | 保存人工修正草稿并重新执行确定性校验 |
| POST | `/assetRecognition/confirm` | `{"id": 12}` | 事务创建正式资产并建立不可变关联 |
| DELETE | `/assetRecognition/delete?id=<id>` | query `id` | 删除未确认任务和临时图片 |

上传只接受 JPG、PNG、WebP、GIF，单张最大 10 MB，每批请求最大 61 MB。任务状态：

| 状态 | 含义 |
| --- | --- |
| `pending` | 等待 Worker 领取 |
| `processing` | 正在读取图片和调用 Vision 模型 |
| `reviewing` | 识别完成，等待人工修正和确认 |
| `completed` | 已创建正式资产，任务不可删除或再次确认 |
| `failed` | 识别或 Schema 校验失败，可重新识别 |
| `deleting` | 临时图片尚未完全清理，只允许重试删除 |

草稿保存示例：

```json
{
  "id": 12,
  "draft": {
    "assetCode": "ASSET-2026-0008",
    "name": "研发笔记本",
    "categoryId": 1,
    "brand": "Example",
    "model": "Pro 14",
    "serialNumber": "SN-UNIQUE-008",
    "specifications": "32GB RAM / 1TB SSD",
    "productionDate": "2026-06-01T00:00:00+08:00",
    "quantity": 1,
    "unit": "台",
    "unitPrice": 8999,
    "currentValue": 8999,
    "supplier": "示例供应商",
    "purchaseDate": "2026-08-01T00:00:00+08:00",
    "warrantyEndDate": "2028-08-01T00:00:00+08:00",
    "recommendedWarrantyMonths": 24,
    "photos": [],
    "remarks": "人工核对铭牌完成"
  }
}
```

服务端忽略客户端提交的 `photos`，始终使用任务原始图片。模型不会生成资产编号、采购单价或当前估值。分类必须存在且启用；序列号完全匹配或去除大小写、空格及分隔符后的标准化匹配都会形成重复候选，并阻止确认。

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

### 4.3 发票识别质量

五个接口使用统一查询参数：`startDate/endDate/provider/model/fileType`；失败明细额外支持 `page/pageSize`。默认统计最近 30 天，结束日期包含当天，时间范围不能超过 366 天。数据范围与发票台账一致。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/invoiceQuality/dashboard` | 总识别量、成功/失败率、平均耗时/尝试、回退率、复核量、修正字段、历史缺口和估算费用 |
| GET | `/invoiceQuality/providerMetrics` | 按 Provider、模型和 MIME 分组的成功率、置信度、耗时、尝试与修正数 |
| GET | `/invoiceQuality/fieldMetrics` | 字段复核量、修改量、修改率、准确率和平均置信度 |
| GET | `/invoiceQuality/failures` | 失败发票、Provider、模型、尝试次数和错误摘要分页明细 |
| GET | `/invoiceQuality/classificationMetrics` | 分类建议、接受、推翻、待决数量和比例 |

字段准确率口径为“已采集复核数据且未被人工修改的字段占比”，不是第三方权威验真准确率。`legacyWithoutFieldData` 单独统计系统启用字段级差异采集前的历史发票。税号和校验码修正记录只保存 SHA-256 摘要，接口不返回对应明文。

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

## 10. AI Gateway 与运营接口

所有 AI 运营接口都位于 JWT + Casbin 私有路由。写接口额外挂载 `OperationRecord()`。默认管理员角色 `888` 在插件初始化时获得菜单和 API 权限，其他角色需单独授权。

| 方法 | 路径 | 主要参数 | 说明 |
| --- | --- | --- | --- |
| GET | `/ai/providers` | - | 获取脱敏后的 Gateway 与 Provider 配置 |
| PUT | `/ai/providers` | Provider 配置 JSON | 更新 Gateway、模型、费用、脱敏和图片策略 |
| GET | `/ai/usage/summary` | - | 当前用户今日请求、Token、本月费用和累计请求 |
| GET | `/ai/invocations` | `page/pageSize/status/module/provider/userId` | 分页查询调用审计 |
| GET | `/ai/quotas` | - | 查询所有配额 |
| PUT | `/ai/quotas` | 配额 JSON | 新建或更新配额 |
| GET | `/ai/prompts` | - | 查询 Prompt 全部版本 |
| POST | `/ai/prompts` | Prompt JSON | 创建新的草稿版本 |
| PUT | `/ai/prompts/activate` | `{promptKey,version}` | 激活指定版本并退役旧版本 |

Provider 更新示例：

```json
{
  "enabled": true,
  "allow-private-endpoints": false,
  "sensitive-words": ["内部项目代号"],
  "allow-vision-modules": ["asset-recognition"],
  "openai-compatible": {
    "enabled": true,
    "base-url": "https://api.example.com/v1",
    "api-key": "<write-only-key>",
    "model": "example-model",
    "timeout-seconds": 60,
    "input-cost-micros-per-million": 1000000,
    "output-cost-micros-per-million": 3000000
  },
  "anthropic": {
    "enabled": false,
    "base-url": "https://api.anthropic.com",
    "model": "",
    "timeout-seconds": 60,
    "input-cost-micros-per-million": 0,
    "output-cost-micros-per-million": 0
  }
}
```

读取配置时不返回任何密钥，只返回 `api-key-configured`。更新时省略或留空 `api-key` 会保留原密钥；`clear-api-key: true` 才会清空。内网、回环、链路本地和私有地址默认禁止作为 Provider Endpoint。

配额示例：

```json
{
  "scopeType": "module",
  "scopeId": "asset-recognition",
  "dailyRequests": 500,
  "dailyTokens": 2000000,
  "monthlyCostMicros": 300000000,
  "maxConcurrency": 4,
  "enabled": true
}
```

`scopeType` 支持 `global/module/authority/user`；数值 `0` 表示对应维度不限制。所有匹配且启用的配额同时生效。

Prompt 创建示例：

```json
{
  "promptKey": "asset-recognition",
  "content": "识别资产铭牌并返回结构化字段。",
  "outputSchema": "{\"type\":\"object\",\"required\":[\"name\"]}"
}
```

服务端在入库前编译 JSON Schema。业务调用指定 `promptKey` 且不指定版本时使用 `active` 版本；指定 `version` 时读取对应历史版本。模型输出不符合 Schema 时返回独立 `schema` 错误并写入审计。

审计表不保存完整 Prompt、Payload、图片、模型输出或 API Key，只保存身份、模块、Provider、模型、Token、估算费用、耗时、状态、错误类型、脱敏数和输入输出哈希。

## 11. 系统管理接口

系统用户、角色、菜单、API、字典、日志、参数、版本、文件、自动代码等接口数量较多，运行时 Swagger 是权威清单。生成文件位于：

- `server/docs/swagger.yaml`
- `server/docs/swagger.json`
- `server/docs/docs.go`

当前 Swagger 包含 123 条框架与系统路径，但部分二次开发插件的注释覆盖仍需补齐。因此：

- 系统管理接口以 Swagger 为准。
- 本文列出的业务插件接口以插件 Router 源码为准。
- 联调前同时检查浏览器 Network、前端 `api/*.js` 和后端 Router。

## 12. 调用示例

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

## 13. 错误处理

客户端应同时处理：

1. HTTP 网络错误和超时。
2. HTTP `401`，清理登录状态并返回登录页。
3. HTTP `200` 但业务 `code=7` 的失败响应。
4. 文件接口返回 JSON 错误而不是预期二进制的情况。
5. 长任务的重试、幂等与用户反馈。

## 14. 接口变更规则

1. 不直接破坏已有字段语义；必要时新增字段并提供兼容期。
2. 新增分页列表必须返回 `list/total/page/pageSize`。
3. 新增写接口必须明确是否使用 `OperationRecord()`。
4. 文件接口必须说明认证、Content-Type、大小限制和删除补偿。
5. 修改路由后同步更新 Router、前端 API 封装、权限初始化、Swagger 和本文。
