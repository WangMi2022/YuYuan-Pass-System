<template>
  <main v-loading="loading" class="blueprint-dashboard">
    <section class="blueprint-hero" aria-labelledby="dashboard-title">
      <div class="blueprint-copy">
        <span class="coordinate">ASSET CONTROL / LIVE DATA</span>
        <h1 id="dashboard-title">{{ greeting }}，{{ userStore.userInfo.nickName || '朋友' }}</h1>
        <p>今日 {{ pendingCount }} 项业务待处理，资产整体健康度保持稳定。</p>
        <div class="hero-buttons">
          <el-button type="primary" :icon="Plus" @click="go('assetInventory')">登记资产</el-button>
          <el-button @click="go('assetDashboard')">
            查看资产大屏
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </div>
      </div>

      <aside class="health-blueprint" aria-label="资产健康度">
        <span>ASSET HEALTH / LIVE</span>
        <strong>{{ healthRate }}<small>%</small></strong>
        <div class="health-line" aria-hidden="true">
          <i :style="{ transform: `scaleX(${healthRate / 100})` }" />
        </div>
        <p><i />{{ formatNumber(controlledQuantity) }} 件资产处于正常受控状态</p>
      </aside>
    </section>

    <section class="metric-strip" aria-label="资产核心指标">
      <article v-for="(item, index) in metrics" :key="item.label">
        <span>{{ String(index + 1).padStart(2, '0') }} / {{ item.label }}</span>
        <div>
          <strong>{{ item.value }}</strong>
          <em>{{ item.hint }}</em>
        </div>
        <svg viewBox="0 0 72 26" preserveAspectRatio="none" aria-hidden="true">
          <polyline :points="item.spark" />
        </svg>
      </article>
    </section>

    <section class="lower-grid">
      <article class="sheet-panel recent-panel">
        <header>
          <div>
            <span>REGISTER / 资产台账</span>
            <h2>最近登记</h2>
          </div>
          <button type="button" class="text-button" @click="go('assetInventory')">
            查看全部 <el-icon><ArrowRight /></el-icon>
          </button>
        </header>

        <div class="recent-table-wrap">
          <table v-if="recentAssets.length" class="recent-table">
            <thead>
              <tr>
                <th>资产编号</th>
                <th>资产名称</th>
                <th>位置 / 保管人</th>
                <th>状态</th>
                <th class="number-cell">入账原值</th>
                <th><span class="sr-only">操作</span></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in recentAssets" :key="row.ID">
                <td><code>{{ row.assetCode }}</code></td>
                <td><strong>{{ row.name }}</strong></td>
                <td>
                  <span class="location-cell">{{ row.location || '位置待补充' }}</span>
                  <small>{{ row.custodian || '未指定保管人' }}</small>
                </td>
                <td>
                  <span class="asset-status" :class="`is-${statusMeta(row.status).tone}`">
                    <i />{{ statusMeta(row.status).label }}
                  </span>
                </td>
                <td class="number-cell">{{ formatCurrency(row.originalValue) }}</td>
                <td>
                  <button type="button" class="row-more" :aria-label="`查看资产 ${row.name}`" @click="go('assetInventory')">
                    <el-icon><MoreFilled /></el-icon>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="empty-state">
            <strong>暂无资产登记</strong>
            <span>完成首条资产建档后，这里会显示最新记录。</span>
          </div>
        </div>
      </article>

      <aside class="sheet-panel task-panel">
        <header>
          <div>
            <span>QUEUE / TODAY</span>
            <h2>今日待办</h2>
          </div>
          <b>{{ pendingCount > 99 ? '99+' : pendingCount }}</b>
        </header>

        <div v-if="tasks.length" class="task-list">
          <button v-for="task in tasks" :key="task.ID" type="button" class="task-line" @click="openTask(task)">
            <time>{{ taskTime(task.CreatedAt || task.businessDate) }}</time>
            <span>
              <strong>{{ taskTitle(task) }}</strong>
              <small>{{ taskDescription(task) }}</small>
            </span>
            <el-icon><ArrowRight /></el-icon>
          </button>
        </div>
        <div v-else class="task-empty">
          <span>✓</span>
          <strong>暂无待处理业务</strong>
          <small>新的草稿单据会显示在这里。</small>
        </div>
      </aside>
    </section>
  </main>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ArrowRight, MoreFilled, Plus } from '@element-plus/icons-vue'
  import { formatCompactCurrency, formatCurrency, formatNumber } from '@/utils/format'
  import { getAssetDashboard } from '@/plugin/asset/api/asset'
  import { getAssetOperationList } from '@/plugin/asset/api/operation'
  import { useUserStore } from '@/pinia/modules/user'

  defineOptions({ name: 'Dashboard' })

  const router = useRouter()
  const userStore = useUserStore()
  const loading = ref(false)

  const dashboard = ref({
    assetKinds: 0,
    totalQuantity: 0,
    categoryCount: 0,
    originalValue: 0,
    currentValue: 0,
    statusSummary: [],
    recentAssets: []
  })
  const tasks = ref([])
  const pendingCount = ref(0)

  const greeting = computed(() => {
    const hour = new Date().getHours()
    if (hour < 6) return '夜深了'
    if (hour < 11) return '早上好'
    if (hour < 14) return '中午好'
    if (hour < 18) return '下午好'
    return '晚上好'
  })

  const maintenanceQuantity = computed(() => Number(
    dashboard.value.statusSummary.find((item) => item.status === 'maintenance')?.quantity || 0
  ))
  const controlledQuantity = computed(() => Math.max(
    Number(dashboard.value.totalQuantity || 0) - maintenanceQuantity.value,
    0
  ))
  const healthRate = computed(() => {
    const total = Number(dashboard.value.totalQuantity || 0)
    return total ? ((controlledQuantity.value / total) * 100).toFixed(1) : '0.0'
  })
  const retentionRate = computed(() => {
    const original = Number(dashboard.value.originalValue || 0)
    return original ? `${((Number(dashboard.value.currentValue || 0) / original) * 100).toFixed(1)}%` : '0.0%'
  })

  const metrics = computed(() => [
    {
      label: '资产实物总量',
      value: formatNumber(dashboard.value.totalQuantity),
      hint: `${formatNumber(dashboard.value.categoryCount)} 类`,
      spark: '2,21 14,17 25,19 37,10 49,14 60,6 70,9'
    },
    {
      label: '资产档案',
      value: formatNumber(dashboard.value.assetKinds),
      hint: `${dashboard.value.recentAssets.length} 条最近登记`,
      spark: '2,20 14,15 25,18 37,8 49,12 60,4 70,7'
    },
    {
      label: '资产原值',
      value: formatCompactCurrency(dashboard.value.originalValue),
      hint: '账面原值',
      spark: '2,22 14,18 25,20 37,12 49,16 60,8 70,11'
    },
    {
      label: '当前估值',
      value: formatCompactCurrency(dashboard.value.currentValue),
      hint: retentionRate.value,
      spark: '2,21 14,16 25,18 37,9 49,13 60,5 70,8'
    }
  ])

  const recentAssets = computed(() => (dashboard.value.recentAssets || []).slice(0, 5))

  const statusMap = {
    pending_inbound: { label: '待入库', tone: 'info' },
    idle: { label: '闲置', tone: 'neutral' },
    in_use: { label: '在用', tone: 'success' },
    maintenance: { label: '维保', tone: 'warning' },
    retired: { label: '已处置', tone: 'danger' }
  }
  const statusMeta = (status) => statusMap[status] || { label: status || '未知', tone: 'neutral' }

  const operationMap = {
    inbound: { label: '资产入库', route: 'assetInbound' },
    issue: { label: '资产领用', route: 'assetIssue' },
    transfer: { label: '资产调拨', route: 'assetTransfer' },
    return: { label: '资产归还', route: 'assetReturn' },
    maintenance: { label: '维修维保', route: 'assetMaintenance' },
    scrap: { label: '资产报废', route: 'assetScrap' }
  }
  const operationMeta = (type) => operationMap[type] || { label: '资产业务', route: 'assetInventory' }

  const taskTime = (value) => {
    if (!value) return '--:--'
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return '--:--'
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
  }
  const taskTitle = (task) => task.reason?.trim() || `${operationMeta(task.type).label}待处理`
  const taskDescription = (task) => {
    const items = task.items || []
    const firstName = items[0]?.assetName || task.orderNo || '业务草稿'
    return `${firstName} · ${items.length} 项档案`
  }

  const go = (name) => {
    if (router.hasRoute(name)) router.push({ name })
  }
  const openTask = (task) => go(operationMeta(task.type).route)

  const loadDashboard = async () => {
    loading.value = true
    try {
      const [dashboardResult, taskResult] = await Promise.allSettled([
        getAssetDashboard(),
        getAssetOperationList({ page: 1, pageSize: 4, status: 'draft' })
      ])

      if (dashboardResult.status === 'fulfilled' && dashboardResult.value.code === 0) {
        const data = dashboardResult.value.data || {}
        dashboard.value = {
          ...dashboard.value,
          ...data,
          statusSummary: data.statusSummary || [],
          recentAssets: data.recentAssets || []
        }
      }
      if (taskResult.status === 'fulfilled' && taskResult.value.code === 0) {
        tasks.value = taskResult.value.data?.list || []
        pendingCount.value = Number(taskResult.value.data?.total || 0)
      }
    } finally {
      loading.value = false
    }
  }

  onMounted(loadDashboard)
</script>

<style scoped lang="scss">
  .blueprint-dashboard {
    --blueprint-accent: var(--na-primary);
    --blueprint-ink: color-mix(in srgb, var(--na-foreground) 92%, #0f2d5c);
    --blueprint-muted: color-mix(in srgb, var(--na-muted-foreground) 88%, #486998);
    width: 100%;
    min-height: 100%;
    padding: 22px 24px 28px;
    background: var(--na-background);
    color: var(--blueprint-ink);
  }

  .blueprint-hero {
    position: relative;
    display: grid;
    min-height: 176px;
    overflow: hidden;
    grid-template-columns: minmax(0, 1fr) 360px;
    align-items: center;
    gap: 30px;
    padding: 26px 28px;
    border: 1px solid color-mix(in srgb, var(--blueprint-accent) 22%, var(--na-border));
    background-color: color-mix(in srgb, var(--na-card) 94%, var(--blueprint-accent));
    background-image:
      linear-gradient(color-mix(in srgb, var(--blueprint-accent) 9%, transparent) 1px, transparent 1px),
      linear-gradient(90deg, color-mix(in srgb, var(--blueprint-accent) 9%, transparent) 1px, transparent 1px),
      linear-gradient(color-mix(in srgb, var(--blueprint-accent) 4%, transparent) 1px, transparent 1px),
      linear-gradient(90deg, color-mix(in srgb, var(--blueprint-accent) 4%, transparent) 1px, transparent 1px);
    background-size: 40px 40px, 40px 40px, 8px 8px, 8px 8px;
  }
  .blueprint-hero::before {
    position: absolute;
    top: 10px;
    right: 12px;
    color: var(--blueprint-accent);
    content: 'A-01';
    font: 600 8px/1 Bahnschrift, 'Segoe UI', sans-serif;
    letter-spacing: .14em;
  }
  .coordinate {
    color: var(--blueprint-muted);
    font: 600 8px/1.4 Bahnschrift, 'Segoe UI', sans-serif;
    letter-spacing: .15em;
  }
  .blueprint-copy h1 {
    margin: 8px 0 5px;
    color: var(--blueprint-ink);
    font-size: clamp(24px, 2.2vw, 32px);
    font-weight: 650;
    letter-spacing: -.04em;
  }
  .blueprint-copy p { margin: 0; color: var(--blueprint-muted); font-size: 12px; }
  .hero-buttons { display: flex; gap: 8px; margin-top: 22px; }
  .hero-buttons :deep(.el-button) { min-height: 36px; }

  .health-blueprint {
    position: relative;
    padding: 17px 20px;
    border: 1px solid color-mix(in srgb, var(--blueprint-accent) 48%, var(--na-border));
    background: color-mix(in srgb, var(--na-card) 94%, transparent);
    box-shadow: 6px 6px 0 color-mix(in srgb, var(--blueprint-accent) 5%, transparent);
  }
  .health-blueprint::before,
  .health-blueprint::after {
    position: absolute;
    width: 7px;
    height: 7px;
    border-color: var(--blueprint-accent);
    border-style: solid;
    content: '';
  }
  .health-blueprint::before { top: -1px; left: -1px; border-width: 1px 0 0 1px; }
  .health-blueprint::after { right: -1px; bottom: -1px; border-width: 0 1px 1px 0; }
  .health-blueprint > span {
    color: var(--blueprint-muted);
    font: 600 8px/1.4 Bahnschrift, 'Segoe UI', sans-serif;
    letter-spacing: .15em;
  }
  .health-blueprint > strong {
    display: block;
    margin: 10px 0 4px;
    color: color-mix(in srgb, var(--blueprint-accent) 58%, var(--blueprint-ink));
    font: 42px/1 Bahnschrift, 'Segoe UI', sans-serif;
  }
  .health-blueprint > strong small { margin-left: 2px; font-size: 15px; }
  .health-line { height: 4px; background: color-mix(in srgb, var(--blueprint-accent) 14%, var(--na-muted)); }
  .health-line i { display: block; width: 100%; height: 100%; transform-origin: left; background: var(--blueprint-accent); transition: transform 300ms ease-out; }
  .health-blueprint p { display: flex; align-items: center; gap: 8px; margin: 13px 0 0; color: var(--blueprint-muted); font-size: 9px; }
  .health-blueprint p i { width: 7px; height: 7px; border-radius: 50%; background: var(--na-success); box-shadow: 0 0 0 4px var(--na-success-soft); }

  .metric-strip {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    margin-top: 12px;
    border: 1px solid var(--na-border);
    background: var(--na-card);
  }
  .metric-strip article {
    position: relative;
    min-width: 0;
    padding: 15px 16px 10px;
    border-right: 1px solid var(--na-border);
    color: var(--blueprint-accent);
  }
  .metric-strip article:last-child { border-right: 0; }
  .metric-strip article > span {
    color: var(--blueprint-muted);
    font: 600 8px/1.4 Bahnschrift, 'Segoe UI', sans-serif;
    letter-spacing: .08em;
  }
  .metric-strip article > div { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; margin-top: 6px; }
  .metric-strip strong { color: var(--blueprint-ink); font: 24px/1.2 Bahnschrift, 'Segoe UI', sans-serif; }
  .metric-strip em { overflow: hidden; color: var(--na-success); font-size: 9px; font-style: normal; text-overflow: ellipsis; white-space: nowrap; }
  .metric-strip svg { position: absolute; right: 15px; bottom: 8px; width: 72px; height: 26px; opacity: .18; }
  .metric-strip polyline { fill: none; stroke: var(--blueprint-accent); stroke-width: 2; vector-effect: non-scaling-stroke; }

  .lower-grid { display: grid; grid-template-columns: minmax(0, 1fr) 300px; gap: 12px; margin-top: 12px; }
  .sheet-panel { min-width: 0; overflow: hidden; border: 1px solid var(--na-border); background: var(--na-card); }
  .sheet-panel > header {
    display: flex;
    height: 58px;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 0 15px;
    border-bottom: 1px solid var(--na-border);
  }
  .sheet-panel header span {
    color: var(--blueprint-muted);
    font: 600 7px/1.4 Bahnschrift, 'Segoe UI', sans-serif;
    letter-spacing: .13em;
  }
  .sheet-panel h2 { margin: 3px 0 0; color: var(--blueprint-ink); font-size: 13px; font-weight: 650; }
  .text-button { display: flex; align-items: center; gap: 6px; border: 0; background: transparent; color: color-mix(in srgb, var(--blueprint-accent) 65%, var(--blueprint-ink)); font-size: 9px; }
  .text-button .el-icon { width: 13px; }

  .recent-table-wrap { width: 100%; overflow-x: auto; }
  .recent-table { width: 100%; min-width: 860px; border-collapse: collapse; table-layout: fixed; }
  .recent-table th {
    height: 39px;
    padding: 0 12px;
    border-bottom: 1px solid var(--na-border);
    color: var(--blueprint-muted);
    font-size: 9px;
    font-weight: 550;
    text-align: left;
  }
  .recent-table th:nth-child(1) { width: 18%; }
  .recent-table th:nth-child(2) { width: 29%; }
  .recent-table th:nth-child(3) { width: 23%; }
  .recent-table th:nth-child(4) { width: 14%; }
  .recent-table th:nth-child(5) { width: 13%; }
  .recent-table th:last-child { width: 36px; }
  .recent-table td { height: 54px; padding: 0 12px; border-bottom: 1px solid var(--na-border); color: var(--blueprint-muted); font-size: 11px; }
  .recent-table tbody tr:last-child td { border-bottom: 0; }
  .recent-table tbody tr:hover { background: color-mix(in srgb, var(--blueprint-accent) 3%, var(--na-card)); }
  .recent-table code { color: color-mix(in srgb, var(--blueprint-accent) 72%, var(--blueprint-ink)); font: 9px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; }
  .recent-table td > strong { color: var(--blueprint-ink); font-size: 11px; font-weight: 550; }
  .recent-table td:nth-child(3) { line-height: 1.3; }
  .location-cell,
  .recent-table td:nth-child(3) small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .recent-table td:nth-child(3) small { margin-top: 3px; color: var(--na-muted-foreground); font-size: 8px; }
  .number-cell { text-align: right !important; font-variant-numeric: tabular-nums; }
  .asset-status { display: inline-flex; align-items: center; gap: 7px; white-space: nowrap; }
  .asset-status i { width: 6px; height: 6px; border-radius: 50%; background: var(--na-muted-foreground); }
  .asset-status.is-success { color: var(--na-success); }
  .asset-status.is-success i { background: var(--na-success); }
  .asset-status.is-warning { color: var(--na-warning); }
  .asset-status.is-warning i { background: var(--na-warning); }
  .asset-status.is-danger { color: var(--na-danger); }
  .asset-status.is-danger i { background: var(--na-danger); }
  .asset-status.is-info { color: var(--na-info); }
  .asset-status.is-info i { background: var(--na-info); }
  .row-more { display: grid; width: 28px; height: 28px; place-items: center; border: 0; border-radius: 5px; background: transparent; color: var(--blueprint-muted); }
  .row-more:hover { background: var(--na-muted); color: var(--blueprint-ink); }

  .task-panel header b { display: grid; width: 28px; height: 28px; place-items: center; border-radius: 50%; background: var(--na-primary-soft); color: var(--blueprint-accent); font: 11px/1 Bahnschrift, 'Segoe UI', sans-serif; }
  .task-line {
    display: grid;
    width: 100%;
    padding: 13px 14px;
    grid-template-columns: 42px minmax(0, 1fr) auto;
    align-items: center;
    gap: 9px;
    border: 0;
    border-bottom: 1px solid var(--na-border);
    background: var(--na-card);
    color: var(--blueprint-ink);
    text-align: left;
  }
  .task-line:hover { background: color-mix(in srgb, var(--blueprint-accent) 3%, var(--na-card)); }
  .task-line time { color: var(--blueprint-accent); font: 10px/1 Bahnschrift, 'Segoe UI', sans-serif; }
  .task-line > span { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
  .task-line strong { overflow: hidden; font-size: 10px; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
  .task-line small { overflow: hidden; color: var(--blueprint-muted); font-size: 8px; text-overflow: ellipsis; white-space: nowrap; }
  .task-line .el-icon { width: 12px; color: var(--na-muted-foreground); }
  .task-empty,
  .empty-state { display: flex; min-height: 210px; align-items: center; justify-content: center; flex-direction: column; gap: 6px; color: var(--blueprint-muted); text-align: center; }
  .task-empty span { display: grid; width: 32px; height: 32px; margin-bottom: 3px; place-items: center; border-radius: 50%; background: var(--na-success-soft); color: var(--na-success); }
  .task-empty strong,
  .empty-state strong { color: var(--blueprint-ink); font-size: 11px; }
  .task-empty small,
  .empty-state span { font-size: 9px; }
  .empty-state { min-height: 310px; }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
  }

  @media (max-width: 1120px) {
    .blueprint-hero { grid-template-columns: minmax(0, 1fr) 300px; }
    .lower-grid { grid-template-columns: 1fr; }
    .task-list { display: grid; grid-template-columns: 1fr 1fr; }
    .task-line:nth-last-child(-n + 2) { border-bottom: 0; }
  }

  @media (max-width: 860px) {
    .blueprint-dashboard { padding: 15px; }
    .blueprint-hero { grid-template-columns: 1fr; }
    .health-blueprint { display: none; }
    .metric-strip { grid-template-columns: 1fr 1fr; }
    .metric-strip article:nth-child(2) { border-right: 0; }
    .metric-strip article:nth-child(-n + 2) { border-bottom: 1px solid var(--na-border); }
  }

  @media (max-width: 560px) {
    .blueprint-hero { min-height: 0; padding: 24px 18px; }
    .blueprint-copy h1 { font-size: 23px; }
    .hero-buttons { align-items: stretch; flex-direction: column; }
    .hero-buttons :deep(.el-button) { width: 100%; margin-left: 0; }
    .metric-strip { grid-template-columns: 1fr; }
    .metric-strip article { border-right: 0; border-bottom: 1px solid var(--na-border); }
    .metric-strip article:last-child { border-bottom: 0; }
    .recent-table { min-width: 0; }
    .recent-table th:nth-child(1) { width: 42%; }
    .recent-table th:nth-child(2) { width: auto; }
    .recent-table th:nth-child(n + 3):nth-child(-n + 5),
    .recent-table td:nth-child(n + 3):nth-child(-n + 5) { display: none; }
    .recent-table code,
    .recent-table td > strong { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .task-list { grid-template-columns: 1fr; }
    .task-line { border-bottom: 1px solid var(--na-border) !important; }
    .task-line:last-child { border-bottom: 0 !important; }
  }

  @media (prefers-reduced-motion: reduce) {
    .health-line i { transition: none; }
  }
</style>
