<template>
  <main class="na-page document-page">
    <AppPageHeader
      title-id="document-title"
      title="文档管理"
      description="支持 MD、TXT、HTML、PDF、Word、Excel 等文档上传、预览与在线版本维护。"
    >
      <template #actions>
        <el-upload
          :show-file-list="false"
          :http-request="uploadFile"
          :before-upload="beforeUpload"
          accept=".txt,.md,.markdown,.html,.htm,.json,.xml,.csv,.yaml,.yml,.ini,.conf,.log,.doc,.docs,.docx,.pdf,.xls,.xlsx"
        >
          <el-button type="primary" size="large" :icon="UploadFilled" :loading="uploading">上传文档</el-button>
        </el-upload>
      </template>
    </AppPageHeader>

    <section class="document-workspace">
      <aside class="na-panel document-list-panel">
        <header class="panel-header">
          <div>
            <h2>已上传文档</h2>
            <span>共 {{ total }} 个文档</span>
          </div>
          <el-button :icon="Refresh" text @click="loadDocuments">刷新</el-button>
        </header>

        <el-form :model="search" class="search-form" @keyup.enter="submitSearch">
          <el-input v-model="search.keyword" clearable placeholder="搜索文档名、后缀或备注" :prefix-icon="Search" />
          <el-select v-model="search.fileExt" clearable placeholder="全部格式">
            <el-option v-for="item in extOptions" :key="item" :label="item.toUpperCase()" :value="item" />
          </el-select>
          <div class="search-actions">
            <el-button type="primary" :icon="Search" @click="submitSearch">查询</el-button>
            <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
          </div>
        </el-form>

        <div v-loading="loading" class="document-cards">
          <div
            v-for="row in tableData"
            :key="row.ID"
            role="button"
            tabindex="0"
            class="doc-card"
            :class="{ active: row.ID === current.ID }"
            @click="openDocument(row)"
            @keyup.enter="openDocument(row)"
          >
            <span class="doc-icon" :class="docType(row).group">
              <el-icon><Document /></el-icon>
            </span>
            <span class="doc-card-body">
              <span class="doc-name">{{ row.title || row.originalName }}</span>
              <span class="doc-subtitle">{{ row.originalName }}</span>
              <span class="doc-meta">
                <el-tag size="small" effect="plain">{{ docType(row).label }}</el-tag>
                <span>{{ fileSize(row.fileSize) }}</span>
                <span>{{ dateText(row.UpdatedAt || row.updatedAt) }}</span>
              </span>
            </span>
            <el-button type="danger" link :icon="Delete" class="doc-delete" @click.stop="removeDocument(row)">删除</el-button>
          </div>

          <el-empty v-if="!loading && !tableData.length" description="暂无文档">
            <el-upload :show-file-list="false" :http-request="uploadFile" :before-upload="beforeUpload">
              <el-button type="primary">上传第一个文档</el-button>
            </el-upload>
          </el-empty>
        </div>

        <footer class="na-pagination pagination-wrap">
          <el-pagination
            v-model:current-page="search.page"
            v-model:page-size="search.pageSize"
            :page-sizes="[10, 20, 50]"
            :total="total"
            size="small"
            layout="total, sizes, prev, pager, next"
            @current-change="loadDocuments"
            @size-change="sizeChanged"
          />
        </footer>
      </aside>

      <section class="na-panel document-editor-panel">
        <template v-if="current.ID">
          <header class="editor-header">
            <div>
              <p class="eyebrow">{{ docType(current).groupLabel }}</p>
              <h2>{{ current.originalName }}</h2>
              <span>{{ current.fileExt?.toUpperCase() }} · {{ fileSize(current.fileSize) }} · {{ current.storageType || 'oss' }}</span>
            </div>
            <div class="editor-actions">
              <el-button :icon="Download" @click="downloadOriginal">下载原文件</el-button>
              <el-button :icon="Refresh" @click="reloadCurrent">重新加载</el-button>
              <el-button type="primary" :icon="Check" :loading="saving" :disabled="!editingStarted" @click="saveDocument">保存在线版本</el-button>
            </div>
          </header>

          <el-alert
            v-if="officeLike"
            class="mode-alert"
            type="info"
            :closable="false"
            show-icon
            :title="officeAlertTitle"
          />

          <el-tabs v-model="activePane" class="document-tabs">
            <el-tab-pane v-if="canPreviewOriginal" label="原文件预览（只读）" name="preview">
              <div class="preview-action-bar">
                <el-alert
                  class="mode-alert"
                  type="warning"
                  :closable="false"
                  show-icon
                  title="这里是原文件只读预览；需要修改内容请点击“开始编辑”。"
                />
                <el-button type="primary" :icon="Edit" @click="startEditing">开始编辑</el-button>
              </div>
              <div v-loading="previewLoading" class="preview-stage">
                <VueOfficeDocx
                  v-if="isDocxPreview && previewSource"
                  :src="previewSource"
                  class="office-viewer"
                  @error="handlePreviewError"
                />
                <VueOfficePdf
                  v-else-if="isPdfPreview && previewSource"
                  :src="previewSource"
                  class="office-viewer pdf-viewer"
                  @error="handlePreviewError"
                />
                <div v-else-if="isExcelPreview && previewSpreadsheetData.length" class="source-spreadsheet-preview">
                  <SpreadsheetEditor
                    :key="`source-excel-${editor.ID}-${previewSpreadsheetData.length}`"
                    :model-value="previewSpreadsheetData"
                    readonly
                  />
                </div>
                <article v-else-if="isMarkdown && previewText" class="source-markdown-preview" v-html="previewMarkdownHtml" />
                <pre v-else-if="isSourceTextPreview && previewText" class="source-text-preview">{{ previewText }}</pre>
                <el-empty v-else-if="previewError" :description="previewError">
                  <div class="empty-actions">
                    <el-button :icon="Refresh" @click="loadPreviewSource(current)">重新加载预览</el-button>
                    <el-button type="primary" :icon="Edit" @click="startEditing">开始编辑</el-button>
                  </div>
                </el-empty>
                <el-empty v-else :description="previewEmptyText">
                  <el-button type="primary" :icon="Edit" @click="startEditing">开始编辑</el-button>
                </el-empty>
              </div>
            </el-tab-pane>

            <el-tab-pane label="在线编辑（可保存）" name="edit">
              <div v-if="!editingStarted" class="edit-locked-card">
                <div>
                  <p class="eyebrow">EDIT LOCKED</p>
                  <h3>当前处于只读预览模式</h3>
                  <span>为避免误改，所有类型文档默认不进入编辑态；点击“开始编辑”后才会显示编辑器和保存能力。</span>
                </div>
                <el-button type="primary" size="large" :icon="Edit" @click="startEditing">开始编辑</el-button>
              </div>
              <el-form v-else label-position="top" class="editor-form">
                <el-form-item label="文档标题">
                  <el-input v-model="editor.title" maxlength="180" placeholder="请输入文档标题" />
                </el-form-item>

                <el-form-item :label="editorLabel">
                  <div class="edit-mode-banner">
                    <el-tag :type="editCapability.type" effect="dark">{{ editCapability.tag }}</el-tag>
                    <span>{{ editCapability.text }}</span>
                  </div>

                  <template v-if="isMarkdown">
                    <div class="markdown-editor">
                      <el-input
                        v-model="editor.content"
                        type="textarea"
                        :rows="18"
                        resize="none"
                        placeholder="请输入 Markdown 内容"
                      />
                      <article class="markdown-preview" v-html="markdownHtml" />
                    </div>
                  </template>

                  <template v-else-if="isPlainText">
                    <el-input
                      v-model="editor.content"
                      class="plain-editor"
                      type="textarea"
                      :rows="20"
                      resize="none"
                      :placeholder="plainPlaceholder"
                    />
                  </template>

                  <template v-else-if="isExcelPreview">
                    <div v-loading="excelLoading" class="excel-preview-panel">
                      <div class="excel-preview-toolbar">
                        <div class="excel-preview-title">
                          <strong>Excel 数据表格</strong>
                          <span>默认以只读表格展示，点击“编辑表格”后进入独立编辑框。</span>
                        </div>
                        <div class="excel-preview-actions">
                          <el-button :icon="Refresh" @click="loadExcelFromOriginal(true)">从原文件重新解析</el-button>
                          <el-button type="primary" :icon="Edit" @click="openExcelEditor">编辑表格</el-button>
                        </div>
                      </div>
                      <el-alert
                        class="excel-tip compact"
                        type="info"
                        :closable="false"
                        show-icon
                        title="当前页面仅展示在线表格数据；保存在线版本不会覆盖 OSS 中的原始 xlsx/xls 文件。"
                      />
                      <el-alert v-if="excelParseError" class="excel-tip" type="warning" :closable="false" show-icon :title="excelParseError" />

                      <div v-if="excelState.spreadsheetData.length" class="excel-readonly-preview">
                        <SpreadsheetEditor
                          :key="`excel-preview-${editor.ID}-${excelState.spreadsheetData.length}`"
                          :model-value="excelState.spreadsheetData"
                          readonly
                        />
                      </div>
                      <el-empty v-else description="暂无可展示的表格数据">
                        <el-button type="primary" :icon="Edit" @click="openExcelEditor">编辑表格</el-button>
                      </el-empty>
                    </div>
                  </template>

                  <template v-else>
                    <div class="rich-editor-wrap">
                      <div v-if="isWordDocument" class="word-toolbar">
                        <el-button v-if="canParseWordOriginal" :loading="wordLoading" :icon="Refresh" @click="loadWordFromOriginal(true)">从原 Word 解析正文</el-button>
                        <span class="word-tip">{{ wordEditTip }}</span>
                      </div>
                      <el-alert v-if="wordLoading" class="excel-tip" type="info" :closable="false" show-icon title="正在解析 Word 正文，解析完成后可在下方富文本区域编辑。" />
                      <el-alert v-if="wordParseError" class="excel-tip" type="warning" :closable="false" show-icon :title="wordParseError" />
                      <RichEdit
                        v-model="editor.content"
                        :height="richEditorHeight"
                      />
                    </div>
                  </template>
                </el-form-item>

                <el-form-item label="备注">
                  <el-input v-model="editor.remarks" type="textarea" :rows="2" maxlength="500" show-word-limit placeholder="补充编辑说明或版本备注" />
                </el-form-item>
              </el-form>
            </el-tab-pane>

          </el-tabs>
        </template>

        <el-empty v-else class="editor-empty" description="请从左侧点击文档名进入预览或在线编辑">
          <el-upload :show-file-list="false" :http-request="uploadFile" :before-upload="beforeUpload">
            <el-button type="primary" :icon="UploadFilled">上传文档</el-button>
          </el-upload>
        </el-empty>
      </section>
    </section>

    <el-dialog
      v-model="excelEditorVisible"
      append-to-body
      destroy-on-close
      class="excel-editor-dialog"
      width="min(96vw, 1480px)"
      top="4vh"
      :close-on-click-modal="false"
    >
      <template #header>
        <div class="excel-dialog-header">
          <div>
            <p class="eyebrow">EXCEL EDITOR</p>
            <h3>{{ current.originalName || '编辑表格' }}</h3>
            <span>在独立编辑框中修改表格，保存后写入系统在线版本。</span>
          </div>
          <el-button :icon="Refresh" @click="loadExcelFromOriginal(true)">从原文件重新解析</el-button>
        </div>
      </template>

      <div class="excel-dialog-body">
        <SpreadsheetEditor
          v-if="excelEditorVisible"
          :model-value="excelState.spreadsheetData"
          @update:model-value="updateSpreadsheetData"
        />
      </div>

      <template #footer>
        <div class="excel-dialog-footer">
          <span>提示：关闭编辑框不会覆盖原文件；只有点击“保存在线版本”才会写入系统在线内容。</span>
          <div>
            <el-button @click="excelEditorVisible = false">关闭</el-button>
            <el-button type="primary" :icon="Check" :loading="saving" @click="saveExcelAndClose">保存在线版本</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Delete, Document, Download, Edit, Refresh, Search, UploadFilled } from '@element-plus/icons-vue'
import { marked } from 'marked'
import '@vue-office/docx/lib/index.css'
import { deleteDocument, downloadDocumentFile, getDocumentDetail, getDocumentList, updateDocumentContent, uploadDocument } from '@/plugin/document/api/document'
import { formatDateText } from '@/utils/format'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { usePagedList } from '@/hooks/usePagedList'

const RichEdit = defineAsyncComponent(() => import('@/components/richtext/rich-edit.vue'))
const SpreadsheetEditor = defineAsyncComponent(() => import('@/plugin/document/components/SpreadsheetEditor.vue'))

const VueOfficeDocx = defineAsyncComponent(() => import('@vue-office/docx'))
const VueOfficePdf = defineAsyncComponent(() => import('@vue-office/pdf'))

const route = useRoute()
const router = useRouter()
const uploading = ref(false)
const saving = ref(false)
const previewLoading = ref(false)
const previewError = ref('')
const previewSource = ref(null)
const previewText = ref('')
const previewSpreadsheetData = ref([])
const activePane = ref('preview')
const editingStarted = ref(false)
const current = ref({})
const editor = reactive({ ID: 0, title: '', content: '', remarks: '' })
const excelState = reactive({ sheets: [], activeSheet: '', spreadsheetData: [] })
const excelLoading = ref(false)
const excelParseError = ref('')
const excelDirty = ref(false)
const excelHasStoredContent = ref(false)
const excelEditorVisible = ref(false)
const wordLoading = ref(false)
const wordParseError = ref('')
const wordHasStoredContent = ref(false)
const extOptions = ['md', 'txt', 'html', 'json', 'csv', 'doc', 'docx', 'pdf', 'xls', 'xlsx']
let previewRequestSeq = 0

const markdownExts = ['md', 'markdown']
const plainTextExts = ['txt', 'json', 'xml', 'csv', 'yaml', 'yml', 'ini', 'conf', 'log']
const docxPreviewExts = ['docx']
const docExts = ['doc', 'docs', 'docx']
const excelPreviewExts = ['xls', 'xlsx']
const pdfPreviewExts = ['pdf']
const officeExts = [...docExts, ...excelPreviewExts, ...pdfPreviewExts]

const normalizeExt = (value = '') => String(value || '').toLowerCase().replace(/^\./, '').trim()
const extFromDoc = (doc = {}) => normalizeExt(doc.fileExt || (doc.originalName || '').split('.').pop())

const fileSize = (size) => {
  const value = Number(size || 0)
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / 1024 / 1024).toFixed(1)} MB`
}

const dateText = formatDateText

const currentExt = computed(() => extFromDoc(current.value))
const isMarkdown = computed(() => markdownExts.includes(currentExt.value))
const isPlainText = computed(() => plainTextExts.includes(currentExt.value))
const isDocxPreview = computed(() => docxPreviewExts.includes(currentExt.value))
const isWordDocument = computed(() => docExts.includes(currentExt.value))
const canParseWordOriginal = computed(() => isDocxPreview.value)
const isExcelPreview = computed(() => excelPreviewExts.includes(currentExt.value))
const isPdfPreview = computed(() => pdfPreviewExts.includes(currentExt.value))
const isHtmlPreview = computed(() => ['html', 'htm'].includes(currentExt.value))
const isSourceTextPreview = computed(() => isMarkdown.value || isPlainText.value || isHtmlPreview.value)
const canPreviewOriginal = computed(() => !!current.value?.ID)
const officeLike = computed(() => officeExts.includes(currentExt.value))
const editorLabel = computed(() => {
  if (isMarkdown.value) return 'Markdown 在线编辑'
  if (isPlainText.value) return '文本在线编辑'
  if (isExcelPreview.value) return 'Excel 数据表格'
  if (isWordDocument.value) return 'Word 在线富文本编辑'
  if (officeLike.value) return '系统在线版本内容'
  return '在线编辑内容'
})

const editCapability = computed(() => {
  if (isExcelPreview.value) {
    return { type: 'info', tag: '只读展示', text: '普通情况下展示数据表格；需要修改时点击“编辑表格”，在独立编辑框中处理。' }
  }
  if (isWordDocument.value) {
    return { type: 'primary', tag: '富文本可编辑', text: '下方富文本区域可直接编辑；DOCX 可从原文件解析正文，DOC/DOCS 可手动维护在线版本。' }
  }
  if (isMarkdown.value || isPlainText.value) {
    return { type: 'success', tag: '文本可编辑', text: '下方文本框可直接编辑，保存后会更新系统在线版本。' }
  }
  if (isPdfPreview.value) {
    return { type: 'warning', tag: '预览为主', text: 'PDF 原文件不可直接改写，可在备注和在线内容中维护说明。' }
  }
  return { type: 'info', tag: '在线版本', text: '可在下方维护系统内的在线版本内容。' }
})

const wordEditTip = computed(() => {
  if (canParseWordOriginal.value) return '可在下方编辑 Word 的系统在线版本，保存不会覆盖原始 docx 文件。'
  return '当前格式暂不自动解析原文，但下方富文本区域可以直接编辑并保存系统在线版本。'
})

const officeAlertTitle = computed(() => {
  if (isExcelPreview.value) return 'Excel 原文件保存在 OSS 中，当前页面默认展示数据表格；点击编辑后可维护系统在线版本，保存不会覆盖原始 xlsx/xls 文件。'
  if (isDocxPreview.value) return 'Word 原文件保存在 OSS 中，当前页面可解析正文为富文本在线版本；保存不会覆盖原始 docx 文件。'
  if (isPdfPreview.value) return 'PDF 原文件保存在 OSS 中，当前页面提供预览和备注维护；PDF 本身不可直接在线编辑。'
  return 'Office 源文件保存在 OSS 中，当前页面提供系统在线版本维护；保存不会覆盖原始文件。'
})
const plainPlaceholder = computed(() => {
  if (currentExt.value === 'csv') return '请输入 CSV 内容'
  if (['json', 'xml', 'yaml', 'yml'].includes(currentExt.value)) return '请输入结构化文本内容'
  return '请输入文本内容'
})
const markdownHtml = computed(() => marked.parse(editor.content || ''))
const previewMarkdownHtml = computed(() => marked.parse(previewText.value || ''))
const previewEmptyText = computed(() => {
  if (isExcelPreview.value) return '正在解析原 Excel 文件'
  if (isSourceTextPreview.value) return '正在读取原文件内容'
  if (isDocxPreview.value || isPdfPreview.value) return '正在准备原文件预览'
  return '当前格式暂不支持内嵌预览，可下载原文件，或点击开始编辑维护在线版本'
})
const richEditorHeight = computed(() => {
  // 文档在线编辑以阅读/录入为主，给 Word 富文本更大的可视区；
  // 使用 clamp 保证大屏更舒展，小屏仍不会把底部备注完全挤出视口。
  if (isWordDocument.value) return 'clamp(620px, calc(100vh - 320px), 860px)'
  return '560px'
})


const excelStoredType = 'asset-center-excel-v1'
const spreadsheetStoredType = 'asset-center-spreadsheet-v2'
const normalizeExcelRows = (rows = []) => {
  const normalized = (Array.isArray(rows) ? rows : [])
    .map((row) => (Array.isArray(row) ? row : [row]).map((cell) => cell == null ? '' : String(cell)))
    .filter((row) => row.some((cell) => String(cell || '').trim() !== ''))
  if (!normalized.length) normalized.push([''])
  return normalized
}

const createBlankSpreadsheetSheet = (name = 'Sheet1') => ({
  name,
  rows: { len: 100 },
  cols: { len: 26 },
  merges: [],
  styles: [],
  validations: [],
  freeze: 'A1'
})

const rowsToSpreadsheetSheet = (sheet, index = 0) => {
  const rows = normalizeExcelRows(sheet?.rows)
  const rowData = { len: Math.max(100, rows.length + 20) }
  let maxCol = 26
  rows.forEach((row, rowIndex) => {
    const cells = {}
    row.forEach((cell, colIndex) => {
      maxCol = Math.max(maxCol, colIndex + 1)
      if (String(cell || '') !== '') {
        cells[colIndex] = { text: String(cell) }
      }
    })
    if (Object.keys(cells).length) {
      rowData[rowIndex] = { cells }
    }
  })
  return {
    ...createBlankSpreadsheetSheet(String(sheet?.name || `Sheet${index + 1}`)),
    rows: rowData,
    cols: { len: maxCol }
  }
}

const spreadsheetSheetToRows = (sheet) => {
  const rowsObj = sheet?.rows || {}
  const rows = []
  Object.keys(rowsObj).forEach((rowKey) => {
    const rowIndex = Number(rowKey)
    if (Number.isNaN(rowIndex)) return
    const row = []
    const cells = rowsObj[rowKey]?.cells || {}
    Object.keys(cells).forEach((colKey) => {
      const colIndex = Number(colKey)
      if (Number.isNaN(colIndex)) return
      row[colIndex] = cells[colKey]?.text == null ? '' : String(cells[colKey].text)
    })
    rows[rowIndex] = row
  })
  return normalizeExcelRows(rows)
}

const normalizeSpreadsheetData = (sheets = []) => {
  const source = Array.isArray(sheets) && sheets.length ? sheets : [createBlankSpreadsheetSheet()]
  return source.map((sheet, index) => {
    if (Array.isArray(sheet?.rows)) {
      return rowsToSpreadsheetSheet(sheet, index)
    }
    const rows = sheet?.rows && typeof sheet.rows === 'object' ? { ...sheet.rows } : { len: 100 }
    const cols = sheet?.cols && typeof sheet.cols === 'object' ? { ...sheet.cols } : { len: 26 }
    rows.len = Math.max(Number(rows.len || 0), 100)
    cols.len = Math.max(Number(cols.len || 0), 26)
    return {
      ...createBlankSpreadsheetSheet(String(sheet?.name || `Sheet${index + 1}`)),
      ...sheet,
      name: String(sheet?.name || `Sheet${index + 1}`),
      rows,
      cols,
      merges: Array.isArray(sheet?.merges) ? sheet.merges : [],
      styles: Array.isArray(sheet?.styles) ? sheet.styles : [],
      validations: Array.isArray(sheet?.validations) ? sheet.validations : [],
      freeze: sheet?.freeze || 'A1'
    }
  })
}

const workbookToSpreadsheetData = (workbook, reader) => {
  const sheets = []
  workbook.SheetNames.forEach((name, sheetIndex) => {
    const worksheet = workbook.Sheets[name]
    const sheet = createBlankSpreadsheetSheet(name || `Sheet${sheetIndex + 1}`)
    if (!worksheet || !worksheet['!ref']) {
      sheets.push(sheet)
      return
    }

    const range = reader.utils.decode_range(worksheet['!ref'])
    range.s = { r: 0, c: 0 }
    const aoa = reader.utils.sheet_to_json(worksheet, {
      raw: false,
      header: 1,
      range,
      defval: ''
    })

    const rows = { len: Math.max(100, range.e.r + 20) }
    const cols = { len: Math.max(26, range.e.c + 6) }
    const styles = []
    const styleIndexMap = new Map()
    const colorFromXlsx = (color) => {
      const value = String(color?.rgb || color?.indexed || '').trim()
      if (!value || value.length < 6 || /^\d+$/.test(value)) return ''
      return `#${value.slice(-6)}`
    }
    const normalizeAlign = (align) => {
      if (['left', 'center', 'right'].includes(align)) return align
      return 'left'
    }
    const normalizeValign = (align) => {
      if (align === 'top') return 'top'
      if (align === 'bottom') return 'bottom'
      return 'middle'
    }
    const styleOfCell = (cell) => {
      const rawStyle = cell?.s
      if (!rawStyle || typeof rawStyle !== 'object') return undefined
      const style = {
        bgcolor: colorFromXlsx(rawStyle.fill?.fgColor) || '#ffffff',
        align: normalizeAlign(rawStyle.alignment?.horizontal),
        valign: normalizeValign(rawStyle.alignment?.vertical),
        textwrap: !!rawStyle.alignment?.wrapText,
        color: colorFromXlsx(rawStyle.font?.color) || '#111827',
        font: {
          name: rawStyle.font?.name || 'Microsoft YaHei',
          size: Number(rawStyle.font?.sz || 10),
          bold: !!rawStyle.font?.bold,
          italic: !!rawStyle.font?.italic
        }
      }
      const key = JSON.stringify(style)
      if (!styleIndexMap.has(key)) {
        styleIndexMap.set(key, styles.length)
        styles.push(style)
      }
      return styleIndexMap.get(key)
    }
    ;(worksheet['!cols'] || []).forEach((col, index) => {
      const width = Number(col?.wpx || (col?.wch ? col.wch * 8 : 0))
      if (width > 0) cols[index] = { width: Math.max(60, Math.round(width)) }
    })
    ;(worksheet['!rows'] || []).forEach((row, index) => {
      const height = Number(row?.hpx || (row?.hpt ? row.hpt * 1.33 : 0))
      if (height > 0) rows[index] = { ...(rows[index] || {}), height: Math.max(24, Math.round(height)) }
    })

    aoa.forEach((row, rowIndex) => {
      const cells = {}
      row.forEach((cell, colIndex) => {
        const ref = reader.utils.encode_cell({ r: rowIndex, c: colIndex })
        const rawCell = worksheet[ref]
        let text = cell == null ? '' : String(cell)
        if (rawCell?.f) text = `=${rawCell.f}`
        const style = styleOfCell(rawCell)
        if (text !== '' || style !== undefined) {
          cells[colIndex] = { text }
          if (style !== undefined) cells[colIndex].style = style
        }
      })
      if (Object.keys(cells).length) {
        rows[rowIndex] = { ...(rows[rowIndex] || {}), cells }
      }
    })

    sheet.merges = []
    ;(worksheet['!merges'] || []).forEach((merge) => {
      const rowIndex = merge.s.r
      const colIndex = merge.s.c
      rows[rowIndex] = rows[rowIndex] || { cells: {} }
      rows[rowIndex].cells = rows[rowIndex].cells || {}
      rows[rowIndex].cells[colIndex] = rows[rowIndex].cells[colIndex] || { text: '' }
      rows[rowIndex].cells[colIndex].merge = [
        merge.e.r - merge.s.r,
        merge.e.c - merge.s.c
      ]
      sheet.merges.push(reader.utils.encode_range(merge))
      rows.len = Math.max(rows.len, merge.e.r + 20)
      cols.len = Math.max(cols.len, merge.e.c + 6)
    })

    sheet.rows = rows
    sheet.cols = cols
    sheet.styles = styles
    sheets.push(sheet)
  })
  return normalizeSpreadsheetData(sheets)
}

const resetExcelEditor = () => {
  excelState.sheets = []
  excelState.activeSheet = ''
  excelState.spreadsheetData = []
  excelParseError.value = ''
  excelDirty.value = false
  excelHasStoredContent.value = false
}

const parseStoredExcelContent = (content) => {
  const text = String(content || '').trim()
  if (!text) return null
  try {
    const data = JSON.parse(text)
    if (data?.type === spreadsheetStoredType && Array.isArray(data.sheets)) {
      return normalizeSpreadsheetData(data.sheets)
    }
    if (data?.type === excelStoredType && Array.isArray(data.sheets)) {
      return normalizeSpreadsheetData(data.sheets.map((sheet, index) => ({
        name: String(sheet?.name || `Sheet${index + 1}`),
        rows: normalizeExcelRows(sheet?.rows)
      })))
    }
    return null
  } catch {
    return null
  }
}

const hydrateExcelFromStoredContent = (content) => {
  const sheets = parseStoredExcelContent(content)
  if (!sheets) {
    resetExcelEditor()
    return false
  }
  const spreadsheetData = normalizeSpreadsheetData(sheets)
  excelState.spreadsheetData = spreadsheetData
  excelState.sheets = spreadsheetData.map((sheet) => ({ name: sheet.name, rows: spreadsheetSheetToRows(sheet) }))
  excelState.activeSheet = spreadsheetData[0]?.name || ''
  excelParseError.value = ''
  excelDirty.value = false
  excelHasStoredContent.value = true
  return true
}

const markExcelDirty = () => {
  excelDirty.value = true
}

const serializeExcelContent = () => JSON.stringify({
  type: spreadsheetStoredType,
  sheets: normalizeSpreadsheetData(excelState.spreadsheetData)
}, null, 2)

const applyExcelSheets = (sheets) => {
  const spreadsheetData = normalizeSpreadsheetData(sheets)
  excelState.spreadsheetData = spreadsheetData
  excelState.sheets = spreadsheetData.map((sheet) => ({ name: sheet.name, rows: spreadsheetSheetToRows(sheet) }))
  excelState.activeSheet = spreadsheetData[0]?.name || ''
  excelParseError.value = ''
  excelDirty.value = false
}

const parseExcelArrayBuffer = async (arrayBuffer, force = false) => {
  if (!isExcelPreview.value || !arrayBuffer) return
  if (!force && excelHasStoredContent.value) return
  excelLoading.value = true
  excelParseError.value = ''
  try {
    const XLSX = await import('xlsx')
    const reader = XLSX.default || XLSX
    const workbook = reader.read(arrayBuffer, { type: 'array', cellDates: true, cellStyles: true })
    const sheets = workbookToSpreadsheetData(workbook, reader)
    if (!sheets.length) {
      excelParseError.value = '原 Excel 文件没有可解析的数据，已创建空白在线表格。'
      applyExcelSheets([createBlankSpreadsheetSheet()])
      return
    }
    applyExcelSheets(sheets)
    excelHasStoredContent.value = false
  } catch (error) {
    excelParseError.value = 'Excel 原文件解析失败，请确认文件格式是否正确；仍可手动新增行列维护在线版本。'
    if (!excelState.spreadsheetData.length) applyExcelSheets([createBlankSpreadsheetSheet()])
  } finally {
    excelLoading.value = false
  }
}

const updateSpreadsheetData = (data) => {
  const previousActiveSheet = excelState.activeSheet
  const spreadsheetData = normalizeSpreadsheetData(data)
  excelState.spreadsheetData = spreadsheetData
  excelState.sheets = spreadsheetData.map((sheet) => ({ name: sheet.name, rows: spreadsheetSheetToRows(sheet) }))
  excelState.activeSheet = spreadsheetData.find((sheet) => sheet.name === previousActiveSheet)?.name || spreadsheetData[0]?.name || ''
  markExcelDirty()
}

const openExcelEditor = () => {
  if (!excelState.spreadsheetData.length) applyExcelSheets([createBlankSpreadsheetSheet()])
  excelEditorVisible.value = true
}

const loadExcelFromOriginal = async (force = false) => {
  if (!isExcelPreview.value) return
  if (!previewSource.value) await loadPreviewSource(current.value)
  if (previewSource.value) await parseExcelArrayBuffer(previewSource.value, force)
}

const loadWordFromArrayBuffer = async (arrayBuffer, force = false) => {
  if (!isDocxPreview.value || !arrayBuffer) return
  if (!force && wordHasStoredContent.value) return
  wordLoading.value = true
  wordParseError.value = ''
  try {
    const mammothModule = await import('mammoth/mammoth.browser')
    const mammoth = mammothModule.default || mammothModule
    const result = await mammoth.convertToHtml({ arrayBuffer })
    const html = String(result?.value || '').trim()
    if (html) {
      editor.content = html
      wordHasStoredContent.value = true
      return
    }
    wordParseError.value = 'Word 原文件未解析到正文内容，可直接在下方手动维护在线版本。'
  } catch {
    wordParseError.value = 'Word 原文件正文解析失败，可直接在下方手动维护在线版本。'
  } finally {
    wordLoading.value = false
  }
}

const loadWordFromOriginal = async (force = false) => {
  if (!isDocxPreview.value) return
  if (!previewSource.value) await loadPreviewSource(current.value)
  if (previewSource.value) await loadWordFromArrayBuffer(previewSource.value, force)
}

const docType = (doc = {}) => {
  const ext = extFromDoc(doc)
  if (markdownExts.includes(ext)) return { label: 'MD', group: 'markdown', groupLabel: 'MARKDOWN' }
  if (plainTextExts.includes(ext)) return { label: ext.toUpperCase(), group: 'text', groupLabel: 'TEXT DOCUMENT' }
  if (docExts.includes(ext)) return { label: ext === 'docx' ? 'DOCX' : 'DOC', group: 'word', groupLabel: 'WORD DOCUMENT' }
  if (excelPreviewExts.includes(ext)) return { label: ext.toUpperCase(), group: 'excel', groupLabel: 'EXCEL DOCUMENT' }
  if (pdfPreviewExts.includes(ext)) return { label: 'PDF', group: 'pdf', groupLabel: 'PDF DOCUMENT' }
  return { label: ext ? ext.toUpperCase() : 'FILE', group: 'file', groupLabel: 'DOCUMENT' }
}

const shouldPreview = (doc = {}) => {
  return !!doc?.ID
}

const cleanupOfficePlaceholder = (content, doc = {}) => {
  const ext = extFromDoc(doc)
  if (!officeExts.includes(ext)) return content || ''
  const value = String(content || '').trim()
  if (value.includes('文档已上传。可在此处维护在线编辑内容')) return ''
  return content || ''
}

const syncEditor = (doc) => {
  current.value = doc || {}
  editor.ID = doc?.ID || 0
  editor.title = doc?.title || ''
  editor.content = cleanupOfficePlaceholder(doc?.content, doc)
  editor.remarks = doc?.remarks || ''
  previewSource.value = null
  previewText.value = ''
  previewSpreadsheetData.value = []
  previewError.value = ''
  editingStarted.value = false
  wordParseError.value = ''
  wordHasStoredContent.value = !!String(editor.content || '').trim()
  if (excelPreviewExts.includes(extFromDoc(doc))) {
    if (!hydrateExcelFromStoredContent(editor.content)) {
      applyExcelSheets([createBlankSpreadsheetSheet()])
      excelHasStoredContent.value = false
    }
  } else {
    resetExcelEditor()
  }
  const defaultPane = doc?.ID ? 'preview' : 'edit'
  activePane.value = defaultPane
  nextTick(() => {
    if (editor.ID === (doc?.ID || 0)) activePane.value = defaultPane
  })
  if (shouldPreview(doc)) {
    loadPreviewSource(doc)
  }
}

const {
  search,
  items: tableData,
  total,
  loading,
  load: loadDocuments,
  submit: submitSearch,
  reset: resetSearch,
  changePageSize: sizeChanged
} = usePagedList({
  defaults: { page: 1, pageSize: 10, keyword: '', fileExt: '' },
  request: getDocumentList
})

const loadPreviewSource = async (doc = current.value) => {
  if (!doc?.ID || !shouldPreview(doc)) return
  const seq = ++previewRequestSeq
  previewLoading.value = true
  previewError.value = ''
  previewSource.value = null
  previewText.value = ''
  previewSpreadsheetData.value = []
  try {
    const res = await downloadDocumentFile({ id: doc.ID })
    if (seq !== previewRequestSeq) return
    const data = res?.data
    if (!data || !data.byteLength) {
      previewError.value = '原文件为空或无法读取'
      return
    }
    previewSource.value = data
    if (isSourceTextPreview.value) {
      const maxPreviewBytes = 1024 * 1024
      const previewBuffer = data.byteLength > maxPreviewBytes ? data.slice(0, maxPreviewBytes) : data
      previewText.value = new TextDecoder('utf-8').decode(previewBuffer)
      if (data.byteLength > maxPreviewBytes) {
        previewText.value += '\n\n……原文件较大，仅预览前 1MB 内容。'
      }
    }
    if (isExcelPreview.value) {
      try {
        const XLSX = await import('xlsx')
        const reader = XLSX.default || XLSX
        const workbook = reader.read(data, { type: 'array', cellDates: true, cellStyles: true })
        previewSpreadsheetData.value = workbookToSpreadsheetData(workbook, reader)
      } catch {
        previewError.value = 'Excel 原文件预览解析失败，可下载原文件查看，或点击开始编辑维护在线版本。'
      }
      await parseExcelArrayBuffer(data, false)
    }
    if (isDocxPreview.value && !String(editor.content || '').trim()) await loadWordFromArrayBuffer(data, false)
  } catch (error) {
    if (seq === previewRequestSeq) {
      previewError.value = '原文件预览加载失败，请下载后查看'
    }
  } finally {
    if (seq === previewRequestSeq) previewLoading.value = false
  }
}

const loadDetail = async (id) => {
  if (!id) {
    syncEditor({})
    return
  }
  const res = await getDocumentDetail({ id })
  if (res.code === 0) {
    syncEditor(res.data)
  }
}

const startEditing = async () => {
  if (!editor.ID) return
  editingStarted.value = true
  activePane.value = 'edit'
  if (isExcelPreview.value && !excelState.spreadsheetData.length) {
    await loadExcelFromOriginal(false)
  }
  if (isDocxPreview.value && !String(editor.content || '').trim()) {
    await loadWordFromOriginal(false)
  }
}

const openDocument = (row) => {
  if (!row?.ID) return
  activePane.value = 'preview'
  editingStarted.value = false
  router.push({ name: 'documentViewer', query: { id: row.ID } })
}

const reloadCurrent = async () => {
  await loadDetail(editor.ID)
}

const beforeUpload = (file) => {
  const maxSize = 100 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error('文档大小不能超过 100MB')
    return false
  }
  return true
}

const uploadFile = async ({ file }) => {
  uploading.value = true
  try {
    const res = await uploadDocument(file)
    if (res.code === 0) {
      ElMessage.success('文档上传成功')
      await loadDocuments()
      openDocument(res.data)
    }
  } finally {
    uploading.value = false
  }
}

const saveDocument = async () => {
  if (!editor.ID) return false
  if (!editingStarted.value) {
    ElMessage.warning('请先点击“开始编辑”后再保存')
    return false
  }
  if (isExcelPreview.value) {
    editor.content = serializeExcelContent()
  }
  saving.value = true
  try {
    const res = await updateDocumentContent({ ...editor })
    if (res.code === 0) {
      ElMessage.success('在线版本保存成功')
      syncEditor(res.data)
      await loadDocuments()
      return true
    }
    return false
  } finally {
    saving.value = false
  }
}

const saveExcelAndClose = async () => {
  const ok = await saveDocument()
  if (ok) excelEditorVisible.value = false
}

const downloadOriginal = async () => {
  if (!current.value?.ID) return
  try {
    const res = await downloadDocumentFile({ id: current.value.ID })
    const blob = new Blob([res.data], { type: current.value.mimeType || 'application/octet-stream' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = current.value.originalName || current.value.title || 'document'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch {
    ElMessage.error('下载原文件失败')
  }
}

const handlePreviewError = () => {
  previewError.value = '当前文件暂不支持在线预览，请下载后查看'
}

const removeDocument = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除文档「${row.title || row.originalName}」吗？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  const res = await deleteDocument({ id: row.ID })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    if (editor.ID === row.ID) {
      syncEditor({})
      router.replace({ name: 'documentViewer' })
    }
    await loadDocuments()
  }
}

watch(
  () => route.query.id,
  (id) => loadDetail(id),
)

onMounted(async () => {
  await loadDocuments()
  if (route.query.id) await loadDetail(route.query.id)
})

onBeforeUnmount(() => {
  previewRequestSeq++
})
</script>

<style scoped lang="scss">
.document-page {
  display: flex;
  flex-direction: column;
}

.editor-header h2,
.panel-header h2 {
  margin: 0;
  color: var(--el-text-color-primary);
}

.panel-header span,
.editor-header span {
  color: var(--el-text-color-secondary);
}

.document-workspace {
  display: grid;
  grid-template-columns: 360px minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.document-list-panel,
.document-editor-panel {
  min-width: 0;
  padding: 18px;
}

.document-editor-panel { margin-top: 0; }

.document-list-panel {
  position: sticky;
  top: 16px;
}

.panel-header,
.editor-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 16px;
}

.search-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 120px;
  gap: 10px;
  margin-bottom: 14px;
}

.search-actions {
  display: flex;
  grid-column: 1 / -1;
  gap: 10px;
}

.document-cards {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 280px;
}

.doc-card {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr) auto;
  gap: 10px;
  align-items: center;
  width: 100%;
  padding: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 14px;
  background: var(--el-fill-color-blank);
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.doc-card:hover,
.doc-card.active {
  border-color: var(--el-color-primary-light-5);
  box-shadow: 0 10px 24px rgb(64 158 255 / 12%);
  transform: translateY(-1px);
}

.doc-card.active {
  background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-bg-color));
}

.doc-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 18px;
}

.doc-icon.excel {
  background: #ecfdf3;
  color: #16a34a;
}

.doc-icon.word,
.doc-icon.markdown {
  background: #eff6ff;
  color: #2563eb;
}

.doc-icon.pdf {
  background: #fef2f2;
  color: #dc2626;
}

.doc-icon.text {
  background: #f8fafc;
  color: #475569;
}

.doc-card-body {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 5px;
}

.doc-name,
.doc-subtitle {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.doc-name {
  color: var(--el-text-color-primary);
  font-weight: 650;
}

.doc-subtitle,
.doc-meta {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.doc-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.doc-delete {
  min-width: 44px;
}

.pagination-wrap {
  margin-top: 14px;
}

.editor-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.mode-alert {
  margin-bottom: 14px;
}

.preview-action-bar {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 14px;
}

.preview-action-bar .mode-alert {
  flex: 1;
  margin-bottom: 0;
}

.document-tabs :deep(.el-tabs__content) {
  overflow: visible;
}

.editor-form :deep(.richtext-wrapper) {
  border-radius: 12px;
  overflow: hidden;
}

.rich-editor-wrap {
  width: 100%;
}

.edit-locked-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  min-height: 220px;
  padding: 28px;
  border: 1px dashed var(--el-color-primary-light-5);
  border-radius: 16px;
  background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-bg-color));
}

.edit-locked-card h3 {
  margin: 0 0 8px;
  color: var(--el-text-color-primary);
  font-size: 18px;
}

.edit-locked-card span {
  color: var(--el-text-color-secondary);
  line-height: 1.7;
}

.edit-mode-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  margin-bottom: 12px;
  padding: 10px 12px;
  border: 1px dashed var(--el-color-primary-light-5);
  border-radius: 12px;
  background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-bg-color));
  color: var(--el-text-color-regular);
  font-size: 13px;
}

.word-toolbar,
.excel-toolbar,
.excel-preview-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.excel-preview-toolbar {
  justify-content: space-between;
  padding: 14px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  background: linear-gradient(135deg, #f8fafc, var(--el-bg-color));
}

.excel-preview-title {
  display: flex;
  min-width: 240px;
  flex-direction: column;
  gap: 4px;
}

.excel-preview-title strong {
  color: var(--el-text-color-primary);
  font-size: 15px;
}

.excel-preview-title span,
.excel-dialog-header span,
.excel-dialog-footer span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.excel-preview-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.word-tip {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.excel-toolbar-tip {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.excel-editor,
.excel-preview-panel {
  width: 100%;
}

.excel-edit-status {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px 12px;
  border-radius: 999px;
  background: #ecfdf3;
  color: #15803d;
  font-size: 13px;
  font-weight: 600;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #22c55e;
  box-shadow: 0 0 0 5px rgb(34 197 94 / 14%);
}

.excel-tip {
  margin-bottom: 12px;
}

.excel-tip.compact {
  margin-top: 10px;
}

.excel-sheet-switcher {
  margin: 12px 0;
  overflow-x: auto;
  white-space: nowrap;
}

.excel-table-wrap {
  max-height: min(60vh, 640px);
  overflow: auto;
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  background: #fff;
}

.excel-preview-table {
  width: max-content;
  min-width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  color: var(--el-text-color-primary);
  font-size: 13px;
  table-layout: fixed;
}

.excel-preview-table th,
.excel-preview-table td {
  min-width: 140px;
  max-width: 320px;
  height: 36px;
  padding: 8px 10px;
  border-right: 1px solid var(--el-border-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: #fff;
  line-height: 1.45;
  text-align: left;
  vertical-align: middle;
  white-space: pre-wrap;
}

.excel-preview-table thead th {
  position: sticky;
  z-index: 2;
  top: 0;
  background: #f8fafc;
  color: var(--el-text-color-secondary);
  font-weight: 700;
  text-align: center;
}

.excel-preview-table .corner-cell,
.excel-preview-table .row-index {
  position: sticky;
  z-index: 3;
  left: 0;
  min-width: 54px;
  width: 54px;
  background: #f8fafc;
  color: var(--el-text-color-secondary);
  text-align: center;
  font-weight: 600;
}

.excel-preview-table .corner-cell {
  top: 0;
  z-index: 4;
}

.excel-preview-table td.is-empty {
  color: transparent;
}

.excel-dialog-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.excel-dialog-header h3 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 18px;
  font-weight: 700;
}

.excel-dialog-body {
  min-height: 560px;
}

.excel-dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

:global(.excel-editor-dialog .el-dialog__body) {
  padding: 0 20px 16px;
}

:global(.excel-editor-dialog .el-dialog__footer) {
  padding: 14px 20px 18px;
  border-top: 1px solid var(--el-border-color-light);
}

:global(.excel-editor-dialog .el-dialog__header) {
  padding: 18px 20px 14px;
  border-bottom: 1px solid var(--el-border-color-light);
}

.markdown-editor {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 14px;
  width: 100%;
}

.markdown-editor :deep(.el-textarea__inner),
.plain-editor :deep(.el-textarea__inner) {
  min-height: 520px !important;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  line-height: 1.65;
}

.markdown-preview {
  min-height: 520px;
  max-height: 640px;
  overflow: auto;
  padding: 16px 18px;
  border: 1px solid var(--el-border-color);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
  color: var(--el-text-color-primary);
  line-height: 1.75;
}

.markdown-preview :deep(h1),
.markdown-preview :deep(h2),
.markdown-preview :deep(h3) {
  margin: 1em 0 0.5em;
  font-weight: 700;
}

.markdown-preview :deep(pre) {
  overflow: auto;
  padding: 12px;
  border-radius: 8px;
  background: #0f172a;
  color: #e2e8f0;
}

.markdown-preview :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

.preview-stage {
  min-height: 620px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-light);
  border-radius: 14px;
  background: #f8fafc;
}

.source-spreadsheet-preview {
  min-height: 620px;
  background: #fff;
}

.source-text-preview,
.source-markdown-preview {
  min-height: 620px;
  max-height: 72vh;
  margin: 0;
  overflow: auto;
  padding: 22px 24px;
  background: #fff;
  color: var(--el-text-color-primary);
}

.source-text-preview {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.source-markdown-preview {
  line-height: 1.8;
}

.empty-actions {
  display: inline-flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: center;
}

.office-viewer {
  width: 100%;
  height: 72vh;
  min-height: 620px;
  background: #fff;
}

.pdf-viewer {
  background: #64748b;
}

.editor-empty {
  min-height: 520px;
}

@media (max-width: 1280px) {
  .document-workspace {
    grid-template-columns: 320px minmax(0, 1fr);
  }
}

@media (max-width: 1100px) {
  .document-workspace,
  .markdown-editor {
    grid-template-columns: 1fr;
  }

  .document-list-panel {
    position: static;
  }
}

@media (max-width: 640px) {
  .panel-header,
  .editor-header {
    flex-direction: column;
  }

  .preview-action-bar,
  .edit-locked-card {
    flex-direction: column;
    align-items: stretch;
  }

  .search-form,
  .doc-card {
    grid-template-columns: 1fr;
  }

  .doc-icon {
    display: none;
  }

  .editor-actions {
    justify-content: flex-start;
  }
}
</style>
