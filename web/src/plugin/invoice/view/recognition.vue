<template>
  <main class="na-page na-page--list invoice-recognition">
    <AppPageHeader title-id="invoice-recognition-title" title="发票识别" description="上传图片或 PDF 发票，系统自动提取字段并给出可解释的分类建议。">
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="loadQueue">刷新队列</el-button>
      </template>
    </AppPageHeader>

    <section class="upload-band">
      <el-upload
        ref="uploadRef"
        drag
        multiple
        :limit="5"
        :auto-upload="false"
        :show-file-list="false"
        :on-change="handleUploadSelection"
        :on-exceed="handleExceed"
        accept="image/jpeg,image/png,application/pdf,.pdf"
        :disabled="uploading"
      >
        <el-icon class="upload-icon"><UploadFilled /></el-icon>
        <div class="upload-copy">
          <strong>{{ uploadHeadline }}</strong>
          <span>一次最多 5 个 JPG / PNG / PDF，单个不超过 10MB，PDF 最多 10 页；原文件仅授权用户可访问。</span>
        </div>
      </el-upload>
      <div class="upload-process" aria-live="polite">
        <div><span>1</span><strong>安全上传</strong><small>校验类型与 SHA256 去重</small></div>
        <i />
        <div><span>2</span><strong>自动识别</strong><small>二维码与 OCR Adapter</small></div>
        <i />
        <div><span>3</span><strong>人工确认</strong><small>确认后才进入统计</small></div>
      </div>
    </section>

    <section class="na-panel queue-panel">
      <div class="na-panel-header queue-heading">
        <div><h2>待处理队列</h2><p>识别中的任务会自动刷新，失败任务可手动重试。</p></div>
        <div class="queue-counts">
          <span>队列共 {{ queueTotal }} 条</span>
          <span><i class="warning" />本页待核对 {{ counts.pending }}</span>
          <span><i class="primary" />本页处理中 {{ counts.processing }}</span>
          <span><i class="danger" />本页失败 {{ counts.failed }}</span>
        </div>
      </div>

      <div v-if="queueError && queueLoaded" class="queue-warning" role="alert">
        <span>刷新失败，当前仍显示上一次成功数据：{{ queueError }}</span>
        <el-button text :icon="Refresh" @click="loadQueue">重试</el-button>
      </div>
      <el-skeleton v-if="loading && !queueLoaded" :rows="6" animated />
      <el-result v-else-if="queueError && !queueLoaded" icon="error" title="识别队列加载失败" :sub-title="queueError">
        <template #extra><el-button type="primary" :icon="Refresh" @click="loadQueue">重新加载</el-button></template>
      </el-result>
      <div v-else-if="queue.length" class="queue-list">
        <button v-for="item in queue" :key="item.ID" type="button" class="queue-row" @click="openReview(item)">
          <div class="file-mark" aria-hidden="true"><el-icon><Document /></el-icon></div>
          <div class="queue-identity">
            <strong>{{ item.sellerName || item.fileName }}</strong>
            <span>{{ item.invoiceNumber || '发票号码待识别' }} · 上传于 {{ dateText(item.CreatedAt) }}</span>
            <span class="queue-mobile-meta">{{ item.category?.name || item.suggestedCategory?.name || '等待人工分类' }} · 置信度 {{ confidence(item) }}</span>
          </div>
          <div class="queue-classification">
            <small>分类建议</small>
            <span>{{ item.category?.name || item.suggestedCategory?.name || '等待人工判断' }}</span>
          </div>
          <div class="queue-confidence">
            <small>识别置信度</small>
            <strong>{{ confidence(item) }}</strong>
          </div>
          <InvoiceStatusTag :status="item.status" />
          <el-icon class="open-icon"><ArrowRight /></el-icon>
        </button>
      </div>
      <AppEmptyState
        v-else
        compact
        title="识别队列为空"
        description="当前没有上传中、识别中、待核对或识别失败的发票。"
        :highlights="['队列已处理完毕']"
      />
      <div v-if="queueTotal > 10" class="na-pagination queue-pagination">
        <el-pagination
          v-model:current-page="queuePage"
          v-model:page-size="queuePageSize"
          :page-sizes="[10, 20, 50]"
          :total="queueTotal"
          layout="total, sizes, prev, pager, next"
          @change="loadQueue"
          @size-change="resetQueuePage"
        />
      </div>
    </section>

    <InvoiceReviewDrawer v-model="reviewVisible" :invoice-id="selectedId" @saved="loadQueue" @confirmed="loadQueue" />
  </main>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ArrowRight, Document, Refresh, UploadFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import InvoiceReviewDrawer from '@/plugin/invoice/components/InvoiceReviewDrawer.vue'
import InvoiceStatusTag from '@/plugin/invoice/components/InvoiceStatusTag.vue'
import { getInvoiceList, uploadInvoices } from '@/plugin/invoice/api/invoice'
import { invoiceDateText } from '@/plugin/invoice/utils/invoice'
import { invoiceUploadError } from '@/plugin/invoice/utils/upload'

defineOptions({ name: 'InvoiceRecognition' })

const loading = ref(false)
const uploadRef = ref()
const uploading = ref(false)
const uploadTotal = ref(0)
const queue = ref([])
const queueTotal = ref(0)
const queuePage = ref(1)
const queuePageSize = ref(10)
const queueLoaded = ref(false)
const queueError = ref('')
const reviewVisible = ref(false)
const selectedId = ref(0)
let refreshTimer
let pendingUploadFiles = []
let uploadScheduled = false
let queueLoadPromise
let queueRefreshPending = false

const dateText = invoiceDateText
const uploadHeadline = computed(() => uploading.value
  ? `正在上传 ${uploadTotal.value} 份发票并创建识别任务`
  : '拖入发票文件，或点击选择')
const counts = computed(() => ({
  pending: queue.value.filter((item) => item.status === 'pending_review').length,
  processing: queue.value.filter((item) => ['uploaded', 'recognizing'].includes(item.status)).length,
  failed: queue.value.filter((item) => item.status === 'recognition_failed').length
}))

const confidence = (item) => {
  const value = Number(item.recognitionConfidence || 0)
  return value > 0 ? `${Math.round(value * 100)}%` : '—'
}

const runQueueRefreshes = async () => {
  loading.value = true
  try {
    do {
      queueRefreshPending = false
      queueError.value = ''
      try {
        const res = await getInvoiceList({ page: queuePage.value, pageSize: queuePageSize.value, excludeStatus: 'confirmed' })
        if (res.code === 0) {
          const nextQueue = res.data?.list || []
          const nextTotal = Number(res.data?.total || 0)
          if (!nextQueue.length && nextTotal > 0 && queuePage.value > 1) {
            queuePage.value--
            queueRefreshPending = true
            continue
          }
          queue.value = nextQueue
          queueTotal.value = nextTotal
          queueLoaded.value = true
        } else {
          queueError.value = res.msg || '无法读取识别队列，请稍后重试'
        }
      } catch (error) {
        queueError.value = error?.message || '无法读取识别队列，请稍后重试'
      }
    } while (queueRefreshPending)
  } finally {
    loading.value = false
  }
}

const loadQueue = () => {
  queueRefreshPending = true
  if (!queueLoadPromise) {
    queueLoadPromise = runQueueRefreshes().finally(() => {
      queueLoadPromise = undefined
    })
  }
  return queueLoadPromise
}

const resetQueuePage = () => {
  queuePage.value = 1
}

const validateUploadFile = (file) => {
  const error = invoiceUploadError(file)
  if (error) {
    ElMessage.error(error)
    return false
  }
  return true
}

const handleExceed = () => {
  ElMessage.warning('一次最多选择 5 个发票文件')
}

const handleUploadSelection = (_file, fileList) => {
  pendingUploadFiles = fileList.map((item) => item.raw).filter(Boolean)
  if (uploadScheduled || uploading.value) return
  uploadScheduled = true
  queueMicrotask(() => {
    uploadScheduled = false
    uploadSelectedFiles()
  })
}

const uploadSelectedFiles = async () => {
  const files = pendingUploadFiles.slice(0, 5).filter(validateUploadFile)
  pendingUploadFiles = []
  uploadRef.value?.clearFiles()
  if (!files.length || uploading.value) return
  uploadTotal.value = files.length
  uploading.value = true
  try {
    const res = await uploadInvoices(files)
    const succeeded = res.data?.succeeded || []
    const failed = res.data?.failed || []
    if (succeeded.length) {
      selectedId.value = Number(succeeded[succeeded.length - 1]?.ID || 0)
      queuePage.value = 1
      await loadQueue()
    }
    if (failed.length === 0 && succeeded.length) {
      ElMessage.success(`${succeeded.length} 份发票已进入识别队列`)
    } else if (succeeded.length) {
      const reason = failed[0]?.message ? `；首个失败原因：${failed[0].message}` : ''
      ElMessage.warning(`上传完成：成功 ${succeeded.length} 份，失败 ${failed.length} 份${reason}`)
    }
  } finally {
    uploading.value = false
  }
}

const openReview = (item) => {
  selectedId.value = Number(item.ID)
  reviewVisible.value = true
}

onMounted(() => {
  loadQueue()
  refreshTimer = window.setInterval(() => {
    if (document.visibilityState === 'visible' && !loading.value && queue.value.some((item) => ['uploaded', 'recognizing'].includes(item.status))) loadQueue()
  }, 5000)
  document.addEventListener('visibilitychange', handleVisibilityChange)
})
const handleVisibilityChange = () => {
  if (document.visibilityState === 'visible' && queue.value.some((item) => ['uploaded', 'recognizing'].includes(item.status))) loadQueue()
}
onUnmounted(() => {
  window.clearInterval(refreshTimer)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped lang="scss">
.upload-band { display: grid; min-width: 0; grid-template-columns: minmax(360px, 1.1fr) minmax(420px, .9fr); align-items: stretch; gap: 14px; margin-bottom: 14px; }
.upload-band :deep(.el-upload), .upload-band :deep(.el-upload-dragger) { width: 100%; height: 100%; }
.upload-band :deep(.el-upload-dragger) { display: flex; min-height: 150px; align-items: center; justify-content: center; gap: 16px; padding: 22px; border-radius: 12px; text-align: left; }
.upload-icon { flex: 0 0 auto; color: var(--na-primary); font-size: 2rem; }
.upload-copy { display: flex; min-width: 0; flex-direction: column; gap: 6px; }
.upload-copy strong { color: var(--na-foreground); font-size: .9375rem; font-weight: 650; }
.upload-copy span { color: var(--na-muted-foreground); font-size: .75rem; line-height: 1.55; }
.upload-process { display: grid; min-width: 0; grid-template-columns: 1fr 24px 1fr 24px 1fr; align-items: center; padding: 18px; border: 1px solid var(--na-border); border-radius: 12px; background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.upload-process > div { display: grid; justify-items: center; gap: 5px; min-width: 0; text-align: center; }
.upload-process div > span { display: grid; width: 28px; height: 28px; place-items: center; border-radius: 50%; background: var(--na-primary-soft); color: var(--na-primary); font-size: .75rem; font-weight: 700; }
.upload-process strong { font-size: .75rem; }
.upload-process small { color: var(--na-muted-foreground); font-size: .625rem; line-height: 1.35; }
.upload-process > i { height: 1px; background: var(--na-border); }
.queue-panel { min-height: 430px; overflow: hidden; }
.queue-heading > div:first-child { min-width: 0; }
.queue-heading h2 { margin: 0; font-size: .9375rem; font-weight: 650; }
.queue-heading p { margin: 3px 0 0; color: var(--na-muted-foreground); font-size: .75rem; }
.queue-counts { display: flex; flex-wrap: wrap; gap: 12px; }
.queue-counts span { display: inline-flex; align-items: center; gap: 6px; color: var(--na-muted-foreground); font-size: .6875rem; }
.queue-counts i { width: 6px; height: 6px; border-radius: 50%; }
.queue-counts .warning { background: var(--na-warning); }
.queue-counts .primary { background: var(--na-primary); }
.queue-counts .danger { background: var(--na-danger); }
.queue-warning { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 12px; padding: 9px 14px; border-bottom: 1px solid var(--na-border); background: var(--na-warning-soft); font-size: .75rem; }
.queue-warning span { min-width: 0; overflow-wrap: anywhere; }
.queue-list { display: grid; }
.queue-pagination { padding: 14px; }
.queue-row { display: grid; min-width: 0; min-height: 54px; grid-template-columns: 32px minmax(180px, 1.5fr) minmax(130px, 1fr) 88px 86px 20px; align-items: center; gap: 10px; padding: 6px 14px; border: 0; border-bottom: 1px solid var(--na-border); background: var(--na-card); color: var(--na-foreground); text-align: left; transition: background-color 160ms ease; }
.queue-row:hover { background: var(--na-table-hover); }
.queue-row:focus-visible { position: relative; z-index: 1; outline: 2px solid var(--na-primary); outline-offset: -3px; }
.file-mark { display: grid; width: 30px; height: 30px; place-items: center; border-radius: 7px; background: var(--na-primary-soft); color: var(--na-primary); font-size: .875rem; }
.queue-identity, .queue-classification, .queue-confidence { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.queue-identity strong, .queue-classification span { overflow: hidden; font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.queue-identity span, .queue-classification small, .queue-confidence small { overflow: hidden; color: var(--na-muted-foreground); font-size: .625rem; text-overflow: ellipsis; white-space: nowrap; }
.queue-identity .queue-mobile-meta { display: none; }
.queue-confidence strong { font-size: .6875rem; font-variant-numeric: tabular-nums; }
.open-icon { color: var(--na-muted-foreground); }
.empty-note { color: var(--na-muted-foreground); font-size: .75rem; }

@media (max-width: 1100px) {
  .upload-band { grid-template-columns: 1fr; }
  .queue-row { grid-template-columns: 32px minmax(0, 1fr) 88px 86px 20px; }
  .queue-classification { display: none; }
  .queue-identity .queue-mobile-meta { display: block; }
}
@media (max-width: 720px) {
  .upload-band { grid-template-columns: minmax(0, 1fr); }
  .upload-process { grid-template-columns: 1fr; gap: 8px; }
  .upload-process > i { width: 1px; height: 12px; justify-self: center; }
  .queue-heading { align-items: flex-start; flex-direction: column; }
  .queue-row { grid-template-columns: 34px minmax(0, 1fr) auto; gap: 9px; padding: 10px 12px; }
  .queue-confidence, .open-icon { display: none; }
  .queue-row :deep(.invoice-status-tag) { grid-column: 3; grid-row: 1; }
}
</style>
