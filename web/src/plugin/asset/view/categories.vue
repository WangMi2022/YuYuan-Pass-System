<template>
  <main class="category-page">
    <section class="page-heading" aria-labelledby="category-title">
      <div>
        <p class="eyebrow">CLASSIFICATION & LOCATION</p>
        <h1 id="category-title">分类管理</h1>
        <p>统一维护资产分类及各业务环节可选位置。</p>
      </div>
      <el-button type="primary" :icon="Plus" size="large" @click="openCreate">
        {{ activeTab === 'category' ? '新增分类' : `新增${activeLocation.label}` }}
      </el-button>
    </section>

    <section class="category-panel">
      <el-tabs v-model="activeTab" class="manage-tabs" @tab-change="changeTab">
        <el-tab-pane label="资产分类" name="category" />
        <el-tab-pane v-for="item in locationTabs" :key="item.type" :label="item.label" :name="item.type" />
      </el-tabs>

      <header class="panel-toolbar">
        <el-input
          v-if="activeTab === 'category'"
          v-model="categorySearch.keyword"
          :prefix-icon="Search"
          clearable
          placeholder="搜索分类名称或编码"
          @keyup.enter="submitSearch"
          @clear="submitSearch"
        />
        <el-input
          v-else
          v-model="locationSearch.keyword"
          :prefix-icon="Search"
          clearable
          :placeholder="`搜索${activeLocation.label}名称、编码或说明`"
          @keyup.enter="submitSearch"
          @clear="submitSearch"
        />
        <div class="toolbar-actions">
          <el-button type="primary" :icon="Search" @click="submitSearch">查询</el-button>
          <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
        </div>
      </header>

      <el-table v-if="activeTab === 'category'" v-loading="loading" :data="tableData" row-key="ID" stripe>
        <el-table-column label="分类" min-width="220">
          <template #default="{ row }">
            <div class="category-name">
              <span class="color-mark" :style="{ backgroundColor: row.color }" />
              <div><strong>{{ row.name }}</strong><small>{{ row.code }}</small></div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="分类说明" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '—' }}</template>
        </el-table-column>
        <el-table-column label="资产档案" width="120" align="right">
          <template #default="{ row }"><strong class="metric">{{ row.assetKinds }}</strong> 种</template>
        </el-table-column>
        <el-table-column label="资产数量" width="120" align="right">
          <template #default="{ row }"><strong class="metric">{{ row.quantity }}</strong> 件</template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="90" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link :icon="Delete" :disabled="row.assetKinds > 0" @click="removeCategory(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无资产分类" /></template>
      </el-table>

      <el-table v-else v-loading="loading" :data="tableData" row-key="ID" stripe>
        <el-table-column label="位置名称" min-width="220">
          <template #default="{ row }">
            <div class="location-name">
              <el-icon><Location /></el-icon>
              <strong>{{ row.name }}</strong>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="code" label="位置编码" min-width="140">
          <template #default="{ row }"><span class="code-text">{{ row.code || '—' }}</span></template>
        </el-table-column>
        <el-table-column prop="description" label="位置说明" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '—' }}</template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="90" align="center" />
        <el-table-column label="启用" width="90" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" :loading="switchingId === row.ID" :aria-label="`${row.name}启用状态`" @change="toggleLocation(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link :icon="Delete" @click="removeLocation(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty :description="`暂无${activeLocation.label}`" /></template>
      </el-table>

      <footer class="pagination-wrap">
        <el-pagination
          v-model:current-page="currentSearch.page"
          v-model:page-size="currentSearch.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          size="small"
          layout="total, sizes, prev, pager, next"
          @current-change="loadCurrent"
          @size-change="sizeChanged"
        />
      </footer>
    </section>

    <el-dialog v-model="categoryDialogVisible" :title="editing ? '编辑资产分类' : '新增资产分类'" width="min(92vw, 560px)" :close-on-click-modal="false">
      <el-form ref="categoryFormRef" :model="categoryForm" :rules="categoryRules" label-position="top">
        <div class="form-grid">
          <el-form-item label="分类名称" prop="name">
            <el-input v-model="categoryForm.name" maxlength="100" placeholder="例如：电脑整机" />
          </el-form-item>
          <el-form-item label="分类编码" prop="code">
            <el-input v-model="categoryForm.code" maxlength="50" placeholder="例如：IT-COMPUTER" />
          </el-form-item>
          <el-form-item label="展示颜色" prop="color">
            <div class="color-field">
              <el-color-picker v-model="categoryForm.color" />
              <el-input v-model="categoryForm.color" maxlength="20" />
            </div>
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="categoryForm.sort" :min="0" :max="9999" controls-position="right" />
          </el-form-item>
        </div>
        <el-form-item label="分类说明">
          <el-input v-model="categoryForm.description" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="说明该分类包含的资产范围" />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="categoryForm.enabled" inline-prompt active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveCategory">保存分类</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="locationDialogVisible" :title="editing ? `编辑${activeLocation.label}` : `新增${activeLocation.label}`" width="min(92vw, 560px)" :close-on-click-modal="false">
      <el-form ref="locationFormRef" :model="locationForm" :rules="locationRules" label-position="top">
        <div class="form-grid">
          <el-form-item label="位置名称" prop="name">
            <el-input v-model="locationForm.name" maxlength="150" :placeholder="activeLocation.example" />
          </el-form-item>
          <el-form-item label="位置编码" prop="code">
            <el-input v-model="locationForm.code" maxlength="50" placeholder="例如：AREA-A，可选" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="locationForm.sort" :min="0" :max="9999" controls-position="right" />
          </el-form-item>
          <el-form-item label="启用状态">
            <el-switch v-model="locationForm.enabled" inline-prompt active-text="启用" inactive-text="停用" />
          </el-form-item>
        </div>
        <el-form-item label="位置说明">
          <el-input v-model="locationForm.description" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="填写适用范围或交接说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="locationDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveLocation">保存位置</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Location, Plus, Refresh, Search } from '@element-plus/icons-vue'
import { createCategory, deleteCategory, getCategoryList, updateCategory } from '@/plugin/asset/api/category'
import { createLocation, deleteLocation, getLocationList, updateLocation } from '@/plugin/asset/api/location'

defineOptions({ name: 'AssetCategories' })

const locationTabs = [
  { type: 'inbound', label: '入库位置', example: '例如：资产仓库 A 区' },
  { type: 'usage', label: '使用位置', example: '例如：研发中心 6F' },
  { type: 'transfer', label: '调入位置', example: '例如：二号办公楼 3F' },
  { type: 'return', label: '归还位置', example: '例如：行政物资库' },
  { type: 'maintenance', label: '维修位置', example: '例如：设备维修中心' },
  { type: 'disposal', label: '处置位置', example: '例如：报废暂存区' }
]

const emptyCategoryForm = () => ({ ID: 0, name: '', code: '', description: '', color: '#6366f1', sort: 0, enabled: true })
const emptyLocationForm = () => ({ ID: 0, name: '', type: activeTab.value, code: '', description: '', sort: 0, enabled: true })
const activeTab = ref('category')
const activeLocation = computed(() => locationTabs.find((item) => item.type === activeTab.value) || locationTabs[0])
const categorySearch = reactive({ page: 1, pageSize: 10, keyword: '' })
const locationSearch = reactive({ page: 1, pageSize: 10, keyword: '' })
const currentSearch = computed(() => activeTab.value === 'category' ? categorySearch : locationSearch)
const tableData = ref([])
const total = ref(0)
const loading = ref(false)
const categoryDialogVisible = ref(false)
const locationDialogVisible = ref(false)
const editing = ref(false)
const saving = ref(false)
const switchingId = ref(0)
const categoryFormRef = ref()
const locationFormRef = ref()
const categoryForm = ref(emptyCategoryForm())
const locationForm = ref({ ID: 0, name: '', type: 'inbound', code: '', description: '', sort: 0, enabled: true })
const categoryRules = {
  name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入分类编码', trigger: 'blur' },
    { pattern: /^[A-Za-z0-9_-]+$/, message: '仅支持字母、数字、中划线和下划线', trigger: 'blur' }
  ],
  color: [{ required: true, message: '请选择展示颜色', trigger: 'change' }]
}
const locationRules = {
  name: [{ required: true, message: '请输入位置名称', trigger: 'blur' }],
  code: [{ pattern: /^[A-Za-z0-9_-]*$/, message: '仅支持字母、数字、中划线和下划线', trigger: 'blur' }]
}

const loadCategories = async () => {
  loading.value = true
  try {
    const res = await getCategoryList(categorySearch)
    if (res.code === 0) {
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
    }
  } finally { loading.value = false }
}

const loadLocations = async () => {
  loading.value = true
  try {
    const res = await getLocationList({ ...locationSearch, type: activeTab.value })
    if (res.code === 0) {
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
    }
  } finally { loading.value = false }
}

const loadCurrent = () => activeTab.value === 'category' ? loadCategories() : loadLocations()
const changeTab = () => {
  tableData.value = []
  total.value = 0
  locationSearch.page = 1
  locationSearch.keyword = ''
  loadCurrent()
}
const submitSearch = () => { currentSearch.value.page = 1; loadCurrent() }
const resetSearch = () => { Object.assign(currentSearch.value, { page: 1, pageSize: 10, keyword: '' }); loadCurrent() }
const sizeChanged = () => { currentSearch.value.page = 1; loadCurrent() }

const openCreate = () => {
  editing.value = false
  if (activeTab.value === 'category') {
    categoryForm.value = emptyCategoryForm()
    categoryDialogVisible.value = true
    return
  }
  locationForm.value = emptyLocationForm()
  locationDialogVisible.value = true
}

const openEdit = (row) => {
  editing.value = true
  if (activeTab.value === 'category') {
    categoryForm.value = { ...emptyCategoryForm(), ...JSON.parse(JSON.stringify(row)) }
    categoryDialogVisible.value = true
    return
  }
  locationForm.value = { ...emptyLocationForm(), ...JSON.parse(JSON.stringify(row)), type: activeTab.value }
  locationDialogVisible.value = true
}

const saveCategory = async () => {
  const valid = await categoryFormRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const res = editing.value ? await updateCategory(categoryForm.value) : await createCategory(categoryForm.value)
    if (res.code === 0) {
      ElMessage.success(editing.value ? '分类已更新' : '分类已创建')
      categoryDialogVisible.value = false
      loadCategories()
    }
  } finally { saving.value = false }
}

const saveLocation = async () => {
  const valid = await locationFormRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const data = { ...locationForm.value, type: activeTab.value }
    const res = editing.value ? await updateLocation(data) : await createLocation(data)
    if (res.code === 0) {
      ElMessage.success(editing.value ? '位置已更新' : '位置已创建')
      locationDialogVisible.value = false
      loadLocations()
    }
  } finally { saving.value = false }
}

const toggleLocation = async (row) => {
  switchingId.value = row.ID
  try {
    const res = await updateLocation({ ...row, type: activeTab.value })
    if (res.code === 0) ElMessage.success(row.enabled ? '位置已启用' : '位置已停用')
    else row.enabled = !row.enabled
  } catch {
    row.enabled = !row.enabled
  } finally { switchingId.value = 0 }
}

const removeCategory = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除分类“${row.name}”吗？`, '删除确认', { type: 'warning' })
  } catch { return }
  const res = await deleteCategory({ id: row.ID })
  if (res.code === 0) { ElMessage.success('分类已删除'); loadCategories() }
}

const removeLocation = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除${activeLocation.value.label}“${row.name}”吗？历史单据中的位置记录不受影响。`, '删除确认', { type: 'warning' })
  } catch { return }
  const res = await deleteLocation({ id: row.ID })
  if (res.code === 0) { ElMessage.success('位置已删除'); loadLocations() }
}

onMounted(loadCategories)
</script>

<style scoped lang="scss">
.category-page { --bg:var(--na-background); --surface:var(--na-card); --text:var(--na-foreground); --muted:var(--na-muted-foreground); --border:var(--na-border); min-height:100%; overflow-x:hidden; padding:20px; background:var(--bg); color:var(--text); }
.page-heading { display:flex; align-items:flex-end; justify-content:space-between; gap:24px; margin-bottom:20px; }
.eyebrow { margin:0 0 5px; color:var(--na-primary); font:600 12px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace; letter-spacing:.12em; }
h1 { margin:0; font-size:30px; line-height:1.2; }
.page-heading p:last-child { margin:8px 0 0; color:var(--muted); font-size:14px; }
.category-panel { overflow:hidden; border:1px solid var(--border); border-radius:var(--na-radius); background:var(--surface); box-shadow:var(--na-shadow-sm); }
.manage-tabs { padding:0 16px; border-bottom:1px solid var(--border); }
.manage-tabs :deep(.el-tabs__header) { margin:0; }
.manage-tabs :deep(.el-tabs__nav-wrap::after) { display:none; }
.manage-tabs :deep(.el-tabs__item) { height:44px; padding:0 16px; font-weight:600; }
.panel-toolbar { display:flex; align-items:center; justify-content:space-between; gap:16px; padding:12px 16px; border-bottom:1px solid var(--border); }
.panel-toolbar > .el-input { width:min(100%,420px); }
.toolbar-actions { display:flex; gap:8px; }
.category-name { display:flex; align-items:center; gap:12px; }
.category-name .color-mark { width:10px; height:36px; border-radius:5px; box-shadow:inset 0 0 0 1px rgb(255 255 255 / 24%); }
.category-name div { display:flex; flex-direction:column; gap:3px; }
.category-name small, .code-text { color:var(--muted); font:12px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace; }
.location-name { display:flex; min-width:0; align-items:center; gap:10px; }
.location-name .el-icon { flex:0 0 auto; color:var(--na-primary); font-size:18px; }
.location-name strong { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.metric { color:var(--na-primary); font-size:16px; font-variant-numeric:tabular-nums; }
.pagination-wrap { display:flex; justify-content:flex-end; padding:12px 16px; border-top:1px solid var(--border); }
.form-grid { display:grid; grid-template-columns:1fr 1fr; gap:0 16px; }
.color-field { display:flex; align-items:center; gap:10px; width:100%; }
.color-field .el-input { flex:1; }
:deep(.el-form-item__label) { color:var(--text); font-weight:600; }
@media (max-width:767px) {
  .category-page { padding:14px; }
  .page-heading { align-items:stretch; flex-direction:column; }
  .page-heading .el-button { align-self:flex-start; }
  .manage-tabs { padding:0 12px; }
  .manage-tabs :deep(.el-tabs__item) { padding:0 14px; }
  .panel-toolbar { align-items:stretch; flex-direction:column; padding:12px; }
  .panel-toolbar > .el-input { width:100%; }
  .toolbar-actions .el-button { flex:1; }
  .form-grid { grid-template-columns:1fr; }
  .pagination-wrap { overflow:hidden; justify-content:flex-start; padding:12px; }
  .pagination-wrap :deep(.el-pagination__sizes), .pagination-wrap :deep(.el-pagination__total) { display:none; }
}
</style>
