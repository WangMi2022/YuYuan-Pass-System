<template>
  <Teleport to="body">
    <section class="leadership-wallboard" aria-labelledby="wallboard-title">
      <header class="wallboard-header">
        <div class="wallboard-brand">
          <Logo :size="2.75" />
          <span>
            <strong>{{ brandingStore.systemName }}</strong>
            <small v-if="brandingStore.subtitle">{{ brandingStore.subtitle }}</small>
          </span>
        </div>

        <div class="wallboard-heading">
          <h1 id="wallboard-title">经营与风险驾驶舱</h1>
          <p>{{ snapshot.dateText }} · 数据范围：当前账号权限</p>
        </div>

        <div class="wallboard-controls">
          <span class="wallboard-freshness"><i />{{ loading ? '数据更新中' : (snapshot.freshnessText || `更新于 ${snapshot.updatedAt || '—'}`) }}</span>
          <button type="button" :disabled="loading" aria-label="刷新大屏数据" @click="$emit('refresh')">
            <el-icon><Refresh /></el-icon><span>刷新</span>
          </button>
          <button type="button" :aria-label="fullscreen ? '退出浏览器全屏' : '进入浏览器全屏'" @click="toggleFullscreen">
            <el-icon><FullScreen /></el-icon><span>{{ fullscreen ? '窗口' : '全屏' }}</span>
          </button>
          <button type="button" aria-label="退出会议室大屏" @click="exitWallboard">
            <el-icon><Close /></el-icon><span>退出</span>
          </button>
        </div>
      </header>

      <main class="wallboard-content" :aria-busy="loading">
        <section class="wallboard-kpis" aria-label="经营关键指标">
          <p v-if="!kpis.length" class="wallboard-kpis-empty">当前账号的统计数据暂不可用</p>
          <article v-for="item in kpis" :key="item.label" :class="`tone-${item.tone}`">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
            <small>{{ item.hint }}</small>
          </article>
        </section>

        <section class="wallboard-grid" :class="{ 'wallboard-grid--flexible': !fullWallboardLayout }">
          <article v-if="snapshot.access?.assets" class="wallboard-panel asset-overview">
            <header>
              <div><span>资产经营</span><h2>资产价值与状态</h2></div>
              <small>{{ moduleStatusText('assets', `${formatNumber(snapshot.asset?.totalQuantity)} 件实物`) }}</small>
            </header>
            <template v-if="snapshot.moduleLoaded?.assets">
            <div class="asset-value-line">
              <div><span>账面原值</span><strong>{{ formatCompactCurrency(snapshot.asset?.originalValue) }}</strong></div>
              <div><span>当前估值</span><strong>{{ formatCompactCurrency(snapshot.asset?.currentValue) }}</strong></div>
              <div><span>资产健康度</span><strong>{{ snapshot.healthRate || '0.0' }}%</strong></div>
            </div>
            <div class="asset-status-wall">
              <div v-for="item in assetStatuses" :key="item.key" class="asset-status-item">
                <span><i :class="`tone-${item.tone}`" />{{ item.label }}</span>
                <div><i :class="`tone-${item.tone}`" :style="{ width: `${item.ratio}%` }" /></div>
                <strong>{{ formatNumber(item.quantity) }}</strong>
              </div>
            </div>
            <footer>
              <span v-for="item in topLocations" :key="item.location">
                <small>{{ item.location || '未标注位置' }}</small>
                <strong>{{ formatNumber(item.quantity) }} 件</strong>
              </span>
              <span v-if="!topLocations.length" class="wallboard-empty">暂无位置分布数据</span>
            </footer>
            </template>
            <div v-else class="wallboard-empty">资产数据暂不可用</div>
          </article>

          <article v-if="snapshot.access?.risk" class="wallboard-panel risk-overview">
            <header>
              <div><span>风险闭环</span><h2>异常态势</h2></div>
              <small>{{ riskSummary }}</small>
            </header>
            <template v-if="snapshot.moduleLoaded?.risk">
            <div class="risk-summary-grid">
              <div><span>开放风险</span><strong>{{ formatNumber(snapshot.risk?.totalOpen) }}</strong></div>
              <div class="is-danger"><span>高风险</span><strong>{{ formatNumber(snapshot.risk?.highOpen) }}</strong></div>
              <div><span>今日新增</span><strong>{{ formatNumber(snapshot.risk?.todayNew) }}</strong></div>
              <div class="is-warning"><span>超期未结</span><strong>{{ formatNumber(snapshot.risk?.overdue) }}</strong></div>
            </div>
            <div v-if="riskTrend.length" class="risk-trend" aria-label="近期风险新增与关闭趋势">
              <div v-for="item in riskTrend" :key="item.date" class="risk-trend-item">
                <div>
                  <i class="risk-new" :style="{ height: `${item.newRatio}%` }" />
                  <i class="risk-resolved" :style="{ height: `${item.resolvedRatio}%` }" />
                </div>
                <small>{{ String(item.date || '').slice(5) }}</small>
              </div>
            </div>
            <div v-else class="wallboard-empty">暂无风险趋势数据</div>
            <footer class="trend-legend"><span><i class="risk-new" />新增</span><span><i class="risk-resolved" />关闭</span></footer>
            </template>
            <div v-else class="wallboard-empty">风险数据暂不可用</div>
          </article>

          <article v-if="snapshot.access?.invoices" class="wallboard-panel invoice-overview">
            <header>
              <div><span>资金流水</span><h2>近六个月确认金额</h2></div>
              <small>{{ moduleStatusText('invoices', `${formatNumber(snapshot.invoice?.confirmedCount)} 张已确认`) }}</small>
            </header>
            <template v-if="snapshot.moduleLoaded?.invoices">
            <div class="invoice-wall-layout">
              <div class="invoice-total-wall">
                <span>价税合计</span>
                <strong>{{ centsToCompactCurrency(snapshot.invoice?.totalCents) }}</strong>
                <small>待核 {{ formatNumber(snapshot.invoice?.pendingCount) }} · 识别失败 {{ formatNumber(snapshot.invoice?.failedCount) }}</small>
              </div>
              <div v-if="invoiceTrend.length" class="invoice-wall-trend" aria-label="发票确认金额趋势">
                <div v-for="item in invoiceTrend" :key="item.month">
                  <span>{{ centsToCompactCurrency(item.totalCents) }}</span>
                  <i :style="{ height: `${item.ratio}%` }" />
                  <small>{{ String(item.month || '').slice(5) }}月</small>
                </div>
              </div>
              <div v-else class="wallboard-empty">确认发票后将生成金额趋势</div>
            </div>
            </template>
            <div v-else class="wallboard-empty">发票数据暂不可用</div>
          </article>

          <aside v-if="hasFocusAccess" class="wallboard-panel focus-overview">
            <header>
              <div><span>今日关注</span><h2>待办与运行状态</h2></div>
            </header>
            <section v-if="hasPendingAccess" class="focus-queue">
              <div><span>全部待处理</span><strong>{{ pendingAvailable ? formatNumber(snapshot.pendingTotal) : '—' }}</strong></div>
              <div v-if="snapshot.access?.assetOperations"><span>资产业务</span><strong>{{ snapshot.moduleLoaded?.assetDrafts ? formatNumber(snapshot.assetDraftTotal) : '—' }}</strong></div>
              <div v-if="snapshot.access?.invoices"><span>发票待核</span><strong>{{ snapshot.moduleLoaded?.invoices ? formatNumber(snapshot.invoice?.pendingCount) : '—' }}</strong></div>
            </section>
            <p v-if="pendingStale" class="focus-stale-note">{{ pendingFreshnessText }}</p>
            <section v-if="snapshot.access?.calendar" class="focus-schedules" aria-label="今日日程">
              <h3>今日日程 <small v-if="snapshot.moduleFailed?.calendar && snapshot.moduleLoaded?.calendar">{{ moduleFreshnessShort('calendar') }}</small></h3>
              <div v-if="snapshot.moduleLoaded?.calendar && snapshot.schedules?.length">
                <p v-for="item in snapshot.schedules.slice(0, 4)" :key="item.id">
                  <time>{{ item.time }}</time><span>{{ item.title }}</span>
                </p>
              </div>
              <p v-else-if="snapshot.moduleLoaded?.calendar" class="wallboard-empty">今日暂无日程</p>
              <p v-else class="wallboard-empty">日程数据暂不可用</p>
            </section>
            <section v-if="snapshot.access?.monitor" class="focus-system" aria-label="服务器运行状态">
              <template v-if="snapshot.moduleLoaded?.monitor">
              <div class="system-heading">
                <span><i :class="`tone-${snapshot.systemHealth?.tone || 'success'}`" />系统{{ snapshot.systemHealth?.label || '运行正常' }}</span>
                <small v-if="snapshot.moduleFailed?.monitor">{{ moduleFreshnessShort('monitor') }}</small>
              </div>
              <dl>
                <div><dt>CPU</dt><dd>{{ percent(snapshot.systemUsage?.cpu) }}</dd></div>
                <div><dt>内存</dt><dd>{{ percent(snapshot.systemUsage?.ram) }}</dd></div>
                <div><dt>磁盘</dt><dd>{{ percent(snapshot.systemUsage?.disk) }}</dd></div>
              </dl>
              </template>
              <div v-else class="wallboard-empty">监控数据暂不可用</div>
            </section>
          </aside>
        </section>
      </main>
    </section>
  </Teleport>
</template>

<script setup>
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { Close, FullScreen, Refresh } from '@element-plus/icons-vue'
import Logo from '@/components/logo/index.vue'
import { useBrandingStore } from '@/pinia'
import { formatCompactCurrency, formatNumber } from '@/utils/format'

const props = defineProps({
  snapshot: { type: Object, required: true },
  loading: { type: Boolean, default: false }
})
const emit = defineEmits(['exit', 'refresh'])
const brandingStore = useBrandingStore()
const fullscreen = ref(typeof document !== 'undefined' && Boolean(document.fullscreenElement))

const assetStatuses = computed(() => props.snapshot.assetStatusRows || [])
const topLocations = computed(() => (props.snapshot.asset?.locationSummary || []).slice(0, 4))
const invoiceTrend = computed(() => props.snapshot.invoiceTrend || [])
const riskTrend = computed(() => {
  const values = (props.snapshot.risk?.trend || []).slice(-10)
  const maximum = Math.max(1, ...values.flatMap((item) => [Number(item.new || 0), Number(item.resolved || 0)]))
  return values.map((item) => ({
    ...item,
    newRatio: Math.max(Number(item.new || 0) ? 8 : 0, Number(item.new || 0) / maximum * 100),
    resolvedRatio: Math.max(Number(item.resolved || 0) ? 8 : 0, Number(item.resolved || 0) / maximum * 100)
  }))
})
const riskSummary = computed(() => moduleStatusText('risk', '实时风险扫描结果'))
const hasPendingAccess = computed(() => Boolean(props.snapshot.access?.assetOperations || props.snapshot.access?.invoices))
const hasFocusAccess = computed(() => Boolean(hasPendingAccess.value || props.snapshot.access?.calendar || props.snapshot.access?.monitor))
const fullWallboardLayout = computed(() => Boolean(
  props.snapshot.access?.assets && props.snapshot.access?.risk && props.snapshot.access?.invoices && hasFocusAccess.value
))
const pendingAvailable = computed(() => Boolean(
  (props.snapshot.access?.assetOperations && props.snapshot.moduleLoaded?.assetDrafts) ||
  (props.snapshot.access?.invoices && props.snapshot.moduleLoaded?.invoices)
))
const pendingStale = computed(() => Boolean(
  (props.snapshot.access?.assetOperations && props.snapshot.moduleLoaded?.assetDrafts && props.snapshot.moduleFailed?.assetDrafts) ||
  (props.snapshot.access?.invoices && props.snapshot.moduleLoaded?.invoices && props.snapshot.moduleFailed?.invoices)
))
const pendingFreshnessText = computed(() => {
  const times = ['assetDrafts', 'invoices']
    .filter((key) => props.snapshot.access?.[key === 'assetDrafts' ? 'assetOperations' : 'invoices'] && props.snapshot.moduleFailed?.[key])
    .map((key) => props.snapshot.moduleUpdatedAt?.[key])
    .filter(Boolean)
  return times.length ? `部分待办显示 ${times.sort().at(-1)} 的上次成功数据` : '部分待办显示上次成功数据'
})
const kpis = computed(() => {
  const items = []
  if (props.snapshot.access?.assets && props.snapshot.moduleLoaded?.assets) {
    items.push({ label: '资产账面原值', value: formatCompactCurrency(props.snapshot.asset?.originalValue), hint: hintWithFreshness('assets', `${formatNumber(props.snapshot.asset?.totalQuantity)} 件实物资产`), tone: 'primary' })
  }
  if (props.snapshot.access?.risk && props.snapshot.moduleLoaded?.risk) {
    items.push({ label: '开放风险', value: `${formatNumber(props.snapshot.risk?.totalOpen)} 项`, hint: hintWithFreshness('risk', `高风险 ${formatNumber(props.snapshot.risk?.highOpen)} 项`), tone: Number(props.snapshot.risk?.highOpen) ? 'danger' : 'success' })
  }
  if (props.snapshot.access?.invoices && props.snapshot.moduleLoaded?.invoices) {
    items.push({ label: '已确认流水', value: centsToCompactCurrency(props.snapshot.invoice?.totalCents), hint: hintWithFreshness('invoices', `${formatNumber(props.snapshot.invoice?.confirmedCount)} 张正式发票`), tone: 'info' })
  }
  if (pendingAvailable.value) {
    items.push({ label: '待处理事项', value: `${formatNumber(props.snapshot.pendingTotal)} 项`, hint: `${hasPendingAccess.value ? '当前权限内处理队列' : '处理队列'}${pendingStale.value ? ' · 上次成功数据' : ''}`, tone: Number(props.snapshot.pendingTotal) ? 'warning' : 'success' })
  }
  return items
})

function centsToCompactCurrency(value) {
  return formatCompactCurrency(Number(value || 0) / 100)
}
function percent(value) {
  return `${Math.min(100, Math.max(0, Number(value || 0))).toFixed(0)}%`
}
function moduleFreshnessShort(key) {
  return props.snapshot.moduleUpdatedAt?.[key] ? `上次成功 ${props.snapshot.moduleUpdatedAt[key]}` : '上次成功数据'
}
function moduleStatusText(key, successText) {
  if (!props.snapshot.moduleLoaded?.[key]) return '数据暂不可用'
  return props.snapshot.moduleFailed?.[key] ? moduleFreshnessShort(key) : successText
}
function hintWithFreshness(key, text) {
  return props.snapshot.moduleFailed?.[key] ? `${text} · 上次成功数据` : text
}
function onFullscreenChange() {
  fullscreen.value = Boolean(document.fullscreenElement)
}
async function toggleFullscreen() {
  try {
    if (document.fullscreenElement && document.exitFullscreen) await document.exitFullscreen()
    else if (document.documentElement.requestFullscreen) await document.documentElement.requestFullscreen()
  } catch {
    // The wallboard remains usable as a viewport-filling overlay.
  }
}
async function exitWallboard() {
  if (document.fullscreenElement && document.exitFullscreen) {
    try { await document.exitFullscreen() } catch { /* ignore browser refusal */ }
  }
  emit('exit')
}

function activateWallboard() {
  if (typeof document === 'undefined') return
  document.body.classList.add('wallboard-open')
  document.addEventListener('fullscreenchange', onFullscreenChange)
  onFullscreenChange()
}
function deactivateWallboard() {
  if (typeof document === 'undefined') return
  document.body.classList.remove('wallboard-open')
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  if (document.fullscreenElement && document.exitFullscreen) {
    document.exitFullscreen().catch(() => {})
  }
}

onMounted(activateWallboard)
onActivated(activateWallboard)
onDeactivated(deactivateWallboard)
onBeforeUnmount(deactivateWallboard)
</script>

<style scoped lang="scss">
.leadership-wallboard {
  --wb-bg: #0e1118;
  --wb-surface: #151a24;
  --wb-surface-raised: #1b2230;
  --wb-border: #2c3546;
  --wb-text: #f3f6fb;
  --wb-muted: #aab4c5;
  --wb-primary: var(--na-primary);
  position: fixed;
  z-index: 1400;
  inset: 0;
  min-width: 1120px;
  overflow: auto;
  background: var(--wb-bg);
  color: var(--wb-text);
  font-family: var(--na-font-sans);
}

.wallboard-header { display: grid; min-height: 88px; grid-template-columns: minmax(260px, .85fr) minmax(420px, 1.2fr) minmax(390px, 1fr); align-items: center; gap: 32px; padding: 16px 32px; border-bottom: 1px solid var(--wb-border); background: var(--wb-surface); }
.wallboard-brand { display: flex; min-width: 0; align-items: center; gap: 14px; }
.wallboard-brand > span { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.wallboard-brand strong, .wallboard-brand small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.wallboard-brand strong { font-size: 1rem; font-weight: 700; }
.wallboard-brand small { color: var(--wb-muted); font-size: .875rem; }
.wallboard-heading { min-width: 0; text-align: center; }
.wallboard-heading h1 { margin: 0; font-size: 2rem; font-weight: 700; line-height: 1.2; }
.wallboard-heading p { margin: 6px 0 0; color: var(--wb-muted); font-size: 1rem; line-height: 1.4; }
.wallboard-controls { display: flex; align-items: center; justify-content: flex-end; gap: 8px; }
.wallboard-freshness { display: inline-flex; align-items: center; gap: 8px; margin-right: 4px; color: var(--wb-muted); font-size: .875rem; font-variant-numeric: tabular-nums; white-space: nowrap; }
.wallboard-freshness i { width: 8px; height: 8px; border-radius: 50%; background: var(--na-success); }
.wallboard-controls button { display: inline-flex; min-width: 72px; min-height: 44px; align-items: center; justify-content: center; gap: 7px; padding: 0 12px; border: 1px solid var(--wb-border); border-radius: 9px; background: var(--wb-surface-raised); color: var(--wb-text); font: 600 .875rem/1 var(--na-font-sans); cursor: pointer; }
.wallboard-controls button:hover { border-color: var(--wb-primary); }
.wallboard-controls button:focus-visible { outline: 3px solid color-mix(in srgb, var(--wb-primary) 44%, transparent); outline-offset: 2px; }
.wallboard-controls button:disabled { cursor: wait; opacity: .55; }

.wallboard-content { display: grid; min-height: calc(100dvh - 88px); grid-template-rows: auto minmax(0, 1fr); gap: 24px; padding: 24px 32px 32px; }
.wallboard-kpis { display: flex; min-height: 128px; border-block: 1px solid var(--wb-border); }
.wallboard-kpis article { display: flex; min-width: 0; flex: 1; flex-direction: column; justify-content: center; gap: 7px; padding: 18px 28px; border-right: 1px solid var(--wb-border); }
.wallboard-kpis article:first-child { flex: 1.25; padding-left: 0; }
.wallboard-kpis article:last-child { border-right: 0; }
.wallboard-kpis-empty { display: grid; flex: 1; place-items: center; margin: 0; color: var(--wb-muted); font-size: 1rem; }
.wallboard-kpis span, .wallboard-kpis small { color: var(--wb-muted); font-size: 1rem; line-height: 1.4; }
.wallboard-kpis strong { overflow: hidden; font-size: 2.5rem; font-variant-numeric: tabular-nums; font-weight: 700; line-height: 1.15; text-overflow: ellipsis; white-space: nowrap; }
.wallboard-kpis .tone-danger strong { color: var(--na-danger); }
.wallboard-kpis .tone-warning strong { color: var(--na-warning); }
.wallboard-kpis .tone-success strong { color: var(--na-success); }
.wallboard-kpis .tone-info strong { color: var(--na-info); }
.wallboard-kpis .tone-primary strong { color: var(--wb-primary); }

.wallboard-grid { display: grid; min-height: 620px; grid-template-areas: "asset risk focus" "invoice invoice focus"; grid-template-columns: minmax(0, 1.28fr) minmax(360px, .92fr) minmax(330px, .72fr); grid-template-rows: minmax(300px, 1.08fr) minmax(260px, .92fr); gap: 16px; }
.wallboard-panel { min-width: 0; overflow: hidden; border: 1px solid var(--wb-border); border-radius: 12px; background: var(--wb-surface); }
.wallboard-panel > header { display: flex; min-height: 72px; align-items: center; justify-content: space-between; gap: 20px; padding: 14px 20px; border-bottom: 1px solid var(--wb-border); }
.wallboard-panel > header div { min-width: 0; }
.wallboard-panel > header span, .wallboard-panel > header small { color: var(--wb-muted); font-size: .875rem; }
.wallboard-panel > header h2 { margin: 4px 0 0; font-size: 1.5rem; font-weight: 600; line-height: 1.3; }
.asset-overview { grid-area: asset; }
.risk-overview { grid-area: risk; }
.invoice-overview { grid-area: invoice; }
.focus-overview { grid-area: focus; }
.wallboard-grid--flexible { grid-template-areas: none; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); grid-template-rows: auto; grid-auto-rows: auto; }
.wallboard-grid--flexible > .wallboard-panel { grid-area: auto; }
.wallboard-empty { display: grid; min-height: 84px; place-items: center; color: var(--wb-muted); font-size: 1rem; }

.asset-value-line { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); padding: 18px 20px; border-bottom: 1px solid var(--wb-border); }
.asset-value-line > div { min-width: 0; padding: 0 18px; border-right: 1px solid var(--wb-border); }
.asset-value-line > div:first-child { padding-left: 0; }
.asset-value-line > div:last-child { padding-right: 0; border-right: 0; }
.asset-value-line span { display: block; color: var(--wb-muted); font-size: .875rem; }
.asset-value-line strong { display: block; overflow: hidden; margin-top: 5px; font-size: 1.5rem; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.asset-status-wall { display: grid; gap: 12px; padding: 16px 20px; }
.asset-status-item { display: grid; grid-template-columns: 110px minmax(0, 1fr) 60px; align-items: center; gap: 12px; }
.asset-status-item > span { display: inline-flex; align-items: center; gap: 8px; color: var(--wb-muted); font-size: 1rem; }
.asset-status-item > span i { width: 8px; height: 8px; border-radius: 50%; }
.asset-status-item > div { height: 7px; overflow: hidden; border-radius: 4px; background: var(--wb-surface-raised); }
.asset-status-item > div i { display: block; height: 100%; border-radius: inherit; }
.asset-status-item > strong { font-size: 1rem; font-variant-numeric: tabular-nums; text-align: right; }
.asset-overview > footer { display: flex; gap: 24px; padding: 12px 20px 16px; border-top: 1px solid var(--wb-border); }
.asset-overview > footer > span { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.asset-overview > footer small { overflow: hidden; color: var(--wb-muted); font-size: .875rem; text-overflow: ellipsis; white-space: nowrap; }
.asset-overview > footer strong { font-size: 1rem; }

.risk-summary-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); padding: 16px 20px; border-bottom: 1px solid var(--wb-border); }
.risk-summary-grid > div { min-width: 0; padding: 0 12px; border-right: 1px solid var(--wb-border); }
.risk-summary-grid > div:first-child { padding-left: 0; }
.risk-summary-grid > div:last-child { padding-right: 0; border-right: 0; }
.risk-summary-grid span { display: block; color: var(--wb-muted); font-size: .875rem; }
.risk-summary-grid strong { display: block; margin-top: 4px; font-size: 1.5rem; font-variant-numeric: tabular-nums; }
.risk-summary-grid .is-danger strong { color: var(--na-danger); }
.risk-summary-grid .is-warning strong { color: var(--na-warning); }
.risk-trend { display: flex; height: 120px; align-items: stretch; gap: 8px; padding: 16px 20px 4px; }
.risk-trend-item { display: grid; min-width: 0; flex: 1; grid-template-rows: 1fr 18px; gap: 5px; }
.risk-trend-item > div { display: flex; align-items: end; justify-content: center; gap: 3px; border-bottom: 1px solid var(--wb-border); }
.risk-trend-item > div i { width: 8px; min-height: 2px; border-radius: 3px 3px 0 0; }
.risk-trend-item small { color: var(--wb-muted); font-size: .75rem; text-align: center; }
.risk-new { background: var(--na-danger); }
.risk-resolved { background: var(--na-success); }
.trend-legend { display: flex; justify-content: flex-end; gap: 16px; padding: 4px 20px 12px; }
.trend-legend span { display: inline-flex; align-items: center; gap: 6px; color: var(--wb-muted); font-size: .75rem; }
.trend-legend i { width: 8px; height: 8px; border-radius: 2px; }

.invoice-wall-layout { display: grid; min-height: calc(100% - 72px); grid-template-columns: 260px minmax(0, 1fr); }
.invoice-total-wall { display: flex; min-width: 0; flex-direction: column; justify-content: center; gap: 8px; padding: 20px 24px; border-right: 1px solid var(--wb-border); background: var(--wb-surface-raised); }
.invoice-total-wall span, .invoice-total-wall small { color: var(--wb-muted); font-size: 1rem; }
.invoice-total-wall strong { overflow: hidden; color: var(--wb-primary); font-size: 2.5rem; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.invoice-wall-trend { display: flex; min-width: 0; align-items: stretch; gap: 18px; padding: 18px 24px 12px; }
.invoice-wall-trend > div { display: grid; min-width: 0; flex: 1; grid-template-rows: 20px minmax(80px, 1fr) 22px; align-items: end; gap: 5px; }
.invoice-wall-trend span, .invoice-wall-trend small { overflow: hidden; color: var(--wb-muted); font-size: .875rem; text-align: center; text-overflow: ellipsis; white-space: nowrap; }
.invoice-wall-trend i { width: min(46px, 70%); min-height: 3px; justify-self: center; border-radius: 5px 5px 0 0; background: var(--wb-primary); }

.focus-queue { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); padding: 18px 16px; border-bottom: 1px solid var(--wb-border); }
.focus-queue > div { min-width: 0; padding: 0 10px; border-right: 1px solid var(--wb-border); }
.focus-queue > div:last-child { border-right: 0; }
.focus-queue span { display: block; color: var(--wb-muted); font-size: .875rem; }
.focus-queue strong { display: block; margin-top: 5px; font-size: 1.5rem; font-variant-numeric: tabular-nums; }
.focus-stale-note { margin: 0; padding: 8px 20px; border-bottom: 1px solid var(--wb-border); background: color-mix(in srgb, var(--na-warning) 12%, transparent); color: var(--na-warning); font-size: .875rem; line-height: 1.4; }
.focus-schedules { padding: 18px 20px; border-bottom: 1px solid var(--wb-border); }
.focus-schedules h3 { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin: 0 0 12px; font-size: 1rem; font-weight: 600; }
.focus-schedules h3 small, .system-heading small { color: var(--na-warning); font-size: .75rem; font-weight: 500; white-space: nowrap; }
.focus-schedules p { display: grid; grid-template-columns: 52px minmax(0, 1fr); gap: 12px; margin: 0; padding: 9px 0; border-bottom: 1px solid var(--wb-border); font-size: 1rem; }
.focus-schedules p:last-child { border-bottom: 0; }
.focus-schedules time { color: var(--wb-muted); font-variant-numeric: tabular-nums; }
.focus-schedules span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.focus-system { padding: 18px 20px; }
.system-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.system-heading > span { display: inline-flex; align-items: center; gap: 8px; font-size: 1rem; font-weight: 600; }
.system-heading i { width: 8px; height: 8px; border-radius: 50%; }
.focus-system dl { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin: 16px 0 0; }
.focus-system dl div { padding: 12px; border-radius: 8px; background: var(--wb-surface-raised); }
.focus-system dt { color: var(--wb-muted); font-size: .875rem; }
.focus-system dd { margin: 5px 0 0; font-size: 1.125rem; font-variant-numeric: tabular-nums; font-weight: 600; }

.tone-primary { background: var(--wb-primary) !important; }
.tone-success { background: var(--na-success) !important; }
.tone-warning { background: var(--na-warning) !important; }
.tone-danger { background: var(--na-danger) !important; }
.tone-info { background: var(--na-info) !important; }

@media (max-width: 1440px) {
  .wallboard-header { grid-template-columns: minmax(220px, .8fr) minmax(360px, 1fr) minmax(330px, .9fr); gap: 20px; padding-inline: 24px; }
  .wallboard-heading h1 { font-size: 1.5rem; }
  .wallboard-content { padding-inline: 24px; }
  .wallboard-grid { grid-template-columns: minmax(0, 1.2fr) minmax(320px, .9fr) minmax(300px, .72fr); }
  .wallboard-kpis strong, .invoice-total-wall strong { font-size: 2rem; }
}

@media (max-width: 1120px) {
  .leadership-wallboard { min-width: 0; }
  .wallboard-header { grid-template-columns: 1fr auto; }
  .wallboard-heading { grid-column: 1 / -1; grid-row: 2; padding-bottom: 8px; text-align: left; }
  .wallboard-controls { grid-column: 2; grid-row: 1; }
  .wallboard-freshness { display: none; }
  .wallboard-grid { grid-template-areas: "asset risk" "invoice invoice" "focus focus"; grid-template-columns: minmax(0, 1fr) minmax(320px, .8fr); grid-template-rows: auto; }
  .wallboard-grid.wallboard-grid--flexible { grid-template-areas: none; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); }
}

@media (max-width: 720px) {
  .wallboard-header { grid-template-columns: minmax(0, 1fr); gap: 14px; padding: 16px; }
  .wallboard-brand { grid-column: 1; grid-row: 1; }
  .wallboard-controls { width: 100%; grid-column: 1; grid-row: 2; }
  .wallboard-controls button { min-width: 0; flex: 1; }
  .wallboard-heading { grid-column: 1; grid-row: 3; padding-bottom: 0; }
  .wallboard-heading h1 { font-size: 1.375rem; }
  .wallboard-heading p { font-size: .875rem; }
  .wallboard-content { gap: 16px; padding: 16px; }
  .wallboard-kpis { display: grid; min-height: 0; grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .wallboard-kpis article, .wallboard-kpis article:first-child { min-height: 112px; padding: 16px; border-right: 0; border-bottom: 1px solid var(--wb-border); }
  .wallboard-kpis article:nth-child(odd) { border-right: 1px solid var(--wb-border); }
  .wallboard-kpis article:last-child,
  .wallboard-kpis article:nth-last-child(2):nth-child(odd) { border-bottom: 0; }
  .wallboard-kpis article:only-child,
  .wallboard-kpis article:last-child:nth-child(odd) { border-right: 0; }
  .wallboard-kpis strong { font-size: 1.75rem; }
  .wallboard-grid { min-height: 0; grid-template-areas: "asset" "risk" "invoice" "focus"; grid-template-columns: minmax(0, 1fr); grid-template-rows: auto; }
  .wallboard-grid.wallboard-grid--flexible { grid-template-areas: none; grid-template-columns: minmax(0, 1fr); }
  .risk-summary-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px 0; }
  .risk-summary-grid > div:nth-child(2) { padding-right: 0; border-right: 0; }
  .risk-summary-grid > div:nth-child(3) { padding-left: 0; }
  .invoice-wall-layout { grid-template-columns: 1fr; }
  .invoice-total-wall { border-right: 0; border-bottom: 1px solid var(--wb-border); }
  .invoice-wall-trend { gap: 8px; padding-inline: 14px; }
  .asset-overview > footer { flex-wrap: wrap; gap: 14px 24px; }
  .focus-queue { grid-template-columns: repeat(auto-fit, minmax(90px, 1fr)); }
}

@media (prefers-reduced-motion: reduce) {
  .leadership-wallboard *, .leadership-wallboard *::before, .leadership-wallboard *::after { scroll-behavior: auto !important; transition: none !important; }
}
</style>

<style>
body.wallboard-open { overflow: hidden; }
</style>
