<template>
  <main class="na-page na-page--list smart-copilot-page">
    <AppPageHeader
      title-id="smart-copilot-title"
      title="业务助手"
      description="只读查询资产、风险、发票、日程和公告；每条回答都保留可追溯的业务记录。"
    >
      <template #actions>
        <el-tooltip content="刷新会话与可查询范围" placement="bottom">
          <el-button circle :icon="Refresh" :loading="loading" aria-label="刷新会话与可查询范围" @click="loadSessions" />
        </el-tooltip>
      </template>
    </AppPageHeader>

    <div class="copilot-layout">
      <aside class="na-panel session-panel" aria-label="业务助手会话">
        <header class="panel-header">
          <div>
            <h2>会话</h2>
            <span>共 {{ sessionTotal }} 个会话</span>
          </div>
          <el-button text :icon="Plus" @click="newSession">新建</el-button>
        </header>

        <div class="session-list">
          <div
            v-for="item in sessions"
            :key="item.ID"
            class="session-row"
            :class="{ 'is-active': item.ID === sessionId }"
          >
            <button
              class="session-item"
              type="button"
              :aria-pressed="item.ID === sessionId"
              @click="openSession(item.ID)"
            >
              <strong>
                <el-icon v-if="sessionLoading && item.ID === sessionId" class="is-loading"><Loading /></el-icon>
                {{ item.title || '未命名会话' }}
              </strong>
              <small>{{ formatSessionTime(item.lastMessageAt) }}</small>
            </button>
            <el-tooltip content="删除会话" placement="top">
              <el-button
                class="session-delete"
                text
                :icon="Delete"
                :loading="deletingSessionId === item.ID"
                :aria-label="`删除会话：${item.title || '未命名会话'}`"
                @click="removeSession(item)"
              />
            </el-tooltip>
          </div>
        </div>

        <el-pagination
          v-if="sessionTotal > sessionPageSize"
          v-model:current-page="sessionPage"
          class="session-pagination"
          small
          :page-size="sessionPageSize"
          :pager-count="3"
          :total="sessionTotal"
          layout="prev, pager, next"
          @current-change="loadSessions"
        />

        <AppEmptyState
          v-if="!sessions.length && !loading"
          compact
          title="暂无会话"
          description="提交第一条业务查询后，会话会保存在这里。"
        />
      </aside>

      <section class="na-panel chat-panel" aria-labelledby="smart-copilot-title">
        <header class="chat-panel-header">
          <div class="chat-panel-title">
            <span class="chat-panel-icon" aria-hidden="true"><el-icon><ChatDotRound /></el-icon></span>
            <div>
              <h2>{{ sessionId ? '当前会话' : '新会话' }}</h2>
              <span>{{ sessionId ? '继续查看和追问业务数据' : '选择建议问题，或直接输入你的问题' }}</span>
            </div>
          </div>
          <span class="access-state"><i aria-hidden="true" />{{ tools.length }} 项数据能力可用</span>
        </header>

        <div ref="chatScroll" class="chat-scroll" aria-live="polite">
          <div v-if="!messages.length && !sessionLoading" class="empty-chat">
            <div class="empty-chat-icon" aria-hidden="true"><el-icon><MagicStick /></el-icon></div>
            <div class="empty-chat-copy">
              <strong>从业务数据开始提问</strong>
              <span>助手只会读取你当前有权限访问的信息，并在回答中附上相关记录。</span>
            </div>
            <div class="question-suggestions" aria-label="常用业务问题">
              <button v-for="item in suggestedQuestions" :key="item" type="button" @click="chooseQuestion(item)">
                <span>{{ item }}</span>
                <el-icon aria-hidden="true"><ArrowRight /></el-icon>
              </button>
            </div>
          </div>

          <article
            v-for="item in messages"
            :key="item.ID || item.clientId"
            class="message"
            :class="item.role === 'user' ? 'message--user' : 'message--assistant'"
          >
            <template v-if="item.role === 'user'">
              <div class="message-role">我</div>
              <div class="user-message-body">{{ item.content }}</div>
              <time v-if="messageTime(item)" class="message-time">{{ formatTime(messageTime(item)) }}</time>
            </template>

            <template v-else>
              <div class="assistant-avatar" aria-hidden="true"><el-icon><MagicStick /></el-icon></div>
              <div class="assistant-message-body">
                <header class="assistant-message-header">
                  <div class="assistant-identity">
                    <strong>业务助手</strong>
                    <span>只读业务查询</span>
                  </div>
                  <div class="assistant-meta">
                    <span
                      class="message-status"
                      :class="item.partial ? 'message-status--partial' : 'message-status--complete'"
                    >
                      <el-icon aria-hidden="true"><CircleCheck /></el-icon>
                      {{ item.partial ? '部分结果' : '已完成' }}
                    </span>
                    <time v-if="messageTime(item)" class="message-time">{{ formatTime(messageTime(item)) }}</time>
                  </div>
                </header>

                <div v-if="messageTools(item).length" class="tool-badges" aria-label="本次查询范围">
                  <span v-for="tool in messageTools(item)" :key="tool" class="tool-badge" :title="tool">
                    {{ toolLabel(tool) }}
                  </span>
                </div>

                <AssistantMarkdown :source="item.content" />

                <p v-if="item.partial" class="partial-note">
                  本次仅展示已完成查询的结果；未完成的查询原因已在回答中说明。
                </p>

                <details v-if="hasStructuredData(item)" class="message-result">
                  <summary>
                    <span class="result-title">查询结果</span>
                    <small>{{ resultDescriptor(item) }}</small>
                    <el-icon class="result-chevron" aria-hidden="true"><ArrowDown /></el-icon>
                  </summary>
                  <div class="result-content">
                    <div v-if="resultFacts(item).length" class="result-facts">
                      <div v-for="fact in resultFacts(item)" :key="fact.label" class="result-fact">
                        <span>{{ fact.label }}</span>
                        <strong>{{ fact.value }}</strong>
                      </div>
                    </div>
                    <div v-if="tableRows(item).length" class="result-table">
                      <el-table :data="tableRows(item)" size="small" max-height="300" table-layout="fixed">
                        <el-table-column
                          v-for="column in tableColumns(item)"
                          :key="column"
                          :prop="column"
                          :label="columnLabel(column)"
                          min-width="128"
                          show-overflow-tooltip
                        >
                          <template #default="{ row }">{{ displayColumnValue(column, row[column]) }}</template>
                        </el-table-column>
                      </el-table>
                    </div>
                  </div>
                </details>

                <footer v-if="citationList(item.citations).length" class="message-citations">
                  <div class="citations-label"><el-icon aria-hidden="true"><Link /></el-icon>相关记录</div>
                  <div class="citation-links">
                    <el-button
                      v-for="citation in citationList(item.citations)"
                      :key="`${citation.type}-${citation.id || citation.label}`"
                      link
                      type="primary"
                      @click="openCitation(citation)"
                    >
                      <span>{{ citation.label }}</span>
                      <el-icon aria-hidden="true"><ArrowRight /></el-icon>
                    </el-button>
                  </div>
                </footer>
              </div>
            </template>
          </article>

          <article v-if="sending" key="copilot-pending" class="message message--pending" aria-label="正在查询业务数据">
            <div class="assistant-avatar" aria-hidden="true"><el-icon><MagicStick /></el-icon></div>
            <div class="pending-copy">
              <span class="assistant-dots" aria-hidden="true"><i /><i /><i /></span>
              正在查询可访问的业务数据
            </div>
          </article>
        </div>

        <form class="composer" @submit.prevent="submitQuestion">
          <div class="composer-heading">
            <label class="composer-label" for="smart-copilot-question">向业务助手提问</label>
            <span>只读查询 · 数据范围跟随当前权限</span>
          </div>
          <div class="composer-control">
            <el-input
              id="smart-copilot-question"
              v-model="question"
              class="composer-input"
              type="textarea"
              :rows="2"
              maxlength="2000"
              show-word-limit
              resize="none"
              placeholder="例如：未来 30 天哪些资产质保到期？"
              :disabled="sending"
            />
            <el-tooltip content="发送查询" placement="top">
              <el-button
                class="composer-submit"
                type="primary"
                circle
                :icon="Position"
                :loading="sending"
                native-type="submit"
                aria-label="发送查询"
              />
            </el-tooltip>
          </div>
        </form>
      </section>

      <aside class="na-panel tools-panel" aria-label="可查询范围">
        <header class="panel-header">
          <div>
            <h2>可查询范围</h2>
            <span>{{ tools.length }} 项能力，已按当前角色过滤</span>
          </div>
        </header>
        <div v-for="item in tools" :key="item.name" class="tool-item">
          <el-icon aria-hidden="true"><CircleCheck /></el-icon>
          <div>
            <strong>{{ toolLabel(item.name) }}</strong>
            <span>{{ item.description }}</span>
          </div>
        </div>
        <AppEmptyState
          v-if="!tools.length && !loading"
          compact
          title="暂无可用查询"
          description="当前角色没有可调用的只读业务工具。"
        />
      </aside>
    </div>
  </main>
</template>

<script setup>
import { nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowDown, ArrowRight, ChatDotRound, CircleCheck, Delete, Link, Loading, MagicStick, Plus, Position, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import AssistantMarkdown from '@/plugin/smart/components/AssistantMarkdown.vue'
import { deleteCopilotSession, getCopilotSession, getCopilotSessions, getCopilotTools, queryCopilot } from '@/plugin/smart/api/smart'

defineOptions({ name: 'SmartCopilot' })

const router = useRouter()
const loading = ref(false)
const sending = ref(false)
const sessionLoading = ref(false)
const deletingSessionId = ref(0)
const sessions = ref([])
const sessionTotal = ref(0)
const sessionPage = ref(1)
const sessionPageSize = ref(10)
const messages = ref([])
const tools = ref([])
const sessionId = ref(0)
const question = ref('')
const chatScroll = ref(null)

const suggestedQuestions = [
  '未来 30 天哪些资产质保到期？',
  '当前有哪些高风险资产？',
  '待核对发票有多少张？',
  '今天有哪些日程和未读公告？'
]

const toolLabels = {
  'asset.search': '资产查询',
  'asset.detail': '资产详情',
  'asset.risk.list': '资产风险',
  'asset.warranty.expiring': '质保到期',
  'asset.custodian.summary': '保管人汇总',
  'asset.operation.summary': '资产流转',
  'invoice.summary': '发票汇总',
  'invoice.pending_reviews': '待复核发票',
  'invoice.failed_recognitions': '识别失败发票',
  'invoice.provider_quality': '识别质量',
  'schedule.today': '个人日程',
  'announcement.unread': '未读公告',
  'knowledge.search': '知识检索'
}

const columnLabels = {
  assetCode: '资产编号', AssetCode: '资产编号', name: '名称', Name: '名称', brand: '品牌', Brand: '品牌', model: '型号', Model: '型号',
  serialNumber: '序列号', SerialNumber: '序列号', status: '状态', Status: '状态', warrantyEndDate: '质保到期日', WarrantyEndDate: '质保到期日',
  custodian: '保管人', Custodian: '保管人', custodianName: '保管人', title: '标题', Title: '标题', severity: '等级', Severity: '等级',
  sellerName: '销售方', SellerName: '销售方', invoiceNumber: '发票号码', InvoiceNumber: '发票号码', totalCents: '含税金额', TotalCents: '含税金额',
  date: '日期', Date: '日期', time: '时间', Time: '时间', type: '类型', Type: '类型', note: '备注', Note: '备注'
}

const factLabels = {
  confirmedCount: '已确认发票', ConfirmedCount: '已确认发票',
  pendingCount: '待复核发票', PendingCount: '待复核发票',
  failedCount: '识别失败', FailedCount: '识别失败',
  unreadCount: '未读公告', UnreadCount: '未读公告',
  totalCount: '记录总数', TotalCount: '记录总数',
  total: '记录总数', count: '记录数量'
}

const countFactKeys = new Set(['confirmedCount', 'ConfirmedCount', 'pendingCount', 'PendingCount', 'failedCount', 'FailedCount'])
const currencyFactKeys = new Set(['totalCents', 'TotalCents'])

const hiddenColumns = new Set(['ID', 'id', 'CreatedAt', 'createdAt', 'UpdatedAt', 'updatedAt', 'DeletedAt', 'deletedAt', 'categoryId', 'CategoryID', 'userId', 'UserID'])
const preferredColumns = ['assetCode', 'AssetCode', 'name', 'Name', 'title', 'Title', 'status', 'Status', 'severity', 'Severity', 'warrantyEndDate', 'WarrantyEndDate', 'date', 'Date', 'time', 'Time', 'sellerName', 'SellerName', 'invoiceNumber', 'InvoiceNumber']

const toolLabel = (tool) => toolLabels[tool] || tool || '业务查询'
const messageTime = (item) => item.generatedAt || item.CreatedAt || item.createdAt
const messageTools = (item) => Array.isArray(item.tools) ? item.tools : String(item.tool || '').split(',').map((value) => value.trim()).filter(Boolean)
const citationList = (value) => Array.isArray(value) ? value : Object.values(value || {})
const createClientMessageID = () => globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`

function tableRows(item) {
  const data = item?.data
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.list)) return data.list
  return []
}

function tableColumns(item) {
  const first = tableRows(item)[0]
  if (!first || typeof first !== 'object') return []
  return Object.keys(first)
    .filter((key) => !hiddenColumns.has(key) && typeof first[key] !== 'object')
    .sort((left, right) => {
      const leftIndex = preferredColumns.indexOf(left)
      const rightIndex = preferredColumns.indexOf(right)
      return (leftIndex < 0 ? preferredColumns.length : leftIndex) - (rightIndex < 0 ? preferredColumns.length : rightIndex)
    })
    .slice(0, 5)
}

function columnLabel(key) {
  return columnLabels[key] || key
}

function displayValue(value) {
  if (value === undefined || value === null || value === '') return '—'
  if (typeof value === 'boolean') return value ? '是' : '否'
  if (typeof value === 'number') return new Intl.NumberFormat('zh-CN').format(value)
  return String(value)
}

function displayColumnValue(key, value) {
  if (currencyFactKeys.has(key) && Number.isFinite(Number(value))) {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY' }).format(Number(value) / 100)
  }
  return displayValue(value)
}

function factLabel(key) {
  return factLabels[key] || columnLabels[key] || toolLabels[key] || key
}

function displayFactValue(key, value) {
  if (currencyFactKeys.has(key) && Number.isFinite(Number(value))) {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY' }).format(Number(value) / 100)
  }
  if (countFactKeys.has(key) && Number.isFinite(Number(value))) return `${displayValue(value)} 张`
  return displayValue(value)
}

function resultFacts(item) {
  const data = item?.data
  if (!data || Array.isArray(data) || typeof data !== 'object') return []
  if (Array.isArray(data.list)) {
    const total = Number.isFinite(Number(data.total)) ? Number(data.total) : data.list.length
    return [{ label: '匹配记录', value: `${total} 条` }]
  }
  return Object.entries(data)
    .filter(([key]) => !['from', 'to', 'label', 'query'].includes(key))
    .slice(0, 4)
    .map(([key, value]) => {
      if (value?.error) return { label: toolLabel(key), value: '未完成' }
      if (Array.isArray(value)) return { label: toolLabel(key), value: `${value.length} 条` }
      if (value && typeof value === 'object') {
        if (Number.isFinite(Number(value.unreadCount))) return { label: toolLabel(key), value: `${value.unreadCount} 条未读` }
        if (Array.isArray(value.list)) return { label: toolLabel(key), value: `${value.list.length} 条` }
        if (Number.isFinite(Number(value.total))) return { label: toolLabel(key), value: `${value.total} 条` }
      }
      return { label: factLabel(key), value: displayFactValue(key, value) }
    })
}

function hasStructuredData(item) {
  return tableRows(item).length > 0 || resultFacts(item).length > 0
}

function resultDescriptor(item) {
  const rows = tableRows(item)
  if (rows.length) return `${rows.length} 条记录`
  const facts = resultFacts(item)
  return facts.length ? `${facts.length} 项汇总` : '查看明细'
}

function formatTime(value) {
  const date = value ? new Date(value) : null
  return date && !Number.isNaN(date.getTime()) ? date.toLocaleString('zh-CN', { hour12: false }) : '—'
}

function formatSessionTime(value) {
  const date = value ? new Date(value) : null
  if (!date || Number.isNaN(date.getTime())) return '暂无消息'
  const today = new Date()
  if (date.toDateString() === today.toDateString()) return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}

async function scrollToLatest() {
  await nextTick()
  const target = chatScroll.value
  if (!target) return
  const reduceMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
  target.scrollTo({ top: target.scrollHeight, behavior: reduceMotion ? 'auto' : 'smooth' })
}

function openCitation(item) {
  const routeName = {
    asset: 'assetInventory',
    risk: 'assetRiskCenter',
    asset_operation: 'assetInbound',
    invoice: 'invoiceLedger',
    invoice_quality: 'invoiceQuality',
    schedule: 'workSchedule',
    announcement: 'anInfo'
  }[item.type]
  const query = Object.fromEntries(new URLSearchParams(item.params || ''))
  if (routeName && router.hasRoute(routeName)) router.push({ name: routeName, query })
  else if (item.path) router.push({ path: item.path, query })
}

async function loadSessions() {
  loading.value = true
  try {
    const [sessionRes, toolRes] = await Promise.all([getCopilotSessions({ paged: true, page: sessionPage.value, pageSize: sessionPageSize.value }), getCopilotTools()])
    if (sessionRes.code === 0) {
      sessions.value = sessionRes.data?.list || []
      sessionTotal.value = Number(sessionRes.data?.total || 0)
    }
    else ElMessage.error(sessionRes.msg || '读取会话失败')
    if (toolRes.code === 0) tools.value = toolRes.data || []
    else ElMessage.error(toolRes.msg || '读取可查询范围失败')
  } catch (error) {
    ElMessage.error('读取业务助手信息失败')
  } finally {
    loading.value = false
  }
}

async function openSession(id) {
  sessionId.value = id
  sessionLoading.value = true
  try {
    const res = await getCopilotSession({ id })
    if (res.code === 0) {
      messages.value = res.data?.messages || []
      await scrollToLatest()
    } else {
      ElMessage.error(res.msg || '读取会话失败')
    }
  } catch (error) {
    ElMessage.error('读取会话失败')
  } finally {
    sessionLoading.value = false
  }
}

function newSession() {
  sessionId.value = 0
  messages.value = []
}

function chooseQuestion(value) {
  question.value = value
}

async function removeSession(item) {
  try {
    await ElMessageBox.confirm(`确认删除会话“${item.title || '未命名会话'}”？`, '删除会话', { type: 'warning' })
    deletingSessionId.value = item.ID
    const res = await deleteCopilotSession({ id: item.ID })
    if (res.code !== 0) {
      ElMessage.error(res.msg || '删除失败')
      return
    }
    if (sessionId.value === item.ID) newSession()
    if (sessions.value.length === 1 && sessionPage.value > 1) sessionPage.value--
    await loadSessions()
    ElMessage.success(res.msg || '会话已删除')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error('删除失败')
  } finally {
    deletingSessionId.value = 0
  }
}

async function submitQuestion() {
  const content = question.value.trim()
  if (!content || sending.value) return
  sending.value = true
  try {
    const res = await queryCopilot({ question: content, sessionId: sessionId.value || undefined })
    if (res.code !== 0) {
      ElMessage.error(res.msg || '查询失败')
      return
    }
    const messageID = createClientMessageID()
    messages.value.push(
      { clientId: `user-${messageID}`, role: 'user', content, createdAt: new Date().toISOString() },
      {
        clientId: `assistant-${messageID}`,
        role: 'assistant',
        content: res.data.answer,
        tool: res.data.tool,
        tools: res.data.tools,
        partial: res.data.partial,
        citations: res.data.citations,
        data: res.data.data,
        generatedAt: res.data.generatedAt
      }
    )
    sessionId.value = res.data.sessionId
    question.value = ''
    sessionPage.value = 1
    await Promise.all([loadSessions(), scrollToLatest()])
  } catch (error) {
    ElMessage.error('查询失败，请稍后重试')
  } finally {
    sending.value = false
  }
}

onMounted(loadSessions)
</script>

<style scoped lang="scss">
.copilot-layout {
  display: grid;
  grid-template-columns: 232px minmax(0, 1fr) 264px;
  align-items: start;
  gap: 12px;
}

.na-panel {
  min-width: 0;
  border: 1px solid var(--na-border);
  border-radius: 12px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
}

.session-panel,
.tools-panel {
  position: sticky;
  top: var(--na-space-sm);
  max-height: calc(100dvh - 170px);
  overflow-y: auto;
  padding: var(--na-space-md);
  background: color-mix(in srgb, var(--na-card) 96%, var(--na-muted));
}

.panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--na-space-xs);
  margin-bottom: var(--na-space-sm);
}

.panel-header h2 { margin: 0; color: var(--na-foreground); font-size: .875rem; font-weight: 700; line-height: 1.4; }
.panel-header span { display: block; margin-top: 3px; color: var(--na-muted-foreground); font-size: .75rem; line-height: 1.4; }

.session-list { display: grid; gap: 4px; }
.session-row {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 30px;
  align-items: center;
  border-radius: 8px;
  transition: background-color 180ms cubic-bezier(.22, 1, .36, 1);
}

.session-row:hover,
.session-row.is-active { background: var(--na-primary-soft); }
.session-row.is-active::before {
  position: absolute;
  top: 9px;
  bottom: 9px;
  left: 0;
  width: 3px;
  border-radius: 3px;
  background: var(--na-primary);
  content: '';
}

.session-item {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 5px;
  padding: 10px 8px 10px 12px;
  border: 0;
  border-radius: 7px;
  background: transparent;
  color: var(--na-foreground);
  font: inherit;
  text-align: left;
  cursor: pointer;
}

.session-item:focus-visible { outline: 3px solid var(--na-ring); outline-offset: -2px; }
.session-item strong { overflow: hidden; font-size: .8125rem; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.session-item small { color: var(--na-muted-foreground); font-size: .6875rem; }
.session-delete { width: 28px; height: 28px; margin: 0; color: var(--na-muted-foreground); opacity: 0; transition: opacity 160ms ease, color 160ms ease; }
.session-row:hover .session-delete,
.session-row:focus-within .session-delete,
.session-row.is-active .session-delete { opacity: 1; }
.session-delete:hover { color: var(--na-danger); }
.session-pagination { justify-content: center; margin-top: var(--na-space-sm); }

.chat-panel {
  display: flex;
  min-height: max(630px, calc(100dvh - 170px));
  flex-direction: column;
  overflow: hidden;
}

.chat-panel-header {
  display: flex;
  min-height: 64px;
  align-items: center;
  justify-content: space-between;
  gap: var(--na-space-md);
  padding: 10px var(--na-space-lg);
  border-bottom: 1px solid var(--na-border);
  background: var(--na-card);
}

.chat-panel-title { display: flex; min-width: 0; align-items: center; gap: 10px; }
.chat-panel-title > div { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.chat-panel-title h2 { margin: 0; color: var(--na-foreground); font-size: .875rem; font-weight: 700; line-height: 1.4; }
.chat-panel-title span { overflow: hidden; color: var(--na-muted-foreground); font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.chat-panel-icon {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  place-items: center;
  border-radius: 8px;
  background: var(--na-primary-soft);
  color: var(--na-primary);
  font-size: 17px;
}
.access-state { display: inline-flex; flex: 0 0 auto; align-items: center; gap: 6px; color: var(--na-muted-foreground); font-size: .75rem; white-space: nowrap; }
.access-state i { width: 7px; height: 7px; border-radius: 50%; background: var(--na-success); box-shadow: 0 0 0 3px var(--na-success-soft); }

.chat-scroll {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 28px clamp(18px, 3vw, 44px);
  background: color-mix(in srgb, var(--na-card) 97%, var(--na-muted));
  scroll-behavior: smooth;
}

.message { min-width: 0; margin: 0 auto 26px; }
.message--user { display: flex; width: min(100%, 960px); align-items: flex-end; flex-direction: column; }
.message-role { margin: 0 4px 5px 0; color: var(--na-muted-foreground); font-size: .75rem; }

.user-message-body {
  max-width: min(72%, 680px);
  padding: 10px 14px;
  border: 1px solid color-mix(in srgb, var(--na-primary) 16%, var(--na-border));
  border-radius: 10px 10px 3px 10px;
  background: var(--na-primary-soft);
  color: var(--na-foreground);
  font-size: .875rem;
  line-height: 1.65;
  overflow-wrap: anywhere;
}

.message-time { color: var(--na-muted-foreground); font-size: .6875rem; font-variant-numeric: tabular-nums; }
.message--user .message-time { margin: 5px 3px 0 0; }

.message--assistant,
.message--pending { display: flex; width: min(100%, 960px); align-items: flex-start; gap: 12px; }

.assistant-avatar {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--na-primary) 22%, var(--na-border));
  border-radius: 8px;
  background: var(--na-primary-soft);
  color: var(--na-primary);
  font-size: 17px;
}

.assistant-message-body {
  min-width: 0;
  flex: 1;
  padding: 2px 0 0;
}

.assistant-message-header,
.assistant-meta,
.tool-badges,
.message-citations,
.composer { display: flex; align-items: center; }

.assistant-message-header { min-height: 34px; justify-content: space-between; gap: var(--na-space-sm); margin-bottom: 10px; }
.assistant-identity { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.assistant-identity strong { color: var(--na-foreground); font-size: .8125rem; font-weight: 650; }
.assistant-identity span { color: var(--na-muted-foreground); font-size: .6875rem; }
.assistant-meta { flex: 0 0 auto; flex-wrap: wrap; justify-content: flex-end; gap: var(--na-space-xs); }

.message-status {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-height: 22px;
  padding: 0 6px;
  border-radius: 5px;
  font-size: .6875rem;
  font-weight: 600;
  white-space: nowrap;
}

.message-status--complete { background: var(--na-success-soft); color: var(--na-action-success); }
.message-status--partial { background: var(--na-warning-soft); color: var(--na-action-warning); }

.tool-badges { flex-wrap: wrap; gap: 5px; margin: 0 0 10px; }
.tool-badge {
  display: inline-flex;
  min-height: 21px;
  align-items: center;
  padding: 0 6px;
  border: 1px solid var(--na-border);
  border-radius: 5px;
  background: var(--na-card);
  color: var(--na-muted-foreground);
  font-size: .6875rem;
  line-height: 1;
}

.partial-note {
  margin: var(--na-space-sm) 0 0;
  padding: var(--na-space-xs) 0 0;
  border-top: 1px solid var(--na-border);
  color: var(--na-action-warning);
  font-size: .75rem;
  line-height: 1.55;
}

.message-result { margin-top: var(--na-space-md); border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-card); overflow: hidden; }
.message-result summary {
  display: flex;
  min-height: 40px;
  align-items: center;
  justify-content: space-between;
  gap: var(--na-space-xs);
  padding: 0 12px;
  color: var(--na-foreground);
  font-size: .75rem;
  font-weight: 600;
  cursor: pointer;
  list-style: none;
}

.message-result summary::-webkit-details-marker { display: none; }
.result-title { display: inline-flex; align-items: center; gap: 6px; }
.result-title::before { width: 3px; height: 14px; border-radius: 2px; background: var(--na-primary); content: ''; }
.result-chevron { margin-left: 3px; color: var(--na-muted-foreground); transition: transform 180ms ease; }
.message-result[open] .result-chevron { transform: rotate(180deg); }
.message-result summary small { margin-left: auto; color: var(--na-muted-foreground); font-size: .6875rem; font-weight: 400; }
.message-result summary:focus-visible { outline: 3px solid var(--na-ring); outline-offset: -2px; }

.result-content { padding: 0 12px 12px; border-top: 1px solid var(--na-border); }
.result-facts { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); }
.result-fact { display: flex; min-width: 0; flex-direction: column; gap: 5px; padding: 12px; border-right: 1px solid var(--na-border); }
.result-fact:first-child { padding-left: 0; }
.result-fact:last-child { padding-right: 0; border-right: 0; }
.result-fact span { color: var(--na-muted-foreground); font-size: .6875rem; }
.result-fact strong { color: var(--na-foreground); font-size: .9375rem; font-weight: 700; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
.result-table { overflow: hidden; margin-top: var(--na-space-sm); border: 1px solid var(--na-border); border-radius: var(--na-radius-sm); }

.message-citations { align-items: flex-start; gap: 10px; margin-top: var(--na-space-md); padding-top: var(--na-space-sm); border-top: 1px solid var(--na-border); }
.citations-label { display: inline-flex; min-height: 24px; flex: 0 0 auto; align-items: center; gap: 5px; color: var(--na-muted-foreground); font-size: .6875rem; }
.citation-links { display: flex; min-width: 0; flex: 1; flex-wrap: wrap; gap: 4px 16px; }
.message-citations .el-button { display: inline-flex; max-width: 100%; min-height: 24px; align-items: center; gap: 5px; margin: 0; font-size: .75rem; white-space: normal; text-align: left; }
.message-citations .el-button span { overflow-wrap: anywhere; }
.message-citations .el-button .el-icon { flex: 0 0 auto; }

.message--pending { animation: message-enter 180ms cubic-bezier(.22, 1, .36, 1); }
.pending-copy { display: inline-flex; min-height: 34px; align-items: center; gap: var(--na-space-xs); color: var(--na-muted-foreground); font-size: .8125rem; }
.assistant-dots { display: inline-flex; align-items: center; gap: 3px; }
.assistant-dots i { display: block; width: 5px; height: 5px; border-radius: 50%; background: var(--na-primary); animation: assistant-dot 900ms ease-in-out infinite; }
.assistant-dots i:nth-child(2) { animation-delay: 120ms; }
.assistant-dots i:nth-child(3) { animation-delay: 240ms; }

.empty-chat {
  display: grid;
  width: min(100%, 760px);
  min-height: 460px;
  grid-template-columns: 42px minmax(0, 1fr);
  align-content: center;
  gap: 6px var(--na-space-sm);
  margin: 0 auto;
  padding: var(--na-space-xl) 0;
}

.empty-chat-icon { display: grid; width: 42px; height: 42px; grid-row: 1 / span 2; place-items: center; border-radius: 8px; background: var(--na-primary-soft); color: var(--na-primary); font-size: 20px; }
.empty-chat-copy { display: flex; min-width: 0; flex-direction: column; gap: 4px; }
.empty-chat-copy strong { color: var(--na-foreground); font-size: .9375rem; }
.empty-chat-copy span { max-width: 60ch; color: var(--na-muted-foreground); font-size: .8125rem; line-height: 1.55; }

.question-suggestions { display: grid; grid-column: 1 / -1; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--na-space-xs); margin-top: var(--na-space-lg); }
.question-suggestions button {
  display: flex;
  min-width: 0;
  min-height: 52px;
  align-items: center;
  justify-content: space-between;
  gap: var(--na-space-xs);
  padding: 10px 12px;
  border: 1px solid var(--na-border);
  border-radius: var(--na-radius-sm);
  background: var(--na-card);
  color: var(--na-foreground);
  font: inherit;
  font-size: .8125rem;
  line-height: 1.45;
  text-align: left;
  cursor: pointer;
  transition: border-color 180ms ease, background-color 180ms ease, color 180ms ease;
}

.question-suggestions button:hover,
.question-suggestions button:focus-visible { border-color: var(--na-primary); background: var(--na-primary-soft); outline: none; }
.question-suggestions button span { min-width: 0; overflow-wrap: anywhere; }
.question-suggestions button .el-icon { flex: 0 0 auto; color: var(--na-primary); }

.composer {
  display: block;
  padding: 12px var(--na-space-lg) 14px;
  border-top: 1px solid var(--na-border);
  background: var(--na-card);
}

.composer-heading { display: flex; align-items: center; justify-content: space-between; gap: var(--na-space-sm); margin-bottom: 7px; }
.composer-label { color: var(--na-foreground); font-size: .75rem; font-weight: 700; white-space: nowrap; }
.composer-heading span { color: var(--na-muted-foreground); font-size: .6875rem; }
.composer-control { display: flex; align-items: flex-end; gap: 10px; }
.composer-input { flex: 1; min-width: 0; }
.composer-input :deep(.el-textarea__inner) { min-height: 58px !important; padding: 10px 58px 10px 12px; border-radius: 8px; box-shadow: 0 0 0 1px var(--na-border) inset; background: color-mix(in srgb, var(--na-card) 96%, var(--na-muted)); line-height: 1.55; }
.composer-input :deep(.el-textarea__inner:focus) { box-shadow: 0 0 0 1px var(--na-primary) inset, 0 0 0 3px var(--na-ring); }
.composer-input :deep(.el-input__count) { right: 10px; bottom: 7px; background: transparent; color: var(--na-muted-foreground); }
.composer-submit { width: 42px; min-width: 42px; height: 42px; margin-bottom: 8px; box-shadow: 0 5px 14px color-mix(in srgb, var(--na-primary) 24%, transparent); }

.tool-item { display: flex; gap: 9px; padding: 9px 0; border-bottom: 1px solid var(--na-border); }
.tool-item:last-child { border-bottom: 0; }
.tool-item > .el-icon { flex: 0 0 auto; margin-top: 2px; color: var(--na-success); font-size: 14px; }
.tool-item div { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.tool-item strong { color: var(--na-foreground); font-size: .75rem; font-weight: 650; }
.tool-item span { color: var(--na-muted-foreground); font-size: .6875rem; line-height: 1.45; }

@keyframes message-enter { from { opacity: 0; transform: translateY(4px); } to { opacity: 1; transform: translateY(0); } }
@keyframes assistant-dot { 0%, 80%, 100% { opacity: .35; transform: translateY(0); } 40% { opacity: 1; transform: translateY(-2px); } }

@media (max-width: 1100px) {
  .copilot-layout { grid-template-columns: 212px minmax(0, 1fr); }
  .tools-panel { display: none; }
}

@media (max-width: 760px) {
  .copilot-layout { display: flex; min-height: 0; flex-direction: column; }
  .session-panel { width: 100%; max-height: 210px; }
  .chat-panel { width: 100%; min-height: calc(100dvh - 330px); }
  .empty-chat { min-height: 390px; padding: var(--na-space-lg) 0; }
}

@media (max-width: 620px) {
  .chat-scroll { padding: var(--na-space-md); }
  .chat-panel-header { padding: 9px var(--na-space-md); }
  .access-state { display: none; }
  .user-message-body { max-width: 92%; }
  .message--assistant,
  .message--pending { gap: var(--na-space-xs); }
  .assistant-avatar { width: 30px; height: 30px; flex-basis: 30px; font-size: 15px; }
  .assistant-message-body { padding: 12px; }
  .assistant-message-header { align-items: flex-start; flex-direction: column; }
  .assistant-meta { width: 100%; justify-content: space-between; }
  .question-suggestions { grid-template-columns: 1fr; }
  .composer { padding: var(--na-space-sm) var(--na-space-md) var(--na-space-md); }
  .composer-heading span { display: none; }
  .result-facts { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .result-fact:nth-child(2) { border-right: 0; }
  .result-fact:nth-child(n + 3) { border-top: 1px solid var(--na-border); }
  .message-citations { flex-direction: column; gap: 4px; }
}

@media (prefers-reduced-motion: reduce) {
  .chat-scroll { scroll-behavior: auto; }
  .session-row,
  .question-suggestions button,
  .result-chevron,
  .message--pending,
  .assistant-dots i { transition: none; animation: none; }
}
</style>
