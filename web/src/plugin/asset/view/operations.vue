<template>
  <main class="operation-page">
    <section class="page-heading" :aria-labelledby="`${operationType}-title`">
      <div>
        <p class="eyebrow">{{ currentMeta.eyebrow }}</p>
        <h1 :id="`${operationType}-title`">{{ currentMeta.title }}</h1>
        <p class="subtitle">{{ currentMeta.description }}</p>
      </div>
      <el-button type="primary" :icon="Plus" size="large" @click="openCreate">
        新增{{ currentMeta.shortLabel }}单
      </el-button>
    </section>

    <section class="filter-panel" aria-label="业务单筛选">
      <el-form :model="search" label-position="top" @keyup.enter="submitSearch">
        <div class="filter-grid">
          <el-form-item label="关键词">
            <el-input
              v-model="search.keyword"
              clearable
              placeholder="业务单号、资产编号或名称"
              :prefix-icon="Search"
            />
          </el-form-item>
          <el-form-item label="单据状态">
            <el-select v-model="search.status" clearable placeholder="全部状态">
              <el-option label="草稿" value="draft" />
              <el-option label="已完成" value="completed" />
            </el-select>
          </el-form-item>
          <el-form-item label="业务日期">
            <el-date-picker
              v-model="search.dateRange"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              unlink-panels
            />
          </el-form-item>
          <div class="filter-actions">
            <el-button type="primary" :icon="Search" @click="submitSearch">查询</el-button>
            <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
          </div>
        </div>
      </el-form>
    </section>

    <section class="table-panel" aria-label="资产业务单列表">
      <header class="panel-header">
        <div>
          <h2>{{ currentMeta.shortLabel }}单据</h2>
          <span>共 {{ total }} 张单据，提交后将同步更新资产档案</span>
        </div>
        <el-button :icon="Refresh" text aria-label="刷新业务单" @click="loadOrders">刷新</el-button>
      </header>

      <el-table v-loading="loading" :data="tableData" row-key="ID" class="operation-table">
        <el-table-column label="业务单号" min-width="190" fixed="left">
          <template #default="{ row }">
            <button type="button" class="order-link" @click="openDetail(row)">{{ row.orderNo }}</button>
          </template>
        </el-table-column>
        <el-table-column label="业务日期" width="120">
          <template #default="{ row }">{{ dateText(row.businessDate) }}</template>
        </el-table-column>
        <el-table-column label="资产明细" min-width="230">
          <template #default="{ row }">
            <div class="asset-summary">
              <strong>{{ firstAssetName(row) }}</strong>
              <span>{{ itemSummary(row) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="目标信息" min-width="190">
          <template #default="{ row }">
            <div class="two-line">
              <span>{{ row.targetLocation || '位置不变' }}</span>
              <small>{{ row.targetCustodian || '责任人不变或清空' }}</small>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="原因" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.reason || row.remarks || '—' }}</template>
        </el-table-column>
        <el-table-column label="创建人" min-width="120">
          <template #default="{ row }">{{ row.createdByName || '系统用户' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="orderStatus(row.status).type" effect="light">{{ orderStatus(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="160">
          <template #default="{ row }">{{ dateTimeText(row.UpdatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right" align="center">
          <template #default="{ row }">
            <template v-if="row.status === 'draft'">
              <el-button type="primary" link :icon="Edit" @click="openEdit(row)">编辑</el-button>
              <el-button type="success" link :loading="processingId === row.ID" @click="submitOrder(row)">提交</el-button>
              <el-button type="danger" link :icon="Delete" @click="removeOrder(row)">删除</el-button>
            </template>
            <el-button v-else type="primary" link :icon="View" @click="openDetail(row)">查看</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty :description="`暂无${currentMeta.shortLabel}单据`">
            <el-button type="primary" @click="openCreate">新建第一张单据</el-button>
          </el-empty>
        </template>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="search.page"
          v-model:page-size="search.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          size="small"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadOrders"
          @size-change="sizeChanged"
        />
      </div>
    </section>

    <el-drawer
      v-model="drawerVisible"
      :size="drawerSize"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <template #header>
        <div class="drawer-title">
          <span>{{ editing ? `编辑${currentMeta.shortLabel}单` : `新增${currentMeta.shortLabel}单` }}</span>
          <small>{{ editing ? formData.orderNo : currentMeta.description }}</small>
        </div>
      </template>

      <el-form ref="formRef" :model="formData" :rules="formRules" label-position="top" class="operation-form">
        <h3>业务信息</h3>
        <div class="form-grid">
          <el-form-item label="业务日期" prop="businessDate">
            <el-date-picker v-model="formData.businessDate" type="date" placeholder="选择业务日期" />
          </el-form-item>
          <el-form-item v-if="currentMeta.showLocation" :label="currentMeta.locationLabel" prop="targetLocation">
            <el-input v-model="formData.targetLocation" maxlength="150" :placeholder="currentMeta.locationPlaceholder" />
          </el-form-item>
          <el-form-item v-if="currentMeta.showCustodian" :label="currentMeta.custodianLabel" prop="targetCustodian">
            <el-input v-model="formData.targetCustodian" maxlength="100" :placeholder="currentMeta.custodianPlaceholder" />
          </el-form-item>
          <el-form-item label="业务原因" prop="reason">
            <el-input v-model="formData.reason" maxlength="500" :placeholder="currentMeta.reasonPlaceholder" />
          </el-form-item>
        </div>
        <el-form-item label="备注">
          <el-input v-model="formData.remarks" type="textarea" :rows="3" maxlength="1000" show-word-limit placeholder="补充审批、交接或处理说明" />
        </el-form-item>

        <h3>资产明细</h3>
        <el-form-item label="选择资产" prop="assetIds">
          <el-select
            v-model="formData.assetIds"
            multiple
            filterable
            collapse-tags
            :max-collapse-tags="3"
            :loading="optionsLoading"
            :placeholder="currentMeta.assetPlaceholder"
          >
            <el-option v-for="asset in assetOptions" :key="asset.ID" :value="asset.ID" :label="`${asset.assetCode} · ${asset.name}`">
              <div class="asset-option">
                <span><strong>{{ asset.assetCode }}</strong>{{ asset.name }}</span>
                <small>{{ assetStatus(asset.status).label }} · {{ asset.quantity }} {{ asset.unit }}</small>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <div class="selected-assets">
          <el-table :data="selectedAssets" row-key="ID" size="small" max-height="300">
            <el-table-column label="资产" min-width="210">
              <template #default="{ row }">
                <div class="asset-summary"><strong>{{ row.name }}</strong><span>{{ row.assetCode }}</span></div>
              </template>
            </el-table-column>
            <el-table-column label="当前状态" width="100" align="center">
              <template #default="{ row }"><el-tag :type="assetStatus(row.status).type" effect="plain">{{ assetStatus(row.status).label }}</el-tag></template>
            </el-table-column>
            <el-table-column label="位置 / 保管人" min-width="170">
              <template #default="{ row }"><div class="two-line"><span>{{ row.location || '未设置位置' }}</span><small>{{ row.custodian || '无保管人' }}</small></div></template>
            </el-table-column>
            <el-table-column label="流转数量" width="110" align="right">
              <template #default="{ row }"><strong class="number">{{ row.quantity }}</strong> {{ row.unit }}</template>
            </el-table-column>
            <el-table-column width="54" align="center">
              <template #default="{ row }">
                <el-tooltip content="移除资产" placement="top">
                  <el-button :icon="Close" text circle aria-label="移除资产" @click="removeSelectedAsset(row.ID)" />
                </el-tooltip>
              </template>
            </el-table-column>
            <template #empty><el-empty :image-size="44" description="尚未选择资产" /></template>
          </el-table>
        </div>
      </el-form>

      <template #footer>
        <div class="drawer-actions">
          <el-button size="large" @click="drawerVisible = false">取消</el-button>
          <el-button size="large" :loading="saving" @click="saveOrder(false)">保存草稿</el-button>
          <el-button type="primary" size="large" :loading="submitting" @click="saveOrder(true)">保存并提交</el-button>
        </div>
      </template>
    </el-drawer>

    <el-dialog v-model="detailVisible" :title="detailOrder?.orderNo || `${currentMeta.shortLabel}单详情`" width="min(94vw, 900px)" append-to-body>
      <div v-if="detailOrder" class="order-detail">
        <el-descriptions :column="detailColumns" border>
          <el-descriptions-item label="业务类型">{{ currentMeta.shortLabel }}</el-descriptions-item>
          <el-descriptions-item label="单据状态"><el-tag :type="orderStatus(detailOrder.status).type">{{ orderStatus(detailOrder.status).label }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="业务日期">{{ dateText(detailOrder.businessDate) }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ detailOrder.createdByName || '系统用户' }}</el-descriptions-item>
          <el-descriptions-item label="目标位置">{{ detailOrder.targetLocation || '—' }}</el-descriptions-item>
          <el-descriptions-item label="目标责任人">{{ detailOrder.targetCustodian || '—' }}</el-descriptions-item>
          <el-descriptions-item label="业务原因" :span="detailColumns">{{ detailOrder.reason || '—' }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="detailColumns">{{ detailOrder.remarks || '—' }}</el-descriptions-item>
        </el-descriptions>

        <h3>资产流转明细</h3>
        <el-table :data="detailOrder.items || []" row-key="ID" border>
          <el-table-column label="资产" min-width="200">
            <template #default="{ row }"><div class="asset-summary"><strong>{{ row.assetName }}</strong><span>{{ row.assetCode }}</span></div></template>
          </el-table-column>
          <el-table-column label="数量" width="90" align="right" prop="quantity" />
          <el-table-column label="状态变化" min-width="160">
            <template #default="{ row }">{{ assetStatus(row.fromStatus).label }} → {{ assetStatus(row.toStatus).label }}</template>
          </el-table-column>
          <el-table-column label="位置变化" min-width="190">
            <template #default="{ row }">{{ row.fromLocation || '未设置' }} → {{ row.toLocation || '未设置' }}</template>
          </el-table-column>
          <el-table-column label="责任人变化" min-width="180">
            <template #default="{ row }">{{ row.fromCustodian || '无' }} → {{ row.toCustodian || '无' }}</template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer><el-button type="primary" @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Close, Delete, Edit, Plus, Refresh, Search, View } from '@element-plus/icons-vue'
import {
  createAssetOperation,
  deleteAssetOperation,
  getAssetOperationDetail,
  getAssetOperationList,
  getOperationAssetOptions,
  submitAssetOperation,
  updateAssetOperation
} from '@/plugin/asset/api/operation'

defineOptions({ name: 'AssetOperations' })

const route = useRoute()
const operationMeta = {
  assetInbound: {
    type: 'inbound', title: '入库管理', shortLabel: '入库', eyebrow: 'ASSET INBOUND',
    description: '登记资产入库位置，提交后资产进入闲置待领用状态。',
    showLocation: true, locationRequired: true, locationLabel: '入库位置', locationPlaceholder: '仓库 / 库区 / 货架',
    showCustodian: false, reasonRequired: false, reasonPlaceholder: '采购入库、退库重入等', assetPlaceholder: '选择待入库的闲置资产'
  },
  assetIssue: {
    type: 'issue', title: '领用管理', shortLabel: '领用', eyebrow: 'ASSET ISSUE',
    description: '记录资产领用人与使用位置，提交后资产进入使用中状态。',
    showLocation: true, locationRequired: false, locationLabel: '使用位置', locationPlaceholder: '部门 / 楼层 / 工位，可选',
    showCustodian: true, custodianRequired: true, custodianLabel: '领用人 / 责任人', custodianPlaceholder: '姓名或部门',
    reasonRequired: false, reasonPlaceholder: '办公领用、项目领用等', assetPlaceholder: '选择可领用的闲置资产'
  },
  assetTransfer: {
    type: 'transfer', title: '调拨管理', shortLabel: '调拨', eyebrow: 'ASSET TRANSFER',
    description: '调整资产所在位置或责任人，保留调拨前后的完整快照。',
    showLocation: true, locationRequired: true, locationLabel: '调入位置', locationPlaceholder: '目标园区 / 楼层 / 房间',
    showCustodian: true, custodianRequired: false, custodianLabel: '新责任人', custodianPlaceholder: '不填写则保持原责任人',
    reasonRequired: false, reasonPlaceholder: '部门调整、位置变更等', assetPlaceholder: '选择闲置或使用中的资产'
  },
  assetReturn: {
    type: 'return', title: '归还管理', shortLabel: '归还', eyebrow: 'ASSET RETURN',
    description: '登记使用或维修资产归还，提交后清空责任人并转为闲置。',
    showLocation: true, locationRequired: true, locationLabel: '归还位置', locationPlaceholder: '仓库 / 库区 / 货架',
    showCustodian: false, reasonRequired: false, reasonPlaceholder: '离职归还、项目结束、维修完成等', assetPlaceholder: '选择使用中或维修中的资产'
  },
  assetMaintenance: {
    type: 'maintenance', title: '维修管理', shortLabel: '维修', eyebrow: 'ASSET MAINTENANCE',
    description: '记录故障原因和送修信息，提交后资产进入维修中状态。',
    showLocation: true, locationRequired: false, locationLabel: '维修位置', locationPlaceholder: '维修点或服务商地址，可选',
    showCustodian: true, custodianRequired: false, custodianLabel: '维修责任人', custodianPlaceholder: '内部负责人或服务商，可选',
    reasonRequired: true, reasonPlaceholder: '请填写故障现象或送修原因', assetPlaceholder: '选择闲置或使用中的资产'
  },
  assetScrap: {
    type: 'scrap', title: '报废管理', shortLabel: '报废', eyebrow: 'ASSET SCRAP',
    description: '记录资产处置原因，提交后资产转为已处置且当前估值归零。',
    showLocation: true, locationRequired: false, locationLabel: '处置位置', locationPlaceholder: '报废库或处置地点，可选',
    showCustodian: false, reasonRequired: true, reasonPlaceholder: '请填写报废、损毁或处置原因', assetPlaceholder: '选择待处置资产'
  }
}

const currentMeta = computed(() => operationMeta[route.name] || operationMeta.assetInbound)
const operationType = computed(() => currentMeta.value.type)
const search = reactive({ page: 1, pageSize: 10, keyword: '', status: '', dateRange: [] })
const tableData = ref([])
const total = ref(0)
const loading = ref(false)
const processingId = ref(0)
const drawerVisible = ref(false)
const editing = ref(false)
const saving = ref(false)
const submitting = ref(false)
const formRef = ref()
const assetOptions = ref([])
const optionsLoading = ref(false)
const detailVisible = ref(false)
const detailOrder = ref(null)

const emptyForm = () => ({
  ID: 0, orderNo: '', type: operationType.value, businessDate: new Date(),
  targetLocation: '', targetCustodian: '', reason: '', remarks: '', assetIds: []
})
const formData = ref(emptyForm())
const drawerSize = computed(() => (window.innerWidth < 768 ? '96%' : '860px'))
const detailColumns = computed(() => (window.innerWidth < 640 ? 1 : 2))
const formRules = computed(() => ({
  businessDate: [{ required: true, message: '请选择业务日期', trigger: 'change' }],
  targetLocation: currentMeta.value.locationRequired ? [{ required: true, message: `请填写${currentMeta.value.locationLabel}`, trigger: 'blur' }] : [],
  targetCustodian: currentMeta.value.custodianRequired ? [{ required: true, message: `请填写${currentMeta.value.custodianLabel}`, trigger: 'blur' }] : [],
  reason: currentMeta.value.reasonRequired ? [{ required: true, message: '请填写业务原因', trigger: 'blur' }] : [],
  assetIds: [{ type: 'array', required: true, min: 1, message: '请至少选择一项资产', trigger: 'change' }]
}))

const assetStatusMap = {
  in_use: { label: '使用中', type: 'success' },
  idle: { label: '闲置', type: 'info' },
  maintenance: { label: '维修中', type: 'warning' },
  retired: { label: '已处置', type: 'danger' }
}
const assetStatus = (value) => assetStatusMap[value] || { label: value || '未知', type: 'info' }
const orderStatus = (value) => value === 'completed' ? { label: '已完成', type: 'success' } : { label: '草稿', type: 'info' }
const dateText = (value) => value ? new Date(value).toLocaleDateString('zh-CN') : '—'
const dateTimeText = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'
const firstAssetName = (row) => row.items?.[0]?.assetName || '暂无资产'
const itemSummary = (row) => {
  const items = row.items || []
  const quantity = items.reduce((sum, item) => sum + Number(item.quantity || 0), 0)
  return `${items.length} 项档案，共 ${quantity} 件（套）`
}

const optionMap = computed(() => new Map(assetOptions.value.map((asset) => [asset.ID, asset])))
const selectedAssets = computed(() => formData.value.assetIds.map((id) => optionMap.value.get(id)).filter(Boolean))

const loadOrders = async () => {
  loading.value = true
  try {
    const [startDate, endDate] = search.dateRange || []
    const res = await getAssetOperationList({
      page: search.page, pageSize: search.pageSize, keyword: search.keyword,
      status: search.status, type: operationType.value, startDate, endDate
    })
    if (res.code === 0) {
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
    }
  } finally {
    loading.value = false
  }
}

const loadAssetOptions = async (extraAssets = []) => {
  optionsLoading.value = true
  try {
    const res = await getOperationAssetOptions({ type: operationType.value })
    const normal = res.code === 0 ? (res.data || []) : []
    const merged = [...normal]
    const seen = new Set(normal.map((asset) => asset.ID))
    extraAssets.forEach((asset) => {
      if (asset?.ID && !seen.has(asset.ID)) {
        merged.push(asset)
        seen.add(asset.ID)
      }
    })
    assetOptions.value = merged
  } finally {
    optionsLoading.value = false
  }
}

const submitSearch = () => { search.page = 1; loadOrders() }
const resetSearch = () => {
  Object.assign(search, { page: 1, pageSize: 10, keyword: '', status: '', dateRange: [] })
  loadOrders()
}
const sizeChanged = () => { search.page = 1; loadOrders() }

const openCreate = async () => {
  editing.value = false
  formData.value = emptyForm()
  await loadAssetOptions()
  drawerVisible.value = true
}

const openEdit = async (row) => {
  const res = await getAssetOperationDetail({ id: row.ID })
  if (res.code !== 0) return
  const order = res.data
  editing.value = true
  const extraAssets = (order.items || []).map((item) => item.asset || {
    ID: item.assetId, assetCode: item.assetCode, name: item.assetName, quantity: item.quantity,
    unit: '件', status: item.fromStatus, location: item.fromLocation, custodian: item.fromCustodian
  })
  await loadAssetOptions(extraAssets)
  formData.value = {
    ID: order.ID, orderNo: order.orderNo, type: order.type,
    businessDate: order.businessDate ? new Date(order.businessDate) : new Date(),
    targetLocation: order.targetLocation || '', targetCustodian: order.targetCustodian || '',
    reason: order.reason || '', remarks: order.remarks || '',
    assetIds: (order.items || []).map((item) => item.assetId)
  }
  drawerVisible.value = true
}

const payload = (submit) => ({ ...formData.value, type: operationType.value, submit })
const saveOrder = async (submit) => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (submit) {
    try {
      await ElMessageBox.confirm(`提交后将立即更新 ${selectedAssets.value.length} 项资产，且不能直接撤销。`, `提交${currentMeta.value.shortLabel}单`, {
        type: currentMeta.value.type === 'scrap' ? 'error' : 'warning',
        confirmButtonText: '确认提交', cancelButtonText: '取消'
      })
    } catch { return }
  }
  if (submit) submitting.value = true
  else saving.value = true
  try {
    const res = editing.value ? await updateAssetOperation(payload(submit)) : await createAssetOperation(payload(submit))
    if (res.code === 0) {
      ElMessage.success(submit ? `${currentMeta.value.shortLabel}单已提交` : '草稿已保存')
      drawerVisible.value = false
      await Promise.all([loadOrders(), loadAssetOptions()])
    }
  } finally {
    saving.value = false
    submitting.value = false
  }
}

const submitOrder = async (row) => {
  try {
    await ElMessageBox.confirm(`确定提交业务单 ${row.orderNo} 吗？提交后将立即更新资产档案。`, '提交业务单', {
      type: operationType.value === 'scrap' ? 'error' : 'warning',
      confirmButtonText: '确认提交', cancelButtonText: '取消'
    })
  } catch { return }
  processingId.value = row.ID
  try {
    const res = await submitAssetOperation({ id: row.ID })
    if (res.code === 0) {
      ElMessage.success('业务单已提交')
      await Promise.all([loadOrders(), loadAssetOptions()])
    }
  } finally {
    processingId.value = 0
  }
}

const removeOrder = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除草稿 ${row.orderNo} 吗？`, '删除草稿', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消'
    })
  } catch { return }
  const res = await deleteAssetOperation({ id: row.ID })
  if (res.code === 0) {
    ElMessage.success('草稿已删除')
    loadOrders()
  }
}

const openDetail = async (row) => {
  const res = await getAssetOperationDetail({ id: row.ID })
  if (res.code === 0) {
    detailOrder.value = res.data
    detailVisible.value = true
  }
}

const removeSelectedAsset = (id) => {
  formData.value.assetIds = formData.value.assetIds.filter((assetID) => assetID !== id)
}

watch(operationType, async () => {
  Object.assign(search, { page: 1, pageSize: 10, keyword: '', status: '', dateRange: [] })
  drawerVisible.value = false
  detailVisible.value = false
  await Promise.all([loadOrders(), loadAssetOptions()])
}, { immediate: true })
</script>

<style scoped lang="scss">
.operation-page { min-height: 100%; overflow-x: hidden; padding: 24px; background: var(--na-background); color: var(--na-foreground); }
.page-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; margin-bottom: 20px; }
.eyebrow { margin: 0 0 5px; color: var(--na-primary); font: 600 12px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; }
h1 { margin: 0; font-size: 30px; line-height: 1.2; }
.subtitle { margin: 8px 0 0; color: var(--na-muted-foreground); font-size: 14px; }
.filter-panel, .table-panel { border: 1px solid var(--na-border); border-radius: var(--na-radius); background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.filter-panel { margin-bottom: 16px; padding: 18px 20px 2px; }
.filter-grid { display: grid; grid-template-columns: minmax(240px, 1.4fr) minmax(150px, .7fr) minmax(300px, 1.2fr) auto; gap: 14px; align-items: end; }
.filter-grid :deep(.el-date-editor) { width: 100%; }
.filter-actions { display: flex; gap: 8px; padding-bottom: 18px; }
.table-panel { overflow: hidden; }
.panel-header { display: flex; align-items: center; justify-content: space-between; gap: 18px; padding: 17px 20px; border-bottom: 1px solid var(--na-border); }
.panel-header h2 { margin: 0 0 3px; font-size: 17px; }
.panel-header span { color: var(--na-muted-foreground); font-size: 12px; }
.operation-table { --el-table-header-bg-color: var(--na-table-header); --el-table-row-hover-bg-color: var(--na-table-hover); }
.order-link { padding: 0; border: 0; background: transparent; color: var(--na-primary); font: 600 12px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; }
.order-link:focus-visible { border-radius: 3px; }
.asset-summary, .two-line { display: flex; min-width: 0; flex-direction: column; gap: 4px; }
.asset-summary strong, .asset-summary span, .two-line span, .two-line small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.asset-summary span, .two-line small { color: var(--na-muted-foreground); font-size: 12px; }
.number { font-variant-numeric: tabular-nums; }
.pagination-wrap { display: flex; justify-content: flex-end; padding: 16px 20px; border-top: 1px solid var(--na-border); }
.drawer-title { display: flex; min-width: 0; flex-direction: column; gap: 4px; }
.drawer-title span { color: var(--na-foreground); font-size: 20px; font-weight: 700; }
.drawer-title small { overflow: hidden; color: var(--na-muted-foreground); font-weight: 400; text-overflow: ellipsis; white-space: nowrap; }
.operation-form h3, .order-detail h3 { margin: 6px 0 18px; padding-bottom: 9px; border-bottom: 1px solid var(--na-border); font-size: 15px; }
.operation-form h3:not(:first-child), .order-detail h3 { margin-top: 26px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; column-gap: 18px; }
.form-grid :deep(.el-date-editor), .form-grid :deep(.el-select), .operation-form :deep(.el-select) { width: 100%; }
.selected-assets { overflow: hidden; border: 1px solid var(--na-border); border-radius: var(--na-radius-sm); }
.asset-option { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 20px; }
.asset-option span { display: flex; min-width: 0; gap: 8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.asset-option small { flex: 0 0 auto; color: var(--na-muted-foreground); }
.drawer-actions { display: flex; justify-content: flex-end; gap: 10px; }
.order-detail { max-height: 68vh; overflow-y: auto; }
:deep(.el-form-item__label) { color: var(--na-foreground); font-weight: 600; }

@media (max-width: 1100px) { .filter-grid { grid-template-columns: 1fr 1fr; } }
@media (max-width: 767px) {
  .operation-page { padding: 14px; }
  .page-heading { align-items: stretch; flex-direction: column; }
  .page-heading .el-button { align-self: flex-start; }
  .filter-grid, .form-grid { grid-template-columns: 1fr; }
  .filter-actions { padding-bottom: 16px; }
  .pagination-wrap { overflow-x: auto; justify-content: flex-start; }
  .drawer-actions { flex-wrap: wrap; }
}
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { transition-duration: .01ms !important; } }
</style>
