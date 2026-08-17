<template>
  <main v-loading="loading" class="na-page na-page--list pending-tasks-page">
    <AppPageHeader
      title-id="pending-tasks-title"
      title="待处理事项"
      description="集中查看尚未提交的资产业务单，以及需要人工处理的发票。"
    >
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="loadPendingTasks">刷新</el-button>
        <el-button :icon="ArrowLeft" @click="emit('back')">返回驾驶舱</el-button>
      </template>
    </AppPageHeader>

    <section class="pending-summary" aria-label="待处理事项汇总">
      <article v-if="canAccessAssetOperations" class="summary-item summary-item--asset">
        <span>资产业务草稿</span>
        <strong>{{ formatNumber(assetTotal) }}<small>项</small></strong>
        <p>按业务类型进入并完成提交</p>
      </article>
      <article v-if="canAccessInvoices" class="summary-item summary-item--invoice">
        <span>待核对发票</span>
        <strong>{{ formatNumber(pendingInvoiceTotal) }}<small>张</small></strong>
        <p>识别完成后等待人工确认</p>
      </article>
      <article v-if="canAccessInvoices" class="summary-item summary-item--failed">
        <span>识别失败发票</span>
        <strong>{{ formatNumber(failedInvoiceTotal) }}<small>张</small></strong>
        <p>重新处理或补充信息后核对</p>
      </article>
    </section>

    <section v-if="canAccessAssetOperations || canAccessInvoices" class="pending-workspace">
      <article v-if="canAccessAssetOperations" class="na-panel pending-panel">
        <header class="na-panel-header pending-panel-heading">
          <div>
            <span>资产管理</span>
            <h2>待提交业务单</h2>
          </div>
          <small>共 {{ formatNumber(assetTotal) }} 项草稿</small>
        </header>

        <div v-if="assetError" class="pending-error" role="alert">{{ assetError }}</div>
        <div v-else-if="assetOrders.length" class="pending-list">
          <button
            v-for="order in assetOrders"
            :key="order.ID"
            type="button"
            class="pending-row"
            :aria-label="`处理${operationMeta(order.type).label}单 ${order.orderNo}`"
            @click="openAssetOperation(order)"
          >
            <span class="task-kind task-kind--asset">{{ operationMeta(order.type).label }}</span>
            <span class="task-main">
              <strong>{{ order.orderNo }}</strong>
              <small>{{ firstAssetName(order) }} · {{ formatDateText(order.businessDate) }}</small>
            </span>
            <span class="task-detail">{{ itemSummary(order) }}</span>
            <el-tag type="info" effect="light" size="small">草稿</el-tag>
            <el-icon class="task-arrow"><ArrowRight /></el-icon>
          </button>
        </div>
        <AppEmptyState
          v-else
          compact
          title="暂无待提交资产业务单"
          description="当前没有停留在草稿状态的入库、领用、调拨、归还、维修或报废单。"
          :highlights="['资产流程已清空']"
        >
          <template #actions><el-button type="primary" @click="openAssetDraft">新建业务单</el-button></template>
        </AppEmptyState>
      </article>

      <article v-if="canAccessInvoices" class="na-panel pending-panel">
        <header class="na-panel-header pending-panel-heading">
          <div>
            <span>流水管理</span>
            <h2>待核对发票</h2>
          </div>
          <el-button text :icon="ArrowRight" @click="openInvoiceQueue">进入识别队列</el-button>
        </header>

        <div v-if="invoiceError" class="pending-error" role="alert">{{ invoiceError }}</div>
        <div v-else-if="invoiceTasks.length" class="pending-list">
          <button
            v-for="invoice in invoiceTasks"
            :key="invoice.ID"
            type="button"
            class="pending-row"
            :aria-label="`处理发票 ${invoice.invoiceNumber || invoice.fileName}`"
            @click="openInvoiceQueue"
          >
            <span class="task-kind" :class="invoice.status === 'recognition_failed' ? 'task-kind--failed' : 'task-kind--invoice'">
              {{ invoice.status === 'recognition_failed' ? '识别失败' : '待核对' }}
            </span>
            <span class="task-main">
              <strong>{{ invoice.sellerName || invoice.fileName }}</strong>
              <small>{{ invoice.invoiceNumber || '发票号码待识别' }} · {{ invoiceDateText(invoice.issueDate || invoice.CreatedAt) }}</small>
            </span>
            <span class="task-detail">{{ invoice.category?.name || invoice.suggestedCategory?.name || '等待人工分类' }}</span>
            <InvoiceStatusTag :status="invoice.status" />
            <el-icon class="task-arrow"><ArrowRight /></el-icon>
          </button>
        </div>
        <AppEmptyState
          v-else
          compact
          title="暂无待处理发票"
          description="当前没有待核对或识别失败的发票任务。"
          :highlights="['发票队列已清空']"
        >
          <template #actions><el-button type="primary" @click="openInvoiceQueue">进入识别队列</el-button></template>
        </AppEmptyState>
      </article>
    </section>

    <el-result
      v-else
      icon="warning"
      title="暂无待处理模块权限"
      sub-title="当前账号没有资产业务或发票识别权限。"
    >
      <template #extra><el-button type="primary" @click="emit('back')">返回驾驶舱</el-button></template>
    </el-result>
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, ArrowRight, Refresh } from '@element-plus/icons-vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import InvoiceStatusTag from '@/plugin/invoice/components/InvoiceStatusTag.vue'
import { getInvoiceList } from '@/plugin/invoice/api/invoice'
import { getAssetOperationList } from '@/plugin/asset/api/operation'
import { invoiceDateText } from '@/plugin/invoice/utils/invoice'
import { formatDateText, formatNumber } from '@/utils/format'

const emit = defineEmits(['back'])
const router = useRouter()
const loading = ref(false)
const assetOrders = ref([])
const assetTotal = ref(0)
const pendingInvoices = ref([])
const pendingInvoiceTotal = ref(0)
const failedInvoices = ref([])
const failedInvoiceTotal = ref(0)
const assetError = ref('')
const invoiceError = ref('')

const operationRoutes = {
  inbound: 'assetInbound',
  issue: 'assetIssue',
  transfer: 'assetTransfer',
  return: 'assetReturn',
  maintenance: 'assetMaintenance',
  scrap: 'assetScrap'
}
const operationTypes = {
  inbound: { label: '入库', tone: 'asset' },
  issue: { label: '领用', tone: 'asset' },
  transfer: { label: '调拨', tone: 'asset' },
  return: { label: '归还', tone: 'asset' },
  maintenance: { label: '维修', tone: 'asset' },
  scrap: { label: '报废', tone: 'asset' }
}

const canAccessAssetOperations = computed(() => Object.values(operationRoutes).some((name) => router.hasRoute(name)))
const canAccessInvoices = computed(() => router.hasRoute('invoiceRecognition'))
const invoiceTasks = computed(() => [...pendingInvoices.value, ...failedInvoices.value]
  .sort((left, right) => new Date(right.CreatedAt || 0) - new Date(left.CreatedAt || 0)))

const operationMeta = (type) => operationTypes[type] || { label: '业务', tone: 'asset' }
const firstAssetName = (order) => order.items?.[0]?.assetName || '暂无资产明细'
const itemSummary = (order) => {
  const items = order.items || []
  const quantity = items.reduce((total, item) => total + Number(item.quantity || 0), 0)
  return `${items.length} 项档案，共 ${quantity} 件（套）`
}

async function loadPendingTasks() {
  loading.value = true
  assetError.value = ''
  invoiceError.value = ''

  const [assetResult, pendingResult, failedResult] = await Promise.allSettled([
    canAccessAssetOperations.value ? getAssetOperationList({ page: 1, pageSize: 12, status: 'draft' }) : Promise.resolve(null),
    canAccessInvoices.value ? getInvoiceList({ page: 1, pageSize: 12, status: 'pending_review' }) : Promise.resolve(null),
    canAccessInvoices.value ? getInvoiceList({ page: 1, pageSize: 12, status: 'recognition_failed' }) : Promise.resolve(null)
  ])

  if (assetResult.status === 'fulfilled' && assetResult.value?.code === 0) {
    assetOrders.value = assetResult.value.data?.list || []
    assetTotal.value = Number(assetResult.value.data?.total || 0)
  } else if (canAccessAssetOperations.value) {
    assetOrders.value = []
    assetTotal.value = 0
    assetError.value = '待提交资产业务单加载失败，请刷新后重试。'
  }

  const updateInvoiceState = (result, target, total) => {
    if (result.status === 'fulfilled' && result.value?.code === 0) {
      target.value = result.value.data?.list || []
      total.value = Number(result.value.data?.total || 0)
      return true
    }
    target.value = []
    total.value = 0
    return false
  }
  const pendingLoaded = updateInvoiceState(pendingResult, pendingInvoices, pendingInvoiceTotal)
  const failedLoaded = updateInvoiceState(failedResult, failedInvoices, failedInvoiceTotal)
  if (canAccessInvoices.value && (!pendingLoaded || !failedLoaded)) {
    invoiceError.value = '待处理发票加载失败，请刷新后重试。'
  }
  loading.value = false
}

function openAssetOperation(order) {
  const routeName = operationRoutes[order.type]
  if (!routeName || !router.hasRoute(routeName)) return
  router.push({ name: routeName, query: { status: 'draft' } })
}
function openInvoiceQueue() {
  if (router.hasRoute('invoiceRecognition')) router.push({ name: 'invoiceRecognition' })
}
function openAssetDraft() {
  const routeName = Object.values(operationRoutes).find((name) => router.hasRoute(name))
  if (routeName) router.push({ name: routeName })
}

onMounted(loadPendingTasks)
</script>

<style scoped lang="scss">
.pending-tasks-page { min-height: 100%; }
.pending-summary { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--na-space-md); margin-bottom: var(--na-space-lg); }
.summary-item { min-width: 0; padding: 16px 18px; border: 1px solid var(--na-border); border-radius: var(--na-radius); background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.summary-item > span, .summary-item p { display: block; overflow: hidden; color: var(--na-muted-foreground); font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.summary-item strong { display: flex; align-items: baseline; gap: 4px; margin: 7px 0 4px; color: var(--na-foreground); font-size: 1.375rem; font-variant-numeric: tabular-nums; font-weight: 700; }
.summary-item strong small { color: var(--na-muted-foreground); font-size: .75rem; font-weight: 600; }
.summary-item p { margin: 0; font-size: .6875rem; }
.summary-item--asset strong { color: var(--na-primary); }
.summary-item--invoice strong { color: var(--na-warning); }
.summary-item--failed strong { color: var(--na-danger); }
.pending-workspace { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--na-space-md); align-items: start; }
.pending-panel { min-width: 0; overflow: hidden; }
.pending-panel-heading > div { min-width: 0; }
.pending-panel-heading span { display: block; color: var(--na-muted-foreground); font-size: .6875rem; }
.pending-panel-heading h2 { margin: 3px 0 0; color: var(--na-foreground); font-size: .9375rem; font-weight: 600; }
.pending-panel-heading > small { color: var(--na-muted-foreground); font-size: .6875rem; white-space: nowrap; }
.pending-error { padding: 10px 14px; border-bottom: 1px solid var(--na-border); background: var(--na-danger-soft); color: var(--na-danger); font-size: .75rem; }
.pending-list { display: grid; }
.pending-row { display: grid; width: 100%; min-width: 0; min-height: 58px; grid-template-columns: auto minmax(130px, 1.25fr) minmax(106px, .9fr) auto 18px; align-items: center; gap: 10px; padding: 7px 14px; border: 0; border-bottom: 1px solid var(--na-border); background: transparent; color: var(--na-foreground); font: inherit; text-align: left; transition: background-color 160ms ease; }
.pending-row:last-child { border-bottom: 0; }
.pending-row:hover { background: var(--na-table-hover); cursor: pointer; }
.pending-row:focus-visible { position: relative; z-index: 1; outline: 2px solid var(--na-primary); outline-offset: -3px; }
.task-kind { display: inline-grid; min-width: 46px; height: 24px; place-items: center; border-radius: 6px; font-size: .6875rem; font-weight: 600; white-space: nowrap; }
.task-kind--asset { color: var(--na-primary); background: var(--na-primary-soft); }
.task-kind--invoice { color: var(--na-warning); background: var(--na-warning-soft); }
.task-kind--failed { color: var(--na-danger); background: var(--na-danger-soft); }
.task-main { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.task-main strong, .task-main small, .task-detail { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.task-main strong { font-size: .75rem; font-weight: 600; }
.task-main small, .task-detail { color: var(--na-muted-foreground); font-size: .6875rem; }
.task-detail { text-align: right; }
.task-arrow { color: var(--na-muted-foreground); font-size: .75rem; }

@media (max-width: 1060px) {
  .pending-workspace { grid-template-columns: 1fr; }
}
@media (max-width: 720px) {
  .pending-summary { grid-template-columns: 1fr; gap: 8px; }
  .pending-row { grid-template-columns: auto minmax(0, 1fr) auto; gap: 8px; padding: 9px 12px; }
  .task-detail, .pending-row :deep(.invoice-status-tag) { display: none; }
  .pending-row :deep(.el-tag) { display: none; }
  .task-arrow { grid-column: 3; grid-row: 1 / span 2; }
}
</style>
