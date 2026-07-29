<template>
  <main class="na-page na-page--list invoice-dashboard">
    <AppPageHeader title-id="invoice-dashboard-title" title="流水总览" description="仅统计已人工确认的发票，待核对与失败任务单独列示。">
      <template #actions>
        <span class="refresh-time">更新于 {{ updateTime || '—' }}</span>
        <el-button :icon="Refresh" :loading="loading" @click="loadDashboard">刷新</el-button>
        <el-button type="primary" :icon="Upload" @click="router.push({ name: 'invoiceRecognition' })">上传发票</el-button>
      </template>
    </AppPageHeader>

    <el-skeleton v-if="loading && !loaded" :rows="8" animated />
    <el-result v-else-if="error && !loaded" icon="error" title="流水数据加载失败" :sub-title="error" class="dashboard-error">
      <template #extra><el-button type="primary" :icon="Refresh" @click="loadDashboard">重新加载</el-button></template>
    </el-result>
    <template v-else>
      <div v-if="error" class="load-warning" role="alert">
        <span>刷新失败，当前仍显示上一次成功数据：{{ error }}</span>
        <el-button text :icon="Refresh" @click="loadDashboard">重试</el-button>
      </div>
      <section class="summary-band" aria-label="流水关键指标">
        <div class="primary-total">
          <span>已确认价税合计</span>
          <strong>{{ money(dashboard.totalCents) }}</strong>
          <small>{{ dashboard.confirmedCount }} 张发票已进入正式统计</small>
        </div>
        <dl class="summary-metrics">
          <div><dt>不含税金额</dt><dd>{{ money(dashboard.amountCents) }}</dd></div>
          <div><dt>税额</dt><dd>{{ money(dashboard.taxCents) }}</dd></div>
          <div class="is-warning"><dt>待人工核对</dt><dd>{{ dashboard.pendingCount }}</dd></div>
          <div class="is-danger"><dt>识别失败</dt><dd>{{ dashboard.failedCount }}</dd></div>
        </dl>
      </section>

      <div class="dashboard-grid">
        <section class="na-panel trend-panel">
          <div class="na-panel-header panel-heading">
            <div><h2>近 12 个月流水趋势</h2><p>按开票日期汇总已确认金额</p></div>
            <span>{{ money(dashboard.totalCents) }}</span>
          </div>
          <div v-if="hasTrend" class="chart-wrap"><Chart :options="trendOption" height="310px" /></div>
          <el-empty v-else description="确认发票后将生成月度趋势" :image-size="82" />
        </section>

        <section class="na-panel category-panel">
          <div class="na-panel-header panel-heading">
            <div><h2>分类构成</h2><p>规则推荐经人工确认后的结果</p></div>
          </div>
          <div v-if="dashboard.categories.length" class="category-layout">
            <Chart :options="categoryOption" height="210px" />
            <div class="category-list">
              <div v-for="item in dashboard.categories.slice(0, 6)" :key="item.categoryId">
                <i :style="{ backgroundColor: item.color || 'var(--na-primary)' }" />
                <span>{{ item.name }}</span>
                <strong>{{ money(item.totalCents) }}</strong>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无分类统计" :image-size="76" />
        </section>
      </div>

      <div class="detail-grid">
        <section class="na-panel supplier-panel">
          <div class="na-panel-header panel-heading">
            <div><h2>供应商支出排行</h2><p>按已确认价税合计排序</p></div>
          </div>
          <div v-if="dashboard.suppliers.length" class="supplier-list">
            <div v-for="(item, index) in dashboard.suppliers" :key="item.name">
              <span class="rank">{{ String(index + 1).padStart(2, '0') }}</span>
              <div><strong>{{ item.name }}</strong><small>{{ item.count }} 张发票</small></div>
              <b>{{ money(item.totalCents) }}</b>
            </div>
          </div>
          <el-empty v-else description="暂无供应商统计" :image-size="72" />
        </section>

        <section class="na-panel recent-panel">
          <div class="na-panel-header panel-heading">
            <div><h2>最近处理</h2><p>包含识别、核对和确认状态</p></div>
            <el-button text @click="router.push({ name: 'invoiceLedger' })">查看台账</el-button>
          </div>
          <div v-if="dashboard.recent.length" class="recent-list">
            <button v-for="item in dashboard.recent" :key="item.ID" type="button" @click="openReview(item)">
              <div class="recent-identity">
                <strong>{{ item.sellerName || item.fileName }}</strong>
                <small>{{ item.invoiceNumber || '号码待核对' }} · {{ dateText(item.issueDate || item.CreatedAt) }}</small>
              </div>
              <span>{{ money(item.totalCents) }}</span>
              <InvoiceStatusTag :status="item.status" />
            </button>
          </div>
          <el-empty v-else description="上传第一张发票开始处理" :image-size="72" />
        </section>
      </div>
    </template>

    <InvoiceReviewDrawer v-model="reviewVisible" :invoice-id="selectedId" @saved="loadDashboard" @confirmed="loadDashboard" />
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { Refresh, Upload } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import Chart from '@/components/charts/index.vue'
import { chartPalette, chartTheme } from '@/components/charts/theme'
import InvoiceReviewDrawer from '@/plugin/invoice/components/InvoiceReviewDrawer.vue'
import InvoiceStatusTag from '@/plugin/invoice/components/InvoiceStatusTag.vue'
import { getInvoiceDashboard } from '@/plugin/invoice/api/invoice'
import { centsToCurrency, invoiceDateText } from '@/plugin/invoice/utils/invoice'
import { useAppStore } from '@/pinia/modules/app'

defineOptions({ name: 'InvoiceDashboard' })

const router = useRouter()
const appStore = useAppStore()
const loading = ref(false)
const loaded = ref(false)
const error = ref('')
const updateTime = ref('')
const reviewVisible = ref(false)
const selectedId = ref(0)
const dashboard = ref({
  confirmedCount: 0, pendingCount: 0, failedCount: 0,
  totalCents: 0, amountCents: 0, taxCents: 0,
  monthlyTrend: [], categories: [], suppliers: [], recent: []
})

const money = centsToCurrency
const dateText = invoiceDateText
const hasTrend = computed(() => dashboard.value.monthlyTrend.some((item) => Number(item.totalCents || 0) > 0))

const trendOption = computed(() => {
  appStore.isDark
  appStore.config.primaryColor
  const theme = chartTheme()
  return {
    aria: { enabled: true, description: '近十二个月已确认发票金额趋势' },
    color: [theme.primary],
    tooltip: { trigger: 'axis', valueFormatter: money, backgroundColor: theme.surface, borderColor: theme.grid, textStyle: { color: theme.text } },
    grid: { left: 20, right: 18, top: 24, bottom: 18, containLabel: true },
    xAxis: {
      type: 'category', boundaryGap: false,
      data: dashboard.value.monthlyTrend.map((item) => item.month.slice(5) + '月'),
      axisLine: { lineStyle: { color: theme.grid } }, axisTick: { show: false }, axisLabel: { color: theme.label }
    },
    yAxis: {
      type: 'value', axisLine: { show: false }, axisTick: { show: false },
      splitLine: { lineStyle: { color: theme.grid } }, axisLabel: { color: theme.label, formatter: (value) => `¥${Math.round(value / 10000) / 100}万` }
    },
    series: [{
      type: 'line', smooth: true, symbol: 'circle', symbolSize: 7,
      data: dashboard.value.monthlyTrend.map((item) => Number(item.totalCents || 0)),
      lineStyle: { width: 3 }, areaStyle: { opacity: 0.1 }, emphasis: { focus: 'series' }
    }]
  }
})

const categoryOption = computed(() => {
  appStore.isDark
  appStore.config.primaryColor
  const theme = chartTheme()
  return {
    aria: { enabled: true, description: '已确认发票分类金额构成' },
    color: dashboard.value.categories.map((item) => item.color).filter(Boolean).concat(chartPalette()),
    tooltip: { trigger: 'item', formatter: (item) => `${item.name}<br/>${money(item.value)}`, backgroundColor: theme.surface, borderColor: theme.grid, textStyle: { color: theme.text } },
    series: [{
      type: 'pie', radius: ['58%', '78%'], center: ['50%', '50%'], avoidLabelOverlap: true,
      label: { show: false }, itemStyle: { borderWidth: 3, borderColor: theme.surface },
      data: dashboard.value.categories.map((item) => ({ name: item.name, value: Number(item.totalCents || 0) }))
    }]
  }
})

const loadDashboard = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await getInvoiceDashboard()
    if (res.code === 0) {
      dashboard.value = {
        ...dashboard.value,
        ...(res.data || {}),
        monthlyTrend: res.data?.monthlyTrend || [],
        categories: res.data?.categories || [],
        suppliers: res.data?.suppliers || [],
        recent: res.data?.recent || []
      }
      updateTime.value = new Date().toLocaleTimeString('zh-CN', { hour12: false, hour: '2-digit', minute: '2-digit' })
      loaded.value = true
    } else {
      error.value = res.msg || '无法读取流水统计，请稍后重试'
    }
  } catch (requestError) {
    error.value = requestError?.message || '无法读取流水统计，请稍后重试'
  } finally {
    loading.value = false
  }
}

const openReview = (item) => {
  selectedId.value = Number(item.ID)
  reviewVisible.value = true
}

onMounted(loadDashboard)
</script>

<style scoped lang="scss">
.dashboard-error { min-height: 420px; }
.load-warning { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; padding: 9px 12px; border-radius: 9px; background: var(--na-warning-soft); color: var(--na-foreground); font-size: .75rem; }
.load-warning span { min-width: 0; overflow-wrap: anywhere; }
.refresh-time { color: var(--na-muted-foreground); font-size: .75rem; }
.summary-band { display: grid; min-width: 0; grid-template-columns: minmax(260px, 1.1fr) minmax(0, 2fr); overflow: hidden; margin-bottom: 14px; border: 1px solid var(--na-border); border-radius: 12px; background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.primary-total { display: flex; min-width: 0; flex-direction: column; justify-content: center; padding: 22px 24px; background: var(--na-primary-soft); }
.primary-total span, .primary-total small, .summary-metrics dt { color: var(--na-muted-foreground); font-size: .75rem; }
.primary-total strong { overflow: hidden; margin: 7px 0 5px; color: var(--na-primary); font-size: 1.75rem; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.summary-metrics { display: grid; min-width: 0; grid-template-columns: repeat(4, 1fr); margin: 0; }
.summary-metrics > div { display: flex; min-width: 0; flex-direction: column; justify-content: center; gap: 7px; padding: 18px; border-left: 1px solid var(--na-border); }
.summary-metrics dd { margin: 0; overflow: hidden; color: var(--na-foreground); font-size: 1.125rem; font-variant-numeric: tabular-nums; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.summary-metrics .is-warning dd { color: var(--na-warning); }
.summary-metrics .is-danger dd { color: var(--na-danger); }
.dashboard-grid, .detail-grid { display: grid; min-width: 0; gap: 14px; }
.dashboard-grid { grid-template-columns: minmax(0, 1.65fr) minmax(300px, .85fr); }
.detail-grid { grid-template-columns: minmax(320px, .9fr) minmax(0, 1.4fr); margin-top: 14px; }
.na-panel { min-width: 0; overflow: hidden; }
.panel-heading > div { min-width: 0; }
.panel-heading h2 { margin: 0; font-size: .9375rem; font-weight: 650; }
.panel-heading p { margin: 3px 0 0; color: var(--na-muted-foreground); font-size: .75rem; }
.panel-heading > span { color: var(--na-primary); font-size: .875rem; font-variant-numeric: tabular-nums; font-weight: 650; }
.chart-wrap { padding: 6px 10px 12px; }
.category-layout { padding: 4px 14px 14px; }
.category-list { display: grid; }
.category-list > div { display: grid; min-height: 34px; grid-template-columns: 8px minmax(0, 1fr) auto; align-items: center; gap: 8px; border-top: 1px solid var(--na-border); }
.category-list i { width: 7px; height: 7px; border-radius: 50%; }
.category-list span { overflow: hidden; font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.category-list strong { font-size: .75rem; font-variant-numeric: tabular-nums; }
.supplier-list { padding: 2px 16px 12px; }
.supplier-list > div { display: grid; min-height: 54px; grid-template-columns: 28px minmax(0, 1fr) auto; align-items: center; gap: 10px; border-bottom: 1px solid var(--na-border); }
.rank { color: var(--na-primary); font: 600 .6875rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; }
.supplier-list div > div { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.supplier-list strong, .recent-identity strong { overflow: hidden; font-size: .8125rem; text-overflow: ellipsis; white-space: nowrap; }
.supplier-list small, .recent-identity small { color: var(--na-muted-foreground); font-size: .6875rem; }
.supplier-list b { font-size: .75rem; font-variant-numeric: tabular-nums; }
.recent-list { display: grid; padding: 2px 16px 12px; }
.recent-list button { display: grid; min-width: 0; min-height: 54px; grid-template-columns: minmax(0, 1fr) 110px 90px; align-items: center; gap: 12px; padding: 0; border: 0; border-bottom: 1px solid var(--na-border); background: transparent; color: var(--na-foreground); text-align: left; }
.recent-list button:hover { background: var(--na-table-hover); }
.recent-identity { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.recent-list button > span { font-size: .75rem; font-variant-numeric: tabular-nums; font-weight: 600; text-align: right; }

@media (max-width: 1180px) {
  .summary-band, .dashboard-grid, .detail-grid { grid-template-columns: 1fr; }
  .summary-metrics > div:first-child { border-left: 0; }
}
@media (max-width: 720px) {
  .summary-metrics { grid-template-columns: 1fr 1fr; }
  .summary-metrics > div:nth-child(3) { border-top: 1px solid var(--na-border); border-left: 0; }
  .summary-metrics > div:nth-child(4) { border-top: 1px solid var(--na-border); }
  .recent-list button { grid-template-columns: minmax(0, 1fr) auto; padding: 8px 0; }
  .recent-list button > span { text-align: left; }
  .recent-list :deep(.invoice-status-tag) { grid-column: 2; grid-row: 1 / span 2; }
}
</style>
