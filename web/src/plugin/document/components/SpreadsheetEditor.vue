<template>
  <div class="spreadsheet-editor-shell">
    <div class="spreadsheet-toolbar">
      <el-tabs v-model="activeSheetIndex" class="sheet-tabs" @tab-change="emitReady">
        <el-tab-pane v-for="(sheet, index) in sheets" :key="sheet.id" :label="sheet.name" :name="index" />
      </el-tabs>
      <div v-if="!readonly" class="spreadsheet-actions">
        <el-button :icon="Plus" title="新增工作表" @click="addSheet" />
        <el-button :icon="Rows3" title="新增行" @click="addRow" />
        <el-button :icon="Columns3" title="新增列" @click="addColumn" />
        <el-button :icon="Trash2" title="删除当前工作表" :disabled="sheets.length <= 1" @click="removeSheet" />
      </div>
    </div>

    <div class="spreadsheet-grid-wrap">
      <table v-if="activeSheet" class="spreadsheet-grid">
        <thead>
          <tr>
            <th class="row-number">#</th>
            <th v-for="columnIndex in activeSheet.columnCount" :key="columnIndex">
              {{ columnLabel(columnIndex - 1) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rowIndex in activeSheet.rowCount" :key="rowIndex">
            <th class="row-number">{{ rowIndex }}</th>
            <td v-for="columnIndex in activeSheet.columnCount" :key="columnIndex">
              <span v-if="readonly">{{ cellValue(rowIndex - 1, columnIndex - 1) }}</span>
              <el-input
                v-else
                :model-value="cellValue(rowIndex - 1, columnIndex - 1)"
                @update:model-value="updateCell(rowIndex - 1, columnIndex - 1, $event)"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { Bottom as Rows3, Delete as Trash2, Plus, Right as Columns3 } from '@element-plus/icons-vue'

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

const minimumRows = 30
const minimumColumns = 12
const maximumRows = 200
const maximumColumns = 50
const sheets = ref([])
const activeSheetIndex = ref(0)
let lastEmittedPayload = ''

const clone = (value) => JSON.parse(JSON.stringify(value ?? null))

const createBlankSheet = (name = 'Sheet1') => ({
  id: crypto.randomUUID(),
  name,
  rowCount: minimumRows,
  columnCount: minimumColumns,
  values: [],
  source: { name, rows: { len: 100 }, cols: { len: 26 }, merges: [], styles: [], validations: [], freeze: 'A1' }
})

const countUsedColumns = (rows = {}) => Object.entries(rows).reduce((maximum, [rowKey, row]) => {
  if (rowKey === 'len' || !row?.cells) return maximum
  const rowMaximum = Object.keys(row.cells).reduce((value, columnKey) => Math.max(value, Number(columnKey) + 1 || 0), 0)
  return Math.max(maximum, rowMaximum)
}, 0)

const toEditableSheet = (source, index) => {
  const rows = source?.rows && typeof source.rows === 'object' ? source.rows : {}
  const usedRows = Object.keys(rows).reduce((maximum, rowKey) => rowKey === 'len' ? maximum : Math.max(maximum, Number(rowKey) + 1 || 0), 0)
  const rowCount = Math.min(maximumRows, Math.max(minimumRows, usedRows))
  const columnCount = Math.min(maximumColumns, Math.max(minimumColumns, countUsedColumns(rows)))
  const values = Array.from({ length: rowCount }, (_, rowIndex) => (
    Array.from({ length: columnCount }, (_, columnIndex) => {
      const value = rows?.[rowIndex]?.cells?.[columnIndex]?.text
      return value == null ? '' : String(value)
    })
  ))
  return {
    id: crypto.randomUUID(),
    name: String(source?.name || `Sheet${index + 1}`),
    rowCount,
    columnCount,
    values,
    source: clone(source || createBlankSheet().source)
  }
}

const activeSheet = computed(() => sheets.value[Number(activeSheetIndex.value)] || sheets.value[0])

const sheetPayload = (sheet) => {
  const source = clone(sheet.source || {})
  const originalRows = source.rows && typeof source.rows === 'object' ? source.rows : {}
  const rows = { len: Math.max(100, sheet.rowCount) }
  for (let rowIndex = 0; rowIndex < sheet.rowCount; rowIndex += 1) {
    const originalRow = originalRows[rowIndex] && typeof originalRows[rowIndex] === 'object' ? clone(originalRows[rowIndex]) : {}
    const originalCells = originalRow.cells && typeof originalRow.cells === 'object' ? originalRow.cells : {}
    const cells = {}
    for (let columnIndex = 0; columnIndex < sheet.columnCount; columnIndex += 1) {
      const text = String(sheet.values[rowIndex]?.[columnIndex] ?? '')
      const originalCell = originalCells[columnIndex] && typeof originalCells[columnIndex] === 'object' ? clone(originalCells[columnIndex]) : {}
      if (text !== '' || Object.keys(originalCell).length > 0) cells[columnIndex] = { ...originalCell, text }
    }
    if (Object.keys(cells).length > 0 || Object.keys(originalRow).some((key) => key !== 'cells')) {
      rows[rowIndex] = { ...originalRow, cells }
    }
  }
  return {
    ...source,
    name: sheet.name,
    rows,
    cols: { ...(source.cols || {}), len: Math.max(26, sheet.columnCount) },
    merges: Array.isArray(source.merges) ? source.merges : [],
    styles: Array.isArray(source.styles) ? source.styles : [],
    validations: Array.isArray(source.validations) ? source.validations : [],
    freeze: source.freeze || 'A1'
  }
}

const emitChange = () => {
  const payload = sheets.value.map(sheetPayload)
  lastEmittedPayload = JSON.stringify(payload)
  emits('update:modelValue', payload)
}

const syncFromProps = (value) => {
  const normalized = Array.isArray(value) && value.length ? value : [createBlankSheet().source]
  const payload = JSON.stringify(normalized)
  if (payload === lastEmittedPayload) return
  sheets.value = normalized.map(toEditableSheet)
  activeSheetIndex.value = Math.min(Number(activeSheetIndex.value) || 0, sheets.value.length - 1)
}

const cellValue = (rowIndex, columnIndex) => activeSheet.value?.values?.[rowIndex]?.[columnIndex] ?? ''

const updateCell = (rowIndex, columnIndex, value) => {
  const sheet = activeSheet.value
  if (!sheet) return
  sheet.values[rowIndex][columnIndex] = String(value ?? '')
  emitChange()
}

const addRow = () => {
  const sheet = activeSheet.value
  if (!sheet || sheet.rowCount >= maximumRows) return
  sheet.values.push(Array.from({ length: sheet.columnCount }, () => ''))
  sheet.rowCount += 1
  emitChange()
}

const addColumn = () => {
  const sheet = activeSheet.value
  if (!sheet || sheet.columnCount >= maximumColumns) return
  sheet.values.forEach((row) => row.push(''))
  sheet.columnCount += 1
  emitChange()
}

const addSheet = () => {
  const sheet = createBlankSheet(`Sheet${sheets.value.length + 1}`)
  sheets.value.push(sheet)
  activeSheetIndex.value = sheets.value.length - 1
  emitChange()
}

const removeSheet = () => {
  if (sheets.value.length <= 1) return
  sheets.value.splice(Number(activeSheetIndex.value), 1)
  activeSheetIndex.value = Math.max(0, Math.min(Number(activeSheetIndex.value), sheets.value.length - 1))
  emitChange()
}

const columnLabel = (index) => {
  let value = index + 1
  let label = ''
  while (value > 0) {
    value -= 1
    label = String.fromCharCode(65 + (value % 26)) + label
    value = Math.floor(value / 26)
  }
  return label
}

const emitReady = () => emits('ready', { sheetCount: sheets.value.length })

watch(() => props.modelValue, syncFromProps, { deep: false, immediate: true })
onMounted(emitReady)
</script>

<style scoped lang="scss">
.spreadsheet-editor-shell {
  width: 100%;
  overflow: hidden;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background: var(--el-bg-color);
}

.spreadsheet-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 52px;
  padding: 0 12px;
  border-bottom: 1px solid var(--el-border-color-light);
}

.sheet-tabs {
  min-width: 0;
  flex: 1;
}

.sheet-tabs :deep(.el-tabs__header) {
  margin: 0;
}

.spreadsheet-actions {
  display: flex;
  gap: 6px;
}

.spreadsheet-grid-wrap {
  height: min(620px, calc(100vh - 250px));
  min-height: 480px;
  overflow: auto;
}

.spreadsheet-grid {
  width: max-content;
  min-width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  table-layout: fixed;
}

.spreadsheet-grid th,
.spreadsheet-grid td {
  width: 140px;
  min-width: 140px;
  height: 36px;
  padding: 0;
  border-right: 1px solid var(--el-border-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}

.spreadsheet-grid thead th {
  position: sticky;
  top: 0;
  z-index: 2;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
}

.spreadsheet-grid .row-number {
  position: sticky;
  left: 0;
  z-index: 1;
  width: 52px;
  min-width: 52px;
  text-align: center;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
}

.spreadsheet-grid thead .row-number {
  z-index: 3;
}

.spreadsheet-grid td > span {
  display: block;
  overflow: hidden;
  padding: 0 8px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spreadsheet-grid :deep(.el-input__wrapper) {
  min-height: 35px;
  padding: 0 8px;
  border-radius: 0;
  box-shadow: none;
}
</style>
