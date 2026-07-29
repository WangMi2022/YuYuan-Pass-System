<template>
  <main class="na-page na-page--list invoice-recognition">
    <AppPageHeader title-id="invoice-recognition-title" title="发票识别" description="上传 JPG 或 PNG 发票，系统自动提取字段并给出可解释的分类建议。">
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="loadQueue">刷新队列</el-button>
      </template>
    </AppPageHeader>

    <section class="upload-band">
      <el-upload
        drag
        :show-file-list="false"
        :http-request="handleUpload"
        :before-upload="beforeUpload"
        accept="image/jpeg,image/png"
        :disabled="uploading"
      >
        <el-icon class="upload-icon"><UploadFilled /></el-icon>
        <div class="upload-copy">
          <strong>{{ uploading ? '正在保存发票并创建识别任务' : '拖入发票图片，或点击选择文件' }}</strong>
          <span>JPG / PNG，单张不超过 10MB；原图保存在私有 RustFS，仅授权用户可访问。</span>
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
          <span><i class="warning" />待核对 {{ counts.pending }}</span>
          <span><i class="primary" />处理中 {{ counts.processing }}</span>
          <span><i class="danger" />失败 {{ counts.failed }}</span>
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
      <el-empty v-else description="当前没有待处理发票">
        <span class="empty-note">上传发票后，识别任务会自动出现在这里。</span>
      </el-empty>
    </section>

    <InvoiceReviewDrawer v-model="reviewVisible" :invoice-id="selectedId" @saved="loadQueue" @confirmed="loadQueue" />
  </main>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ArrowRight, Document, Refresh, UploadFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import InvoiceReviewDrawer from '@/plugin/invoice/components/InvoiceReviewDrawer.vue'
import InvoiceStatusTag from '@/plugin/invoice/components/InvoiceStatusTag.vue'
import { getInvoiceList, uploadInvoice } from '@/plugin/invoice/api/invoice'
import { invoiceDateText } from '@/plugin/invoice/utils/invoice'

defineOptions({ name: 'InvoiceRecognition' })

const loading = ref(false)
const uploading = ref(false)
const queue = ref([])
const queueLoaded = ref(false)
const queueError = ref('')
const reviewVisible = ref(false)
const selectedId = ref(0)
let refreshTimer

const dateText = invoiceDateText
const counts = computed(() => ({
  pending: queue.value.filter((item) => item.status === 'pending_review').length,
  processing: queue.value.filter((item) => ['uploaded', 'recognizing'].includes(item.status)).length,
  failed: queue.value.filter((item) => item.status === 'recognition_failed').length
}))

const confidence = (item) => {
  const value = Number(item.recognitionConfidence || 0)
  return value > 0 ? `${Math.round(value * 100)}%` : '—'
}

const loadQueue = async () => {
  if (loading.value) return
  loading.value = true
  queueError.value = ''
  try {
    const res = await getInvoiceList({ page: 1, pageSize: 50 })
    if (res.code === 0) {
      queue.value = (res.data?.list || []).filter((item) => item.status !== 'confirmed')
      queueLoaded.value = true
    } else {
      queueError.value = res.msg || '无法读取识别队列，请稍后重试'
    }
  } catch (error) {
    queueError.value = error?.message || '无法读取识别队列，请稍后重试'
  } finally {
    loading.value = false
  }
}

const beforeUpload = (file) => {
  if (!['image/jpeg', 'image/png'].includes(file.type)) {
    ElMessage.error('仅支持 JPG、PNG 发票图片')
    return false
  }
  if (file.size <= 0 || file.size > 10 * 1024 * 1024) {
    ElMessage.error('发票图片大小必须在 10MB 以内')
    return false
  }
  return true
}

const handleUpload = async (options) => {
  uploading.value = true
  try {
    const res = await uploadInvoice(options.file)
    if (res.code === 0) {
      options.onSuccess(res.data)
      ElMessage.success('发票已进入识别队列')
      await loadQueue()
      selectedId.value = Number(res.data?.ID || 0)
    } else {
      options.onError(new Error(res.msg || '上传失败'))
    }
  } catch (error) {
    options.onError(error)
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
.queue-row { display: grid; min-width: 0; min-height: 68px; grid-template-columns: 38px minmax(180px, 1.5fr) minmax(130px, 1fr) 96px 90px 24px; align-items: center; gap: 12px; padding: 8px 16px; border: 0; border-bottom: 1px solid var(--na-border); background: var(--na-card); color: var(--na-foreground); text-align: left; transition: background-color 160ms ease; }
.queue-row:hover { background: var(--na-table-hover); }
.queue-row:focus-visible { position: relative; z-index: 1; outline: 2px solid var(--na-primary); outline-offset: -3px; }
.file-mark { display: grid; width: 34px; height: 34px; place-items: center; border-radius: 8px; background: var(--na-primary-soft); color: var(--na-primary); }
.queue-identity, .queue-classification, .queue-confidence { display: flex; min-width: 0; flex-direction: column; gap: 4px; }
.queue-identity strong, .queue-classification span { overflow: hidden; font-size: .8125rem; text-overflow: ellipsis; white-space: nowrap; }
.queue-identity span, .queue-classification small, .queue-confidence small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.queue-identity .queue-mobile-meta { display: none; }
.queue-confidence strong { font-size: .75rem; font-variant-numeric: tabular-nums; }
.open-icon { color: var(--na-muted-foreground); }
.empty-note { color: var(--na-muted-foreground); font-size: .75rem; }

@media (max-width: 1100px) {
  .upload-band { grid-template-columns: 1fr; }
  .queue-row { grid-template-columns: 38px minmax(0, 1fr) 100px 90px 24px; }
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
