<template>
  <main class="na-page na-page--list asset-recognition-page">
    <AppPageHeader title-id="asset-recognition-title" title="智能建档" description="上传照片或铭牌，生成待确认资产草稿。">
      <template #actions><el-button :icon="Refresh" :loading="loading" @click="loadJobs">刷新任务</el-button></template>
    </AppPageHeader>

    <section class="na-panel upload-panel">
      <div class="section-heading"><div><h2>创建识别任务</h2><p>最多上传 6 张图片，用户确认前不会创建正式资产。</p></div><el-tag type="info">图片仅用于本次建档</el-tag></div>
      <el-upload
        ref="uploadRef"
        v-model:file-list="uploadFiles"
        class="recognition-upload"
        drag
        multiple
        :limit="6"
        :auto-upload="false"
        accept="image/jpeg,image/png,image/webp,image/gif"
        :on-exceed="onExceed"
        :before-upload="beforeUpload"
      >
        <el-icon class="upload-icon"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖入资产照片，或 <em>选择文件</em></div>
        <template #tip><div class="el-upload__tip">支持 JPG、PNG、WebP、GIF；单张不超过 10MB</div></template>
      </el-upload>
      <div class="upload-actions"><el-button type="primary" :icon="MagicStick" :loading="creating" :disabled="!uploadFiles.length" @click="createJob">开始智能识别</el-button><el-button text @click="clearUpload">清空</el-button></div>
    </section>

    <section class="na-panel task-panel">
      <div class="section-heading"><div><h2>我的识别任务</h2><p>任务完成后进入人工复核，确认后与正式资产建立不可变关联。</p></div><el-select v-model="statusFilter" clearable placeholder="全部状态" class="status-filter" @change="loadJobs"><el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" /></el-select></div>
      <el-table :data="jobs" stripe class="task-table">
        <el-table-column label="任务" min-width="180"><template #default="{ row }"><div class="task-identity"><strong>#{{ row.ID }}</strong><span>{{ (row.fileKeys || []).length }} 张图片</span></div></template></el-table-column>
        <el-table-column label="识别结果" min-width="220"><template #default="{ row }"><div class="task-result"><strong>{{ row.draft?.name || row.result?.name || '等待识别' }}</strong><span>{{ row.draft?.serialNumber || row.result?.serialNumber || '序列号待确认' }}</span></div></template></el-table-column>
        <el-table-column label="状态" width="120"><template #default="{ row }"><el-tag :type="statusMeta(row.status).type">{{ statusMeta(row.status).label }}</el-tag></template></el-table-column>
        <el-table-column label="尝试" width="82" align="right"><template #default="{ row }">{{ row.attempts || 0 }} / {{ row.maxAttempts || 3 }}</template></el-table-column>
        <el-table-column label="更新时间" width="170"><template #default="{ row }">{{ dateText(row.updatedAt || row.createdAt) }}</template></el-table-column>
        <el-table-column label="操作" width="190" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openJob(row)">查看任务</el-button><el-button v-if="row.status === 'failed'" link type="warning" :loading="retryingId === row.ID" @click="retryJob(row)">重试</el-button><el-button v-if="!['completed', 'processing'].includes(row.status)" link type="danger" :loading="deletingId === row.ID" @click="deleteJob(row)">{{ row.status === 'deleting' ? '重试删除' : '删除' }}</el-button></template></el-table-column>
        <template #empty>
          <AppEmptyState
            compact
            :title="statusFilter ? '当前状态没有识别任务' : '还没有智能建档任务'"
            description="在上方上传资产照片或铭牌后，系统会生成待人工确认的资产草稿。"
            :highlights="['最多上传 6 张图片', '确认前不写入正式资产', '任务保留识别与修正状态']"
          >
            <template v-if="statusFilter" #actions>
              <el-button :icon="Refresh" @click="statusFilter = ''; loadJobs()">查看全部任务</el-button>
            </template>
          </AppEmptyState>
        </template>
      </el-table>
      <el-pagination v-if="total > pageSize" v-model:current-page="page" :page-size="pageSize" :total="total" layout="prev, pager, next" @current-change="loadJobs" />
    </section>

    <el-drawer v-model="detailVisible" :size="drawerSize" append-to-body destroy-on-close>
      <template #header><div class="drawer-heading"><strong>智能建档任务 #{{ activeJob?.ID }}</strong><el-tag v-if="activeJob" :type="statusMeta(activeJob.status).type">{{ statusMeta(activeJob.status).label }}</el-tag></div></template>
      <el-skeleton v-if="detailLoading && !activeJob" :rows="8" animated />
      <template v-else-if="activeJob">
        <el-alert v-if="activeJob.lastError" type="error" :title="activeJob.lastError" show-icon :closable="false" class="detail-alert" />
        <div class="recognition-layout">
          <section class="evidence-column">
            <h3>输入图片</h3>
            <div class="photo-grid"><el-image v-for="(photo, index) in activeJob.fileKeys || []" :key="photo.key || index" :src="photoUrl(photo)" fit="cover" :preview-src-list="(activeJob.fileKeys || []).map(photoUrl)" :initial-index="index" class="evidence-photo"><template #error><div class="photo-error">图片不可用</div></template></el-image></div>
            <div v-if="activeJob.result?.rawText" class="raw-text"><span>铭牌原文</span><pre>{{ activeJob.result.rawText }}</pre></div>
            <div v-if="confidenceEntries.length" class="confidence-list"><h3>字段置信度</h3><div v-for="item in confidenceEntries" :key="item.field" class="confidence-item"><span>{{ confidenceLabel(item.field) }}</span><el-progress :percentage="item.percentage" :status="item.percentage < 70 ? 'warning' : 'success'" :stroke-width="6" /></div></div>
            <div v-if="activeJob.warnings?.length" class="warning-list"><div v-for="warning in activeJob.warnings" :key="`${warning.code}-${warning.field}`" :class="['warning-item', warning.severity]"><el-icon><WarningFilled /></el-icon><span>{{ warning.message }}</span></div></div>
            <div v-if="activeJob.duplicateCandidates?.length" class="duplicate-list"><h3>重复候选资产</h3><div v-for="candidate in activeJob.duplicateCandidates" :key="candidate.assetId" class="duplicate-item"><strong>{{ candidate.assetCode }} · {{ candidate.name }}</strong><span>{{ candidate.serialNumber }} · {{ candidate.matchType === 'exact' ? '完全匹配' : '标准化匹配' }}</span></div></div>
          </section>
          <section class="draft-column">
            <el-form ref="draftFormRef" :model="draft" :rules="draftRules" label-position="top" class="draft-form">
              <div class="form-title"><h3>待确认草稿</h3><small>低置信字段请人工核对</small></div>
              <el-alert v-if="activeJob.status === 'processing' || activeJob.status === 'pending'" type="info" title="识别处理中，页面会自动刷新" show-icon :closable="false" class="detail-alert" />
              <div class="form-grid">
                <el-form-item label="资产编号" prop="assetCode"><el-input v-model="draft.assetCode" maxlength="80" placeholder="由业务人员填写" /></el-form-item>
                <el-form-item label="资产名称" prop="name"><el-input v-model="draft.name" maxlength="150" /></el-form-item>
                <el-form-item label="资产分类" prop="categoryId"><el-select v-model="draft.categoryId" filterable placeholder="选择分类"><el-option v-for="category in categories" :key="category.ID" :label="category.name" :value="category.ID" /></el-select></el-form-item>
                <el-form-item label="品牌"><el-input v-model="draft.brand" maxlength="100" /></el-form-item>
                <el-form-item label="型号"><el-input v-model="draft.model" maxlength="120" /></el-form-item>
                <el-form-item label="序列号"><el-input v-model="draft.serialNumber" maxlength="120" /></el-form-item>
                <el-form-item label="生产日期"><el-date-picker v-model="draft.productionDate" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" /></el-form-item>
                <el-form-item label="推荐质保（月）"><el-input-number v-model="draft.recommendedWarrantyMonths" :min="0" :max="120" controls-position="right" /></el-form-item>
              </div>
              <el-form-item label="规格参数"><el-input v-model="draft.specifications" type="textarea" :rows="2" maxlength="1000" /></el-form-item>
              <div class="form-grid valuation-grid">
                <el-form-item label="数量" prop="quantity"><el-input-number v-model="draft.quantity" :min="1" controls-position="right" /></el-form-item>
                <el-form-item label="计量单位"><el-input v-model="draft.unit" maxlength="30" /></el-form-item>
                <el-form-item label="采购单价（元）" prop="unitPrice"><el-input-number v-model="draft.unitPrice" :min="0" :precision="2" controls-position="right" /></el-form-item>
                <el-form-item label="当前估值（元）" prop="currentValue"><el-input-number v-model="draft.currentValue" :min="0" :max="originalValue || 999999999" :precision="2" controls-position="right" /></el-form-item>
                <el-form-item label="供应商"><el-input v-model="draft.supplier" maxlength="150" /></el-form-item>
                <el-form-item label="购置日期"><el-date-picker v-model="draft.purchaseDate" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" /></el-form-item>
                <el-form-item label="质保到期日"><el-date-picker v-model="draft.warrantyEndDate" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" /></el-form-item>
              </div>
              <el-form-item label="备注"><el-input v-model="draft.remarks" type="textarea" :rows="2" maxlength="1000" /></el-form-item>
            </el-form>
          </section>
        </div>
      </template>
      <template #footer><div class="drawer-actions"><el-button @click="detailVisible = false">关闭</el-button><el-button v-if="activeJob && ['reviewing'].includes(activeJob.status)" :loading="saving" @click="saveDraft">保存草稿</el-button><el-button v-if="activeJob && activeJob.status === 'reviewing'" type="primary" :loading="confirming" @click="confirmJob">确认建档</el-button><el-button v-if="activeJob && activeJob.status === 'failed'" type="warning" :loading="retrying" @click="retryJob(activeJob)">重新识别</el-button></div></template>
    </el-drawer>
  </main>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useWindowSize } from '@vueuse/core'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MagicStick, Refresh, UploadFilled, WarningFilled } from '@element-plus/icons-vue'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { assetPhotoUrl } from '@/plugin/asset/utils/photo'
import {
  confirmAssetRecognition,
  createAssetRecognition,
  deleteAssetRecognition,
  getAssetRecognitionDetail,
  getAssetRecognitionList,
  getCategoryOptions,
  retryAssetRecognition,
  saveAssetRecognitionDraft
} from '@/plugin/asset/api/asset'

defineOptions({ name: 'AssetRecognition' })

const statusOptions = [
  { value: 'pending', label: '排队中', type: 'info' }, { value: 'processing', label: '识别中', type: 'primary' },
  { value: 'reviewing', label: '待确认', type: 'warning' }, { value: 'completed', label: '已建档', type: 'success' }, { value: 'failed', label: '失败', type: 'danger' },
  { value: 'deleting', label: '待清理', type: 'warning' }
]
const statusMeta = (status) => statusOptions.find((item) => item.value === status) || { label: status || '未知', type: 'info' }
const statusFilter = ref('')
const uploadRef = ref()
const uploadFiles = ref([])
const jobs = ref([])
const categories = ref([])
const page = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const creating = ref(false)
const detailLoading = ref(false)
const saving = ref(false)
const confirming = ref(false)
const retrying = ref(false)
const retryingId = ref(null)
const deletingId = ref(null)
const detailVisible = ref(false)
const activeJob = ref(null)
const draft = ref(emptyDraft())
const draftFormRef = ref()
let pollingTimer

function emptyDraft () {
  return { assetCode: '', name: '', categoryId: undefined, brand: '', model: '', serialNumber: '', specifications: '', productionDate: '', quantity: 1, unit: '件', unitPrice: 0, currentValue: 0, supplier: '', purchaseDate: '', warrantyEndDate: '', recommendedWarrantyMonths: 0, photos: [], remarks: '' }
}
const { width: viewportWidth } = useWindowSize()
const drawerSize = computed(() => (viewportWidth.value <= 1200 ? '96%' : '1120px'))
const originalValue = computed(() => Number(draft.value.quantity || 0) * Number(draft.value.unitPrice || 0))
const draftRules = { assetCode: [{ required: true, message: '请输入资产编号', trigger: 'blur' }], name: [{ required: true, message: '请输入资产名称', trigger: 'blur' }], categoryId: [{ required: true, message: '请选择资产分类', trigger: 'change' }], quantity: [{ required: true, message: '数量必须大于 0', trigger: 'change' }] }
const dateText = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'
const photoUrl = (photo) => assetPhotoUrl(photo, import.meta.env.VITE_BASE_API)
const confidenceLabels = { name: '资产名称', brand: '品牌', model: '型号', serialNumber: '序列号', specifications: '规格参数', productionDate: '生产日期', recommendedCategoryCode: '推荐分类', recommendedUnit: '计量单位', recommendedWarrantyMonths: '质保月数' }
const confidenceLabel = (field) => confidenceLabels[field] || field
const confidenceEntries = computed(() => Object.entries(activeJob.value?.fieldConfidences || {}).map(([field, value]) => ({ field, percentage: Math.round(Number(value || 0) * 100) })).sort((left, right) => left.percentage - right.percentage))

const loadCategories = async () => { const response = await getCategoryOptions(); if (response?.code === 0) categories.value = response.data || [] }
const loadJobs = async () => {
  loading.value = true
  try {
    const response = await getAssetRecognitionList({ page: page.value, pageSize, status: statusFilter.value })
    if (response?.code === 0) { jobs.value = response.data?.list || []; total.value = Number(response.data?.total || 0) }
  } finally { loading.value = false }
}
const beforeUpload = (file) => {
  if (!['image/jpeg', 'image/png', 'image/webp', 'image/gif'].includes(file.type)) { ElMessage.error('仅支持 JPG、PNG、WebP、GIF 图片'); return false }
  if (file.size > 10 * 1024 * 1024) { ElMessage.error('单张图片不能超过 10MB'); return false }
  return true
}
const onExceed = () => ElMessage.warning('每个识别任务最多上传 6 张图片')
const clearUpload = () => { uploadFiles.value = []; uploadRef.value?.clearFiles() }
const createJob = async () => {
  const files = uploadFiles.value.map((item) => item.raw).filter(Boolean)
  if (!files.length) return
  creating.value = true
  try {
    const response = await createAssetRecognition(files)
    if (response?.code === 0) { ElMessage.success('识别任务已创建'); clearUpload(); await loadJobs() }
  } finally { creating.value = false }
}
const normalizeJob = (job) => {
  activeJob.value = job
  draft.value = { ...emptyDraft(), ...(job.draft || {}) }
}
const openJob = async (row) => {
  detailVisible.value = true
  detailLoading.value = true
  try {
    const response = await getAssetRecognitionDetail({ id: row.ID })
    if (response?.code === 0) { normalizeJob(response.data); startPolling() }
  } finally { detailLoading.value = false }
}
const refreshActiveJob = async () => {
  if (!activeJob.value?.ID) return
  const response = await getAssetRecognitionDetail({ id: activeJob.value.ID })
  if (response?.code === 0) { normalizeJob(response.data); if (!['pending', 'processing'].includes(response.data.status)) stopPolling() }
}
const startPolling = () => { stopPolling(); if (['pending', 'processing'].includes(activeJob.value?.status)) pollingTimer = window.setInterval(refreshActiveJob, 3000) }
const stopPolling = () => { if (pollingTimer) { window.clearInterval(pollingTimer); pollingTimer = undefined } }
const saveDraft = async () => {
  const valid = await draftFormRef.value?.validate().catch(() => false)
  if (!valid || !activeJob.value) return
  saving.value = true
  try {
    const response = await saveAssetRecognitionDraft({ id: activeJob.value.ID, draft: draft.value })
    if (response?.code === 0) { ElMessage.success('草稿已保存'); normalizeJob(response.data); await loadJobs() }
  } finally { saving.value = false }
}
const confirmJob = async () => {
  if (confirming.value) return
  confirming.value = true
  try {
    const valid = await draftFormRef.value?.validate().catch(() => false)
    if (!valid || !activeJob.value) return
    const confirmed = await ElMessageBox.confirm('确认后将创建正式资产，任务与资产关系不可撤销。', '确认建档', { type: 'warning', confirmButtonText: '确认建档', cancelButtonText: '再检查' }).catch(() => false)
    if (!confirmed) return
    const saved = await saveAssetRecognitionDraft({ id: activeJob.value.ID, draft: draft.value })
    if (saved?.code !== 0) return
    const response = await confirmAssetRecognition({ id: activeJob.value.ID })
    if (response?.code === 0) { ElMessage.success('资产建档成功'); detailVisible.value = false; stopPolling(); await loadJobs() }
  } finally { confirming.value = false }
}
const retryJob = async (row) => {
  if (!row?.ID || retryingId.value !== null) return
  retryingId.value = row.ID
  retrying.value = true
  try {
    const response = await retryAssetRecognition({ id: row.ID })
    if (response?.code === 0) { ElMessage.success('任务已重新排队'); await loadJobs(); if (activeJob.value?.ID === row.ID) { await refreshActiveJob(); startPolling() } }
  } finally {
    retrying.value = false
    retryingId.value = null
  }
}
const deleteJob = async (row) => {
  if (!row?.ID || deletingId.value !== null) return
  deletingId.value = row.ID
  try {
    const confirmed = await ElMessageBox.confirm('删除后会清理本次任务图片，不能恢复。', '删除任务', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' }).catch(() => false)
    if (!confirmed) return
    const response = await deleteAssetRecognition({ id: row.ID })
    if (response?.code === 0) { ElMessage.success('识别任务已删除'); if (activeJob.value?.ID === row.ID) { detailVisible.value = false; stopPolling() } await loadJobs() }
  } finally {
    deletingId.value = null
  }
}

onMounted(async () => { await loadCategories(); await loadJobs() })
watch(detailVisible, (visible) => { if (!visible) stopPolling() })
onBeforeUnmount(stopPolling)
</script>

<style scoped lang="scss">
.asset-recognition-page { min-width: 0; }
.upload-panel, .task-panel { padding: 16px; margin-bottom: 14px; }
.section-heading, .drawer-heading, .upload-actions, .form-title { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.section-heading > div, .drawer-heading { min-width: 0; }
.section-heading { margin-bottom: 14px; }.section-heading h2 { margin: 0 0 4px; font-size: 16px; }.section-heading p, .form-title small { margin: 0; color: var(--na-muted-foreground); font-size: 12px; }
.recognition-upload :deep(.el-upload-dragger) { padding: 22px 14px; border-color: var(--na-border); background: var(--na-muted); }.upload-icon { margin-bottom: 8px; color: var(--na-primary); font-size: 30px; }.recognition-upload em { color: var(--na-primary); font-style: normal; }.recognition-upload :deep(.el-upload__tip) { color: var(--na-muted-foreground); }.upload-actions { justify-content: flex-end; margin-top: 14px; }
.status-filter { width: 140px; flex: 0 0 auto; }.task-panel { min-width: 0; overflow: hidden; }.task-panel :deep(.el-pagination) { justify-content: flex-end; margin-top: 14px; }.task-identity, .task-result { display: flex; flex-direction: column; gap: 4px; min-width: 0; }.task-identity span, .task-result span { color: var(--na-muted-foreground); font-size: 12px; }.task-result strong, .task-result span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.drawer-heading { flex-wrap: wrap; justify-content: flex-start; }.drawer-heading strong { font-size: 18px; }.detail-alert { margin-bottom: 14px; }.recognition-layout { display: grid; grid-template-columns: minmax(260px, .82fr) minmax(430px, 1.4fr); gap: 22px; }.evidence-column, .draft-column { min-width: 0; }.evidence-column h3, .duplicate-list h3, .confidence-list h3 { margin: 0 0 12px; font-size: 14px; }.photo-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; }.evidence-photo { width: 100%; aspect-ratio: 1; border: 1px solid var(--na-border); border-radius: 7px; background: var(--na-muted); }.photo-error { display: grid; height: 100%; place-items: center; color: var(--na-muted-foreground); font-size: 12px; }.raw-text { margin-top: 16px; }.raw-text > span { color: var(--na-muted-foreground); font-size: 12px; }.raw-text pre { max-height: 180px; overflow: auto; margin: 7px 0 0; padding: 10px; border: 1px solid var(--na-border); border-radius: 6px; background: var(--na-muted); color: var(--na-foreground); font: 12px/1.55 ui-monospace, SFMono-Regular, Menlo, monospace; white-space: pre-wrap; overflow-wrap: anywhere; }.confidence-list { margin-top: 18px; }.confidence-item { display: grid; grid-template-columns: 90px minmax(0, 1fr); align-items: center; gap: 8px; margin-bottom: 8px; }.confidence-item > span { color: var(--na-muted-foreground); font-size: 12px; }.warning-list { display: grid; gap: 7px; margin-top: 14px; }.warning-item { display: flex; gap: 7px; align-items: flex-start; padding: 8px 9px; border-radius: 6px; background: var(--na-warning-soft); color: var(--na-warning); font-size: 12px; }.warning-item.error { background: var(--na-danger-soft); color: var(--na-danger); }.duplicate-list { margin-top: 18px; }.duplicate-item { display: flex; flex-direction: column; gap: 4px; padding: 9px 10px; border: 1px solid var(--na-danger); border-radius: 6px; background: var(--na-danger-soft); }.duplicate-item + .duplicate-item { margin-top: 7px; }.duplicate-item span { color: var(--na-muted-foreground); font-size: 12px; }.form-title { flex-wrap: wrap; justify-content: flex-start; margin-bottom: 14px; }.form-title h3 { margin: 0; font-size: 16px; }.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0 16px; }.form-grid :deep(.el-select), .form-grid :deep(.el-date-editor), .form-grid :deep(.el-input-number) { width: 100%; }.valuation-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.drawer-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 9px; }.drawer-actions > .el-button { margin-left: 0; }
@media (max-width: 900px) { .recognition-layout { grid-template-columns: 1fr; }.evidence-column { order: 2; }.draft-column { order: 1; } }
@media (max-width: 600px) { .section-heading { align-items: flex-start; flex-direction: column; }.status-filter { width: 100%; }.form-grid, .valuation-grid { grid-template-columns: 1fr; }.upload-panel, .task-panel { padding: 12px; }.upload-actions { align-items: stretch; flex-direction: column; }.upload-actions > .el-button { width: 100%; margin-left: 0; }.drawer-actions { justify-content: stretch; }.drawer-actions > .el-button { flex: 1 1 calc(50% - 5px); } }
</style>
