<template>
  <main class="na-page na-page--list smart-report-page">
    <AppPageHeader title-id="smart-report-title" title="智能日报" description="按日汇总异常、待办、趋势和系统健康指标；模型不可用时仍保留结构化数据。">
      <template #actions><el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button><el-button type="primary" :icon="MagicStick" :loading="generating" @click="generate">生成今日日报</el-button></template>
    </AppPageHeader>
    <AppEmptyState
      v-if="!loading && !report"
      title="今日尚未生成智能日报"
      description="生成后会汇总资产、风险、发票、协作和 AI 服务指标；模型不可用时仍保留确定性统计。"
      :highlights="['异常与待办集中汇总', '业务指标可追溯跳转', '支持应用内订阅提醒']"
    >
      <template #actions><el-button type="primary" :icon="MagicStick" :loading="generating" @click="generate">生成今日日报</el-button></template>
    </AppEmptyState>
    <section v-if="report" class="na-panel report-hero">
      <div class="report-hero__content">
        <span class="eyebrow">{{ report.reportDate?.slice?.(0, 10) || '今日' }}</span>
        <div class="report-hero__summary" aria-label="日报摘要">
          <h2 class="report-hero__title">{{ heroReportHeading }}</h2>
          <AssistantMarkdown v-if="heroReportBody" :source="heroReportBody" />
          <p v-else class="report-hero__empty">日报暂无摘要。</p>
        </div>
        <small>生成方式：{{ generationLabel(report.generatedBy) }} · {{ formatTime(report.generatedAt) }}</small>
      </div>
      <div class="report-hero__actions">
        <el-tag type="success" effect="plain">已生成</el-tag>
        <el-button type="primary" plain :icon="Document" @click="showReport(report)">阅读完整日报</el-button>
        <el-dropdown :disabled="Boolean(exportingFormat)" @command="downloadReport($event, report)">
          <el-button :icon="Download" :loading="Boolean(exportingFormat)">
            下载日报
            <el-icon v-if="!exportingFormat" class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="item in REPORT_EXPORT_FORMATS" :key="item.value" :command="item.value">
                <el-icon><component :is="exportFormatIcon(item.value)" /></el-icon>{{ item.label }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </section>
    <section v-if="report" class="metric-grid"><article v-for="item in metricCards" :key="item.key" class="na-panel metric-card" role="button" tabindex="0" @click="openMetric(item)" @keydown.enter.space.prevent="openMetric(item)"><span>{{ item.label }}</span><strong>{{ item.value }}</strong><small>{{ item.hint }}</small></article></section>
    <section v-if="report" class="na-panel detail-panel"><header class="panel-header"><div><h2>详细指标</h2><p>指标来自业务表确定性统计，模型仅负责摘要。</p></div></header><div class="detail-grid"><div v-for="item in detailMetrics" :key="item.key" class="detail-item"><span>{{ item.label }}</span><strong>{{ item.value }}</strong></div></div></section>
    <div class="report-grid">
      <section class="na-panel history-panel">
        <header class="panel-header"><div><h2>历史日报</h2><p>点击行或“查看日报”，阅读完整摘要和业务指标。</p></div></header>
        <el-table v-loading="loading && !loaded" :data="reports" row-key="ID" class="report-table" @row-click="openReport">
          <el-table-column prop="reportDate" label="日期" min-width="120"><template #default="{ row }">{{ row.reportDate?.slice?.(0, 10) }}</template></el-table-column>
          <el-table-column label="生成方式" width="150"><template #default="{ row }">{{ generationLabel(row.generatedBy) }}</template></el-table-column>
          <el-table-column label="摘要" min-width="280">
            <template #default="{ row }"><span class="history-summary">{{ row.summary || '日报暂无摘要' }}</span></template>
          </el-table-column>
          <el-table-column label="操作" width="112" fixed="right" align="center">
            <template #default="{ row }"><el-button type="primary" link :icon="View" :loading="detailLoading && openingReportId === row.ID" @click.stop="openReport(row)">查看日报</el-button></template>
          </el-table-column>
          <template #empty><AppEmptyState compact title="暂无历史日报" description="生成第一份日报后，可在这里按日期回看业务摘要。" /></template>
        </el-table>
        <div class="pagination"><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="loadHistory" /></div>
      </section>
      <section class="na-panel subscription-panel"><header class="panel-header"><div><h2>日报订阅</h2><p>保留应用内提醒，发送时间可按用户调整。</p></div><el-switch v-model="subscription.enabled" active-text="启用" /></header><el-form label-position="top"><el-form-item label="发送时间"><el-time-picker v-model="deliveryTime" format="HH:mm" value-format="HH:mm" placeholder="选择时间" /></el-form-item><el-form-item label="渠道"><el-checkbox-group v-model="channels"><el-checkbox label="in_app">应用内</el-checkbox><el-checkbox label="email">邮件</el-checkbox></el-checkbox-group></el-form-item><el-button type="primary" :loading="saving" @click="saveSubscription">保存订阅</el-button></el-form></section>
    </div>
    <section class="na-panel deliveries-panel">
      <header class="panel-header delivery-header">
        <div><h2>发送记录</h2><p>最近 30 条投递结果，按发送时间倒序排列。</p></div>
        <div v-if="deliveries.length" class="delivery-summary" aria-label="投递结果汇总">
          <span>共 {{ deliverySummary.total }} 条</span>
          <span class="is-success"><i />成功 {{ deliverySummary.sent }}</span>
          <span v-if="deliverySummary.sending" class="is-sending"><i />发送中 {{ deliverySummary.sending }}</span>
          <span v-if="deliverySummary.failed" class="is-failed"><i />失败 {{ deliverySummary.failed }}</span>
        </div>
      </header>
      <el-table :data="deliveries" size="small" class="delivery-table" :row-class-name="deliveryRowClass">
        <el-table-column label="发送渠道" width="160">
          <template #default="{ row }">
            <div class="delivery-channel">
              <el-icon><component :is="deliveryChannelMeta(row.channel).icon" /></el-icon>
              <span><strong>{{ deliveryChannelMeta(row.channel).label }}</strong><small>{{ deliveryChannelMeta(row.channel).hint }}</small></span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="投递状态" width="126">
          <template #default="{ row }"><el-tag :type="deliveryStatusMeta(row.status).type" :effect="row.status === 'sent' ? 'plain' : 'light'">{{ deliveryStatusMeta(row.status).label }}</el-tag></template>
        </el-table-column>
        <el-table-column label="投递结果" min-width="280">
          <template #default="{ row }">
            <div class="delivery-result" :class="`is-${row.status}`">
              <strong>{{ deliveryResultTitle(row) }}</strong>
              <span :title="row.status === 'failed' ? deliveryResultDescription(row) : ''">{{ deliveryResultDescription(row) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="尝试" width="118"><template #default="{ row }"><span class="delivery-attempt">{{ deliveryAttemptLabel(row) }}</span></template></el-table-column>
        <el-table-column label="发送时间" width="176">
          <template #default="{ row }"><time class="delivery-time" :datetime="row.sentAt || row.lastAttempt"><strong>{{ formatDeliveryDate(row.sentAt || row.lastAttempt) }}</strong><span>{{ formatDeliveryClock(row.sentAt || row.lastAttempt) }}</span></time></template>
        </el-table-column>
        <template #empty><AppEmptyState compact title="暂无日报发送记录" description="启用订阅并生成日报后，投递结果会显示在这里。" /></template>
      </el-table>
    </section>

    <el-drawer v-model="detailVisible" :size="drawerSize" append-to-body destroy-on-close :close-on-click-modal="false" class="smart-report-drawer">
      <template #header>
        <div class="report-detail-header">
          <div class="report-detail-heading">
            <span>{{ detailReport?.reportDate?.slice?.(0, 10) || '智能日报' }}</span>
            <h2>智能日报详情</h2>
            <p v-if="detailReport">{{ generationLabel(detailReport.generatedBy) }} · 生成于 {{ formatTime(detailReport.generatedAt) }}</p>
          </div>
          <el-dropdown v-if="detailReport" :disabled="Boolean(exportingFormat)" @command="downloadReport($event, detailReport)">
            <el-button :icon="Download" :loading="Boolean(exportingFormat)">
              下载
              <el-icon v-if="!exportingFormat" class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="item in REPORT_EXPORT_FORMATS" :key="item.value" :command="item.value">
                  <el-icon><component :is="exportFormatIcon(item.value)" /></el-icon>{{ item.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </template>
      <el-skeleton v-if="detailLoading" animated :rows="10" class="report-detail-skeleton" />
      <div v-else-if="detailReport" class="report-reader">
        <section class="report-reader__section report-reader__summary">
          <div class="report-section-heading"><h3>日报正文</h3><el-tag type="success" effect="plain">已生成</el-tag></div>
          <AssistantMarkdown :source="detailReportSummary" />
        </section>
        <section v-for="group in detailMetricGroups" :key="group.key" class="report-reader__section">
          <div class="report-section-heading"><h3>{{ group.label }}</h3><span>{{ group.description }}</span></div>
          <dl class="report-metric-list">
            <div v-for="item in group.items" :key="item.key"><dt>{{ item.label }}</dt><dd>{{ item.value }}</dd></div>
          </dl>
        </section>
      </div>
      <AppEmptyState v-else compact title="日报详情不可用" description="请关闭后重新选择日报。" />
    </el-drawer>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowDown, Bell, Document, Download, Grid, MagicStick, Message, Refresh, Tickets, View } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import AssistantMarkdown from '@/plugin/smart/components/AssistantMarkdown.vue'
import { generateSmartReport, getSmartReport, getSmartReportDeliveries, getSmartReportSubscription, getSmartReports, getTodaySmartReport, saveSmartReportSubscription } from '@/plugin/smart/api/smart'
import { exportSmartReport, formatMicrosMoney, REPORT_EXPORT_FORMATS } from '@/plugin/smart/utils/reportExport'
import { normalizeReportSummary, reportSummaryBody, reportSummaryHeading } from '@/plugin/smart/utils/reportSummary'
import { useAppStore } from '@/pinia'

defineOptions({ name: 'SmartReport' })
const router = useRouter()
const appStore = useAppStore()
const loading = ref(false); const loaded = ref(false); const generating = ref(false); const saving = ref(false); const report = ref(null); const reports = ref([]); const deliveries = ref([]); const total = ref(0); const page = ref(1); const pageSize = ref(10); const subscription = reactive({ enabled: true, deliveryTime: '09:00', channels: 'in_app' }); const deliveryTime = ref('09:00'); const channels = ref(['in_app'])
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailReport = ref(null)
const openingReportId = ref(null)
const exportingFormat = ref('')
const drawerSize = computed(() => appStore.drawerSize === '100%' ? '100%' : 'min(92vw, 780px)')
const heroReportHeading = computed(() => reportSummaryHeading(report.value?.summary))
const heroReportBody = computed(() => reportSummaryBody(report.value?.summary))
const detailReportSummary = computed(() => normalizeReportSummary(detailReport.value?.summary || '日报暂无正文。'))
const deliverySummary = computed(() => deliveries.value.reduce((summary, item) => {
  summary.total += 1
  if (Object.hasOwn(summary, item.status)) summary[item.status] += 1
  return summary
}, { total: 0, sent: 0, sending: 0, failed: 0 }))
const metricCards = computed(() => { const m = report.value?.metrics || {}; return [{ key: 'assets', label: '资产新增', value: m.assets?.created || 0, hint: `今日完成流转 ${m.assets?.todayOperationTotal || 0} 单`, route: 'assetInventory' }, { key: 'risks', label: '开放风险', value: m.risks?.open || 0, hint: `今日新增 ${m.risks?.new || 0} · 处理 ${m.risks?.resolved || 0}`, route: 'assetRiskCenter' }, { key: 'invoices', label: '待复核发票', value: m.invoices?.pendingReview || 0, hint: `低置信度 ${m.invoices?.lowConfidence || 0} · 失败 ${m.invoices?.recognitionFailed || 0}`, route: 'invoiceLedger', query: { status: 'pending_review' } }, { key: 'collaboration', label: '今日待办', value: m.collaboration?.todaySchedules || 0, hint: `未读公告 ${m.collaboration?.unreadAnnouncements || 0} 条`, route: 'workSchedule' }, { key: 'system', label: 'AI 调用', value: m.system?.aiCalls || 0, hint: `失败率 ${percentage(m.system?.aiFailureRate)}`, route: 'aiOperations' }] })
const detailMetrics = computed(() => { const m = report.value?.metrics || {}; return [{ key: 'pending', label: '待入库资产', value: m.assets?.pendingInbound || 0 }, { key: 'long-use', label: '长期在用', value: m.assets?.longTermInUse || 0 }, { key: 'maintenance', label: '维修超期', value: m.assets?.maintenanceOverdue || 0 }, { key: 'warranty-60', label: '60 天内过保', value: m.assets?.warrantyExpiring60d || 0 }, { key: 'warranty-90', label: '90 天内过保', value: m.assets?.warrantyExpiring90d || 0 }, { key: 'recognized', label: '今日识别发票', value: m.invoices?.todayRecognized || 0 }, { key: 'reviewed', label: '今日复核发票', value: m.invoices?.todayReviewed || 0 }, { key: 'confirmed', label: '今日确认发票', value: m.invoices?.todayConfirmed || 0 }, { key: 'backlog', label: '识别积压', value: m.invoices?.recognitionBacklog || 0 }, { key: 'today-amount', label: '今日确认金额', value: money(m.invoices?.confirmedTodayCents) }, { key: 'month-amount', label: '本月确认金额', value: money(m.invoices?.confirmedMonthCents) }, { key: 'ai-duration', label: 'AI 平均耗时', value: `${m.system?.aiAverageDurationMs || 0} ms` }, { key: 'ai-cost', label: 'AI 估算费用', value: formatMicrosMoney(m.system?.aiEstimatedCostMicros) }] })
const detailMetricGroups = computed(() => reportMetricGroups(detailReport.value))
function reportMetricGroups(reportValue) {
  const m = reportValue?.metrics || {}
  const operations = m.assets?.todayOperations || {}
  return [
    {
      key: 'assets', label: '资产运营', description: '资产状态、到期风险和当日流转', items: [
        { key: 'created', label: '当日新增', value: count(m.assets?.created) },
        { key: 'operation-total', label: '完成流转', value: count(m.assets?.todayOperationTotal) },
        { key: 'inbound', label: '入库', value: count(operations.inbound) },
        { key: 'issue', label: '领用', value: count(operations.issue) },
        { key: 'transfer', label: '调拨', value: count(operations.transfer) },
        { key: 'return', label: '退库', value: count(operations.return) },
        { key: 'maintenance-operation', label: '维修流转', value: count(operations.maintenance) },
        { key: 'scrap', label: '报废', value: count(operations.scrap) },
        { key: 'pending', label: '待入库资产', value: count(m.assets?.pendingInbound) },
        { key: 'long-use', label: '长期在用', value: count(m.assets?.longTermInUse) },
        { key: 'maintenance', label: '维修超期', value: count(m.assets?.maintenanceOverdue) },
        { key: 'warranty-30', label: '30 天内过保', value: count(m.assets?.warrantyExpiring30d) },
        { key: 'warranty-60', label: '60 天内过保', value: count(m.assets?.warrantyExpiring60d) },
        { key: 'warranty-90', label: '90 天内过保', value: count(m.assets?.warrantyExpiring90d) }
      ]
    },
    {
      key: 'risks', label: '风险处置', description: '开放风险及当日处理进度', items: [
        { key: 'open', label: '开放风险', value: count(m.risks?.open) },
        { key: 'new', label: '当日新增', value: count(m.risks?.new) },
        { key: 'resolved', label: '当日处理', value: count(m.risks?.resolved) }
      ]
    },
    {
      key: 'invoices', label: '发票处理', description: '识别、复核、确认和费用汇总', items: [
        { key: 'uploaded', label: '当日上传', value: count(m.invoices?.todayUploaded) },
        { key: 'recognized', label: '当日识别', value: count(m.invoices?.todayRecognized) },
        { key: 'reviewed', label: '当日复核', value: count(m.invoices?.todayReviewed) },
        { key: 'confirmed', label: '当日确认', value: count(m.invoices?.todayConfirmed) },
        { key: 'pending-review', label: '待复核', value: count(m.invoices?.pendingReview) },
        { key: 'low-confidence', label: '低置信度', value: count(m.invoices?.lowConfidence) },
        { key: 'failed', label: '识别失败', value: count(m.invoices?.recognitionFailed) },
        { key: 'backlog', label: '识别积压', value: count(m.invoices?.recognitionBacklog) },
        { key: 'provider-failure', label: '供应商失败率', value: percentage(m.invoices?.providerFailureRate) },
        { key: 'today-amount', label: '当日确认金额', value: money(m.invoices?.confirmedTodayCents) },
        { key: 'week-amount', label: '本周确认金额', value: money(m.invoices?.confirmedWeekCents) },
        { key: 'month-amount', label: '本月确认金额', value: money(m.invoices?.confirmedMonthCents) }
      ]
    },
    {
      key: 'collaboration', label: '协作提醒', description: '个人日程与公告阅读状态', items: [
        { key: 'schedules', label: '当日日程', value: count(m.collaboration?.todaySchedules) },
        { key: 'announcements', label: '未读公告', value: count(m.collaboration?.unreadAnnouncements) }
      ]
    },
    {
      key: 'system', label: 'AI 服务', description: '当日调用质量、耗时与成本', items: [
        { key: 'calls', label: '调用次数', value: count(m.system?.aiCalls) },
        { key: 'failures', label: '失败次数', value: count(m.system?.aiFailures) },
        { key: 'failure-rate', label: '失败率', value: percentage(m.system?.aiFailureRate) },
        { key: 'duration', label: '平均耗时', value: `${count(m.system?.aiAverageDurationMs)} ms` },
        { key: 'cost', label: '估算费用', value: formatMicrosMoney(m.system?.aiEstimatedCostMicros) }
      ]
    }
  ]
}
function percentage(value) { return `${Number(value || 0).toFixed(1)}%` }
function money(cents) { return `¥${(Number(cents || 0) / 100).toFixed(2)}` }
function count(value) { return Number(value || 0).toLocaleString('zh-CN') }
function generationLabel(value) {
  if (value === 'deterministic+model') return '业务统计 + AI 摘要'
  if (value === 'deterministic') return '业务统计'
  return value || '业务统计'
}
function exportFormatIcon(format) {
  if (format === 'xlsx') return Grid
  if (format === 'md') return Tickets
  return Document
}
async function downloadReport(format, reportValue) {
  if (!reportValue || exportingFormat.value) return
  exportingFormat.value = format
  try {
    const filename = await exportSmartReport(reportValue, reportMetricGroups(reportValue), format, {
      generationLabel: generationLabel(reportValue.generatedBy)
    })
    ElMessage.success(`${filename} 已开始下载`)
  } catch (error) {
    ElMessage.error(error?.message || '日报导出失败，请稍后重试')
  } finally {
    exportingFormat.value = ''
  }
}
function openMetric(item) { if (item.route && router.hasRoute(item.route)) router.push({ name: item.route, query: item.query }) }
function formatTime(value) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—' }
function deliveryChannelMeta(channel) {
  if (channel === 'email') return { label: '邮件', hint: '账户绑定邮箱', icon: Message }
  if (channel === 'in_app') return { label: '应用内', hint: '站内通知中心', icon: Bell }
  return { label: channel || '未知渠道', hint: '未识别的投递渠道', icon: Bell }
}
function deliveryStatusMeta(status) {
  if (status === 'sent') return { label: '发送成功', type: 'success' }
  if (status === 'sending') return { label: '发送中', type: 'warning' }
  if (status === 'failed') return { label: '发送失败', type: 'danger' }
  return { label: status || '未知', type: 'info' }
}
function deliveryResultTitle(row) {
  if (row.status === 'sent') return row.channel === 'email' ? '邮件已发送' : '应用内通知已送达'
  if (row.status === 'sending') return '正在执行投递'
  if (row.status === 'failed') return '本次投递未完成'
  return '等待投递结果'
}
function deliveryResultDescription(row) {
  if (row.status === 'sent') return row.channel === 'email' ? '日报已发送至账户绑定邮箱' : '日报提醒已进入站内通知中心'
  if (row.status === 'sending') return `系统正在执行第 ${Math.max(Number(row.retryCount) || 1, 1)} 次投递`
  if (row.status === 'failed') return row.error || '系统将在下一调度周期继续尝试'
  return '暂未收到渠道返回结果'
}
function deliveryAttemptLabel(row) {
  const attempts = Number(row.retryCount) || 0
  if (attempts === 0) return '尚未尝试'
  if (row.status === 'sent' && attempts === 1) return '首次成功'
  if (row.status === 'sent') return `重试 ${attempts - 1} 次`
  if (row.status === 'failed' && attempts >= 3) return '已达上限'
  return `第 ${attempts} 次`
}
function deliveryDate(value) {
  const date = value ? new Date(value) : null
  return date && !Number.isNaN(date.getTime()) ? date : null
}
function formatDeliveryDate(value) {
  const date = deliveryDate(value)
  if (!date) return '—'
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
function formatDeliveryClock(value) {
  const date = deliveryDate(value)
  return date ? date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false }) : '时间未知'
}
function deliveryRowClass({ row }) { return row?.status ? `delivery-row--${row.status}` : '' }
function showReport(value) { detailReport.value = value; detailVisible.value = true }
async function loadToday() { const res = await getTodaySmartReport(); if (res.code === 0) report.value = res.data; else ElMessage.error(res.msg || '读取日报失败') }
async function loadHistory() { const res = await getSmartReports({ page: page.value, pageSize: pageSize.value }); if (res.code === 0) { reports.value = res.data?.list || []; total.value = res.data?.total || 0 } }
async function loadSubscription() { const res = await getSmartReportSubscription(); if (res.code === 0) { Object.assign(subscription, res.data || {}); deliveryTime.value = res.data?.deliveryTime || '09:00'; channels.value = (res.data?.channels || 'in_app').split(',').filter(Boolean) } }
async function loadDeliveries() { const res = await getSmartReportDeliveries(); if (res.code === 0) deliveries.value = res.data || [] }
async function load() { loading.value = true; try { await Promise.all([loadToday(), loadHistory(), loadSubscription(), loadDeliveries()]) } finally { loading.value = false; loaded.value = true } }
async function generate() { generating.value = true; try { const res = await generateSmartReport(); if (res.code === 0) { report.value = res.data; await loadHistory(); ElMessage.success(res.msg || '日报已生成') } else ElMessage.error(res.msg || '生成失败') } finally { generating.value = false } }
async function saveSubscription() { saving.value = true; try { const res = await saveSmartReportSubscription({ enabled: subscription.enabled, deliveryTime: deliveryTime.value, channels: channels.value.join(',') }); if (res.code === 0) { ElMessage.success(res.msg || '订阅已保存'); await loadDeliveries() } else ElMessage.error(res.msg || '保存失败') } finally { saving.value = false } }
async function openReport(row) {
  const reportId = row?.ID
  if (!reportId || detailLoading.value) return
  openingReportId.value = reportId
  detailReport.value = null
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await getSmartReport({ id: reportId })
    if (res.code === 0) detailReport.value = res.data
    else ElMessage.error(res.msg || '读取日报详情失败')
  } catch {
    ElMessage.error('读取日报详情失败，请稍后重试')
  } finally {
    detailLoading.value = false
    openingReportId.value = null
  }
}
onMounted(load)
</script>

<style scoped lang="scss">
.report-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 14px;
  padding: 20px;
}

.report-hero__content {
  min-width: 0;
  flex: 1 1 auto;
}

.report-hero__actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 10px;
}

.eyebrow { color: var(--el-color-primary); font-size: .75rem; }

.report-hero__summary {
  max-width: 920px;
  max-height: 8rem;
  margin: 8px 0 10px;
  overflow: hidden;
}

.report-hero__title {
  margin: 0 0 6px;
  color: var(--na-foreground);
  font-size: 1.05rem;
  line-height: 1.45;
}

.report-hero__empty { margin: 0; color: var(--na-muted-foreground); font-size: .84rem; }

.report-hero__summary :deep(.assistant-markdown) {
  color: var(--na-foreground);
  font-size: .84rem;
  line-height: 1.58;
}

.report-hero__summary :deep(h1) {
  margin: 0 0 6px;
  font-size: 1.05rem;
  line-height: 1.45;
}

.report-hero__summary :deep(h2),
.report-hero__summary :deep(h3),
.report-hero__summary :deep(h4) {
  margin: 6px 0 3px;
  font-size: .875rem;
  line-height: 1.45;
}

.report-hero__summary :deep(p),
.report-hero__summary :deep(ul),
.report-hero__summary :deep(ol) { margin-bottom: 4px; }

.report-hero small,
.panel-header p,
.metric-card small { color: var(--na-muted-foreground); }

.metric-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 14px;
}

.metric-card {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  cursor: pointer;
  transition: border-color .16s ease, box-shadow .16s ease;
}

.metric-card:hover,
.metric-card:focus-visible {
  border-color: var(--el-color-primary);
  box-shadow: var(--na-shadow-md);
  outline: none;
}

.metric-card span { color: var(--na-muted-foreground); font-size: .78rem; }
.metric-card strong { font-size: 1.45rem; font-variant-numeric: tabular-nums; }
.metric-card small { min-height: 30px; font-size: .72rem; line-height: 1.4; }
.detail-panel { margin-bottom: 14px; }

.detail-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  border-top: 1px solid var(--na-border);
  border-left: 1px solid var(--na-border);
}

.detail-item {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
  padding: 11px 12px;
  border-right: 1px solid var(--na-border);
  border-bottom: 1px solid var(--na-border);
}

.detail-item span { color: var(--na-muted-foreground); font-size: .72rem; }
.detail-item strong { overflow-wrap: anywhere; font-size: .9rem; font-variant-numeric: tabular-nums; }
.report-grid { display: grid; grid-template-columns: minmax(0, 1fr) 320px; gap: 14px; }

.na-panel {
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--na-border);
  border-radius: 8px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
}

.panel-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.panel-header h2 { margin: 0; font-size: .95rem; }
.panel-header p { margin: 5px 0 0; font-size: .75rem; }
.subscription-panel { align-self: start; }
.subscription-panel .el-time-picker { width: 100%; }
.deliveries-panel { margin-top: 14px; }
.delivery-header { align-items: center; }

.delivery-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 14px;
  padding: 7px 10px;
  border-radius: 6px;
  background: var(--na-muted);
  color: var(--na-muted-foreground);
  font-size: .75rem;
  font-variant-numeric: tabular-nums;
}

.delivery-summary span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  white-space: nowrap;
}

.delivery-summary i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.delivery-summary .is-success { color: var(--el-color-success); }
.delivery-summary .is-sending { color: var(--el-color-warning); }
.delivery-summary .is-failed { color: var(--el-color-danger); }

.delivery-table :deep(.el-table__header th.el-table__cell) {
  height: 44px;
  background: var(--na-muted);
}

.delivery-table :deep(.el-table__row td.el-table__cell) { padding: 11px 0; }
.delivery-table :deep(.delivery-row--failed > td.el-table__cell) { background: var(--el-color-danger-light-9); }
.delivery-table :deep(.delivery-row--sending > td.el-table__cell) { background: var(--el-color-warning-light-9); }

.delivery-channel {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.delivery-channel > .el-icon {
  width: 30px;
  height: 30px;
  flex: 0 0 30px;
  border-radius: 6px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: .9rem;
}

.delivery-channel > span,
.delivery-result,
.delivery-time {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
}

.delivery-channel strong,
.delivery-result strong,
.delivery-time strong {
  color: var(--na-foreground);
  font-size: .8rem;
  font-weight: 650;
  line-height: 1.35;
}

.delivery-channel small,
.delivery-result span,
.delivery-time span {
  color: var(--na-muted-foreground);
  font-size: .7rem;
  line-height: 1.4;
}

.delivery-result span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.delivery-result.is-failed strong,
.delivery-result.is-failed span { color: var(--el-color-danger); }
.delivery-result.is-sending strong { color: var(--el-color-warning-dark-2); }
.delivery-attempt { color: var(--na-muted-foreground); font-size: .75rem; white-space: nowrap; }
.delivery-time { font-style: normal; font-variant-numeric: tabular-nums; }
.pagination { display: flex; justify-content: flex-end; margin-top: 14px; }
.report-table :deep(.el-table__row) { cursor: pointer; }
.history-summary {
  display: -webkit-box;
  overflow: hidden;
  color: var(--na-muted-foreground);
  line-height: 1.5;
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.report-detail-header {
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-right: 8px;
}

.report-detail-heading { min-width: 0; }

.report-detail-heading > span {
  display: block;
  margin-bottom: 5px;
  color: var(--el-color-primary);
  font-size: .75rem;
  font-variant-numeric: tabular-nums;
}

.report-detail-heading h2 {
  margin: 0;
  color: var(--na-foreground);
  font-size: 1.1rem;
  line-height: 1.4;
}

.report-detail-heading p {
  margin: 5px 0 0;
  color: var(--na-muted-foreground);
  font-size: .75rem;
}

.report-detail-skeleton { padding: 6px 2px; }
.report-reader { color: var(--na-foreground); }

.report-reader__section {
  padding: 22px 0;
  border-bottom: 1px solid var(--na-border);
}

.report-reader__section:first-child { padding-top: 4px; }
.report-reader__section:last-child { border-bottom: 0; }

.report-reader__summary :deep(.assistant-markdown) {
  max-width: 72ch;
  font-size: .9375rem;
  line-height: 1.8;
}

.report-section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.report-section-heading h3 {
  margin: 0;
  color: var(--na-foreground);
  font-size: .9375rem;
  line-height: 1.5;
}

.report-section-heading > span {
  color: var(--na-muted-foreground);
  font-size: .75rem;
  line-height: 1.5;
  text-align: right;
}

.report-metric-list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 0;
  border-top: 1px solid var(--na-border);
  border-left: 1px solid var(--na-border);
}

.report-metric-list > div {
  min-width: 0;
  padding: 11px 12px;
  border-right: 1px solid var(--na-border);
  border-bottom: 1px solid var(--na-border);
  background: var(--na-card);
}

.report-metric-list dt {
  color: var(--na-muted-foreground);
  font-size: .72rem;
  line-height: 1.4;
}

.report-metric-list dd {
  margin: 5px 0 0;
  color: var(--na-foreground);
  font-size: .9rem;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

:global(.smart-report-drawer .el-drawer__header) {
  margin-bottom: 0;
  padding: 20px 22px 16px;
  border-bottom: 1px solid var(--na-border);
}

:global(.smart-report-drawer .el-drawer__body) { padding: 18px 22px 28px; }

@media (max-width: 1100px) {
  .metric-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .detail-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .report-grid { grid-template-columns: 1fr; }
}

@media (max-width: 650px) {
  .metric-grid,
  .detail-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .report-hero { flex-direction: column; }
  .report-hero__summary { max-height: 10rem; }
  .report-hero__actions { width: 100%; flex-wrap: wrap; }
  .metric-card strong { font-size: 1.2rem; }
  .report-section-heading { flex-direction: column; gap: 4px; }
  .report-section-heading > span { text-align: left; }
  .report-metric-list { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .delivery-header { align-items: flex-start; }
  .delivery-summary { width: 100%; }
  .report-detail-header { align-items: flex-start; flex-direction: column; gap: 10px; }
  :global(.smart-report-drawer .el-drawer__header) { padding: 16px; }
  :global(.smart-report-drawer .el-drawer__body) { padding: 16px; }
}

@media (prefers-reduced-motion: reduce) {
  .metric-card { transition: none; }
}
</style>
