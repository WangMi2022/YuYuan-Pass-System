# New API 风格前端主题

本主题参考 `QuantumNous/new-api` 当前默认前端的设计语言，并针对
Vue 3、Element Plus 和资产管理业务重新实现。

## 设计语言

- 48px 全局顶栏、240px 紧凑侧栏和轻量标签导航
- 浅色背景、白色表面、低对比边框、低阴影和 8–16px 圆角体系
- 蓝色主色、语义化成功/警告/危险色与数据图表色
- 系统无衬线字体栈、紧凑数据表、清晰的标题层级
- 完整亮色/深色主题、键盘焦点、减少动画和移动端响应式规则

## 核心文件

```text
web/src/style/new-api-theme.scss            全局令牌和 Element Plus 覆盖
web/src/view/layout/                         顶栏、侧栏、标签页、内容外壳
web/src/view/login/index.vue                 登录页
web/src/view/dashboard/index.vue             工作台
web/src/plugin/asset/view/                   资产档案、分类与大屏
web/public/asset-logo.svg                    新品牌图标
```

## 全局覆盖范围

- Button、Input、Select、DatePicker、Textarea、Switch
- Table、Pagination、Tag、Alert、Empty、Skeleton、Loading
- Card、Dialog、Drawer、Dropdown、Popover、Message
- `gva-search-box`、`gva-table-box`、`gva-form-box`
- 四种导航布局、亮色/深色模式以及 390px 移动端

## 调整主题

基础语义变量位于 `new-api-theme.scss` 的 `:root` 和 `html.dark`。
默认主色与布局密度位于 `web/src/pinia/modules/app.js`。

```bash
cd deploy/docker-dev
./up.sh
```

构建后的默认访问地址：`http://<服务器IP>:8080`。
