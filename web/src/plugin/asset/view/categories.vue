<template>
  <main class="category-page">
    <section class="page-heading" aria-labelledby="category-title">
      <div>
        <p class="eyebrow">CLASSIFICATION</p>
        <h1 id="category-title">资产分类</h1>
        <p>维护座椅、电脑、办公设备等分类口径，确保统计维度统一。</p>
      </div>
      <el-button type="primary" :icon="Plus" size="large" @click="openCreate">新增分类</el-button>
    </section>

    <section class="category-panel">
      <header class="panel-toolbar">
        <el-input v-model="search.keyword" :prefix-icon="Search" clearable placeholder="搜索分类名称或编码" @keyup.enter="submitSearch" />
        <div>
          <el-button type="primary" :icon="Search" @click="submitSearch">查询</el-button>
          <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
        </div>
      </header>

      <el-table v-loading="loading" :data="tableData" row-key="ID" stripe>
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

      <footer class="pagination-wrap">
        <el-pagination
          v-model:current-page="search.page"
          v-model:page-size="search.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @current-change="loadCategories"
          @size-change="sizeChanged"
        />
      </footer>
    </section>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑资产分类' : '新增资产分类'" width="min(92vw, 560px)" :close-on-click-modal="false">
      <el-form ref="formRef" :model="formData" :rules="rules" label-position="top">
        <div class="form-grid">
          <el-form-item label="分类名称" prop="name">
            <el-input v-model="formData.name" maxlength="100" placeholder="例如：电脑整机" />
          </el-form-item>
          <el-form-item label="分类编码" prop="code">
            <el-input v-model="formData.code" maxlength="50" placeholder="例如：IT-COMPUTER" />
          </el-form-item>
          <el-form-item label="展示颜色" prop="color">
            <div class="color-field">
              <el-color-picker v-model="formData.color" />
              <el-input v-model="formData.color" maxlength="20" />
            </div>
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="formData.sort" :min="0" :max="9999" controls-position="right" />
          </el-form-item>
        </div>
        <el-form-item label="分类说明">
          <el-input v-model="formData.description" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="说明该分类包含的资产范围" />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="formData.enabled" inline-prompt active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveCategory">保存分类</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Plus, Refresh, Search } from '@element-plus/icons-vue'
import { createCategory, deleteCategory, getCategoryList, updateCategory } from '@/plugin/asset/api/category'

defineOptions({ name: 'AssetCategories' })

const emptyForm = () => ({ ID: 0, name: '', code: '', description: '', color: '#334155', sort: 0, enabled: true })
const search = reactive({ page: 1, pageSize: 10, keyword: '' })
const tableData = ref([])
const total = ref(0)
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref(false)
const saving = ref(false)
const formRef = ref()
const formData = ref(emptyForm())
const rules = {
  name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入分类编码', trigger: 'blur' },
    { pattern: /^[A-Za-z0-9_-]+$/, message: '仅支持字母、数字、中划线和下划线', trigger: 'blur' }
  ],
  color: [{ required: true, message: '请选择展示颜色', trigger: 'change' }]
}

const loadCategories = async () => {
  loading.value = true
  try {
    const res = await getCategoryList(search)
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } finally { loading.value = false }
}
const submitSearch = () => { search.page = 1; loadCategories() }
const resetSearch = () => { Object.assign(search, { page: 1, pageSize: 10, keyword: '' }); loadCategories() }
const sizeChanged = () => { search.page = 1; loadCategories() }
const openCreate = () => { editing.value = false; formData.value = emptyForm(); dialogVisible.value = true }
const openEdit = (row) => { editing.value = true; formData.value = { ...emptyForm(), ...JSON.parse(JSON.stringify(row)) }; dialogVisible.value = true }

const saveCategory = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const res = editing.value ? await updateCategory(formData.value) : await createCategory(formData.value)
    if (res.code === 0) {
      ElMessage.success(editing.value ? '分类已更新' : '分类已创建')
      dialogVisible.value = false
      loadCategories()
    }
  } finally { saving.value = false }
}

const removeCategory = async (row) => {
  await ElMessageBox.confirm(`确定删除分类“${row.name}”吗？`, '删除确认', { type: 'warning' })
  const res = await deleteCategory({ id: row.ID })
  if (res.code === 0) { ElMessage.success('分类已删除'); loadCategories() }
}

onMounted(loadCategories)
</script>

<style scoped lang="scss">
.category-page { --bg:var(--na-background); --surface:var(--na-card); --text:var(--na-foreground); --muted:var(--na-muted-foreground); --border:var(--na-border); min-height:100%; padding:24px; background:var(--bg); color:var(--text); }
.page-heading { display:flex; align-items:flex-end; justify-content:space-between; gap:24px; margin-bottom:20px; }
.eyebrow { margin:0 0 5px; color:var(--na-primary); font:600 12px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace; letter-spacing:.12em; }
h1 { margin:0; font-size:clamp(26px,3vw,36px); letter-spacing:-.02em; }
.page-heading p:last-child { margin:8px 0 0; color:var(--muted); }
.category-panel { overflow:hidden; border:1px solid var(--border); border-radius:var(--na-radius); background:var(--surface); box-shadow:var(--na-shadow-sm); }
.panel-toolbar { display:flex; align-items:center; justify-content:space-between; gap:16px; padding:18px 20px; border-bottom:1px solid var(--border); }
.panel-toolbar > .el-input { width:min(100%,420px); }
.category-name { display:flex; align-items:center; gap:12px; }
.category-name .color-mark { width:10px; height:36px; border-radius:5px; box-shadow:inset 0 0 0 1px rgb(255 255 255 / 24%); }
.category-name div { display:flex; flex-direction:column; gap:3px; }
.category-name small { color:var(--muted); font:12px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace; }
.metric { color:var(--na-primary); font-size:16px; font-variant-numeric:tabular-nums; }
.pagination-wrap { display:flex; justify-content:flex-end; padding:18px 20px; border-top:1px solid var(--border); }
.form-grid { display:grid; grid-template-columns:1fr 1fr; gap:0 16px; }
.color-field { display:flex; align-items:center; gap:10px; width:100%; }
.color-field .el-input { flex:1; }
:deep(.el-form-item__label) { color:var(--text); font-weight:600; }
@media (max-width:767px) { .category-page{padding:14px}.page-heading{align-items:stretch;flex-direction:column}.panel-toolbar{align-items:stretch;flex-direction:column}.panel-toolbar>div{display:flex}.form-grid{grid-template-columns:1fr}.pagination-wrap{overflow-x:auto;justify-content:flex-start} }
</style>
