<template>
  <div class="notification-center">
    <el-popover
      v-model:visible="popoverVisible"
      placement="bottom-end"
      trigger="click"
      :width="popoverWidth"
      popper-class="announcement-popper na-notification-popover"
      :show-arrow="false"
      transition="el-zoom-in-top"
    >
      <template #reference>
        <el-badge :value="unreadCount" :max="99" :hidden="unreadCount === 0" class="notification-badge">
          <button
            type="button"
            class="na-icon-button notification-trigger"
            :class="{ 'has-unread': unreadCount > 0 }"
            :aria-label="notificationLabel"
          >
            <el-icon class="notification-trigger__icon"><Bell /></el-icon>
          </button>
        </el-badge>
      </template>

      <section class="notification-panel" aria-label="公告中心">
        <!-- 头部标题与全部已读操作 -->
        <header class="notification-header">
          <div class="notification-header__title-group">
            <div class="notification-header__icon-box">
              <el-icon><Bell /></el-icon>
            </div>
            <div>
              <div class="notification-header__title-row">
                <h2>公告中心</h2>
                <span v-if="unreadCount > 0" class="notification-header__count-badge">
                  <span class="pulse-dot" />
                  {{ unreadCount }} 条待处理
                </span>
                <span v-else class="notification-header__count-badge is-clean">
                  <el-icon><CircleCheck /></el-icon>
                  全已读
                </span>
              </div>
              <p class="notification-header__subtext">聚合协同日程、系统公告与资产安全告警</p>
            </div>
          </div>
          <el-button
            v-if="unreadCount > 0"
            type="primary"
            link
            class="notification-header__read-all-btn"
            @click="readAll"
          >
            <el-icon><Check /></el-icon>
            全部已读
          </el-button>
        </header>

        <!-- 分类筛选导航条 -->
        <nav class="notification-tabs" aria-label="消息分类">
          <button
            type="button"
            class="notification-tab"
            :class="{ 'is-active': activeTab === 'all' }"
            @click="activeTab = 'all'"
          >
            全部
            <span class="tab-counter">{{ notifications.length }}</span>
          </button>
          <button
            type="button"
            class="notification-tab"
            :class="{ 'is-active': activeTab === 'unread' }"
            @click="activeTab = 'unread'"
          >
            未读
            <span v-if="unreadCount > 0" class="tab-counter is-unread">{{ unreadCount }}</span>
            <span v-else class="tab-counter">0</span>
          </button>
          <button
            type="button"
            class="notification-tab"
            :class="{ 'is-active': activeTab === 'schedule' }"
            @click="activeTab = 'schedule'"
          >
            日程
            <span class="tab-counter">{{ scheduleCount }}</span>
          </button>
          <button
            type="button"
            class="notification-tab"
            :class="{ 'is-active': activeTab === 'announcement' }"
            @click="activeTab = 'announcement'"
          >
            公告
            <span class="tab-counter">{{ announcementCount }}</span>
          </button>
        </nav>

        <!-- 消息列表滚动区 -->
        <el-scrollbar v-if="displayedNotifications.length" max-height="410px" class="notification-scroll">
          <div class="notification-list">
            <button
              v-for="item in displayedNotifications"
              :key="item.key"
              type="button"
              class="notification-item"
              :class="{
                'is-unread': !item.isRead,
                [`is-kind-${item.kind}`]: true
              }"
              @click="openNotification(item)"
            >
              <!-- 类别专属图标盒子 -->
              <div class="notification-avatar-box" :class="`type--${item.kind}`">
                <el-icon v-if="item.kind === 'schedule'"><Calendar /></el-icon>
                <el-icon v-else-if="item.kind === 'asset-risk'"><Warning /></el-icon>
                <el-icon v-else><ChatLineRound /></el-icon>
                <span v-if="!item.isRead" class="avatar-unread-dot" aria-hidden="true" />
              </div>

              <!-- 消息主体内容 -->
              <div class="notification-content">
                <div class="notification-content__top">
                  <strong class="notification-title" :title="item.title">{{ item.title }}</strong>
                  <time class="notification-time">{{ formatNoticeTime(notificationTime(item)) }}</time>
                </div>

                <p class="notification-summary">{{ plainText(item.content) }}</p>

                <div class="notification-meta">
                  <div class="notification-meta__tags">
                    <span class="notification-kind-tag" :class="`kind-${item.kind}`">
                      {{ notificationKind(item) }}
                    </span>
                    <span class="notification-publisher">
                      {{ item.kind === 'schedule' ? '工作日历' : item.publisher || '系统通知' }}
                    </span>
                  </div>

                  <!-- 悬停快捷“标为已读”或“查看”操作 -->
                  <div class="notification-actions">
                    <el-tooltip v-if="!item.isRead" content="标记为已读" placement="top" :show-after="400">
                      <span
                        class="quick-action-btn"
                        role="button"
                        tabindex="0"
                        aria-label="标记已读"
                        @click.stop="markReadSingle(item, $event)"
                        @keydown.enter.stop="markReadSingle(item, $event)"
                      >
                        <el-icon><Check /></el-icon>
                      </span>
                    </el-tooltip>
                    <el-icon class="action-arrow"><Right /></el-icon>
                  </div>
                </div>
              </div>
            </button>
          </div>
        </el-scrollbar>

        <!-- 空数据状态 -->
        <div v-else class="notification-empty">
          <div class="notification-empty__icon-circle">
            <el-icon v-if="activeTab === 'unread'"><CircleCheck /></el-icon>
            <el-icon v-else-if="activeTab === 'schedule'"><Calendar /></el-icon>
            <el-icon v-else><Bell /></el-icon>
          </div>
          <h4>{{ emptyTitle }}</h4>
          <p>{{ emptyDescription }}</p>
        </div>

        <!-- 底部快捷操作栏 -->
        <footer class="notification-footer">
          <button type="button" class="footer-link-btn" @click="navToSchedule">
            <el-icon><Calendar /></el-icon>
            <span>工作日历</span>
          </button>
          <div class="footer-divider" />
          <button type="button" class="footer-link-btn" @click="navToAnnouncements">
            <el-icon><Document /></el-icon>
            <span>公告管理</span>
          </button>
        </footer>
      </section>
    </el-popover>

    <!-- 公告详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      :title="currentNotice?.title || '公告详情'"
      width="min(92vw, 680px)"
      append-to-body
      destroy-on-close
      class="announcement-detail-modal"
    >
      <div v-if="currentNotice" class="announcement-detail">
        <header class="detail-header-card">
          <div class="detail-header-card__info">
            <span class="detail-kind-pill">{{ notificationKind(currentNotice) }}</span>
            <div class="detail-publisher-line">
              <span class="publisher-name">{{ currentNotice.publisher || '系统管理员' }}</span>
              <span class="meta-dot">·</span>
              <time class="publish-time">
                <el-icon><Clock /></el-icon>
                {{ formatFullTime(currentNotice.publishedAt || currentNotice.CreatedAt) }}
              </time>
            </div>
          </div>
        </header>

        <article class="announcement-rich-content prose-content" v-html="safeContent" />

        <div v-if="attachments.length" class="announcement-attachments">
          <h4>
            <el-icon><FolderOpened /></el-icon>
            相关附件 ({{ attachments.length }})
          </h4>
          <div class="attachment-cards">
            <a
              v-for="file in attachments"
              :key="file.uid || file.url"
              :href="getUrl(file.url)"
              target="_blank"
              rel="noopener noreferrer"
              class="attachment-card"
            >
              <el-icon class="attachment-icon"><Document /></el-icon>
              <div class="attachment-info">
                <span class="attachment-name">{{ file.name || '查看附件' }}</span>
                <span class="attachment-action">点击预览与下载</span>
              </div>
            </a>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="detailVisible = false">我知道了</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Bell,
  Calendar,
  ChatLineRound,
  Check,
  CircleCheck,
  Clock,
  Document,
  FolderOpened,
  Right,
  Warning
} from '@element-plus/icons-vue'
import { ElNotification } from 'element-plus'
import {
  getNotifications,
  markAllNotificationsRead,
  markNotificationRead
} from '@/plugin/announcement/api/info'
import {
  getWorkScheduleNotifications,
  markAllWorkScheduleNotificationsRead,
  markWorkScheduleNotificationRead
} from '@/api/workSchedule'
import { getBaseUrl } from '@/utils/format'
import { getUrl } from '@/utils/image'
import { useUserStore } from '@/pinia/modules/user'

const userStore = useUserStore()
const router = useRouter()
const notifications = ref([])
const unreadCount = ref(0)
const popoverVisible = ref(false)
const detailVisible = ref(false)
const currentNotice = ref(null)
const popoverWidth = ref(410)
const activeTab = ref('all')

let streamController
let reconnectTimer
let pollTimer
let disposed = false
const pollInterval = 60000

const notificationLabel = computed(() => unreadCount.value ? `公告中心，${unreadCount.value} 条未读` : '公告中心，无未读消息')
const attachments = computed(() => Array.isArray(currentNotice.value?.attachments) ? currentNotice.value.attachments : [])

const scheduleCount = computed(() => notifications.value.filter((i) => i.kind === 'schedule').length)
const announcementCount = computed(() => notifications.value.filter((i) => i.kind === 'announcement').length)

const displayedNotifications = computed(() => {
  if (activeTab.value === 'unread') {
    return notifications.value.filter((item) => !item.isRead)
  }
  if (activeTab.value === 'schedule') {
    return notifications.value.filter((item) => item.kind === 'schedule')
  }
  if (activeTab.value === 'announcement') {
    return notifications.value.filter((item) => item.kind === 'announcement')
  }
  return notifications.value
})

const emptyTitle = computed(() => {
  if (activeTab.value === 'unread') return '暂无未读消息'
  if (activeTab.value === 'schedule') return '暂无日程待办'
  if (activeTab.value === 'announcement') return '暂无系统公告'
  return '暂无通知消息'
})

const emptyDescription = computed(() => {
  if (activeTab.value === 'unread') return '太棒了！所有通知均已处理完毕'
  if (activeTab.value === 'schedule') return '近期工作日历暂无待处理日程'
  if (activeTab.value === 'announcement') return '系统近期无新发布的全局公告'
  return '当前没有任何待查看或待处理的事项'
})

const plainText = (html = '') => {
  const text = String(html).replace(/<[^>]*>/g, ' ').replace(/&nbsp;/gi, ' ').replace(/\s+/g, ' ').trim()
  return text || '点击查看详情'
}

const sanitizeHTML = (html = '') => {
  const doc = new DOMParser().parseFromString(String(html), 'text/html')
  doc.querySelectorAll('script,style,iframe,object,embed,form').forEach((node) => node.remove())
  doc.querySelectorAll('*').forEach((node) => {
    Array.from(node.attributes).forEach((attr) => {
      const name = attr.name.toLowerCase()
      const value = attr.value.trim().toLowerCase()
      if (name.startsWith('on') || ((name === 'href' || name === 'src') && value.startsWith('javascript:'))) {
        node.removeAttribute(attr.name)
      }
    })
  })
  return doc.body.innerHTML
}

const safeContent = computed(() => sanitizeHTML(currentNotice.value?.content || ''))

const formatNoticeTime = (value) => {
  if (!value) return ''
  const date = new Date(value)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  if (diff >= 0 && diff < 60 * 1000) return '刚刚'
  if (diff >= 0 && diff < 60 * 60 * 1000) return `${Math.floor(diff / 60000)} 分钟前`
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}

const formatFullTime = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : ''

const notificationTime = (item) => item.occurrenceAt || item.publishedAt || item.CreatedAt

const notificationKind = (item) => item.kind === 'schedule' ? '日程' : item.kind === 'asset-risk' ? '资产风险' : '公告'

const normalizeNotification = (item, kind) => ({
  ...item,
  kind,
  key: `${kind}-${item.ID ?? item.id}`,
  notificationID: Number(item.ID ?? item.id)
})

const notificationTimestamp = (item) => {
  const time = new Date(notificationTime(item)).getTime()
  return Number.isNaN(time) ? 0 : time
}

const loadNotifications = async () => {
  const [announcementResult, scheduleResult] = await Promise.allSettled([
    getNotifications({ limit: 16 }),
    getWorkScheduleNotifications({ limit: 16 })
  ])
  const sources = []
  let unread = 0
  let hasSuccessfulSource = false

  if (announcementResult.status === 'fulfilled' && announcementResult.value.code === 0) {
    const data = announcementResult.value.data || {}
    sources.push(...(data.list || []).map((item) => normalizeNotification(item, 'announcement')))
    unread += Number(data.unreadCount || 0)
    hasSuccessfulSource = true
  }
  if (scheduleResult.status === 'fulfilled' && scheduleResult.value.code === 0) {
    const data = scheduleResult.value.data || {}
    sources.push(...(data.list || []).map((item) => normalizeNotification(item, 'schedule')))
    unread += Number(data.unreadCount || 0)
    hasSuccessfulSource = true
  }

  if (hasSuccessfulSource) {
    notifications.value = sources.sort((left, right) => notificationTimestamp(right) - notificationTimestamp(left)).slice(0, 16)
    unreadCount.value = unread
  }
}

const markRead = async (item) => {
  if (item.isRead) return true
  item.isRead = true
  unreadCount.value = Math.max(0, unreadCount.value - 1)
  try {
    const result = item.kind === 'schedule'
      ? await markWorkScheduleNotificationRead({ id: item.notificationID })
      : await markNotificationRead({ id: item.notificationID })
    if (result.code === 0) return true
  } catch {
    // Fall through to a reload so the badge remains accurate.
  }
  await loadNotifications()
  return false
}

const markReadSingle = async (item, e) => {
  if (e) e.stopPropagation()
  await markRead(item)
}

const openNotification = async (item) => {
  popoverVisible.value = false
  await markRead(item)
  if (item.kind === 'schedule') {
    const dueDate = new Date(item.occurrenceAt)
    const date = Number.isNaN(dueDate.getTime()) ? '' : [
      dueDate.getFullYear(),
      String(dueDate.getMonth() + 1).padStart(2, '0'),
      String(dueDate.getDate()).padStart(2, '0')
    ].join('-')
    if (router.hasRoute('workSchedule')) {
      router.push({ name: 'workSchedule', query: date ? { date } : undefined })
    }
    return
  }
  currentNotice.value = item
  detailVisible.value = true
}

const readAll = async () => {
  const results = await Promise.allSettled([
    markAllNotificationsRead(),
    markAllWorkScheduleNotificationsRead()
  ])
  const succeeded = results.every((result) => result.status === 'fulfilled' && result.value.code === 0)
  if (succeeded) {
    notifications.value.forEach((item) => { item.isRead = true })
    unreadCount.value = 0
  } else {
    await loadNotifications()
  }
}

const navToSchedule = () => {
  popoverVisible.value = false
  if (router.hasRoute('workSchedule')) {
    router.push({ name: 'workSchedule' })
  }
}

const navToAnnouncements = () => {
  popoverVisible.value = false
  if (router.hasRoute('Info')) {
    router.push({ name: 'Info' })
  } else if (router.hasRoute('announcementInfo')) {
    router.push({ name: 'announcementInfo' })
  }
}

const handleStreamBlock = async (block) => {
  const lines = block.split('\n')
  const eventName = lines.find((line) => line.startsWith('event:'))?.slice(6).trim()
  if (eventName !== 'notification' && eventName !== 'announcement') return
  const data = lines.filter((line) => line.startsWith('data:')).map((line) => line.slice(5).trim()).join('')
  let event = {}
  try { event = JSON.parse(data) } catch { return }
  if (event.kind !== 'asset-risk') await loadNotifications()
  ElNotification({
    title: event.kind === 'schedule' ? '日程提醒' : event.kind === 'asset-risk' ? '资产风险提醒' : '新公告提醒',
    message: event.title || (event.kind === 'schedule' ? '有一条日程已到时间，请及时处理' : event.kind === 'asset-risk' ? '发现一条高风险资产事件，请及时处理' : '有一条新公告，请及时查看'),
    type: event.kind === 'asset-risk' ? 'warning' : 'info',
    duration: event.kind === 'asset-risk' ? 6500 : 5000,
    position: 'top-right',
    onClick: () => {
      if (event.kind === 'asset-risk' && router.hasRoute('assetRiskCenter')) {
        router.push({ name: 'assetRiskCenter', query: event.id ? { riskId: event.id } : undefined })
      }
    }
  })
}

const startPolling = () => {
  if (disposed || pollTimer) return
  pollTimer = setInterval(loadNotifications, pollInterval)
}

const stopPolling = () => {
  clearInterval(pollTimer)
  pollTimer = undefined
}

const scheduleReconnect = () => {
  if (disposed) return
  clearTimeout(reconnectTimer)
  reconnectTimer = setTimeout(startStream, 5000)
}

const startStream = async () => {
  if (disposed || !userStore.token) {
    startPolling()
    return
  }
  streamController?.abort()
  const controller = new AbortController()
  streamController = controller
  try {
    const response = await fetch(`${getBaseUrl()}/info/stream`, {
      headers: { Accept: 'text/event-stream', 'x-token': userStore.token },
      cache: 'no-store',
      signal: controller.signal
    })
    if (!response.ok || !response.body) throw new Error(`SSE ${response.status}`)
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    while (!disposed) {
      const { value, done } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true }).replace(/\r\n/g, '\n')
      const blocks = buffer.split('\n\n')
      buffer = blocks.pop() || ''
      for (const block of blocks) await handleStreamBlock(block)
    }
    if (!disposed && streamController === controller) {
      startPolling()
      scheduleReconnect()
    }
  } catch (error) {
    if (error?.name !== 'AbortError' && !disposed && streamController === controller) {
      startPolling()
      scheduleReconnect()
    }
  }
}

const updatePopoverWidth = () => {
  popoverWidth.value = Math.max(320, Math.min(410, window.innerWidth - 24))
}

const handleVisibility = () => {
  if (!document.hidden) loadNotifications()
}

onMounted(() => {
  updatePopoverWidth()
  window.addEventListener('resize', updatePopoverWidth)
  document.addEventListener('visibilitychange', handleVisibility)
  loadNotifications()
  startPolling()
  startStream()
})

onBeforeUnmount(() => {
  disposed = true
  streamController?.abort()
  clearTimeout(reconnectTimer)
  stopPolling()
  window.removeEventListener('resize', updatePopoverWidth)
  document.removeEventListener('visibilitychange', handleVisibility)
})
</script>

<style scoped lang="scss">
.notification-center {
  display: inline-flex;
}

.notification-badge {
  display: inline-flex;

  :deep(.el-badge__content.is-fixed) {
    top: 2px;
    right: 3px;
    transform: translate(25%, -25%);
    height: 16px;
    min-width: 16px;
    padding: 0 4px;
    border: 2px solid var(--na-card, #ffffff);
    border-radius: 10px;
    background: #ef4444;
    color: #ffffff;
    font-size: 10px;
    font-weight: 700;
    line-height: 12px;
    box-shadow: 0 2px 6px rgba(239, 68, 68, 0.4);
  }
}

.notification-trigger {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--na-border, #e2e0ec);
  border-radius: 10px;
  background: color-mix(in srgb, var(--na-muted, #f3f2f7) 45%, var(--na-card, #ffffff));
  color: var(--na-muted-foreground);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);

  &:hover {
    border-color: color-mix(in srgb, var(--na-primary) 35%, var(--na-border));
    background: var(--na-card, #ffffff);
    color: var(--na-primary);
    transform: translateY(-1px);
    box-shadow: 0 2px 8px color-mix(in srgb, var(--na-foreground) 4%, transparent);

    .notification-trigger__icon {
      color: var(--na-primary);
    }
  }

  &.has-unread {
    color: var(--na-foreground);
  }

  &__icon {
    font-size: 16px;
    transition: color 180ms ease, transform 180ms ease;
  }
}

.notification-panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin: -12px;
  border-radius: 16px;
  background: var(--na-card, #ffffff);
  color: var(--na-foreground, #19172c);
  box-shadow: 0 16px 36px -4px rgba(17, 24, 39, 0.12), 0 6px 14px -2px rgba(17, 24, 39, 0.06);
}

/* 头部样式 */
.notification-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 18px 12px;
  border-bottom: 1px solid var(--na-border, #e2e0ec);
  background: linear-gradient(180deg, var(--na-primary-soft, rgba(109, 93, 251, 0.05)) 0%, transparent 100%);

  &__title-group {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__icon-box {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.12));
    color: var(--na-primary, #6d5dfb);
    font-size: 18px;
    flex-shrink: 0;
  }

  &__title-row {
    display: flex;
    align-items: center;
    gap: 8px;

    h2 {
      margin: 0;
      color: var(--na-foreground, #19172c);
      font-size: 15px;
      font-weight: 700;
      letter-spacing: -0.01em;
    }
  }

  &__count-badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 2px 7px;
    border-radius: 100px;
    background: var(--na-danger-soft, rgba(220, 38, 38, 0.1));
    color: var(--na-danger, #dc2626);
    font-size: 11px;
    font-weight: 600;

    .pulse-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: var(--na-danger, #dc2626);
      animation: beacon-pulse 1.8s infinite;
    }

    &.is-clean {
      background: var(--na-success-soft, rgba(5, 150, 105, 0.1));
      color: var(--na-success, #059669);
    }
  }

  &__subtext {
    margin: 3px 0 0;
    color: var(--na-muted-foreground, #706b82);
    font-size: 12px;
  }

  &__read-all-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 5px 9px;
    border-radius: 8px;
    font-size: 12px;
    font-weight: 600;
    color: var(--na-primary, #6d5dfb);
    transition: all 0.15s ease;

    &:hover {
      background: var(--na-primary-soft, rgba(109, 93, 251, 0.12));
    }
  }
}

/* 分类导航条 */
.notification-tabs {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-bottom: 1px solid var(--na-border, #e2e0ec);
  background: var(--na-muted, #f3f2f7);
}

.notification-tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 11px;
  border: 0;
  border-radius: 7px;
  background: transparent;
  color: var(--na-muted-foreground, #706b82);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.18s ease;

  &:hover {
    color: var(--na-foreground, #19172c);
    background: color-mix(in srgb, var(--na-card) 60%, transparent);
  }

  &.is-active {
    background: var(--na-card, #ffffff);
    color: var(--na-primary, #6d5dfb);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  }

  .tab-counter {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 17px;
    height: 17px;
    padding: 0 4px;
    border-radius: 10px;
    background: var(--na-border, #e2e0ec);
    color: var(--na-muted-foreground, #706b82);
    font-size: 10px;
    font-weight: 700;

    &.is-unread {
      background: var(--na-danger, #dc2626);
      color: #ffffff;
    }
  }
}

/* 列表与消息卡片 */
.notification-scroll {
  padding: 4px 6px;
}

.notification-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 6px 4px;
}

.notification-item {
  position: relative;
  display: flex;
  align-items: flex-start;
  width: 100%;
  gap: 12px;
  padding: 12px;
  border: 1px solid transparent;
  border-radius: 10px;
  background: var(--na-card, #ffffff);
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);

  &:hover {
    border-color: var(--na-border-strong, #d5d1e2);
    background: var(--na-table-hover, #f6f4fc);
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(109, 93, 251, 0.06);

    .action-arrow {
      opacity: 1;
      transform: translateX(2px);
    }

    .quick-action-btn {
      opacity: 1;
      pointer-events: auto;
    }
  }

  &.is-unread {
    border-color: color-mix(in srgb, var(--na-primary) 22%, var(--na-border));
    background: color-mix(in srgb, var(--na-primary) 3.5%, var(--na-card));

    &::before {
      content: '';
      position: absolute;
      top: 10px;
      bottom: 10px;
      left: 0;
      width: 3px;
      border-radius: 0 3px 3px 0;
      background: var(--na-primary, #6d5dfb);
    }

    .notification-title {
      font-weight: 700;
      color: var(--na-foreground, #19172c);
    }
  }
}

/* 消息图标盒子 */
.notification-avatar-box {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  font-size: 16px;
  flex-shrink: 0;
  transition: transform 0.2s ease;

  &.type--schedule {
    background: var(--na-warning-soft, rgba(217, 119, 6, 0.12));
    color: var(--na-warning, #d97706);
  }

  &.type--announcement {
    background: var(--na-info-soft, rgba(2, 132, 199, 0.12));
    color: var(--na-info, #0284c7);
  }

  &.type--asset-risk {
    background: var(--na-danger-soft, rgba(220, 38, 38, 0.12));
    color: var(--na-danger, #dc2626);
  }

  .avatar-unread-dot {
    position: absolute;
    top: -2px;
    right: -2px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--na-primary, #6d5dfb);
    border: 2px solid var(--na-card, #ffffff);
    box-shadow: 0 0 0 2px rgba(109, 93, 251, 0.2);
  }
}

/* 内容详情区 */
.notification-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  gap: 4px;

  &__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
}

.notification-title {
  display: block;
  overflow: hidden;
  color: var(--na-foreground, #19172c);
  font-size: 13.5px;
  font-weight: 600;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-time {
  color: var(--na-muted-foreground, #706b82);
  font-size: 11px;
  white-space: nowrap;
  flex-shrink: 0;
}

.notification-summary {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  margin: 0;
  color: var(--na-muted-foreground, #706b82);
  font-size: 12px;
  line-height: 1.5;
  text-overflow: ellipsis;
}

.notification-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 4px;

  &__tags {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }
}

.notification-kind-tag {
  display: inline-flex;
  align-items: center;
  height: 19px;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 10.5px;
  font-weight: 650;
  line-height: 1;
  flex-shrink: 0;

  &.kind-schedule {
    background: var(--na-warning-soft, rgba(217, 119, 6, 0.12));
    color: var(--na-warning, #d97706);
  }

  &.kind-announcement {
    background: var(--na-info-soft, rgba(2, 132, 199, 0.12));
    color: var(--na-info, #0284c7);
  }

  &.kind-asset-risk {
    background: var(--na-danger-soft, rgba(220, 38, 38, 0.12));
    color: var(--na-danger, #dc2626);
  }
}

.notification-publisher {
  overflow: hidden;
  color: var(--na-muted-foreground, #706b82);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.quick-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--na-primary-soft, rgba(109, 93, 251, 0.12));
  color: var(--na-primary, #6d5dfb);
  font-size: 12px;
  opacity: 0;
  pointer-events: none;
  transition: all 0.15s ease;

  &:hover {
    background: var(--na-primary, #6d5dfb);
    color: #ffffff;
    transform: scale(1.1);
  }
}

.action-arrow {
  color: var(--na-muted-foreground, #706b82);
  font-size: 12px;
  opacity: 0.4;
  transition: all 0.2s ease;
}

/* 空状态 */
.notification-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 44px 20px;
  text-align: center;

  &__icon-circle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    margin-bottom: 12px;
    border-radius: 50%;
    background: var(--na-muted, #f3f2f7);
    color: var(--na-muted-foreground, #706b82);
    font-size: 24px;
  }

  h4 {
    margin: 0 0 4px;
    color: var(--na-foreground, #19172c);
    font-size: 14px;
    font-weight: 600;
  }

  p {
    margin: 0;
    color: var(--na-muted-foreground, #706b82);
    font-size: 12px;
  }
}

/* 底部操作条 */
.notification-footer {
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding: 10px 14px;
  border-top: 1px solid var(--na-border, #e2e0ec);
  background: var(--na-table-header, #faf9fc);
}

.footer-link-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--na-muted-foreground, #706b82);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.08));
    color: var(--na-primary, #6d5dfb);
  }
}

.footer-divider {
  width: 1px;
  height: 14px;
  background: var(--na-border, #e2e0ec);
}

/* 公告详情弹窗样式 */
.announcement-detail {
  min-height: 180px;
}

.detail-header-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  margin-bottom: 18px;
  border-radius: 10px;
  background: var(--na-muted, #f3f2f7);
  border: 1px solid var(--na-border, #e2e0ec);

  &__info {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }
}

.detail-kind-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--na-primary, #6d5dfb);
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
}

.detail-publisher-line {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--na-muted-foreground, #706b82);
  font-size: 12.5px;

  .publisher-name {
    font-weight: 600;
    color: var(--na-foreground, #19172c);
  }

  .meta-dot {
    color: var(--na-border-strong, #d5d1e2);
  }

  .publish-time {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
}

.prose-content {
  overflow-wrap: anywhere;
  color: var(--na-foreground, #19172c);
  font-size: 14.5px;
  line-height: 1.8;
  padding: 6px 2px;

  :deep(p) {
    margin: 0 0 12px;
  }

  :deep(img) {
    max-width: 100%;
    border-radius: 8px;
    height: auto;
  }

  :deep(table) {
    display: block;
    max-width: 100%;
    overflow-x: auto;
    border-collapse: collapse;
    margin: 12px 0;
  }

  :deep(th), :deep(td) {
    padding: 8px 12px;
    border: 1px solid var(--na-border, #e2e0ec);
  }

  :deep(th) {
    background: var(--na-muted, #f3f2f7);
  }
}

.announcement-attachments {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--na-border, #e2e0ec);

  h4 {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 12px;
    color: var(--na-foreground, #19172c);
    font-size: 13.5px;
    font-weight: 650;
  }
}

.attachment-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 10px;
}

.attachment-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid var(--na-border, #e2e0ec);
  border-radius: 8px;
  background: var(--na-card, #ffffff);
  color: inherit;
  text-decoration: none;
  transition: all 0.15s ease;

  &:hover {
    border-color: var(--na-primary, #6d5dfb);
    background: var(--na-primary-soft, rgba(109, 93, 251, 0.05));
    transform: translateY(-1px);

    .attachment-icon {
      color: var(--na-primary, #6d5dfb);
    }
  }

  .attachment-icon {
    font-size: 20px;
    color: var(--na-muted-foreground, #706b82);
    flex-shrink: 0;
    transition: color 0.15s ease;
  }

  .attachment-info {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .attachment-name {
    overflow: hidden;
    color: var(--na-foreground, #19172c);
    font-size: 13px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .attachment-action {
    color: var(--na-primary, #6d5dfb);
    font-size: 11px;
    margin-top: 2px;
  }
}

@media (prefers-reduced-motion: reduce) {
  * {
    scroll-behavior: auto !important;
    transition: none !important;
    animation: none !important;
  }
}
</style>
