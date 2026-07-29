<template>
  <main class="na-page na-page--list invoice-categories">
    <AppPageHeader title-id="invoice-categories-title" title="分类规则" description="维护统计口径与可解释的规则评分；达到 60 分才自动推荐分类。">
      <template #actions>
        <el-button :icon="Refresh" :loading="currentLoading" @click="refreshCurrent">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">{{ activeView === 'categories' ? '新增分类' : '新增规则' }}</el-button>
      </template>
    </AppPageHeader>

    <div class="view-switch">
      <el-segmented v-model="activeView" :options="viewOptions" @change="refreshCurrent" />
      <p>{{ activeView === 'categories' ? '分类是正式统计口径，停用后不会出现在新的核对表单中。' : '多条命中规则会累加分值，并保留具体命中原因供人工核对。' }}</p>
    </div>

    <section v-if="activeView === 'categories'" class="na-panel collection-panel">
      <div class="collection-toolbar">
        <el-input v-model="categorySearch.keyword" clearable :prefix-icon="Search" placeholder="搜索分类名称或编码" @keyup.enter="categorySubmit" />
        <el-select v-model="categorySearch.enabled" clearable placeholder="全部状态" @change="categorySubmit">
          <el-option label="启用" :value="true" />
          <el-option label="停用" :value="false" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="categorySubmit">查询</el-button>
      </div>
      <div class="category-header category-grid" aria-hidden="true">
        <span>分类</span><span>说明</span><span>排序</span><span>状态</span><span>操作</span>
      </div>
      <div v-if="categoryError && categoryLoaded" class="collection-warning" role="alert">
        <span>刷新失败，当前仍显示上一次成功数据：{{ categoryError }}</span>
        <el-button text :icon="Refresh" @click="categoryLoad">重试</el-button>
      </div>
      <el-skeleton v-if="categoryLoading && !categoryLoaded" :rows="6" animated />
      <el-result v-else-if="categoryError && !categoryLoaded" icon="error" title="分类加载失败" :sub-title="categoryError">
        <template #extra><el-button type="primary" :icon="Refresh" @click="categoryLoad">重新加载</el-button></template>
      </el-result>
      <div v-else-if="categories.length" class="category-list">
        <div v-for="item in categories" :key="item.ID" class="category-row category-grid">
          <div class="category-name"><i :style="{ backgroundColor: item.color }" /><div><strong>{{ item.name }}</strong><small>{{ item.code }}</small><small class="mobile-detail">{{ item.description || '暂无分类说明' }} · 排序 {{ item.sort }}</small></div></div>
          <p>{{ item.description || '暂无分类说明' }}</p>
          <span class="sort-value">{{ item.sort }}</span>
          <el-tag :type="item.enabled ? 'success' : 'info'" effect="light">{{ item.enabled ? '启用' : '停用' }}</el-tag>
          <div class="row-actions">
            <el-button :icon="Edit" type="primary" text :disabled="isPending('category', item.ID)" aria-label="编辑分类" @click="openCategory(item)" />
            <el-button :icon="Delete" type="danger" text :loading="isPending('category', item.ID)" :disabled="isPending('category', item.ID)" aria-label="删除分类" @click="removeCategory(item)" />
          </div>
        </div>
      </div>
      <el-empty v-else description="暂无发票分类" />
      <div v-if="categoryTotal > 0" class="na-pagination">
        <el-pagination v-model:current-page="categorySearch.page" :page-size="categorySearch.pageSize" :total="categoryTotal" layout="total, prev, pager, next" @current-change="categoryChangePage" />
      </div>
    </section>

    <section v-else class="na-panel collection-panel">
      <div class="collection-toolbar collection-toolbar--rules">
        <el-input v-model="ruleSearch.keyword" clearable :prefix-icon="Search" placeholder="搜索规则名称或关键词" @keyup.enter="ruleSubmit" />
        <el-select v-model="ruleSearch.categoryId" clearable placeholder="全部分类" @change="ruleSubmit">
          <el-option v-for="item in categoryOptions" :key="item.ID" :label="item.name" :value="item.ID" />
        </el-select>
        <el-select v-model="ruleSearch.enabled" clearable placeholder="全部状态" @change="ruleSubmit">
          <el-option label="启用" :value="true" /><el-option label="停用" :value="false" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="ruleSubmit">查询</el-button>
      </div>
      <div class="rule-header rule-grid" aria-hidden="true">
        <span>规则</span><span>匹配内容</span><span>目标分类</span><span>分值</span><span>状态</span><span>操作</span>
      </div>
      <div v-if="ruleError && ruleLoaded" class="collection-warning" role="alert">
        <span>刷新失败，当前仍显示上一次成功数据：{{ ruleError }}</span>
        <el-button text :icon="Refresh" @click="ruleLoad">重试</el-button>
      </div>
      <el-skeleton v-if="ruleLoading && !ruleLoaded" :rows="6" animated />
      <el-result v-else-if="ruleError && !ruleLoaded" icon="error" title="规则加载失败" :sub-title="ruleError">
        <template #extra><el-button type="primary" :icon="Refresh" @click="ruleLoad">重新加载</el-button></template>
      </el-result>
      <div v-else-if="rules.length" class="rule-list">
        <div v-for="item in rules" :key="item.ID" class="rule-row rule-grid">
          <div class="rule-name"><strong>{{ item.name }}</strong><small>{{ fieldLabel(item.field) }} · {{ matchLabel(item.matchType) }}</small><small class="mobile-detail">“{{ item.pattern }}” · {{ item.category?.name || '分类不存在' }} · +{{ item.weight }}</small></div>
          <code>{{ item.pattern }}</code>
          <div class="category-name compact"><i :style="{ backgroundColor: item.category?.color }" /><strong>{{ item.category?.name || '分类不存在' }}</strong></div>
          <strong class="score-value">+{{ item.weight }}</strong>
          <el-tag :type="item.enabled ? 'success' : 'info'" effect="light">{{ item.enabled ? '启用' : '停用' }}</el-tag>
          <div class="row-actions">
            <el-button :icon="Edit" type="primary" text :disabled="isPending('rule', item.ID)" aria-label="编辑规则" @click="openRule(item)" />
            <el-button :icon="Delete" type="danger" text :loading="isPending('rule', item.ID)" :disabled="isPending('rule', item.ID)" aria-label="删除规则" @click="removeRule(item)" />
          </div>
        </div>
      </div>
      <el-empty v-else description="暂无分类规则" />
      <div v-if="ruleTotal > 0" class="na-pagination">
        <el-pagination v-model:current-page="ruleSearch.page" :page-size="ruleSearch.pageSize" :total="ruleTotal" layout="total, prev, pager, next" @current-change="ruleChangePage" />
      </div>
    </section>

    <el-drawer v-model="drawerVisible" :size="drawerSize" destroy-on-close :close-on-click-modal="false">
      <template #header>
        <div class="drawer-heading"><span>{{ drawerTitle }}</span><small>{{ editing ? '修改后将影响后续识别与统计' : '建立可维护、可解释的业务口径' }}</small></div>
      </template>

      <el-form v-if="drawerType === 'category'" ref="categoryFormRef" :model="categoryForm" :rules="categoryRules" label-position="top">
        <el-form-item label="分类名称" prop="name"><el-input v-model="categoryForm.name" maxlength="100" /></el-form-item>
        <el-form-item label="分类编码" prop="code"><el-input v-model="categoryForm.code" maxlength="50" placeholder="例如 SOFTWARE" /></el-form-item>
        <el-form-item label="分类说明"><el-input v-model="categoryForm.description" type="textarea" :rows="3" maxlength="500" /></el-form-item>
        <div class="form-grid">
          <el-form-item label="标识颜色"><el-color-picker v-model="categoryForm.color" :predefine="predefinedColors" /></el-form-item>
          <el-form-item label="排序"><el-input-number v-model="categoryForm.sort" :min="0" :max="999" /></el-form-item>
        </div>
        <el-form-item label="启用状态"><el-switch v-model="categoryForm.enabled" inline-prompt active-text="启用" inactive-text="停用" /></el-form-item>
      </el-form>

      <el-form v-else ref="ruleFormRef" :model="ruleForm" :rules="ruleFormRules" label-position="top">
        <el-form-item label="规则名称" prop="name"><el-input v-model="ruleForm.name" maxlength="120" /></el-form-item>
        <el-form-item label="目标分类" prop="categoryId">
          <el-select v-model="ruleForm.categoryId" filterable placeholder="选择目标分类">
            <el-option v-for="item in categoryOptions" :key="item.ID" :label="item.name" :value="item.ID" />
          </el-select>
        </el-form-item>
        <div class="form-grid">
          <el-form-item label="匹配字段" prop="field">
            <el-select v-model="ruleForm.field"><el-option v-for="item in fieldOptions" :key="item.value" :label="item.label" :value="item.value" /></el-select>
          </el-form-item>
          <el-form-item label="匹配方式" prop="matchType">
            <el-select v-model="ruleForm.matchType"><el-option label="包含" value="contains" /><el-option label="完全一致" value="exact" /></el-select>
          </el-form-item>
        </div>
        <el-form-item label="匹配内容" prop="pattern"><el-input v-model="ruleForm.pattern" maxlength="300" /></el-form-item>
        <div class="score-editor">
          <el-form-item label="命中分值" prop="weight"><el-slider v-model="ruleForm.weight" :min="1" :max="100" show-input /></el-form-item>
          <p>同一分类的规则分值会累加；总分达到 60 时才自动填入分类建议。</p>
        </div>
        <div class="form-grid">
          <el-form-item label="执行优先级"><el-input-number v-model="ruleForm.priority" :min="0" :max="999" /></el-form-item>
          <el-form-item label="启用状态"><el-switch v-model="ruleForm.enabled" inline-prompt active-text="启用" inactive-text="停用" /></el-form-item>
        </div>
      </el-form>

      <template #footer>
        <div class="drawer-actions"><el-button @click="drawerVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></div>
      </template>
    </el-drawer>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { Delete, Edit, Plus, Refresh, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import {
  createInvoiceCategory, createInvoiceRule, deleteInvoiceCategory, deleteInvoiceRule,
  getInvoiceCategoryList, getInvoiceCategoryOptions, getInvoiceRuleList,
  updateInvoiceCategory, updateInvoiceRule
} from '@/plugin/invoice/api/invoice'
import { usePagedList } from '@/hooks/usePagedList'
import { useAppStore } from '@/pinia/modules/app'

defineOptions({ name: 'InvoiceCategories' })

const appStore = useAppStore()
const activeView = ref('categories')
const drawerVisible = ref(false)
const drawerType = ref('category')
const editing = ref(false)
const saving = ref(false)
const pendingActions = ref(new Set())
const categoryFormRef = ref()
const ruleFormRef = ref()
const categoryOptions = ref([])
const viewOptions = [{ label: '分类口径', value: 'categories' }, { label: '智能规则', value: 'rules' }]
const fieldOptions = [
  { value: 'all', label: '全部识别内容' }, { value: 'seller', label: '销售方名称' },
  { value: 'item', label: '商品或服务' }, { value: 'type', label: '发票类型' }, { value: 'raw', label: '识别原文' }
]
const predefinedColors = ['#2563EB', '#7C3AED', '#0891B2', '#4F46E5', '#D97706', '#059669', '#DC2626', '#64748B']

const emptyCategory = () => ({ ID: 0, name: '', code: '', description: '', color: '#2563EB', sort: 10, enabled: true })
const emptyRule = () => ({ ID: 0, name: '', field: 'all', matchType: 'contains', pattern: '', weight: 70, priority: 0, enabled: true, categoryId: undefined })
const categoryForm = reactive(emptyCategory())
const ruleForm = reactive(emptyRule())
const categoryRules = { name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }], code: [{ required: true, message: '请输入分类编码', trigger: 'blur' }] }
const ruleFormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请选择目标分类', trigger: 'change' }],
  field: [{ required: true, message: '请选择匹配字段', trigger: 'change' }],
  matchType: [{ required: true, message: '请选择匹配方式', trigger: 'change' }],
  pattern: [{ required: true, message: '请输入匹配内容', trigger: 'blur' }],
  weight: [{ required: true, message: '请设置规则分值', trigger: 'change' }]
}

const drawerSize = computed(() => appStore.drawerSize === '100%' ? '100%' : '520px')
const drawerTitle = computed(() => `${editing.value ? '编辑' : '新增'}${drawerType.value === 'category' ? '发票分类' : '分类规则'}`)

const {
  search: categorySearch, items: categories, total: categoryTotal, loading: categoryLoading,
  loaded: categoryLoaded, error: categoryError,
  load: categoryLoad, submit: categorySubmit, changePage: categoryChangePage
} = usePagedList({ defaults: { page: 1, pageSize: 20, keyword: '', enabled: undefined }, request: getInvoiceCategoryList })
const {
  search: ruleSearch, items: rules, total: ruleTotal, loading: ruleLoading,
  loaded: ruleLoaded, error: ruleError,
  load: ruleLoad, submit: ruleSubmit, changePage: ruleChangePage
} = usePagedList({ defaults: { page: 1, pageSize: 20, keyword: '', categoryId: undefined, enabled: undefined }, request: getInvoiceRuleList })
const currentLoading = computed(() => activeView.value === 'categories' ? categoryLoading.value : ruleLoading.value)

const loadCategoryOptions = async () => {
  const res = await getInvoiceCategoryOptions()
  if (res.code === 0) categoryOptions.value = res.data || []
}
const refreshCurrent = () => activeView.value === 'categories' ? categoryLoad() : Promise.all([ruleLoad(), loadCategoryOptions()])
const fieldLabel = (value) => fieldOptions.find((item) => item.value === value)?.label || value
const matchLabel = (value) => value === 'exact' ? '完全一致' : '包含'
const pendingKey = (type, id) => `${type}:${Number(id)}`
const isPending = (type, id) => pendingActions.value.has(pendingKey(type, id))
const setPending = (type, id, pending) => {
  const next = new Set(pendingActions.value)
  const key = pendingKey(type, id)
  if (pending) next.add(key)
  else next.delete(key)
  pendingActions.value = next
}

const openCreate = () => {
  editing.value = false
  if (activeView.value === 'categories') {
    drawerType.value = 'category'
    Object.assign(categoryForm, emptyCategory())
  } else {
    drawerType.value = 'rule'
    Object.assign(ruleForm, emptyRule())
    loadCategoryOptions()
  }
  drawerVisible.value = true
}
const openCategory = (item) => {
  editing.value = true
  drawerType.value = 'category'
  Object.assign(categoryForm, emptyCategory(), JSON.parse(JSON.stringify(item)))
  drawerVisible.value = true
}
const openRule = (item) => {
  editing.value = true
  drawerType.value = 'rule'
  Object.assign(ruleForm, emptyRule(), JSON.parse(JSON.stringify(item)))
  loadCategoryOptions()
  drawerVisible.value = true
}

const save = async () => {
  if (saving.value) return
  const formInstance = drawerType.value === 'category' ? categoryFormRef.value : ruleFormRef.value
  const valid = await formInstance?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const isCategory = drawerType.value === 'category'
    const payload = isCategory ? categoryForm : ruleForm
    const request = isCategory
      ? (editing.value ? updateInvoiceCategory : createInvoiceCategory)
      : (editing.value ? updateInvoiceRule : createInvoiceRule)
    const res = await request({ ...payload })
    if (res.code === 0) {
      ElMessage.success(`${isCategory ? '分类' : '规则'}已保存`)
      drawerVisible.value = false
      await Promise.all([refreshCurrent(), loadCategoryOptions()])
    }
  } finally {
    saving.value = false
  }
}

const removeCategory = async (item) => {
  if (isPending('category', item.ID)) return
  setPending('category', item.ID, true)
  try {
    await ElMessageBox.confirm(`确定删除分类“${item.name}”吗？`, '删除分类', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' })
    const res = await deleteInvoiceCategory({ id: item.ID })
    if (res.code === 0) { ElMessage.success('分类已删除'); await Promise.all([categoryLoad(), loadCategoryOptions()]) }
  } catch (action) {
    if (action !== 'cancel' && action !== 'close') ElMessage.error(action?.message || '分类删除失败')
  } finally {
    setPending('category', item.ID, false)
  }
}
const removeRule = async (item) => {
  if (isPending('rule', item.ID)) return
  setPending('rule', item.ID, true)
  try {
    await ElMessageBox.confirm(`确定删除规则“${item.name}”吗？`, '删除规则', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' })
    const res = await deleteInvoiceRule({ id: item.ID })
    if (res.code === 0) { ElMessage.success('规则已删除'); await ruleLoad() }
  } catch (action) {
    if (action !== 'cancel' && action !== 'close') ElMessage.error(action?.message || '规则删除失败')
  } finally {
    setPending('rule', item.ID, false)
  }
}

onMounted(categoryLoad)
</script>

<style scoped lang="scss">
.view-switch { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 18px; margin-bottom: 14px; }
.view-switch p { margin: 0; color: var(--na-muted-foreground); font-size: .75rem; text-align: right; }
.collection-panel { min-height: 480px; overflow: hidden; }
.collection-toolbar { display: grid; grid-template-columns: minmax(220px, 1fr) 160px auto; gap: 8px; padding: 12px 16px; border-bottom: 1px solid var(--na-border); }
.collection-toolbar--rules { grid-template-columns: minmax(220px, 1fr) 180px 140px auto; }
.collection-warning { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 12px; padding: 9px 14px; border-bottom: 1px solid var(--na-border); background: var(--na-warning-soft); font-size: .75rem; }
.collection-warning span { min-width: 0; overflow-wrap: anywhere; }
.category-grid, .rule-grid { display: grid; min-width: 0; align-items: center; gap: 12px; }
.category-grid { grid-template-columns: minmax(170px, 1fr) minmax(200px, 1.7fr) 70px 80px 92px; }
.rule-grid { grid-template-columns: minmax(160px, 1.2fr) minmax(130px, 1fr) minmax(120px, .8fr) 64px 76px 92px; }
.category-header, .rule-header { min-height: 42px; padding: 0 16px; border-bottom: 1px solid var(--na-border); background: var(--na-table-header); color: var(--na-muted-foreground); font-size: .6875rem; font-weight: 620; }
.category-row, .rule-row { min-height: 60px; padding: 8px 16px; border-bottom: 1px solid var(--na-border); }
.category-row:hover, .rule-row:hover { background: var(--na-table-hover); }
.category-name { display: flex; min-width: 0; align-items: center; gap: 9px; }
.category-name i { width: 9px; height: 9px; flex: 0 0 9px; border-radius: 50%; background: var(--na-border-strong); }
.category-name > div, .rule-name { display: flex; min-width: 0; flex-direction: column; gap: 4px; }
.category-name strong, .rule-name strong { overflow: hidden; font-size: .8125rem; text-overflow: ellipsis; white-space: nowrap; }
.category-name small, .rule-name small { color: var(--na-muted-foreground); font-size: .6875rem; }
.category-name .mobile-detail, .rule-name .mobile-detail { display: none; }
.category-row p { margin: 0; overflow: hidden; color: var(--na-muted-foreground); font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.sort-value, .score-value { font-size: .75rem; font-variant-numeric: tabular-nums; }
.rule-row code { overflow: hidden; padding: 4px 7px; border-radius: 6px; background: var(--na-muted); color: var(--na-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.category-name.compact strong { font-size: .75rem; }
.score-value { color: var(--na-primary); }
.row-actions { display: flex; justify-content: flex-end; gap: 2px; }
.row-actions :deep(.el-button) { width: 30px; min-width: 30px; min-height: 30px; padding: 0; }
.row-actions :deep(.el-button + .el-button) { margin-left: 0; }
.drawer-heading { display: flex; flex-direction: column; gap: 4px; }
.drawer-heading span { color: var(--na-foreground); font-size: 1.125rem; font-weight: 650; }
.drawer-heading small { color: var(--na-muted-foreground); font-size: .75rem; font-weight: 400; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-grid :deep(.el-select), .form-grid :deep(.el-input-number), :deep(.el-select) { width: 100%; }
.score-editor { margin-bottom: 18px; padding: 14px; border: 1px solid var(--na-border); border-radius: 10px; background: var(--na-muted); }
.score-editor :deep(.el-form-item) { margin-bottom: 8px; }
.score-editor p { margin: 0; color: var(--na-muted-foreground); font-size: .6875rem; line-height: 1.55; }
.drawer-actions { display: flex; justify-content: flex-end; gap: 8px; }

@media (max-width: 1020px) {
  .category-grid { grid-template-columns: minmax(160px, 1fr) minmax(180px, 1.5fr) 76px 92px; }
  .category-grid > :nth-child(3) { display: none; }
  .rule-grid { grid-template-columns: minmax(150px, 1.2fr) minmax(120px, 1fr) minmax(110px, .8fr) 70px 92px; }
  .rule-grid > :nth-child(4) { display: none; }
}
@media (max-width: 720px) {
  .view-switch { align-items: stretch; flex-direction: column; }
  .view-switch p { text-align: left; }
  .collection-toolbar, .collection-toolbar--rules { grid-template-columns: 1fr; }
  .category-header, .rule-header { display: none; }
  .category-row, .rule-row { grid-template-columns: minmax(0, 1fr) auto; grid-template-rows: auto auto; gap: 6px 10px; padding: 12px 14px; }
  .category-row > :nth-child(2), .category-row > :nth-child(3), .rule-row > :nth-child(2), .rule-row > :nth-child(4) { display: none; }
  .category-name .mobile-detail, .rule-name .mobile-detail { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .category-row :deep(.el-tag), .rule-row :deep(.el-tag) { grid-column: 2; grid-row: 1; }
  .row-actions { grid-column: 2; grid-row: 2; }
  .form-grid { grid-template-columns: 1fr; }
}
@media (pointer: coarse) {
  .row-actions :deep(.el-button) { width: 44px; min-width: 44px; min-height: 44px; }
}
</style>
