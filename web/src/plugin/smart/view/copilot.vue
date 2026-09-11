<template>
  <main class="na-page na-page--list smart-copilot-page">
    <AppPageHeader
      title-id="smart-copilot-title"
      title="业务助手"
      description="企业级多维业务数据只读分析助手，穿透资产全生命周期、财务发票与协同办公；每条回答附带审计证据与事实单据。"
    >
      <template #actions>
        <div class="header-status-badge" aria-label="AI 业务助手运行状态">
          <span class="pulse-dot" aria-hidden="true" />
          <span class="status-text">只读安全穿透已启用</span>
        </div>
        <el-tooltip content="刷新会话与可查询范围" placement="bottom">
          <el-button :icon="Refresh" :loading="loading" aria-label="刷新会话与可查询范围" @click="loadSessions">
            刷新数据
          </el-button>
        </el-tooltip>
      </template>
    </AppPageHeader>

    <div class="copilot-layout">
      <!-- 1. 左侧：会话中心 (Session Hub) -->
      <aside class="na-panel session-panel" aria-label="业务助手会话">
        <header class="panel-header">
          <div class="panel-header__meta">
            <h2>会话记录</h2>
            <span class="session-counter">共 {{ sessionTotal }} 个会话</span>
          </div>
          <el-button
            type="primary"
            size="small"
            class="btn-new-session"
            :icon="Plus"
            @click="newSession"
          >
            新对话
          </el-button>
        </header>

        <!-- 历史会话过滤检索 -->
        <div class="session-search">
          <el-input
            v-model="sessionSearchQuery"
            placeholder="搜索历史会话..."
            size="small"
            clearable
            :prefix-icon="Search"
            aria-label="搜索历史会话"
          />
        </div>

        <div class="session-list">
          <div
            v-for="item in filteredSessions"
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
              <div class="session-item__header">
                <el-icon class="session-item__icon" aria-hidden="true">
                  <Loading v-if="sessionLoading && item.ID === sessionId" class="is-loading" />
                  <ChatDotRound v-else />
                </el-icon>
                <strong :title="item.title || '未命名会话'">
                  {{ item.title || '未命名会话' }}
                </strong>
              </div>
              <div class="session-item__footer">
                <small>{{ formatSessionTime(item.lastMessageAt) }}</small>
                <span v-if="item.ID === sessionId" class="active-badge">当前对话</span>
              </div>
            </button>

            <el-tooltip content="删除会话" placement="top">
              <el-button
                class="session-delete"
                text
                :icon="Delete"
                :loading="deletingSessionId === item.ID"
                :aria-label="`删除会话：${item.title || '未命名会话'}`"
                @click.stop="removeSession(item)"
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
          description="提交第一条业务查询后，会话将自动保存在这里。"
        />
        <div v-else-if="!filteredSessions.length && sessionSearchQuery" class="session-empty-filtered">
          <el-icon><Search /></el-icon>
          <span>未找到包含“{{ sessionSearchQuery }}”的会话</span>
        </div>
      </aside>

      <!-- 2. 中间：智能工作台 (Chat Studio) -->
      <section class="na-panel chat-panel" aria-labelledby="smart-copilot-title">
        <header class="chat-panel-header">
          <div class="chat-panel-title">
            <div class="chat-panel-avatar" aria-hidden="true">
              <el-icon><Cpu /></el-icon>
            </div>
            <div class="chat-panel-heading">
              <div class="chat-panel-title-row">
                <h2>{{ currentSessionTitle }}</h2>
                <el-tag v-if="sessionId" size="small" effect="plain" type="primary" class="session-tag">
                  ID: #{{ sessionId }}
                </el-tag>
              </div>
              <span>{{ sessionId ? '基于当前会话历史上下文持续提问，支持下钻与穿透' : '选择场景化推荐问题，或直接在下方输入您的业务疑问' }}</span>
            </div>
          </div>
          <div class="chat-panel-actions">
            <span class="access-state">
              <i class="state-dot" aria-hidden="true" />
              <span>{{ tools.length }} 项数据能力可用</span>
            </span>
            <el-button
              v-if="messages.length"
              size="small"
              text
              :icon="Plus"
              class="btn-reset-chat"
              @click="newSession"
            >
              开启新会话
            </el-button>
          </div>
        </header>

        <div ref="chatScroll" class="chat-scroll" aria-live="polite">
          <!-- 2.1 无对话时的空状态与场景引导 -->
          <div v-if="!messages.length && !sessionLoading" class="empty-chat">
            <!-- 头部 AI 形象与信任标识 -->
            <div class="hero-header">
              <div class="hero-badge-wrap">
                <div class="hero-badge-aura" aria-hidden="true" />
                <div class="hero-badge" aria-hidden="true">
                  <svg class="hero-robot-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M12 8V4H8" />
                    <rect x="4" y="8" width="16" height="12" rx="2" />
                    <path d="M2 14h2M20 14h2M9 13v2M15 13v2" />
                  </svg>
                </div>
              </div>

              <div class="hero-copy">
                <h3>智能业务助手 · Copilot</h3>
                <p>实时穿透资产台账、流转单据、发票状态、工作日程与制度文档，每一条回答均提供可核验的事实源证据。</p>
              </div>

              <div class="hero-trust-badges">
                <div class="trust-pill">
                  <el-icon><Lock /></el-icon>
                  <span>严格角色权限隔离</span>
                </div>
                <div class="trust-pill">
                  <el-icon><Cpu /></el-icon>
                  <span>毫秒级全域只读穿透</span>
                </div>
                <div class="trust-pill">
                  <el-icon><Finished /></el-icon>
                  <span>回答全链路事实留痕</span>
                </div>
              </div>
            </div>

            <!-- 场景化推荐问答卡片 -->
            <div class="scenario-section">
              <div class="scenario-title">
                <span>推荐业务探索场景</span>
                <small>点击卡片即可一键填入查询</small>
              </div>
              <div class="scenario-grid">
                <button
                  v-for="card in scenarioCards"
                  :key="card.title"
                  type="button"
                  class="scenario-card"
                  @click="chooseQuestion(card.title)"
                >
                  <div class="scenario-card__icon" :style="{ color: card.color, background: `${card.color}15` }">
                    <component :is="card.icon" />
                  </div>
                  <div class="scenario-card__content">
                    <div class="scenario-card__category">
                      <el-tag size="small" :type="card.tagType" effect="light">{{ card.category }}</el-tag>
                    </div>
                    <strong class="scenario-card__title">{{ card.title }}</strong>
                    <span class="scenario-card__desc">{{ card.desc }}</span>
                  </div>
                  <el-icon class="scenario-card__arrow" aria-hidden="true"><ArrowRight /></el-icon>
                </button>
              </div>
            </div>
          </div>

          <!-- 2.2 消息对话流 -->
          <article
            v-for="item in messages"
            :key="item.ID || item.clientId"
            class="message"
            :class="item.role === 'user' ? 'message--user' : 'message--assistant'"
          >
            <!-- 用户发问气泡 -->
            <template v-if="item.role === 'user'">
              <div class="user-message-container">
                <div class="user-message-meta">
                  <span class="user-role-label">我</span>
                  <time v-if="messageTime(item)" class="message-time">{{ formatTime(messageTime(item)) }}</time>
                </div>
                <div class="user-message-bubble">
                  {{ item.content }}
                </div>
              </div>
              <div class="user-avatar" aria-hidden="true">
                {{ userInitials }}
              </div>
            </template>

            <!-- 业务助手回答气泡 -->
            <template v-else>
              <div class="assistant-avatar" aria-hidden="true">
                <svg class="assistant-robot" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 8V4H8" />
                  <rect x="4" y="8" width="16" height="12" rx="2" />
                  <path d="M2 14h2M20 14h2M9 13v2M15 13v2" />
                </svg>
              </div>
              <div class="assistant-message-body">
                <header class="assistant-message-header">
                  <div class="assistant-identity">
                    <div class="identity-name">
                      <strong>业务助手</strong>
                      <span class="copilot-pill">Copilot</span>
                    </div>
                    <span class="identity-role">安全只读业务引擎</span>
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

                <!-- 调用的数据工具徽章 -->
                <div v-if="messageTools(item).length" class="tool-badges" aria-label="本次查询范围">
                  <span class="tool-badges-title"><el-icon><Cpu /></el-icon> 穿透能力：</span>
                  <span v-for="tool in messageTools(item)" :key="tool" class="tool-badge" :title="tool">
                    {{ toolLabel(tool) }}
                  </span>
                </div>

                <!-- 回答主体 Markdown -->
                <div class="assistant-markdown-container">
                  <AssistantMarkdown :source="item.content" />
                </div>

                <p v-if="item.partial" class="partial-note">
                  本次仅展示已完成查询的结果；未完成的查询原因已在回答中说明。
                </p>

                <!-- 结构化数据结果面板 (KPI + 数据表格) -->
                <details v-if="hasStructuredData(item)" class="message-result" open>
                  <summary>
                    <span class="result-title">
                      <el-icon><Tickets /></el-icon>
                      <span>结构化事实数据</span>
                    </span>
                    <small>{{ resultDescriptor(item) }}</small>
                    <el-icon class="result-chevron" aria-hidden="true"><ArrowDown /></el-icon>
                  </summary>
                  <div class="result-content">
                    <div v-if="resultFacts(item).length" class="result-facts">
                      <div v-for="fact in resultFacts(item)" :key="fact.label" class="result-fact">
                        <span class="fact-label">{{ fact.label }}</span>
                        <strong class="fact-value">{{ fact.value }}</strong>
                      </div>
                    </div>
                    <div v-if="tableRows(item).length" class="result-table">
                      <el-table :data="tableRows(item)" size="small" max-height="300" stripe table-layout="fixed">
                        <el-table-column
                          v-for="column in tableColumns(item)"
                          :key="column"
                          :prop="column"
                          :label="columnLabel(column)"
                          min-width="128"
                          show-overflow-tooltip
                        >
                          <template #default="{ row }">
                            <span class="cell-content">{{ displayColumnValue(column, row[column]) }}</span>
                          </template>
                        </el-table-column>
                      </el-table>
                    </div>
                  </div>
                </details>

                <!-- 证据溯源单据链接 -->
                <footer v-if="citationList(item.citations).length" class="message-citations">
                  <div class="citations-label">
                    <el-icon aria-hidden="true"><Link /></el-icon>
                    <span>关联事实记录：</span>
                  </div>
                  <div class="citation-links">
                    <el-button
                      v-for="citation in citationList(item.citations)"
                      :key="`${citation.type}-${citation.id || citation.label}`"
                      class="citation-btn"
                      size="small"
                      type="primary"
                      plain
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

          <!-- 思考中动态卡片 -->
          <article v-if="sending" key="copilot-pending" class="message message--pending" aria-label="业务助手正在查询业务数据">
            <div class="assistant-avatar assistant-avatar--thinking" aria-hidden="true">
              <svg class="assistant-robot" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 8V4H8" />
                <rect x="4" y="8" width="16" height="12" rx="2" />
                <path d="M2 14h2M20 14h2M9 13v2M15 13v2" />
              </svg>
            </div>
            <div class="pending-container">
              <div class="pending-header">
                <strong>业务助手思考中</strong>
                <span class="assistant-dots" aria-hidden="true"><i /><i /><i /></span>
              </div>
              <span class="pending-tip">正在实时调取权限内资产、财务与协同业务数据库...</span>
            </div>
          </article>
        </div>

        <!-- 2.3 底部交互式输入台 (Composer) -->
        <div class="composer-container">
          <!-- 快捷提示胶囊 -->
          <div class="quick-prompts-bar">
            <span class="quick-prompts-label">快捷问询：</span>
            <div class="quick-prompts-scroll">
              <button
                v-for="prompt in quickPrompts"
                :key="prompt.label"
                type="button"
                class="quick-prompt-pill"
                @click="chooseQuestion(prompt.query)"
              >
                {{ prompt.label }}
              </button>
            </div>
          </div>

          <form class="composer" @submit.prevent="submitQuestion">
            <div class="composer-box">
              <el-input
                id="smart-copilot-question"
                ref="questionInputRef"
                v-model="question"
                class="composer-input"
                type="textarea"
                :rows="2"
                maxlength="2000"
                show-word-limit
                resize="none"
                placeholder="向业务助手提问，例如：未来 30 天哪些资产质保到期？（按 Enter 发送，Shift + Enter 换行）"
                :disabled="sending"
                @keydown="handleKeydown"
              />
              <el-tooltip content="发送查询 (Enter)" placement="top">
                <el-button
                  class="composer-submit"
                  type="primary"
                  :icon="Position"
                  :loading="sending"
                  :disabled="!question.trim() || sending"
                  native-type="submit"
                  aria-label="发送查询"
                />
              </el-tooltip>
            </div>
            <div class="composer-footer">
              <div class="composer-guard">
                <el-icon><Lock /></el-icon>
                <span>只读安全穿透 · 数据严格限定于当前登录账号权限范围</span>
              </div>
              <span class="composer-hint">Enter 发送 / Shift + Enter 换行</span>
            </div>
          </form>
        </div>
      </section>

      <!-- 3. 右侧：可查询能力矩阵 (Capabilities Directory) -->
      <aside class="na-panel tools-panel" aria-label="可查询范围">
        <header class="panel-header">
          <div class="panel-header__meta">
            <h2>可查询范围</h2>
            <span class="tools-counter">{{ tools.length }} 项数据能力已就绪</span>
          </div>
        </header>

        <!-- 分类切换 Tabs -->
        <div class="tool-category-tabs">
          <button
            v-for="cat in toolCategories"
            :key="cat.key"
            type="button"
            class="category-tab-btn"
            :class="{ 'is-active': activeToolCategory === cat.key }"
            @click="activeToolCategory = cat.key"
          >
            {{ cat.label }}
            <span class="tab-count">{{ getCategoryCount(cat) }}</span>
          </button>
        </div>

        <div class="tools-list">
          <div
            v-for="item in filteredTools"
            :key="item.name"
            class="tool-card"
            @click="chooseQuestion(getToolSampleQuery(item.name))"
          >
            <div class="tool-card__icon" :style="{ color: getToolColor(item.name), background: `${getToolColor(item.name)}15` }">
              <component :is="getToolIcon(item.name)" />
            </div>
            <div class="tool-card__info">
              <div class="tool-card__header">
                <strong>{{ toolLabel(item.name) }}</strong>
                <el-icon class="tool-card__prompt-hint" title="点击填入此能力测试问询"><ArrowRight /></el-icon>
              </div>
              <p>{{ item.description }}</p>
            </div>
          </div>
        </div>

        <AppEmptyState
          v-if="!tools.length && !loading"
          compact
          title="暂无可用查询"
          description="当前角色暂无可调用的只读业务工具。"
        />
        <div v-else-if="!filteredTools.length" class="tool-empty-filtered">
          <span>该分类下暂无已授权数据能力</span>
        </div>
      </aside>
    </div>
  </main>
</template>

<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  ArrowDown,
  ArrowRight,
  Box,
  Calendar,
  ChatDotRound,
  CircleCheck,
  Cpu,
  Delete,
  Finished,
  Link,
  Loading,
  Lock,
  Plus,
  Position,
  Refresh,
  Search,
  Tickets,
  WarningFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import AssistantMarkdown from '@/plugin/smart/components/AssistantMarkdown.vue'
import { deleteCopilotSession, getCopilotSession, getCopilotSessions, getCopilotTools, queryCopilot } from '@/plugin/smart/api/smart'
import { useUserStore } from '@/pinia/modules/user'

defineOptions({ name: 'SmartCopilot' })

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const sending = ref(false)
const sessionLoading = ref(false)
const deletingSessionId = ref(0)
const sessions = ref([])
const sessionTotal = ref(0)
const sessionPage = ref(1)
const sessionPageSize = ref(10)
const sessionSearchQuery = ref('')
const messages = ref([])
const tools = ref([])
const sessionId = ref(0)
const question = ref('')
const chatScroll = ref(null)
const questionInputRef = ref(null)
const activeToolCategory = ref('all')

const userInitials = computed(() => {
  const name = userStore.userInfo?.nickName || userStore.userInfo?.userName || '我'
  return String(name).trim().slice(0, 2)
})

const currentSessionTitle = computed(() => {
  if (!sessionId.value) return '新对话 · 智能问询'
  const found = sessions.value.find((s) => s.ID === sessionId.value)
  return found?.title || '当前会话'
})

const filteredSessions = computed(() => {
  if (!sessionSearchQuery.value.trim()) return sessions.value
  const q = sessionSearchQuery.value.trim().toLowerCase()
  return sessions.value.filter((item) => (item.title || '').toLowerCase().includes(q))
})

const scenarioCards = [
  {
    icon: Box,
    category: '资产与质保',
    tagType: 'primary',
    color: '#3b82f6',
    title: '未来 30 天哪些资产质保到期？',
    desc: '自动扫描临期设备并定位责任保管人'
  },
  {
    icon: WarningFilled,
    category: '风控与异常',
    tagType: 'warning',
    color: '#f59e0b',
    title: '当前有哪些高风险资产？',
    desc: '排查未闭环资产风险事件与状态异常'
  },
  {
    icon: Tickets,
    category: '财务与发票',
    tagType: 'success',
    color: '#10b981',
    title: '待核对发票有多少张？',
    desc: '穿透待复核单据与识别失败发票'
  },
  {
    icon: Calendar,
    category: '办公与协同',
    tagType: 'info',
    color: '#8b5cf6',
    title: '今天有哪些日程和未读公告？',
    desc: '一览今日日程待办与未读重要通知'
  }
]

const quickPrompts = [
  { label: '临期资产盘点', query: '未来 30 天哪些资产质保到期？' },
  { label: '高风险资产', query: '当前有哪些高风险资产？' },
  { label: '待复核发票', query: '待核对发票有多少张？' },
  { label: '今日日程与公告', query: '今天有哪些日程和未读公告？' },
  { label: '保管人资产汇总', query: '按保管人汇总资产数量和价值' }
]

const toolCategories = [
  { key: 'all', label: '全部' },
  { key: 'asset', label: '资产设备', prefix: 'asset.' },
  { key: 'invoice', label: '财务发票', prefix: 'invoice.' },
  { key: 'office', label: '协同知识', prefix: ['schedule.', 'announcement.', 'knowledge.'] }
]

const sampleQueries = {
  'asset.search': '按状态查询在用资产清单',
  'asset.detail': '查询指定编号资产的详细配置与使用人',
  'asset.risk.list': '当前有哪些高风险资产？',
  'asset.warranty.expiring': '未来 30 天哪些资产质保到期？',
  'asset.custodian.summary': '按保管人汇总资产数量和价值',
  'asset.operation.summary': '汇总近期资产流转与调拨单据',
  'invoice.summary': '查询当前权限范围内的发票汇总',
  'invoice.pending_reviews': '待核对发票有多少张？',
  'invoice.failed_recognitions': '有哪些识别失败的发票？',
  'invoice.provider_quality': '查阅发票识别 Provider 质量',
  'schedule.today': '今天有哪些日程和待办事项？',
  'announcement.unread': '今天有哪些未读公告？',
  'knowledge.search': '检索资产盘点管理与报废制度'
}

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

function getCategoryCount(cat) {
  if (cat.key === 'all') return tools.value.length
  if (Array.isArray(cat.prefix)) {
    return tools.value.filter((t) => cat.prefix.some((p) => t.name.startsWith(p))).length
  }
  return tools.value.filter((t) => t.name.startsWith(cat.prefix)).length
}

const filteredTools = computed(() => {
  if (activeToolCategory.value === 'all') return tools.value
  const cat = toolCategories.find((c) => c.key === activeToolCategory.value)
  if (!cat) return tools.value
  if (Array.isArray(cat.prefix)) {
    return tools.value.filter((t) => cat.prefix.some((p) => t.name.startsWith(p)))
  }
  return tools.value.filter((t) => t.name.startsWith(cat.prefix))
})

function getToolIcon(toolName) {
  if (toolName.startsWith('asset.')) return Box
  if (toolName.startsWith('invoice.')) return Tickets
  return Calendar
}

function getToolColor(toolName) {
  if (toolName.startsWith('asset.')) return '#3b82f6'
  if (toolName.startsWith('invoice.')) return '#10b981'
  return '#8b5cf6'
}

function getToolSampleQuery(toolName) {
  return sampleQueries[toolName] || `查询 ${toolLabel(toolName)} 业务数据`
}

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
    if (total === 0 && data.list.length === 0) return []
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
  if (rows.length) return `${rows.length} 条记录明细`
  const facts = resultFacts(item)
  return facts.length ? `${facts.length} 项关键指标` : '查看明细'
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
    const [sessionRes, toolRes] = await Promise.all([
      getCopilotSessions({ paged: true, page: sessionPage.value, pageSize: sessionPageSize.value }),
      getCopilotTools()
    ])
    if (sessionRes.code === 0) {
      sessions.value = sessionRes.data?.list || []
      sessionTotal.value = Number(sessionRes.data?.total || 0)
    } else {
      ElMessage.error(sessionRes.msg || '读取会话失败')
    }
    if (toolRes.code === 0) {
      tools.value = toolRes.data || []
    } else {
      ElMessage.error(toolRes.msg || '读取可查询范围失败')
    }
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
  nextTick(() => {
    questionInputRef.value?.focus?.()
  })
}

function chooseQuestion(value) {
  question.value = value
  nextTick(() => {
    questionInputRef.value?.focus?.()
  })
}

function handleKeydown(e) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    submitQuestion()
  }
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
  await scrollToLatest()
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
.smart-copilot-page {
  padding-bottom: var(--na-space-md);
}

.header-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border: 1px solid color-mix(in srgb, var(--na-success) 24%, var(--na-border));
  border-radius: 999px;
  background: color-mix(in srgb, var(--na-success) 8%, var(--na-card));
  color: var(--na-action-success);
  font-size: 0.8125rem;
  font-weight: 600;

  .pulse-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--na-success);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--na-success) 30%, transparent);
    animation: pulse-ring 2s infinite cubic-bezier(0.45, 0, 0.55, 1);
  }
}

@keyframes pulse-ring {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 color-mix(in srgb, var(--na-success) 40%, transparent); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px transparent; }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 transparent; }
}

.copilot-layout {
  display: grid;
  grid-template-columns: 248px minmax(0, 1fr) 280px;
  align-items: stretch;
  gap: 14px;
}

.na-panel {
  min-width: 0;
  border: 1px solid var(--na-border);
  border-radius: 14px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
}

/* 1. 左侧会话面板 */
.session-panel {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 182px);
  min-height: 640px;
  padding: 16px;
  background: color-mix(in srgb, var(--na-card) 96%, var(--na-muted));
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 12px;

  &__meta {
    h2 {
      margin: 0;
      color: var(--na-foreground);
      font-size: 0.9375rem;
      font-weight: 700;
    }
    .session-counter,
    .tools-counter {
      display: block;
      margin-top: 2px;
      color: var(--na-muted-foreground);
      font-size: 0.75rem;
    }
  }

  .btn-new-session {
    border-radius: 8px;
    font-weight: 600;
  }
}

.session-search {
  margin-bottom: 10px;

  :deep(.el-input__wrapper) {
    border-radius: 8px;
  }
}

.session-list {
  display: grid;
  flex: 1;
  align-content: start;
  gap: 6px;
  overflow-y: auto;
  padding-right: 2px;
}

.session-row {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 30px;
  align-items: center;
  border: 1px solid transparent;
  border-radius: 10px;
  transition: all 180ms cubic-bezier(0.22, 1, 0.36, 1);

  &:hover {
    border-color: color-mix(in srgb, var(--na-primary) 20%, var(--na-border));
    background: color-mix(in srgb, var(--na-primary-soft) 45%, var(--na-card));
  }

  &.is-active {
    border-color: color-mix(in srgb, var(--na-primary) 35%, var(--na-border));
    background: color-mix(in srgb, var(--na-primary-soft) 85%, var(--na-card));
    box-shadow: 0 2px 8px color-mix(in srgb, var(--na-primary) 12%, transparent);

    &::before {
      position: absolute;
      top: 10px;
      bottom: 10px;
      left: 0;
      width: 3px;
      border-radius: 0 4px 4px 0;
      background: var(--na-primary);
      content: '';
    }
  }
}

.session-item {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 4px;
  padding: 10px 10px 10px 12px;
  border: 0;
  border-radius: 9px;
  background: transparent;
  color: var(--na-foreground);
  font: inherit;
  text-align: left;
  cursor: pointer;

  &:focus-visible {
    outline: 2px solid var(--na-ring);
    outline-offset: -2px;
  }

  &__header {
    display: flex;
    align-items: center;
    gap: 7px;
    min-width: 0;

    .session-item__icon {
      flex: 0 0 16px;
      color: var(--na-primary);
      font-size: 15px;
    }

    strong {
      flex: 1;
      overflow: hidden;
      font-size: 0.8125rem;
      font-weight: 600;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  &__footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-left: 23px;

    small {
      color: var(--na-muted-foreground);
      font-size: 0.6875rem;
    }

    .active-badge {
      padding: 1px 6px;
      border-radius: 4px;
      background: var(--na-primary);
      color: #fff;
      font-size: 0.625rem;
      font-weight: 600;
      line-height: 1.2;
    }
  }
}

.session-delete {
  width: 28px;
  height: 28px;
  margin: 0 4px 0 0;
  color: var(--na-muted-foreground);
  opacity: 0;
  transition: all 160ms ease;

  .session-row:hover &,
  .session-row:focus-within &,
  .session-row.is-active & {
    opacity: 1;
  }

  &:hover {
    background: var(--na-danger-soft);
    color: var(--na-danger);
  }
}

.session-empty-filtered {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px 8px;
  color: var(--na-muted-foreground);
  font-size: 0.75rem;
  text-align: center;
}

.session-pagination {
  justify-content: center;
  margin-top: 10px;
}

/* 2. 中间智能工作台 */
.chat-panel {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 182px);
  min-height: 640px;
  overflow: hidden;
  background: var(--na-card);
}

.chat-panel-header {
  display: flex;
  min-height: 68px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 24px;
  border-bottom: 1px solid var(--na-border);
  background: color-mix(in srgb, var(--na-card) 98%, var(--na-muted));
}

.chat-panel-title {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;

  .chat-panel-avatar {
    display: grid;
    width: 40px;
    height: 40px;
    flex: 0 0 40px;
    place-items: center;
    border: 1px solid color-mix(in srgb, var(--na-primary) 28%, var(--na-border));
    border-radius: 10px;
    background: linear-gradient(135deg, var(--na-primary-soft), color-mix(in srgb, var(--na-primary-soft) 40%, var(--na-card)));
    color: var(--na-primary);
    font-size: 20px;
  }

  .chat-panel-heading {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 3px;

    .chat-panel-title-row {
      display: flex;
      align-items: center;
      gap: 8px;

      h2 {
        margin: 0;
        color: var(--na-foreground);
        font-size: 0.9375rem;
        font-weight: 700;
        line-height: 1.4;
      }
    }

    span {
      overflow: hidden;
      color: var(--na-muted-foreground);
      font-size: 0.75rem;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

.chat-panel-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 14px;

  .access-state {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border: 1px solid var(--na-border);
    border-radius: 999px;
    background: var(--na-card);
    color: var(--na-foreground);
    font-size: 0.75rem;
    font-weight: 600;

    .state-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: var(--na-success);
      box-shadow: 0 0 0 2px var(--na-success-soft);
    }
  }

  .btn-reset-chat {
    border-radius: 6px;
  }
}

.chat-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 24px clamp(18px, 3.5vw, 48px);
  background: color-mix(in srgb, var(--na-card) 97%, var(--na-muted));
  scroll-behavior: smooth;
}

/* 2.1 空状态 Hero 引导 */
.empty-chat {
  display: flex;
  width: min(100%, 820px);
  min-height: 100%;
  flex-direction: column;
  justify-content: center;
  margin: 0 auto;
  padding: 20px 0;
}

.hero-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: 32px;

  .hero-badge-wrap {
    position: relative;
    display: grid;
    place-items: center;
    margin-bottom: 16px;

    .hero-badge-aura {
      position: absolute;
      inset: -10px;
      border-radius: 50%;
      background: radial-gradient(circle, color-mix(in srgb, var(--na-primary) 35%, transparent) 0%, transparent 70%);
      filter: blur(8px);
      animation: hero-glow 3s ease-in-out infinite alternate;
    }

    .hero-badge {
      position: relative;
      display: grid;
      width: 60px;
      height: 60px;
      place-items: center;
      border: 1px solid color-mix(in srgb, var(--na-primary) 40%, var(--na-border));
      border-radius: 16px;
      background: linear-gradient(135deg, var(--na-primary), #4f46e5);
      color: #fff;
      box-shadow: 0 8px 24px color-mix(in srgb, var(--na-primary) 35%, transparent);

      .hero-robot-icon {
        width: 32px;
        height: 32px;
      }
    }
  }

  .hero-copy {
    max-width: 620px;
    margin-bottom: 18px;

    h3 {
      margin: 0 0 8px;
      color: var(--na-foreground);
      font-size: 1.35rem;
      font-weight: 800;
      letter-spacing: -0.02em;
    }

    p {
      margin: 0;
      color: var(--na-muted-foreground);
      font-size: 0.875rem;
      line-height: 1.65;
    }
  }

  .hero-trust-badges {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 10px;

    .trust-pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 5px 12px;
      border: 1px solid var(--na-border);
      border-radius: 999px;
      background: var(--na-card);
      color: var(--na-muted-foreground);
      font-size: 0.75rem;
      font-weight: 500;

      .el-icon {
        color: var(--na-primary);
      }
    }
  }
}

@keyframes hero-glow {
  0% { transform: scale(0.9); opacity: 0.5; }
  100% { transform: scale(1.15); opacity: 0.9; }
}

.scenario-section {
  .scenario-title {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 12px;
    padding: 0 4px;

    span {
      color: var(--na-foreground);
      font-size: 0.8125rem;
      font-weight: 700;
    }

    small {
      color: var(--na-muted-foreground);
      font-size: 0.6875rem;
    }
  }

  .scenario-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .scenario-card {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 14px;
    padding: 14px 16px;
    border: 1px solid var(--na-border);
    border-radius: 12px;
    background: var(--na-card);
    color: var(--na-foreground);
    text-align: left;
    cursor: pointer;
    transition: all 200ms cubic-bezier(0.22, 1, 0.36, 1);

    &:hover {
      border-color: color-mix(in srgb, var(--na-primary) 50%, var(--na-border));
      background: color-mix(in srgb, var(--na-primary-soft) 40%, var(--na-card));
      box-shadow: 0 6px 18px color-mix(in srgb, var(--na-primary) 10%, transparent);
      transform: translateY(-2px);

      .scenario-card__arrow {
        transform: translateX(3px);
        color: var(--na-primary);
      }
    }

    &__icon {
      display: grid;
      width: 38px;
      height: 38px;
      flex: 0 0 38px;
      place-items: center;
      border-radius: 10px;
      font-size: 19px;
    }

    &__content {
      display: flex;
      flex: 1;
      min-width: 0;
      flex-direction: column;
      gap: 3px;
    }

    &__category {
      margin-bottom: 2px;
    }

    &__title {
      overflow: hidden;
      color: var(--na-foreground);
      font-size: 0.8125rem;
      font-weight: 650;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    &__desc {
      overflow: hidden;
      color: var(--na-muted-foreground);
      font-size: 0.6875rem;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    &__arrow {
      flex: 0 0 auto;
      color: var(--na-muted-foreground);
      font-size: 15px;
      transition: transform 180ms ease, color 180ms ease;
    }
  }
}

/* 2.2 对话气泡 */
.message {
  min-width: 0;
  margin: 0 auto 24px;
}

.message--user {
  display: flex;
  width: min(100%, 960px);
  align-items: flex-start;
  justify-content: flex-end;
  gap: 12px;

  .user-message-container {
    display: flex;
    max-width: min(76%, 680px);
    flex-direction: column;
    align-items: flex-end;
  }

  .user-message-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 5px;

    .user-role-label {
      color: var(--na-foreground);
      font-size: 0.75rem;
      font-weight: 600;
    }
  }

  .user-message-bubble {
    padding: 12px 16px;
    border: 1px solid color-mix(in srgb, var(--na-primary) 30%, var(--na-border));
    border-radius: 14px 2px 14px 14px;
    background: linear-gradient(135deg, color-mix(in srgb, var(--na-primary-soft) 85%, var(--na-card)), color-mix(in srgb, var(--na-primary-soft) 40%, var(--na-card)));
    color: var(--na-foreground);
    font-size: 0.875rem;
    line-height: 1.68;
    box-shadow: 0 2px 8px color-mix(in srgb, var(--na-primary) 10%, transparent);
    overflow-wrap: anywhere;
  }

  .user-avatar {
    display: grid;
    width: 36px;
    height: 36px;
    flex: 0 0 36px;
    place-items: center;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--na-primary), #0284c7);
    color: #fff;
    font-size: 0.75rem;
    font-weight: 700;
    box-shadow: 0 3px 10px color-mix(in srgb, var(--na-primary) 30%, transparent);
  }
}

.message--assistant,
.message--pending {
  display: flex;
  width: min(100%, 960px);
  align-items: flex-start;
  gap: 12px;
}

.assistant-avatar {
  position: relative;
  display: grid;
  width: 36px;
  height: 36px;
  flex: 0 0 36px;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--na-primary) 32%, var(--na-border));
  border-radius: 10px;
  background: linear-gradient(135deg, var(--na-primary-soft), color-mix(in srgb, var(--na-primary-soft) 30%, var(--na-card)));
  color: var(--na-primary);
  box-shadow: 0 2px 8px color-mix(in srgb, var(--na-primary) 12%, transparent);

  .assistant-robot {
    width: 20px;
    height: 20px;
  }

  &--thinking {
    background: color-mix(in srgb, var(--na-primary-soft) 90%, var(--na-card));

    &::before {
      position: absolute;
      inset: -4px;
      border: 2px solid color-mix(in srgb, var(--na-primary) 20%, transparent);
      border-top-color: var(--na-primary);
      border-right-color: color-mix(in srgb, var(--na-primary) 70%, transparent);
      border-radius: 50%;
      content: '';
      animation: assistant-orbit 1.2s linear infinite;
    }

    .assistant-robot {
      animation: assistant-robot-pulse 1.6s ease-in-out infinite;
    }
  }
}

.assistant-message-body {
  flex: 1;
  min-width: 0;
  padding: 14px 18px;
  border: 1px solid var(--na-border);
  border-radius: 2px 14px 14px 14px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
}

.assistant-message-header {
  display: flex;
  min-height: 32px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid color-mix(in srgb, var(--na-border) 70%, transparent);

  .assistant-identity {
    display: flex;
    align-items: baseline;
    gap: 8px;

    .identity-name {
      display: flex;
      align-items: center;
      gap: 6px;

      strong {
        color: var(--na-foreground);
        font-size: 0.875rem;
        font-weight: 700;
      }

      .copilot-pill {
        padding: 1px 6px;
        border-radius: 4px;
        background: var(--na-primary-soft);
        color: var(--na-primary);
        font-size: 0.625rem;
        font-weight: 700;
      }
    }

    .identity-role {
      color: var(--na-muted-foreground);
      font-size: 0.6875rem;
    }
  }

  .assistant-meta {
    display: flex;
    align-items: center;
    gap: 8px;

    .message-status {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      padding: 2px 8px;
      border-radius: 6px;
      font-size: 0.6875rem;
      font-weight: 600;

      &--complete {
        background: var(--na-success-soft);
        color: var(--na-action-success);
      }
      &--partial {
        background: var(--na-warning-soft);
        color: var(--na-action-warning);
      }
    }

    .message-time {
      color: var(--na-muted-foreground);
      font-size: 0.6875rem;
    }
  }
}

.tool-badges {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  width: fit-content;
  max-width: 100%;
  margin-bottom: 10px;
  padding: 4px 10px;
  border: 1px solid color-mix(in srgb, var(--na-border) 60%, transparent);
  border-radius: 6px;
  background: color-mix(in srgb, var(--na-card) 95%, var(--na-muted));

  .tool-badges-title {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--na-muted-foreground);
    font-size: 0.6875rem;
    font-weight: 600;
  }

  .tool-badge {
    display: inline-flex;
    align-items: center;
    padding: 2px 8px;
    border: 1px solid color-mix(in srgb, var(--na-primary) 30%, var(--na-border));
    border-radius: 6px;
    background: var(--na-primary-soft);
    color: var(--na-primary);
    font-size: 0.6875rem;
    font-weight: 600;
  }
}

.assistant-markdown-container {
  min-width: 0;
  overflow-wrap: anywhere;
}

.partial-note {
  margin: 12px 0 0;
  padding: 8px 12px;
  border-left: 3px solid var(--na-warning);
  border-radius: 0 6px 6px 0;
  background: var(--na-warning-soft);
  color: var(--na-action-warning);
  font-size: 0.75rem;
  line-height: 1.5;
}

/* 结构化数据面板 */
.message-result {
  margin-top: 14px;
  border: 1px solid var(--na-border);
  border-radius: 10px;
  background: color-mix(in srgb, var(--na-card) 98%, var(--na-muted));
  overflow: hidden;

  summary {
    display: flex;
    min-height: 42px;
    align-items: center;
    justify-content: space-between;
    padding: 0 14px;
    color: var(--na-foreground);
    font-size: 0.8125rem;
    font-weight: 650;
    cursor: pointer;
    list-style: none;

    &::-webkit-details-marker { display: none; }

    &:hover {
      background: color-mix(in srgb, var(--na-primary-soft) 30%, transparent);
    }

    .result-title {
      display: inline-flex;
      align-items: center;
      gap: 7px;
      color: var(--na-primary);
    }

    small {
      margin-left: auto;
      margin-right: 8px;
      color: var(--na-muted-foreground);
      font-size: 0.6875rem;
      font-weight: 500;
    }

    .result-chevron {
      color: var(--na-muted-foreground);
      transition: transform 180ms ease;
    }
  }

  &[open] summary .result-chevron {
    transform: rotate(180deg);
  }

  .result-content {
    padding: 12px;
    border-top: 1px solid var(--na-border);
  }

  .result-facts {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 8px;
    margin-bottom: 10px;

    .result-fact {
      display: flex;
      flex-direction: column;
      gap: 4px;
      padding: 10px 12px;
      border: 1px solid var(--na-border);
      border-radius: 8px;
      background: var(--na-card);

      .fact-label {
        color: var(--na-muted-foreground);
        font-size: 0.6875rem;
      }

      .fact-value {
        color: var(--na-foreground);
        font-size: 1rem;
        font-weight: 700;
        font-variant-numeric: tabular-nums;
      }
    }
  }

  .result-table {
    border: 1px solid var(--na-border);
    border-radius: 8px;
    overflow: hidden;

    .cell-content {
      font-size: 0.75rem;
      font-variant-numeric: tabular-nums;
    }
  }
}

/* 事实依据链接 */
.message-citations {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
  margin-top: 14px;
  padding-top: 10px;
  border-top: 1px solid color-mix(in srgb, var(--na-border) 60%, transparent);

  .citations-label {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    color: var(--na-muted-foreground);
    font-size: 0.75rem;
    font-weight: 600;
  }

  .citation-links {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;

    .citation-btn {
      border-radius: 6px;
      font-size: 0.75rem;
    }
  }
}

/* 思考中卡片 */
.message--pending {
  animation: message-enter 180ms cubic-bezier(0.22, 1, 0.36, 1);

  .pending-container {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 12px 16px;
    border: 1px solid color-mix(in srgb, var(--na-primary) 30%, var(--na-border));
    border-radius: 2px 14px 14px 14px;
    background: var(--na-card);

    .pending-header {
      display: flex;
      align-items: center;
      gap: 8px;

      strong {
        color: var(--na-primary);
        font-size: 0.8125rem;
      }
    }

    .pending-tip {
      color: var(--na-muted-foreground);
      font-size: 0.75rem;
    }
  }
}

.assistant-dots {
  display: inline-flex;
  align-items: center;
  gap: 3px;

  i {
    display: block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--na-primary);
    animation: assistant-dot 1.05s ease-in-out infinite;

    &:nth-child(2) { animation-delay: 140ms; }
    &:nth-child(3) { animation-delay: 280ms; }
  }
}

/* 2.3 底部交互输入台 */
.composer-container {
  padding: 10px 24px 16px;
  border-top: 1px solid var(--na-border);
  background: var(--na-card);
}

.quick-prompts-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  overflow: hidden;

  .quick-prompts-label {
    flex: 0 0 auto;
    color: var(--na-muted-foreground);
    font-size: 0.6875rem;
    font-weight: 600;
  }

  .quick-prompts-scroll {
    display: flex;
    align-items: center;
    gap: 6px;
    overflow-x: auto;
    scrollbar-width: none;

    &::-webkit-scrollbar { display: none; }
  }

  .quick-prompt-pill {
    flex: 0 0 auto;
    padding: 3px 10px;
    border: 1px solid var(--na-border);
    border-radius: 999px;
    background: color-mix(in srgb, var(--na-card) 95%, var(--na-muted));
    color: var(--na-foreground);
    font-size: 0.6875rem;
    cursor: pointer;
    transition: all 160ms ease;

    &:hover {
      border-color: var(--na-primary);
      background: var(--na-primary-soft);
      color: var(--na-primary);
    }
  }
}

.composer-box {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  padding: 6px 8px 6px 12px;
  border: 1px solid var(--na-border);
  border-radius: 12px;
  background: color-mix(in srgb, var(--na-card) 96%, var(--na-muted));
  transition: all 180ms ease;

  &:focus-within {
    border-color: var(--na-primary);
    box-shadow: 0 0 0 3px var(--na-ring), 0 4px 14px color-mix(in srgb, var(--na-primary) 12%, transparent);
    background: var(--na-card);
  }
}

.composer-input {
  flex: 1;

  :deep(.el-textarea__inner) {
    min-height: 52px !important;
    padding: 4px 0;
    border: 0 !important;
    box-shadow: none !important;
    background: transparent !important;
    color: var(--na-foreground);
    font-size: 0.875rem;
    line-height: 1.6;
  }

  :deep(.el-input__count) {
    right: 4px;
    bottom: 2px;
    background: transparent;
    color: var(--na-muted-foreground);
    font-size: 0.6875rem;
  }
}

.composer-submit {
  width: 40px;
  min-width: 40px;
  height: 40px;
  margin-bottom: 2px;
  border-radius: 10px;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--na-primary) 30%, transparent);
  transition: transform 160ms ease;

  &:hover:not(:disabled) {
    transform: scale(1.05);
  }
}

.composer-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
  padding: 0 2px;

  .composer-guard {
    display: flex;
    align-items: center;
    gap: 5px;
    color: var(--na-muted-foreground);
    font-size: 0.6875rem;

    .el-icon {
      color: var(--na-success);
    }
  }

  .composer-hint {
    color: var(--na-muted-foreground);
    font-size: 0.6875rem;
  }
}

/* 3. 右侧可查询能力矩阵 */
.tools-panel {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 182px);
  min-height: 640px;
  padding: 16px;
  background: color-mix(in srgb, var(--na-card) 96%, var(--na-muted));
}

.tool-category-tabs {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 12px;
  padding: 3px;
  border-radius: 8px;
  background: color-mix(in srgb, var(--na-card) 85%, var(--na-muted));

  .category-tab-btn {
    flex: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 5px 0;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: var(--na-muted-foreground);
    font-size: 0.6875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 160ms ease;

    .tab-count {
      padding: 0 4px;
      border-radius: 999px;
      background: color-mix(in srgb, var(--na-muted) 80%, transparent);
      font-size: 0.625rem;
    }

    &.is-active {
      background: var(--na-card);
      color: var(--na-primary);
      box-shadow: 0 1px 4px var(--na-shadow-sm);

      .tab-count {
        background: var(--na-primary-soft);
        color: var(--na-primary);
      }
    }
  }
}

.tools-list {
  display: grid;
  flex: 1;
  align-content: start;
  gap: 8px;
  overflow-y: auto;
  padding-right: 2px;
}

.tool-card {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid var(--na-border);
  border-radius: 10px;
  background: var(--na-card);
  cursor: pointer;
  transition: all 180ms cubic-bezier(0.22, 1, 0.36, 1);

  &:hover {
    border-color: color-mix(in srgb, var(--na-primary) 40%, var(--na-border));
    box-shadow: 0 4px 12px color-mix(in srgb, var(--na-primary) 8%, transparent);
    transform: translateY(-1px);

    .tool-card__prompt-hint {
      color: var(--na-primary);
      transform: translateX(2px);
    }
  }

  &__icon {
    display: grid;
    width: 30px;
    height: 30px;
    flex: 0 0 30px;
    place-items: center;
    margin-top: 1px;
    border-radius: 8px;
    font-size: 15px;
  }

  &__info {
    flex: 1;
    min-width: 0;

    .tool-card__header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 3px;

      strong {
        color: var(--na-foreground);
        font-size: 0.75rem;
        font-weight: 700;
      }

      .tool-card__prompt-hint {
        color: var(--na-muted-foreground);
        font-size: 12px;
        transition: transform 160ms ease, color 160ms ease;
      }
    }

    p {
      margin: 0;
      color: var(--na-muted-foreground);
      font-size: 0.6875rem;
      line-height: 1.45;
    }
  }
}

.tool-empty-filtered {
  padding: 24px 8px;
  color: var(--na-muted-foreground);
  font-size: 0.75rem;
  text-align: center;
}

/* 动效关键帧 */
@keyframes message-enter {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes assistant-orbit {
  to { transform: rotate(360deg); }
}

@keyframes assistant-robot-pulse {
  0%, 100% { opacity: 0.75; transform: scale(0.95); }
  50% { opacity: 1; transform: scale(1); }
}

@keyframes assistant-dot {
  0%, 65%, 100% { opacity: 0.3; transform: translateY(0) scale(0.8); }
  32% { opacity: 1; transform: translateY(-3px) scale(1); }
}

/* 响应式适配 */
@media (max-width: 1200px) {
  .copilot-layout {
    grid-template-columns: 220px minmax(0, 1fr);
  }
  .tools-panel {
    display: none;
  }
}

@media (max-width: 768px) {
  .copilot-layout {
    display: flex;
    flex-direction: column;
  }

  .session-panel {
    width: 100%;
    height: auto;
    max-height: 220px;
    min-height: auto;
  }

  .chat-panel {
    width: 100%;
    height: calc(100dvh - 360px);
    min-height: 520px;
  }

  .scenario-section .scenario-grid {
    grid-template-columns: 1fr;
  }

  .message-result .result-facts {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (prefers-reduced-motion: reduce) {
  .chat-scroll { scroll-behavior: auto; }
  .session-row,
  .scenario-card,
  .tool-card,
  .composer-box,
  .result-chevron,
  .message--pending,
  .assistant-avatar--thinking::before,
  .assistant-avatar--thinking .assistant-robot,
  .assistant-dots i,
  .hero-badge-aura {
    transition: none !important;
    animation: none !important;
  }
}
</style>
