<template>
  <main class="na-page na-page--list quality-page">
    <AppPageHeader title-id="invoice-quality-title" title="发票识别质量" description="按识别任务、字段修正和分类采纳情况查看质量闭环。">
      <template #actions>
        <el-button type="primary" plain :icon="Refresh" :loading="loading" @click="loadQuality">刷新数据</el-button>
      </template>
    </AppPageHeader>

    <section class="scope-bar" aria-label="分析范围">
      <div class="scope-heading">
        <span class="scope-icon" aria-hidden="true"><el-icon><Filter /></el-icon></span>
        <div><strong>分析范围</strong><small>{{ hasFilters ? '已按条件筛选质量数据' : '当前展示全部识别数据' }}</small></div>
      </div>
      <el-date-picker v-model="dateRange" class="date-filter" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" />
      <el-input v-model="filters.provider" clearable placeholder="筛选 Provider" class="filter-input" />
      <el-input v-model="filters.model" clearable placeholder="筛选模型版本" class="filter-input" />
      <el-button v-if="hasFilters" text class="clear-filter" @click="clearQualityFilters">重置</el-button>
    </section>

    <el-result v-if="error && !loaded" icon="error" title="质量数据加载失败" :sub-title="error">
      <template #extra><el-button type="primary" :icon="Refresh" @click="loadQuality">重新加载</el-button></template>
    </el-result>
    <template v-else>
      <div v-if="error" class="load-warning" role="alert">{{ error }}<el-button text :icon="Refresh" @click="loadQuality">重试</el-button></div>
      <section class="metric-band" aria-label="识别质量指标">
        <article v-for="item in summaryCards" :key="item.label" class="metric-card" :class="item.tone">
          <span class="metric-label"><i aria-hidden="true" />{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
          <small>{{ item.note }}</small>
        </article>
      </section>

      <div class="quality-grid">
        <section class="na-panel quality-panel">
          <div class="panel-heading">
            <div><h2>Provider 与模型</h2><p>按文件类型拆分识别成功率、置信度和修正量</p></div>
            <span class="record-count">{{ providerMetrics.length }} 组</span>
          </div>
          <el-table :data="providerMetrics" class="quality-table provider-table" size="small">
            <el-table-column label="Provider" min-width="170">
              <template #default="{ row }"><div class="provider-name"><i aria-hidden="true" /><strong>{{ row.provider || '未标记' }}</strong></div></template>
            </el-table-column>
            <el-table-column label="模型" min-width="145" show-overflow-tooltip>
              <template #default="{ row }"><span class="model-name">{{ row.model || '默认模型' }}</span></template>
            </el-table-column>
            <el-table-column label="文件类型" min-width="118"><template #default="{ row }"><span class="file-type">{{ row.fileType || '—' }}</span></template></el-table-column>
            <el-table-column prop="total" label="总量" width="72" align="right" />
            <el-table-column label="成功率" width="138" align="right">
              <template #default="{ row }"><div class="rate-cell"><strong :class="accuracyTone(row.successRate)">{{ percent(row.successRate) }}</strong><span class="rate-track"><i :style="{ width: `${safeRate(row.successRate)}%` }" /></span></div></template>
            </el-table-column>
            <el-table-column label="平均置信度" width="112" align="right"><template #default="{ row }"><span class="numeric-value">{{ confidencePercent(row.averageConfidence) }}</span></template></el-table-column>
            <el-table-column label="修正字段" width="92" align="right"><template #default="{ row }"><span :class="['correction-count', { 'has-correction': row.correctedFields }]">{{ row.correctedFields || 0 }}</span></template></el-table-column>
            <template #empty>
              <AppEmptyState compact :title="hasFilters ? '当前筛选范围没有识别数据' : '暂无识别质量数据'" description="完成发票识别后，将按 Provider、模型和文件类型形成质量统计。">
                <template v-if="hasFilters" #actions><el-button :icon="Refresh" @click="clearQualityFilters">清除筛选</el-button></template>
              </AppEmptyState>
            </template>
          </el-table>
        </section>

        <section class="na-panel classification-panel">
          <div class="panel-heading"><div><h2>分类推荐效果</h2><p>只统计已有建议分类的任务</p></div></div>
          <div class="classification-overview">
            <div class="classification-total"><span>分类建议</span><strong>{{ classification.suggested }}</strong><small>累计任务</small></div>
            <div class="acceptance-rate"><span>采纳率</span><strong>{{ percent(classification.acceptanceRate) }}</strong></div>
          </div>
          <div class="classification-track" aria-label="分类建议处理分布">
            <i class="is-accepted" :style="{ width: classificationShare(classification.accepted) }" />
            <i class="is-overridden" :style="{ width: classificationShare(classification.overridden) }" />
            <i class="is-pending" :style="{ width: classificationShare(classification.pending) }" />
          </div>
          <dl class="classification-list">
            <div v-for="item in classificationRows" :key="item.label">
              <dt><i :class="item.tone" aria-hidden="true" /><span>{{ item.label }}</span><small>{{ item.share }}</small></dt>
              <dd>{{ item.value }}</dd>
            </div>
          </dl>
        </section>
      </div>

      <section class="na-panel quality-panel field-panel">
        <div class="panel-heading"><div><h2>字段准确率与人工修改率</h2><p>只统计已完成字段级复核采集的发票</p></div><el-tag v-if="dashboard.legacyWithoutFieldData" type="warning" effect="light">历史无字段数据 {{ dashboard.legacyWithoutFieldData }}</el-tag></div>
        <el-table :data="fieldMetrics" class="quality-table" size="small">
          <el-table-column prop="label" label="字段" min-width="150" />
          <el-table-column prop="reviewed" label="复核量" width="100" align="right" />
          <el-table-column prop="modified" label="人工修改" width="100" align="right" />
          <el-table-column label="修改率" width="120" align="right"><template #default="{ row }"><span class="numeric-value">{{ percent(row.modificationRate) }}</span></template></el-table-column>
          <el-table-column label="准确率" width="152" align="right"><template #default="{ row }"><div class="rate-cell"><strong :class="accuracyTone(row.accuracyRate)">{{ percent(row.accuracyRate) }}</strong><span class="rate-track"><i :style="{ width: `${safeRate(row.accuracyRate)}%` }" /></span></div></template></el-table-column>
          <el-table-column label="平均置信度" width="130" align="right"><template #default="{ row }"><span class="numeric-value">{{ confidencePercent(row.averageConfidence) }}</span></template></el-table-column>
          <template #empty><AppEmptyState compact title="尚无字段级复核指标" description="人工复核并保存字段修正后，这里会统计修改率、准确率和平均置信度。" /></template>
        </el-table>
      </section>

      <section class="na-panel quality-panel failure-panel">
        <div class="panel-heading"><div><h2>失败原因</h2><p>展示当前筛选范围内已结束的失败任务，便于定位 Provider 或模型问题</p></div><span class="record-count is-danger-count">{{ failuresTotal }} 条</span></div>
        <el-table :data="failures" class="quality-table" size="small">
          <el-table-column prop="fileName" label="文件" min-width="190" show-overflow-tooltip />
          <el-table-column prop="provider" label="Provider" min-width="120" />
          <el-table-column prop="model" label="模型" min-width="120" show-overflow-tooltip />
          <el-table-column label="尝试" width="90" align="right"><template #default="{ row }">{{ row.attempts }} / {{ row.maxAttempts }}</template></el-table-column>
          <el-table-column prop="error" label="失败原因" min-width="280" show-overflow-tooltip />
          <el-table-column label="时间" width="165"><template #default="{ row }">{{ dateText(row.updatedAt || row.createdAt) }}</template></el-table-column>
          <template #empty><AppEmptyState compact title="当前范围没有失败任务" description="识别任务运行正常，或当前筛选条件下没有已结束的失败记录。" :highlights="['可继续按日期、Provider 和模型筛选', '失败任务会保留错误原因与尝试次数']" /></template>
        </el-table>
        <el-pagination v-if="failuresTotal > pageSize" v-model:current-page="failurePage" :page-size="pageSize" :total="failuresTotal" layout="prev, pager, next" @current-change="loadFailures" />
      </section>
    </template>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Filter, Refresh } from '@element-plus/icons-vue'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import {
  getInvoiceQualityClassificationMetrics,
  getInvoiceQualityDashboard,
  getInvoiceQualityFailures,
  getInvoiceQualityFieldMetrics,
  getInvoiceQualityProviderMetrics
} from '@/plugin/invoice/api/invoice'

defineOptions({ name: 'InvoiceQuality' })

const loading = ref(false)
const loaded = ref(false)
const error = ref('')
const dateRange = ref([])
const filters = reactive({ provider: '', model: '' })
const dashboard = ref({})
const providerMetrics = ref([])
const fieldMetrics = ref([])
const classification = ref({ suggested: 0, accepted: 0, overridden: 0, pending: 0, acceptanceRate: 0 })
const failures = ref([])
const failuresTotal = ref(0)
const failurePage = ref(1)
const pageSize = 10

const query = computed(() => ({
  startDate: dateRange.value?.[0], endDate: dateRange.value?.[1],
  provider: filters.provider, model: filters.model
}))
const hasFilters = computed(() => Boolean(
  dateRange.value?.length || filters.provider || filters.model
))
const clearQualityFilters = () => {
  dateRange.value = []
  filters.provider = ''
  filters.model = ''
}
const percent = (value) => `${Number(value || 0).toFixed(2)}%`
const confidencePercent = (value) => percent(Number(value || 0) * 100)
const safeRate = (value) => Math.min(100, Math.max(0, Number(value || 0)))
const dateText = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'
const accuracyTone = (value) => Number(value || 0) >= 95 ? 'is-good' : Number(value || 0) >= 80 ? 'is-medium' : 'is-low'
const classificationShare = (value) => `${classification.value.suggested ? safeRate(Number(value || 0) / classification.value.suggested * 100) : 0}%`
const classificationRows = computed(() => [
  { label: '已采纳', value: classification.value.accepted || 0, share: classificationShare(classification.value.accepted), tone: 'tone-success' },
  { label: '已推翻', value: classification.value.overridden || 0, share: classificationShare(classification.value.overridden), tone: 'tone-warning' },
  { label: '待复核', value: classification.value.pending || 0, share: classificationShare(classification.value.pending), tone: 'tone-muted' }
])
const summaryCards = computed(() => [
  { label: '识别总量', value: dashboard.value.totalRecognitions || 0, note: `已复核 ${dashboard.value.reviewedInvoices || 0} 张`, tone: '' },
  { label: '识别成功率', value: percent(dashboard.value.successRate), note: `成功 ${dashboard.value.successfulRecognitions || 0} 张`, tone: 'is-primary' },
  { label: '平均耗时', value: `${dashboard.value.averageDurationMs || 0} ms`, note: `平均尝试 ${Number(dashboard.value.averageAttempts || 0).toFixed(2)} 次`, tone: '' },
  { label: '多模态回退率', value: percent(dashboard.value.multimodalFallbackRate), note: '成功任务中的回退比例', tone: 'is-warning' },
  { label: '字段修正', value: dashboard.value.correctedFields || 0, note: `历史无字段数据 ${dashboard.value.legacyWithoutFieldData || 0}`, tone: '' },
  { label: '失败率', value: percent(dashboard.value.failureRate), note: `失败 ${dashboard.value.failedRecognitions || 0} 张`, tone: 'is-danger' }
])

const readData = (response, fallback = []) => response?.code === 0 ? (response.data ?? fallback) : fallback
const loadFailures = async () => {
  const response = await getInvoiceQualityFailures({ ...query.value, page: failurePage.value, pageSize })
  if (response?.code === 0) {
    failures.value = response.data?.list || []
    failuresTotal.value = Number(response.data?.total || 0)
  }
}
const loadQuality = async () => {
  loading.value = true
  error.value = ''
  try {
    const [dashboardRes, providerRes, fieldRes, classificationRes] = await Promise.all([
      getInvoiceQualityDashboard(query.value), getInvoiceQualityProviderMetrics(query.value),
      getInvoiceQualityFieldMetrics(query.value), getInvoiceQualityClassificationMetrics(query.value)
    ])
    if ([dashboardRes, providerRes, fieldRes, classificationRes].some((response) => response?.code !== 0)) throw new Error('部分质量指标加载失败')
    dashboard.value = readData(dashboardRes, {})
    providerMetrics.value = readData(providerRes)
    fieldMetrics.value = readData(fieldRes)
    classification.value = readData(classificationRes, classification.value)
    failurePage.value = 1
    await loadFailures()
    loaded.value = true
  } catch (requestError) {
    error.value = requestError?.message || '质量数据加载失败'
    ElMessage.error(error.value)
  } finally {
    loading.value = false
  }
}

let filterTimer
watch([dateRange, () => filters.provider, () => filters.model], () => {
  clearTimeout(filterTimer)
  filterTimer = setTimeout(loadQuality, 260)
})
onMounted(loadQuality)
</script>

<style scoped lang="scss">
.quality-page { min-width: 0; }
.quality-page :deep(.na-page-header > :first-child), .quality-page :deep(.na-page-actions) { min-width: 0; }
.scope-bar { display: grid; min-width: 0; grid-template-columns: minmax(190px, 1fr) minmax(280px, 360px) 170px 170px auto; align-items: center; gap: 10px; margin-bottom: 14px; padding: 12px 14px; border: 1px solid var(--na-border); border-radius: var(--na-radius); background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.scope-heading { display: flex; min-width: 0; align-items: center; gap: 10px; }
.scope-heading > div { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.scope-heading strong { color: var(--na-foreground); font-size: .8125rem; font-weight: 650; }
.scope-heading small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.scope-icon { display: grid; width: 30px; height: 30px; flex: 0 0 auto; place-items: center; border-radius: 7px; background: var(--na-primary-soft); color: var(--na-primary); }
.filter-input, .date-filter { width: 100%; }
.clear-filter { margin-left: 0; }
.load-warning { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; padding: 9px 12px; border: 1px solid color-mix(in srgb, var(--na-warning) 28%, var(--na-border)); border-radius: 8px; background: var(--na-warning-soft); color: var(--na-foreground); font-size: .75rem; }
.metric-band { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); overflow: hidden; margin-bottom: 14px; border: 1px solid var(--na-border); border-radius: var(--na-radius); background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.metric-card { display: flex; min-width: 0; min-height: 112px; flex-direction: column; justify-content: center; gap: 6px; padding: 15px 18px; border-right: 1px solid var(--na-border); background: transparent; }
.metric-card:last-child { border-right: 0; }
.metric-label { display: flex; align-items: center; gap: 7px; color: var(--na-muted-foreground); font-size: .6875rem; }
.metric-label i { width: 6px; height: 6px; flex: 0 0 auto; border-radius: 50%; background: var(--na-muted-foreground); }
.metric-card strong { overflow: hidden; color: var(--na-foreground); font-size: 1.375rem; font-variant-numeric: tabular-nums; font-weight: 700; letter-spacing: -.02em; text-overflow: ellipsis; white-space: nowrap; }
.metric-card small { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.metric-card.is-primary .metric-label i { background: var(--na-primary); }.metric-card.is-primary strong { color: var(--na-primary); }
.metric-card.is-warning .metric-label i { background: var(--na-warning); }.metric-card.is-warning strong { color: var(--na-warning); }
.metric-card.is-danger .metric-label i { background: var(--na-danger); }.metric-card.is-danger strong { color: var(--na-danger); }
.quality-grid { display: grid; grid-template-columns: minmax(0, 2.2fr) minmax(290px, .8fr); gap: 14px; margin-bottom: 14px; }
.quality-panel, .classification-panel { min-width: 0; overflow: hidden; padding: 16px 18px 18px; }
.panel-heading { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.panel-heading > div { min-width: 0; }
.panel-heading h2 { margin: 0 0 4px; color: var(--na-foreground); font-size: .9375rem; font-weight: 650; }
.panel-heading p { overflow: hidden; margin: 0; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.record-count { flex: 0 0 auto; padding: 4px 8px; border-radius: 999px; background: var(--na-muted); color: var(--na-muted-foreground); font-size: .6875rem; font-variant-numeric: tabular-nums; }
.is-danger-count { background: var(--na-danger-soft); color: var(--na-danger); }
.quality-table { --el-table-header-bg-color: var(--na-table-header); --el-table-row-hover-bg-color: color-mix(in srgb, var(--na-primary) 4%, var(--na-card)); overflow: hidden; border: 1px solid var(--na-border); border-radius: 8px; }
.quality-table :deep(th.el-table__cell) { height: 42px; color: var(--na-muted-foreground); font-size: .6875rem; font-weight: 650; }
.quality-table :deep(td.el-table__cell) { height: 48px; color: var(--na-foreground); font-size: .75rem; }
.provider-name { display: flex; min-width: 0; align-items: center; gap: 8px; }
.provider-name i { width: 7px; height: 7px; flex: 0 0 auto; border-radius: 50%; background: var(--na-primary); box-shadow: 0 0 0 3px var(--na-primary-soft); }
.provider-name strong, .model-name { overflow: hidden; font-size: .75rem; font-weight: 620; text-overflow: ellipsis; white-space: nowrap; }
.file-type { display: inline-flex; max-width: 100%; padding: 3px 7px; overflow: hidden; border: 1px solid var(--na-border); border-radius: 5px; background: var(--na-muted); color: var(--na-muted-foreground); font-family: ui-monospace, SFMono-Regular, Consolas, monospace; font-size: .625rem; text-overflow: ellipsis; white-space: nowrap; }
.numeric-value, .correction-count { font-variant-numeric: tabular-nums; }
.correction-count { display: inline-grid; min-width: 24px; height: 24px; place-items: center; border-radius: 6px; color: var(--na-muted-foreground); }
.correction-count.has-correction { background: var(--na-warning-soft); color: var(--na-warning); font-weight: 650; }
.rate-cell { display: inline-grid; width: 100%; min-width: 90px; grid-template-columns: 52px minmax(34px, 1fr); align-items: center; gap: 8px; }
.rate-cell strong { font-size: .6875rem; font-variant-numeric: tabular-nums; font-weight: 680; text-align: right; }
.rate-track { height: 5px; overflow: hidden; border-radius: 3px; background: var(--na-muted); }
.rate-track i { display: block; height: 100%; border-radius: inherit; background: var(--na-primary); }
.is-good { color: var(--na-success) !important; }.is-medium { color: var(--na-warning) !important; }.is-low { color: var(--na-danger) !important; }
.classification-panel { display: flex; flex-direction: column; }
.classification-overview { display: grid; grid-template-columns: 1fr auto; align-items: end; gap: 18px; padding: 8px 0 14px; }
.classification-total { display: grid; min-width: 0; grid-template-columns: auto 1fr; align-items: baseline; gap: 2px 8px; }
.classification-total span, .classification-total small, .acceptance-rate span { color: var(--na-muted-foreground); font-size: .6875rem; }
.classification-total strong { grid-row: 2; color: var(--na-foreground); font-size: 2rem; font-variant-numeric: tabular-nums; font-weight: 720; letter-spacing: -.04em; }
.classification-total small { grid-row: 2; }
.acceptance-rate { display: flex; align-items: flex-end; flex-direction: column; gap: 4px; }
.acceptance-rate strong { color: var(--na-primary); font-size: 1.125rem; font-variant-numeric: tabular-nums; }
.classification-track { display: flex; height: 7px; overflow: hidden; border-radius: 4px; background: var(--na-muted); }
.classification-track i { display: block; height: 100%; }.classification-track .is-accepted { background: var(--na-success); }.classification-track .is-overridden { background: var(--na-warning); }.classification-track .is-pending { background: var(--na-muted-foreground); }
.classification-list { display: grid; gap: 0; margin: 12px 0 0; }
.classification-list div { display: flex; min-height: 44px; align-items: center; justify-content: space-between; gap: 12px; border-bottom: 1px solid var(--na-border); }
.classification-list div:last-child { border-bottom: 0; }
.classification-list dt { display: grid; min-width: 0; grid-template-columns: 7px auto auto; align-items: center; gap: 7px; margin: 0; color: var(--na-foreground); font-size: .75rem; }
.classification-list dt > i { width: 7px; height: 7px; border-radius: 50%; }.tone-success { background: var(--na-success); }.tone-warning { background: var(--na-warning); }.tone-muted { background: var(--na-muted-foreground); }
.classification-list dt small { color: var(--na-muted-foreground); font-size: .625rem; font-variant-numeric: tabular-nums; }
.classification-list dd { margin: 0; color: var(--na-foreground); font-size: .875rem; font-variant-numeric: tabular-nums; font-weight: 680; }
.field-panel, .failure-panel { margin-bottom: 14px; }
.field-panel :deep(.el-table__empty-block) { min-height: 116px; }
.field-panel :deep(.na-empty-state.is-compact) { min-height: 114px; padding-top: 14px; padding-bottom: 14px; }
.failure-panel :deep(.el-pagination) { justify-content: flex-end; margin-top: 14px; }
.quality-grid > .na-panel + .na-panel { margin-top: 0; }
@media (max-width: 1600px) {
  .scope-bar { grid-template-columns: minmax(180px, 1fr) minmax(260px, 1.4fr) minmax(140px, .7fr) minmax(140px, .7fr) auto; }
  .metric-band { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .metric-card:nth-child(3) { border-right: 0; }.metric-card:nth-child(-n + 3) { border-bottom: 1px solid var(--na-border); }
}
@media (max-width: 1180px) {
  .scope-bar { grid-template-columns: minmax(0, 1fr) minmax(260px, 1.4fr) minmax(140px, .8fr); }
  .scope-heading { grid-row: span 2; }.clear-filter { justify-self: end; }
  .quality-grid { grid-template-columns: 1fr; }
}
@media (max-width: 767px) {
  .quality-page :deep(.na-page-actions) { align-items: stretch; }.quality-page :deep(.na-page-actions > .el-button) { width: 100%; margin-left: 0; }
  .scope-bar { grid-template-columns: 1fr; padding: 12px; }.scope-heading { grid-row: auto; }.clear-filter { width: 100%; justify-self: stretch; }
  .metric-band { grid-template-columns: repeat(2, minmax(0, 1fr)); }.metric-card, .metric-card:nth-child(3) { border-right: 1px solid var(--na-border); border-bottom: 1px solid var(--na-border); }.metric-card:nth-child(2n) { border-right: 0; }.metric-card:nth-last-child(-n + 2) { border-bottom: 0; }
  .quality-panel, .classification-panel { padding: 14px 12px; }.panel-heading { align-items: flex-start; flex-direction: column; }.panel-heading p { white-space: normal; }
}
@media (max-width: 440px) {
  .metric-band { grid-template-columns: 1fr; }.metric-card, .metric-card:nth-child(2n), .metric-card:nth-last-child(-n + 2) { min-height: 96px; border-right: 0; border-bottom: 1px solid var(--na-border); }.metric-card:last-child { border-bottom: 0; }
  .classification-overview { gap: 10px; }
}
</style>
