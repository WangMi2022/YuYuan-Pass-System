<template>
  <main class="na-page na-page--list quality-page">
    <AppPageHeader title-id="invoice-quality-title" title="发票识别质量" description="按识别任务、字段修正和分类采纳情况查看质量闭环。">
      <template #actions>
        <el-date-picker v-model="dateRange" class="date-filter" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" />
        <el-input v-model="filters.provider" clearable placeholder="Provider" class="filter-input" />
        <el-input v-model="filters.model" clearable placeholder="模型版本" class="filter-input" />
        <el-button :icon="Refresh" :loading="loading" @click="loadQuality">刷新</el-button>
      </template>
    </AppPageHeader>

    <el-skeleton v-if="loading && !loaded" :rows="8" animated />
    <el-result v-else-if="error && !loaded" icon="error" title="质量数据加载失败" :sub-title="error">
      <template #extra><el-button type="primary" :icon="Refresh" @click="loadQuality">重新加载</el-button></template>
    </el-result>
    <template v-else>
      <div v-if="error" class="load-warning" role="alert">{{ error }}<el-button text :icon="Refresh" @click="loadQuality">重试</el-button></div>
      <section class="metric-grid" aria-label="识别质量指标">
        <article v-for="item in summaryCards" :key="item.label" class="metric-card" :class="item.tone">
          <span>{{ item.label }}</span><strong>{{ item.value }}</strong><small>{{ item.note }}</small>
        </article>
      </section>

      <div class="quality-grid">
        <section class="na-panel quality-panel">
          <div class="panel-heading"><div><h2>Provider 与模型</h2><p>按文件类型拆分识别成功率和修正量</p></div></div>
          <el-table :data="providerMetrics" stripe size="small" empty-text="暂无识别数据">
            <el-table-column prop="provider" label="Provider" min-width="145" />
            <el-table-column prop="model" label="模型" min-width="130" show-overflow-tooltip />
            <el-table-column prop="fileType" label="文件类型" width="112" />
            <el-table-column prop="total" label="总量" width="72" align="right" />
            <el-table-column label="成功率" width="98" align="right"><template #default="{ row }">{{ percent(row.successRate) }}</template></el-table-column>
            <el-table-column label="平均置信度" width="112" align="right"><template #default="{ row }">{{ confidencePercent(row.averageConfidence) }}</template></el-table-column>
            <el-table-column prop="correctedFields" label="修正字段" width="92" align="right" />
          </el-table>
        </section>

        <section class="na-panel classification-panel">
          <div class="panel-heading"><div><h2>分类推荐效果</h2><p>只统计已有建议分类的任务</p></div></div>
          <div class="classification-total"><strong>{{ classification.suggested }}</strong><span>次分类建议</span></div>
          <dl class="classification-list">
            <div><dt>采纳</dt><dd class="is-success">{{ classification.accepted }}</dd></div>
            <div><dt>推翻</dt><dd class="is-warning">{{ classification.overridden }}</dd></div>
            <div><dt>待复核</dt><dd>{{ classification.pending }}</dd></div>
            <div><dt>采纳率</dt><dd>{{ percent(classification.acceptanceRate) }}</dd></div>
          </dl>
        </section>
      </div>

      <section class="na-panel quality-panel field-panel">
        <div class="panel-heading"><div><h2>字段准确率与人工修改率</h2><p>只统计已完成字段级复核采集的发票</p></div><el-tag v-if="dashboard.legacyWithoutFieldData" type="warning">历史无字段数据 {{ dashboard.legacyWithoutFieldData }}</el-tag></div>
        <el-table :data="fieldMetrics" stripe size="small" empty-text="完成复核后生成字段指标">
          <el-table-column prop="label" label="字段" min-width="150" />
          <el-table-column prop="reviewed" label="复核量" width="100" align="right" />
          <el-table-column prop="modified" label="人工修改" width="100" align="right" />
          <el-table-column label="修改率" width="120" align="right"><template #default="{ row }">{{ percent(row.modificationRate) }}</template></el-table-column>
          <el-table-column label="准确率" width="120" align="right"><template #default="{ row }"><strong :class="accuracyTone(row.accuracyRate)">{{ percent(row.accuracyRate) }}</strong></template></el-table-column>
          <el-table-column label="平均置信度" width="130" align="right"><template #default="{ row }">{{ confidencePercent(row.averageConfidence) }}</template></el-table-column>
        </el-table>
      </section>

      <section class="na-panel quality-panel failure-panel">
        <div class="panel-heading"><div><h2>失败原因</h2><p>展示当前筛选范围内已结束的失败任务</p></div><span class="muted">{{ failuresTotal }} 条</span></div>
        <el-table :data="failures" stripe size="small" empty-text="暂无失败任务">
          <el-table-column prop="fileName" label="文件" min-width="190" show-overflow-tooltip />
          <el-table-column prop="provider" label="Provider" min-width="120" />
          <el-table-column prop="model" label="模型" min-width="120" show-overflow-tooltip />
          <el-table-column label="尝试" width="90" align="right"><template #default="{ row }">{{ row.attempts }} / {{ row.maxAttempts }}</template></el-table-column>
          <el-table-column prop="error" label="失败原因" min-width="280" show-overflow-tooltip />
          <el-table-column label="时间" width="165"><template #default="{ row }">{{ dateText(row.updatedAt || row.createdAt) }}</template></el-table-column>
        </el-table>
        <el-pagination v-if="failuresTotal > pageSize" v-model:current-page="failurePage" :page-size="pageSize" :total="failuresTotal" layout="prev, pager, next" @current-change="loadFailures" />
      </section>
    </template>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
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
const percent = (value) => `${Number(value || 0).toFixed(2)}%`
const confidencePercent = (value) => percent(Number(value || 0) * 100)
const dateText = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'
const accuracyTone = (value) => Number(value || 0) >= 95 ? 'is-good' : Number(value || 0) >= 80 ? 'is-medium' : 'is-low'
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
.filter-input { width: 150px; }
.quality-page :deep(.na-page-header > :first-child), .quality-page :deep(.na-page-actions) { min-width: 0; }
.quality-page :deep(.na-page-actions) { flex-wrap: wrap; justify-content: flex-end; max-width: 100%; }
.date-filter { width: 430px; max-width: 100%; }
.load-warning { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; padding: 9px 12px; border-radius: 8px; background: var(--na-warning-soft); color: var(--na-foreground); font-size: 12px; }
.metric-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 12px; margin-bottom: 14px; }
.metric-card { display: flex; min-width: 0; flex-direction: column; gap: 7px; padding: 16px; border: 1px solid var(--na-border); border-radius: 10px; background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.metric-card span, .metric-card small, .panel-heading p, .muted { color: var(--na-muted-foreground); font-size: 12px; }
.metric-card strong { overflow: hidden; color: var(--na-foreground); font-size: 22px; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.metric-card.is-primary strong { color: var(--na-primary); }
.metric-card.is-warning strong { color: var(--na-warning); }
.metric-card.is-danger strong { color: var(--na-danger); }
.quality-grid { display: grid; grid-template-columns: minmax(0, 1.8fr) minmax(270px, .8fr); gap: 14px; margin-bottom: 14px; }
.quality-panel { min-width: 0; overflow: hidden; padding: 16px; }
.panel-heading { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.panel-heading h2 { margin: 0 0 4px; color: var(--na-foreground); font-size: 16px; }
.panel-heading p { margin: 0; }
.classification-total { display: flex; align-items: baseline; gap: 8px; padding: 12px 0 18px; }
.classification-total strong { color: var(--na-primary); font-size: 34px; font-variant-numeric: tabular-nums; }
.classification-total span { color: var(--na-muted-foreground); font-size: 12px; }
.classification-list { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin: 0; }
.classification-list div { display: flex; align-items: center; justify-content: space-between; padding: 11px 12px; border: 1px solid var(--na-border); border-radius: 7px; background: var(--na-muted); }
.classification-list dt { color: var(--na-muted-foreground); font-size: 12px; }
.classification-list dd { margin: 0; color: var(--na-foreground); font-size: 18px; font-weight: 700; }
.is-success, .is-good { color: var(--na-success) !important; }
.is-warning { color: var(--na-warning) !important; }
.is-medium { color: var(--na-warning); }.is-low { color: var(--na-danger); }
.field-panel, .failure-panel { margin-bottom: 14px; }
.failure-panel :deep(.el-pagination) { justify-content: flex-end; margin-top: 14px; }
.quality-grid > .na-panel + .na-panel { margin-top: 0; }
@media (max-width: 1600px) {
  .quality-page :deep(.na-page-header) { align-items: stretch; flex-direction: column; gap: 14px; }
  .quality-page :deep(.na-page-actions) { justify-content: flex-start; }
  .metric-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .quality-grid { grid-template-columns: 1fr; }
}
@media (max-width: 767px) {
  .quality-page :deep(.na-page-actions) { align-items: stretch; }
  .date-filter, .filter-input { flex: 1 1 100%; width: 100%; }
  .quality-page :deep(.na-page-actions > .el-button) { width: 100%; margin-left: 0; }
  .metric-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .quality-panel { padding: 12px; }
  .panel-heading { align-items: flex-start; flex-direction: column; }
}
@media (max-width: 420px) { .metric-grid { grid-template-columns: 1fr; } }
</style>
