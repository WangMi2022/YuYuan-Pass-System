<template>
  <main class="na-page na-page--list ai-operations-page">
    <AppPageHeader title-id="ai-operations-title" title="AI 运营" description="统一管理模型供应商、调用配额、Prompt 版本及不含原文内容的调用审计。">
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="loadAll">刷新</el-button>
        <el-button type="primary" :icon="Check" :loading="savingProviders" @click="saveProviders">保存配置</el-button>
      </template>
    </AppPageHeader>

    <section class="summary-band" aria-label="AI 用量摘要">
      <div><span>今日调用</span><strong>{{ usage.todayRequests || 0 }}</strong><small>当前账户成功请求</small></div>
      <div><span>今日 Token</span><strong>{{ number(usage.todayTokens) }}</strong><small>输入与输出合计</small></div>
      <div><span>本月费用</span><strong>{{ money(usage.monthCostMicros) }}</strong><small>按 Provider 单价估算</small></div>
      <div><span>累计调用</span><strong>{{ number(usage.totalRequests) }}</strong><small>当前账户成功请求</small></div>
    </section>

    <el-tabs v-model="activeTab" class="operation-tabs">
      <el-tab-pane label="Provider" name="providers">
        <section class="na-panel config-panel">
          <div class="na-panel-header"><div><h2>统一模型 Provider</h2><p>密钥为写入即隐藏字段；留空会保留当前密钥。</p></div><el-switch v-model="providers.enabled" active-text="启用 Gateway" /></div>
          <el-alert v-if="!providers.enabled" type="info" :closable="false" title="Gateway 当前关闭，现有自动代码兼容接口不受影响。" />
          <el-form label-position="top" class="provider-form">
            <el-form-item label="允许内网端点"><el-switch v-model="providers['allow-private-endpoints']" active-text="允许" inactive-text="拒绝" /></el-form-item>
            <div class="policy-grid">
              <el-form-item label="业务敏感词">
                <el-select v-model="providers['sensitive-words']" multiple filterable allow-create default-first-option placeholder="输入敏感词后回车" />
              </el-form-item>
              <el-form-item label="允许发送图片的模块">
                <el-select v-model="providers['allow-vision-modules']" multiple filterable allow-create default-first-option placeholder="输入模块标识后回车" />
              </el-form-item>
            </div>
            <div class="provider-grid">
              <section v-for="provider in providerList" :key="provider.key" class="provider-section">
                <header><div><h3>{{ provider.label }}</h3><span>{{ provider.hint }}</span></div><el-switch v-model="providers[provider.key].enabled" /></header>
                <el-form-item label="Base URL"><el-input v-model="providers[provider.key]['base-url']" placeholder="https://api.example.com/v1" /></el-form-item>
                <el-form-item label="模型"><el-input v-model="providers[provider.key].model" placeholder="模型名称" /></el-form-item>
                <el-form-item :label="providers[provider.key]['api-key-configured'] ? '替换 API Key（已配置）' : 'API Key'"><el-input v-model="providers[provider.key]['api-key']" type="password" show-password :disabled="providers[provider.key]['clear-api-key']" placeholder="留空不变" /></el-form-item>
                <el-checkbox v-if="providers[provider.key]['api-key-configured']" v-model="providers[provider.key]['clear-api-key']" class="clear-key-control">清除已配置 API Key</el-checkbox>
                <div class="cost-grid">
                  <el-form-item label="超时（秒）"><el-input-number v-model="providers[provider.key]['timeout-seconds']" :min="1" :max="120" controls-position="right" /></el-form-item>
                  <el-form-item label="输入费用（微单位）/ 百万 Token"><el-input-number v-model="providers[provider.key]['input-cost-micros-per-million']" :min="0" controls-position="right" /></el-form-item>
                  <el-form-item label="输出费用（微单位）/ 百万 Token"><el-input-number v-model="providers[provider.key]['output-cost-micros-per-million']" :min="0" controls-position="right" /></el-form-item>
                </div>
              </section>
            </div>
          </el-form>
        </section>
      </el-tab-pane>

      <el-tab-pane label="调用审计" name="invocations">
        <section class="na-panel table-panel">
          <header class="na-panel-header"><div><h2>模型调用审计</h2><p>仅保留用量、哈希和状态，不保存 Prompt、图片或模型输出。</p></div><el-button :icon="Search" @click="loadInvocations">查询</el-button></header>
          <div class="filter-row"><el-select v-model="invocationSearch.status" clearable placeholder="全部状态"><el-option label="成功" value="success" /><el-option label="失败" value="failed" /><el-option label="已阻断" value="blocked" /></el-select><el-input v-model="invocationSearch.module" clearable placeholder="业务模块" /><el-input v-model="invocationSearch.provider" clearable placeholder="Provider" /><el-input-number v-model="invocationSearch.userId" :min="1" :controls="false" placeholder="用户 ID" /></div>
          <el-table v-loading="invocationLoading" :data="invocations" row-key="ID">
            <el-table-column prop="CreatedAt" label="时间" min-width="165"><template #default="{ row }">{{ dateTime(row.CreatedAt) }}</template></el-table-column>
            <el-table-column prop="userId" label="用户" width="80" /><el-table-column prop="module" label="模块" min-width="110" /><el-table-column prop="operation" label="操作" min-width="140" /><el-table-column prop="provider" label="Provider" min-width="145" /><el-table-column prop="model" label="模型" min-width="145" show-overflow-tooltip />
            <el-table-column label="Token" width="120" align="right"><template #default="{ row }">{{ number(Number(row.inputTokens) + Number(row.outputTokens)) }}</template></el-table-column><el-table-column label="费用" width="120" align="right"><template #default="{ row }">{{ money(row.estimatedCostMicros) }}</template></el-table-column><el-table-column prop="durationMs" label="耗时" width="105" align="right"><template #default="{ row }">{{ row.durationMs }} ms</template></el-table-column>
            <el-table-column prop="errorType" label="错误类型" width="105"><template #default="{ row }">{{ row.errorType || '—' }}</template></el-table-column><el-table-column label="状态" width="100" align="center"><template #default="{ row }"><el-tag :type="statusMeta(row.status).type" effect="light">{{ statusMeta(row.status).label }}</el-tag></template></el-table-column>
          </el-table>
          <div class="na-pagination"><el-pagination v-model:current-page="invocationSearch.page" v-model:page-size="invocationSearch.pageSize" :total="invocationTotal" layout="total, sizes, prev, pager, next" :page-sizes="[10, 20, 50, 100]" @current-change="loadInvocations" @size-change="loadInvocations" /></div>
        </section>
      </el-tab-pane>

      <el-tab-pane label="配额" name="quotas">
        <section class="na-panel table-panel"><header class="na-panel-header"><div><h2>用量配额</h2><p>所有已启用的范围同时生效；数值为 0 表示不限制。</p></div><el-button type="primary" :icon="Plus" @click="openQuota">新增配额</el-button></header>
          <el-table v-loading="quotaLoading" :data="quotas" row-key="ID"><el-table-column prop="scopeType" label="范围" width="130" /><el-table-column prop="scopeId" label="范围标识" min-width="180" /><el-table-column prop="dailyRequests" label="每日请求" width="120" /><el-table-column prop="dailyTokens" label="每日 Token" width="130" /><el-table-column prop="monthlyCostMicros" label="月预算" width="150"><template #default="{ row }">{{ money(row.monthlyCostMicros) }}</template></el-table-column><el-table-column prop="maxConcurrency" label="最大并发" width="120" /><el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag></template></el-table-column><el-table-column label="操作" width="90"><template #default="{ row }"><el-button text type="primary" :icon="Edit" @click="openQuota(row)">编辑</el-button></template></el-table-column></el-table>
        </section>
      </el-tab-pane>

      <el-tab-pane label="Prompt" name="prompts">
        <section class="na-panel table-panel"><header class="na-panel-header"><div><h2>Prompt 模板版本</h2><p>创建草稿后激活；每个标识只会保留一个活跃版本。</p></div><el-button type="primary" :icon="Plus" @click="openPrompt">创建版本</el-button></header>
          <el-table v-loading="promptLoading" :data="prompts" row-key="ID"><el-table-column prop="promptKey" label="标识" min-width="170" /><el-table-column prop="version" label="版本" width="85" /><el-table-column prop="status" label="状态" width="105"><template #default="{ row }"><el-tag :type="promptStatus(row.status).type">{{ promptStatus(row.status).label }}</el-tag></template></el-table-column><el-table-column prop="CreatedAt" label="创建时间" min-width="165"><template #default="{ row }">{{ dateTime(row.CreatedAt) }}</template></el-table-column><el-table-column label="内容" min-width="240" show-overflow-tooltip><template #default="{ row }">{{ row.content }}</template></el-table-column><el-table-column label="操作" width="100"><template #default="{ row }"><el-button v-if="row.status !== 'active'" text type="primary" :icon="Check" @click="activatePrompt(row)">激活</el-button></template></el-table-column></el-table>
        </section>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="quotaDialogVisible" :title="quotaForm.ID ? '编辑 AI 配额' : '新增 AI 配额'" width="min(560px, calc(100vw - 32px))" destroy-on-close>
      <el-form label-position="top"><div class="dialog-grid"><el-form-item label="范围"><el-select v-model="quotaForm.scopeType"><el-option label="全局" value="global" /><el-option label="模块" value="module" /><el-option label="角色" value="authority" /><el-option label="用户" value="user" /></el-select></el-form-item><el-form-item label="范围标识"><el-input v-model="quotaForm.scopeId" :placeholder="quotaForm.scopeType === 'global' ? 'global' : '模块名、角色ID或用户ID'" /></el-form-item></div><div class="dialog-grid"><el-form-item label="每日请求"><el-input-number v-model="quotaForm.dailyRequests" :min="0" /></el-form-item><el-form-item label="每日 Token"><el-input-number v-model="quotaForm.dailyTokens" :min="0" /></el-form-item><el-form-item label="月预算（微单位）"><el-input-number v-model="quotaForm.monthlyCostMicros" :min="0" /></el-form-item><el-form-item label="最大并发"><el-input-number v-model="quotaForm.maxConcurrency" :min="0" /></el-form-item></div><el-switch v-model="quotaForm.enabled" active-text="启用配额" /></el-form>
      <template #footer><el-button @click="quotaDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingQuota" @click="submitQuota">保存</el-button></template>
    </el-dialog>

    <el-dialog v-model="promptDialogVisible" title="创建 Prompt 版本" width="min(720px, calc(100vw - 32px))" destroy-on-close>
      <el-form label-position="top"><el-form-item label="Prompt 标识"><el-input v-model="promptForm.promptKey" placeholder="例如 asset-draft-v1" /></el-form-item><el-form-item label="Prompt 内容"><el-input v-model="promptForm.content" type="textarea" :rows="8" maxlength="131072" show-word-limit /></el-form-item><el-form-item label="输出 JSON Schema（可选）"><el-input v-model="promptForm.outputSchema" type="textarea" :rows="4" /></el-form-item></el-form>
      <template #footer><el-button @click="promptDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingPrompt" @click="submitPrompt">创建草稿</el-button></template>
    </el-dialog>
  </main>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { Check, Edit, Plus, Refresh, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { activateAIPrompt, createAIPrompt, getAIInvocations, getAIProviders, getAIPrompts, getAIQuotas, getAIUsageSummary, saveAIQuota, updateAIProviders } from '@/plugin/aioperations/api/operations'

defineOptions({ name: 'AIOperations' })

const loading = ref(false)
const savingProviders = ref(false)
const activeTab = ref('providers')
const usage = ref({})
const invocations = ref([])
const invocationTotal = ref(0)
const invocationLoading = ref(false)
const quotas = ref([])
const quotaLoading = ref(false)
const prompts = ref([])
const promptLoading = ref(false)
const quotaDialogVisible = ref(false)
const promptDialogVisible = ref(false)
const savingQuota = ref(false)
const savingPrompt = ref(false)
const providers = reactive(defaultProviders())
const providerList = [{ key: 'openai-compatible', label: 'OpenAI Compatible', hint: '支持 OpenAI Chat Completions 兼容服务' }, { key: 'anthropic', label: 'Anthropic', hint: '使用 Messages API' }]
const invocationSearch = reactive({ page: 1, pageSize: 20, status: '', module: '', provider: '', userId: undefined })
const quotaForm = reactive(defaultQuota())
const promptForm = reactive({ promptKey: '', content: '', outputSchema: '' })

function defaultProvider() { return { enabled: false, 'base-url': '', 'api-key': '', 'api-key-configured': false, 'clear-api-key': false, model: '', 'timeout-seconds': 60, 'input-cost-micros-per-million': 0, 'output-cost-micros-per-million': 0 } }
function defaultProviders() { return { enabled: false, 'allow-private-endpoints': false, 'sensitive-words': [], 'allow-vision-modules': [], 'openai-compatible': defaultProvider(), anthropic: { ...defaultProvider(), 'base-url': 'https://api.anthropic.com' } } }
function defaultQuota() { return { ID: 0, scopeType: 'global', scopeId: 'global', dailyRequests: 0, dailyTokens: 0, monthlyCostMicros: 0, maxConcurrency: 0, enabled: true } }
function applyProviders(value = {}) { const source = { ...defaultProviders(), ...value }; Object.assign(providers, source); for (const key of providerList.map((item) => item.key)) Object.assign(providers[key], defaultProvider(), source[key] || {}) }
function number(value) { return new Intl.NumberFormat('zh-CN').format(Number(value || 0)) }
function money(value) { return `¥${(Number(value || 0) / 1000000).toFixed(4)}` }
function dateTime(value) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—' }
function statusMeta(status) { return ({ success: { label: '成功', type: 'success' }, failed: { label: '失败', type: 'danger' }, blocked: { label: '已阻断', type: 'warning' } }[status] || { label: status || '未知', type: 'info' }) }
function promptStatus(status) { return ({ active: { label: '已激活', type: 'success' }, draft: { label: '草稿', type: 'info' }, retired: { label: '已停用', type: 'warning' } }[status] || { label: status, type: 'info' }) }
async function loadProviders() {
  const res = await getAIProviders()
  if (res.code === 0) {
    applyProviders(res.data)
  } else {
    ElMessage.error(res.msg || '无法读取 Provider 配置')
  }
}
async function loadUsage() { const res = await getAIUsageSummary(); if (res.code === 0) usage.value = res.data || {} }
async function loadInvocations() { invocationLoading.value = true; try { const res = await getAIInvocations(invocationSearch); if (res.code === 0) { invocations.value = res.data?.list || []; invocationTotal.value = res.data?.total || 0 } else ElMessage.error(res.msg || '无法读取调用审计') } finally { invocationLoading.value = false } }
async function loadQuotas() { quotaLoading.value = true; try { const res = await getAIQuotas(); if (res.code === 0) quotas.value = res.data || [] } finally { quotaLoading.value = false } }
async function loadPrompts() { promptLoading.value = true; try { const res = await getAIPrompts(); if (res.code === 0) prompts.value = res.data || [] } finally { promptLoading.value = false } }
async function loadAll() { loading.value = true; try { await Promise.all([loadProviders(), loadUsage(), loadInvocations(), loadQuotas(), loadPrompts()]) } finally { loading.value = false } }
async function saveProviders() {
  savingProviders.value = true
  try {
    const res = await updateAIProviders(providers)
    if (res.code === 0) {
      applyProviders(res.data)
      ElMessage.success(res.msg || 'Provider 配置已保存')
    } else {
      ElMessage.error(res.msg || '保存失败')
    }
  } finally {
    savingProviders.value = false
  }
}
function openQuota(row) { Object.assign(quotaForm, defaultQuota(), row || {}); quotaDialogVisible.value = true }
async function submitQuota() { if (quotaForm.scopeType === 'global' && !quotaForm.scopeId) quotaForm.scopeId = 'global'; savingQuota.value = true; try { const res = await saveAIQuota(quotaForm); if (res.code === 0) { ElMessage.success(res.msg || '配额已保存'); quotaDialogVisible.value = false; await loadQuotas() } else ElMessage.error(res.msg || '保存失败') } finally { savingQuota.value = false } }
function openPrompt() { Object.assign(promptForm, { promptKey: '', content: '', outputSchema: '' }); promptDialogVisible.value = true }
async function submitPrompt() { savingPrompt.value = true; try { const res = await createAIPrompt(promptForm); if (res.code === 0) { ElMessage.success(res.msg || 'Prompt 草稿已创建'); promptDialogVisible.value = false; await loadPrompts() } else ElMessage.error(res.msg || '创建失败') } finally { savingPrompt.value = false } }
async function activatePrompt(row) { try { await ElMessageBox.confirm(`确认激活 ${row.promptKey} 的 V${row.version}？当前活跃版本将退役。`, '激活 Prompt', { type: 'warning' }); const res = await activateAIPrompt({ promptKey: row.promptKey, version: row.version }); if (res.code === 0) { ElMessage.success(res.msg || 'Prompt 已激活'); await loadPrompts() } else ElMessage.error(res.msg || '激活失败') } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.error('激活失败') } }

onMounted(loadAll)
</script>

<style scoped lang="scss">
.summary-band { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); margin-bottom: 14px; overflow: hidden; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.summary-band > div { display: flex; min-width: 0; flex-direction: column; gap: 5px; padding: 18px 20px; border-right: 1px solid var(--na-border); }
.summary-band > div:last-child { border-right: 0; }.summary-band span, .summary-band small { color: var(--na-muted-foreground); font-size: .75rem; }.summary-band strong { overflow: hidden; color: var(--na-foreground); font-size: 1.35rem; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.operation-tabs :deep(.el-tabs__header) { margin-bottom: 14px; }.config-panel, .table-panel { padding: 18px; }.na-panel-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; }.na-panel-header h2 { margin: 0; color: var(--na-foreground); font-size: 1rem; }.na-panel-header p { margin: 5px 0 0; color: var(--na-muted-foreground); font-size: .78rem; }
.provider-form { margin-top: 16px; }.policy-grid, .provider-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }.policy-grid :deep(.el-select) { width: 100%; }.provider-section { min-width: 0; padding: 16px; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-surface-muted); }.provider-section header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 14px; }.provider-section h3 { margin: 0 0 4px; color: var(--na-foreground); font-size: .92rem; }.provider-section header span { color: var(--na-muted-foreground); font-size: .74rem; line-height: 1.45; }.clear-key-control { margin: -6px 0 14px; }.cost-grid, .dialog-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 12px; }.cost-grid :deep(.el-form-item:last-child) { grid-column: span 2; }
.filter-row { display: grid; grid-template-columns: 150px repeat(2, minmax(0, 1fr)) 150px; gap: 10px; margin-bottom: 14px; }.filter-row :deep(.el-input-number) { width: 100%; }.na-pagination { display: flex; justify-content: flex-end; margin-top: 14px; }
@media (max-width: 800px) { .summary-band, .policy-grid, .provider-grid { grid-template-columns: 1fr; }.summary-band > div { border-right: 0; border-bottom: 1px solid var(--na-border); }.summary-band > div:last-child { border-bottom: 0; }.filter-row, .cost-grid, .dialog-grid { grid-template-columns: 1fr; }.cost-grid :deep(.el-form-item:last-child) { grid-column: auto; } }
</style>
