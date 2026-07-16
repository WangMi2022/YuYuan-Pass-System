<template>
  <main class="monitor-page" aria-labelledby="server-monitor-title">
    <header class="monitor-header">
      <div>
        <p class="eyebrow">SERVER MONITORING</p>
        <h1 id="server-monitor-title">服务器负载</h1>
        <p class="subtitle">持续采集 CPU、内存、系统负载与磁盘使用情况，保留当前页面最近 30 次采样。</p>
      </div>
      <div class="monitor-actions">
        <span class="collector-state" :class="{ paused }" aria-live="polite">
          <i />{{ paused ? '采集已暂停' : '实时采集中' }}
        </span>
        <el-select v-model="refreshSeconds" class="interval-select" aria-label="采集周期">
          <el-option label="每 5 秒" :value="5" />
          <el-option label="每 10 秒" :value="10" />
          <el-option label="每 30 秒" :value="30" />
          <el-option label="每 60 秒" :value="60" />
        </el-select>
        <el-button :icon="paused ? VideoPlay : VideoPause" @click="togglePause">
          {{ paused ? '继续采集' : '暂停采集' }}
        </el-button>
        <el-button type="primary" :icon="Refresh" :loading="loading" @click="reload(true)">立即采集</el-button>
      </div>
    </header>

    <el-alert
      v-if="errorMessage"
      class="monitor-alert"
      type="error"
      title="服务器负载采集失败"
      :description="errorMessage"
      show-icon
      :closable="false"
    />

    <section class="kpi-grid" aria-label="服务器核心负载指标">
      <article class="metric-card">
        <span class="metric-icon cpu"><el-icon><Cpu /></el-icon></span>
        <div><span>CPU 平均使用率</span><strong>{{ percent(cpuUsage) }}</strong><small>{{ state.cpu?.cores || 0 }} 个物理核心</small></div>
        <el-progress :percentage="cpuUsage" :show-text="false" :stroke-width="5" :color="progressColors" />
      </article>
      <article class="metric-card">
        <span class="metric-icon memory"><el-icon><Coin /></el-icon></span>
        <div><span>内存使用率</span><strong>{{ percent(ramUsage) }}</strong><small>{{ formatCapacity(state.ram?.usedMb) }} / {{ formatCapacity(state.ram?.totalMb) }}</small></div>
        <el-progress :percentage="ramUsage" :show-text="false" :stroke-width="5" :color="progressColors" />
      </article>
      <article class="metric-card">
        <span class="metric-icon disk"><el-icon><FolderOpened /></el-icon></span>
        <div><span>磁盘最高使用率</span><strong>{{ percent(diskUsage) }}</strong><small>{{ diskCount }} 个挂载点正在监控</small></div>
        <el-progress :percentage="diskUsage" :show-text="false" :stroke-width="5" :color="progressColors" />
      </article>
      <article class="metric-card">
        <span class="metric-icon load"><el-icon><DataLine /></el-icon></span>
        <div><span>系统 1 分钟负载</span><strong>{{ number(state.load?.load1) }}</strong><small>5 分钟 {{ number(state.load?.load5) }} · 15 分钟 {{ number(state.load?.load15) }}</small></div>
        <el-progress :percentage="normalizedLoad" :show-text="false" :stroke-width="5" :color="progressColors" />
      </article>
    </section>

    <section class="primary-grid">
      <article class="monitor-card trend-card">
        <header class="card-header">
          <div><h2>负载采集趋势</h2><p>CPU、内存和磁盘最高使用率</p></div>
          <span>{{ history.length }} / 30 个采样点</span>
        </header>
        <Chart v-if="history.length" :options="trendOption" height="330px" />
        <el-empty v-else description="等待首次采样" :image-size="72" />
      </article>

      <article class="monitor-card runtime-card">
        <header class="card-header">
          <div><h2>运行环境</h2><p>主机与应用运行时信息</p></div>
        </header>
        <dl class="runtime-list">
          <div><dt>主机名称</dt><dd>{{ state.host?.hostname || '—' }}</dd></div>
          <div><dt>操作系统</dt><dd>{{ platformText }}</dd></div>
          <div><dt>内核版本</dt><dd>{{ state.host?.kernelVersion || '—' }}</dd></div>
          <div><dt>运行时</dt><dd>{{ state.os?.goVersion || '—' }}</dd></div>
          <div><dt>编译器</dt><dd>{{ state.os?.compiler || '—' }}</dd></div>
          <div><dt>逻辑 CPU</dt><dd>{{ state.os?.numCpu || 0 }} 核</dd></div>
          <div><dt>Goroutine</dt><dd>{{ state.os?.numGoroutine || 0 }}</dd></div>
          <div><dt>主机运行时长</dt><dd>{{ uptimeText }}</dd></div>
        </dl>
        <footer class="sample-time">最近采集：{{ collectedAtText }}</footer>
      </article>
    </section>

    <section class="detail-grid">
      <article class="monitor-card">
        <header class="card-header">
          <div><h2>CPU 核心负载</h2><p>各逻辑核心当前使用率</p></div>
          <span>{{ cpuCores.length }} 个逻辑核心</span>
        </header>
        <div v-if="cpuCores.length" class="core-grid">
          <div v-for="(usage, index) in cpuCores" :key="index" class="core-row">
            <span>CPU {{ String(index + 1).padStart(2, '0') }}</span>
            <el-progress :percentage="safePercent(usage)" :stroke-width="7" :color="progressColors" />
          </div>
        </div>
        <el-empty v-else description="暂无 CPU 数据" :image-size="64" />
      </article>

      <article class="monitor-card">
        <header class="card-header">
          <div><h2>磁盘使用情况</h2><p>配置挂载点的容量与占用</p></div>
          <span>{{ diskCount }} 个挂载点</span>
        </header>
        <div v-if="disks.length" class="disk-list">
          <div v-for="disk in disks" :key="disk.mountPoint" class="disk-row">
            <div class="disk-title"><strong>{{ disk.mountPoint }}</strong><span>{{ disk.usedGb }} GB / {{ disk.totalGb }} GB</span></div>
            <el-progress :percentage="safePercent(disk.usedPercent)" :stroke-width="8" :color="progressColors" />
          </div>
        </div>
        <el-empty v-else description="暂无磁盘数据" :image-size="64" />
      </article>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { Coin, Cpu, DataLine, FolderOpened, Refresh, VideoPause, VideoPlay } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import Chart from '@/components/charts/index.vue'
import { getSystemState } from '@/api/system'
import { useAppStore } from '@/pinia'

defineOptions({ name: 'State' })

const appStore = useAppStore()
const state = ref({})
const history = ref([])
const loading = ref(false)
const paused = ref(false)
const refreshSeconds = ref(10)
const errorMessage = ref('')
let timer = null

const progressColors = [
  { color: '#059669', percentage: 60 },
  { color: '#d97706', percentage: 80 },
  { color: '#dc2626', percentage: 100 }
]
const safePercent = (value) => Math.min(100, Math.max(0, Math.round(Number(value || 0))))
const average = (values = []) => values.length ? values.reduce((sum, value) => sum + Number(value || 0), 0) / values.length : 0
const number = (value) => Number(value || 0).toFixed(2)
const percent = (value) => `${safePercent(value)}%`
const cpuCores = computed(() => state.value.cpu?.cpus || [])
const cpuUsage = computed(() => safePercent(average(cpuCores.value)))
const ramUsage = computed(() => safePercent(state.value.ram?.usedPercent))
const disks = computed(() => state.value.disk || [])
const diskCount = computed(() => disks.value.length)
const diskUsage = computed(() => safePercent(Math.max(0, ...disks.value.map((item) => Number(item.usedPercent || 0)))))
const normalizedLoad = computed(() => safePercent((Number(state.value.load?.load1 || 0) / Math.max(Number(state.value.os?.numCpu || 1), 1)) * 100))
const platformText = computed(() => [state.value.host?.platform, state.value.host?.platformVersion].filter(Boolean).join(' ') || state.value.os?.goos || '—')
const collectedAtText = computed(() => state.value.collectedAt ? new Date(state.value.collectedAt).toLocaleString('zh-CN', { hour12: false }) : '—')
const uptimeText = computed(() => {
  let seconds = Number(state.value.host?.uptime || 0)
  const days = Math.floor(seconds / 86400); seconds %= 86400
  const hours = Math.floor(seconds / 3600); seconds %= 3600
  const minutes = Math.floor(seconds / 60)
  return `${days} 天 ${hours} 小时 ${minutes} 分钟`
})
const formatCapacity = (mb) => {
  const value = Number(mb || 0)
  return value >= 1024 ? `${(value / 1024).toFixed(1)} GB` : `${value} MB`
}

const chartText = computed(() => appStore.isDark ? '#cbd5e1' : '#475569')
const chartGrid = computed(() => appStore.isDark ? '#263244' : '#e2e8f0')
const trendOption = computed(() => ({
  animationDuration: 240,
  grid: { top: 32, right: 24, bottom: 32, left: 48, containLabel: true },
  tooltip: { trigger: 'axis', valueFormatter: (value) => `${Number(value).toFixed(0)}%` },
  legend: { top: 2, right: 18, textStyle: { color: chartText.value }, data: ['CPU', '内存', '磁盘'] },
  xAxis: { type: 'category', boundaryGap: false, data: history.value.map((item) => item.time), axisLabel: { color: chartText.value }, axisLine: { lineStyle: { color: chartGrid.value } }, axisTick: { show: false } },
  yAxis: { type: 'value', min: 0, max: 100, axisLabel: { color: chartText.value, formatter: '{value}%' }, splitLine: { lineStyle: { color: chartGrid.value, type: 'dashed' } } },
  series: [
    { name: 'CPU', type: 'line', smooth: true, showSymbol: false, data: history.value.map((item) => item.cpu), lineStyle: { width: 2, color: '#2563eb' }, areaStyle: { color: 'rgba(37,99,235,.08)' } },
    { name: '内存', type: 'line', smooth: true, showSymbol: false, data: history.value.map((item) => item.ram), lineStyle: { width: 2, color: '#059669' } },
    { name: '磁盘', type: 'line', smooth: true, showSymbol: false, data: history.value.map((item) => item.disk), lineStyle: { width: 2, color: '#d97706' } }
  ]
}))

const appendHistory = () => {
  const collected = state.value.collectedAt ? new Date(state.value.collectedAt) : new Date()
  history.value = [...history.value, {
    time: collected.toLocaleTimeString('zh-CN', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    cpu: cpuUsage.value, ram: ramUsage.value, disk: diskUsage.value
  }].slice(-30)
}

const reload = async (manual = false) => {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getSystemState()
    if (res.code !== 0 || !res.data?.server) throw new Error(res.msg || '接口未返回有效数据')
    state.value = res.data.server
    errorMessage.value = ''
    appendHistory()
    if (manual) ElMessage.success('服务器负载采集完成')
  } catch (error) {
    errorMessage.value = error?.message || '请检查服务器状态接口与访问权限'
  } finally {
    loading.value = false
  }
}

const stopTimer = () => {
  if (timer) window.clearInterval(timer)
  timer = null
}
const startTimer = () => {
  stopTimer()
  if (!paused.value) timer = window.setInterval(() => reload(false), refreshSeconds.value * 1000)
}
const togglePause = () => {
  paused.value = !paused.value
  startTimer()
}

watch(refreshSeconds, startTimer)
onMounted(async () => {
  await reload(false)
  startTimer()
})
onUnmounted(stopTimer)
</script>

<style scoped lang="scss">
.monitor-page { min-height: 100%; overflow-x: hidden; padding: 24px; background: var(--na-background); color: var(--na-foreground); }
.monitor-header { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; margin-bottom: 18px; }
.eyebrow { margin: 0 0 5px; color: var(--na-primary); font: 600 12px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; }
h1 { margin: 0; font-size: clamp(27px, 3vw, 36px); line-height: 1.2; }
.subtitle { margin: 8px 0 0; color: var(--na-muted-foreground); }
.monitor-actions { display: flex; align-items: center; justify-content: flex-end; gap: 8px; flex-wrap: wrap; }
.collector-state { display: inline-flex; min-height: 32px; align-items: center; gap: 7px; color: #047857; font-size: 12px; font-weight: 600; }
.collector-state i { width: 8px; height: 8px; border-radius: 50%; background: #10b981; }
.collector-state.paused { color: var(--na-muted-foreground); }
.collector-state.paused i { background: #94a3b8; }
.interval-select { width: 112px; }
.monitor-alert { margin-bottom: 16px; }
.kpi-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 14px; margin-bottom: 14px; }
.metric-card { display: grid; grid-template-columns: 46px minmax(0, 1fr); gap: 12px; padding: 17px; border: 1px solid var(--na-border); border-radius: var(--na-radius); background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.metric-icon { display: grid; width: 46px; height: 46px; place-items: center; border-radius: 11px; font-size: 21px; }
.metric-icon.cpu { color: #1d4ed8; background: #dbeafe; }.metric-icon.memory { color: #047857; background: #d1fae5; }.metric-icon.disk { color: #b45309; background: #fef3c7; }.metric-icon.load { color: #6d28d9; background: #ede9fe; }
.metric-card>div { display: flex; min-width: 0; flex-direction: column; }.metric-card span { color: var(--na-muted-foreground); font-size: 12px; }.metric-card strong { margin: 3px 0; font-size: 26px; font-variant-numeric: tabular-nums; }.metric-card small { overflow: hidden; color: var(--na-muted-foreground); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }.metric-card :deep(.el-progress) { grid-column: 1/-1; }
.primary-grid { display: grid; grid-template-columns: minmax(0, 1.8fr) minmax(300px, .7fr); gap: 14px; margin-bottom: 14px; }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.monitor-card { overflow: hidden; border: 1px solid var(--na-border); border-radius: var(--na-radius); background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.card-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 17px 20px; border-bottom: 1px solid var(--na-border); }
.card-header h2 { margin: 0; font-size: 16px; }.card-header p { margin: 4px 0 0; color: var(--na-muted-foreground); font-size: 12px; }.card-header>span { color: var(--na-muted-foreground); font-size: 12px; }
.runtime-list { margin: 0; padding: 8px 20px; }.runtime-list div { display: grid; grid-template-columns: 112px minmax(0, 1fr); gap: 12px; padding: 10px 0; border-bottom: 1px solid var(--na-border); }.runtime-list div:last-child { border-bottom: 0; }.runtime-list dt { color: var(--na-muted-foreground); font-size: 12px; }.runtime-list dd { overflow-wrap: anywhere; margin: 0; font-size: 12px; font-weight: 600; text-align: right; }.sample-time { padding: 12px 20px; border-top: 1px solid var(--na-border); color: var(--na-muted-foreground); font-size: 11px; text-align: right; }
.core-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px 20px; max-height: 340px; overflow-y: auto; padding: 18px 20px; }.core-row { display: grid; grid-template-columns: 64px minmax(0, 1fr); align-items: center; gap: 10px; }.core-row>span { color: var(--na-muted-foreground); font: 11px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; }
.disk-list { display: grid; gap: 16px; padding: 18px 20px; }.disk-row { display: grid; gap: 9px; }.disk-title { display: flex; justify-content: space-between; gap: 16px; }.disk-title strong { overflow: hidden; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }.disk-title span { flex: 0 0 auto; color: var(--na-muted-foreground); font-size: 11px; font-variant-numeric: tabular-nums; }
:global(html.dark) .metric-icon.cpu { color: #93c5fd; background: #172554; }:global(html.dark) .metric-icon.memory { color: #6ee7b7; background: #064e3b; }:global(html.dark) .metric-icon.disk { color: #fcd34d; background: #451a03; }:global(html.dark) .metric-icon.load { color: #c4b5fd; background: #2e1065; }
@media (max-width: 1200px) { .kpi-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.primary-grid { grid-template-columns: 1fr; } }
@media (max-width: 800px) { .monitor-page { padding: 14px; }.monitor-header { align-items: stretch; flex-direction: column; }.monitor-actions { justify-content: flex-start; }.detail-grid { grid-template-columns: 1fr; }.core-grid { grid-template-columns: 1fr; } }
@media (max-width: 520px) { .kpi-grid { grid-template-columns: 1fr; }.monitor-actions .el-button { flex: 1; }.collector-state { width: 100%; }.runtime-list div { grid-template-columns: 1fr; gap: 4px; }.runtime-list dd { text-align: left; } }
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { transition-duration: .01ms !important; } }
</style>
