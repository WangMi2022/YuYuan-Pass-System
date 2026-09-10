<template>
  <div class="dashboard-route-view">
    <!-- Keep an element root for the layout's out-in transition when switching to Teleport. -->
    <LeadershipWallboard
      v-if="isWallboardView"
      :snapshot="wallboardSnapshot"
      :loading="loading"
      @refresh="loadDashboard"
      @exit="closeWallboard"
    />

    <PendingTasks v-else-if="isPendingView" @back="closePendingItems" />

    <main v-else class="dashboard-page">
      <AppPageHeader
        title-id="dashboard-title"
        title="首页驾驶舱"
        description="汇总当前权限范围内的资产、流水、风险与日程。"
      >
        <template #actions>
          <span class="updated-at">{{ refreshText }}</span>
          <el-button v-if="canOpenWallboard" class="cockpit-fullscreen-btn" type="primary" plain :icon="FullScreen" @click="openWallboard">大屏驾驶舱</el-button>
          <el-button :icon="Refresh" :loading="loading" @click="loadDashboard">刷新</el-button>
          <div v-if="access.assetInventory || access.invoiceRecognition" class="header-primary-actions">
            <el-button v-if="access.assetInventory" type="primary" :icon="Plus" @click="go('assetInventory')">登记资产</el-button>
            <el-button v-if="access.invoiceRecognition" type="primary" :icon="Tickets" @click="go('invoiceRecognition')">上传发票</el-button>
          </div>
        </template>
      </AppPageHeader>

      <section class="workbench-band" aria-labelledby="workbench-title">
        <HeroCanvas class="workbench-hero-canvas" />
        <div class="workbench-copy">
          <div class="workbench-time-row">
            <span class="cockpit-live-tag">
              <i class="live-pulse-dot" />
              <span>COCKPIT TELEMETRY</span>
            </span>
            <p class="current-date">{{ currentDateText }} · {{ currentTimeText }}</p>
          </div>
          <h2 id="workbench-title">{{ greeting }}，{{ userStore.userInfo.nickName || userStore.userInfo.userName || '用户' }}</h2>
          <p>{{ overviewText }}</p>
          <div class="quick-actions" aria-label="常用操作">
            <el-button v-if="access.invoiceLedger" text :icon="DocumentChecked" @click="go('invoiceLedger')">发票台账</el-button>
            <el-button v-if="access.calendar" text :icon="Calendar" @click="go('workSchedule')">日程总览</el-button>
            <el-button v-if="access.audit" text :icon="Clock" @click="go('operation')">操作历史</el-button>
          </div>
        </div>

        <button
          type="button"
          class="runtime-summary"
          :class="{ 'is-actionable': access.monitor }"
          :disabled="!access.monitor"
          aria-label="服务器监控摘要"
          @click="go('state')"
        >
          <div class="runtime-ecg-track" aria-hidden="true">
            <svg viewBox="0 0 120 16" preserveAspectRatio="none" class="runtime-ecg-svg">
              <path class="runtime-ecg-path" d="M 0 8 L 36 8 L 42 2 L 48 14 L 54 4 L 60 11 L 66 8 L 120 8" />
            </svg>
          </div>
          <span class="runtime-topline">
            <span class="runtime-heading" :class="`health-${systemHealth.tone}`"><i />{{ moduleLoaded.monitor ? systemHealth.label : '服务器监控' }}</span>
            <small>{{ moduleLoaded.monitor ? (moduleFailed.monitor ? moduleFreshnessShort('monitor') : `采集于 ${serverCollectedAt}`) : '暂无可用数据' }}</small>
          </span>
          <dl>
            <div>
              <dt>CPU</dt><dd>{{ moduleLoaded.monitor ? percent(systemUsage.cpu) : '—' }}</dd>
              <span class="runtime-track"><i :class="`tone-${usageTone(systemUsage.cpu)}`" :style="{ width: `${moduleLoaded.monitor ? safePercent(systemUsage.cpu) : 0}%` }" /></span>
            </div>
            <div>
              <dt>内存</dt><dd>{{ moduleLoaded.monitor ? percent(systemUsage.ram) : '—' }}</dd>
              <span class="runtime-track"><i :class="`tone-${usageTone(systemUsage.ram)}`" :style="{ width: `${moduleLoaded.monitor ? safePercent(systemUsage.ram) : 0}%` }" /></span>
            </div>
            <div>
              <dt>磁盘</dt><dd>{{ moduleLoaded.monitor ? percent(systemUsage.disk) : '—' }}</dd>
              <span class="runtime-track"><i :class="`tone-${usageTone(systemUsage.disk)}`" :style="{ width: `${moduleLoaded.monitor ? safePercent(systemUsage.disk) : 0}%` }" /></span>
            </div>
          </dl>
        </button>
      </section>

      <section v-if="metrics.length" class="metric-band" aria-label="核心业务指标">
        <component
          :is="metric.action ? 'button' : 'article'"
          v-for="(metric, idx) in metrics"
          :key="metric.label"
          class="metric-item"
          :class="[`metric-item--${metric.tone}`, { 'metric-item--actionable': metric.action }]"
          :type="metric.action ? 'button' : undefined"
          :aria-label="metric.action ? metric.actionLabel : undefined"
          :style="{ '--metric-delay': `${idx * 45}ms` }"
          @click="handleMetricClick(metric)"
        >
          <div class="metric-accent-stripe" />
          <div class="metric-copy">
            <div class="metric-title-row">
              <span>{{ metric.label }}</span>
              <span class="metric-status-glow" :class="`tone-${metric.tone}`" />
            </div>
            <strong>
              <AnimatedValue
                v-if="metric.numValue !== undefined"
                :value="metric.numValue"
                :format="metric.format"
              />
              <template v-else>{{ metric.value }}</template>
            </strong>
            <small :class="metric.tone === 'warning' ? 'is-warning' : ''">{{ metric.hint }}</small>
          </div>
          <div class="metric-icon-box" :class="`metric-${metric.tone}`">
            <el-icon class="metric-icon"><component :is="metric.icon" /></el-icon>
          </div>
        </component>
      </section>

      <section class="dashboard-workspace">
        <div class="business-column">
          <article v-if="access.assets" class="na-panel dashboard-panel asset-panel">
            <header class="na-panel-header panel-heading">
              <div>
                <span>资产管理</span>
                <h2>资产状态</h2>
              </div>
              <el-button v-if="access.assetInventory" text :icon="ArrowRight" @click="go('assetInventory')">资产档案</el-button>
            </header>

            <p v-if="moduleFailed.assets && moduleLoaded.assets" class="module-stale-notice">{{ moduleFreshnessText('assets') }}</p>
            <template v-if="moduleLoaded.assets">
              <dl class="asset-summary">
                <div><dt>资产档案</dt><dd><AnimatedValue :value="Number(assetDashboard.assetKinds || 0)" :format="formatNumber" /></dd><small>{{ formatNumber(assetDashboard.categoryCount) }} 个分类</small></div>
                <div><dt>账面原值</dt><dd><AnimatedValue :value="Number(assetDashboard.originalValue || 0)" :format="formatCompactCurrency" /></dd><small>当前估值 {{ formatCompactCurrency(assetDashboard.currentValue) }}</small></div>
                <div><dt>资产健康度</dt><dd><AnimatedValue :value="Number(healthRate || 0)" :format="(val) => `${Number(val).toFixed(1)}%`" /></dd><small>{{ formatNumber(controlledQuantity) }} 件处于受控状态</small></div>
              </dl>

              <div class="asset-detail-grid">
                <section class="asset-status-section" aria-label="资产状态分布">
                  <div class="section-mini-heading"><span>状态分布</span><small>共 {{ formatNumber(assetDashboard.totalQuantity) }} 件</small></div>
                  <div class="asset-status-list">
                    <div v-for="status in assetStatusRows" :key="status.key" class="asset-status-row">
                      <span class="status-label"><i :class="`tone-${status.tone}`" />{{ status.label }}</span>
                      <div class="progress-track"><i :class="`tone-${status.tone}`" :style="{ width: `${status.ratio}%` }" /></div>
                      <strong>{{ formatNumber(status.quantity) }}</strong>
                    </div>
                  </div>
                </section>

                <section class="asset-recent-section" aria-label="最近登记资产">
                  <div class="section-mini-heading"><span>最近登记</span><small>位置 / 状态 / 原值</small></div>
                  <div class="asset-recent-table-head"><span>资产</span><span>位置 / 状态</span><span>原值</span></div>
                  <div v-if="recentAssets.length" class="asset-recent-list">
                    <button v-for="item in recentAssets" :key="item.ID" type="button" @click="go('assetInventory')">
                      <div class="asset-identity">
                        <strong>{{ item.name }}</strong>
                        <small>{{ item.assetCode || '编号待补充' }}</small>
                      </div>
                      <div class="asset-place">
                        <span>{{ item.location || '位置待补充' }}</span>
                        <small :class="`status-${statusMeta(item.status).tone}`">{{ statusMeta(item.status).label }}</small>
                      </div>
                      <b>{{ formatCurrency(item.originalValue) }}</b>
                    </button>
                  </div>
                  <div v-else class="inline-empty">暂无资产登记</div>
                </section>
              </div>
            </template>
            <div v-else class="panel-placeholder">资产数据暂不可用</div>
          </article>

          <article v-if="access.invoices" class="na-panel dashboard-panel invoice-panel">
            <header class="na-panel-header panel-heading">
              <div>
                <span>流水管理</span>
                <h2>发票处理</h2>
              </div>
              <el-button v-if="access.invoiceDashboard" text :icon="ArrowRight" @click="go('invoiceDashboard')">流水总览</el-button>
            </header>

            <p v-if="moduleFailed.invoices && moduleLoaded.invoices" class="module-stale-notice">{{ moduleFreshnessText('invoices') }}</p>
            <template v-if="moduleLoaded.invoices">
              <div class="invoice-workspace">
                <div class="invoice-total">
                  <span>已确认价税合计</span>
                  <strong><AnimatedValue :value="Number(invoiceDashboard.totalCents || 0)" :format="centsToCurrency" /></strong>
                  <small>{{ formatNumber(invoiceDashboard.confirmedCount) }} 张已进入正式统计</small>
                  <dl class="invoice-breakdown">
                    <div><dt>不含税金额</dt><dd><AnimatedValue :value="Number(invoiceDashboard.amountCents || 0)" :format="centsToCurrency" /></dd></div>
                    <div><dt>税额</dt><dd><AnimatedValue :value="Number(invoiceDashboard.taxCents || 0)" :format="centsToCurrency" /></dd></div>
                  </dl>
                </div>

                <div class="trend-section">
                  <div class="invoice-trend-heading">
                    <div><span>近 6 个月确认金额</span><small>按开票日期汇总</small></div>
                    <div class="invoice-exceptions" aria-label="发票异常事项">
                      <span class="is-warning">待核 {{ formatNumber(invoiceDashboard.pendingCount) }}</span>
                      <span class="is-danger">失败 {{ formatNumber(invoiceDashboard.failedCount) }}</span>
                    </div>
                  </div>
                  <div v-if="invoiceTrend.length" class="invoice-trend" aria-label="近六个月已确认发票金额趋势">
                    <div v-for="item in invoiceTrend" :key="item.month" class="trend-item">
                      <span class="trend-value">{{ centsToCompactCurrency(item.totalCents) }}</span>
                      <div class="trend-bar"><i :style="{ height: `${item.ratio}%` }" /></div>
                      <small>{{ monthText(item.month) }}</small>
                    </div>
                  </div>
                  <div v-else class="inline-empty">确认发票后将生成月度趋势</div>
                </div>
              </div>
            </template>
            <div v-else class="panel-placeholder">流水数据暂不可用</div>
          </article>
        </div>

        <aside class="support-column">
          <article v-if="access.risk" class="na-panel dashboard-panel daily-risk-panel">
            <header class="na-panel-header panel-heading">
              <div><span>风险闭环</span><h2>领导关注</h2></div>
              <el-button text :icon="ArrowRight" @click="go('assetRiskCenter')">风险中心</el-button>
            </header>
            <p v-if="moduleFailed.risk && moduleLoaded.risk" class="module-stale-notice">{{ moduleFreshnessText('risk') }}</p>
            <template v-if="moduleLoaded.risk">
              <div class="daily-risk-summary">
                <div class="daily-risk-total">
                  <span>开放风险</span>
                  <strong :class="riskDashboard.highOpen ? 'is-danger' : 'is-success'">
                    <AnimatedValue :value="Number(riskDashboard.totalOpen || 0)" :format="formatNumber" />
                    <i v-if="riskDashboard.highOpen" class="danger-pulse-dot" title="存在高风险项" />
                  </strong>
                  <small>高风险 {{ formatNumber(riskDashboard.highOpen) }} 项</small>
                </div>
                <dl>
                  <div><dt>今日新增</dt><dd><AnimatedValue :value="Number(riskDashboard.todayNew || 0)" :format="formatNumber" /></dd></div>
                  <div><dt>超期未结</dt><dd class="is-warning"><AnimatedValue :value="Number(riskDashboard.overdue || 0)" :format="formatNumber" /></dd></div>
                </dl>
              </div>
              <div v-if="riskTrendBars.length" class="daily-risk-trend" aria-label="近七日风险新增与关闭趋势">
                <span v-for="item in riskTrendBars" :key="item.date" :title="`${item.date} 新增 ${item.new} · 关闭 ${item.resolved}`">
                  <i class="risk-new" :style="{ height: `${item.newRatio}%` }" />
                  <i class="risk-resolved" :style="{ height: `${item.resolvedRatio}%` }" />
                </span>
              </div>
            </template>
            <div v-else class="panel-placeholder">风险数据暂不可用</div>
          </article>

          <article v-if="access.calendar" class="na-panel dashboard-panel schedule-panel">
            <header class="na-panel-header panel-heading">
              <div><span>工作日历</span><h2>今日日程</h2></div>
              <el-button text :icon="ArrowRight" @click="go('workSchedule')">查看日历</el-button>
            </header>
            <p v-if="moduleFailed.calendar && moduleLoaded.calendar" class="module-stale-notice">{{ moduleFreshnessText('calendar') }}</p>
            <div v-if="moduleLoaded.calendar && todaySchedules.length" class="schedule-list">
              <button v-for="item in todaySchedules" :key="item.id" type="button" @click="go('workSchedule')">
                <i :style="{ background: item.color }" />
                <time>{{ item.time }}</time>
                <span>
                  <strong>{{ item.title }}</strong>
                  <small>{{ item.typeLabel }}<template v-if="item.repeatLabel"> · {{ item.repeatLabel }}</template></small>
                </span>
              </button>
            </div>
            <div v-else-if="moduleLoaded.calendar" class="side-empty"><Calendar /><span>今日暂无日程</span></div>
            <div v-else class="panel-placeholder">日程数据暂不可用</div>
            <footer v-if="moduleLoaded.calendar && todaySchedules.length" class="schedule-footer">今日共 {{ formatNumber(todaySchedules.length) }} 项安排</footer>
          </article>

          <article v-if="access.audit" class="na-panel dashboard-panel audit-panel">
            <header class="na-panel-header panel-heading">
              <div><span>审计平台</span><h2>最近操作</h2></div>
              <el-button text :icon="ArrowRight" @click="go('operation')">全部记录</el-button>
            </header>
            <p v-if="moduleFailed.audit && moduleLoaded.audit" class="module-stale-notice">{{ moduleFreshnessText('audit') }}</p>
            <div v-if="moduleLoaded.audit && recentOperations.length" class="operation-list">
              <div class="operation-table-head"><span>方法</span><span>请求路径</span><span>时间</span><span>状态</span></div>
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

        </aside>
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowRight,
  Box,
  Calendar,
  Clock,
  Coin,
  DocumentChecked,
  FullScreen,
  Plus,
  Refresh,
  Tickets,
  WarningFilled
} from '@element-plus/icons-vue'
import { dateKey, recurrenceLabel, scheduleMatchesDate } from '@/utils/workCalendar'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import PendingTasks from '@/view/dashboard/PendingTasks.vue'
import LeadershipWallboard from '@/view/dashboard/LeadershipWallboard.vue'
import AnimatedValue from '@/view/dashboard/wallboard/AnimatedValue.vue'
import HeroCanvas from '@/components/three/HeroCanvas.vue'
import { formatCompactCurrency, formatCurrency, formatNumber } from '@/utils/format'
import { getAssetDashboard } from '@/plugin/asset/api/asset'
import { getAssetRiskDashboard } from '@/plugin/asset/api/risk'
import { getAssetOperationList } from '@/plugin/asset/api/operation'
import { getInvoiceDashboard } from '@/plugin/invoice/api/invoice'
import { centsToCurrency } from '@/plugin/invoice/utils/invoice'
import { getSysOperationRecordList } from '@/api/sysOperationRecord'
import { getSystemState } from '@/api/system'
import { getWorkSchedules, importLegacyWorkSchedules } from '@/api/workSchedule'
import { useUserStore } from '@/pinia/modules/user'

defineOptions({ name: 'Dashboard' })

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const moduleKeys = ['assets', 'assetDrafts', 'invoices', 'risk', 'calendar', 'audit', 'monitor']
const loading = ref(false)
const dashboardActive = ref(true)
const updatedAt = ref('')
const assetDashboard = ref(createAssetDashboard())
const invoiceDashboard = ref(createInvoiceDashboard())
const riskDashboard = ref(createRiskDashboard())
const assetDraftTotal = ref(0)
const recentOperations = ref([])
const systemState = ref(createSystemState())
const moduleLoaded = ref(createModuleState())
const moduleRefreshed = ref(createModuleState())
const moduleFailed = ref(createModuleState())
const moduleUpdatedAt = ref(createModuleState(''))
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
const assetOperationRouteNames = ['assetInbound', 'assetIssue', 'assetTransfer', 'assetReturn', 'assetMaintenance', 'assetScrap']

const access = computed(() => ({
  assetInventory: router.hasRoute('assetInventory'),
  assetOperations: assetOperationRouteNames.some((name) => router.hasRoute(name)),
  invoiceDashboard: router.hasRoute('invoiceDashboard'),
  invoiceLedger: router.hasRoute('invoiceLedger'),
  invoiceRecognition: router.hasRoute('invoiceRecognition'),
  assets: router.hasRoute('assetDashboard'),
  invoices: router.hasRoute('invoiceDashboard'),
  calendar: router.hasRoute('workSchedule'),
  audit: router.hasRoute('operation'),
  risk: router.hasRoute('assetRiskCenter'),
  monitor: router.hasRoute('state')
}))
const isPendingView = computed(() => route.query.view === 'pending')
const isWallboardView = computed(() => dashboardActive.value && route.query.view === 'wallboard')
const canOpenWallboard = computed(() => access.value.assets || access.value.invoices || access.value.risk || access.value.calendar || access.value.monitor)
const canOpenPendingTasks = computed(() => access.value.assetOperations || access.value.invoiceRecognition)
const requestedModules = computed(() => [
  access.value.assets && 'assets',
  access.value.assetOperations && 'assetDrafts',
  access.value.invoices && 'invoices',
  access.value.risk && 'risk',
  access.value.calendar && 'calendar',
  access.value.audit && 'audit',
  access.value.monitor && 'monitor'
].filter(Boolean))
const refreshedModuleCount = computed(() => requestedModules.value.filter((key) => moduleRefreshed.value[key]).length)
const availableModuleCount = computed(() => requestedModules.value.filter((key) => moduleLoaded.value[key]).length)

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
const currentTime = ref(new Date())
let clockTimer = null
function startClock() {
  if (clockTimer) clearInterval(clockTimer)
  currentTime.value = new Date()
  clockTimer = setInterval(() => {
    if (dashboardActive.value) {
      currentTime.value = new Date()
    }
  }, 1000)
}
function stopClock() {
  if (clockTimer) {
    clearInterval(clockTimer)
    clockTimer = null
  }
}
const currentTimeText = computed(() => {
  return currentTime.value.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
})
const refreshText = computed(() => {
  if (loading.value) return '正在刷新数据'
  const total = requestedModules.value.length
  const refreshed = refreshedModuleCount.value
  const available = availableModuleCount.value
  if (!total) return '当前账号暂无统计权限'
  if (!updatedAt.value) return available ? '数据已就绪' : '数据暂不可用'
  if (!refreshed) return available ? `刷新失败 · 上次成功于 ${updatedAt.value}` : '数据暂不可用'
  if (refreshed < total) return `部分更新于 ${updatedAt.value}`
  return `刷新于 ${updatedAt.value}`
})
const pendingTotal = computed(() => {
  const assetTotal = moduleLoaded.value.assetDrafts ? Number(assetDraftTotal.value || 0) : 0
  const invoiceTotal = moduleLoaded.value.invoices
    ? Number(invoiceDashboard.value.pendingCount || 0) + Number(invoiceDashboard.value.failedCount || 0)
    : 0
  return assetTotal + invoiceTotal
})
const overviewText = computed(() => {
  const counts = []
  if (access.value.assetOperations && moduleLoaded.value.assetDrafts) counts.push(`${formatNumber(assetDraftTotal.value)} 项资产业务待处理`)
  if (access.value.invoices && moduleLoaded.value.invoices) counts.push(`${formatNumber(invoiceDashboard.value.pendingCount)} 张发票待核对`)
  if (access.value.invoices && moduleLoaded.value.invoices && invoiceDashboard.value.failedCount) counts.push(`${formatNumber(invoiceDashboard.value.failedCount)} 张识别失败`)
  if (access.value.risk && moduleLoaded.value.risk) counts.push(`${formatNumber(riskDashboard.value.totalOpen)} 项开放风险`)
  if (counts.length) {
    const staleHint = requestedModules.value.some((key) => moduleFailed.value[key] && moduleLoaded.value[key]) ? ' 部分数据来自上次成功刷新。' : ''
    return `当前有 ${counts.join('，')}。${staleHint}`
  }
  if (availableModuleCount.value < requestedModules.value.length) return '部分统计数据暂不可用，可稍后刷新重试。'
  return '当前没有待处理事项，可从左侧菜单进入已获授权的业务模块。'
})
const maintenanceQuantity = computed(() => Number(assetDashboard.value.statusSummary.find((item) => item.status === 'maintenance')?.quantity || 0))
const controlledQuantity = computed(() => Math.max(Number(assetDashboard.value.totalQuantity || 0) - maintenanceQuantity.value, 0))
const healthRate = computed(() => {
  const total = Number(assetDashboard.value.totalQuantity || 0)
  return total ? ((controlledQuantity.value / total) * 100).toFixed(1) : '0.0'
})
const metrics = computed(() => {
  const items = []
  if (access.value.assets && moduleLoaded.value.assets) {
    items.push({
      label: '资产实物总量',
      numValue: Number(assetDashboard.value.totalQuantity || 0),
      format: (val) => `${formatNumber(val)} 件`,
      value: `${formatNumber(assetDashboard.value.totalQuantity)} 件`,
      hint: hintWithFreshness('assets', `${formatNumber(assetDashboard.value.categoryCount)} 个分类 · ${formatNumber(assetDashboard.value.assetKinds)} 份资产档案`),
      tone: 'primary',
      icon: Box
    })
    items.push({
      label: '资产账面原值',
      numValue: Number(assetDashboard.value.originalValue || 0),
      format: (val) => formatCompactCurrency(val),
      value: formatCompactCurrency(assetDashboard.value.originalValue),
      hint: hintWithFreshness('assets', `当前估值 ${formatCompactCurrency(assetDashboard.value.currentValue)}`),
      tone: 'success',
      icon: Coin
    })
  }
  if (access.value.invoices && moduleLoaded.value.invoices) {
    items.push({
      label: '已确认发票',
      numValue: Number(invoiceDashboard.value.totalCents || 0),
      format: (val) => centsToCurrency(val),
      value: centsToCurrency(invoiceDashboard.value.totalCents),
      hint: hintWithFreshness('invoices', `${formatNumber(invoiceDashboard.value.confirmedCount)} 张已进入正式统计`),
      tone: 'info',
      icon: Tickets
    })
  }
  if ((access.value.assetOperations && moduleLoaded.value.assetDrafts) || (access.value.invoices && moduleLoaded.value.invoices)) {
    const pendingHints = []
    if (access.value.assetOperations && moduleLoaded.value.assetDrafts) pendingHints.push(`${formatNumber(assetDraftTotal.value)} 项资产业务`)
    if (access.value.invoices && moduleLoaded.value.invoices) {
      pendingHints.push(`${formatNumber(invoiceDashboard.value.pendingCount)} 张待核对`)
      pendingHints.push(`${formatNumber(invoiceDashboard.value.failedCount)} 张失败`)
    }
    const pendingStale = (moduleFailed.value.assetDrafts && moduleLoaded.value.assetDrafts) || (moduleFailed.value.invoices && moduleLoaded.value.invoices)
    items.push({
      label: '待处理事项',
      numValue: Number(pendingTotal.value || 0),
      format: (val) => `${formatNumber(val)} 项`,
      value: `${formatNumber(pendingTotal.value)} 项`,
      hint: `${pendingHints.join(' · ')}${pendingStale ? ' · 上次成功数据' : ''}`,
      tone: 'warning',
      icon: WarningFilled,
      action: canOpenPendingTasks.value ? 'pending' : '',
      actionLabel: canOpenPendingTasks.value ? '查看待处理事项' : ''
    })
  }
  if (access.value.risk && moduleLoaded.value.risk) {
    items.push({
      label: '开放风险',
      numValue: Number(riskDashboard.value.totalOpen || 0),
      format: (val) => `${formatNumber(val)} 项`,
      value: `${formatNumber(riskDashboard.value.totalOpen)} 项`,
      hint: hintWithFreshness('risk', `高风险 ${formatNumber(riskDashboard.value.highOpen)} 项 · 今日新增 ${formatNumber(riskDashboard.value.todayNew)} 项`),
      tone: riskDashboard.value.highOpen ? 'danger' : 'success',
      icon: WarningFilled,
      action: 'risk',
      actionLabel: '查看风险中心'
    })
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
const riskTrendBars = computed(() => {
  const values = (riskDashboard.value.trend || []).slice(-7)
  const maximum = Math.max(1, ...values.flatMap((item) => [Number(item.new || 0), Number(item.resolved || 0)]))
  return values.map((item) => ({
    ...item,
    newRatio: Math.max(Number(item.new || 0) ? 10 : 0, Number(item.new || 0) / maximum * 100),
    resolvedRatio: Math.max(Number(item.resolved || 0) ? 10 : 0, Number(item.resolved || 0) / maximum * 100)
  }))
})
const systemUsage = computed(() => {
  const cpuValues = Array.isArray(systemState.value.cpu?.cpus) ? systemState.value.cpu.cpus : []
  const cpu = cpuValues.length
    ? cpuValues.reduce((total, value) => total + Number(value || 0), 0) / cpuValues.length
    : 0
  const disks = Array.isArray(systemState.value.disk) ? systemState.value.disk : []
  return {
    cpu,
    ram: Number(systemState.value.ram?.usedPercent || 0),
    disk: Math.max(0, ...disks.map((item) => Number(item?.usedPercent || 0)))
  }
})
const systemHealth = computed(() => {
  const peak = Math.max(systemUsage.value.cpu, systemUsage.value.ram, systemUsage.value.disk)
  if (peak >= 90) return { label: '资源紧张', tone: 'danger' }
  if (peak >= 75) return { label: '资源需关注', tone: 'warning' }
  return { label: '运行正常', tone: 'success' }
})
const serverCollectedAt = computed(() => {
  const date = new Date(systemState.value.collectedAt)
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
})
const todaySchedules = computed(() => calendarEvents.value
  .filter((item) => scheduleMatchesDate(item, dateKey(new Date())))
  .sort((left, right) => String(left.time).localeCompare(String(right.time)))
  .slice(0, 4)
  .map((item) => {
    const type = calendarTypes.value.find((entry) => entry.value === item.type) || defaultScheduleTypes[0]
    return { ...item, typeLabel: type.label, color: type.color, repeatLabel: recurrenceLabel(item) }
  }))
const wallboardSnapshot = computed(() => ({
  dateText: currentDateText.value,
  updatedAt: updatedAt.value,
  freshnessText: refreshText.value,
  access: access.value,
  moduleLoaded: moduleLoaded.value,
  moduleFailed: moduleFailed.value,
  moduleUpdatedAt: moduleUpdatedAt.value,
  asset: assetDashboard.value,
  assetStatusRows: assetStatusRows.value,
  invoice: invoiceDashboard.value,
  invoiceTrend: invoiceTrend.value,
  risk: riskDashboard.value,
  healthRate: healthRate.value,
  pendingTotal: pendingTotal.value,
  assetDraftTotal: assetDraftTotal.value,
  schedules: todaySchedules.value,
  systemHealth: systemHealth.value,
  systemUsage: systemUsage.value
}))
const assetStatusOrder = [
  { key: 'in_use', label: '在用', tone: 'success' },
  { key: 'idle', label: '闲置', tone: 'info' },
  { key: 'maintenance', label: '维修维保', tone: 'warning' },
  { key: 'pending_inbound', label: '待入库', tone: 'primary' },
  { key: 'retired', label: '已处置', tone: 'danger' }
]
const statusMap = Object.fromEntries(assetStatusOrder.map((item) => [item.key, item]))

function createAssetDashboard() {
  return { assetKinds: 0, totalQuantity: 0, categoryCount: 0, originalValue: 0, currentValue: 0, statusSummary: [], locationSummary: [], recentAssets: [] }
}
function createInvoiceDashboard() {
  return { confirmedCount: 0, pendingCount: 0, failedCount: 0, totalCents: 0, amountCents: 0, taxCents: 0, monthlyTrend: [] }
}
function createModuleState(defaultValue = false) {
  return Object.fromEntries(moduleKeys.map((key) => [key, defaultValue]))
}
function hintWithFreshness(key, text) {
  return moduleFailed.value[key] && moduleLoaded.value[key] ? `${text} · 上次成功数据` : text
}
function moduleFreshnessShort(key) {
  return moduleUpdatedAt.value[key] ? `上次成功 ${moduleUpdatedAt.value[key]}` : '上次成功数据'
}
function moduleFreshnessText(key) {
  return moduleUpdatedAt.value[key]
    ? `本次刷新失败，显示 ${moduleUpdatedAt.value[key]} 的上次成功数据`
    : '本次刷新失败，当前显示上次成功数据'
}
function createRiskDashboard() {
  return { totalOpen: 0, highOpen: 0, todayNew: 0, overdue: 0, trend: [], byCategory: [], bySeverity: [], byStatus: [], byCustodian: [], recentEvents: [], latestScan: null, generatedAt: '' }
}
function createSystemState() {
  return { collectedAt: '', cpu: { cpus: [] }, ram: { usedPercent: 0 }, disk: [] }
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
function safePercent(value) {
  return Math.min(100, Math.max(0, Number(value || 0)))
}
function percent(value) {
  return `${safePercent(value).toFixed(0)}%`
}
function usageTone(value) {
  const usage = safePercent(value)
  if (usage >= 90) return 'danger'
  if (usage >= 75) return 'warning'
  return 'success'
}
function go(name) {
  if (router.hasRoute(name)) router.push({ name })
}
function handleMetricClick(metric) {
  if (metric.action === 'pending') openPendingItems()
  if (metric.action === 'risk') go('assetRiskCenter')
}
function openPendingItems() {
  router.push({ path: route.path, query: { ...route.query, view: 'pending' } })
}
function closePendingItems() {
  const query = { ...route.query }
  delete query.view
  router.replace({ path: route.path, query })
}
function openWallboard() {
  router.push({ path: route.path, query: { ...route.query, view: 'wallboard' } })
}
function closeWallboard() {
  const query = { ...route.query }
  delete query.view
  router.replace({ path: route.path, query })
}
function loadCalendarTypes() {
  if (!access.value.calendar) return
  try {
    const savedTypes = JSON.parse(window.localStorage.getItem(typeStorageKey) || '[]')
    calendarTypes.value = Array.isArray(savedTypes) && savedTypes.length ? savedTypes : defaultScheduleTypes
  } catch {
    calendarTypes.value = defaultScheduleTypes
  }
}
function readLegacyCalendarSchedules() {
  try {
    const savedEvents = JSON.parse(window.localStorage.getItem(eventStorageKey) || '[]')
    return Array.isArray(savedEvents) ? savedEvents.filter((item) => item && item.id && item.title && item.date && item.time) : []
  } catch {
    return []
  }
}
function ensureCalendarTypes(schedules) {
  const knownTypes = new Set(calendarTypes.value.map((item) => item.value))
  const additionalTypes = []
  for (const schedule of schedules) {
    if (!schedule.type || knownTypes.has(schedule.type)) continue
    knownTypes.add(schedule.type)
    additionalTypes.push({
      value: schedule.type,
      label: '自定义日程',
      color: defaultScheduleTypes[(calendarTypes.value.length + additionalTypes.length) % defaultScheduleTypes.length].color
    })
  }
  if (additionalTypes.length) calendarTypes.value = [...calendarTypes.value, ...additionalTypes]
}
async function loadCalendarSchedules() {
  if (!access.value.calendar) return
  loadCalendarTypes()
  const legacySchedules = readLegacyCalendarSchedules()
  if (legacySchedules.length) {
    try {
      const imported = await importLegacyWorkSchedules({
        schedules: legacySchedules.map(({ id, ...schedule }) => ({ ...schedule, clientKey: id }))
      })
      if (imported.code === 0) window.localStorage.removeItem(eventStorageKey)
    } catch {
      // Keep the local data until a successful migration from the calendar page.
    }
  }
  try {
    const result = await getWorkSchedules()
    if (result.code === 0) {
      calendarEvents.value = (result.data || []).map((item) => ({ ...item, id: String(item.id ?? item.ID) }))
      ensureCalendarTypes(calendarEvents.value)
      return true
    }
  } catch {
    return false
  }
  return false
}
async function loadAssets() {
  const result = await getAssetDashboard()
  if (result.code === 0) {
    const data = result.data || {}
    assetDashboard.value = {
      ...createAssetDashboard(),
      ...data,
      statusSummary: data.statusSummary || [],
      locationSummary: data.locationSummary || [],
      recentAssets: data.recentAssets || []
    }
    return true
  }
  return false
}
async function loadAssetDrafts() {
  const result = await getAssetOperationList({ page: 1, pageSize: 1, status: 'draft' })
  if (result.code === 0) {
    assetDraftTotal.value = Number(result.data?.total || 0)
    return true
  }
  return false
}
async function loadInvoices() {
  const result = await getInvoiceDashboard()
  if (result.code === 0) {
    const data = result.data || {}
    invoiceDashboard.value = { ...createInvoiceDashboard(), ...data, monthlyTrend: data.monthlyTrend || [] }
    return true
  }
  return false
}
async function loadRisk() {
  const result = await getAssetRiskDashboard()
  if (result.code === 0) {
    const data = result.data || {}
    riskDashboard.value = { ...createRiskDashboard(), ...data, trend: data.trend || [] }
    return true
  }
  return false
}
async function loadAudit() {
  const result = await getSysOperationRecordList({ page: 1, pageSize: 4 })
  if (result.code === 0) {
    recentOperations.value = result.data?.list || []
    return true
  }
  return false
}
async function loadMonitor() {
  const result = await getSystemState()
  if (result.code === 0 && result.data?.server) {
    systemState.value = { ...createSystemState(), ...result.data.server }
    return true
  }
  return false
}
async function loadDashboard() {
  if (loading.value) return
  loading.value = true
  moduleRefreshed.value = createModuleState()
  moduleFailed.value = createModuleState()
  try {
    const tasks = [
      { key: 'assets', enabled: access.value.assets, run: loadAssets },
      { key: 'assetDrafts', enabled: access.value.assetOperations, run: loadAssetDrafts },
      { key: 'invoices', enabled: access.value.invoices, run: loadInvoices },
      { key: 'risk', enabled: access.value.risk, run: loadRisk },
      { key: 'calendar', enabled: access.value.calendar, run: loadCalendarSchedules },
      { key: 'audit', enabled: access.value.audit, run: loadAudit },
      { key: 'monitor', enabled: access.value.monitor, run: loadMonitor }
    ].filter((task) => task.enabled)
    const results = await Promise.allSettled(tasks.map((task) => task.run()))
    const refreshed = createModuleState()
    const failed = createModuleState()
    const available = { ...moduleLoaded.value }
    const timestamps = { ...moduleUpdatedAt.value }
    const completedAt = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
    results.forEach((result, index) => {
      const key = tasks[index].key
      const succeeded = result.status === 'fulfilled' && result.value === true
      refreshed[key] = succeeded
      failed[key] = !succeeded
      if (succeeded) {
        available[key] = true
        timestamps[key] = completedAt
      }
    })
    moduleLoaded.value = available
    moduleRefreshed.value = refreshed
    moduleFailed.value = failed
    moduleUpdatedAt.value = timestamps
    if (results.some((result) => result.status === 'fulfilled' && result.value === true)) {
      updatedAt.value = completedAt
    }
  } finally {
    loading.value = false
  }
}

let wallboardTimer = null
function syncWallboardTimer(enabled) {
  if (wallboardTimer) {
    window.clearInterval(wallboardTimer)
    wallboardTimer = null
  }
  if (enabled) {
    wallboardTimer = window.setInterval(() => {
      if (isWallboardView.value && !loading.value) loadDashboard()
    }, 60 * 1000)
  }
}

onMounted(() => {
  if (!isPendingView.value) loadDashboard()
  syncWallboardTimer(isWallboardView.value)
  startClock()
})
let activatedOnce = false
onActivated(() => {
  dashboardActive.value = true
  if (activatedOnce && !isPendingView.value) loadDashboard()
  activatedOnce = true
  syncWallboardTimer(isWallboardView.value)
  startClock()
})
onDeactivated(() => {
  dashboardActive.value = false
  syncWallboardTimer(false)
  stopClock()
})
watch(isPendingView, (pending, wasPending) => {
  if (!pending && wasPending) loadDashboard()
})
watch(isWallboardView, (wallboard, wasWallboard) => {
  syncWallboardTimer(wallboard)
  if (!wallboard && wasWallboard && dashboardActive.value && !isPendingView.value) loadDashboard()
})
onBeforeUnmount(() => {
  syncWallboardTimer(false)
  stopClock()
})
</script>

<style scoped lang="scss">
.dashboard-route-view { min-height: 100%; }
.dashboard-page { min-height: 100%; padding: 18px 20px 24px; background: var(--na-background); color: var(--na-foreground); }
.updated-at { color: var(--na-muted-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; white-space: nowrap; }
.header-primary-actions { display: flex; flex-wrap: wrap; gap: 8px; }
.header-primary-actions :deep(.el-button) { min-width: 104px; margin-left: 0; }

.cockpit-fullscreen-btn {
  font-weight: 600;
  border-color: var(--na-primary) !important;
  color: var(--na-primary) !important;
  box-shadow: 0 0 10px -2px var(--na-primary-soft);
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1) !important;
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 0 16px 0 var(--na-primary-soft);
  }
}

.workbench-band {
  position: relative;
  overflow: hidden;
  display: flex;
  min-width: 0;
  min-height: 138px;
  align-items: stretch;
  justify-content: space-between;
  gap: 28px;
  margin-bottom: var(--na-space-lg);
  padding: 20px 24px;
  border: 1px solid rgba(109, 93, 251, 0.22);
  border-radius: calc(var(--na-radius) + 2px);
  background: linear-gradient(135deg, rgba(109, 93, 251, 0.08) 0%, rgba(246, 248, 255, 0.94) 42%, rgba(250, 245, 255, 0.88) 100%), var(--na-card);
  box-shadow: 0 10px 28px -8px rgba(109, 93, 251, 0.14), 0 2px 6px 0 rgba(0, 0, 0, 0.03);
  animation: panelFadeIn 0.35s cubic-bezier(0.16, 1, 0.3, 1) both;

  &::before,
  &::after {
    content: '';
    position: absolute;
    width: 12px;
    height: 12px;
    pointer-events: none;
    z-index: 2;
  }
  &::before {
    top: 7px;
    left: 7px;
    border-top: 2px solid var(--na-primary);
    border-left: 2px solid var(--na-primary);
    border-top-left-radius: 4px;
  }
  &::after {
    top: 7px;
    right: 7px;
    border-top: 2px solid var(--na-primary);
    border-right: 2px solid var(--na-primary);
    border-top-right-radius: 4px;
  }
}

:root.dark .workbench-band {
  border-color: rgba(129, 140, 248, 0.3);
  background: radial-gradient(circle at 18% 28%, rgba(109, 93, 251, 0.25), transparent 48%),
              radial-gradient(circle at 82% 72%, rgba(59, 130, 246, 0.18), transparent 42%),
              linear-gradient(135deg, rgba(15, 23, 42, 0.96), rgba(30, 27, 75, 0.9)),
              var(--na-card);
  box-shadow: 0 12px 36px -8px rgba(0, 0, 0, 0.55), inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.workbench-hero-canvas {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  pointer-events: none;
}
.workbench-copy {
  position: relative;
  z-index: 1;
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  justify-content: center;
}
.workbench-time-row {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}
.cockpit-live-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px;
  border-radius: 12px;
  background: var(--na-primary-soft);
  border: 1px solid rgba(109, 93, 251, 0.25);
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  color: var(--na-primary);
}
.live-pulse-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--na-primary);
  box-shadow: 0 0 0 3px var(--na-primary-soft);
  animation: pulseGlow 2.4s infinite ease-in-out;
}
.current-date {
  margin: 0;
  color: var(--na-primary);
  font-size: .75rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.01em;
}
.workbench-copy h2 {
  margin: 0;
  font-size: 1.375rem;
  font-weight: 700;
  letter-spacing: -0.01em;
}
.workbench-copy > p:last-of-type {
  margin: 7px 0 0;
  color: var(--na-muted-foreground);
  font-size: .8125rem;
  line-height: 1.5;
}
.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 8px;
  margin-top: 10px;
}
.quick-actions :deep(.el-button) {
  height: 26px;
  padding: 0 6px;
  font-size: .75rem;
  transition: all 0.2s ease;
  &:hover {
    transform: translateY(-1px);
    color: var(--na-primary);
  }
}

.runtime-summary {
  position: relative;
  z-index: 1;
  display: grid;
  width: min(390px, 38%);
  min-width: 330px;
  grid-template-columns: 1fr;
  gap: 10px;
  padding: 0 0 0 26px;
  border: 0;
  border-left: 1px solid var(--na-border);
  background: transparent;
  color: var(--na-foreground);
  text-align: left;
  transition: border-color 0.2s ease;
}
.runtime-summary.is-actionable { cursor: pointer; }
.runtime-summary.is-actionable:hover .runtime-heading { color: var(--na-primary); }
.runtime-summary:disabled { cursor: default; opacity: 1; }

.runtime-ecg-track {
  position: relative;
  width: 100%;
  height: 14px;
  overflow: hidden;
  opacity: 0.85;
}
.runtime-ecg-svg {
  width: 100%;
  height: 100%;
}
.runtime-ecg-path {
  fill: none;
  stroke: var(--na-primary);
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-dasharray: 50 100;
  animation: ecgFlow 2.8s linear infinite;
}
@keyframes ecgFlow {
  0% { stroke-dashoffset: 150; }
  100% { stroke-dashoffset: 0; }
}

.runtime-topline { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 16px; }
.runtime-topline > small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.runtime-heading {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: .75rem;
  font-weight: 600;
  transition: color 160ms ease;
  white-space: nowrap;
}
.runtime-heading i {
  position: relative;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--na-success);
  box-shadow: 0 0 0 3px var(--na-success-soft);
  animation: pulseGlow 2.2s infinite ease-in-out;
}
.runtime-heading.health-warning i { background: var(--na-warning); box-shadow: 0 0 0 3px var(--na-warning-soft); }
.runtime-heading.health-danger i { background: var(--na-danger); box-shadow: 0 0 0 3px var(--na-warning-soft); animation: pulseDanger 1.6s infinite; }
.runtime-summary dl { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 26px; margin: 0; }
.runtime-summary dl div { display: grid; min-width: 0; grid-template-columns: 1fr auto; align-items: end; gap: 4px 8px; }
.runtime-summary dt { color: var(--na-muted-foreground); font-size: .6875rem; }
.runtime-summary dd { margin: 0; color: var(--na-foreground); font-size: 1rem; font-variant-numeric: tabular-nums; font-weight: 700; }
.runtime-track { grid-column: 1 / -1; height: 5px; overflow: hidden; border-radius: 3px; background: var(--na-muted); }
.runtime-track i {
  display: block;
  height: 100%;
  border-radius: inherit;
  transition: width 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}

.metric-band {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
  gap: 12px;
  margin-bottom: var(--na-space-lg);
  overflow: visible;
  animation: panelFadeIn 0.4s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.metric-item {
  position: relative;
  display: flex;
  min-width: 0;
  min-height: 108px;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 18px 20px 16px;
  border: 1px solid var(--na-border);
  border-radius: var(--na-radius);
  background: var(--na-card);
  box-shadow: 0 2px 6px 0 rgba(0, 0, 0, 0.03);
  overflow: hidden;
  transition: transform 0.22s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.22s cubic-bezier(0.16, 1, 0.3, 1), border-color 0.22s ease;
  animation: metricItemEntrance 0.45s cubic-bezier(0.16, 1, 0.3, 1) both;
  animation-delay: var(--metric-delay, 0ms);

  &:hover {
    transform: translateY(-3px);
    box-shadow: 0 10px 24px -4px var(--metric-glow, rgba(99, 102, 241, 0.16));
    border-color: var(--metric-border, var(--na-primary));
    .metric-icon {
      transform: scale(1.12) rotate(4deg);
    }
  }
}
button.metric-item { width: 100%; color: inherit; font: inherit; text-align: left; }
.metric-item--actionable { cursor: pointer; }
.metric-item--actionable:focus-visible { position: relative; z-index: 1; outline: 2px solid var(--na-primary); outline-offset: -2px; }

.metric-accent-stripe {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--metric-accent-color, var(--na-primary));
  opacity: 0.85;
}

.metric-title-row {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.metric-status-glow {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--metric-accent-color, var(--na-primary));
  box-shadow: 0 0 0 2px var(--metric-glow, var(--na-primary-soft));
  animation: pulseGlow 2.5s infinite;
}

.metric-copy { display: flex; min-width: 0; flex: 1; flex-direction: column; justify-content: center; gap: 4px; }
.metric-copy > span, .metric-title-row > span, .metric-copy small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.metric-copy strong { overflow: hidden; color: var(--na-foreground); font-size: 1.35rem; font-variant-numeric: tabular-nums; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; margin-top: 2px; }
.metric-copy small.is-warning { color: var(--na-warning); }

.metric-icon-box {
  display: inline-grid;
  width: 38px;
  height: 38px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: var(--na-radius-sm);
  background: var(--metric-soft-bg, var(--na-primary-soft));
  transition: transform 0.22s ease, box-shadow 0.22s ease;
}
.metric-icon {
  font-size: 1.15rem;
  color: var(--metric-accent-color, var(--na-primary));
  transition: transform 0.22s ease;
}

.metric-item--primary {
  --metric-accent-color: var(--na-primary);
  --metric-border: var(--na-primary);
  --metric-glow: rgba(99, 102, 241, 0.22);
  --metric-soft-bg: var(--na-primary-soft);
}
.metric-item--success {
  --metric-accent-color: var(--na-success);
  --metric-border: var(--na-success);
  --metric-glow: rgba(16, 185, 129, 0.22);
  --metric-soft-bg: var(--na-success-soft);
}
.metric-item--info {
  --metric-accent-color: #06B6D4;
  --metric-border: #06B6D4;
  --metric-glow: rgba(6, 182, 212, 0.22);
  --metric-soft-bg: rgba(6, 182, 212, 0.12);
}
.metric-item--warning {
  --metric-accent-color: var(--na-warning);
  --metric-border: var(--na-warning);
  --metric-glow: rgba(245, 158, 11, 0.25);
  --metric-soft-bg: var(--na-warning-soft);
}
.metric-item--danger {
  --metric-accent-color: var(--na-danger);
  --metric-border: var(--na-danger);
  --metric-glow: rgba(239, 68, 68, 0.25);
  --metric-soft-bg: var(--na-danger-soft);
}

.dashboard-workspace {
  display: grid;
  min-width: 0;
  grid-template-columns: minmax(0, 1.65fr) minmax(340px, .84fr);
  gap: var(--na-space-md);
  animation: panelFadeIn 0.45s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.business-column, .support-column { display: grid; min-width: 0; align-content: start; gap: var(--na-space-md); }
.dashboard-panel {
  min-width: 0;
  overflow: hidden;
  transition: box-shadow 0.22s ease, border-color 0.22s ease;
  &:hover {
    box-shadow: 0 4px 18px -4px rgba(0, 0, 0, 0.05);
  }
}
.panel-heading > div { min-width: 0; }
.panel-heading span { display: block; color: var(--na-muted-foreground); font-size: .6875rem; }
.panel-heading h2 { margin: 3px 0 0; color: var(--na-foreground); font-size: .9375rem; font-weight: 600; }
.panel-heading :deep(.el-button) { font-size: .75rem; }
.panel-placeholder, .inline-empty { display: grid; min-height: 84px; place-items: center; color: var(--na-muted-foreground); font-size: .75rem; }
.module-stale-notice { margin: 0; padding: 8px 16px; border-bottom: 1px solid var(--na-border); background: var(--na-warning-soft); color: var(--na-warning); font-size: .75rem; line-height: 1.5; }

.asset-summary { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); margin: 0; padding: 15px 20px; border-bottom: 1px solid var(--na-border); }
.asset-summary div { min-width: 0; padding: 0 20px; border-left: 1px solid var(--na-border); }
.asset-summary div:first-child { padding-left: 0; border-left: 0; }
.asset-summary div:last-child { padding-right: 0; }
.asset-summary dt, .asset-summary small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.asset-summary dd { overflow: hidden; margin: 6px 0 3px; color: var(--na-foreground); font-size: 1.125rem; font-variant-numeric: tabular-nums; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.asset-detail-grid { display: grid; min-width: 0; grid-template-columns: minmax(320px, .96fr) minmax(360px, 1.04fr); }
.asset-status-section, .asset-recent-section { min-width: 0; }
.asset-status-section { border-right: 1px solid var(--na-border); }
.section-mini-heading { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 12px; padding: 13px 20px 10px; }
.section-mini-heading > span { color: var(--na-foreground); font-size: .75rem; font-weight: 600; }
.section-mini-heading small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.asset-status-list { display: grid; gap: 11px; padding: 8px 20px 18px; }
.asset-status-row { display: grid; min-width: 0; grid-template-columns: 78px minmax(0, 1fr) 42px; align-items: center; gap: 10px; }
.status-label { display: inline-flex; align-items: center; gap: 6px; color: var(--na-muted-foreground); font-size: .75rem; white-space: nowrap; }
.status-label i { width: 6px; height: 6px; border-radius: 50%; background: var(--na-muted-foreground); }
.progress-track { height: 5px; overflow: hidden; border-radius: 3px; background: var(--na-muted); }
.progress-track > i {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: var(--na-primary);
  transition: width 0.8s cubic-bezier(0.16, 1, 0.3, 1);
  position: relative;
  overflow: hidden;
  &::after {
    content: '';
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.35), transparent);
    animation: trackShimmer 3.2s infinite ease-in-out;
  }
}
.tone-success { background: var(--na-success) !important; }
.tone-warning { background: var(--na-warning) !important; }
.tone-danger { background: var(--na-danger) !important; }
.tone-info { background: var(--na-info) !important; }
.tone-primary { background: var(--na-primary) !important; }
.asset-status-row > strong { color: var(--na-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; text-align: right; }
.asset-recent-table-head { display: grid; margin: 0 20px; padding: 8px 12px; grid-template-columns: minmax(0, 1.25fr) minmax(90px, .8fr) minmax(92px, .6fr); gap: 12px; border-radius: 6px; background: var(--na-table-header); color: var(--na-muted-foreground); font-size: .6875rem; }
.asset-recent-table-head span:last-child { text-align: right; }
.asset-recent-list { display: grid; }
.asset-recent-list button {
  display: grid;
  min-width: 0;
  min-height: 52px;
  grid-template-columns: minmax(0, 1.25fr) minmax(90px, .8fr) minmax(92px, .6fr);
  align-items: center;
  gap: 12px;
  margin: 0 20px;
  padding: 7px 12px;
  border: 0;
  border-bottom: 1px solid var(--na-border);
  background: transparent;
  color: var(--na-foreground);
  text-align: left;
  transition: background-color 0.16s ease, transform 0.16s ease;
  &:hover {
    background: var(--na-table-hover);
    transform: translateX(2px);
  }
}
.asset-recent-list button:last-child { border-bottom: 0; }
.operation-list button:hover, .schedule-list button:hover {
  background: var(--na-table-hover);
  transform: translateX(2px);
}
.asset-identity, .asset-place { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.asset-identity strong, .asset-place span { overflow: hidden; font-size: .75rem; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.asset-identity small, .asset-place small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.asset-place small.status-success { color: var(--na-success); }
.asset-place small.status-warning { color: var(--na-warning); }
.asset-place small.status-danger { color: var(--na-danger); }
.asset-place small.status-info { color: var(--na-info); }
.asset-place small.status-primary { color: var(--na-primary); }
.asset-recent-list b { overflow: hidden; color: var(--na-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; text-align: right; text-overflow: ellipsis; white-space: nowrap; }

.invoice-workspace { display: grid; min-width: 0; grid-template-columns: minmax(230px, .8fr) minmax(0, 1.7fr); }
.invoice-total { display: flex; min-width: 0; flex-direction: column; justify-content: center; padding: 20px; background: var(--na-primary-soft); }
.invoice-total > span, .invoice-total > small, .invoice-breakdown dt { color: var(--na-muted-foreground); font-size: .6875rem; }
.invoice-total > strong { overflow: hidden; margin: 7px 0 3px; color: var(--na-primary); font-size: 1.375rem; font-variant-numeric: tabular-nums; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.invoice-breakdown { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; margin: 15px 0 0; padding-top: 13px; border-top: 1px solid var(--na-ring); }
.invoice-breakdown div { min-width: 0; }
.invoice-breakdown dd { overflow: hidden; margin: 5px 0 0; color: var(--na-foreground); font-size: .75rem; font-variant-numeric: tabular-nums; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.trend-section { min-width: 0; padding: 15px 20px 12px; }
.invoice-trend-heading { display: flex; min-width: 0; align-items: flex-start; justify-content: space-between; gap: 14px; }
.invoice-trend-heading > div:first-child { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.invoice-trend-heading > div:first-child > span { color: var(--na-foreground); font-size: .75rem; font-weight: 600; }
.invoice-trend-heading small { color: var(--na-muted-foreground); font-size: .6875rem; }
.invoice-exceptions { display: flex; flex: 0 0 auto; gap: 6px; }
.invoice-exceptions span { display: inline-flex; min-height: 24px; align-items: center; padding: 0 8px; border-radius: 6px; font-size: .6875rem; font-weight: 600; }
.invoice-exceptions .is-warning { color: var(--na-warning); background: var(--na-warning-soft); }
.invoice-exceptions .is-danger { color: var(--na-danger); background: var(--na-danger-soft); }
.invoice-trend { display: grid; height: 124px; grid-template-columns: repeat(6, minmax(0, 1fr)); align-items: end; gap: 10px; padding: 9px 0 0; }
.trend-item {
  display: grid;
  min-width: 0;
  height: 100%;
  grid-template-rows: 16px minmax(0, 1fr) 15px;
  align-items: end;
  gap: 4px;
  cursor: default;
  &:hover {
    .trend-value { color: var(--na-primary); font-weight: 700; }
    .trend-bar i { transform: scaleX(1.12); filter: brightness(1.18); }
  }
}
.trend-value { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-align: center; text-overflow: ellipsis; white-space: nowrap; transition: color 0.16s ease; }
.trend-bar { display: flex; height: 100%; align-items: end; justify-content: center; border-bottom: 1px solid var(--na-border); }
.trend-bar i {
  width: min(32px, 62%);
  min-height: 2px;
  border-radius: 4px 4px 0 0;
  background: var(--na-primary);
  transform-origin: bottom center;
  transition: height 0.75s cubic-bezier(0.16, 1, 0.3, 1), transform 0.2s ease, filter 0.2s ease;
}
.trend-item small { color: var(--na-muted-foreground); font-size: .6875rem; text-align: center; }

.schedule-list, .operation-list { display: grid; }
.schedule-list button, .operation-list button {
  display: grid;
  min-width: 0;
  min-height: 48px;
  align-items: center;
  border: 0;
  border-bottom: 1px solid var(--na-border);
  background: transparent;
  color: var(--na-foreground);
  text-align: left;
  transition: background-color 0.16s ease, transform 0.16s ease;
}
.schedule-list button { grid-template-columns: 6px 42px minmax(0, 1fr); gap: 10px; padding: 6px 16px; }
.schedule-list button:last-child, .operation-list button:last-child { border-bottom: 0; }
.schedule-list i { width: 6px; height: 24px; border-radius: 3px; }
.schedule-list time { color: var(--na-muted-foreground); font-size: .6875rem; font-variant-numeric: tabular-nums; }
.schedule-list span { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.schedule-list strong { overflow: hidden; font-size: .75rem; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.schedule-list small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.side-empty { display: grid; min-height: 102px; place-items: center; align-content: center; gap: 7px; color: var(--na-muted-foreground); font-size: .75rem; }
.side-empty :deep(svg) { width: 22px; height: 22px; color: var(--na-primary); }
.schedule-footer { display: block; overflow: hidden; padding: 8px 16px; border-top: 1px solid var(--na-border); color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.operation-table-head, .operation-list button { display: grid; min-width: 0; grid-template-columns: 42px minmax(0, 1fr) 40px 34px; align-items: center; gap: 8px; }
.operation-table-head { margin: 12px 14px 4px; padding: 7px 8px; border-radius: 6px; background: var(--na-table-header); color: var(--na-muted-foreground); font-size: .6875rem; }
.operation-table-head span:last-child { text-align: center; }
.operation-list button { min-height: 40px; padding: 0 22px; border: 0; border-bottom: 1px solid var(--na-border); }
.request-method { overflow: hidden; color: var(--na-primary); font: 600 .6875rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; text-overflow: ellipsis; white-space: nowrap; }
.request-path { overflow: hidden; font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.operation-list time { color: var(--na-muted-foreground); font-size: .6875rem; font-variant-numeric: tabular-nums; }
.operation-list i { display: inline-grid; min-width: 30px; place-items: center; border-radius: 4px; color: var(--na-on-primary); font-size: .6875rem; font-style: normal; font-variant-numeric: tabular-nums; line-height: 18px; }
.operation-list i.request-ok { color: var(--na-success); background: var(--na-success-soft); }
.operation-list i.request-error { color: var(--na-danger); background: var(--na-danger-soft); }
.daily-risk-summary { display: grid; grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr); border-bottom: 1px solid var(--na-border); }
.daily-risk-total { display: flex; min-width: 0; flex-direction: column; justify-content: center; padding: var(--na-space-md); border-right: 1px solid var(--na-border); }
.daily-risk-total span, .daily-risk-total small, .daily-risk-summary dt { color: var(--na-muted-foreground); font-size: .75rem; }
.daily-risk-total strong {
  display: inline-flex;
  align-items: center;
  margin: 5px 0 3px;
  color: var(--na-foreground);
  font-size: 1.5rem;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  line-height: 1.15;
}
.daily-risk-total strong.is-danger { color: var(--na-danger); }
.daily-risk-total strong.is-success { color: var(--na-success); }
.danger-pulse-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  margin-left: 6px;
  border-radius: 50%;
  background: var(--na-danger);
  box-shadow: 0 0 0 3px var(--na-danger-soft);
  animation: pulseDanger 1.8s infinite;
}
.daily-risk-summary dl { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: center; margin: 0; padding: var(--na-space-md) var(--na-space-sm); }
.daily-risk-summary dl div { min-width: 0; padding: 0 var(--na-space-xs); }
.daily-risk-summary dd { margin: 5px 0 0; color: var(--na-foreground); font-size: 1rem; font-variant-numeric: tabular-nums; font-weight: 600; }
.daily-risk-summary dd.is-warning { color: var(--na-warning); }
.daily-risk-trend { display: flex; height: 72px; align-items: stretch; gap: var(--na-space-xs); padding: var(--na-space-sm) var(--na-space-md); }
.daily-risk-trend > span {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: end;
  justify-content: center;
  gap: 2px;
  border-bottom: 1px solid var(--na-border);
  transition: background-color 0.16s ease;
  &:hover {
    background: var(--na-muted);
  }
}
.daily-risk-trend i {
  width: min(7px, 34%);
  min-height: 2px;
  border-radius: 2px 2px 0 0;
  transition: height 0.75s cubic-bezier(0.16, 1, 0.3, 1);
}
.daily-risk-trend .risk-new { background: var(--na-danger); }
.daily-risk-trend .risk-resolved { background: var(--na-success); }

/* Animations */
@keyframes panelFadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes metricItemEntrance {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulseGlow {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 0 0 3px var(--na-primary-soft);
  }
  50% {
    transform: scale(1.15);
    box-shadow: 0 0 0 6px var(--na-primary-soft);
  }
}

@keyframes pulseDanger {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 0 0 3px var(--na-danger-soft);
  }
  50% {
    transform: scale(1.22);
    box-shadow: 0 0 0 6px var(--na-danger-soft);
  }
}

@keyframes trackShimmer {
  0% {
    left: -100%;
  }
  50%, 100% {
    left: 100%;
  }
}

@media (max-width: 1120px) {
  .dashboard-workspace { grid-template-columns: 1fr; }
  .support-column { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 860px) {
  .dashboard-page { padding: 15px; }
  .workbench-band { flex-direction: column; }
  .runtime-summary { width: 100%; min-width: 0; padding: 14px 0 0; border-top: 1px solid var(--na-border); border-left: 0; }
  .metric-band { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .metric-item:nth-child(2n) { border-right: 0; }
  .metric-item:nth-child(-n + 2) { border-bottom: 1px solid var(--na-border); }
}
@media (max-width: 720px) {
  .asset-detail-grid, .invoice-workspace { grid-template-columns: 1fr; }
  .asset-status-section { border-right: 0; border-bottom: 1px solid var(--na-border); }
  .support-column { grid-template-columns: 1fr; }
}
@media (max-width: 640px) {
  .dashboard-page { padding: 12px; }
  .workbench-band { padding: 15px; }
  .runtime-summary dl { gap: 12px; }
  .runtime-topline { align-items: flex-start; flex-direction: column; gap: 4px; }
  .asset-summary { grid-template-columns: 1fr; gap: 12px; }
  .asset-summary div, .asset-summary div:first-child, .asset-summary div:last-child { padding: 0; border-left: 0; }
  .asset-recent-table-head, .asset-recent-list button { grid-template-columns: minmax(0, 1fr) auto; }
  .asset-recent-table-head span:nth-child(2), .asset-place { display: none; }
  .asset-recent-list button { grid-template-columns: minmax(0, 1fr) auto; }
  .invoice-trend-heading { align-items: flex-start; flex-direction: column; }
  .operation-list button { grid-template-columns: 38px minmax(0, 1fr) 34px; }
  .operation-table-head { grid-template-columns: 38px minmax(0, 1fr) 34px; }
  .operation-table-head span:nth-child(3) { display: none; }
  .operation-list time { display: none; }
  .metric-band { grid-template-columns: 1fr; }
  .metric-item, .metric-item:nth-child(2n), .metric-item:nth-child(-n + 2) { min-height: 86px; border-right: 0; border-bottom: 1px solid var(--na-border); }
  .metric-item:last-child { border-bottom: 0; }
  .daily-risk-summary { grid-template-columns: 1fr; }
  .daily-risk-total { border-right: 0; border-bottom: 1px solid var(--na-border); }
}
@media (prefers-reduced-motion: reduce) {
  .workbench-band, .metric-band, .dashboard-workspace, .metric-item { animation: none; }
  .progress-track > i, .runtime-track i, .trend-bar i { transition: none; }
  .progress-track > i::after { animation: none; display: none; }
  .live-pulse-dot, .danger-pulse-dot, .runtime-heading i { animation: none; }
}
</style>
