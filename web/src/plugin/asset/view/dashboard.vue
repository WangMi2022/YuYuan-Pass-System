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
              <el-progress :percentage="locationPercent(item.quantity)" :show-text="false" :stroke-width="7" color="#059669" />
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
  in_use: { label: '使用中', color: '#059669' },
  idle: { label: '闲置', color: '#64748b' },
  maintenance: { label: '维修中', color: '#d97706' },
  retired: { label: '已处置', color: '#dc2626' }
}
const activeCategories = computed(() => dashboard.value.categorySummary.filter((item) => Number(item.quantity) > 0))
const hasCategoryData = computed(() => activeCategories.value.length > 0)
const categoryListShouldScroll = computed(() => activeCategories.value.length > 4)
const categoryScrollGroups = computed(() => categoryListShouldScroll.value
  ? [activeCategories.value, activeCategories.value]
  : [activeCategories.value])
const categoryScrollDuration = computed(() => Math.max(12, activeCategories.value.length * 2.4))
const statusRows = computed(() => dashboard.value.statusSummary.map((item) => ({ ...item, ...(statusConfig[item.status] || { label: item.status, color: '#64748b' }) })))
const hasStatusData = computed(() => statusRows.value.some((item) => item.quantity > 0))
const chartText = computed(() => appStore.isDark ? '#cbd5e1' : '#475569')
const chartGrid = computed(() => appStore.isDark ? '#263244' : '#e2e8f0')

const categoryOption = computed(() => ({
  animationDuration: 450,
  grid: { top: 18, right: 26, bottom: 55, left: 62, containLabel: true },
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' }, valueFormatter: (value) => `${Number(value).toFixed(2)} 万元` },
  xAxis: { type: 'category', data: activeCategories.value.map((item) => item.categoryName), axisLabel: { color: chartText.value, rotate: activeCategories.value.length > 5 ? 24 : 0, interval: 0 }, axisLine: { lineStyle: { color: chartGrid.value } }, axisTick: { show: false } },
  yAxis: { type: 'value', name: '万元', nameTextStyle: { color: chartText.value }, axisLabel: { color: chartText.value }, splitLine: { lineStyle: { color: chartGrid.value, type: 'dashed' } } },
  series: [{
    name: '当前估值', type: 'bar', barMaxWidth: 42,
    data: activeCategories.value.map((item) => ({ value: Number(item.currentValue || 0) / 10000, itemStyle: { color: item.color, borderRadius: [6, 6, 0, 0] } })),
    label: { show: true, position: 'top', color: chartText.value, formatter: ({ value }) => value > 0 ? value.toFixed(1) : '' }
  }]
}))

const statusOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: ({ name, value, percent: p }) => `${name}<br/>${value} 件（${p}%）` },
  legend: { show: false },
  series: [{
    type: 'pie', radius: ['58%', '78%'], center: ['50%', '48%'], avoidLabelOverlap: true,
    itemStyle: { borderColor: appStore.isDark ? '#111827' : '#ffffff', borderWidth: 4, borderRadius: 5 },
    label: { show: false },
    data: statusRows.value.map((item) => ({ name: item.label, value: item.quantity, itemStyle: { color: item.color } }))
  }],
  graphic: [{ type: 'text', left: 'center', top: '38%', style: { text: String(dashboard.value.totalQuantity), fill: appStore.isDark ? '#f8fafc' : '#0f172a', fontSize: 28, fontWeight: 700 } }, { type: 'text', left: 'center', top: '54%', style: { text: '资产总量', fill: chartText.value, fontSize: 12 } }]
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
  --bg:var(--na-background); --surface:var(--na-card); --surface-2:var(--na-muted); --text:var(--na-foreground); --muted:var(--na-muted-foreground); --border:var(--na-border);
  position:relative; isolation:isolate; min-height:100%; padding:24px; background:var(--bg); color:var(--text); overflow:auto;
}
.asset-dashboard::before { position:absolute; z-index:0; inset:0; content:""; background-image:linear-gradient(rgb(37 99 235 / 3.5%) 1px,transparent 1px),linear-gradient(90deg,rgb(5 150 105 / 3%) 1px,transparent 1px); background-size:40px 40px; pointer-events:none; }
.asset-dashboard>* { position:relative; z-index:1; }
.asset-dashboard:fullscreen { min-height:100vh; }
.dashboard-header { display:flex; align-items:flex-end; justify-content:space-between; gap:24px; margin-bottom:14px; }
.eyebrow { margin:0 0 5px; color:var(--na-primary); font:600 12px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace; letter-spacing:.14em; }
h1 { margin:0; font-size:clamp(27px,3vw,38px); letter-spacing:-.025em; }
.dashboard-header p:last-child { margin:8px 0 0; color:var(--muted); }
.header-actions { display:flex; align-items:center; gap:10px; flex-wrap:wrap; justify-content:flex-end; }
.update-time { display:inline-flex; align-items:center; gap:7px; margin-right:5px; color:var(--muted); font-size:12px; }
.update-time i { width:7px; height:7px; border-radius:50%; background:#10b981; box-shadow:0 0 0 0 rgb(16 185 129 / 24%); animation:data-status-pulse 2.8s ease-out infinite; }
.dashboard-flow-rail { position:relative; height:3px; margin-bottom:17px; overflow:hidden; border-radius:999px; background:linear-gradient(90deg,rgb(37 99 235 / 8%),rgb(14 165 233 / 28%) 34%,rgb(5 150 105 / 24%) 68%,rgb(217 119 6 / 8%)); }
.dashboard-flow-rail::after { position:absolute; inset:0 auto 0 -34%; width:34%; content:""; background:linear-gradient(90deg,transparent,rgb(56 189 248 / 20%),rgb(37 99 235 / 90%),rgb(16 185 129 / 76%),transparent); filter:drop-shadow(0 0 5px rgb(56 189 248 / 58%)); animation:dashboard-flow 5.8s linear infinite; }
.dashboard-flow-rail span { position:absolute; z-index:1; top:50%; right:0; width:6px; height:6px; border:1px solid rgb(255 255 255 / 76%); border-radius:50%; background:#10b981; box-shadow:0 0 9px rgb(16 185 129 / 68%); transform:translateY(-50%); }
.kpi-grid { display:grid; grid-template-columns:repeat(5,minmax(0,1fr)); gap:14px; margin-bottom:14px; }
.kpi-card { --kpi-flow:#64748b; position:relative; display:flex; align-items:center; gap:14px; min-height:124px; overflow:hidden; padding:18px; border:1px solid color-mix(in srgb,var(--border) 78%,var(--kpi-flow) 22%); border-radius:var(--na-radius); background:var(--surface); box-shadow:var(--na-shadow-sm),inset 0 1px 0 rgb(255 255 255 / 42%); }
.kpi-card::before { position:absolute; top:0; left:0; width:42%; height:2px; content:""; background:linear-gradient(90deg,var(--kpi-flow),transparent); opacity:.72; }
.kpi-card.violet{--kpi-flow:#7c3aed}.kpi-card.blue{--kpi-flow:#2563eb}.kpi-card.green{--kpi-flow:#059669}.kpi-card.amber{--kpi-flow:#d97706}
.kpi-icon { display:grid; place-items:center; width:46px; height:46px; flex:0 0 46px; border-radius:12px; font-size:22px; }
.kpi-icon.slate{color:#334155;background:#e2e8f0}.kpi-icon.violet{color:#7c3aed;background:#ede9fe}.kpi-icon.blue{color:#2563eb;background:#dbeafe}.kpi-icon.green{color:#047857;background:#d1fae5}.kpi-icon.amber{color:#b45309;background:#fef3c7}
.kpi-card>div:last-child { display:flex; flex-direction:column; min-width:0; }
.kpi-card span { color:var(--muted); font-size:13px; }
.kpi-card strong { margin:4px 0 3px; overflow:hidden; color:var(--text); font-size:clamp(21px,2vw,28px); font-variant-numeric:tabular-nums; text-overflow:ellipsis; white-space:nowrap; }
.kpi-card small { color:var(--muted); font-size:11px; }
.dashboard-grid { display:grid; grid-template-columns:minmax(0,1.7fr) minmax(290px,.75fr) minmax(320px,.85fr); gap:14px; }
.dashboard-card { position:relative; overflow:hidden; border:1px solid color-mix(in srgb,var(--border) 84%,#3b82f6 16%); border-radius:var(--na-radius); background:var(--surface); box-shadow:var(--na-shadow-sm),inset 0 1px 0 rgb(255 255 255 / 38%); }
.dashboard-card::before { position:absolute; z-index:2; top:0; left:20px; width:110px; height:2px; content:""; background:linear-gradient(90deg,#2563eb,#0ea5e9 48%,#059669 78%,transparent); opacity:.62; pointer-events:none; }
.card-header { display:flex; align-items:center; justify-content:space-between; gap:16px; padding:18px 20px 0; }
.card-header h2 { margin:0; font-size:16px; }
.card-header p { margin:4px 0 0; color:var(--muted); font-size:12px; }
.chart-wrap { position:relative; padding:2px 8px 0; }
.category-chart-card .chart-wrap { overflow:hidden; }
.category-chart-card .chart-wrap::after { position:absolute; z-index:2; top:18px; bottom:22px; left:-24%; width:24%; content:""; background:linear-gradient(90deg,transparent,rgb(56 189 248 / 4%),rgb(37 99 235 / 12%),rgb(16 185 129 / 7%),transparent); border-right:1px solid rgb(56 189 248 / 20%); filter:drop-shadow(4px 0 7px rgb(37 99 235 / 10%)); pointer-events:none; animation:chart-scan 8.8s ease-in-out infinite; }
.category-value-list { max-height:164px; overflow:hidden; margin:0 20px 16px; }
.category-value-list.is-scrolling { height:164px; mask-image:linear-gradient(to bottom,transparent 0,#000 7px,#000 calc(100% - 7px),transparent 100%); }
.category-value-track { will-change:transform; }
.category-value-list.is-scrolling .category-value-track { animation:category-value-scroll var(--category-scroll-duration) linear infinite; }
.category-value-list.is-scrolling:hover .category-value-track { animation-play-state:paused; }
.category-value-group { width:100%; }
.category-value-row { display:grid; grid-template-columns:10px minmax(100px,1fr) 70px 130px; align-items:center; gap:8px; min-height:36px; border-top:1px solid var(--border); font-size:12px; }
.category-value-group:first-child .category-value-row:first-child { border-top:0; }
.legend-dot { width:8px; height:8px; border-radius:50%; }
.category-label { color:var(--text); font-weight:600; }
.category-value-row>span:nth-child(3) { color:var(--muted); text-align:right; }
.category-value-row strong { text-align:right; font-variant-numeric:tabular-nums; }
.status-list { display:grid; grid-template-columns:1fr 1fr; gap:8px 12px; padding:0 20px 20px; }
.status-list>div { display:grid; grid-template-columns:1fr auto; gap:3px 8px; padding:9px 10px; border-radius:8px; background:var(--surface-2); }
.status-list span { display:flex; align-items:center; gap:6px; color:var(--muted); font-size:12px; }.status-list span i{width:7px;height:7px;border-radius:50%}
.status-list strong { font-variant-numeric:tabular-nums; }.status-list small{grid-column:2;color:var(--muted);font-size:10px;text-align:right}
.location-list { padding:14px 20px 20px; }
.location-row { display:flex; align-items:center; gap:11px; min-height:52px; border-bottom:1px solid var(--border); }.location-row:last-child{border-bottom:0}
.rank { width:25px; color:#94a3b8; font:600 12px/1 ui-monospace,SFMono-Regular,Menlo,monospace; }.rank.top{color:#047857}
.location-progress { flex:1; min-width:0; }.location-progress>div{display:flex;justify-content:space-between;gap:8px;margin-bottom:6px}.location-progress strong{overflow:hidden;font-size:12px;text-overflow:ellipsis;white-space:nowrap}.location-progress span{color:var(--muted);font-size:10px}
.location-quantity { min-width:54px; text-align:right; font-size:12px; font-variant-numeric:tabular-nums; }
.bottom-grid { display:grid; grid-template-columns:1.15fr 1fr; gap:14px; margin-top:14px; }
.category-cards { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:10px; padding:16px 20px 20px; }
.mini-category-card { position:relative; display:grid; grid-template-columns:1fr auto; gap:8px; overflow:hidden; min-height:104px; padding:14px; border:1px solid var(--border); border-radius:11px; background:var(--surface-2); }
.mini-accent { position:absolute; inset:0 auto 0 0; width:4px; }.mini-title{display:flex;flex-direction:column;gap:3px;min-width:0}.mini-title strong{overflow:hidden;font-size:13px;text-overflow:ellipsis;white-space:nowrap}.mini-title span{color:var(--muted);font-size:10px}.mini-quantity{align-self:end;font-size:22px;font-variant-numeric:tabular-nums}.mini-quantity small{color:var(--muted);font-size:11px;font-weight:400}.mini-value{align-self:end;color:var(--na-primary);font-size:12px;font-weight:600;text-align:right}
.recent-table { margin-top:10px; }.asset-cell{display:flex;flex-direction:column;gap:3px}.asset-cell span{color:var(--muted);font:10px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace}.table-money{color:var(--na-primary);font-variant-numeric:tabular-nums}
:global(html.dark) .kpi-icon.slate{color:#cbd5e1;background:#263244}:global(html.dark) .kpi-icon.violet{color:#c4b5fd;background:#2e1d50}:global(html.dark) .kpi-icon.blue{color:#93c5fd;background:#172554}:global(html.dark) .kpi-icon.green{color:#6ee7b7;background:#064e3b}:global(html.dark) .kpi-icon.amber{color:#fcd34d;background:#451a03}
:global(html.dark) .asset-dashboard::before{background-image:linear-gradient(rgb(96 165 250 / 5%) 1px,transparent 1px),linear-gradient(90deg,rgb(52 211 153 / 4%) 1px,transparent 1px)}
:global(html.dark) .kpi-card,:global(html.dark) .dashboard-card{box-shadow:var(--na-shadow-sm),inset 0 1px 0 rgb(255 255 255 / 5%)}
@media (max-width:1400px){.kpi-grid{grid-template-columns:repeat(3,1fr)}.dashboard-grid{grid-template-columns:1.5fr 1fr}.location-card{grid-column:1/-1}.location-list{display:grid;grid-template-columns:1fr 1fr;column-gap:24px}.category-cards{grid-template-columns:repeat(3,1fr)}}
@media (max-width:900px){.dashboard-header{align-items:stretch;flex-direction:column}.header-actions{justify-content:flex-start}.kpi-grid{grid-template-columns:repeat(2,1fr)}.dashboard-grid,.bottom-grid{grid-template-columns:1fr}.location-card{grid-column:auto}.category-cards{grid-template-columns:repeat(2,1fr)}}
@media (max-width:560px){.asset-dashboard{padding:14px}.kpi-grid{grid-template-columns:1fr}.location-list{grid-template-columns:1fr}.category-cards{grid-template-columns:1fr}.category-value-row{grid-template-columns:10px 1fr 55px}.category-value-row strong{grid-column:2/-1}.header-actions .el-button{flex:1}.update-time{width:100%;margin-bottom:4px}}
@keyframes category-value-scroll { to { transform:translateY(-50%); } }
@keyframes dashboard-flow { to { left:100%; } }
@keyframes chart-scan { 0%,12%{left:-24%;opacity:0} 20%{opacity:1} 62%{opacity:.8} 72%,100%{left:106%;opacity:0} }
@keyframes data-status-pulse { 0%{box-shadow:0 0 0 0 rgb(16 185 129 / 30%)} 70%,100%{box-shadow:0 0 0 7px rgb(16 185 129 / 0%)} }
@media (prefers-reduced-motion:reduce){*,*::before,*::after{animation-duration:.01ms!important;transition-duration:.01ms!important}.dashboard-flow-rail::after,.category-chart-card .chart-wrap::after,.update-time i{animation:none!important}.category-value-list.is-scrolling{overflow-y:auto;mask-image:none}.category-value-list.is-scrolling .category-value-track{animation:none!important;transform:none}}
</style>
