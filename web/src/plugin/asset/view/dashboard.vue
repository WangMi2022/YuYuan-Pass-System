<template>
  <main
    ref="screenRef"
    v-loading="loading"
    class="asset-dashboard"
    :class="'is-' + viewMode"
  >
    <div class="dashboard-content">
      <AppPageHeader
        title-id="asset-dashboard-title"
        title="资产可视化大屏"
        description="在全景指挥舱与模块矩阵之间切换，实时查看资产价值、状态、分类与空间分布。"
      >
        <template #actions>
          <div class="view-mode-switch" role="group" aria-label="大屏布局">
            <button
              v-for="mode in viewModes"
              :key="mode.value"
              type="button"
              :class="{ active: viewMode === mode.value }"
              :aria-pressed="viewMode === mode.value"
              @click="selectViewMode(mode.value)"
            >
              <el-icon><component :is="mode.icon" /></el-icon>
              <span>{{ mode.label }}</span>
            </button>
          </div>
          <span class="update-time"><i /> 数据更新于 {{ updateTime }}</span>
          <el-button :icon="Refresh" :loading="loading" @click="loadDashboard">刷新</el-button>
          <el-button type="primary" :icon="FullScreen" @click="toggleFullScreen">全屏展示</el-button>
        </template>
      </AppPageHeader>

      <Transition name="dashboard-mode" mode="out-in">
        <section
          v-if="viewMode === 'panorama'"
          key="panorama"
          class="panorama-board"
          aria-label="全景指挥舱"
        >
          <div class="panorama-main">
            <article class="dashboard-panel value-wing">
              <PanelHeading title="资产价值结构" description="分类价值与资产投入" />
              <div class="value-metrics">
                <div>
                  <span>资产原值</span>
                  <strong>{{ formatCompactCurrency(dashboard.originalValue) }}</strong>
                  <small>全生命周期累计投入</small>
                </div>
                <div>
                  <span>当前估值</span>
                  <strong>{{ formatCompactCurrency(dashboard.currentValue) }}</strong>
                  <small>价值保有率 {{ retentionRate }}</small>
                </div>
              </div>
              <div v-if="hasCategoryData" class="category-chart">
                <Chart :options="categoryBarOption" height="228px" />
              </div>
              <el-empty v-else description="登记资产后显示分类价值" :image-size="72" />
            </article>

            <article class="dashboard-panel orbit-panel">
              <PanelHeading title="资产运营态势" description="基于当前台账实时汇总">
                <span class="live-badge"><i /> 实时数据</span>
              </PanelHeading>
              <div class="orbit-canvas" aria-label="资产核心指标">
                <div class="orbit-layer orbit-large"><i /><i /><i /></div>
                <div class="orbit-layer orbit-small"><i /><i /></div>
                <div class="orbit-core">
                  <span>在册资产</span>
                  <strong>{{ formatNumber(dashboard.totalQuantity) }}</strong>
                  <small>件资产实物</small>
                </div>
                <div class="orbit-stat stat-a"><strong>{{ dashboard.assetKinds }}</strong><span>资产档案</span></div>
                <div class="orbit-stat stat-b"><strong>{{ dashboard.categoryCount }}</strong><span>资产分类</span></div>
                <div class="orbit-stat stat-c"><strong>{{ dashboard.locationSummary.length }}</strong><span>覆盖空间</span></div>
                <div class="orbit-stat stat-d"><strong>{{ retentionRate }}</strong><span>价值保有率</span></div>
              </div>
            </article>

            <article class="dashboard-panel status-wing">
              <PanelHeading title="资产状态构成" description="按实物数量统计" />
              <div v-if="hasStatusData" class="status-chart">
                <Chart :options="statusOption" height="248px" />
              </div>
              <el-empty v-else description="暂无状态数据" :image-size="72" />
              <div class="status-list">
                <div v-for="item in statusRows" :key="item.status">
                  <span><i :style="{ background: item.color }" />{{ item.label }}</span>
                  <strong>{{ item.quantity }}</strong>
                  <small>{{ formatPercent(item.quantity, dashboard.totalQuantity) }}</small>
                </div>
              </div>
            </article>
          </div>

          <div class="panorama-footer">
            <article class="dashboard-panel footer-panel">
              <PanelHeading title="空间热度" description="资产数量排名" />
              <div v-if="dashboard.locationSummary.length" class="location-list">
                <div
                  v-for="(item, index) in dashboard.locationSummary.slice(0, 6)"
                  :key="item.location"
                  class="location-row"
                >
                  <span class="rank" :class="{ top: index < 3 }">{{ String(index + 1).padStart(2, '0') }}</span>
                  <div>
                    <strong>{{ item.location }}</strong>
                    <i><em :style="{ width: locationPercent(item.quantity) + '%' }" /></i>
                  </div>
                  <b>{{ item.quantity }} 件</b>
                </div>
              </div>
              <el-empty v-else description="补充位置后显示排行" :image-size="64" />
            </article>

            <article class="dashboard-panel footer-panel value-structure">
              <PanelHeading title="价值保有结构" description="原值、当前估值与累计价值减少" />
              <div class="retention-summary">
                <div class="retention-ring" :style="{ '--retention': retentionRateNumber * 3.6 + 'deg' }">
                  <span><strong>{{ retentionRate }}</strong><small>保有率</small></span>
                </div>
                <div class="value-segments">
                  <div v-for="item in valueSegments" :key="item.label">
                    <span>{{ item.label }}</span>
                    <strong>{{ formatCompactCurrency(item.value) }}</strong>
                    <i><em :style="{ width: item.percent + '%', background: item.color }" /></i>
                  </div>
                </div>
              </div>
            </article>

            <article class="dashboard-panel footer-panel">
              <PanelHeading title="最近登记" description="最新录入的资产档案" />
              <div v-if="recentAssets.length" class="recent-list">
                <div v-for="item in recentAssets.slice(0, 5)" :key="item.ID || item.id || item.assetCode" class="recent-row">
                  <span class="asset-mark" />
                  <div><strong>{{ item.name }}</strong><small>{{ item.assetCode || '暂无资产编码' }}</small></div>
                  <span>{{ item.category?.name || '未分类' }}</span>
                  <b>{{ item.quantity }} {{ item.unit }}</b>
                </div>
              </div>
              <el-empty v-else description="暂无资产" :image-size="64" />
            </article>
          </div>
        </section>

        <section
          v-else
          key="matrix"
          class="matrix-board"
          aria-label="业务模块矩阵"
        >
          <article class="matrix-card matrix-total">
            <span>资产实物总量</span>
            <strong>{{ formatNumber(dashboard.totalQuantity) }}</strong>
            <small>{{ formatNumber(dashboard.assetKinds) }} 种资产档案</small>
            <el-icon><Box /></el-icon>
          </article>

          <article class="matrix-card matrix-value">
            <div>
              <span>当前估值</span>
              <strong>{{ formatCompactCurrency(dashboard.currentValue) }}</strong>
              <small>原值 {{ formatCompactCurrency(dashboard.originalValue) }}</small>
            </div>
            <div class="matrix-retention-bar">
              <span><b>价值保有率</b><strong>{{ retentionRate }}</strong></span>
              <i><em :style="{ width: retentionRateNumber + '%' }" /></i>
            </div>
          </article>

          <article class="matrix-card matrix-status">
            <PanelHeading title="状态构成" description="资产实物分布" />
            <div v-if="hasStatusData" class="matrix-status-chart">
              <Chart :options="statusOption" height="250px" />
            </div>
            <el-empty v-else description="暂无状态数据" :image-size="72" />
            <div class="matrix-status-list">
              <span v-for="item in statusRows" :key="item.status">
                <i :style="{ background: item.color }" />
                <b>{{ item.label }}</b>
                <strong>{{ item.quantity }}</strong>
              </span>
            </div>
          </article>

          <article class="matrix-card matrix-category">
            <PanelHeading title="分类矩阵" description="按资产数量与当前估值展示" />
            <div v-if="activeCategories.length" class="category-matrix">
              <div
                v-for="(item, index) in activeCategories.slice(0, 8)"
                :key="item.categoryId"
                class="category-tile"
                :class="{ featured: index === 0 }"
                :style="{ '--category-color': item.color }"
              >
                <span>{{ item.categoryName }}</span>
                <strong>{{ item.quantity }}<small> 件</small></strong>
                <b>{{ formatCompactCurrency(item.currentValue) }}</b>
              </div>
            </div>
            <el-empty v-else description="暂无分类数据" :image-size="72" />
          </article>

          <article class="matrix-card matrix-locations">
            <PanelHeading title="空间排行" description="TOP 6" />
            <div v-if="dashboard.locationSummary.length" class="matrix-location-list">
              <div v-for="(item, index) in dashboard.locationSummary.slice(0, 6)" :key="item.location">
                <span>{{ String(index + 1).padStart(2, '0') }}</span>
                <div><b>{{ item.location }}</b><i><em :style="{ width: locationPercent(item.quantity) + '%' }" /></i></div>
                <strong>{{ item.quantity }}</strong>
              </div>
            </div>
            <el-empty v-else description="暂无位置数据" :image-size="64" />
          </article>

          <article class="matrix-card matrix-recent">
            <PanelHeading title="最近登记" description="最新资产档案" />
            <div v-if="recentAssets.length" class="matrix-recent-list">
              <div v-for="item in recentAssets.slice(0, 6)" :key="item.ID || item.id || item.assetCode">
                <span>{{ item.assetCode || '—' }}</span>
                <b>{{ item.name }}</b>
                <small>{{ item.category?.name || '未分类' }}</small>
                <strong>{{ formatCompactCurrency(item.currentValue) }}</strong>
              </div>
            </div>
            <el-empty v-else description="暂无资产" :image-size="64" />
          </article>

          <article class="matrix-card matrix-retention">
            <span>价值保有率</span>
            <strong>{{ retentionRate }}</strong>
            <small>当前估值 / 资产原值</small>
            <div class="mini-ring" :style="{ '--retention': retentionRateNumber * 3.6 + 'deg' }" />
          </article>

          <article class="matrix-card matrix-depreciation">
            <span>累计价值减少</span>
            <strong>{{ formatCompactCurrency(dashboard.depreciation) }}</strong>
            <small>{{ depreciationRate }} · 相对资产原值</small>
            <el-icon><DataAnalysis /></el-icon>
          </article>
        </section>
      </Transition>
    </div>
  </main>
</template>

<script setup>
import { computed, defineComponent, h, markRaw, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Box, DataAnalysis, DataBoard, FullScreen, Grid, Refresh } from '@element-plus/icons-vue'
import Chart from '@/components/charts/index.vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { chartPalette } from '@/components/charts/theme'
import { formatCompactCurrency, formatNumber, formatPercent } from '@/utils/format'
import { getAssetDashboard } from '@/plugin/asset/api/asset'
import { useAppStore } from '@/pinia'

defineOptions({ name: 'AssetDashboard' })

const PanelHeading = defineComponent({
  name: 'AssetDashboardPanelHeading',
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' }
  },
  setup(props, { slots }) {
    return () => h('header', { class: 'panel-heading' }, [
      h('div', [
        h('h2', props.title),
        props.description ? h('p', props.description) : null
      ]),
      slots.default?.()
    ])
  }
})

const VIEW_STORAGE_KEY = 'asset-dashboard-layout'
const validViewModes = ['panorama', 'matrix']
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const screenRef = ref()
const loading = ref(false)
const updateTime = ref('—')

const viewModes = [
  { value: 'panorama', label: '全景指挥舱', icon: markRaw(DataBoard) },
  { value: 'matrix', label: '模块矩阵', icon: markRaw(Grid) }
]

const isValidViewMode = (value) => typeof value === 'string' && validViewModes.includes(value)
const savedViewMode = (() => {
  try {
    return localStorage.getItem(VIEW_STORAGE_KEY)
  } catch {
    return null
  }
})()
const viewMode = ref(isValidViewMode(route.query.layout)
  ? route.query.layout
  : isValidViewMode(savedViewMode) ? savedViewMode : 'panorama')

const dashboard = ref({
  assetKinds: 0,
  totalQuantity: 0,
  categoryCount: 0,
  originalValue: 0,
  currentValue: 0,
  depreciation: 0,
  categorySummary: [],
  statusSummary: [],
  locationSummary: [],
  recentAssets: []
})

const selectViewMode = (mode) => {
  if (!isValidViewMode(mode)) return
  viewMode.value = mode
  try {
    localStorage.setItem(VIEW_STORAGE_KEY, mode)
  } catch {
    // The URL still preserves the selected layout when storage is unavailable.
  }
  if (route.query.layout !== mode) {
    router.replace({ query: { ...route.query, layout: mode } })
  }
}

watch(
  () => route.query.layout,
  (mode) => {
    if (isValidViewMode(mode) && mode !== viewMode.value) {
      viewMode.value = mode
      try {
        localStorage.setItem(VIEW_STORAGE_KEY, mode)
      } catch {
        // Ignore storage restrictions; route state remains authoritative.
      }
    }
  }
)

const retentionRateNumber = computed(() => {
  const original = Number(dashboard.value.originalValue || 0)
  if (!original) return 0
  return Math.min(100, Math.max(0, Number(dashboard.value.currentValue || 0) / original * 100))
})
const retentionRate = computed(() => retentionRateNumber.value.toFixed(1) + '%')
const depreciationRate = computed(() => {
  const original = Number(dashboard.value.originalValue || 0)
  if (!original) return '0.0%'
  return (Number(dashboard.value.depreciation || 0) / original * 100).toFixed(1) + '%'
})

const chartColors = computed(() => {
  void appStore.isDark
  void appStore.config.primaryColor
  const palette = chartPalette()
  return {
    primary: palette[0] || '#6d5dfb',
    palette,
    text: '#19172c',
    muted: '#706b82',
    grid: '#e2e0ec',
    surface: '#ffffff'
  }
})

const activeCategories = computed(() => {
  const palette = chartColors.value.palette
  return (dashboard.value.categorySummary || [])
    .filter((item) => Number(item.quantity) > 0)
    .map((item, index) => ({
      ...item,
      color: item.color || palette[index % palette.length]
    }))
})
const hasCategoryData = computed(() => activeCategories.value.length > 0)
const recentAssets = computed(() => dashboard.value.recentAssets || [])

const statusConfig = {
  pending_inbound: { label: '待入库', color: '#0284c7' },
  in_use: { label: '使用中', color: '#059669' },
  idle: { label: '闲置', color: '#8a8697' },
  maintenance: { label: '维修中', color: '#d97706' },
  retired: { label: '已处置', color: '#dc2626' }
}
const statusOrder = ['pending_inbound', 'idle', 'in_use', 'maintenance', 'retired']
const statusRows = computed(() => (dashboard.value.statusSummary || [])
  .map((item) => ({
    ...item,
    ...(statusConfig[item.status] || { label: item.status, color: '#8a8697' })
  }))
  .sort((a, b) => statusOrder.indexOf(a.status) - statusOrder.indexOf(b.status)))
const hasStatusData = computed(() => statusRows.value.some((item) => Number(item.quantity) > 0))

const categoryBarOption = computed(() => {
  const theme = chartColors.value
  return {
    animationDuration: 420,
    grid: { top: 4, right: 40, bottom: 4, left: 10, containLabel: true },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      valueFormatter: (value) => Number(value).toFixed(1) + ' 万元'
    },
    xAxis: {
      type: 'value',
      axisLabel: { show: false },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { show: false }
    },
    yAxis: {
      type: 'category',
      inverse: true,
      data: activeCategories.value.slice(0, 7).map((item) => item.categoryName),
      axisLabel: { color: theme.muted, fontSize: 11 },
      axisLine: { show: false },
      axisTick: { show: false }
    },
    series: [{
      type: 'bar',
      barWidth: 7,
      showBackground: true,
      backgroundStyle: { color: '#efedf4', borderRadius: 4 },
      label: {
        show: true,
        position: 'right',
        color: theme.text,
        fontSize: 10,
        formatter: ({ value }) => Number(value).toFixed(1)
      },
      data: activeCategories.value.slice(0, 7).map((item) => ({
        value: Number(item.currentValue || 0) / 10000,
        itemStyle: { color: item.color, borderRadius: 4 }
      }))
    }]
  }
})

const statusOption = computed(() => {
  const theme = chartColors.value
  return {
    animationDuration: 420,
    tooltip: {
      trigger: 'item',
      formatter: ({ name, value, percent }) => name + '<br/>' + value + ' 件（' + percent + '%）'
    },
    series: [{
      type: 'pie',
      radius: ['62%', '80%'],
      center: ['50%', '48%'],
      avoidLabelOverlap: true,
      itemStyle: { borderColor: theme.surface, borderWidth: 4, borderRadius: 4 },
      label: { show: false },
      data: statusRows.value.map((item) => ({
        name: item.label,
        value: Number(item.quantity || 0),
        itemStyle: { color: item.color }
      }))
    }],
    graphic: [
      {
        type: 'text',
        left: 'center',
        top: '37%',
        style: {
          text: formatNumber(dashboard.value.totalQuantity),
          fill: theme.text,
          fontSize: 30,
          fontWeight: 700
        }
      },
      {
        type: 'text',
        left: 'center',
        top: '54%',
        style: { text: '资产总量', fill: theme.muted, fontSize: 11 }
      }
    ]
  }
})

const valueSegments = computed(() => {
  const original = Math.max(Number(dashboard.value.originalValue || 0), 1)
  return [
    {
      label: '资产原值',
      value: dashboard.value.originalValue,
      percent: 100,
      color: '#6d5dfb'
    },
    {
      label: '当前估值',
      value: dashboard.value.currentValue,
      percent: Math.max(3, Number(dashboard.value.currentValue || 0) / original * 100),
      color: '#10b981'
    },
    {
      label: '累计价值减少',
      value: dashboard.value.depreciation,
      percent: Math.max(3, Number(dashboard.value.depreciation || 0) / original * 100),
      color: '#f59e0b'
    }
  ]
})

const locationPercent = (quantity) => {
  const max = Math.max(
    ...(dashboard.value.locationSummary || []).map((item) => Number(item.quantity || 0)),
    1
  )
  return Math.max(5, Math.round(Number(quantity || 0) / max * 100))
}

const loadDashboard = async () => {
  loading.value = true
  try {
    const res = await getAssetDashboard()
    if (res.code === 0) {
      const data = res.data || {}
      dashboard.value = {
        ...dashboard.value,
        ...data,
        categorySummary: data.categorySummary || [],
        statusSummary: data.statusSummary || [],
        locationSummary: data.locationSummary || [],
        recentAssets: data.recentAssets || []
      }
      updateTime.value = new Date().toLocaleString('zh-CN', { hour12: false })
    }
  } finally {
    loading.value = false
  }
}

const toggleFullScreen = async () => {
  if (!document.fullscreenElement) await screenRef.value?.requestFullscreen()
  else await document.exitFullscreen()
}

onMounted(loadDashboard)
</script>

<style scoped lang="scss">
.asset-dashboard {
  --canvas: #f6f5fa;
  --surface: #ffffff;
  --surface-soft: #f3f2f7;
  --text: #19172c;
  --muted: #706b82;
  --border: #e2e0ec;
  --border-strong: #d5d1e2;
  --primary: var(--na-primary, #6d5dfb);
  --primary-soft: color-mix(in srgb, var(--primary) 10%, #fff);
  position: relative;
  min-height: 100%;
  overflow: auto;
  background: var(--canvas);
  color: var(--text);
}

.asset-dashboard:fullscreen {
  min-height: 100vh;
}

.dashboard-content {
  min-width: 1160px;
  padding: 24px;
}

.asset-dashboard :deep(.na-page-title) { color: var(--text); }
.asset-dashboard :deep(.na-page-description) { color: var(--muted); }

.update-time {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  margin-right: 2px;
  color: var(--muted);
  font-size: 12px;
}

.update-time i,
.live-badge i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #059669;
  box-shadow: 0 0 0 4px rgb(5 150 105 / 9%);
}

.view-mode-switch {
  display: inline-flex;
  gap: 3px;
  padding: 3px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface-soft);
}

.view-mode-switch button {
  display: inline-flex;
  min-height: 30px;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border: 0;
  border-radius: 7px;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  font: 600 12px/1 var(--na-font-sans);
  transition: background-color 180ms cubic-bezier(.22, 1, .36, 1), color 180ms ease;
}

.view-mode-switch button:hover {
  color: var(--text);
}

.view-mode-switch button.active {
  background: #fff;
  color: var(--primary);
  box-shadow: 0 2px 4px rgb(53 45 111 / 8%);
}

.view-mode-switch button:focus-visible {
  outline: 3px solid color-mix(in srgb, var(--primary) 20%, transparent);
  outline-offset: 1px;
}

.dashboard-panel,
.matrix-card {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 14px;
  background: var(--surface);
  box-shadow: 0 2px 4px rgb(53 45 111 / 4%);
}

.panel-heading {
  display: flex;
  min-height: 58px;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px 0;
}

.panel-heading h2 {
  margin: 0;
  color: var(--text);
  font-size: 14px;
  font-weight: 650;
}

.panel-heading p {
  margin: 3px 0 0;
  color: var(--muted);
  font-size: 11px;
}

.live-badge {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #047857;
  font-size: 11px;
}

/* Panorama layout */
.panorama-board {
  display: grid;
  gap: 12px;
}

.panorama-main {
  display: grid;
  min-height: 510px;
  grid-template-columns: minmax(270px, 23%) minmax(520px, 1fr) minmax(280px, 24%);
  gap: 12px;
}

.value-metrics {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 8px 16px 0;
}

.value-metrics > div {
  min-width: 0;
  padding: 13px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface-soft);
}

.value-metrics span,
.value-metrics small {
  display: block;
  color: var(--muted);
  font-size: 10px;
}

.value-metrics strong {
  display: block;
  overflow: hidden;
  margin: 5px 0 3px;
  color: var(--text);
  font-size: 20px;
  font-variant-numeric: tabular-nums;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.category-chart {
  padding: 12px 8px 4px;
}

.orbit-panel {
  position: relative;
  background: radial-gradient(circle at 50% 56%, #eeebff 0, #f9f8fd 42%, #fff 72%);
}

.orbit-heading {
  position: relative;
  z-index: 4;
}

.orbit-canvas {
  position: absolute;
  inset: 58px 0 0;
  overflow: hidden;
}

.orbit-layer,
.orbit-core {
  position: absolute;
  top: 51%;
  left: 50%;
  border-radius: 50%;
  transform: translate(-50%, -50%);
}

.orbit-layer {
  border: 1px solid color-mix(in srgb, var(--primary) 27%, var(--border));
}

.orbit-layer::before,
.orbit-layer::after {
  position: absolute;
  border: 1px dashed color-mix(in srgb, var(--primary) 15%, var(--border));
  border-radius: 50%;
  content: '';
}

.orbit-layer::before { inset: 13%; }
.orbit-layer::after { inset: 29%; }
.orbit-large { width: 390px; height: 390px; animation: orbit-spin 26s linear infinite; }
.orbit-small { width: 280px; height: 280px; animation: orbit-spin-reverse 19s linear infinite; }

.orbit-layer > i {
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--primary);
  box-shadow: 0 0 0 5px color-mix(in srgb, var(--primary) 10%, transparent);
}

.orbit-layer > i:nth-child(1) { top: -4px; left: 50%; }
.orbit-layer > i:nth-child(2) { right: 9%; bottom: 18%; }
.orbit-layer > i:nth-child(3) { bottom: 8%; left: 15%; }

.orbit-core {
  z-index: 2;
  display: grid;
  width: 168px;
  height: 168px;
  place-content: center;
  border: 1px solid color-mix(in srgb, var(--primary) 25%, var(--border));
  background: #fff;
  text-align: center;
  box-shadow: 0 4px 8px rgb(53 45 111 / 8%), inset 0 0 0 11px #f7f5ff;
}

.orbit-core span,
.orbit-core small {
  color: var(--muted);
  font-size: 10px;
}

.orbit-core strong {
  margin: 3px 0;
  color: var(--text);
  font-size: 48px;
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.orbit-stat {
  position: absolute;
  z-index: 3;
  display: flex;
  min-width: 112px;
  flex-direction: column;
  gap: 3px;
  padding: 10px 12px;
  border: 1px solid color-mix(in srgb, var(--primary) 22%, var(--border));
  border-radius: 9px;
  background: #fff;
  box-shadow: 0 3px 7px rgb(53 45 111 / 7%);
}

.orbit-stat strong {
  color: var(--primary);
  font-size: 18px;
  font-variant-numeric: tabular-nums;
}

.orbit-stat span { color: var(--muted); font-size: 10px; }
.stat-a { top: 14%; left: 9%; }
.stat-b { top: 17%; right: 8%; }
.stat-c { bottom: 13%; left: 8%; }
.stat-d { right: 7%; bottom: 16%; }

.status-chart {
  padding: 0 8px;
}

.status-list {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 10px;
  padding: 0 16px 14px;
}

.status-list > div {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 3px 7px;
  padding: 8px 2px;
  border-bottom: 1px solid var(--border);
}

.status-list span {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--muted);
  font-size: 11px;
}

.status-list span i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.status-list strong { font-size: 12px; font-variant-numeric: tabular-nums; }
.status-list small { grid-column: 2; color: var(--muted); font-size: 10px; text-align: right; }

.panorama-footer {
  display: grid;
  min-height: 230px;
  grid-template-columns: .92fr 1.08fr 1fr;
  gap: 12px;
}

.footer-panel {
  min-height: 230px;
}

.location-list,
.recent-list {
  padding: 4px 16px 12px;
}

.location-row {
  display: grid;
  min-height: 29px;
  grid-template-columns: 25px minmax(0, 1fr) 50px;
  align-items: center;
  gap: 8px;
}

.location-row .rank {
  color: var(--muted);
  font: 600 10px/1 ui-monospace, SFMono-Regular, Menlo, monospace;
}

.location-row .rank.top { color: var(--primary); }
.location-row > div { display: grid; grid-template-columns: minmax(75px, 1fr) 1.2fr; align-items: center; gap: 8px; }
.location-row strong { overflow: hidden; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.location-row i,
.value-segments i,
.matrix-location-list i,
.matrix-retention-bar > i {
  display: block;
  height: 5px;
  overflow: hidden;
  border-radius: 999px;
  background: #eceaf1;
}
.location-row i em,
.value-segments i em,
.matrix-location-list i em,
.matrix-retention-bar > i em {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: var(--primary);
}
.location-row > b { font-size: 11px; text-align: right; }

.retention-summary {
  display: grid;
  grid-template-columns: 128px 1fr;
  align-items: center;
  gap: 20px;
  padding: 6px 20px 16px;
}

.retention-ring,
.mini-ring {
  display: grid;
  border-radius: 50%;
  background: conic-gradient(var(--primary) var(--retention), #ebe9f1 0);
}

.retention-ring {
  width: 116px;
  height: 116px;
  place-items: center;
}

.retention-ring::before,
.mini-ring::before {
  border-radius: 50%;
  background: var(--surface);
  content: '';
}

.retention-ring::before { width: 88px; height: 88px; }

.retention-ring > span {
  position: absolute;
  display: flex;
  flex-direction: column;
  text-align: center;
}

.retention-ring strong { font-size: 22px; }
.retention-ring small { margin-top: 3px; color: var(--muted); font-size: 10px; }
.value-segments { display: grid; gap: 12px; }
.value-segments > div { display: grid; grid-template-columns: 1fr auto; gap: 5px 10px; }
.value-segments span { color: var(--muted); font-size: 11px; }
.value-segments strong { font-size: 11px; font-variant-numeric: tabular-nums; }
.value-segments i { grid-column: 1/-1; }

.recent-row {
  display: grid;
  min-height: 31px;
  grid-template-columns: 7px minmax(0, 1fr) 68px 48px;
  align-items: center;
  gap: 8px;
  border-bottom: 1px solid var(--border);
}

.recent-row:last-child { border-bottom: 0; }
.asset-mark { width: 6px; height: 6px; border-radius: 50%; background: var(--primary); }
.recent-row > div { display: flex; min-width: 0; flex-direction: column; gap: 1px; }
.recent-row strong,
.recent-row small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.recent-row strong { font-size: 11px; }
.recent-row small,
.recent-row > span { color: var(--muted); font-size: 9px; }
.recent-row > b { font-size: 10px; text-align: right; }

/* Matrix layout */
.matrix-board {
  display: grid;
  grid-template-columns: 1.05fr 1fr 1.2fr .95fr;
  grid-template-rows: 215px 300px 220px;
  gap: 10px;
}

.matrix-card {
  position: relative;
  padding: 16px;
}

.matrix-card .panel-heading {
  min-height: 42px;
  padding: 0;
}

.matrix-total {
  grid-column: 1;
  grid-row: 1;
  background: #e7f4ff;
}

.matrix-total > span,
.matrix-value > div > span,
.matrix-retention > span,
.matrix-depreciation > span {
  color: var(--muted);
  font-size: 11px;
}

.matrix-total > strong,
.matrix-value > div > strong,
.matrix-retention > strong,
.matrix-depreciation > strong {
  display: block;
  margin: 8px 0 5px;
  color: var(--text);
  font-size: 42px;
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.matrix-total > small,
.matrix-value > div > small,
.matrix-retention > small,
.matrix-depreciation > small {
  color: var(--muted);
  font-size: 10px;
}

.matrix-total > .el-icon,
.matrix-depreciation > .el-icon {
  position: absolute;
  right: 20px;
  bottom: 16px;
  color: #0284c7;
  font-size: 72px;
  opacity: .24;
}

.matrix-value {
  grid-column: 2/4;
  grid-row: 1;
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  gap: 28px;
  background: #f0edff;
}

.matrix-value > div:first-child {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.matrix-retention-bar {
  display: flex;
  width: min(52%, 420px);
  flex-direction: column;
  justify-content: center;
}

.matrix-retention-bar span { display: flex; justify-content: space-between; margin-bottom: 10px; }
.matrix-retention-bar span b { font-size: 11px; }
.matrix-retention-bar span strong { color: var(--primary); font-size: 18px; }
.matrix-retention-bar > i { height: 8px; background: rgb(109 93 251 / 13%); }

.matrix-status {
  grid-column: 4;
  grid-row: 1/3;
}

.matrix-status-chart { margin: 0 -2px; }
.matrix-status-list { display: grid; gap: 2px; }
.matrix-status-list > span {
  display: grid;
  grid-template-columns: 7px 1fr auto;
  align-items: center;
  gap: 7px;
  padding: 7px 2px;
  border-bottom: 1px solid var(--border);
  font-size: 10px;
}
.matrix-status-list i { width: 6px; height: 6px; border-radius: 50%; }
.matrix-status-list b { color: var(--muted); font-weight: 500; }

.matrix-category {
  grid-column: 1/3;
  grid-row: 2;
}

.category-matrix {
  display: grid;
  height: calc(100% - 48px);
  grid-template-columns: 1.5fr repeat(3, 1fr);
  grid-template-rows: 1fr 1fr;
  gap: 8px;
  padding-top: 8px;
}

.category-tile {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: center;
  padding: 12px;
  border: 1px solid color-mix(in srgb, var(--category-color) 25%, var(--border));
  border-radius: 9px;
  background: color-mix(in srgb, var(--category-color) 10%, #fff);
}

.category-tile.featured {
  grid-row: 1/3;
}

.category-tile span { overflow: hidden; color: var(--muted); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.category-tile strong { margin: 8px 0 4px; color: var(--category-color); font-size: 23px; font-variant-numeric: tabular-nums; }
.category-tile strong small { font-size: 10px; font-weight: 500; }
.category-tile b { color: var(--text); font-size: 10px; }

.matrix-locations {
  grid-column: 3;
  grid-row: 2;
}

.matrix-location-list { padding-top: 8px; }
.matrix-location-list > div {
  display: grid;
  min-height: 37px;
  grid-template-columns: 24px 1fr 28px;
  align-items: center;
  gap: 8px;
}
.matrix-location-list > div > span { color: var(--primary); font: 600 10px/1 ui-monospace, SFMono-Regular, Menlo, monospace; }
.matrix-location-list > div > div { display: grid; grid-template-columns: 1fr .9fr; align-items: center; gap: 8px; }
.matrix-location-list b { overflow: hidden; font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.matrix-location-list strong { font-size: 11px; text-align: right; }

.matrix-recent {
  grid-column: 1/3;
  grid-row: 3;
}

.matrix-recent-list {
  display: grid;
  grid-template-columns: 1fr 1fr;
  column-gap: 24px;
  padding-top: 6px;
}

.matrix-recent-list > div {
  display: grid;
  min-height: 47px;
  grid-template-columns: 84px minmax(0, 1fr) 64px;
  align-items: center;
  gap: 8px;
  border-bottom: 1px solid var(--border);
}
.matrix-recent-list span { overflow: hidden; color: var(--primary); font: 10px/1 ui-monospace, SFMono-Regular, Menlo, monospace; text-overflow: ellipsis; white-space: nowrap; }
.matrix-recent-list b { overflow: hidden; font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.matrix-recent-list small { display: none; }
.matrix-recent-list strong { color: var(--muted); font-size: 10px; text-align: right; }

.matrix-retention {
  grid-column: 3;
  grid-row: 3;
  background: #fff4d6;
}

.matrix-retention > strong { font-size: 38px; }
.mini-ring {
  position: absolute;
  right: 18px;
  bottom: 16px;
  width: 70px;
  height: 70px;
  place-items: center;
}
.mini-ring::before { width: 50px; height: 50px; background: #fff4d6; }

.matrix-depreciation {
  grid-column: 4;
  grid-row: 3;
  background: var(--primary);
}

.matrix-depreciation > span,
.matrix-depreciation > small { color: rgb(255 255 255 / 76%); }
.matrix-depreciation > strong { color: #fff; font-size: 32px; }
.matrix-depreciation > .el-icon { color: #fff; }

.dashboard-mode-enter-active,
.dashboard-mode-leave-active {
  transition: opacity 180ms ease, transform 180ms cubic-bezier(.22, 1, .36, 1);
}

.dashboard-mode-enter-from,
.dashboard-mode-leave-to {
  opacity: 0;
  transform: translateY(3px);
}

@keyframes orbit-spin {
  to { transform: translate(-50%, -50%) rotate(360deg); }
}

@keyframes orbit-spin-reverse {
  to { transform: translate(-50%, -50%) rotate(-360deg); }
}

@media (max-width: 1400px) {
  .dashboard-content { min-width: 1060px; }
  .panorama-main { grid-template-columns: minmax(245px, 22%) minmax(470px, 1fr) minmax(255px, 23%); }
  .value-metrics { grid-template-columns: 1fr; }
  .orbit-large { width: 340px; height: 340px; }
  .orbit-small { width: 240px; height: 240px; }
  .matrix-board { grid-template-columns: 1fr 1fr 1.15fr .9fr; }
}

@media (max-width: 1100px) {
  .dashboard-content { min-width: 0; }
  .panorama-main { grid-template-columns: 1fr 1fr; }
  .orbit-panel { grid-column: 1/-1; grid-row: 1; min-height: 480px; }
  .value-metrics { grid-template-columns: 1fr 1fr; }
  .panorama-footer { grid-template-columns: 1fr; }
  .matrix-board { grid-template-columns: 1fr 1fr; grid-template-rows: none; }
  .matrix-card { grid-column: auto; grid-row: auto; min-height: 210px; }
  .matrix-value,
  .matrix-category,
  .matrix-recent { grid-column: 1/-1; }
  .matrix-status { min-height: 460px; }
}

@media (max-width: 760px) {
  .dashboard-content { padding: 14px; }
  .view-mode-switch { width: 100%; order: -1; }
  .view-mode-switch button { flex: 1; justify-content: center; }
  .update-time { width: 100%; }
  .panorama-main,
  .matrix-board { grid-template-columns: 1fr; }
  .orbit-panel,
  .matrix-value,
  .matrix-category,
  .matrix-recent { grid-column: auto; }
  .value-metrics { grid-template-columns: 1fr; }
  .orbit-large { width: 310px; height: 310px; }
  .orbit-small { width: 220px; height: 220px; }
  .stat-a { left: 4%; }
  .stat-b { right: 4%; }
  .stat-c { left: 4%; }
  .stat-d { right: 4%; }
  .matrix-value { flex-direction: column; }
  .matrix-retention-bar { width: 100%; }
  .category-matrix { grid-template-columns: 1fr 1fr; grid-template-rows: auto; }
  .category-tile.featured { grid-row: auto; }
  .matrix-recent-list { grid-template-columns: 1fr; }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: .01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: .01ms !important;
  }
}
</style>
