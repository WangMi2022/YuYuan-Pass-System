<template>
  <main ref="screenRef" class="asset-dashboard">
    <div class="dashboard-content">
      <AppPageHeader
        title-id="asset-dashboard-title"
        title="资产可视化大屏"
        description="实时汇总资产分类、数量、原值、当前估值与使用状态。"
      >
        <template #actions>
          <span class="update-time"><i /> 数据更新于 {{ updateTime }}</span>
          <el-button :icon="Refresh" :loading="loading" @click="loadDashboard">刷新</el-button>
          <el-button type="primary" :icon="FullScreen" @click="toggleFullScreen">全屏展示</el-button>
        </template>
      </AppPageHeader>

      <section class="kpi-grid" aria-label="资产核心指标">
        <article v-for="item in kpis" :key="item.label" class="kpi-card" :style="{ '--kpi-color': item.color }">
          <div class="kpi-icon"><el-icon><component :is="item.icon" /></el-icon></div>
          <div>
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
            <small>{{ item.hint }}</small>
          </div>
        </article>
      </section>

      <section class="dashboard-grid">
        <article class="dashboard-card category-chart-card">
          <header class="card-header">
            <div><h2>分类资产价值</h2><p>各资产分类当前估值，单位：万元</p></div>
            <el-tag type="success" effect="plain">{{ dashboard.categoryCount }} 个分类</el-tag>
          </header>
          <div v-if="hasCategoryData" class="chart-wrap">
            <Chart :options="categoryOption" height="340px" />
          </div>
          <el-empty v-else description="登记资产后显示分类价值" :image-size="92" />
          <div
            class="category-value-list"
            :class="{ 'is-scrolling': categoryListShouldScroll }"
            aria-label="分类统计明细"
          >
            <div
              class="category-value-track"
              :style="{ '--category-scroll-duration': `${categoryScrollDuration}s` }"
            >
              <div
                v-for="(group, groupIndex) in categoryScrollGroups"
                :key="groupIndex"
                class="category-value-group"
                :aria-hidden="groupIndex > 0 ? 'true' : undefined"
              >
                <div v-for="item in group" :key="`${groupIndex}-${item.categoryId}`" class="category-value-row">
                  <span class="legend-dot" :style="{ background: item.color }" />
                  <span class="category-label">{{ item.categoryName }}</span>
                  <span>{{ item.quantity }} 件</span>
                  <strong>{{ formatCurrency(item.currentValue) }}</strong>
                </div>
              </div>
            </div>
          </div>
        </article>

        <article class="dashboard-card status-card">
          <header class="card-header">
            <div><h2>资产状态构成</h2><p>按实物数量统计</p></div>
          </header>
          <div v-if="hasStatusData" class="chart-wrap status-chart">
            <Chart :options="statusOption" height="260px" />
          </div>
          <el-empty v-else description="暂无状态数据" :image-size="78" />
          <div class="status-list">
            <div v-for="item in statusRows" :key="item.status">
              <span><i :style="{ background: `var(${item.colorVar})` }" />{{ item.label }}</span>
              <strong>{{ item.quantity }}</strong>
              <small>{{ formatPercent(item.quantity, dashboard.totalQuantity) }}</small>
            </div>
          </div>
        </article>

        <article class="dashboard-card location-card">
          <header class="card-header">
            <div><h2>资产位置排行</h2><p>按实物数量展示前 8 个位置</p></div>
          </header>
          <div v-if="dashboard.locationSummary.length" class="location-list">
            <div v-for="(item, index) in dashboard.locationSummary" :key="item.location" class="location-row">
              <span class="rank" :class="{ top: index < 3 }">{{ String(index + 1).padStart(2, '0') }}</span>
              <div class="location-progress">
                <div><strong>{{ item.location }}</strong><span>{{ formatCurrency(item.value) }}</span></div>
                <el-progress :percentage="locationPercent(item.quantity)" :show-text="false" :stroke-width="5" color="var(--na-primary)" />
              </div>
              <strong class="location-quantity">{{ item.quantity }} 件</strong>
            </div>
          </div>
          <el-empty v-else description="补充资产位置后显示排行" :image-size="78" />
        </article>
      </section>

      <section class="bottom-grid">
        <article class="dashboard-card category-overview">
          <header class="card-header">
            <div><h2>分类概览</h2><p>快速识别资产规模与结构</p></div>
          </header>
          <div class="category-cards">
            <div v-for="item in dashboard.categorySummary" :key="item.categoryId" class="mini-category-card">
              <span class="mini-accent" :style="{ background: item.color }" />
              <div class="mini-title"><strong>{{ item.categoryName }}</strong><span>{{ item.assetKinds }} 种档案</span></div>
              <strong class="mini-quantity">{{ item.quantity }}<small> 件</small></strong>
              <span class="mini-value">{{ formatCompactCurrency(item.currentValue) }}</span>
            </div>
          </div>
        </article>

        <article class="dashboard-card recent-card">
          <header class="card-header">
            <div><h2>最近登记</h2><p>最新录入的资产档案</p></div>
          </header>
          <el-table :data="dashboard.recentAssets" size="small" class="recent-table">
            <el-table-column label="资产" min-width="170">
              <template #default="{ row }"><div class="asset-cell"><strong>{{ row.name }}</strong><span>{{ row.assetCode }}</span></div></template>
            </el-table-column>
            <el-table-column label="分类" min-width="110"><template #default="{ row }">{{ row.category?.name || '—' }}</template></el-table-column>
            <el-table-column label="数量" width="80" align="right"><template #default="{ row }">{{ row.quantity }} {{ row.unit }}</template></el-table-column>
            <el-table-column label="当前估值" width="130" align="right"><template #default="{ row }"><strong class="table-money">{{ formatCurrency(row.currentValue) }}</strong></template></el-table-column>
            <template #empty><el-empty description="暂无资产" :image-size="60" /></template>
          </el-table>
        </article>
      </section>
    </div>
  </main>
</template>

<script setup>
import { computed, markRaw, onMounted, ref } from 'vue'
import { Box, Coin, CollectionTag, DataAnalysis, FullScreen, Goods, Refresh } from '@element-plus/icons-vue'
import Chart from '@/components/charts/index.vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { chartTheme } from '@/components/charts/theme'
import { formatCurrency, formatCompactCurrency, formatNumber, formatPercent } from '@/utils/format'
import { getAssetDashboard } from '@/plugin/asset/api/asset'
import { useAppStore } from '@/pinia'

defineOptions({ name: 'AssetDashboard' })

const appStore = useAppStore()
const screenRef = ref()
const loading = ref(false)
const updateTime = ref('—')
const dashboard = ref({
  assetKinds: 0, totalQuantity: 0, categoryCount: 0, originalValue: 0, currentValue: 0, depreciation: 0,
  categorySummary: [], statusSummary: [], locationSummary: [], recentAssets: []
})

const retentionRate = computed(() => dashboard.value.originalValue ? `${((dashboard.value.currentValue / dashboard.value.originalValue) * 100).toFixed(1)}%` : '0.0%')

const kpis = computed(() => [
  { label: '资产实物总量', value: formatNumber(dashboard.value.totalQuantity), hint: `${dashboard.value.assetKinds} 种资产档案`, icon: markRaw(Box), color: 'var(--na-chart-1)' },
  { label: '资产分类', value: dashboard.value.categoryCount, hint: '统一分类统计口径', icon: markRaw(CollectionTag), color: 'var(--na-chart-5)' },
  { label: '资产原值', value: formatCompactCurrency(dashboard.value.originalValue), hint: '数量 × 采购单价', icon: markRaw(Coin), color: 'var(--na-chart-2)' },
  { label: '当前估值', value: formatCompactCurrency(dashboard.value.currentValue), hint: `价值保有率 ${retentionRate.value}`, icon: markRaw(Goods), color: 'var(--na-chart-3)' },
  { label: '累计价值减少', value: formatCompactCurrency(dashboard.value.depreciation), hint: '原值与当前估值差额', icon: markRaw(DataAnalysis), color: 'var(--na-chart-4)' }
])

/* colorVar drives DOM swatches via var(); ECharts gets concrete values from chartTheme(). */
const statusConfig = {
  pending_inbound: { label: '待入库', colorVar: '--na-info', themeKey: 'info' },
  in_use: { label: '使用中', colorVar: '--na-success', themeKey: 'success' },
  idle: { label: '闲置', colorVar: '--na-muted-foreground', themeKey: 'muted' },
  maintenance: { label: '维修中', colorVar: '--na-warning', themeKey: 'warning' },
  retired: { label: '已处置', colorVar: '--na-danger', themeKey: 'danger' }
}
const statusOrder = ['pending_inbound', 'idle', 'in_use', 'maintenance', 'retired']
const activeCategories = computed(() => dashboard.value.categorySummary.filter((item) => Number(item.quantity) > 0))
const hasCategoryData = computed(() => activeCategories.value.length > 0)
const categoryListShouldScroll = computed(() => activeCategories.value.length > 4)
const categoryScrollGroups = computed(() => categoryListShouldScroll.value
  ? [activeCategories.value, activeCategories.value]
  : [activeCategories.value])
const categoryScrollDuration = computed(() => Math.max(12, activeCategories.value.length * 2.4))
const statusRows = computed(() => dashboard.value.statusSummary
  .map((item) => ({ ...item, ...(statusConfig[item.status] || { label: item.status, colorVar: '--na-muted-foreground', themeKey: 'muted' }) }))
  .sort((a, b) => statusOrder.indexOf(a.status) - statusOrder.indexOf(b.status)))
const hasStatusData = computed(() => statusRows.value.some((item) => item.quantity > 0))

/* Recomputed on dark-mode / primary-color changes so canvas charts follow the theme. */
const theme = computed(() => {
  void appStore.isDark
  void appStore.config.primaryColor
  return chartTheme()
})

const categoryOption = computed(() => ({
  animationDuration: 450,
  grid: { top: 18, right: 26, bottom: 55, left: 62, containLabel: true },
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' }, valueFormatter: (value) => `${Number(value).toFixed(2)} 万元` },
  xAxis: { type: 'category', data: activeCategories.value.map((item) => item.categoryName), axisLabel: { color: theme.value.label, rotate: activeCategories.value.length > 5 ? 24 : 0, interval: 0 }, axisLine: { lineStyle: { color: theme.value.grid } }, axisTick: { show: false } },
  yAxis: { type: 'value', name: '万元', nameTextStyle: { color: theme.value.label }, axisLabel: { color: theme.value.label }, splitLine: { lineStyle: { color: theme.value.grid, type: 'dashed' } } },
  series: [{
    name: '当前估值', type: 'bar', barMaxWidth: 42,
    data: activeCategories.value.map((item) => ({ value: Number(item.currentValue || 0) / 10000, itemStyle: { color: item.color, borderRadius: [3, 3, 0, 0] } })),
    label: { show: true, position: 'top', color: theme.value.label, formatter: ({ value }) => value > 0 ? value.toFixed(1) : '' }
  }]
}))

const statusOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: ({ name, value, percent: p }) => `${name}<br/>${value} 件（${p}%）` },
  legend: { show: false },
  series: [{
    type: 'pie', radius: ['58%', '78%'], center: ['50%', '48%'], avoidLabelOverlap: true,
    itemStyle: { borderColor: theme.value.surface, borderWidth: 4, borderRadius: 4 },
    label: { show: false },
    data: statusRows.value.map((item) => ({ name: item.label, value: item.quantity, itemStyle: { color: theme.value[item.themeKey] || theme.value.muted } }))
  }],
  graphic: [{ type: 'text', left: 'center', top: '38%', style: { text: String(dashboard.value.totalQuantity), fill: theme.value.text, fontSize: 28, fontWeight: 700 } }, { type: 'text', left: 'center', top: '54%', style: { text: '资产总量', fill: theme.value.label, fontSize: 12 } }]
}))

const locationPercent = (quantity) => {
  const max = Math.max(...dashboard.value.locationSummary.map((item) => Number(item.quantity || 0)), 1)
  return Math.round((Number(quantity || 0) / max) * 100)
}

const loadDashboard = async () => {
  loading.value = true
  try {
    const res = await getAssetDashboard()
    if (res.code === 0) {
      dashboard.value = { ...dashboard.value, ...res.data,
        categorySummary: res.data.categorySummary || [], statusSummary: res.data.statusSummary || [],
        locationSummary: res.data.locationSummary || [], recentAssets: res.data.recentAssets || [] }
      updateTime.value = new Date().toLocaleString('zh-CN', { hour12: false })
    }
  } finally { loading.value = false }
}

const toggleFullScreen = async () => {
  if (!document.fullscreenElement) await screenRef.value?.requestFullscreen()
  else await document.exitFullscreen()
}

onMounted(loadDashboard)
</script>

<style scoped lang="scss">
.asset-dashboard {
  --surface: var(--na-card); --text: var(--na-foreground); --muted: var(--na-muted-foreground); --border: var(--na-border);
  position: relative;
  min-height: 100%; overflow: auto; background: var(--na-background); color: var(--text);
}
.asset-dashboard:fullscreen { min-height: 100vh; }
.dashboard-content { padding: 24px; }
.update-time { display: inline-flex; align-items: center; gap: 7px; margin-right: 4px; color: var(--muted); font-size: 12px; }
.update-time i { width: 6px; height: 6px; border-radius: 50%; background: var(--na-success); }
.kpi-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 12px; margin-bottom: 12px; }
.kpi-card {
  display: flex; min-width: 0; min-height: 108px; align-items: center; gap: 14px; padding: 16px;
  border: 1px solid var(--border); border-radius: var(--na-radius);
  background: var(--surface);
  box-shadow: var(--na-shadow-sm);
  transition: border-color 150ms ease, transform 150ms ease;
}
.kpi-card:hover { border-color: color-mix(in srgb, var(--kpi-color) 36%, var(--border)); transform: translateY(-1px); }
.kpi-icon {
  display: grid; width: 40px; height: 40px; place-items: center; flex: 0 0 40px;
  border-radius: 10px; font-size: 18px;
  color: var(--kpi-color);
  background: color-mix(in srgb, var(--kpi-color) 12%, transparent);
}
.kpi-card > div:last-child { display: flex; min-width: 0; flex-direction: column; }
.kpi-card span { color: var(--muted); font-size: 12px; }
.kpi-card strong { overflow: hidden; margin: 4px 0 2px; color: var(--text); font-size: 22px; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.kpi-card small { color: var(--muted); font-size: 11px; }
.dashboard-grid { display: grid; grid-template-columns: minmax(0, 1.7fr) minmax(280px, .72fr) minmax(310px, .82fr); gap: 12px; }
.dashboard-card {
  overflow: hidden; border: 1px solid var(--border); border-radius: var(--na-radius);
  background: var(--surface);
  box-shadow: var(--na-shadow-sm);
}
.card-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 16px 18px 0; }
.card-header h2 { margin: 0; font-size: 14px; font-weight: 600; }
.card-header p { margin: 3px 0 0; color: var(--muted); font-size: 12px; }
.chart-wrap { position: relative; padding: 0 6px; }
.category-value-list { max-height: 158px; overflow: hidden; margin: 0 18px 14px; }
.category-value-list.is-scrolling { height: 158px; mask-image: linear-gradient(to bottom, transparent 0, #000 7px, #000 calc(100% - 7px), transparent 100%); }
.category-value-track { will-change: transform; }
.category-value-list.is-scrolling .category-value-track { animation: category-value-scroll var(--category-scroll-duration) linear infinite; }
.category-value-list.is-scrolling:hover .category-value-track { animation-play-state: paused; }
.category-value-group { width: 100%; }
.category-value-row { display: grid; min-height: 35px; grid-template-columns: 10px minmax(100px, 1fr) 70px 130px; align-items: center; gap: 8px; border-top: 1px solid var(--border); font-size: 12px; }
.category-value-group:first-child .category-value-row:first-child { border-top: 0; }
.legend-dot { width: 7px; height: 7px; border-radius: 50%; }
.category-label { color: var(--text); font-weight: 550; }
.category-value-row > span:nth-child(3) { color: var(--muted); text-align: right; }
.category-value-row strong { text-align: right; font-variant-numeric: tabular-nums; }
.status-list { display: grid; grid-template-columns: 1fr 1fr; gap: 0 12px; padding: 0 18px 16px; }
.status-list > div { display: grid; grid-template-columns: 1fr auto; gap: 3px 8px; padding: 9px 2px; border-bottom: 1px solid var(--border); }
.status-list span { display: flex; align-items: center; gap: 6px; color: var(--muted); font-size: 12px; }
.status-list span i { width: 7px; height: 7px; border-radius: 50%; }
.status-list strong { font-variant-numeric: tabular-nums; }
.status-list small { grid-column: 2; color: var(--muted); font-size: 11px; text-align: right; }
.location-list { padding: 12px 18px 18px; }
.location-row { display: flex; min-height: 50px; align-items: center; gap: 10px; border-bottom: 1px solid var(--border); }
.location-row:last-child { border-bottom: 0; }
.rank { width: 24px; color: var(--muted); font: 600 11px/1 ui-monospace, SFMono-Regular, Menlo, monospace; }
.rank.top { color: var(--na-primary); }
.location-progress { min-width: 0; flex: 1; }
.location-progress > div { display: flex; justify-content: space-between; gap: 8px; margin-bottom: 5px; }
.location-progress strong { overflow: hidden; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.location-progress span { color: var(--muted); font-size: 11px; }
.location-quantity { min-width: 54px; font-size: 12px; font-variant-numeric: tabular-nums; text-align: right; }
.bottom-grid { display: grid; grid-template-columns: 1.15fr 1fr; gap: 12px; margin-top: 12px; }
.category-cards { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); padding: 14px 18px 18px; }
.mini-category-card { position: relative; display: grid; min-height: 94px; grid-template-columns: 1fr auto; gap: 7px; padding: 12px 14px; border-right: 1px solid var(--border); border-bottom: 1px solid var(--border); background: transparent; }
.mini-category-card:nth-child(4n) { border-right: 0; }
.mini-accent { position: absolute; inset: 12px auto 12px 0; width: 2px; border-radius: 2px; }
.mini-title { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.mini-title strong { overflow: hidden; font-size: 12.5px; text-overflow: ellipsis; white-space: nowrap; }
.mini-title span { color: var(--muted); font-size: 11px; }
.mini-quantity { align-self: end; font-size: 20px; font-variant-numeric: tabular-nums; }
.mini-quantity small { color: var(--muted); font-size: 11px; font-weight: 400; }
.mini-value { align-self: end; color: var(--na-primary); font-size: 12px; font-weight: 600; text-align: right; }
.recent-table { margin-top: 8px; }
.asset-cell { display: flex; flex-direction: column; gap: 3px; }
.asset-cell span { color: var(--muted); font: 11px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; }
.table-money { color: var(--na-primary); font-variant-numeric: tabular-nums; }
@media (max-width: 1400px) { .kpi-grid { grid-template-columns: repeat(3, 1fr); } .dashboard-grid { grid-template-columns: 1.5fr 1fr; } .location-card { grid-column: 1/-1; } .location-list { display: grid; grid-template-columns: 1fr 1fr; column-gap: 24px; } .category-cards { grid-template-columns: repeat(3, 1fr); } .mini-category-card:nth-child(4n) { border-right: 1px solid var(--border); } .mini-category-card:nth-child(3n) { border-right: 0; } }
@media (max-width: 900px) { .kpi-grid { grid-template-columns: repeat(2, 1fr); } .dashboard-grid, .bottom-grid { grid-template-columns: 1fr; } .location-card { grid-column: auto; } .category-cards { grid-template-columns: repeat(2, 1fr); } .mini-category-card:nth-child(3n) { border-right: 1px solid var(--border); } .mini-category-card:nth-child(2n) { border-right: 0; } }
@media (max-width: 560px) { .dashboard-content { padding: 14px; } .kpi-grid { grid-template-columns: 1fr; } .location-list { grid-template-columns: 1fr; } .category-cards { grid-template-columns: 1fr; } .mini-category-card { border-right: 0; } .category-value-row { grid-template-columns: 10px 1fr 55px; } .category-value-row strong { grid-column: 2/-1; } .update-time { width: 100%; margin-bottom: 4px; } }
@keyframes category-value-scroll { to { transform: translateY(-50%); } }
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { animation-duration: .01ms !important; transition-duration: .01ms !important; } .category-value-list.is-scrolling { overflow-y: auto; mask-image: none; } .category-value-list.is-scrolling .category-value-track { animation: none !important; transform: none; } }
</style>
