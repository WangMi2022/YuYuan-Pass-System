<template>
  <main class="asset-page">
    <section class="page-heading" aria-labelledby="asset-inventory-title">
      <div>
        <p class="eyebrow">ASSET INVENTORY</p>
        <h1 id="asset-inventory-title">资产档案</h1>
        <p class="subtitle">统一登记资产数量、采购原值、当前估值、责任人与实物照片。</p>
      </div>
      <el-button type="primary" :icon="Plus" size="large" @click="openCreate">新增资产</el-button>
    </section>

    <section class="filter-panel" aria-label="资产筛选">
      <el-form :model="search" label-position="top" @keyup.enter="loadAssets">
        <div class="filter-grid">
          <el-form-item label="关键词">
            <el-input v-model="search.keyword" clearable placeholder="编号、名称、品牌、型号、责任人" :prefix-icon="Search" />
          </el-form-item>
          <el-form-item label="资产分类">
            <el-select v-model="search.categoryId" clearable placeholder="全部分类">
              <el-option v-for="item in categories" :key="item.ID" :label="item.name" :value="item.ID" />
            </el-select>
          </el-form-item>
          <el-form-item label="资产状态">
            <el-select v-model="search.status" clearable placeholder="全部状态">
              <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="存放位置">
            <el-input v-model="search.location" clearable placeholder="例如：A 区三楼" />
          </el-form-item>
          <div class="filter-actions">
            <el-button type="primary" :icon="Search" @click="submitSearch">查询</el-button>
            <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
          </div>
        </div>
      </el-form>
    </section>

    <section class="table-panel">
      <header class="panel-header">
        <div>
          <h2>资产清单</h2>
          <span>共 {{ total }} 条档案</span>
        </div>
        <el-button :icon="Refresh" text aria-label="刷新资产列表" @click="loadAssets">刷新</el-button>
      </header>

      <el-table v-loading="loading" :data="tableData" row-key="ID" stripe class="asset-table">
        <el-table-column label="资产" min-width="250" fixed="left">
          <template #default="{ row }">
            <div class="asset-identity">
              <el-image
                v-if="row.photos?.length"
                class="asset-thumb"
                :src="row.photos[0].url"
                :preview-src-list="row.photos.map((item) => item.url)"
                preview-teleported
                fit="cover"
                lazy
                :alt="`${row.name}实物照片`"
              />
              <div v-else class="asset-thumb asset-thumb--empty" aria-hidden="true">
                <el-icon><Picture /></el-icon>
              </div>
              <div class="identity-copy">
                <strong>{{ row.name }}</strong>
                <span>{{ row.assetCode }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="分类" min-width="130">
          <template #default="{ row }">
            <span class="category-pill">
              <i :style="{ background: row.category?.color || '#64748b' }" />
              {{ row.category?.name || '未分类' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="品牌 / 型号" min-width="160">
          <template #default="{ row }">
            <div class="two-line"><span>{{ row.brand || '—' }}</span><small>{{ row.model || '—' }}</small></div>
          </template>
        </el-table-column>
        <el-table-column label="数量" min-width="100" align="right">
          <template #default="{ row }"><strong class="number">{{ row.quantity }}</strong> {{ row.unit }}</template>
        </el-table-column>
        <el-table-column label="资产原值" min-width="140" align="right">
          <template #default="{ row }"><span class="money">{{ currency(row.originalValue) }}</span></template>
        </el-table-column>
        <el-table-column label="当前估值" min-width="140" align="right">
          <template #default="{ row }"><span class="money money--current">{{ currency(row.currentValue) }}</span></template>
        </el-table-column>
        <el-table-column label="状态" min-width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="statusMeta(row.status).type" effect="light">{{ statusMeta(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="位置 / 保管人" min-width="170">
          <template #default="{ row }">
            <div class="two-line"><span>{{ row.location || '位置待补充' }}</span><small>{{ row.custodian || '未指定保管人' }}</small></div>
          </template>
        </el-table-column>
        <el-table-column label="购置日期" min-width="120">
          <template #default="{ row }">{{ dateText(row.purchaseDate) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link :icon="Delete" @click="removeAsset(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂时没有资产档案">
            <el-button type="primary" @click="openCreate">登记第一项资产</el-button>
          </el-empty>
        </template>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="search.page"
          v-model:page-size="search.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadAssets"
          @size-change="sizeChanged"
        />
      </div>
    </section>

    <el-drawer v-model="drawerVisible" :size="drawerSize" destroy-on-close :close-on-click-modal="false">
      <template #header>
        <div class="drawer-title">
          <span>{{ editing ? '编辑资产' : '新增资产' }}</span>
          <small>{{ editing ? formData.assetCode : '建立一份可统计、可追溯的资产档案' }}</small>
        </div>
      </template>
      <el-form ref="formRef" :model="formData" :rules="rules" label-position="top" class="asset-form">
        <h3>基础信息</h3>
        <div class="form-grid">
          <el-form-item label="资产编号" prop="assetCode">
            <el-input v-model="formData.assetCode" maxlength="80" placeholder="例如：IT-2026-0001" />
          </el-form-item>
          <el-form-item label="资产名称" prop="name">
            <el-input v-model="formData.name" maxlength="150" placeholder="例如：研发笔记本电脑" />
          </el-form-item>
          <el-form-item label="资产分类" prop="categoryId">
            <el-select v-model="formData.categoryId" placeholder="选择分类" filterable>
              <el-option v-for="item in categories" :key="item.ID" :label="item.name" :value="item.ID" />
            </el-select>
          </el-form-item>
          <el-form-item label="资产状态" prop="status">
            <el-select v-model="formData.status">
              <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="品牌">
            <el-input v-model="formData.brand" maxlength="100" placeholder="品牌名称" />
          </el-form-item>
          <el-form-item label="规格型号">
            <el-input v-model="formData.model" maxlength="120" placeholder="产品型号或规格" />
          </el-form-item>
          <el-form-item label="序列号">
            <el-input v-model="formData.serialNumber" maxlength="120" placeholder="设备 SN，可选" />
          </el-form-item>
          <el-form-item label="计量单位">
            <el-input v-model="formData.unit" maxlength="30" placeholder="件 / 台 / 套" />
          </el-form-item>
        </div>

        <h3>数量与计价</h3>
        <div class="valuation-box">
          <el-form-item label="数量" prop="quantity">
            <el-input-number v-model="formData.quantity" :min="1" :max="999999" controls-position="right" @change="syncCurrentValue" />
          </el-form-item>
          <el-form-item label="采购单价（元）" prop="unitPrice">
            <el-input-number v-model="formData.unitPrice" :min="0" :precision="2" :step="100" controls-position="right" @change="syncCurrentValue" />
          </el-form-item>
          <div class="computed-value">
            <span>自动计算原值</span>
            <strong>{{ currency(originalPreview) }}</strong>
          </div>
          <el-form-item label="当前估值（元）" prop="currentValue">
            <el-input-number v-model="formData.currentValue" :min="0" :max="originalPreview || 999999999" :precision="2" :step="100" controls-position="right" />
          </el-form-item>
        </div>

        <h3>使用与责任信息</h3>
        <div class="form-grid">
          <el-form-item label="存放位置">
            <el-input v-model="formData.location" maxlength="150" placeholder="园区 / 楼层 / 房间" />
          </el-form-item>
          <el-form-item label="保管人 / 使用人">
            <el-input v-model="formData.custodian" maxlength="100" placeholder="姓名或部门" />
          </el-form-item>
          <el-form-item label="供应商">
            <el-input v-model="formData.supplier" maxlength="150" placeholder="供应商名称" />
          </el-form-item>
          <el-form-item label="购置日期">
            <el-date-picker v-model="formData.purchaseDate" type="date" placeholder="选择购置日期" />
          </el-form-item>
          <el-form-item label="质保到期日">
            <el-date-picker v-model="formData.warrantyEndDate" type="date" placeholder="选择质保到期日" />
          </el-form-item>
        </div>

        <h3>实物照片</h3>
        <el-form-item>
          <el-upload
            class="photo-uploader"
            list-type="picture-card"
            :file-list="formData.photos"
            :http-request="uploadPhoto"
            :before-upload="beforePhotoUpload"
            :on-preview="previewPhoto"
            :on-remove="removePhoto"
            :on-exceed="photoExceed"
            :limit="6"
            accept="image/jpeg,image/png,image/webp,image/gif"
            multiple
          >
            <el-icon><Plus /></el-icon>
            <template #tip><div class="el-upload__tip">最多 6 张，单张不超过 10MB，文件保存到 RustFS。</div></template>
          </el-upload>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="formData.remarks" type="textarea" :rows="3" maxlength="1000" show-word-limit placeholder="补充使用、盘点或维修说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="drawer-actions">
          <el-button size="large" @click="drawerVisible = false">取消</el-button>
          <el-button type="primary" size="large" :loading="saving" @click="saveAsset">保存资产</el-button>
        </div>
      </template>
    </el-drawer>

    <el-dialog v-model="previewVisible" title="资产照片" width="min(92vw, 900px)" append-to-body>
      <img class="preview-image" :src="previewUrl" alt="资产实物大图" />
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Picture, Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  createAsset,
  deleteAsset,
  getAssetList,
  getCategoryOptions,
  updateAsset,
  uploadAssetPhoto
} from '@/plugin/asset/api/asset'

defineOptions({ name: 'AssetInventory' })

const statusOptions = [
  { value: 'in_use', label: '使用中', type: 'success' },
  { value: 'idle', label: '闲置', type: 'info' },
  { value: 'maintenance', label: '维修中', type: 'warning' },
  { value: 'retired', label: '已处置', type: 'danger' }
]

const emptyForm = () => ({
  ID: 0,
  assetCode: '',
  name: '',
  categoryId: undefined,
  brand: '',
  model: '',
  serialNumber: '',
  quantity: 1,
  unit: '件',
  unitPrice: 0,
  currentValue: 0,
  status: 'in_use',
  location: '',
  custodian: '',
  supplier: '',
  purchaseDate: null,
  warrantyEndDate: null,
  photos: [],
  remarks: ''
})

const search = reactive({ page: 1, pageSize: 10, keyword: '', categoryId: undefined, status: '', location: '' })
const tableData = ref([])
const categories = ref([])
const total = ref(0)
const loading = ref(false)
const drawerVisible = ref(false)
const editing = ref(false)
const saving = ref(false)
const formRef = ref()
const formData = ref(emptyForm())
const previewVisible = ref(false)
const previewUrl = ref('')

const drawerSize = computed(() => (window.innerWidth < 768 ? '96%' : '780px'))
const originalPreview = computed(() => Number(formData.value.quantity || 0) * Number(formData.value.unitPrice || 0))
const rules = {
  assetCode: [{ required: true, message: '请输入资产编号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入资产名称', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请选择资产分类', trigger: 'change' }],
  quantity: [{ required: true, message: '数量必须大于 0', trigger: 'change' }],
  unitPrice: [{ required: true, message: '请输入采购单价', trigger: 'change' }],
  currentValue: [{ required: true, message: '请输入当前估值', trigger: 'change' }]
}

const currency = (value) => new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY', minimumFractionDigits: 2 }).format(Number(value || 0))
const dateText = (value) => (value ? new Date(value).toLocaleDateString('zh-CN') : '—')
const statusMeta = (value) => statusOptions.find((item) => item.value === value) || { label: value || '未知', type: 'info' }

const loadCategories = async () => {
  const res = await getCategoryOptions()
  if (res.code === 0) categories.value = res.data || []
}

const loadAssets = async () => {
  loading.value = true
  try {
    const res = await getAssetList(search)
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

const submitSearch = () => { search.page = 1; loadAssets() }
const resetSearch = () => {
  Object.assign(search, { page: 1, pageSize: 10, keyword: '', categoryId: undefined, status: '', location: '' })
  loadAssets()
}
const sizeChanged = () => { search.page = 1; loadAssets() }

const openCreate = () => {
  editing.value = false
  formData.value = emptyForm()
  drawerVisible.value = true
}

const openEdit = (row) => {
  editing.value = true
  formData.value = {
    ...emptyForm(),
    ...JSON.parse(JSON.stringify(row)),
    purchaseDate: row.purchaseDate ? new Date(row.purchaseDate) : null,
    warrantyEndDate: row.warrantyEndDate ? new Date(row.warrantyEndDate) : null,
    photos: (row.photos || []).map((item, index) => ({ ...item, uid: item.key || index, status: 'success' }))
  }
  drawerVisible.value = true
}

const syncCurrentValue = () => {
  if (!editing.value || !formData.value.currentValue) formData.value.currentValue = originalPreview.value
}

const saveAsset = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const payload = {
      ...formData.value,
      photos: (formData.value.photos || []).map(({ name, key, url }) => ({ name, key, url }))
    }
    const res = editing.value ? await updateAsset(payload) : await createAsset(payload)
    if (res.code === 0) {
      ElMessage.success(editing.value ? '资产已更新' : '资产已登记')
      drawerVisible.value = false
      await loadAssets()
    }
  } finally {
    saving.value = false
  }
}

const removeAsset = async (row) => {
  await ElMessageBox.confirm(`确定删除资产“${row.name}”吗？删除后统计数据会同步更新。`, '删除确认', {
    type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消'
  })
  const res = await deleteAsset({ id: row.ID })
  if (res.code === 0) {
    ElMessage.success('资产已删除')
    loadAssets()
  }
}

const beforePhotoUpload = (file) => {
  if (!['image/jpeg', 'image/png', 'image/webp', 'image/gif'].includes(file.type)) {
    ElMessage.error('仅支持 JPG、PNG、WebP、GIF 图片')
    return false
  }
  if (file.size > 10 * 1024 * 1024) {
    ElMessage.error('单张图片不能超过 10MB')
    return false
  }
  return true
}

const uploadPhoto = async (options) => {
  try {
    const res = await uploadAssetPhoto(options.file)
    if (res.code === 0) {
      const photo = { ...res.data, uid: res.data.key, status: 'success' }
      formData.value.photos.push(photo)
      options.onSuccess(photo)
      ElMessage.success('照片已保存到 RustFS')
    } else {
      options.onError(new Error(res.msg || '上传失败'))
    }
  } catch (error) {
    options.onError(error)
  }
}

const removePhoto = (file) => {
  const key = file.key || file.response?.key
  formData.value.photos = formData.value.photos.filter((item) => item.key !== key)
}
const previewPhoto = (file) => { previewUrl.value = file.url; previewVisible.value = true }
const photoExceed = () => ElMessage.warning('每项资产最多上传 6 张照片')

onMounted(async () => {
  await loadCategories()
  await loadAssets()
})
</script>

<style scoped lang="scss">
.asset-page {
  --asset-bg: var(--na-background);
  --asset-surface: var(--na-card);
  --asset-text: var(--na-foreground);
  --asset-muted: var(--na-muted-foreground);
  --asset-border: var(--na-border);
  min-height: 100%; padding: 24px; background: var(--asset-bg); color: var(--asset-text);
}
.page-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; margin-bottom: 20px; }
.eyebrow { margin: 0 0 5px; color: var(--na-primary); font: 600 12px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; }
h1 { margin: 0; font-size: clamp(26px, 3vw, 36px); line-height: 1.2; letter-spacing: -.02em; }
.subtitle { margin: 8px 0 0; color: var(--asset-muted); }
.filter-panel, .table-panel { border: 1px solid var(--asset-border); border-radius: var(--na-radius); background: var(--asset-surface); box-shadow: var(--na-shadow-sm); }
.filter-panel { padding: 18px 20px 2px; margin-bottom: 16px; }
.filter-grid { display: grid; grid-template-columns: minmax(220px, 1.5fr) repeat(3, minmax(160px, 1fr)) auto; gap: 14px; align-items: end; }
.filter-actions { display: flex; gap: 8px; padding-bottom: 18px; }
.table-panel { overflow: hidden; }
.panel-header { display: flex; align-items: center; justify-content: space-between; padding: 18px 20px; border-bottom: 1px solid var(--asset-border); }
.panel-header h2 { margin: 0 0 3px; font-size: 17px; }
.panel-header span { color: var(--asset-muted); font-size: 13px; }
.asset-table { --el-table-header-bg-color: #f8fafc; --el-table-row-hover-bg-color: #f1f5f9; }
.asset-identity { display: flex; align-items: center; gap: 12px; min-width: 0; }
.asset-thumb { width: 48px; height: 48px; flex: 0 0 48px; border-radius: 10px; border: 1px solid var(--asset-border); background: #f1f5f9; }
.asset-thumb--empty { display: grid; place-items: center; color: #94a3b8; font-size: 20px; }
.identity-copy, .two-line { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.identity-copy strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.identity-copy span, .two-line small { color: var(--asset-muted); font: 12px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; }
.category-pill { display: inline-flex; align-items: center; gap: 7px; }
.category-pill i { width: 8px; height: 8px; border-radius: 50%; }
.number, .money { font-variant-numeric: tabular-nums; }
.money--current { color: var(--na-primary); font-weight: 600; }
.pagination-wrap { display: flex; justify-content: flex-end; padding: 18px 20px; border-top: 1px solid var(--asset-border); }
.drawer-title { display: flex; flex-direction: column; gap: 4px; }
.drawer-title span { color: var(--asset-text); font-size: 20px; font-weight: 700; }
.drawer-title small { color: var(--asset-muted); font-weight: 400; }
.asset-form h3 { margin: 8px 0 18px; padding-bottom: 9px; border-bottom: 1px solid var(--asset-border); font-size: 15px; }
.asset-form h3:not(:first-child) { margin-top: 26px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; column-gap: 18px; }
.valuation-box { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; padding: 18px; border: 1px solid var(--asset-border); border-radius: var(--na-radius); background: var(--na-muted); }
.valuation-box :deep(.el-input-number), .form-grid :deep(.el-select), .form-grid :deep(.el-date-editor) { width: 100%; }
.computed-value { display: flex; flex-direction: column; justify-content: center; gap: 4px; min-height: 64px; }
.computed-value span { color: var(--asset-muted); font-size: 13px; }
.computed-value strong { color: var(--na-primary); font-size: 22px; font-variant-numeric: tabular-nums; }
.drawer-actions { display: flex; justify-content: flex-end; gap: 10px; }
.preview-image { display: block; width: 100%; max-height: 70vh; object-fit: contain; border-radius: 10px; background: #0f172a; }
:deep(.el-form-item__label) { color: var(--asset-text); font-weight: 600; }
:deep(.el-upload-list--picture-card .el-upload-list__item), :deep(.el-upload--picture-card) { width: 112px; height: 112px; }

:global(html.dark) .asset-table { --el-table-header-bg-color: var(--na-table-header); --el-table-row-hover-bg-color: var(--na-table-hover); }

@media (max-width: 1200px) { .filter-grid { grid-template-columns: repeat(2, minmax(180px, 1fr)); } .filter-actions { align-self: end; } }
@media (max-width: 767px) {
  .asset-page { padding: 14px; }
  .page-heading { align-items: stretch; flex-direction: column; }
  .filter-grid, .form-grid, .valuation-box { grid-template-columns: 1fr; }
  .filter-actions { padding-bottom: 16px; }
  .pagination-wrap { overflow-x: auto; justify-content: flex-start; }
}
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { scroll-behavior: auto !important; transition-duration: .01ms !important; } }
</style>
