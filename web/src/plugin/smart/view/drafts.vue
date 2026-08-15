<template>
  <main class="na-page na-page--list smart-drafts-page">
    <AppPageHeader title-id="smart-drafts-title" title="智能草稿" description="从公告提取日程，或生成资产业务单草稿；确认后才写入正式业务记录。">
      <template #actions><el-button :icon="Refresh" :loading="loading" @click="loadDrafts">刷新</el-button></template>
    </AppPageHeader>
    <div class="draft-layout">
      <section class="na-panel create-panel"><header class="panel-header"><h2>生成草稿</h2><el-tag type="warning" effect="plain">需人工确认</el-tag></header><el-tabs v-model="activeTab">
        <el-tab-pane label="公告转日程" name="schedule"><el-form label-position="top" @submit.prevent><el-form-item label="公告 ID"><el-input-number v-model="announcementId" :min="1" :controls="false" placeholder="输入已发布公告 ID" /></el-form-item><el-button type="primary" :icon="MagicStick" :loading="creating" @click="extract">提取日程草稿</el-button></el-form></el-tab-pane>
        <el-tab-pane label="资产业务单" name="operation"><el-form label-position="top"><div class="form-grid"><el-form-item label="业务类型"><el-select v-model="operation.operationType" clearable placeholder="可从说明自动识别"><el-option label="入库" value="inbound" /><el-option label="领用" value="issue" /><el-option label="调拨" value="transfer" /><el-option label="归还" value="return" /><el-option label="维修" value="maintenance" /><el-option label="报废" value="scrap" /></el-select></el-form-item><el-form-item label="业务日期"><el-date-picker v-model="operation.businessDate" type="date" value-format="YYYY-MM-DD" /></el-form-item></div><el-form-item label="资产候选"><el-select v-model="selectedAssetIDs" multiple filterable remote reserve-keyword :remote-method="loadCandidates" :disabled="!operation.operationType" placeholder="先选择业务类型，再搜索资产"><el-option v-for="item in candidateOptions" :key="item.ID" :label="`${item.assetCode} · ${item.name}`" :value="item.ID" /></el-select></el-form-item><el-form-item label="资产 ID（逗号分隔，可选）"><el-input v-model="assetIDs" placeholder="也可以直接输入，例如 12,15" /></el-form-item><el-form-item label="目标位置"><el-input v-model="operation.targetLocation" /></el-form-item><el-form-item label="领用人"><el-input v-model="operation.targetCustodian" /></el-form-item><el-form-item label="原因或说明"><el-input v-model="operation.instruction" type="textarea" :rows="3" maxlength="500" show-word-limit /></el-form-item><el-button type="primary" :icon="MagicStick" :loading="creating" @click="createDraft">生成业务草稿</el-button></el-form></el-tab-pane>
      </el-tabs></section>
      <section class="na-panel list-panel">
        <header class="panel-header"><div><h2>待确认草稿</h2><p>确认日程会写入个人日历；确认业务单只创建草稿，不会自动提交。</p></div><el-select v-model="draftType" clearable placeholder="全部类型" @change="loadDrafts"><el-option label="日程" value="schedule" /><el-option label="业务单" value="operation" /></el-select></header>
        <el-table v-loading="loading" :data="drafts" row-key="ID">
          <el-table-column type="expand" width="44"><template #default="{ row }"><div class="draft-detail"><div v-for="item in draftDetails(row)" :key="item.label" class="draft-detail-item"><span>{{ item.label }}</span><strong>{{ item.value }}</strong></div></div></template></el-table-column>
          <el-table-column prop="draftType" label="类型" width="90"><template #default="{ row }">{{ row.draftType === 'schedule' ? '日程' : '业务单' }}</template></el-table-column>
          <el-table-column prop="confidence" label="置信度" width="100"><template #default="{ row }">{{ Math.round(Number(row.confidence || 0) * 100) }}%</template></el-table-column>
          <el-table-column prop="CreatedAt" label="创建时间" width="170"><template #default="{ row }">{{ formatTime(row.CreatedAt) }}</template></el-table-column>
          <el-table-column label="内容" min-width="260" show-overflow-tooltip><template #default="{ row }">{{ draftTitle(row) }}</template></el-table-column>
          <el-table-column label="操作" width="110" fixed="right"><template #default="{ row }"><el-button v-if="row.status === 'draft' && !isExpired(row)" text type="primary" @click="accept(row)">确认写入</el-button><el-tag v-else :type="draftStatusType(row)" size="small">{{ draftStatusLabel(row) }}</el-tag></template></el-table-column>
          <template #empty>
            <AppEmptyState
              compact
              :title="draftType ? '当前类型没有待确认草稿' : '还没有智能草稿'"
              description="可从已发布公告提取日程，或根据资产业务说明生成待确认业务单。"
              :highlights="['生成结果需人工确认', '高风险业务不会自动提交', '草稿保留来源与有效期']"
            >
              <template #actions>
                <el-button @click="activeTab = 'schedule'">公告转日程</el-button>
                <el-button type="primary" @click="activeTab = 'operation'">生成业务草稿</el-button>
              </template>
            </AppEmptyState>
          </template>
        </el-table>
      </section>
    </div>
  </main>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { MagicStick, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { acceptSmartDraft, createOperationDraft, extractAnnouncementSchedule, getOperationAssetCandidates, getSmartDrafts } from '@/plugin/smart/api/smart'

defineOptions({ name: 'SmartDrafts' })
const loading = ref(false); const creating = ref(false); const activeTab = ref('schedule'); const announcementId = ref(); const assetIDs = ref(''); const selectedAssetIDs = ref([]); const candidateOptions = ref([]); const draftType = ref(''); const drafts = ref([]); const operation = reactive({ operationType: '', businessDate: '', targetLocation: '', targetCustodian: '', instruction: '' })
function formatTime(value) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—' }
function draftTitle(row) { const p = row.payload || {}; return p.title || p.instruction || `${p.operationType || '智能'} 草稿` }
function operationLabel(value) { return ({ inbound: '入库', issue: '领用', transfer: '调拨', return: '归还', maintenance: '维修', scrap: '报废' })[value] || value || '—' }
function isExpired(row) { return Boolean(row.expiresAt && new Date(row.expiresAt).getTime() <= Date.now()) }
function draftStatusLabel(row) { if (isExpired(row) && row.status === 'draft') return '已过期'; return ({ accepted: '已确认', processing: '处理中', discarded: '已放弃' })[row.status] || row.status }
function draftStatusType(row) { if (isExpired(row) && row.status === 'draft') return 'danger'; return ({ accepted: 'success', processing: 'warning', discarded: 'info' })[row.status] || 'info' }
function draftDetails(row) { const p = row.payload || {}; const expires = formatTime(row.expiresAt); if (row.draftType === 'schedule') return [{ label: '来源', value: row.sourceId ? `公告 #${row.sourceId}` : '公告' }, { label: '标题', value: p.title || '—' }, { label: '日期时间', value: `${p.date || '未识别'} ${p.time || ''}`.trim() }, { label: '地点', value: p.location || '—' }, { label: '待办', value: Array.isArray(p.todos) && p.todos.length ? p.todos.join('；') : '—' }, { label: '公告原文', value: p.note || '—' }, { label: '有效期', value: expires }]; return [{ label: '业务类型', value: operationLabel(p.operationType) }, { label: '资产 ID', value: Array.isArray(p.assetIds) ? p.assetIds.join(', ') : '—' }, { label: '业务日期', value: p.businessDate || '—' }, { label: '目标位置', value: p.targetLocation || '—' }, { label: '领用人', value: p.targetCustodian || '—' }, { label: '原因', value: p.reason || p.instruction || '—' }, { label: '备注', value: p.remarks || '—' }, { label: '有效期', value: expires }] }
async function loadDrafts() { loading.value = true; try { const res = await getSmartDrafts({ draftType: draftType.value || undefined }); if (res.code === 0) drafts.value = res.data || [] } finally { loading.value = false } }
async function extract() { if (!announcementId.value) return; creating.value = true; try { const res = await extractAnnouncementSchedule({ announcementId: announcementId.value }); if (res.code === 0) { ElMessage.success(res.msg || '草稿已生成'); await loadDrafts() } else ElMessage.error(res.msg || '提取失败') } finally { creating.value = false } }
async function loadCandidates(keyword = '') { if (!operation.operationType) return; const res = await getOperationAssetCandidates({ operationType: operation.operationType, keyword }); if (res.code === 0) candidateOptions.value = res.data || [] }
async function createDraft() { const manualIDs = assetIDs.value.split(',').map((item) => Number(item.trim())).filter((item) => item > 0); const ids = [...new Set([...selectedAssetIDs.value, ...manualIDs])]; if (!ids.length || !operation.instruction.trim()) { ElMessage.warning('请选择或填写资产，并补充业务说明'); return } creating.value = true; try { const res = await createOperationDraft({ ...operation, assetIds: ids }); if (res.code === 0) { ElMessage.success(res.msg || '草稿已生成'); await loadDrafts() } else ElMessage.error(res.msg || '生成失败') } finally { creating.value = false } }
watch(() => operation.operationType, () => { selectedAssetIDs.value = []; candidateOptions.value = []; loadCandidates() })
async function accept(row) { try { await ElMessageBox.confirm('确认后会写入日历或创建资产业务草稿，仍不会自动提交高风险动作。', '确认智能草稿', { type: 'warning' }); const res = await acceptSmartDraft({ id: row.ID }); if (res.code === 0) { ElMessage.success(res.msg || '已确认'); await loadDrafts() } else ElMessage.error(res.msg || '确认失败') } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.error('确认失败') } }
onMounted(loadDrafts)
</script>

<style scoped lang="scss">
.draft-layout { display: grid; grid-template-columns: 360px minmax(0, 1fr); gap: 14px; }.na-panel { min-width: 0; padding: 16px; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-card); box-shadow: var(--na-shadow-sm); }.panel-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 14px; }.panel-header h2 { margin: 0; font-size: .95rem; }.panel-header p { margin: 5px 0 0; color: var(--na-muted-foreground); font-size: .75rem; }.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0 12px; }.form-grid .el-date-editor, .form-grid .el-select, .create-panel .el-input-number { width: 100%; }.list-panel { overflow: hidden; }.list-panel .el-select { width: 130px; }.draft-detail { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 18px; padding: 4px 16px 12px 52px; }.draft-detail-item { display: grid; grid-template-columns: 86px minmax(0, 1fr); gap: 8px; padding: 8px 0; border-bottom: 1px solid var(--na-border); }.draft-detail-item span { color: var(--na-muted-foreground); font-size: .75rem; }.draft-detail-item strong { overflow-wrap: anywhere; font-size: .78rem; font-weight: 500; white-space: pre-wrap; }
@media (max-width: 1000px) { .draft-layout { grid-template-columns: 1fr; } }
@media (max-width: 600px) { .form-grid, .draft-detail { grid-template-columns: 1fr; } .panel-header { flex-direction: column; } .list-panel .el-select { width: 100%; }.draft-detail { padding-left: 14px; }.draft-detail-item { grid-template-columns: 72px minmax(0, 1fr); } }
</style>
