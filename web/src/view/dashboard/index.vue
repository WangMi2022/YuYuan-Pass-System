<template>
  <main v-loading="loading" class="dashboard-page">
    <AppPageHeader
      title-id="dashboard-title"
      title="首页驾驶舱"
      description="汇总当前权限范围内的资产、流水、日程与运行状态。"
    >
      <template #actions>
        <span class="updated-at">更新于 {{ updatedAt || '—' }}</span>
        <el-button :icon="Refresh" :loading="loading" @click="loadDashboard">刷新</el-button>
        <el-button v-if="access.assets" type="primary" :icon="Plus" @click="go('assetInventory')">登记资产</el-button>
        <el-button v-else-if="access.invoices" type="primary" :icon="Tickets" @click="go('invoiceRecognition')">上传发票</el-button>
      </template>
    </AppPageHeader>

    <section class="workbench-band" aria-labelledby="workbench-title">
      <div class="workbench-copy">
        <p class="current-date">{{ currentDateText }}</p>
        <h2 id="workbench-title">{{ greeting }}，{{ userStore.userInfo.nickName || userStore.userInfo.userName || '用户' }}</h2>
        <p>{{ overviewText }}</p>
        <div class="quick-actions" aria-label="常用操作">
          <el-button v-if="access.invoices" text :icon="DocumentChecked" @click="go('invoiceLedger')">发票台账</el-button>
          <el-button v-if="access.calendar" text :icon="Calendar" @click="go('workSchedule')">日程总览</el-button>
          <el-button v-if="access.audit" text :icon="Clock" @click="go('operation')">操作历史</el-button>
        </div>
      </div>

      <button
        v-if="access.monitor && moduleLoaded.monitor"
        type="button"
        class="runtime-summary"
        aria-label="查看服务器负载"
        @click="go('state')"
      >
        <span class="runtime-heading" :class="`health-${systemHealth.tone}`"><i />{{ systemHealth.label }}</span>
        <dl>
          <div><dt>CPU</dt><dd>{{ percent(systemUsage.cpu) }}</dd></div>
          <div><dt>内存</dt><dd>{{ percent(systemUsage.ram) }}</dd></div>
          <div><dt>磁盘</dt><dd>{{ percent(systemUsage.disk) }}</dd></div>
        </dl>
        <small>最近采集 {{ serverCollectedAt }}</small>
      </button>
    </section>

    <section v-if="metrics.length" class="metric-band" aria-label="核心业务指标">
      <article v-for="metric in metrics" :key="metric.label" class="metric-item">
        <span>{{ metric.label }}</span>
        <strong>{{ metric.value }}</strong>
        <small :class="metric.tone">{{ metric.hint }}</small>
      </article>
    </section>

    <section class="dashboard-workspace">
      <div class="business-column">
        <article v-if="access.assets" class="na-panel dashboard-panel asset-panel">
          <header class="na-panel-header panel-heading">
            <div>
              <span>资产管理</span>
              <h2>资产状态</h2>
            </div>
            <el-button text :icon="ArrowRight" @click="go('assetInventory')">资产档案</el-button>
          </header>

          <template v-if="moduleLoaded.assets">
            <dl class="asset-summary">
              <div><dt>资产档案</dt><dd>{{ formatNumber(assetDashboard.assetKinds) }}</dd><small>{{ formatNumber(assetDashboard.categoryCount) }} 个分类</small></div>
              <div><dt>账面原值</dt><dd>{{ formatCompactCurrency(assetDashboard.originalValue) }}</dd><small>当前估值 {{ formatCompactCurrency(assetDashboard.currentValue) }}</small></div>
              <div><dt>资产健康度</dt><dd>{{ healthRate }}%</dd><small>{{ formatNumber(controlledQuantity) }} 件处于受控状态</small></div>
            </dl>

            <div class="asset-status-list" aria-label="资产状态分布">
              <div v-for="status in assetStatusRows" :key="status.key" class="asset-status-row">
                <span class="status-label"><i :class="`tone-${status.tone}`" />{{ status.label }}</span>
                <div class="progress-track"><i :class="`tone-${status.tone}`" :style="{ width: `${status.ratio}%` }" /></div>
                <strong>{{ formatNumber(status.quantity) }}</strong>
              </div>
            </div>

            <div class="compact-list-heading"><span>最近登记</span><span>位置 / 状态</span></div>
            <div v-if="recentAssets.length" class="asset-recent-list">
              <button v-for="item in recentAssets" :key="item.ID" type="button" @click="go('assetInventory')">
                <div class="asset-identity">
                  <strong>{{ item.name }}</strong>
                  <small>{{ item.assetCode || '编号待补充' }}</small>
                </div>
                <div class="asset-place">
                  <span>{{ item.location || '位置待补充' }}</span>
                  <small>{{ statusMeta(item.status).label }}</small>
                </div>
                <b>{{ formatCurrency(item.originalValue) }}</b>
              </button>
            </div>
            <div v-else class="inline-empty">暂无资产登记</div>
          </template>
          <div v-else class="panel-placeholder">资产数据暂不可用</div>
        </article>

        <article v-if="access.invoices" class="na-panel dashboard-panel invoice-panel">
          <header class="na-panel-header panel-heading">
            <div>
              <span>流水管理</span>
              <h2>发票处理</h2>
            </div>
            <el-button text :icon="ArrowRight" @click="go('invoiceDashboard')">流水总览</el-button>
          </header>

          <template v-if="moduleLoaded.invoices">
            <div class="invoice-summary">
              <div class="invoice-total"><span>已确认价税合计</span><strong>{{ centsToCurrency(invoiceDashboard.totalCents) }}</strong><small>{{ formatNumber(invoiceDashboard.confirmedCount) }} 张已进入正式统计</small></div>
              <dl>
                <div><dt>不含税金额</dt><dd>{{ centsToCurrency(invoiceDashboard.amountCents) }}</dd></div>
                <div><dt>税额</dt><dd>{{ centsToCurrency(invoiceDashboard.taxCents) }}</dd></div>
                <div class="is-warning"><dt>待人工核对</dt><dd>{{ formatNumber(invoiceDashboard.pendingCount) }}</dd></div>
                <div class="is-danger"><dt>识别失败</dt><dd>{{ formatNumber(invoiceDashboard.failedCount) }}</dd></div>
              </dl>
            </div>

            <div class="trend-section">
              <div class="compact-list-heading"><span>近 6 个月确认金额</span><span>按开票日期汇总</span></div>
              <div v-if="invoiceTrend.length" class="invoice-trend" aria-label="近六个月已确认发票金额趋势">
                <div v-for="item in invoiceTrend" :key="item.month" class="trend-item">
                  <span class="trend-value">{{ centsToCompactCurrency(item.totalCents) }}</span>
                  <div class="trend-bar"><i :style="{ height: `${item.ratio}%` }" /></div>
                  <small>{{ monthText(item.month) }}</small>
                </div>
              </div>
              <div v-else class="inline-empty">确认发票后将生成月度趋势</div>
            </div>
          </template>
          <div v-else class="panel-placeholder">流水数据暂不可用</div>
        </article>
      </div>

      <aside class="support-column">
        <article v-if="access.calendar" class="na-panel dashboard-panel schedule-panel">
          <header class="na-panel-header panel-heading">
            <div><span>工作日历</span><h2>今日日程</h2></div>
            <el-button text :icon="ArrowRight" @click="go('workSchedule')">查看日历</el-button>
          </header>
          <div v-if="todaySchedules.length" class="schedule-list">
            <button v-for="item in todaySchedules" :key="item.id" type="button" @click="go('workSchedule')">
              <i :style="{ background: item.color }" />
              <time>{{ item.time }}</time>
              <span><strong>{{ item.title }}</strong><small>{{ item.typeLabel }}</small></span>
            </button>
          </div>
          <div v-else class="side-empty"><Calendar /><span>今日暂无日程</span></div>
        </article>

        <article v-if="access.audit" class="na-panel dashboard-panel audit-panel">
          <header class="na-panel-header panel-heading">
            <div><span>审计平台</span><h2>最近操作</h2></div>
            <el-button text :icon="ArrowRight" @click="go('operation')">全部记录</el-button>
          </header>
          <div v-if="moduleLoaded.audit && recentOperations.length" class="operation-list">
            <button v-for="item in recentOperations" :key="item.ID" type="button" @click="go('operation')">
              <span class="request-method">{{ item.method || 'HTTP' }}</span>
              <span class="request-path">{{ item.path || '请求路径待记录' }}</span>
              <time>{{ operationTime(item.CreatedAt) }}</time>
              <i :class="isRequestError(item.status) ? 'request-error' : 'request-ok'">{{ item.status || '—' }}</i>
            </button>
          </div>
          <div v-else-if="moduleLoaded.audit" class="side-empty"><DocumentChecked /><span>暂无操作记录</span></div>
          <div v-else class="panel-placeholder">操作记录暂不可用</div>
        </article>

        <article v-if="access.monitor" class="na-panel dashboard-panel load-panel">
          <header class="na-panel-header panel-heading">
            <div><span>监控状态</span><h2>服务器负载</h2></div>
            <el-button text :icon="ArrowRight" @click="go('state')">详细负载</el-button>
          </header>
          <template v-if="moduleLoaded.monitor">
            <div class="load-list">
              <div v-for="item in systemLoadItems" :key="item.label">
                <span><el-icon><component :is="item.icon" /></el-icon>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
                <div class="progress-track"><i :class="`tone-${item.tone}`" :style="{ width: `${item.percent}%` }" /></div>
              </div>
            </div>
            <footer>主机 {{ systemState.host?.hostname || '—' }} · {{ serverCollectedAt }}</footer>
          </template>
          <div v-else class="panel-placeholder">服务器状态暂不可用</div>
        </article>
      </aside>
    </section>
  </main>
</template>

<script setup>
import { computed, onActivated, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  ArrowRight,
  Calendar,
  Clock,
  Cpu,
  DocumentChecked,
  FolderOpened,
  Plus,
  Refresh,
  Tickets
} from '@element-plus/icons-vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { formatCompactCurrency, formatCurrency, formatNumber } from '@/utils/format'
import { getAssetDashboard } from '@/plugin/asset/api/asset'
import { getAssetOperationList } from '@/plugin/asset/api/operation'
import { getInvoiceDashboard } from '@/plugin/invoice/api/invoice'
import { centsToCurrency } from '@/plugin/invoice/utils/invoice'
import { getSystemState } from '@/api/system'
import { getSysOperationRecordList } from '@/api/sysOperationRecord'
import { useUserStore } from '@/pinia/modules/user'

defineOptions({ name: 'Dashboard' })

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const updatedAt = ref('')
const assetDashboard = ref(createAssetDashboard())
const invoiceDashboard = ref(createInvoiceDashboard())
const assetDraftTotal = ref(0)
const recentOperations = ref([])
const systemState = ref({})
const moduleLoaded = ref({ assets: false, invoices: false, audit: false, monitor: false })
const calendarEvents = ref([])
const calendarTypes = ref([])

const eventStorageKey = 'gva-work-calendar-events'
const typeStorageKey = 'gva-work-calendar-types'
const defaultScheduleTypes = [
  { value: 'task', label: '工作任务', color: '#4f7cf3' },
  { value: 'meeting', label: '会议沟通', color: '#7a61d4' },
  { value: 'asset', label: '资产盘点', color: '#18a678' },
  { value: 'reminder', label: '到期提醒', color: '#d9773c' }
]

const access = computed(() => ({
  assets: router.hasRoute('assetDashboard') || router.hasRoute('assetInventory'),
  invoices: router.hasRoute('invoiceDashboard') || router.hasRoute('invoiceLedger'),
  calendar: router.hasRoute('workSchedule'),
  audit: router.hasRoute('operation'),
  monitor: router.hasRoute('state')
}))

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '夜深了'
  if (hour < 11) return '早上好'
  if (hour < 14) return '中午好'
  if (hour < 18) return '下午好'
  return '晚上好'
})
const currentDateText = computed(() => new Intl.DateTimeFormat('zh-CN', {
  year: 'numeric', month: 'long', day: 'numeric', weekday: 'long'
}).format(new Date()))
const pendingTotal = computed(() => Number(assetDraftTotal.value || 0) + Number(invoiceDashboard.value.pendingCount || 0) + Number(invoiceDashboard.value.failedCount || 0))
const overviewText = computed(() => {
  const counts = []
  if (access.value.assets) counts.push(`${formatNumber(assetDraftTotal.value)} 项资产业务待处理`)
  if (access.value.invoices) counts.push(`${formatNumber(invoiceDashboard.value.pendingCount)} 张发票待核对`)
  return counts.length ? `当前有 ${counts.join('，')}。` : '从左侧菜单进入已获授权的业务模块。'
})
const maintenanceQuantity = computed(() => Number(assetDashboard.value.statusSummary.find((item) => item.status === 'maintenance')?.quantity || 0))
const controlledQuantity = computed(() => Math.max(Number(assetDashboard.value.totalQuantity || 0) - maintenanceQuantity.value, 0))
const healthRate = computed(() => {
  const total = Number(assetDashboard.value.totalQuantity || 0)
  return total ? ((controlledQuantity.value / total) * 100).toFixed(1) : '0.0'
})
const metrics = computed(() => {
  const items = []
  if (access.value.assets) {
    items.push({ label: '资产实物总量', value: formatNumber(assetDashboard.value.totalQuantity), hint: `${formatNumber(assetDashboard.value.categoryCount)} 个分类`, tone: '' })
    items.push({ label: '资产账面原值', value: formatCompactCurrency(assetDashboard.value.originalValue), hint: `当前估值 ${formatCompactCurrency(assetDashboard.value.currentValue)}`, tone: '' })
  }
  if (access.value.invoices) {
    items.push({ label: '已确认发票', value: centsToCurrency(invoiceDashboard.value.totalCents), hint: `${formatNumber(invoiceDashboard.value.confirmedCount)} 张已确认`, tone: '' })
    items.push({ label: '待处理事项', value: formatNumber(pendingTotal.value), hint: `${formatNumber(invoiceDashboard.value.failedCount)} 张识别失败`, tone: pendingTotal.value ? 'is-warning' : '' })
  }
  return items
})
const recentAssets = computed(() => assetDashboard.value.recentAssets.slice(0, 4))
const assetStatusRows = computed(() => assetStatusOrder.map((item) => {
  const quantity = Number(assetDashboard.value.statusSummary.find((summary) => summary.status === item.key)?.quantity || 0)
  const total = Number(assetDashboard.value.totalQuantity || 0)
  return { ...item, quantity, ratio: total ? Math.max((quantity / total) * 100, quantity ? 3 : 0) : 0 }
}))
const invoiceTrend = computed(() => {
  const values = invoiceDashboard.value.monthlyTrend.slice(-6)
  const maximum = Math.max(...values.map((item) => Number(item.totalCents || 0)), 0)
  return values.map((item) => ({ ...item, ratio: maximum ? Math.max((Number(item.totalCents || 0) / maximum) * 100, Number(item.totalCents || 0) ? 5 : 0) : 0 }))
})
const todaySchedules = computed(() => calendarEvents.value
  .filter((item) => item.date === dateKey(new Date()))
  .sort((left, right) => String(left.time).localeCompare(String(right.time)))
  .slice(0, 4)
  .map((item) => {
    const type = calendarTypes.value.find((entry) => entry.value === item.type) || defaultScheduleTypes[0]
    return { ...item, typeLabel: type.label, color: type.color }
  }))
const systemUsage = computed(() => ({
  cpu: average(systemState.value.cpu?.cpus),
  ram: Number(systemState.value.ram?.usedPercent || 0),
  disk: Math.max(0, ...(systemState.value.disk || []).map((item) => Number(item.usedPercent || 0)))
}))
const systemHealth = computed(() => {
  const maximum = Math.max(systemUsage.value.cpu, systemUsage.value.ram, systemUsage.value.disk)
  if (maximum >= 80) return { label: '资源压力较高', tone: 'danger' }
  if (maximum >= 60) return { label: '运行负载偏高', tone: 'warning' }
  return { label: '系统运行正常', tone: 'success' }
})
const systemLoadItems = computed(() => [
  { label: 'CPU 平均使用率', value: percent(systemUsage.value.cpu), percent: safePercent(systemUsage.value.cpu), icon: Cpu, tone: usageTone(systemUsage.value.cpu) },
  { label: '内存使用率', value: percent(systemUsage.value.ram), percent: safePercent(systemUsage.value.ram), icon: DocumentChecked, tone: usageTone(systemUsage.value.ram) },
  { label: '磁盘最高使用率', value: percent(systemUsage.value.disk), percent: safePercent(systemUsage.value.disk), icon: FolderOpened, tone: usageTone(systemUsage.value.disk) }
])
const serverCollectedAt = computed(() => systemState.value.collectedAt
  ? new Date(systemState.value.collectedAt).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
  : '—')

const assetStatusOrder = [
  { key: 'in_use', label: '在用', tone: 'success' },
  { key: 'idle', label: '闲置', tone: 'info' },
  { key: 'maintenance', label: '维修维保', tone: 'warning' },
  { key: 'pending_inbound', label: '待入库', tone: 'primary' },
  { key: 'retired', label: '已处置', tone: 'danger' }
]
const statusMap = Object.fromEntries(assetStatusOrder.map((item) => [item.key, item]))

function createAssetDashboard() {
  return { assetKinds: 0, totalQuantity: 0, categoryCount: 0, originalValue: 0, currentValue: 0, statusSummary: [], recentAssets: [] }
}
function createInvoiceDashboard() {
  return { confirmedCount: 0, pendingCount: 0, failedCount: 0, totalCents: 0, amountCents: 0, taxCents: 0, monthlyTrend: [] }
}
function dateKey(date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
function safePercent(value) {
  return Math.min(100, Math.max(0, Math.round(Number(value || 0))))
}
function percent(value) {
  return `${safePercent(value)}%`
}
function average(values = []) {
  return values.length ? values.reduce((sum, value) => sum + Number(value || 0), 0) / values.length : 0
}
function usageTone(value) {
  const percentValue = safePercent(value)
  if (percentValue >= 80) return 'danger'
  if (percentValue >= 60) return 'warning'
  return 'success'
}
function statusMeta(status) {
  return statusMap[status] || { label: status || '未知', tone: 'info' }
}
function centsToCompactCurrency(value) {
  return formatCompactCurrency(Number(value || 0) / 100)
}
function monthText(value) {
  return value ? `${String(value).slice(5)}月` : '—'
}
function operationTime(value) {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
}
function isRequestError(status) {
  return Number(status || 0) >= 400
}
function go(name) {
  if (router.hasRoute(name)) router.push({ name })
}
function readCalendarStorage() {
  if (!access.value.calendar) return
  try {
    const savedTypes = JSON.parse(window.localStorage.getItem(typeStorageKey) || '[]')
    calendarTypes.value = Array.isArray(savedTypes) && savedTypes.length ? savedTypes : defaultScheduleTypes
  } catch {
    calendarTypes.value = defaultScheduleTypes
  }
  try {
    const savedEvents = JSON.parse(window.localStorage.getItem(eventStorageKey) || '[]')
    calendarEvents.value = Array.isArray(savedEvents) ? savedEvents.filter((item) => item && item.id && item.title && item.date && item.time) : []
  } catch {
    calendarEvents.value = []
  }
}
function handleStorageChange(event) {
  if (event.key === eventStorageKey || event.key === typeStorageKey) readCalendarStorage()
}
async function loadAssets() {
  const [dashboardResult, operationResult] = await Promise.allSettled([
    getAssetDashboard(),
    getAssetOperationList({ page: 1, pageSize: 1, status: 'draft' })
  ])
  if (dashboardResult.status === 'fulfilled' && dashboardResult.value.code === 0) {
    const data = dashboardResult.value.data || {}
    assetDashboard.value = { ...createAssetDashboard(), ...data, statusSummary: data.statusSummary || [], recentAssets: data.recentAssets || [] }
    moduleLoaded.value.assets = true
  }
  if (operationResult.status === 'fulfilled' && operationResult.value.code === 0) {
    assetDraftTotal.value = Number(operationResult.value.data?.total || 0)
  }
}
async function loadInvoices() {
  const result = await getInvoiceDashboard()
  if (result.code === 0) {
    const data = result.data || {}
    invoiceDashboard.value = { ...createInvoiceDashboard(), ...data, monthlyTrend: data.monthlyTrend || [] }
    moduleLoaded.value.invoices = true
  }
}
async function loadAudit() {
  const result = await getSysOperationRecordList({ page: 1, pageSize: 4 })
  if (result.code === 0) {
    recentOperations.value = result.data?.list || []
    moduleLoaded.value.audit = true
  }
}
async function loadMonitor() {
  const result = await getSystemState()
  if (result.code === 0 && result.data?.server) {
    systemState.value = result.data.server
    moduleLoaded.value.monitor = true
  }
}
async function loadDashboard() {
  loading.value = true
  moduleLoaded.value = { assets: false, invoices: false, audit: false, monitor: false }
  try {
    readCalendarStorage()
    await Promise.allSettled([
      access.value.assets ? loadAssets() : Promise.resolve(),
      access.value.invoices ? loadInvoices() : Promise.resolve(),
      access.value.audit ? loadAudit() : Promise.resolve(),
      access.value.monitor ? loadMonitor() : Promise.resolve()
    ])
    updatedAt.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  window.addEventListener('storage', handleStorageChange)
  loadDashboard()
})
onActivated(readCalendarStorage)
onUnmounted(() => window.removeEventListener('storage', handleStorageChange))
</script>

<style scoped lang="scss">
.dashboard-page { min-height: 100%; padding: 18px 20px 24px; background: var(--na-background); color: var(--na-foreground); }
.updated-at { color: var(--na-muted-foreground); font-size: .75rem; white-space: nowrap; }

.workbench-band { display: flex; min-width: 0; align-items: stretch; justify-content: space-between; gap: 20px; margin-bottom: 12px; padding: 18px 20px; border: 1px solid var(--na-border); border-radius: var(--na-radius); background: var(--na-card); }
.workbench-copy { min-width: 0; }
.current-date { margin: 0 0 5px; color: var(--na-primary); font-size: .75rem; font-weight: 650; }
.workbench-copy h2 { margin: 0; font-size: 1.35rem; font-weight: 680; letter-spacing: 0; }
.workbench-copy > p:last-of-type { margin: 7px 0 0; color: var(--na-muted-foreground); font-size: .8125rem; }
.quick-actions { display: flex; flex-wrap: wrap; gap: 2px 6px; margin-top: 10px; }
.quick-actions :deep(.el-button) { height: 28px; padding: 0 4px; font-size: .75rem; }

.runtime-summary { display: grid; min-width: 300px; grid-template-columns: 1fr; gap: 10px; padding: 2px 0 2px 20px; border: 0; border-left: 1px solid var(--na-border); background: transparent; color: var(--na-foreground); text-align: left; }
.runtime-summary:hover .runtime-heading { color: var(--na-primary); }
.runtime-heading { display: inline-flex; align-items: center; gap: 7px; font-size: .75rem; font-weight: 630; transition: color 160ms ease; }
.runtime-heading i { width: 7px; height: 7px; border-radius: 50%; background: var(--na-success); box-shadow: 0 0 0 3px var(--na-success-soft); }
.runtime-heading.health-warning i { background: var(--na-warning); box-shadow: 0 0 0 3px var(--na-warning-soft); }
.runtime-heading.health-danger i { background: var(--na-danger); box-shadow: 0 0 0 3px var(--na-danger-soft); }
.runtime-summary dl { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin: 0; }
.runtime-summary dl div { min-width: 0; }
.runtime-summary dt, .runtime-summary small { color: var(--na-muted-foreground); font-size: .6875rem; }
.runtime-summary dd { margin: 4px 0 0; color: var(--na-foreground); font-size: 1rem; font-variant-numeric: tabular-nums; font-weight: 670; }

.metric-band { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); margin-bottom: 12px; border: 1px solid var(--na-border); border-radius: var(--na-radius); background: var(--na-card); }
.metric-item { display: flex; min-width: 0; min-height: 94px; flex-direction: column; justify-content: center; gap: 4px; padding: 14px 18px; border-right: 1px solid var(--na-border); }
.metric-item:last-child { border-right: 0; }
.metric-item > span, .metric-item small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.metric-item strong { overflow: hidden; color: var(--na-foreground); font-size: 1.25rem; font-variant-numeric: tabular-nums; font-weight: 680; text-overflow: ellipsis; white-space: nowrap; }
.metric-item small.is-warning { color: var(--na-warning); }

.dashboard-workspace { display: grid; min-width: 0; grid-template-columns: minmax(0, 1.65fr) minmax(310px, .9fr); gap: 12px; }
.business-column, .support-column { display: grid; min-width: 0; align-content: start; gap: 12px; }
.dashboard-panel { min-width: 0; overflow: hidden; }
.panel-heading > div { min-width: 0; }
.panel-heading span { display: block; color: var(--na-muted-foreground); font-size: .6875rem; }
.panel-heading h2 { margin: 3px 0 0; color: var(--na-foreground); font-size: .9375rem; font-weight: 660; }
.panel-heading :deep(.el-button) { font-size: .75rem; }
.panel-placeholder, .inline-empty { display: grid; min-height: 84px; place-items: center; color: var(--na-muted-foreground); font-size: .75rem; }

.asset-summary { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); margin: 0; padding: 14px 16px; border-bottom: 1px solid var(--na-border); }
.asset-summary div { min-width: 0; padding: 0 14px; border-left: 1px solid var(--na-border); }
.asset-summary div:first-child { padding-left: 0; border-left: 0; }
.asset-summary div:last-child { padding-right: 0; }
.asset-summary dt, .asset-summary small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.asset-summary dd { overflow: hidden; margin: 5px 0 3px; color: var(--na-foreground); font-size: 1rem; font-variant-numeric: tabular-nums; font-weight: 670; text-overflow: ellipsis; white-space: nowrap; }
.asset-status-list { display: grid; gap: 8px; padding: 14px 16px 13px; }
.asset-status-row { display: grid; min-width: 0; grid-template-columns: 82px minmax(0, 1fr) 38px; align-items: center; gap: 10px; }
.status-label { display: inline-flex; align-items: center; gap: 6px; color: var(--na-muted-foreground); font-size: .75rem; white-space: nowrap; }
.status-label i { width: 6px; height: 6px; border-radius: 50%; background: var(--na-muted-foreground); }
.progress-track { height: 5px; overflow: hidden; border-radius: 3px; background: var(--na-muted); }
.progress-track > i { display: block; height: 100%; border-radius: inherit; background: var(--na-primary); transition: width 180ms ease; }
.tone-success { background: var(--na-success) !important; }
.tone-warning { background: var(--na-warning) !important; }
.tone-danger { background: var(--na-danger) !important; }
.tone-info { background: var(--na-info) !important; }
.tone-primary { background: var(--na-primary) !important; }
.asset-status-row > strong { color: var(--na-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; text-align: right; }
.compact-list-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 9px 16px; border-top: 1px solid var(--na-border); border-bottom: 1px solid var(--na-border); background: var(--na-table-header); color: var(--na-muted-foreground); font-size: .6875rem; }
.compact-list-heading span:last-child { text-align: right; }
.asset-recent-list { display: grid; }
.asset-recent-list button { display: grid; min-width: 0; min-height: 52px; grid-template-columns: minmax(0, 1.35fr) minmax(90px, .8fr) minmax(100px, .65fr); align-items: center; gap: 12px; padding: 7px 16px; border: 0; border-bottom: 1px solid var(--na-border); background: transparent; color: var(--na-foreground); text-align: left; }
.asset-recent-list button:last-child { border-bottom: 0; }
.asset-recent-list button:hover, .operation-list button:hover, .schedule-list button:hover { background: var(--na-table-hover); }
.asset-identity, .asset-place { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.asset-identity strong, .asset-place span { overflow: hidden; font-size: .75rem; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.asset-identity small, .asset-place small { overflow: hidden; color: var(--na-muted-foreground); font-size: .625rem; text-overflow: ellipsis; white-space: nowrap; }
.asset-recent-list b { overflow: hidden; color: var(--na-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; text-align: right; text-overflow: ellipsis; white-space: nowrap; }

.invoice-summary { display: grid; grid-template-columns: minmax(200px, .9fr) minmax(0, 1.6fr); border-bottom: 1px solid var(--na-border); }
.invoice-total { display: flex; min-width: 0; flex-direction: column; justify-content: center; padding: 16px; background: var(--na-primary-soft); }
.invoice-total span, .invoice-total small, .invoice-summary dt { color: var(--na-muted-foreground); font-size: .6875rem; }
.invoice-total strong { overflow: hidden; margin: 5px 0 3px; color: var(--na-primary); font-size: 1.25rem; font-variant-numeric: tabular-nums; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.invoice-summary dl { display: grid; grid-template-columns: 1fr 1fr; margin: 0; }
.invoice-summary dl div { display: flex; min-width: 0; flex-direction: column; justify-content: center; gap: 4px; padding: 13px 16px; border-left: 1px solid var(--na-border); border-bottom: 1px solid var(--na-border); }
.invoice-summary dl div:nth-child(3), .invoice-summary dl div:nth-child(4) { border-bottom: 0; }
.invoice-summary dd { overflow: hidden; margin: 0; color: var(--na-foreground); font-size: .875rem; font-variant-numeric: tabular-nums; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.invoice-summary .is-warning dd { color: var(--na-warning); }
.invoice-summary .is-danger dd { color: var(--na-danger); }
.trend-section { padding-bottom: 12px; }
.invoice-trend { display: grid; height: 118px; grid-template-columns: repeat(6, minmax(0, 1fr)); align-items: end; gap: 10px; padding: 8px 16px 0; }
.trend-item { display: grid; min-width: 0; height: 100%; grid-template-rows: 16px minmax(0, 1fr) 15px; align-items: end; gap: 4px; }
.trend-value { overflow: hidden; color: var(--na-muted-foreground); font-size: .5625rem; text-align: center; text-overflow: ellipsis; white-space: nowrap; }
.trend-bar { display: flex; height: 100%; align-items: end; justify-content: center; border-bottom: 1px solid var(--na-border); }
.trend-bar i { width: min(24px, 64%); min-height: 2px; border-radius: 3px 3px 0 0; background: var(--na-primary); }
.trend-item small { color: var(--na-muted-foreground); font-size: .625rem; text-align: center; }

.schedule-list, .operation-list { display: grid; }
.schedule-list button { display: grid; min-width: 0; min-height: 48px; grid-template-columns: 7px 40px minmax(0, 1fr); align-items: center; gap: 8px; padding: 6px 14px; border: 0; border-bottom: 1px solid var(--na-border); background: transparent; color: var(--na-foreground); text-align: left; }
.schedule-list button:last-child, .operation-list button:last-child { border-bottom: 0; }
.schedule-list i { width: 6px; height: 24px; border-radius: 3px; }
.schedule-list time { color: var(--na-muted-foreground); font-size: .6875rem; font-variant-numeric: tabular-nums; }
.schedule-list span { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.schedule-list strong { overflow: hidden; font-size: .75rem; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.schedule-list small { overflow: hidden; color: var(--na-muted-foreground); font-size: .625rem; text-overflow: ellipsis; white-space: nowrap; }
.side-empty { display: grid; min-height: 102px; place-items: center; align-content: center; gap: 7px; color: var(--na-muted-foreground); font-size: .75rem; }
.side-empty :deep(svg) { width: 22px; height: 22px; color: var(--na-primary); }
.operation-list button { display: grid; min-width: 0; min-height: 40px; grid-template-columns: 40px minmax(0, 1fr) 38px 34px; align-items: center; gap: 8px; padding: 0 14px; border: 0; border-bottom: 1px solid var(--na-border); background: transparent; color: var(--na-foreground); text-align: left; }
.request-method { overflow: hidden; color: var(--na-primary); font: 600 .625rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; text-overflow: ellipsis; white-space: nowrap; }
.request-path { overflow: hidden; font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.operation-list time { color: var(--na-muted-foreground); font-size: .625rem; font-variant-numeric: tabular-nums; }
.operation-list i { display: inline-grid; min-width: 30px; place-items: center; border-radius: 4px; color: var(--na-on-primary); font-size: .5625rem; font-style: normal; font-variant-numeric: tabular-nums; line-height: 18px; }
.operation-list i.request-ok { color: var(--na-success); background: var(--na-success-soft); }
.operation-list i.request-error { color: var(--na-danger); background: var(--na-danger-soft); }
.load-list { display: grid; gap: 12px; padding: 14px; }
.load-list > div { display: grid; min-width: 0; grid-template-columns: minmax(0, 1fr) 42px; gap: 6px 10px; align-items: center; }
.load-list span { display: inline-flex; min-width: 0; align-items: center; gap: 6px; overflow: hidden; color: var(--na-muted-foreground); font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.load-list .el-icon { color: var(--na-primary); font-size: .875rem; }
.load-list strong { color: var(--na-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; text-align: right; }
.load-list .progress-track { grid-column: 1 / -1; }
.load-panel footer { display: block; overflow: hidden; padding: 9px 14px; border-top: 1px solid var(--na-border); color: var(--na-muted-foreground); font-size: .625rem; text-overflow: ellipsis; white-space: nowrap; }

@media (max-width: 1120px) {
  .dashboard-workspace { grid-template-columns: 1fr; }
  .support-column { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .load-panel { grid-column: 1 / -1; }
}
@media (max-width: 860px) {
  .dashboard-page { padding: 15px; }
  .workbench-band { flex-direction: column; }
  .runtime-summary { min-width: 0; padding: 14px 0 0; border-top: 1px solid var(--na-border); border-left: 0; }
  .metric-item:nth-child(2n) { border-right: 0; }
}
@media (max-width: 640px) {
  .dashboard-page { padding: 12px; }
  .workbench-band { padding: 15px; }
  .support-column { grid-template-columns: 1fr; }
  .load-panel { grid-column: auto; }
  .asset-summary { grid-template-columns: 1fr; gap: 12px; }
  .asset-summary div, .asset-summary div:first-child, .asset-summary div:last-child { padding: 0; border-left: 0; }
  .asset-recent-list button { grid-template-columns: minmax(0, 1fr) auto; }
  .asset-recent-list b { display: none; }
  .invoice-summary { grid-template-columns: 1fr; }
  .invoice-summary dl div { border-left: 0; }
  .operation-list button { grid-template-columns: 38px minmax(0, 1fr) 34px; }
  .operation-list time { display: none; }
  .metric-band { grid-template-columns: 1fr; }
  .metric-item, .metric-item:nth-child(2n) { min-height: 76px; border-right: 0; border-bottom: 1px solid var(--na-border); }
  .metric-item:last-child { border-bottom: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .progress-track > i, .runtime-heading { transition: none; }
}
</style>
