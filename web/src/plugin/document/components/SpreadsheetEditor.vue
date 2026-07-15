<template>
  <div class="spreadsheet-editor-shell">
    <div ref="containerRef" class="spreadsheet-editor-host" />
  </div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import 'x-data-spreadsheet/dist/xspreadsheet'
import 'x-data-spreadsheet/dist/locale/zh-cn'
import 'x-data-spreadsheet/dist/xspreadsheet.css'

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => []
  },
  readonly: {
    type: Boolean,
    default: false
  }
})

const emits = defineEmits(['update:modelValue', 'ready'])

const containerRef = ref(null)
let spreadsheet = null
let applying = false
let lastPayload = ''
let hostHeight = 620
let resizeObserver = null
let suppressPropSyncUntil = 0
let suppressPropSyncTimer = null

const clone = (value) => JSON.parse(JSON.stringify(value || []))

const blankSheet = () => ({
  name: 'Sheet1',
  rows: { len: 100 },
  cols: { len: 26 },
  merges: [],
  styles: []
})

const normalizeData = (value) => {
  const data = Array.isArray(value) && value.length ? value : [blankSheet()]
  return data.map((sheet, index) => ({
    name: sheet?.name || `Sheet${index + 1}`,
    rows: sheet?.rows || { len: 100 },
    cols: sheet?.cols || { len: 26 },
    merges: Array.isArray(sheet?.merges) ? sheet.merges : [],
    styles: Array.isArray(sheet?.styles) ? sheet.styles : [],
    validations: Array.isArray(sheet?.validations) ? sheet.validations : [],
    freeze: sheet?.freeze || 'A1'
  }))
}

const canonicalData = (value) => normalizeData(value).map((sheet, index) => {
  const rows = sheet.rows && typeof sheet.rows === 'object' ? { ...sheet.rows } : { len: 100 }
  const cols = sheet.cols && typeof sheet.cols === 'object' ? { ...sheet.cols } : { len: 26 }
  rows.len = Math.max(Number(rows.len || 0), 100)
  cols.len = Math.max(Number(cols.len || 0), 26)
  return {
    name: sheet.name || `Sheet${index + 1}`,
    rows,
    cols,
    merges: Array.isArray(sheet.merges) ? sheet.merges : [],
    styles: Array.isArray(sheet.styles) ? sheet.styles : [],
    validations: Array.isArray(sheet.validations) ? sheet.validations : [],
    freeze: sheet.freeze || 'A1'
  }
})

const payloadOf = (value) => JSON.stringify(canonicalData(value))

const suppressNextSelfPropSync = () => {
  suppressPropSyncUntil = Date.now() + 350
  if (suppressPropSyncTimer) clearTimeout(suppressPropSyncTimer)
  suppressPropSyncTimer = setTimeout(() => {
    suppressPropSyncUntil = 0
    suppressPropSyncTimer = null
  }, 360)
}

const calculateSpreadsheetHeight = () => {
  const top = containerRef.value?.getBoundingClientRect?.().top || 0
  const viewport = window.innerHeight || document.documentElement.clientHeight || 760
  // 让表格区域和当前可视区保持自然比例：不再强行拉到很高，
  // 同时为底部“备注”等表单项预留空间，避免出现一截生硬的空白延伸。
  return Math.round(Math.min(660, Math.max(520, viewport - top - 128)))
}

const spreadsheetHeight = () => {
  // x-data-spreadsheet 的 view.height 必须读取宿主真实高度。
  // 如果这里再次按 viewport 计算，页面滚动或 Tab 重排后会和外层 CSS 高度不一致，
  // 底部工具条/网格就会出现“不自然延伸”的视觉错位。
  return Math.round(containerRef.value?.clientHeight || hostHeight)
}

const syncHostHeight = () => {
  if (!containerRef.value) return
  hostHeight = calculateSpreadsheetHeight()
  containerRef.value.style.setProperty('--spreadsheet-editor-height', `${hostHeight}px`)
}

const loadSpreadsheetData = (value) => {
  if (!spreadsheet) return
  const data = canonicalData(value)
  const payload = payloadOf(data)
  if (payload === lastPayload) return
  // 用户在单元格中每输入一个字符都会向父组件同步一次数据。
  // 父组件回写 props 时，如果这里立刻 loadData，会重建表格并把选区重置到 A1。
  // 因此对“本组件刚刚 emit 出去”的回写只更新快照，不重新加载表格实例。
  if (Date.now() < suppressPropSyncUntil) {
    lastPayload = payload
    return
  }
  applying = true
  lastPayload = payload
  spreadsheet.loadData(clone(data))
  requestAnimationFrame(() => spreadsheet?.reRender?.())
  applying = false
}

const emitChange = () => {
  if (!spreadsheet || applying) return
  const data = canonicalData(spreadsheet.getData ? spreadsheet.getData() : props.modelValue)
  lastPayload = payloadOf(data)
  suppressNextSelfPropSync()
  emits('update:modelValue', clone(data))
}

const resizeSpreadsheet = () => {
  syncHostHeight()
  requestAnimationFrame(() => spreadsheet?.reRender?.())
}

const initSpreadsheet = async () => {
  await nextTick()
  if (!containerRef.value || spreadsheet) return

  const Spreadsheet = window.x_spreadsheet
  if (!Spreadsheet) return
  syncHostHeight()
  Spreadsheet?.locale?.('zh-cn')
  spreadsheet = Spreadsheet(containerRef.value, {
    mode: props.readonly ? 'read' : 'edit',
    showToolbar: !props.readonly,
    showContextmenu: !props.readonly,
    showGrid: true,
    view: {
      height: spreadsheetHeight,
      width: () => containerRef.value?.clientWidth || 960
    },
    row: {
      len: 100,
      height: 30
    },
    col: {
      len: 26,
      width: 140,
      indexWidth: 52,
      minWidth: 60
    },
    style: {
      bgcolor: '#ffffff',
      align: 'left',
      valign: 'middle',
      textwrap: true,
      strike: false,
      underline: false,
      color: '#111827',
      font: {
        name: 'Microsoft YaHei',
        size: 10,
        bold: false,
        italic: false
      }
    }
  })

  spreadsheet.change(emitChange)
  spreadsheet.on?.('cell-edited', emitChange)
  loadSpreadsheetData(props.modelValue)
  window.addEventListener('resize', resizeSpreadsheet)
  if (window.ResizeObserver) {
    resizeObserver = new ResizeObserver(resizeSpreadsheet)
    resizeObserver.observe(containerRef.value)
  }
  emits('ready', spreadsheet)
}

watch(
  () => props.modelValue,
  (value) => loadSpreadsheetData(value),
  { deep: false }
)

onMounted(initSpreadsheet)

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeSpreadsheet)
  resizeObserver?.disconnect?.()
  resizeObserver = null
  if (suppressPropSyncTimer) clearTimeout(suppressPropSyncTimer)
  suppressPropSyncTimer = null
  suppressPropSyncUntil = 0
  if (containerRef.value) {
    containerRef.value.innerHTML = ''
  }
  spreadsheet = null
})
</script>

<style scoped lang="scss">
.spreadsheet-editor-shell {
  width: 100%;
  overflow: hidden;
  border: 1px solid var(--el-border-color-light);
  border-radius: 10px;
  background: #f8fafc;
}

.spreadsheet-editor-host {
  width: 100%;
  height: var(--spreadsheet-editor-height, 620px);
  min-height: 520px;
  max-height: 660px;
}

.spreadsheet-editor-host :deep(.x-spreadsheet) {
  width: 100%;
  height: 100% !important;
  overflow: hidden;
  border-radius: 10px;
  font-family: "Microsoft YaHei", Arial, sans-serif;
}

.spreadsheet-editor-host :deep(.x-spreadsheet-toolbar) {
  border-bottom-color: var(--el-border-color-light);
  background: #f8fafc;
}

.spreadsheet-editor-host :deep(.x-spreadsheet-bottombar) {
  border-top: 1px solid var(--el-border-color-light);
  background: #f8fafc;
}

/*
  x-data-spreadsheet 的工具栏 tooltip 会用 ::before 画一个黑色菱形箭头。
  在当前后台布局中 tooltip 容易被顶部导航裁切，只剩黑色菱形悬浮在页面上。
  这里隐藏该组件自带 tooltip，不影响单元格编辑/右键菜单/工具栏功能。
*/
:global(.x-spreadsheet-tooltip) {
  display: none !important;
}

:global(.x-spreadsheet-tooltip::before) {
  display: none !important;
}
</style>
