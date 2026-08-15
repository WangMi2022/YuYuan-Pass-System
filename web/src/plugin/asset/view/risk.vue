<template>
  <main class="na-page na-page--list asset-risk-page">
    <AppPageHeader
      title-id="asset-risk-title"
      title="资产风险中心"
      description="持续检查资产状态、价值、归还、维修、质保与重复数据，并保留可追溯的处理证据。"
    >
      <template #actions>
        <span class="scan-state" :class="`is-${latestScan.status || 'idle'}`">
          <i />{{ latestScanText }}
        </span>
        <el-button :icon="Refresh" :loading="refreshing" :disabled="scanInProgress" @click="refreshAll">刷新</el-button>
        <el-button type="primary" :icon="Search" :loading="scanInProgress" @click="startScan()">立即扫描</el-button>
      </template>
    </AppPageHeader>

    <section v-loading="dashboardLoading" class="risk-kpis" aria-label="风险关键指标">
      <div>
        <span>待处理风险</span>
        <strong>{{ formatNumber(dashboard.totalOpen) }}</strong>
        <small>待处理与已确认事件</small>
      </div>
      <div class="is-danger">
        <span>高风险</span>
        <strong>{{ formatNumber(dashboard.highOpen) }}</strong>
        <small>高风险与严重风险</small>
      </div>
      <div class="is-info">
        <span>今日新增</span>
        <strong>{{ formatNumber(dashboard.todayNew) }}</strong>
        <small>今天首次命中的事件</small>
      </div>
      <div class="is-warning">
        <span>处理逾期</span>
        <strong>{{ formatNumber(dashboard.overdue) }}</strong>
        <small>超过 7 天仍未关闭</small>
      </div>
    </section>

    <section class="na-panel risk-insights" aria-label="风险趋势与分布">
      <div class="chart-region trend-region">
        <header>
          <div><h2>30 天变化</h2><p>新增风险与已关闭风险</p></div>
          <span>{{ formatDate(dashboard.generatedAt) }}</span>
        </header>
        <Chart v-if="hasTrend" :options="trendOptions" height="250px" />
        <el-empty v-else description="完成首次扫描后显示趋势" :image-size="64" />
      </div>
      <div class="chart-region distribution-region">
        <header>
          <div><h2>风险类型分布</h2><p>当前未关闭事件</p></div>
        </header>
        <Chart v-if="dashboard.byCategory.length" :options="categoryOptions" height="250px" />
        <el-empty v-else description="当前没有待处理风险" :image-size="64" />
      </div>
    </section>

    <section class="na-panel risk-workspace">
      <el-tabs v-model="activeTab" class="risk-tabs" @tab-change="handleTabChange">
        <el-tab-pane name="events">
          <template #label><span class="tab-label"><el-icon><Warning /></el-icon>风险事件</span></template>
          <div class="event-toolbar">
            <el-input v-model="searchForm.keyword" clearable :prefix-icon="Search" placeholder="资产编号、名称、规则或保管人" @keyup.enter="submitSearch" />
            <el-select v-model="searchForm.status" clearable placeholder="全部状态">
              <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-select v-model="searchForm.severity" clearable placeholder="全部等级">
              <el-option v-for="item in severityOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-select v-model="searchForm.category" clearable placeholder="全部类型">
              <el-option v-for="item in categoryFilters" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-button type="primary" :icon="Search" @click="submitSearch">查询</el-button>
            <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
          </div>

          <div v-if="selectedEvents.length" class="batch-toolbar">
            <span>已选择 {{ selectedEvents.length }} 条</span>
            <el-button type="primary" plain :icon="Check" @click="acknowledgeSelected">批量确认</el-button>
            <el-button :icon="User" @click="openAssign(selectedEvents)">批量分配</el-button>
          </div>

          <el-table
            v-loading="eventsLoading"
            :data="events"
            row-key="ID"
            stripe
            class="risk-table"
            @selection-change="selectedEvents = $event"
            @row-dblclick="openDetail"
          >
            <el-table-column type="selection" width="44" :selectable="isSelectableRisk" />
            <el-table-column label="风险" min-width="300" fixed="left">
              <template #default="{ row }">
                <button type="button" class="risk-identity" @click="openDetail(row)">
                  <span class="severity-mark" :class="`is-${row.severity}`" />
                  <span><strong>{{ row.title }}</strong><small>{{ row.description }}</small></span>
                </button>
              </template>
            </el-table-column>
            <el-table-column label="等级" width="96" align="center">
              <template #default="{ row }"><el-tag :type="severityMeta(row.severity).type" effect="light">{{ severityMeta(row.severity).label }}</el-tag></template>
            </el-table-column>
            <el-table-column label="风险类型" width="120">
              <template #default="{ row }">{{ categoryLabel(row.category) }}</template>
            </el-table-column>
            <el-table-column label="关联资产" min-width="190">
              <template #default="{ row }">
                <div class="asset-cell"><strong>{{ row.asset?.name || '资产已删除' }}</strong><small>{{ row.asset?.assetCode || `ID ${row.assetId}` }}</small></div>
              </template>
            </el-table-column>
            <el-table-column label="保管人" min-width="110">
              <template #default="{ row }">{{ row.asset?.custodian || '未登记' }}</template>
            </el-table-column>
            <el-table-column label="状态" width="108" align="center">
              <template #default="{ row }"><el-tag :type="statusMeta(row.status).type" effect="plain">{{ statusMeta(row.status).label }}</el-tag></template>
            </el-table-column>
            <el-table-column label="最近命中" width="166">
              <template #default="{ row }">{{ formatDate(row.lastDetectedAt) }}</template>
            </el-table-column>
            <el-table-column label="处理人" min-width="110">
              <template #default="{ row }">{{ row.assignedToName || '未分配' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="110" fixed="right" align="center">
              <template #default="{ row }"><el-button type="primary" link :icon="View" @click="openDetail(row)">查看</el-button></template>
            </el-table-column>
            <template #empty>
              <el-empty description="当前筛选条件下没有风险事件">
                <el-button type="primary" :icon="Search" :loading="scanInProgress" @click="startScan()">运行风险扫描</el-button>
              </el-empty>
            </template>
          </el-table>
          <div class="na-pagination risk-pagination">
            <el-pagination
              v-model:current-page="searchForm.page"
              v-model:page-size="searchForm.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="eventTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadEvents"
              @size-change="handlePageSize"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane name="rules">
          <template #label><span class="tab-label"><el-icon><Setting /></el-icon>规则配置</span></template>
          <header class="workspace-header">
            <div><h2>检测规则</h2><p>阈值或等级变更会生成新版本，历史事件继续保留原版本证据。</p></div>
            <el-button :icon="Refresh" :loading="rulesLoading" @click="loadRules">刷新规则</el-button>
          </header>
          <el-table v-loading="rulesLoading" :data="rules" row-key="ID" stripe>
            <el-table-column label="规则" min-width="260">
              <template #default="{ row }"><div class="rule-cell"><strong>{{ row.name }}</strong><code>{{ row.code }}</code></div></template>
            </el-table-column>
            <el-table-column label="类型" width="120"><template #default="{ row }">{{ categoryLabel(row.category) }}</template></el-table-column>
            <el-table-column label="等级" width="100" align="center"><template #default="{ row }"><el-tag :type="severityMeta(row.severity).type">{{ severityMeta(row.severity).label }}</el-tag></template></el-table-column>
            <el-table-column label="阈值" min-width="220"><template #default="{ row }"><span class="parameter-summary">{{ parametersSummary(row.parameters) }}</span></template></el-table-column>
            <el-table-column label="版本" width="80" align="center"><template #default="{ row }">v{{ row.version }}</template></el-table-column>
            <el-table-column label="状态" width="90" align="center"><template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag></template></el-table-column>
            <el-table-column label="操作" width="96" fixed="right" align="center"><template #default="{ row }"><el-button type="primary" link :icon="Edit" @click="openRule(row)">配置</el-button></template></el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane name="scans">
          <template #label><span class="tab-label"><el-icon><Clock /></el-icon>扫描记录</span></template>
          <header class="workspace-header">
            <div><h2>扫描运行</h2><p>扫描按批次提交并记录游标，失败任务可从上次完成位置继续。</p></div>
            <el-button :icon="Refresh" :loading="scansLoading" @click="loadScans">刷新记录</el-button>
          </header>
          <el-table v-loading="scansLoading" :data="scans" row-key="ID" stripe>
            <el-table-column label="运行 ID" width="100"><template #default="{ row }">#{{ row.ID }}</template></el-table-column>
            <el-table-column label="触发方式" width="100"><template #default="{ row }">{{ row.triggerType === 'scheduled' ? '定时扫描' : '手动扫描' }}</template></el-table-column>
            <el-table-column label="状态" width="100" align="center"><template #default="{ row }"><el-tag :type="scanStatusMeta(row.status).type">{{ scanStatusMeta(row.status).label }}</el-tag></template></el-table-column>
            <el-table-column prop="scannedAssets" label="扫描资产" width="100" align="right" />
            <el-table-column prop="newEvents" label="新增" width="86" align="right" />
            <el-table-column prop="updatedEvents" label="更新" width="86" align="right" />
            <el-table-column prop="closedEvents" label="关闭" width="86" align="right" />
            <el-table-column label="开始时间" width="166"><template #default="{ row }">{{ formatDate(row.startedAt) }}</template></el-table-column>
            <el-table-column label="完成时间" width="166"><template #default="{ row }">{{ formatDate(row.finishedAt) || '—' }}</template></el-table-column>
            <el-table-column label="结果" min-width="220"><template #default="{ row }"><span class="scan-error">{{ row.errorMessage || '扫描过程正常' }}</span></template></el-table-column>
            <el-table-column label="操作" width="100" fixed="right" align="center">
              <template #default="{ row }"><el-button v-if="canResumeScan(row)" type="primary" link :icon="RefreshRight" @click="startScan(row.ID)">继续扫描</el-button><span v-else>—</span></template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </section>

    <el-drawer v-model="detailVisible" :size="drawerSize" destroy-on-close :close-on-click-modal="false" class="risk-detail-drawer">
      <template #header>
        <div v-if="detail.event" class="detail-heading">
          <div><span class="severity-mark" :class="`is-${detail.event.severity}`" /><strong>{{ detail.event.title }}</strong></div>
          <p>{{ detail.event.ruleCode }} · v{{ detail.event.ruleVersion }}</p>
        </div>
      </template>
      <div v-loading="detailLoading" class="detail-content">
        <template v-if="detail.event">
          <section class="detail-status-row">
            <el-tag :type="severityMeta(detail.event.severity).type">{{ severityMeta(detail.event.severity).label }}</el-tag>
            <el-tag :type="statusMeta(detail.event.status).type" effect="plain">{{ statusMeta(detail.event.status).label }}</el-tag>
            <span>首次发现 {{ formatDate(detail.event.firstDetectedAt) }}</span>
          </section>
          <section class="detail-section asset-snapshot">
            <header><h3>关联资产</h3><el-button type="primary" link :icon="ArrowRight" @click="goToAsset">查看资产档案</el-button></header>
            <dl>
              <div><dt>资产名称</dt><dd>{{ detail.event.asset?.name || '资产已删除' }}</dd></div>
              <div><dt>资产编号</dt><dd>{{ detail.event.asset?.assetCode || `ID ${detail.event.assetId}` }}</dd></div>
              <div><dt>当前状态</dt><dd>{{ assetStatusLabel(detail.event.asset?.status) }}</dd></div>
              <div><dt>保管人</dt><dd>{{ detail.event.asset?.custodian || '未登记' }}</dd></div>
              <div><dt>存放位置</dt><dd>{{ detail.event.asset?.location || '未登记' }}</dd></div>
              <div><dt>当前估值</dt><dd>{{ formatCurrency(detail.event.asset?.currentValue) }}</dd></div>
            </dl>
          </section>
          <section class="detail-section explanation-section">
            <h3>风险说明</h3>
            <p>{{ detail.event.description }}</p>
            <div class="recommendation"><el-icon><Opportunity /></el-icon><span><strong>推荐动作</strong>{{ detail.event.recommendation }}</span></div>
          </section>
          <section class="detail-section">
            <h3>证据快照</h3>
            <dl class="evidence-list">
              <div v-for="item in evidenceRows" :key="item.key"><dt>{{ item.label }}</dt><dd><pre v-if="item.complex">{{ item.value }}</pre><span v-else>{{ item.value }}</span></dd></div>
            </dl>
          </section>
          <section class="detail-section">
            <h3>处理记录</h3>
            <el-timeline v-if="detail.logs.length">
              <el-timeline-item v-for="item in detail.logs" :key="item.ID" :timestamp="formatDate(item.CreatedAt)" placement="top">
                <div class="timeline-entry"><strong>{{ actionLabel(item.action) }}</strong><span>{{ item.actorName || '系统扫描' }}</span><p>{{ item.note || `${statusMeta(item.fromState).label} → ${statusMeta(item.toState).label}` }}</p></div>
              </el-timeline-item>
            </el-timeline>
            <el-empty v-else description="暂无处理记录" :image-size="56" />
          </section>
        </template>
      </div>
      <template #footer>
        <div v-if="detail.event" class="detail-actions">
          <el-button :icon="User" @click="openAssign([detail.event])">{{ detail.event.assignedTo ? '重新分配' : '分配处理人' }}</el-button>
          <el-button v-if="detail.event.status === 'open'" type="primary" plain :icon="Check" @click="acknowledgeOne">确认风险</el-button>
          <el-button v-if="['open', 'acknowledged'].includes(detail.event.status)" type="success" :icon="CircleCheck" @click="openAction('resolve')">标记解决</el-button>
          <el-button v-if="['open', 'acknowledged'].includes(detail.event.status)" type="warning" plain :icon="Hide" @click="openAction('ignore')">忽略</el-button>
          <el-button v-if="['resolved', 'ignored'].includes(detail.event.status)" type="primary" :icon="RefreshRight" @click="openAction('reopen')">重新打开</el-button>
        </div>
      </template>
    </el-drawer>

    <el-dialog v-model="actionVisible" :title="actionDialogTitle" width="min(92vw, 520px)" destroy-on-close>
      <el-form label-position="top">
        <el-form-item :label="actionType === 'reopen' ? '复核说明' : '处理说明'" required>
          <el-input v-model="actionNote" type="textarea" :rows="4" maxlength="500" show-word-limit :placeholder="actionPlaceholder" />
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="actionVisible = false">取消</el-button><el-button type="primary" :loading="actionSaving" @click="submitAction">确认</el-button></template>
    </el-dialog>

    <el-dialog v-model="assignVisible" title="分配风险处理人" width="min(92vw, 480px)" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="处理人">
          <el-select v-model="assignedTo" clearable filterable placeholder="选择处理人，清空表示取消分配" style="width: 100%">
            <el-option v-for="item in userOptions" :key="item.ID" :label="item.nickName || item.userName" :value="item.ID"><span>{{ item.nickName || item.userName }}</span><small class="user-option">{{ item.userName }}</small></el-option>
          </el-select>
        </el-form-item>
        <p class="dialog-note">高风险分配后会向在线处理人发送实时提醒。</p>
      </el-form>
      <template #footer><el-button @click="assignVisible = false">取消</el-button><el-button type="primary" :loading="assignSaving" @click="saveAssignment">保存分配</el-button></template>
    </el-dialog>

    <el-dialog v-model="ruleVisible" title="配置风险规则" width="min(94vw, 620px)" destroy-on-close>
      <div v-if="editingRule" class="rule-dialog">
        <header><strong>{{ editingRule.name }}</strong><code>{{ editingRule.code }} · v{{ editingRule.version }}</code></header>
        <p>{{ editingRule.description }}</p>
        <el-form label-position="top">
          <div class="rule-form-grid">
            <el-form-item label="风险等级">
              <el-select v-model="ruleForm.severity"><el-option v-for="item in severityOptions" :key="item.value" :label="item.label" :value="item.value" /></el-select>
            </el-form-item>
            <el-form-item label="启用规则"><el-switch v-model="ruleForm.enabled" inline-prompt active-text="启用" inactive-text="停用" /></el-form-item>
          </div>
          <el-form-item v-for="key in Object.keys(ruleForm.parameters)" :key="key" :label="parameterLabel(key)">
            <el-select v-if="Array.isArray(ruleForm.parameters[key])" v-model="ruleForm.parameters[key]" multiple filterable allow-create default-first-option style="width: 100%" />
            <el-input-number v-else v-model="ruleForm.parameters[key]" :min="parameterMin(key)" :max="parameterMax(key)" controls-position="right" style="width: 100%" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer><el-button @click="ruleVisible = false">取消</el-button><el-button type="primary" :loading="ruleSaving" @click="saveRule">保存并升版</el-button></template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowRight, Check, CircleCheck, Clock, Edit, Hide, Opportunity, Refresh,
  RefreshRight, Search, Setting, User, View, Warning
} from '@element-plus/icons-vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import Chart from '@/components/charts/index.vue'
import { chartPalette, chartTheme } from '@/components/charts/theme'
import { useAppStore } from '@/pinia'
import { getUserList } from '@/api/user'
import { formatCurrency, formatDate, formatNumber } from '@/utils/format'
import { createRiskScanPoller } from '@/plugin/asset/utils/riskScan'
import {
  acknowledgeAssetRisk, assignAssetRisk, getAssetRiskDashboard, getAssetRiskDetail,
  getAssetRiskList, getAssetRiskRules, getAssetRiskScans, ignoreAssetRisk,
  reopenAssetRisk, resolveAssetRisk, startAssetRiskScan, updateAssetRiskRule
} from '@/plugin/asset/api/risk'

defineOptions({ name: 'AssetRiskCenter' })

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const activeTab = ref('events')
const refreshing = ref(false)
const scanStarting = ref(false)
const activeScan = ref(null)
const dashboardLoading = ref(false)
const eventsLoading = ref(false)
const rulesLoading = ref(false)
const scansLoading = ref(false)
const detailLoading = ref(false)
const dashboard = ref({ totalOpen: 0, highOpen: 0, todayNew: 0, overdue: 0, byCategory: [], bySeverity: [], byStatus: [], byCustodian: [], trend: [], recentEvents: [], latestScan: null, generatedAt: null })
const events = ref([])
const eventTotal = ref(0)
const rules = ref([])
const scans = ref([])
const selectedEvents = ref([])
const detailVisible = ref(false)
const detail = ref({ event: null, logs: [] })
const userOptions = ref([])
const searchForm = reactive({ page: 1, pageSize: 20, keyword: '', status: '', severity: '', category: '' })
let scanPollErrorShown = false

const severityOptions = [
  { value: 'critical', label: '严重风险', type: 'danger' },
  { value: 'high', label: '高风险', type: 'danger' },
  { value: 'medium', label: '中风险', type: 'warning' },
  { value: 'low', label: '低风险', type: 'info' }
]
const statusOptions = [
  { value: 'open', label: '待处理', type: 'danger' },
  { value: 'acknowledged', label: '已确认', type: 'warning' },
  { value: 'resolved', label: '已解决', type: 'success' },
  { value: 'ignored', label: '已忽略', type: 'info' }
]
const categoryFilters = [
  { value: 'status', label: '状态异常' }, { value: 'value', label: '价值异常' },
  { value: 'return', label: '归还与调拨' }, { value: 'maintenance', label: '维修异常' },
  { value: 'warranty', label: '质保风险' }, { value: 'duplicate', label: '重复数据' }
]
const assetStatusLabels = { pending_inbound: '待入库', idle: '闲置', in_use: '使用中', maintenance: '维修中', retired: '已处置' }
const actionLabels = { detected: '首次发现', auto_reopen: '自动重新打开', auto_resolve: '自动解决', acknowledge: '确认风险', resolve: '标记解决', ignore: '忽略风险', reopen: '重新打开', assign: '分配处理人' }

const latestScan = computed(() => activeScan.value || dashboard.value.latestScan || {})
const scanInProgress = computed(() => scanStarting.value || latestScan.value.status === 'running')
const latestScanText = computed(() => {
  const scan = latestScan.value
  if (!scan.ID) return '尚未扫描'
  if (scan.status === 'running') return `扫描中 · ${scan.scannedAssets || 0} 项`
  if (scan.status === 'failed') return `扫描失败 · 可继续 #${scan.ID}`
  return `最近扫描 ${formatDate(scan.finishedAt || scan.startedAt)}`
})
const hasTrend = computed(() => dashboard.value.trend.some((item) => Number(item.new) || Number(item.resolved)))
const drawerSize = computed(() => window.innerWidth < 768 ? '96%' : '720px')

const trendOptions = computed(() => {
  void appStore.isDark
  void appStore.config.primaryColor
  const theme = chartTheme()
  const palette = chartPalette()
  return {
    animationDuration: 180,
    tooltip: { trigger: 'axis' },
    legend: { right: 0, top: 0, textStyle: { color: theme.muted } },
    grid: { left: 12, right: 12, top: 42, bottom: 8, containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: dashboard.value.trend.map((item) => item.date.slice(5)), axisLine: { lineStyle: { color: theme.grid } }, axisLabel: { color: theme.label, interval: 4 } },
    yAxis: { type: 'value', minInterval: 1, axisLabel: { color: theme.label }, splitLine: { lineStyle: { color: theme.grid } } },
    series: [
      { name: '新增风险', type: 'line', smooth: true, symbolSize: 5, data: dashboard.value.trend.map((item) => item.new), lineStyle: { width: 2, color: theme.danger }, itemStyle: { color: theme.danger }, areaStyle: { color: 'rgba(220, 38, 38, .08)' } },
      { name: '关闭风险', type: 'line', smooth: true, symbolSize: 5, data: dashboard.value.trend.map((item) => item.resolved), lineStyle: { width: 2, color: palette[2] }, itemStyle: { color: palette[2] } }
    ]
  }
})

const categoryOptions = computed(() => {
  void appStore.isDark
  void appStore.config.primaryColor
  const theme = chartTheme()
  const palette = chartPalette()
  const items = [...dashboard.value.byCategory].reverse()
  return {
    animationDuration: 180,
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 8, right: 24, top: 12, bottom: 8, containLabel: true },
    xAxis: { type: 'value', minInterval: 1, axisLabel: { color: theme.label }, splitLine: { lineStyle: { color: theme.grid } } },
    yAxis: { type: 'category', data: items.map((item) => item.label), axisLabel: { color: theme.text }, axisLine: { show: false }, axisTick: { show: false } },
    series: [{ type: 'bar', data: items.map((item, index) => ({ value: item.count, itemStyle: { color: palette[index % palette.length], borderRadius: [0, 3, 3, 0] } })), barMaxWidth: 18, label: { show: true, position: 'right', color: theme.muted } }]
  }
})

const severityMeta = (value) => severityOptions.find((item) => item.value === value) || { label: value || '未知', type: 'info' }
const statusMeta = (value) => statusOptions.find((item) => item.value === value) || { label: value || '未知', type: 'info' }
const scanStatusMeta = (value) => ({ running: { label: '运行中', type: 'warning' }, success: { label: '已完成', type: 'success' }, failed: { label: '失败', type: 'danger' } })[value] || { label: value || '未知', type: 'info' }
const categoryLabel = (value) => categoryFilters.find((item) => item.value === value)?.label || value || '未知'
const assetStatusLabel = (value) => assetStatusLabels[value] || value || '未知'
const actionLabel = (value) => actionLabels[value] || value
const canResumeScan = (scan) => scan.status === 'failed' && scans.value[0]?.ID === scan.ID

const loadDashboard = async () => {
  dashboardLoading.value = true
  try {
    const res = await getAssetRiskDashboard()
    if (res.code === 0) dashboard.value = { ...dashboard.value, ...res.data }
  } finally {
    dashboardLoading.value = false
  }
}
const loadEvents = async () => {
  eventsLoading.value = true
  try {
    const res = await getAssetRiskList({ ...searchForm })
    if (res.code === 0) {
      events.value = res.data?.list || []
      eventTotal.value = Number(res.data?.total || 0)
    }
  } finally {
    eventsLoading.value = false
  }
}
const loadRules = async () => {
  rulesLoading.value = true
  try {
    const res = await getAssetRiskRules()
    if (res.code === 0) rules.value = res.data || []
  } finally {
    rulesLoading.value = false
  }
}
const loadScans = async () => {
  scansLoading.value = true
  try {
    const res = await getAssetRiskScans({ page: 1, pageSize: 50 })
    if (res.code === 0) scans.value = res.data?.list || []
  } finally {
    scansLoading.value = false
  }
}
const requestScanStatus = async (runId) => {
  const res = await getAssetRiskScans({ page: 1, pageSize: 50 })
  if (res.code !== 0) throw new Error(res.msg || '无法读取扫描状态')
  return (res.data?.list || []).find((item) => Number(item.ID) === Number(runId))
}
const refreshAfterScan = async (run) => {
  try {
    await Promise.all([loadDashboard(), loadEvents(), loadScans()])
    if (run.status === 'success') ElMessage.success('资产风险扫描完成，列表已更新')
    else ElMessage.error(run.errorMessage || '资产风险扫描失败，请查看扫描记录')
  } finally {
    activeScan.value = null
  }
}
const scanPoller = createRiskScanPoller({
  requestStatus: requestScanStatus,
  onProgress(run) {
    scanPollErrorShown = false
    activeScan.value = run
    dashboard.value = { ...dashboard.value, latestScan: run }
  },
  onComplete: refreshAfterScan,
  onError() {
    if (scanPollErrorShown) return
    scanPollErrorShown = true
    ElMessage.warning('扫描状态暂时无法读取，系统将继续重试')
  }
})
const startScanPolling = (scan) => {
  if (!scan?.ID || scan.status !== 'running') {
    scanPoller.stop()
    activeScan.value = null
    return
  }
  scanPollErrorShown = false
  activeScan.value = scan
  scanPoller.start(scan.ID)
}
const refreshAll = async () => {
  refreshing.value = true
  try {
    await Promise.all([loadDashboard(), loadEvents(), loadRules(), loadScans()])
    startScanPolling(dashboard.value.latestScan)
  } finally {
    refreshing.value = false
  }
}
const submitSearch = () => { searchForm.page = 1; loadEvents() }
const resetSearch = () => { Object.assign(searchForm, { page: 1, pageSize: searchForm.pageSize, keyword: '', status: '', severity: '', category: '' }); loadEvents() }
const handlePageSize = () => { searchForm.page = 1; loadEvents() }
const handleTabChange = (name) => { if (name === 'rules' && !rules.value.length) loadRules(); if (name === 'scans') loadScans() }

const startScan = async (runId = 0) => {
  if (scanInProgress.value) return
  scanStarting.value = true
  try {
    const res = await startAssetRiskScan({ runId })
    if (res.code !== 0) return
    dashboard.value = { ...dashboard.value, latestScan: res.data }
    startScanPolling(res.data)
    ElMessage.success(runId ? '扫描任务已从上次游标继续' : '资产风险扫描已启动')
  } finally {
    scanStarting.value = false
  }
}

const openDetail = async (row) => {
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await getAssetRiskDetail({ id: row.ID })
    if (res.code === 0) detail.value = res.data
  } finally {
    detailLoading.value = false
  }
}
const syncRiskDetailFromRoute = async (value) => {
  const riskID = Number(value)
  if (!Number.isInteger(riskID) || riskID <= 0) {
    detailVisible.value = false
    detail.value = { event: null, logs: [] }
    return
  }
  if (detail.value.event?.ID === riskID && detailVisible.value) return
  await openDetail({ ID: riskID })
}
const refreshDetail = async () => {
  if (!detail.value.event?.ID) return
  const res = await getAssetRiskDetail({ id: detail.value.event.ID })
  if (res.code === 0) detail.value = res.data
}
const goToAsset = () => {
  const keyword = detail.value.event?.asset?.assetCode
  if (router.hasRoute('assetInventory')) router.push({ name: 'assetInventory', query: keyword ? { keyword } : undefined })
}

const evidenceLabels = { status: '资产状态', custodian: '保管人', currentValue: '当前估值', originalValue: '资产原值', difference: '金额差异', pendingDays: '待入库天数', thresholdDays: '规则阈值天数', idleDays: '闲置天数', idleSince: '闲置起始时间', maintenanceDays: '维修天数', maintenanceSince: '维修起始时间', inUseDays: '连续使用天数', inUseSince: '领用起始时间', warrantyEndDate: '质保到期日', remainingDays: '剩余天数', overdueDays: '过期天数', serialNumber: '序列号', normalizedSerial: '规范化序列号', count: '统计次数', thresholdCount: '次数阈值', windowDays: '统计窗口天数', operatedAt: '相关流转时间', assets: '重复资产', similarAssetId: '相似资产 ID', similarAssetCode: '相似资产编号', current: '资产当前数据', latestRecord: '最近流转记录', timeSource: '时间推导来源', minValue: '高价值阈值', categoryCode: '资产分类编码', createdAt: '资产创建时间' }
const evidenceRows = computed(() => Object.entries(detail.value.event?.evidence || {}).map(([key, value]) => ({ key, label: evidenceLabels[key] || key, complex: typeof value === 'object' && value !== null, value: formatEvidence(value) })))
const formatEvidence = (value) => {
  if (value === null || value === undefined || value === '') return '未登记'
  if (typeof value === 'number') return Number.isInteger(value) ? formatNumber(value) : value.toFixed(2)
  if (typeof value === 'boolean') return value ? '是' : '否'
  if (typeof value === 'object') return JSON.stringify(value, null, 2)
  const date = new Date(value)
  if (/^\d{4}-\d{2}-\d{2}T/.test(String(value)) && !Number.isNaN(date.getTime())) return formatDate(value)
  return String(value)
}

const isSelectableRisk = (row) => row.status === 'open'
const acknowledgeSelected = async () => {
  const ids = selectedEvents.value.filter((item) => item.status === 'open').map((item) => item.ID)
  if (!ids.length) return ElMessage.warning('请选择待处理风险')
  try {
    await ElMessageBox.confirm(`确认接手 ${ids.length} 条风险？`, '批量确认', { type: 'warning' })
  } catch {
    return
  }
  const res = await acknowledgeAssetRisk({ ids, note: '批量确认并进入处理' })
  if (res.code === 0) { ElMessage.success('风险已确认'); await Promise.all([loadEvents(), loadDashboard()]) }
}
const acknowledgeOne = async () => {
  const res = await acknowledgeAssetRisk({ ids: [detail.value.event.ID], note: '已确认并进入处理' })
  if (res.code === 0) { ElMessage.success('风险已确认'); await Promise.all([refreshDetail(), loadEvents(), loadDashboard()]) }
}

const actionVisible = ref(false)
const actionType = ref('resolve')
const actionNote = ref('')
const actionSaving = ref(false)
const actionDialogTitle = computed(() => ({ resolve: '解决风险', ignore: '忽略风险', reopen: '重新打开风险' })[actionType.value])
const actionPlaceholder = computed(() => ({ resolve: '说明已采取的修正措施或核验结果', ignore: '说明该风险为何可以忽略', reopen: '说明重新打开的原因和待复核事项' })[actionType.value])
const openAction = (type) => { actionType.value = type; actionNote.value = ''; actionVisible.value = true }
const submitAction = async () => {
  if (!actionNote.value.trim()) return ElMessage.warning('请填写处理说明')
  actionSaving.value = true
  try {
    const request = { resolve: resolveAssetRisk, ignore: ignoreAssetRisk, reopen: reopenAssetRisk }[actionType.value]
    const res = await request({ ids: [detail.value.event.ID], note: actionNote.value.trim() })
    if (res.code === 0) {
      ElMessage.success('风险状态已更新')
      actionVisible.value = false
      actionSaving.value = false
      void Promise.allSettled([refreshDetail(), loadEvents(), loadDashboard()])
    }
  } finally {
    actionSaving.value = false
  }
}

const assignVisible = ref(false)
const assignSaving = ref(false)
const assignedTo = ref()
const assigningIDs = ref([])
const loadUsers = async () => {
  if (userOptions.value.length) return
  const res = await getUserList({ page: 1, pageSize: 100 })
  if (res.code === 0) userOptions.value = (res.data?.list || []).filter((item) => Number(item.enable) === 1)
}
const openAssign = async (rows) => {
  assigningIDs.value = rows.map((item) => item.ID)
  assignedTo.value = rows.length === 1 ? rows[0].assignedTo || undefined : undefined
  assignVisible.value = true
  await loadUsers()
}
const saveAssignment = async () => {
  assignSaving.value = true
  try {
    const res = await assignAssetRisk({ ids: assigningIDs.value, assignedTo: assignedTo.value || 0 })
    if (res.code === 0) {
      ElMessage.success('处理人已更新')
      assignVisible.value = false
      await loadEvents()
      await refreshDetail()
    }
  } finally {
    assignSaving.value = false
  }
}

const ruleVisible = ref(false)
const ruleSaving = ref(false)
const editingRule = ref(null)
const ruleForm = reactive({ severity: 'medium', enabled: true, parameters: {} })
const openRule = (rule) => {
  editingRule.value = rule
  ruleForm.severity = rule.severity
  ruleForm.enabled = Boolean(rule.enabled)
  ruleForm.parameters = JSON.parse(JSON.stringify(rule.parameters || {}))
  ruleVisible.value = true
}
const saveRule = async () => {
  ruleSaving.value = true
  try {
    const res = await updateAssetRiskRule({ ID: editingRule.value.ID, severity: ruleForm.severity, enabled: ruleForm.enabled, parameters: ruleForm.parameters })
    if (res.code === 0) {
      ElMessage.success('规则已升版')
      ruleVisible.value = false
      await loadRules()
    }
  } finally {
    ruleSaving.value = false
  }
}
const parameterLabels = { days: '持续天数阈值', minValue: '资产价值阈值（元）', categoryCodes: '重要资产分类编码', windowDays: '统计窗口（天）', count: '次数阈值', maxDistance: '编号最大编辑距离', minLength: '编号最短长度' }
const parameterLabel = (key) => parameterLabels[key] || key
const parameterMin = (key) => key === 'minValue' ? 0 : key === 'count' ? 2 : key === 'minLength' ? 3 : 1
const parameterMax = (key) => key === 'minValue' ? 1e15 : key === 'count' ? 1000 : key === 'minLength' ? 50 : key === 'maxDistance' ? 2 : 3650
const parametersSummary = (parameters = {}) => {
  const entries = Object.entries(parameters)
  if (!entries.length) return '无需阈值'
  return entries.map(([key, value]) => `${parameterLabel(key)}：${Array.isArray(value) ? value.join('、') : value}`).join('；')
}

onMounted(async () => {
  await refreshAll()
  await syncRiskDetailFromRoute(route.query.riskId)
})
watch(() => route.query.riskId, syncRiskDetailFromRoute)
onBeforeUnmount(scanPoller.stop)
</script>

<style scoped lang="scss">
.asset-risk-page { display: flex; flex-direction: column; gap: 16px; }
.scan-state { display: inline-flex; align-items: center; gap: 7px; min-height: 32px; color: var(--na-muted-foreground); font-size: 12px; white-space: nowrap; }
.scan-state i { width: 7px; height: 7px; border-radius: 50%; background: var(--na-muted-foreground); }
.scan-state.is-running { color: var(--na-warning); }
.scan-state.is-running i { background: var(--na-warning); animation: scan-pulse 1.4s ease-out infinite; }
.scan-state.is-failed { color: var(--na-danger); }
.scan-state.is-failed i { background: var(--na-danger); }
.scan-state.is-success { color: var(--na-success); }
.scan-state.is-success i { background: var(--na-success); }
@keyframes scan-pulse { 0% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--na-warning) 35%, transparent); } 100% { box-shadow: 0 0 0 7px transparent; } }

.risk-kpis { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); overflow: hidden; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-card); }
.risk-kpis > div { min-width: 0; padding: 18px 22px; border-right: 1px solid var(--na-border); }
.risk-kpis > div:last-child { border-right: 0; }
.risk-kpis span, .risk-kpis small { display: block; color: var(--na-muted-foreground); }
.risk-kpis span { font-size: 13px; font-weight: 650; }
.risk-kpis strong { display: block; margin: 6px 0 3px; color: var(--na-foreground); font-size: 30px; line-height: 1; }
.risk-kpis small { overflow: hidden; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.risk-kpis .is-danger strong { color: var(--na-danger); }
.risk-kpis .is-warning strong { color: var(--na-warning); }
.risk-kpis .is-info strong { color: var(--na-info); }

.risk-insights { display: grid; grid-template-columns: minmax(0, 1.5fr) minmax(320px, .8fr); overflow: hidden; padding: 0; border-radius: 8px; }
.chart-region { min-width: 0; padding: 18px 20px 12px; }
.distribution-region { border-left: 1px solid var(--na-border); }
.chart-region > header, .workspace-header, .detail-section > header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.chart-region h2, .workspace-header h2, .detail-section h3 { margin: 0; color: var(--na-foreground); font-size: 15px; font-weight: 700; }
.chart-region p, .workspace-header p { margin: 3px 0 0; color: var(--na-muted-foreground); font-size: 12px; }
.chart-region > header > span { color: var(--na-muted-foreground); font-size: 11px; }

.risk-workspace { overflow: hidden; padding: 0 18px 18px; border-radius: 8px; }
.risk-tabs :deep(.el-tabs__header) { margin: 0 0 16px; }
.risk-tabs :deep(.el-tabs__nav-wrap::after) { height: 1px; background: var(--na-border); }
.tab-label { display: inline-flex; align-items: center; gap: 6px; }
.event-toolbar { display: grid; grid-template-columns: minmax(260px, 1fr) repeat(3, minmax(130px, 170px)) auto auto; gap: 10px; margin-bottom: 14px; }
.batch-toolbar { display: flex; align-items: center; gap: 10px; margin: -2px 0 12px; padding: 9px 12px; border: 1px solid var(--na-border); border-radius: 6px; background: var(--na-primary-soft); }
.batch-toolbar span { margin-right: auto; color: var(--na-accent-foreground); font-size: 13px; font-weight: 650; }
.risk-table { width: 100%; }
.risk-identity { display: flex; width: 100%; align-items: flex-start; gap: 10px; padding: 0; border: 0; background: transparent; color: inherit; text-align: left; }
.risk-identity > span:last-child, .asset-cell, .rule-cell { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.risk-identity strong, .asset-cell strong, .rule-cell strong { overflow: hidden; color: var(--na-foreground); font-size: 13px; font-weight: 680; text-overflow: ellipsis; white-space: nowrap; }
.risk-identity small, .asset-cell small { overflow: hidden; color: var(--na-muted-foreground); font-size: 11px; line-height: 1.45; text-overflow: ellipsis; white-space: nowrap; }
.severity-mark { display: inline-block; width: 8px; height: 8px; flex: 0 0 auto; margin-top: 5px; border-radius: 50%; background: var(--na-muted-foreground); }
.severity-mark.is-critical { background: #991b1b; }
.severity-mark.is-high { background: var(--na-danger); }
.severity-mark.is-medium { background: var(--na-warning); }
.severity-mark.is-low { background: var(--na-info); }
.risk-pagination { padding-top: 16px; }
.workspace-header { align-items: center; margin-bottom: 14px; }
.rule-cell code, .rule-dialog code { color: var(--na-muted-foreground); font-size: 11px; }
.parameter-summary, .scan-error { display: block; overflow: hidden; color: var(--na-muted-foreground); font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }

.detail-heading div { display: flex; align-items: center; gap: 9px; }
.detail-heading strong { color: var(--na-foreground); font-size: 16px; }
.detail-heading p { margin: 4px 0 0 17px; color: var(--na-muted-foreground); font-size: 11px; }
.detail-content { min-height: 260px; }
.detail-status-row { display: flex; align-items: center; gap: 8px; padding-bottom: 14px; border-bottom: 1px solid var(--na-border); }
.detail-status-row > span:last-child { margin-left: auto; color: var(--na-muted-foreground); font-size: 11px; }
.detail-section { padding: 18px 0; border-bottom: 1px solid var(--na-border); }
.detail-section:last-child { border-bottom: 0; }
.asset-snapshot dl, .evidence-list { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px 18px; margin: 14px 0 0; }
.asset-snapshot dl > div, .evidence-list > div { min-width: 0; }
dt { margin-bottom: 4px; color: var(--na-muted-foreground); font-size: 11px; }
dd { margin: 0; color: var(--na-foreground); font-size: 13px; overflow-wrap: anywhere; }
.explanation-section > p { margin: 10px 0 12px; color: var(--na-foreground); font-size: 13px; line-height: 1.7; }
.recommendation { display: flex; gap: 10px; padding: 12px; border-radius: 6px; background: var(--na-warning-soft); color: var(--na-action-warning); }
.recommendation .el-icon { margin-top: 2px; }
.recommendation span { display: flex; flex-direction: column; gap: 3px; font-size: 12px; line-height: 1.55; }
.recommendation strong { color: var(--na-foreground); font-size: 12px; }
.evidence-list { grid-template-columns: 1fr; gap: 10px; }
.evidence-list > div { display: grid; grid-template-columns: 150px minmax(0, 1fr); padding-bottom: 9px; border-bottom: 1px dashed var(--na-border); }
.evidence-list > div:last-child { border-bottom: 0; }
.evidence-list pre { overflow: auto; max-height: 220px; margin: 0; padding: 10px; border-radius: 5px; background: var(--na-muted); color: var(--na-foreground); font: 11px/1.6 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; white-space: pre-wrap; }
.timeline-entry { padding: 10px 12px; border: 1px solid var(--na-border); border-radius: 6px; }
.timeline-entry strong { margin-right: 8px; font-size: 13px; }
.timeline-entry span { color: var(--na-muted-foreground); font-size: 11px; }
.timeline-entry p { margin: 5px 0 0; color: var(--na-foreground); font-size: 12px; line-height: 1.55; }
.detail-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
.detail-actions .el-button + .el-button { margin-left: 0; }

.dialog-note { margin: 0; color: var(--na-muted-foreground); font-size: 12px; }
.user-option { float: right; margin-left: 24px; color: var(--na-muted-foreground); }
.rule-dialog > header { display: flex; flex-direction: column; gap: 4px; padding-bottom: 12px; border-bottom: 1px solid var(--na-border); }
.rule-dialog > p { margin: 12px 0 18px; color: var(--na-muted-foreground); font-size: 13px; line-height: 1.6; }
.rule-form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

@media (max-width: 1100px) {
  .asset-risk-page :deep(.na-page-header) { align-items: stretch; flex-direction: column; gap: 12px; }
  .asset-risk-page :deep(.na-page-actions) { flex-wrap: wrap; justify-content: flex-start; }
  .event-toolbar { grid-template-columns: minmax(240px, 1fr) repeat(2, minmax(130px, 1fr)) auto; }
  .event-toolbar > :nth-child(4) { grid-column: 1 / 2; }
}
@media (max-width: 820px) {
  .risk-kpis { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .risk-kpis > div:nth-child(2) { border-right: 0; }
  .risk-kpis > div:nth-child(-n+2) { border-bottom: 1px solid var(--na-border); }
  .risk-insights { grid-template-columns: 1fr; }
  .distribution-region { border-top: 1px solid var(--na-border); border-left: 0; }
  .event-toolbar { grid-template-columns: 1fr 1fr; }
  .event-toolbar > * { width: 100%; }
}
@media (max-width: 560px) {
  .asset-risk-page :deep(.na-page-actions) { gap: 8px; }
  .asset-risk-page :deep(.na-page-actions) .scan-state { flex: 1 0 100%; }
  .risk-kpis { grid-template-columns: 1fr; }
  .risk-kpis > div { border-right: 0; border-bottom: 1px solid var(--na-border); }
  .risk-kpis > div:last-child { border-bottom: 0; }
  .event-toolbar { grid-template-columns: 1fr; }
  .event-toolbar > :nth-child(4) { grid-column: auto; }
  .asset-snapshot dl, .rule-form-grid { grid-template-columns: 1fr; }
  .evidence-list > div { grid-template-columns: 1fr; gap: 4px; }
  .detail-status-row { flex-wrap: wrap; }
  .detail-status-row > span:last-child { width: 100%; margin-left: 0; }
}
@media (prefers-reduced-motion: reduce) { .scan-state.is-running i { animation: none; } }
</style>
