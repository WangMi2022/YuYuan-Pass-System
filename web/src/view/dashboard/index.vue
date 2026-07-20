<template>
  <main class="orbit-workbench">
    <header class="orbit-welcome">
      <div>
        <span>{{ todayText }}</span>
        <h1>{{ greeting }}，{{ userStore.userInfo.nickName || '朋友' }}</h1>
        <p>让每一件资产沿正确的轨道流转，让每一次决策都有可靠数据。</p>
      </div>
      <div class="orbit-welcome__actions">
        <el-button :icon="DataAnalysis" @click="go('assetDashboard')">资产大屏</el-button>
        <el-button type="primary" :icon="Plus" @click="go('assetInventory')">登记新资产</el-button>
      </div>
    </header>

    <section class="orbit-overview">
      <LifecycleOrbit
        :formatted-total="formatNumber(summary.totalQuantity)"
        :retention-rate="retentionRate"
        :retention-number="retentionNumber"
        :items="lifecycleItems"
        @select="openAssets"
      />
      <DashboardSidePanel
        :attention-title="attentionTitle"
        :attention-count="attentionCount"
        :attentions="attentionItems"
        :actions="visibleQuickActions"
        @action="go"
      />
    </section>

    <section class="orbit-metrics" aria-label="资产概览">
      <article v-for="item in metrics" :key="item.label">
        <span>{{ item.label }}</span>
        <div>
          <strong>{{ item.value }}</strong>
          <small>{{ item.hint }}</small>
        </div>
        <span class="orbit-metrics__icon" :style="{ '--metric-color': item.color }">
          <el-icon><component :is="item.icon" /></el-icon>
        </span>
      </article>
    </section>
  </main>
</template>

<script setup>
import { computed, markRaw, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Box,
  Coin,
  CollectionTag,
  DataAnalysis,
  Delete,
  Files,
  Goods,
  Plus,
  RefreshLeft,
  Service,
  Switch,
  User
} from '@element-plus/icons-vue'
import LifecycleOrbit from '@/components/dashboard/LifecycleOrbit.vue'
import DashboardSidePanel from '@/components/dashboard/DashboardSidePanel.vue'
import { formatCompactCurrency, formatDateText, formatNumber } from '@/utils/format'
import { getAssetDashboard } from '@/plugin/asset/api/asset'
import { getNotifications } from '@/plugin/announcement/api/info'
import { useUserStore } from '@/pinia/modules/user'

defineOptions({ name: 'Dashboard' })

const router = useRouter()
const userStore = useUserStore()

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '夜深了'
  if (hour < 11) return '早上好'
  if (hour < 14) return '中午好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const todayText = new Date().toLocaleDateString('zh-CN', {
  year: 'numeric', month: 'long', day: 'numeric', weekday: 'long'
})

const summary = ref({
  totalQuantity: 0,
  assetKinds: 0,
  categoryCount: 0,
  originalValue: 0,
  currentValue: 0,
  depreciation: 0,
  statusSummary: []
})

const retentionNumber = computed(() => {
  const original = Number(summary.value.originalValue || 0)
  if (!original) return 0
  return Math.min(100, Math.max(0, Number(((Number(summary.value.currentValue || 0) / original) * 100).toFixed(1))))
})

const retentionRate = computed(() => `${retentionNumber.value.toFixed(1)}%`)

const statusDefinitions = [
  { status: 'pending_inbound', label: '待入库', color: 'var(--na-info)', icon: markRaw(Box) },
  { status: 'idle', label: '闲置', color: 'var(--na-muted-foreground)', icon: markRaw(Files) },
  { status: 'in_use', label: '使用中', color: 'var(--na-success)', icon: markRaw(Goods) },
  { status: 'maintenance', label: '维修中', color: 'var(--na-warning)', icon: markRaw(Service) },
  { status: 'retired', label: '已处置', color: 'var(--na-danger)', icon: markRaw(Delete) }
]

const lifecycleItems = computed(() => {
  const statusMap = Object.fromEntries(
    (summary.value.statusSummary || []).map((item) => [item.status, item])
  )
  return statusDefinitions.map((definition) => {
    const data = statusMap[definition.status] || {}
    const quantity = Number(data.quantity || 0)
    return {
      ...definition,
      quantity,
      formattedQuantity: formatNumber(quantity),
      assetKinds: formatNumber(Number(data.assetKinds || 0))
    }
  })
})

const metrics = computed(() => [
  {
    label: '资产实物总量', value: formatNumber(summary.value.totalQuantity),
    hint: `${formatNumber(summary.value.assetKinds)} 份档案`, icon: markRaw(Box), color: 'var(--na-chart-1)'
  },
  {
    label: '资产分类', value: formatNumber(summary.value.categoryCount),
    hint: '统一分类统计口径', icon: markRaw(CollectionTag), color: 'var(--na-chart-5)'
  },
  {
    label: '资产原值', value: formatCompactCurrency(summary.value.originalValue),
    hint: '数量 × 采购单价', icon: markRaw(Coin), color: 'var(--na-chart-2)'
  },
  {
    label: '当前估值', value: formatCompactCurrency(summary.value.currentValue),
    hint: `价值保有率 ${retentionRate.value}`, icon: markRaw(Goods), color: 'var(--na-chart-3)'
  },
  {
    label: '累计价值减少', value: formatCompactCurrency(summary.value.depreciation),
    hint: '原值与当前估值差额', icon: markRaw(DataAnalysis), color: 'var(--na-chart-4)'
  }
])

const quickActions = [
  { route: 'assetInbound', title: '资产入库', desc: '确认新资产入账', icon: markRaw(Box) },
  { route: 'assetIssue', title: '资产领用', desc: '办理人员领用', icon: markRaw(User) },
  { route: 'assetTransfer', title: '资产调拨', desc: '跨部门流转资产', icon: markRaw(Switch) },
  { route: 'assetReturn', title: '资产归还', desc: '完成资产交回', icon: markRaw(RefreshLeft) },
  { route: 'assetMaintenance', title: '维修申请', desc: '登记维修维保', icon: markRaw(Service) },
  { route: 'assetScrap', title: '报废处置', desc: '发起报废流程', icon: markRaw(Delete) }
]

const visibleQuickActions = computed(() => quickActions.filter((item) => router.hasRoute(item.route)).slice(0, 4))

const notices = ref([])
const unreadCount = computed(() => notices.value.filter((item) => !item.isRead).length)

const fallbackAttentionItems = computed(() => {
  const byStatus = Object.fromEntries(lifecycleItems.value.map((item) => [item.status, item]))
  return [
    {
      key: 'maintenance',
      title: `${byStatus.maintenance?.formattedQuantity || 0} 件资产正在维修`,
      meta: '可进入维修管理查看处理进度',
      tone: 1
    },
    {
      key: 'pending',
      title: `${byStatus.pending_inbound?.formattedQuantity || 0} 件资产等待入库`,
      meta: '完成验收后即可进入资产台账',
      tone: 2
    },
    {
      key: 'idle',
      title: `${byStatus.idle?.formattedQuantity || 0} 件资产当前闲置`,
      meta: '可结合领用与调拨计划提高利用率',
      tone: 'success'
    }
  ]
})

const attentionItems = computed(() => {
  if (!notices.value.length) return fallbackAttentionItems.value
  return notices.value.slice(0, 3).map((item, index) => ({
    key: item.ID || index,
    title: item.title,
    meta: `${item.publisher || '系统管理员'} · ${formatDateText(item.publishedAt || item.CreatedAt)}`,
    tone: item.isRead ? index : index === 0 ? 0 : index
  }))
})

const attentionTitle = computed(() => notices.value.length ? '最新公告' : '状态焦点')
const attentionCount = computed(() => notices.value.length ? unreadCount.value : 0)

const go = (name) => {
  if (router.hasRoute(name)) router.push({ name })
}

const openAssets = () => go('assetInventory')

onMounted(async () => {
  try {
    const res = await getAssetDashboard()
    if (res.code === 0) {
      summary.value = {
        ...summary.value,
        ...res.data,
        statusSummary: res.data?.statusSummary || []
      }
    }
  } catch {
    // 工作台概览加载失败不阻塞其他模块。
  }

  try {
    const res = await getNotifications({ limit: 6 })
    if (res.code === 0) notices.value = res.data?.list || res.data || []
  } catch {
    // 公告加载失败时回退到真实资产状态摘要。
  }
})
</script>

<style scoped lang="scss">
.orbit-workbench {
  min-height: 100%;
  padding: 26px;
  background:
    radial-gradient(100% 80% at 50% -20%, var(--na-primary-soft), transparent 48%),
    var(--na-background);
  color: var(--na-foreground);
}

.orbit-welcome {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  max-width: 1480px;
  margin: 0 auto 20px;
}

.orbit-welcome > div:first-child > span {
  color: var(--na-muted-foreground);
  font-size: 11px;
}

.orbit-welcome h1 {
  margin: 5px 0 0;
  font-size: clamp(23px, 2vw, 29px);
  font-weight: 660;
  letter-spacing: -.035em;
}

.orbit-welcome p {
  margin: 7px 0 0;
  color: var(--na-muted-foreground);
  font-size: 12px;
}

.orbit-welcome__actions {
  display: flex;
  flex: 0 0 auto;
  gap: 8px;
}

.orbit-overview {
  display: grid;
  max-width: 1480px;
  margin: 0 auto;
  grid-template-columns: minmax(0, 1fr) minmax(280px, 330px);
  gap: 14px;
}

.orbit-metrics {
  display: grid;
  max-width: 1480px;
  margin: 14px auto 0;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  border: 1px solid var(--na-border);
  border-radius: 15px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
}

.orbit-metrics article {
  position: relative;
  display: flex;
  min-width: 0;
  min-height: 94px;
  padding: 16px 48px 16px 18px;
  flex-direction: column;
  justify-content: center;
  border-right: 1px solid var(--na-border);
}

.orbit-metrics article:last-child { border-right: 0; }

.orbit-metrics article > span:first-child {
  color: var(--na-muted-foreground);
  font-size: 10px;
}

.orbit-metrics article > div {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
  margin-top: 5px;
}

.orbit-metrics strong {
  overflow: hidden;
  font-size: 19px;
  font-variant-numeric: tabular-nums;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.orbit-metrics small {
  overflow: hidden;
  color: var(--na-muted-foreground);
  font-size: 8px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.orbit-metrics__icon {
  position: absolute;
  top: 16px;
  right: 15px;
  display: grid;
  width: 29px;
  height: 29px;
  place-items: center;
  border-radius: 9px;
  background: color-mix(in srgb, var(--metric-color) 11%, transparent);
  color: var(--metric-color);
  font-size: 14px;
}

@media (max-width: 1200px) {
  .orbit-overview { grid-template-columns: minmax(0, 1fr) 290px; }
  .orbit-metrics { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .orbit-metrics article:nth-child(3) { border-right: 0; }
  .orbit-metrics article:nth-child(-n + 3) { border-bottom: 1px solid var(--na-border); }
}

@media (max-width: 980px) {
  .orbit-overview { grid-template-columns: 1fr; }
  .orbit-metrics { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .orbit-metrics article:nth-child(odd) { border-right: 1px solid var(--na-border); }
  .orbit-metrics article:nth-child(even) { border-right: 0; }
  .orbit-metrics article:nth-child(-n + 4) { border-bottom: 1px solid var(--na-border); }
}

@media (max-width: 720px) {
  .orbit-workbench { padding: 17px 14px 24px; }
  .orbit-welcome { align-items: stretch; flex-direction: column; gap: 15px; }
  .orbit-welcome__actions { flex-wrap: wrap; }
}

@media (max-width: 520px) {
  .orbit-metrics { grid-template-columns: 1fr; }
  .orbit-metrics article,
  .orbit-metrics article:nth-child(odd) {
    border-right: 0;
    border-bottom: 1px solid var(--na-border);
  }
  .orbit-metrics article:last-child { border-bottom: 0; }
}
</style>
