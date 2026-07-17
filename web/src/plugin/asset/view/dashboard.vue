<template>
  <main ref="screenRef" class="asset-dashboard">
    <header class="dashboard-header">
      <div>
        <p class="eyebrow">ASSET INTELLIGENCE</p>
        <h1>资产可视化大屏</h1>
        <p>实时汇总资产分类、数量、原值、当前估值与使用状态。</p>
      </div>
      <div class="header-actions">
        <span class="update-time"><i /> 数据更新于 {{ updateTime }}</span>
        <el-button :icon="Refresh" :loading="loading" @click="loadDashboard">刷新</el-button>
        <el-button type="primary" :icon="FullScreen" @click="toggleFullScreen">全屏展示</el-button>
      </div>
    </header>

    <div class="dashboard-flow-rail" aria-hidden="true">
      <span />
    </div>

    <section class="kpi-grid" aria-label="资产核心指标">
      <article v-for="item in kpis" :key="item.label" class="kpi-card" :class="item.tone">
        <div class="kpi-icon" :class="item.tone"><el-icon><component :is="item.icon" /></el-icon></div>
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
                <strong>{{ currency(item.currentValue) }}</strong>
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
            <span><i :style="{ background: item.color }" />{{ item.label }}</span>
            <strong>{{ item.quantity }}</strong>
            <small>{{ percent(item.quantity, dashboard.totalQuantity) }}</small>
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
              <div><strong>{{ item.location }}</strong><span>{{ currency(item.value) }}</span></div>
              <el-progress :percentage="locationPercent(item.quantity)" :show-text="false" :stroke-width="6" color="#20aaa6" />
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
            <span class="mini-value">{{ shortCurrency(item.currentValue) }}</span>
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
          <el-table-column label="当前估值" width="130" align="right"><template #default="{ row }"><strong class="table-money">{{ currency(row.currentValue) }}</strong></template></el-table-column>
          <template #empty><el-empty description="暂无资产" :image-size="60" /></template>
        </el-table>
      </article>
    </section>
  </main>
</template>

<script setup>
import { computed, markRaw, onMounted, ref } from 'vue'
import { Box, Coin, CollectionTag, DataAnalysis, FullScreen, Goods, Refresh } from '@element-plus/icons-vue'
import Chart from '@/components/charts/index.vue'
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

const currency = (value) => new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY', minimumFractionDigits: 2 }).format(Number(value || 0))
const shortCurrency = (value) => `¥ ${new Intl.NumberFormat('zh-CN', { notation: 'compact', maximumFractionDigits: 1 }).format(Number(value || 0))}`
const percent = (value, total) => `${total ? ((value / total) * 100).toFixed(1) : '0.0'}%`
const retentionRate = computed(() => dashboard.value.originalValue ? `${((dashboard.value.currentValue / dashboard.value.originalValue) * 100).toFixed(1)}%` : '0.0%')

const kpis = computed(() => [
  { label: '资产实物总量', value: new Intl.NumberFormat('zh-CN').format(dashboard.value.totalQuantity), hint: `${dashboard.value.assetKinds} 种资产档案`, icon: markRaw(Box), tone: 'slate' },
  { label: '资产分类', value: dashboard.value.categoryCount, hint: '统一分类统计口径', icon: markRaw(CollectionTag), tone: 'violet' },
  { label: '资产原值', value: shortCurrency(dashboard.value.originalValue), hint: '数量 × 采购单价', icon: markRaw(Coin), tone: 'blue' },
  { label: '当前估值', value: shortCurrency(dashboard.value.currentValue), hint: `价值保有率 ${retentionRate.value}`, icon: markRaw(Goods), tone: 'green' },
  { label: '累计价值减少', value: shortCurrency(dashboard.value.depreciation), hint: '原值与当前估值差额', icon: markRaw(DataAnalysis), tone: 'amber' }
])

const statusConfig = {
  pending_inbound: { label: '待入库', color: '#2563eb' },
  in_use: { label: '使用中', color: '#059669' },
  idle: { label: '闲置', color: '#64748b' },
  maintenance: { label: '维修中', color: '#d97706' },
  retired: { label: '已处置', color: '#dc2626' }
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
  .map((item) => ({ ...item, ...(statusConfig[item.status] || { label: item.status, color: '#64748b' }) }))
  .sort((a, b) => statusOrder.indexOf(a.status) - statusOrder.indexOf(b.status)))
const hasStatusData = computed(() => statusRows.value.some((item) => item.quantity > 0))
const chartText = computed(() => appStore.isDark ? '#aeb9ba' : '#7d8995')
const chartGrid = computed(() => appStore.isDark ? '#3b4548' : '#edf1f3')

const categoryOption = computed(() => ({
  animationDuration: 450,
  grid: { top: 18, right: 26, bottom: 55, left: 62, containLabel: true },
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' }, valueFormatter: (value) => `${Number(value).toFixed(2)} 万元` },
  xAxis: { type: 'category', data: activeCategories.value.map((item) => item.categoryName), axisLabel: { color: chartText.value, rotate: activeCategories.value.length > 5 ? 24 : 0, interval: 0 }, axisLine: { lineStyle: { color: chartGrid.value } }, axisTick: { show: false } },
  yAxis: { type: 'value', name: '万元', nameTextStyle: { color: chartText.value }, axisLabel: { color: chartText.value }, splitLine: { lineStyle: { color: chartGrid.value, type: 'dashed' } } },
  series: [{
    name: '当前估值', type: 'bar', barMaxWidth: 42,
    data: activeCategories.value.map((item) => ({ value: Number(item.currentValue || 0) / 10000, itemStyle: { color: item.color, borderRadius: [2, 2, 0, 0] } })),
    label: { show: true, position: 'top', color: chartText.value, formatter: ({ value }) => value > 0 ? value.toFixed(1) : '' }
  }]
}))

const statusOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: ({ name, value, percent: p }) => `${name}<br/>${value} 件（${p}%）` },
  legend: { show: false },
  series: [{
    type: 'pie', radius: ['58%', '78%'], center: ['50%', '48%'], avoidLabelOverlap: true,
    itemStyle: { borderColor: appStore.isDark ? '#292f32' : '#ffffff', borderWidth: 4, borderRadius: 3 },
    label: { show: false },
    data: statusRows.value.map((item) => ({ name: item.label, value: item.quantity, itemStyle: { color: item.color } }))
  }],
  graphic: [{ type: 'text', left: 'center', top: '38%', style: { text: String(dashboard.value.totalQuantity), fill: appStore.isDark ? '#eef3f3' : '#27313c', fontSize: 28, fontWeight: 700 } }, { type: 'text', left: 'center', top: '54%', style: { text: '资产总量', fill: chartText.value, fontSize: 12 } }]
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
  --bg: var(--na-background); --surface: var(--na-card); --surface-2: var(--na-muted); --text: var(--na-foreground); --muted: var(--na-muted-foreground); --border: var(--na-border);
  min-height: 100%; overflow: auto; padding: 20px; background: var(--bg); color: var(--text);
}
.asset-dashboard:fullscreen { min-height: 100vh; }
.dashboard-header { display: flex; align-items: flex-end; justify-content: space-between; gap: 20px; margin-bottom: 14px; }
.eyebrow { margin: 0 0 4px; color: var(--na-primary); font: 600 11px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; }
h1 { margin: 0; font-size: 24px; }
.dashboard-header p:last-child { margin: 6px 0 0; color: var(--muted); font-size: 13px; }
.header-actions { display: flex; align-items: center; justify-content: flex-end; flex-wrap: wrap; gap: 8px; }
.update-time { display: inline-flex; align-items: center; gap: 7px; margin-right: 4px; color: var(--muted); font-size: 11px; }
.update-time i { width: 6px; height: 6px; border-radius: 50%; background: var(--na-success); }
.dashboard-flow-rail { position: relative; height: 1px; margin-bottom: 14px; background: var(--border); }
.dashboard-flow-rail span { position: absolute; top: -1px; left: 0; width: 86px; height: 2px; background: var(--na-primary); }
.kpi-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 12px; margin-bottom: 12px; }
.kpi-card { display: flex; min-width: 0; min-height: 112px; align-items: center; gap: 13px; padding: 16px; border: 1px solid var(--border); border-radius: var(--na-radius); background: var(--surface); box-shadow: var(--na-shadow-sm); }
.kpi-icon { display: grid; width: 42px; height: 42px; place-items: center; flex: 0 0 42px; border-radius: 50%; font-size: 19px; }
.kpi-icon.slate { color: #5d6974; background: #f0f3f5; }
.kpi-icon.violet { color: #7378b8; background: #f0f0fa; }
.kpi-icon.blue { color: #4c8fb8; background: #eaf4f8; }
.kpi-icon.green { color: #168b7b; background: #e6f5f1; }
.kpi-icon.amber { color: #b9812b; background: #faf3e5; }
.kpi-card > div:last-child { display: flex; min-width: 0; flex-direction: column; }
.kpi-card span { color: var(--muted); font-size: 12px; }
.kpi-card strong { overflow: hidden; margin: 4px 0 2px; color: var(--text); font-size: 24px; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.kpi-card small { color: var(--muted); font-size: 10px; }
.dashboard-grid { display: grid; grid-template-columns: minmax(0, 1.7fr) minmax(280px, .72fr) minmax(310px, .82fr); gap: 12px; }
.dashboard-card { overflow: hidden; border: 1px solid var(--border); border-radius: var(--na-radius); background: var(--surface); box-shadow: var(--na-shadow-sm); }
.card-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 16px 18px 0; }
.card-header h2 { margin: 0; font-size: 15px; font-weight: 650; }
.card-header p { margin: 3px 0 0; color: var(--muted); font-size: 11px; }
.chart-wrap { position: relative; padding: 0 6px; }
.category-value-list { max-height: 158px; overflow: hidden; margin: 0 18px 14px; }
.category-value-list.is-scrolling { height: 158px; mask-image: linear-gradient(to bottom, transparent 0, #000 7px, #000 calc(100% - 7px), transparent 100%); }
.category-value-track { will-change: transform; }
.category-value-list.is-scrolling .category-value-track { animation: category-value-scroll var(--category-scroll-duration) linear infinite; }
.category-value-list.is-scrolling:hover .category-value-track { animation-play-state: paused; }
.category-value-group { width: 100%; }
.category-value-row { display: grid; min-height: 35px; grid-template-columns: 10px minmax(100px, 1fr) 70px 130px; align-items: center; gap: 8px; border-top: 1px solid var(--border); font-size: 11px; }
.category-value-group:first-child .category-value-row:first-child { border-top: 0; }
.legend-dot { width: 7px; height: 7px; border-radius: 50%; }
.category-label { color: var(--text); font-weight: 600; }
.category-value-row > span:nth-child(3) { color: var(--muted); text-align: right; }
.category-value-row strong { text-align: right; font-variant-numeric: tabular-nums; }
.status-list { display: grid; grid-template-columns: 1fr 1fr; gap: 0 12px; padding: 0 18px 16px; }
.status-list > div { display: grid; grid-template-columns: 1fr auto; gap: 3px 8px; padding: 9px 2px; border-bottom: 1px solid var(--border); }
.status-list span { display: flex; align-items: center; gap: 6px; color: var(--muted); font-size: 11px; }
.status-list span i { width: 7px; height: 7px; border-radius: 50%; }
.status-list strong { font-variant-numeric: tabular-nums; }
.status-list small { grid-column: 2; color: var(--muted); font-size: 10px; text-align: right; }
.location-list { padding: 12px 18px 18px; }
.location-row { display: flex; min-height: 50px; align-items: center; gap: 10px; border-bottom: 1px solid var(--border); }
.location-row:last-child { border-bottom: 0; }
.rank { width: 24px; color: #9aa5ad; font: 600 11px/1 ui-monospace, SFMono-Regular, Menlo, monospace; }
.rank.top { color: var(--na-primary); }
.location-progress { min-width: 0; flex: 1; }
.location-progress > div { display: flex; justify-content: space-between; gap: 8px; margin-bottom: 5px; }
.location-progress strong { overflow: hidden; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.location-progress span { color: var(--muted); font-size: 10px; }
.location-quantity { min-width: 54px; font-size: 11px; font-variant-numeric: tabular-nums; text-align: right; }
.bottom-grid { display: grid; grid-template-columns: 1.15fr 1fr; gap: 12px; margin-top: 12px; }
.category-cards { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); padding: 14px 18px 18px; }
.mini-category-card { position: relative; display: grid; min-height: 94px; grid-template-columns: 1fr auto; gap: 7px; padding: 12px 14px; border-right: 1px solid var(--border); border-bottom: 1px solid var(--border); background: transparent; }
.mini-category-card:nth-child(4n) { border-right: 0; }
.mini-accent { position: absolute; inset: 12px auto 12px 0; width: 2px; }
.mini-title { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.mini-title strong { overflow: hidden; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.mini-title span { color: var(--muted); font-size: 10px; }
.mini-quantity { align-self: end; font-size: 20px; font-variant-numeric: tabular-nums; }
.mini-quantity small { color: var(--muted); font-size: 10px; font-weight: 400; }
.mini-value { align-self: end; color: var(--na-primary); font-size: 11px; font-weight: 600; text-align: right; }
.recent-table { margin-top: 8px; }
.asset-cell { display: flex; flex-direction: column; gap: 3px; }
.asset-cell span { color: var(--muted); font: 10px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; }
.table-money { color: var(--na-primary); font-variant-numeric: tabular-nums; }
:global(html.dark) .kpi-icon.slate { color: #bec7cb; background: #384044; }
:global(html.dark) .kpi-icon.violet { color: #b7b8dc; background: #39394d; }
:global(html.dark) .kpi-icon.blue { color: #92c4d8; background: #293e46; }
:global(html.dark) .kpi-icon.green { color: #7fd8ca; background: #29433f; }
:global(html.dark) .kpi-icon.amber { color: #d7b579; background: #443a2a; }
@media (max-width: 1400px) { .kpi-grid { grid-template-columns: repeat(3, 1fr); } .dashboard-grid { grid-template-columns: 1.5fr 1fr; } .location-card { grid-column: 1/-1; } .location-list { display: grid; grid-template-columns: 1fr 1fr; column-gap: 24px; } .category-cards { grid-template-columns: repeat(3, 1fr); } .mini-category-card:nth-child(4n) { border-right: 1px solid var(--border); } .mini-category-card:nth-child(3n) { border-right: 0; } }
@media (max-width: 900px) { .dashboard-header { align-items: stretch; flex-direction: column; } .header-actions { justify-content: flex-start; } .kpi-grid { grid-template-columns: repeat(2, 1fr); } .dashboard-grid, .bottom-grid { grid-template-columns: 1fr; } .location-card { grid-column: auto; } .category-cards { grid-template-columns: repeat(2, 1fr); } .mini-category-card:nth-child(3n) { border-right: 1px solid var(--border); } .mini-category-card:nth-child(2n) { border-right: 0; } }
@media (max-width: 560px) { .asset-dashboard { padding: 14px; } .kpi-grid { grid-template-columns: 1fr; } .location-list { grid-template-columns: 1fr; } .category-cards { grid-template-columns: 1fr; } .mini-category-card { border-right: 0; } .category-value-row { grid-template-columns: 10px 1fr 55px; } .category-value-row strong { grid-column: 2/-1; } .header-actions .el-button { flex: 1; } .update-time { width: 100%; margin-bottom: 4px; } }
@keyframes category-value-scroll { to { transform: translateY(-50%); } }
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { animation-duration: .01ms !important; transition-duration: .01ms !important; } .category-value-list.is-scrolling { overflow-y: auto; mask-image: none; } .category-value-list.is-scrolling .category-value-track { animation: none !important; transform: none; } }
</style>
