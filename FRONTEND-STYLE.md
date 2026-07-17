# SaaS 风格前端主题

现代 SaaS 简约风（Linear / Vercel 气质）：浅色为主、大留白、1px 细边框、
克制阴影、精致微交互；暗色模式为近黑低饱和。工作台横幅与资产大屏内置
Three.js 粒子特效，颜色跟随主题色与亮暗模式实时联动。

## 设计语言

- 56px 顶栏、40px 历史标签条、可配置宽度侧栏（默认 220px）
- zinc 中性灰阶：亮色底 `#FAFAFA` / 卡片 `#FFF`；暗色底 `#09090B` / 卡片 `#111113`
- 默认主色 indigo `#6366F1`，用户可在设置面板切换 13 个预设色或自定义
- 语义状态色（success/warning/danger/info）附带 soft 底色令牌
- 六色分类图表色板（`--na-chart-1..6`，chart-1 跟随主色）
- 10px 卡片圆角、极轻阴影、150ms 微交互过渡
- 完整亮/暗主题、键盘焦点、prefers-reduced-motion 与移动端响应式

## 核心文件

```text
web/src/style/theme/tokens.scss              设计令牌唯一真源（:root + html.dark）
web/src/style/theme/element-bridge.scss      --el-* ← --na-* 变量桥接
web/src/style/theme/base.scss                文档基础（字体/滚动条/选区/焦点）
web/src/style/theme/components.scss          Element Plus 组件精修
web/src/style/theme/shell.scss               顶栏、侧栏、标签条、页面外壳类
web/src/style/theme/legacy-bridge.scss       系统页换肤（gva-* 容器 + 归一化层）
web/src/components/three/useThreeScene.js    Three.js 场景生命周期封装
web/src/components/three/HeroCanvas.vue      工作台横幅波浪粒子
web/src/components/three/AmbientCanvas.vue   资产大屏漂移粒子场
web/src/components/charts/theme.js           ECharts 从 CSS 变量取色
web/src/view/layout/                         布局外壳与设置面板
web/src/view/dashboard/index.vue             首页工作台
web/src/plugin/asset/view/                   资产档案、分类、流转与大屏
```

## 主题运行机制

- 运行时主色注入：`setBodyPrimaryColor`（`web/src/utils/format.js`）向
  `documentElement` 写入 `--na-primary` 系、`--na-chart-1` 与全套
  `--el-color-primary*`；令牌文件中的主色仅作首帧兜底。
- 暗色切换：`html.dark` class（`useDark`），令牌在 `tokens.scss` 的
  `html.dark` 块中整体翻转。
- 用户 12 项主题配置（主色/暗色/布局四模式/侧栏三尺寸/标签页/灰色/色弱/
  水印/切换动画/全局尺寸）存于 `pinia/modules/app.js`，经 localStorage
  `originSetting` 与后端 `sys_user.origin_setting` 双持久化。
- ECharts 等 canvas 绘制场景通过 `chartTheme()` 在构建 option 时解析
  CSS 变量实际值，依赖 `isDark`/`primaryColor` 触发重算。

## 全局覆盖范围

- Button、Input、Select、DatePicker、Textarea、Switch、Checkbox/Radio
- Table、Pagination、Tag、Alert、Empty、Skeleton、Loading、Tree、Steps
- Card、Dialog、Drawer、Dropdown、Popper、Message、Notification
- `gva-search-box`、`gva-table-box`、`gva-form-box`、`gva-btn-list`
  与遗留 Tailwind 类归一化（系统管理页零改码换肤）
- 四种导航布局、亮/暗模式与移动端断点

## 调整主题

- 设计令牌：`web/src/style/theme/tokens.scss` 的 `:root` 与 `html.dark`
- 默认主色与预设色板：`web/src/pinia/modules/app.js`、
  `web/src/view/layout/setting/components/themeColorPicker.vue`
- 页面外壳复用类：`.na-page / .na-page-header / .na-panel / .na-toolbar`
  （见 `shell.scss`），业务页优先复用而非自写容器

```bash
cd deploy/docker-dev
./up.sh
```

构建后的默认访问地址：`http://<服务器IP>:8080`。
