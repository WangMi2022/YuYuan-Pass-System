<template>
  <div class="notification-center">
    <el-popover
      v-model:visible="popoverVisible"
      placement="bottom-end"
      trigger="click"
      :width="popoverWidth"
      popper-class="announcement-popper"
    >
      <template #reference>
        <el-badge :value="unreadCount" :max="99" :hidden="unreadCount === 0" class="notification-badge">
          <button type="button" class="na-icon-button notification-trigger" :aria-label="notificationLabel">
            <el-icon><Bell /></el-icon>
          </button>
        </el-badge>
      </template>

      <section class="notification-panel" aria-label="公告中心">
        <header class="notification-header">
          <div>
            <h2>公告中心</h2>
            <p>{{ unreadCount ? `${unreadCount} 条消息待查看` : '没有未读消息' }}</p>
          </div>
          <el-button v-if="unreadCount" type="primary" link @click="readAll">全部已读</el-button>
        </header>

        <el-scrollbar v-if="notifications.length" max-height="420px">
          <div class="notification-list">
            <button
              v-for="item in notifications"
              :key="item.key"
              type="button"
              class="notification-item"
              :class="{ unread: !item.isRead }"
              @click="openNotification(item)"
            >
              <span class="notification-state" aria-hidden="true" />
              <span class="notification-content">
                <strong>{{ item.title }}</strong>
                <span class="notification-summary">{{ plainText(item.content) }}</span>
                <span class="notification-meta">
                  <span class="notification-kind" :class="`is-${item.kind}`">{{ notificationKind(item) }}</span>
                  <span>{{ item.kind === 'schedule' ? '工作日历' : item.publisher || '系统管理员' }}</span>
                  <time>{{ formatNoticeTime(notificationTime(item)) }}</time>
                </span>
              </span>
            </button>
          </div>
        </el-scrollbar>

        <el-empty v-else :image-size="64" description="暂无提醒" />
      </section>
    </el-popover>

    <el-dialog
      v-model="detailVisible"
      :title="currentNotice?.title || '公告详情'"
      width="min(92vw, 720px)"
      append-to-body
      destroy-on-close
    >
      <div v-if="currentNotice" class="announcement-detail">
        <div class="announcement-detail-meta">
          <span>{{ currentNotice.publisher || '系统管理员' }}</span>
          <time>{{ formatFullTime(currentNotice.publishedAt || currentNotice.CreatedAt) }}</time>
        </div>
        <article class="announcement-rich-content" v-html="safeContent" />
        <div v-if="attachments.length" class="announcement-attachments">
          <h3>相关附件</h3>
          <a
            v-for="file in attachments"
            :key="file.uid || file.url"
            :href="getUrl(file.url)"
            target="_blank"
            rel="noopener noreferrer"
          >{{ file.name || '查看附件' }}</a>
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
import { Bell } from '@element-plus/icons-vue'
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
const popoverWidth = ref(380)

let streamController
let reconnectTimer
let pollTimer
let disposed = false
const pollInterval = 60000

const notificationLabel = computed(() => unreadCount.value ? `公告中心，${unreadCount.value} 条未读` : '公告中心，无未读消息')
const attachments = computed(() => Array.isArray(currentNotice.value?.attachments) ? currentNotice.value.attachments : [])

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
  if (date.toDateString() === now.toDateString()) return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}

const formatFullTime = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : ''

const notificationTime = (item) => item.occurrenceAt || item.publishedAt || item.CreatedAt

const notificationKind = (item) => item.kind === 'schedule' ? '日程' : '公告'

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
    getNotifications({ limit: 12 }),
    getWorkScheduleNotifications({ limit: 12 })
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
    notifications.value = sources.sort((left, right) => notificationTimestamp(right) - notificationTimestamp(left)).slice(0, 12)
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

const handleStreamBlock = async (block) => {
  const lines = block.split('\n')
  const eventName = lines.find((line) => line.startsWith('event:'))?.slice(6).trim()
  if (eventName !== 'notification' && eventName !== 'announcement') return
  const data = lines.filter((line) => line.startsWith('data:')).map((line) => line.slice(5).trim()).join('')
  let event = {}
  try { event = JSON.parse(data) } catch { return }
  await loadNotifications()
  ElNotification({
    title: event.kind === 'schedule' ? '日程提醒' : '新公告提醒',
    message: event.title || (event.kind === 'schedule' ? '有一条日程已到时间，请及时处理' : '有一条新公告，请及时查看'),
    type: 'info',
    duration: 5000,
    position: 'top-right'
  })
}

const startPolling = () => {
  if (disposed || pollTimer) return
  // SSE is immediate on a single instance; polling keeps the inbox fresh when
  // a load balancer routes the stream and reminder worker to different nodes.
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
  popoverWidth.value = Math.max(280, Math.min(380, window.innerWidth - 24))
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
.notification-center { display: inline-flex; }
.notification-badge { display: inline-flex; }
.notification-badge :deep(.el-badge__content.is-fixed) { transform: translateY(-35%) translateX(55%); }
.notification-trigger { position: relative; }
.notification-panel { overflow: hidden; margin: -12px; border-radius: 10px; background: var(--el-bg-color); }
.notification-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 16px 18px 13px; border-bottom: 1px solid var(--el-border-color-lighter); }
.notification-header h2 { margin: 0; color: var(--el-text-color-primary); font-size: 16px; font-weight: 700; }
.notification-header p { margin: 3px 0 0; color: var(--el-text-color-secondary); font-size: 12px; }
.notification-list { padding: 6px; }
.notification-item { display: flex; width: 100%; gap: 10px; padding: 12px; border: 0; border-radius: 8px; background: transparent; color: inherit; text-align: left; cursor: pointer; }
.notification-item:hover,
.notification-item:focus-visible { background: var(--el-fill-color-light); outline: none; }
.notification-item:focus-visible { box-shadow: inset 0 0 0 2px var(--el-color-primary-light-5); }
.notification-state { width: 7px; height: 7px; flex: 0 0 auto; margin-top: 7px; border-radius: 50%; background: var(--el-border-color); }
.notification-item.unread .notification-state { background: var(--el-color-primary); box-shadow: 0 0 0 3px var(--el-color-primary-light-9); }
.notification-content { display: flex; min-width: 0; flex: 1; flex-direction: column; gap: 4px; }
.notification-content strong { overflow: hidden; color: var(--el-text-color-primary); font-size: 14px; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.notification-item.unread .notification-content strong { font-weight: 750; }
.notification-summary { overflow: hidden; color: var(--el-text-color-secondary); font-size: 12px; line-height: 1.5; text-overflow: ellipsis; white-space: nowrap; }
.notification-meta { display: flex; justify-content: space-between; gap: 12px; color: var(--el-text-color-placeholder); font-size: 11px; }
.notification-meta > span:not(.notification-kind) { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.notification-kind { display: inline-flex; flex: 0 0 auto; align-items: center; min-height: 18px; padding: 0 5px; border-radius: 4px; color: var(--el-color-primary); background: var(--el-color-primary-light-9); font-size: 10px; font-weight: 650; }
.notification-kind.is-schedule { color: var(--el-color-warning); background: var(--el-color-warning-light-9); }
.announcement-detail { min-height: 180px; }
.announcement-detail-meta { display: flex; gap: 14px; margin-bottom: 18px; padding-bottom: 12px; border-bottom: 1px solid var(--el-border-color-lighter); color: var(--el-text-color-secondary); font-size: 12px; }
.announcement-rich-content { overflow-wrap: anywhere; color: var(--el-text-color-primary); font-size: 14px; line-height: 1.75; }
.announcement-rich-content :deep(img) { max-width: 100%; height: auto; }
.announcement-rich-content :deep(table) { display: block; max-width: 100%; overflow-x: auto; border-collapse: collapse; }
.announcement-attachments { margin-top: 22px; padding-top: 14px; border-top: 1px solid var(--el-border-color-lighter); }
.announcement-attachments h3 { margin: 0 0 10px; font-size: 14px; }
.announcement-attachments a { display: inline-flex; margin: 0 10px 8px 0; color: var(--el-color-primary); font-size: 13px; }
@media (prefers-reduced-motion: reduce) { * { scroll-behavior: auto !important; transition: none !important; } }
</style>
