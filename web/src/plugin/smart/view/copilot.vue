<template>
  <main class="na-page smart-copilot-page">
    <AppPageHeader title-id="smart-copilot-title" title="业务助手" description="只读查询资产、风险、发票、日程和公告；每条回答都保留数据引用。">
      <template #actions><el-button :icon="Refresh" :loading="loading" @click="loadSessions">刷新</el-button></template>
    </AppPageHeader>
    <div class="copilot-layout">
      <aside class="na-panel session-panel">
        <header class="panel-header"><h2>会话</h2><el-button text :icon="Plus" @click="newSession">新会话</el-button></header>
        <div v-for="item in sessions" :key="item.ID" class="session-row" :class="{ active: item.ID === sessionId }">
          <button class="session-item" @click="openSession(item.ID)">{{ item.title || '未命名会话' }}<small>{{ formatTime(item.lastMessageAt) }}</small></button>
          <el-button class="session-delete" text :icon="Delete" title="删除会话" aria-label="删除会话" @click="removeSession(item)" />
        </div>
        <el-empty v-if="!sessions.length" description="暂无会话" :image-size="64" />
      </aside>
      <section class="na-panel chat-panel">
        <div class="chat-scroll" aria-live="polite">
          <div v-for="(item, index) in messages" :key="item.ID || index" class="message" :class="item.role === 'user' ? 'message-user' : 'message-assistant'">
            <div class="message-role">{{ item.role === 'user' ? '我' : '助手' }}<span v-if="item.tool">{{ item.tool }}</span></div>
            <p>{{ item.content }}</p>
            <div v-if="citationList(item.citations).length" class="citations message-citations"><span>数据引用</span><el-button v-for="citation in citationList(item.citations)" :key="`${citation.type}-${citation.id || citation.label}`" link type="primary" :icon="ArrowRight" @click="openCitation(citation)">{{ citation.label }}</el-button></div>
          </div>
          <div v-if="!messages.length" class="empty-chat"><el-icon><ChatDotRound /></el-icon><strong>问我资产和运营数据</strong><span>例如：查询未来 30 天质保到期的资产</span></div>
          <div v-if="result" class="result-card">
            <div class="result-meta"><el-tag size="small" type="info">{{ result.tool }}</el-tag><span>只读查询 · {{ formatTime(result.generatedAt) }}</span></div>
            <p class="result-answer">{{ result.answer }}</p>
            <el-table v-if="tableRows.length" :data="tableRows" size="small" max-height="260">
              <el-table-column v-for="column in tableColumns" :key="column" :prop="column" :label="columnLabel(column)" min-width="130" show-overflow-tooltip />
            </el-table>
            <div v-if="result.citations?.length" class="citations"><span>数据引用</span><el-button v-for="item in result.citations" :key="`${item.type}-${item.id || item.label}`" link type="primary" :icon="ArrowRight" @click="openCitation(item)">{{ item.label }}</el-button></div>
          </div>
        </div>
        <form class="composer" @submit.prevent="submitQuestion"><el-input v-model="question" type="textarea" :rows="2" maxlength="2000" show-word-limit resize="none" placeholder="输入只读业务问题" :disabled="sending" /><el-button type="primary" :icon="Position" :loading="sending" native-type="submit">查询</el-button></form>
      </section>
      <aside class="na-panel tools-panel"><header class="panel-header"><h2>可用 Tool</h2></header><div v-for="item in tools" :key="item.name" class="tool-item"><el-icon><CircleCheck /></el-icon><div><strong>{{ item.name }}</strong><span>{{ item.description }}</span></div></div><el-empty v-if="!tools.length" description="当前角色暂无可用 Tool" :image-size="56" /></aside>
    </div>
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowRight, ChatDotRound, CircleCheck, Delete, Plus, Position, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { deleteCopilotSession, getCopilotSession, getCopilotSessions, getCopilotTools, queryCopilot } from '@/plugin/smart/api/smart'

defineOptions({ name: 'SmartCopilot' })
const router = useRouter()
const loading = ref(false); const sending = ref(false); const sessions = ref([]); const messages = ref([]); const tools = ref([]); const result = ref(null); const sessionId = ref(0); const question = ref('')
const tableRows = computed(() => { const data = result.value?.data; if (Array.isArray(data)) return data; if (Array.isArray(data?.list)) return data.list; return [] })
const tableColumns = computed(() => { const first = tableRows.value[0]; return first ? Object.keys(first).filter((key) => typeof first[key] !== 'object').slice(0, 6) : [] })
function columnLabel(key) { return ({ assetCode: '资产编号', name: '名称', status: '状态', title: '标题', severity: '等级', sellerName: '销售方', totalCents: '含税金额', date: '日期', time: '时间' }[key] || key) }
function formatTime(value) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—' }
function citationList(value) { return Array.isArray(value) ? value : Object.values(value || {}) }
function openCitation(item) { const routeName = ({ asset: 'assetInventory', risk: 'assetRiskCenter', asset_operation: 'assetInbound', invoice: 'invoiceLedger', invoice_quality: 'invoiceQuality', schedule: 'workSchedule', announcement: 'anInfo' })[item.type]; const query = Object.fromEntries(new URLSearchParams(item.params || '')); if (routeName && router.hasRoute(routeName)) router.push({ name: routeName, query }); else if (item.path) router.push({ path: item.path, query }) }
async function loadSessions() { loading.value = true; try { const [sessionRes, toolRes] = await Promise.all([getCopilotSessions(), getCopilotTools()]); if (sessionRes.code === 0) sessions.value = sessionRes.data || []; if (toolRes.code === 0) tools.value = toolRes.data || [] } finally { loading.value = false } }
async function openSession(id) { sessionId.value = id; result.value = null; const res = await getCopilotSession({ id }); if (res.code === 0) messages.value = res.data?.messages || [] }
function newSession() { sessionId.value = 0; messages.value = []; result.value = null }
async function removeSession(item) { try { await ElMessageBox.confirm(`确认删除会话“${item.title || '未命名会话'}”？`, '删除会话', { type: 'warning' }); const res = await deleteCopilotSession({ id: item.ID }); if (res.code !== 0) { ElMessage.error(res.msg || '删除失败'); return } if (sessionId.value === item.ID) newSession(); await loadSessions(); ElMessage.success(res.msg || '会话已删除') } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.error('删除失败') } }
async function submitQuestion() { if (!question.value.trim()) return; sending.value = true; try { const res = await queryCopilot({ question: question.value.trim(), sessionId: sessionId.value || undefined }); if (res.code !== 0) { ElMessage.error(res.msg || '查询失败'); return } result.value = res.data; messages.value.push({ role: 'user', content: question.value.trim() }, { role: 'assistant', content: res.data.answer, tool: res.data.tool }); sessionId.value = res.data.sessionId; question.value = ''; await loadSessions() } finally { sending.value = false } }
onMounted(loadSessions)
</script>

<style scoped lang="scss">
.copilot-layout { display: grid; grid-template-columns: 220px minmax(0, 1fr) 250px; gap: 14px; min-height: calc(100vh - 170px); }.na-panel { min-width: 0; padding: 16px; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-card); box-shadow: var(--na-shadow-sm); }.panel-header { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 12px; }.panel-header h2 { margin: 0; font-size: .95rem; }.session-row { display: grid; grid-template-columns: minmax(0, 1fr) 30px; align-items: center; border-radius: 6px; }.session-row:hover, .session-row.active { background: var(--na-surface-muted); }.session-item { display: flex; min-width: 0; flex-direction: column; gap: 4px; padding: 10px; border: 0; background: transparent; color: var(--na-foreground); text-align: left; cursor: pointer; }.session-item small { color: var(--na-muted-foreground); }.session-delete { width: 28px; height: 28px; margin: 0; color: var(--na-muted-foreground); }.session-delete:hover { color: var(--el-color-danger); }.chat-panel { display: flex; min-height: 620px; flex-direction: column; }.chat-scroll { flex: 1; min-height: 0; overflow: auto; padding: 4px 2px 16px; }.message { max-width: 82%; margin: 0 0 12px; padding: 10px 12px; border-radius: 8px; }.message-user { margin-left: auto; background: var(--el-color-primary-light-9); }.message-assistant { border: 1px solid var(--na-border); background: var(--na-surface-muted); }.message-role { display: flex; justify-content: space-between; margin-bottom: 5px; color: var(--na-muted-foreground); font-size: .72rem; }.message-role span { color: var(--el-color-primary); }.message p, .result-answer { margin: 0; line-height: 1.6; white-space: pre-wrap; }.message-citations { padding-top: 8px; border-top: 1px solid var(--na-border); }.empty-chat { display: flex; min-height: 360px; flex-direction: column; align-items: center; justify-content: center; gap: 8px; color: var(--na-muted-foreground); }.empty-chat strong { color: var(--na-foreground); }.result-card { margin: 4px 0 14px; padding: 12px; border: 1px solid var(--na-border); border-radius: 8px; }.result-meta, .citations { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 10px; color: var(--na-muted-foreground); font-size: .75rem; }.citations { margin: 12px 0 0; }.citations .el-button { max-width: 100%; margin: 0; white-space: normal; text-align: left; }.composer { display: flex; align-items: flex-end; gap: 10px; padding-top: 12px; border-top: 1px solid var(--na-border); }.composer .el-input { flex: 1; }.tools-panel { align-self: start; }.tool-item { display: flex; gap: 8px; padding: 9px 0; border-bottom: 1px solid var(--na-border); }.tool-item:last-child { border-bottom: 0; }.tool-item .el-icon { flex: 0 0 auto; color: var(--el-color-success); }.tool-item div { display: flex; min-width: 0; flex-direction: column; gap: 3px; }.tool-item strong { font-size: .78rem; }.tool-item span { color: var(--na-muted-foreground); font-size: .72rem; line-height: 1.4; }
@media (max-width: 1100px) { .copilot-layout { grid-template-columns: 190px minmax(0, 1fr); }.tools-panel { display: none; } }
@media (max-width: 700px) { .copilot-layout { display: flex; flex-direction: column; min-height: 0; }.session-panel { max-height: 180px; overflow: auto; }.chat-panel { min-height: calc(100vh - 240px); }.composer { flex-direction: column; align-items: stretch; }.composer .el-button { width: 100%; }.message { max-width: 94%; } }
</style>
