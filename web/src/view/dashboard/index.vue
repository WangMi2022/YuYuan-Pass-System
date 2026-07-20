<template>
  <main class="workbench">
    <!-- 问候横幅 · Three.js 波浪粒子 -->
    <section class="hero na-panel">
      <HeroCanvas />
      <div class="hero-content">
        <div>
          <p class="hero-kicker">{{ todayText }}</p>
          <h1>{{ greeting }}，{{ userStore.userInfo.nickName || '朋友' }}</h1>
          <p class="hero-subtitle">欢迎回到资产管理中心，今天也保持从容高效。</p>
        </div>
        <div class="hero-actions">
          <el-button type="primary" :icon="DataAnalysis" @click="go('assetDashboard')">进入资产大屏</el-button>
          <el-button :icon="Plus" @click="go('assetInventory')">登记资产</el-button>
        </div>
      </div>
    </section>

    <!-- 资产概览 -->
    <section class="kpi-row" aria-label="资产概览">
      <article v-for="item in kpis" :key="item.label" class="kpi na-panel" :style="{ '--kpi-color': item.color }">
        <div class="kpi-icon"><el-icon><component :is="item.icon" /></el-icon></div>
        <div class="kpi-body">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </article>
    </section>

    <div class="workbench-grid">
      <!-- 快捷入口 -->
      <section class="na-panel">
        <header class="na-panel-header"><h2 class="na-panel-title">快捷入口</h2></header>
        <div class="shortcut-grid">
          <button
            v-for="item in visibleShortcuts"
            :key="item.route"
            type="button"
            class="shortcut"
            @click="go(item.route)"
          >
            <span class="shortcut-icon"><el-icon><component :is="item.icon" /></el-icon></span>
            <span class="shortcut-text">
              <strong>{{ item.title }}</strong>
              <small>{{ item.desc }}</small>
            </span>
            <el-icon class="shortcut-arrow"><ArrowRight /></el-icon>
          </button>
        </div>
      </section>

      <!-- 最新公告 -->
      <section class="na-panel">
        <header class="na-panel-header">
          <h2 class="na-panel-title">最新公告</h2>
          <el-tag v-if="unreadCount" type="danger" effect="plain" round size="small">{{ unreadCount }} 未读</el-tag>
        </header>
        <div v-if="notices.length" class="notice-list">
          <div v-for="item in notices" :key="item.ID" class="notice-item" :class="{ 'is-unread': !item.isRead }">
            <span class="notice-dot" />
            <div class="notice-body">
              <strong>{{ item.title }}</strong>
              <small>{{ item.publisher || '系统管理员' }} · {{ formatDateText(item.publishedAt || item.CreatedAt) }}</small>
            </div>
          </div>
        </div>
        <el-empty v-else description="暂无公告" :image-size="72" />
      </section>
    </div>
  </main>
</template>

<script setup>
  import { computed, markRaw, onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import {
    ArrowRight, Box, Coin, CollectionTag, DataAnalysis, Document, Files,
    Goods, Plus, Service, Setting, User
  } from '@element-plus/icons-vue'
  import HeroCanvas from '@/components/three/HeroCanvas.vue'
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

  const summary = ref({ totalQuantity: 0, assetKinds: 0, categoryCount: 0, originalValue: 0, currentValue: 0 })
  const kpis = computed(() => [
    { label: '资产实物总量', value: formatNumber(summary.value.totalQuantity), icon: markRaw(Box), color: 'var(--na-chart-1)' },
    { label: '资产档案', value: formatNumber(summary.value.assetKinds), icon: markRaw(Files), color: 'var(--na-chart-2)' },
    { label: '资产分类', value: formatNumber(summary.value.categoryCount), icon: markRaw(CollectionTag), color: 'var(--na-chart-5)' },
    { label: '资产原值', value: formatCompactCurrency(summary.value.originalValue), icon: markRaw(Coin), color: 'var(--na-chart-4)' },
    { label: '当前估值', value: formatCompactCurrency(summary.value.currentValue), icon: markRaw(Goods), color: 'var(--na-chart-3)' }
  ])

  const shortcuts = [
    { icon: markRaw(DataAnalysis), title: '资产大屏', desc: '总览资产结构与价值', route: 'assetDashboard' },
    { icon: markRaw(Files), title: '资产档案', desc: '登记与维护资产信息', route: 'assetInventory' },
    { icon: markRaw(Document), title: '文档中心', desc: '在线预览与协作编辑', route: 'documentViewer' },
    { icon: markRaw(User), title: '用户管理', desc: '维护账号与角色指派', route: 'user' },
    { icon: markRaw(Service), title: '角色管理', desc: '配置权限与数据范围', route: 'authority' },
    { icon: markRaw(Setting), title: '系统设置', desc: '登录外观与系统参数', route: 'systemSetting' }
  ]
  // 仅展示当前用户权限内已注册的路由
  const visibleShortcuts = computed(() => shortcuts.filter((item) => router.hasRoute(item.route)))

  const go = (name) => {
    if (router.hasRoute(name)) router.push({ name })
  }

  const notices = ref([])
  const unreadCount = computed(() => notices.value.filter((item) => !item.isRead).length)

  onMounted(async () => {
    try {
      const res = await getAssetDashboard()
      if (res.code === 0) summary.value = { ...summary.value, ...res.data }
    } catch { /* 概览加载失败不阻塞工作台 */ }
    try {
      const res = await getNotifications({ limit: 6 })
      if (res.code === 0) notices.value = res.data?.list || res.data || []
    } catch { /* 公告加载失败不阻塞工作台 */ }
  })
</script>

<style scoped lang="scss">
  .workbench {
    min-height: 100%;
    width: 100%;
    max-width: 1440px;
    margin: 0 auto;
    padding: 28px 32px 38px;
    background: transparent;
    color: var(--na-foreground);
  }

  /* Hero */
  .hero {
    position: relative;
    overflow: hidden;
    min-height: 90px;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }
  .hero-canvas { opacity: .18; }
  .hero-content {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 20px;
    padding: 10px 0 20px;
  }
  .hero-kicker { margin: 0 0 7px; color: var(--na-muted-foreground); font-size: 10px; }
  .hero h1 { margin: 0; font-size: 25px; font-weight: 570; letter-spacing: -.035em; }
  .hero-subtitle { margin: 7px 0 0; color: var(--na-muted-foreground); font-size: 12px; }
  .hero-actions { display: flex; flex: 0 0 auto; gap: 8px; }

  /* KPI row */
  .kpi-row {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 0;
    overflow: hidden;
    margin-top: 0;
    border: 1px solid var(--na-border);
    border-radius: 12px;
    background: var(--na-card);
    box-shadow: var(--na-shadow-sm);
  }
  .kpi {
    display: flex;
    align-items: center;
    gap: 0;
    padding: 16px 18px;
    border: 0;
    border-right: 1px solid var(--na-border);
    border-radius: 0;
    background: transparent;
    box-shadow: none;
    transition: background-color 150ms ease;
  }
  .kpi:last-child { border-right: 0; }
  .kpi:hover { background: color-mix(in srgb, var(--na-primary) 3%, var(--na-card)); }
  .kpi-icon {
    display: none;
  }
  .kpi-body { display: flex; min-width: 0; flex-direction: column; }
  .kpi-body span { color: var(--na-muted-foreground); font-size: 10px; }
  .kpi-body strong {
    overflow: hidden;
    margin-top: 5px;
    font-size: 19px;
    font-weight: 570;
    font-variant-numeric: tabular-nums;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Two-column grid */
  .workbench-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 285px;
    gap: 14px;
    margin-top: 14px;
  }

  .shortcut-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
    padding: 14px 16px 16px;
  }
  .shortcut {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    border: 1px solid var(--na-border);
    border-radius: var(--na-radius-sm);
    background: transparent;
    color: var(--na-foreground);
    text-align: left;
    transition: border-color 150ms ease, background-color 150ms ease;
  }
  .shortcut:hover { border-color: color-mix(in srgb, var(--na-primary) 40%, var(--na-border)); background: var(--na-primary-soft); }
  .shortcut:hover .shortcut-arrow { translate: 2px 0; color: var(--na-primary); }
  .shortcut-icon {
    display: grid;
    width: 36px;
    height: 36px;
    place-items: center;
    flex: 0 0 36px;
    border-radius: 9px;
    color: var(--na-primary);
    background: var(--na-primary-soft);
    font-size: 16px;
  }
  .shortcut-text { display: flex; min-width: 0; flex: 1; flex-direction: column; gap: 2px; }
  .shortcut-text strong { font-size: 13px; font-weight: 550; }
  .shortcut-text small { overflow: hidden; color: var(--na-muted-foreground); font-size: 11.5px; text-overflow: ellipsis; white-space: nowrap; }
  .shortcut-arrow { color: var(--na-muted-foreground); font-size: 13px; transition: translate 150ms ease, color 150ms ease; }

  .notice-list { padding: 6px 16px 12px; }
  .notice-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 11px 2px;
    border-bottom: 1px solid var(--na-border);
  }
  .notice-item:last-child { border-bottom: 0; }
  .notice-dot {
    flex: 0 0 6px;
    width: 6px;
    height: 6px;
    margin-top: 6px;
    border-radius: 50%;
    background: var(--na-border-strong);
  }
  .notice-item.is-unread .notice-dot { background: var(--na-primary); box-shadow: 0 0 0 3px var(--na-primary-soft); }
  .notice-body { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
  .notice-body strong {
    overflow: hidden;
    font-size: 13px;
    font-weight: 500;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .notice-item.is-unread .notice-body strong { font-weight: 650; }
  .notice-body small { color: var(--na-muted-foreground); font-size: 11.5px; }

  @media (max-width: 1200px) {
    .kpi-row { grid-template-columns: repeat(3, 1fr); }
    .kpi { border-bottom: 1px solid var(--na-border); }
    .kpi:nth-child(3) { border-right: 0; }
    .kpi:nth-child(n + 4) { border-bottom: 0; }
  }
  @media (max-width: 900px) {
    .workbench { padding: 14px; }
    .workbench-grid { grid-template-columns: 1fr; }
    .hero-content { align-items: stretch; flex-direction: column; }
    .hero-actions { flex-wrap: wrap; }
    .kpi-row { grid-template-columns: repeat(2, 1fr); }
    .kpi:nth-child(3) { border-right: 1px solid var(--na-border); }
    .kpi:nth-child(even) { border-right: 0; }
    .kpi:nth-child(n + 4) { border-bottom: 1px solid var(--na-border); }
    .kpi:last-child { border-right: 0; border-bottom: 0; }
    .shortcut-grid { grid-template-columns: 1fr; }
  }
  @media (max-width: 560px) {
    .kpi-row { grid-template-columns: 1fr; }
    .kpi { border-right: 0; border-bottom: 1px solid var(--na-border); }
    .kpi:last-child { border-bottom: 0; }
  }
</style>
