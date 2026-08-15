<template>
  <main class="na-page na-page--list invoice-ledger">
    <AppPageHeader title-id="invoice-ledger-title" title="发票台账" description="查询、核对和追踪发票状态；只有已确认记录计入正式统计。">
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
      </template>
    </AppPageHeader>

    <section class="na-panel filter-panel">
      <el-form label-position="top" @submit.prevent="submitFilters">
        <div class="filter-grid">
          <el-form-item label="关键词">
            <el-input v-model="search.keyword" clearable placeholder="发票号、销售方或购买方" :prefix-icon="Search" @keyup.enter="submitFilters" />
          </el-form-item>
          <el-form-item label="处理状态">
            <el-select v-model="search.status" clearable placeholder="全部状态">
              <el-option v-for="item in invoiceStatuses" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="发票分类">
            <el-select v-model="search.categoryId" clearable filterable placeholder="全部分类">
              <el-option v-for="item in categories" :key="item.ID" :label="item.name" :value="item.ID" />
            </el-select>
          </el-form-item>
          <el-form-item label="开票日期">
            <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" />
          </el-form-item>
          <div class="filter-actions">
            <el-button :icon="RefreshLeft" @click="resetFilters">重置</el-button>
            <el-button type="primary" :icon="Search" @click="submitFilters">查询</el-button>
          </div>
        </div>
      </el-form>
    </section>

    <section class="na-panel ledger-panel">
      <div class="ledger-header ledger-grid" aria-hidden="true">
        <span>发票信息</span><span>销售方</span><span>分类</span><span>开票日期</span><span>价税合计</span><span>{{ verificationEnabled ? '处理 / 查验' : '处理状态' }}</span><span>操作</span>
      </div>
      <el-skeleton v-if="loading && !loaded" :rows="7" animated />
      <el-result v-else-if="error && !loaded" icon="error" title="发票台账加载失败" :sub-title="error">
        <template #extra><el-button type="primary" :icon="Refresh" @click="load">重新加载</el-button></template>
      </el-result>
      <template v-else>
        <div v-if="error" class="ledger-warning" role="alert">
          <span>刷新失败，当前仍显示上一次成功数据：{{ error }}</span>
          <el-button text :icon="Refresh" @click="load">重试</el-button>
        </div>
        <div v-if="items.length" class="ledger-list" role="list">
          <article v-for="item in items" :key="item.ID" class="ledger-row ledger-grid" role="listitem">
            <button type="button" class="invoice-identity" @click="openReview(item)">
              <strong>{{ item.invoiceNumber || '号码待核对' }}</strong>
              <span>{{ item.invoiceType || item.fileName }}</span>
              <small class="mobile-details">{{ item.sellerName || '销售方待核对' }} · {{ item.category?.name || item.suggestedCategory?.name || '未分类' }} · {{ dateText(item.issueDate) }}</small>
            </button>
            <div class="seller-cell"><strong>{{ item.sellerName || '销售方待核对' }}</strong><span>{{ item.sellerTaxNo || '税号未识别' }}</span></div>
            <div class="category-cell"><i :style="{ backgroundColor: item.category?.color || item.suggestedCategory?.color || 'var(--na-border-strong)' }" /><span>{{ item.category?.name || item.suggestedCategory?.name || '未分类' }}</span></div>
            <time>{{ dateText(item.issueDate) }}</time>
            <strong class="money-cell">{{ money(item.totalCents) }}</strong>
            <div class="status-stack">
              <InvoiceStatusTag :status="item.status" />
              <InvoiceVerificationTag v-if="verificationEnabled" :status="item.verificationStatus" />
            </div>
            <div class="row-actions">
              <el-tooltip content="查看或核对" placement="top">
                <el-button :icon="View" text :disabled="isPending(item.ID)" aria-label="查看或核对发票" @click="openReview(item)" />
              </el-tooltip>
              <el-tooltip v-if="item.status === 'recognition_failed'" content="重新识别" placement="top">
                <el-button :icon="RefreshRight" type="warning" text :loading="isPending(item.ID, 'retry')" :disabled="isPending(item.ID)" aria-label="重新识别发票" @click="retry(item)" />
              </el-tooltip>
              <el-tooltip v-if="item.status === 'confirmed' && canReopen" content="重新打开并编辑" placement="top">
                <el-button :icon="EditPen" type="primary" text :loading="isPending(item.ID, 'reopen')" :disabled="isPending(item.ID)" aria-label="重新打开发票并编辑" @click="reopen(item)" />
              </el-tooltip>
              <el-tooltip content="删除" placement="top">
                <el-button :icon="Delete" type="danger" text :loading="isPending(item.ID, 'delete')" :disabled="isPending(item.ID)" aria-label="删除发票" @click="remove(item)" />
              </el-tooltip>
            </div>
          </article>
        </div>
        <AppEmptyState
          v-else
          compact
          title="没有符合条件的发票"
          description="调整关键词、状态、分类或开票日期后重新查询。"
          :highlights="['当前筛选结果为 0 条']"
        />
        <div v-if="total > 0" class="na-pagination">
          <el-pagination
            v-model:current-page="search.page"
            :page-size="search.pageSize"
            :total="total"
            :pager-count="5"
            layout="total, prev, pager, next"
            @current-change="changePage"
          />
        </div>
      </template>
    </section>

    <InvoiceReviewDrawer v-model="reviewVisible" :invoice-id="selectedId" @saved="load" @confirmed="load" @reopened="load" />
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { Delete, EditPen, Refresh, RefreshLeft, RefreshRight, Search, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import AppEmptyState from '@/components/page/AppEmptyState.vue'
import InvoiceReviewDrawer from '@/plugin/invoice/components/InvoiceReviewDrawer.vue'
import InvoiceStatusTag from '@/plugin/invoice/components/InvoiceStatusTag.vue'
import InvoiceVerificationTag from '@/plugin/invoice/components/InvoiceVerificationTag.vue'
import { deleteInvoice, getInvoiceCapabilities, getInvoiceCategoryOptions, getInvoiceList, reopenInvoice, retryInvoice } from '@/plugin/invoice/api/invoice'
import { centsToCurrency, invoiceDateText, invoiceStatuses } from '@/plugin/invoice/utils/invoice'
import { usePagedList } from '@/hooks/usePagedList'
import { useUserStore } from '@/pinia/modules/user'

defineOptions({ name: 'InvoiceLedger' })

const categories = ref([])
const userStore = useUserStore()
const dateRange = ref([])
const reviewVisible = ref(false)
const selectedId = ref(0)
const pendingActions = ref(new Set())
const verificationEnabled = ref(true)
const money = centsToCurrency
const dateText = invoiceDateText
const canReopen = computed(() => Number(userStore.userInfo.authorityId) === 888)

const {
  search,
  items,
  total,
  loading,
  loaded,
  error,
  load,
  submit,
  reset,
  changePage,
  reloadAfterRemoval
} = usePagedList({
  defaults: { page: 1, pageSize: 20, keyword: '', status: '', categoryId: undefined, direction: '', startDate: '', endDate: '' },
  request: getInvoiceList
})

const submitFilters = () => {
  search.startDate = dateRange.value?.[0] || ''
  search.endDate = dateRange.value?.[1] || ''
  return submit()
}
const resetFilters = () => {
  dateRange.value = []
  return reset()
}

const openReview = (item) => {
  selectedId.value = Number(item.ID)
  reviewVisible.value = true
}

const actionKey = (id, action) => `${action}:${Number(id)}`
const isPending = (id, action) => action
  ? pendingActions.value.has(actionKey(id, action))
  : ['retry', 'reopen', 'delete'].some((name) => pendingActions.value.has(actionKey(id, name)))
const setPending = (id, action, pending) => {
  const next = new Set(pendingActions.value)
  const key = actionKey(id, action)
  if (pending) next.add(key)
  else next.delete(key)
  pendingActions.value = next
}

const retry = async (item) => {
  if (isPending(item.ID)) return
  setPending(item.ID, 'retry', true)
  try {
    const res = await retryInvoice({ id: item.ID })
    if (res.code === 0) {
      ElMessage.success('已重新加入识别队列')
      await load()
    }
  } finally {
    setPending(item.ID, 'retry', false)
  }
}

const reopen = async (item) => {
  if (isPending(item.ID) || !canReopen.value) return
  try {
    await ElMessageBox.confirm(
      `重新打开“${item.invoiceNumber || item.fileName}”后，它会暂时移出正式统计。完成修改后需要再次确认，是否继续？`,
      '重新打开发票',
      { type: 'warning', confirmButtonText: '重新打开', cancelButtonText: '取消' }
    )
  } catch (action) {
    if (action !== 'cancel' && action !== 'close') ElMessage.error(action?.message || '无法重新打开发票')
    return
  }

  setPending(item.ID, 'reopen', true)
  try {
    const res = await reopenInvoice({ id: item.ID })
    if (res.code === 0) {
      ElMessage.success('发票已重新打开，可以继续编辑')
      await load()
      openReview(res.data || item)
    }
  } finally {
    setPending(item.ID, 'reopen', false)
  }
}

const remove = async (item) => {
  if (isPending(item.ID)) return
  setPending(item.ID, 'delete', true)
  try {
    await ElMessageBox.confirm(`确定删除发票“${item.invoiceNumber || item.fileName}”及其原图吗？`, '删除发票', {
      type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    const res = await deleteInvoice({ id: item.ID })
    if (res.code === 0) {
      ElMessage.success('发票已删除')
      await reloadAfterRemoval()
    }
  } catch (action) {
    if (action !== 'cancel' && action !== 'close') ElMessage.error(action?.message || '发票删除失败')
  } finally {
    setPending(item.ID, 'delete', false)
  }
}

const loadCategories = async () => {
  const res = await getInvoiceCategoryOptions()
  if (res.code === 0) categories.value = res.data || []
}

const loadCapabilities = async () => {
  const res = await getInvoiceCapabilities().catch(() => null)
  if (res?.code === 0) verificationEnabled.value = res.data?.verificationEnabled !== false
}

onMounted(() => Promise.all([loadCategories(), loadCapabilities(), load()]))
</script>

<style scoped lang="scss">
.filter-panel { padding: 14px 16px 0; }
.filter-grid { display: grid; min-width: 0; grid-template-columns: minmax(210px, 1.35fr) minmax(140px, .7fr) minmax(150px, .8fr) minmax(250px, 1.2fr) auto; align-items: end; gap: 12px; }
.filter-grid :deep(.el-select), .filter-grid :deep(.el-date-editor) { width: 100%; }
.filter-actions { display: flex; gap: 8px; padding-bottom: 18px; }
.ledger-panel { min-height: 470px; overflow: hidden; }
.ledger-warning { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 12px; padding: 9px 14px; border-bottom: 1px solid var(--na-border); background: var(--na-warning-soft); font-size: .75rem; }
.ledger-warning span { min-width: 0; overflow-wrap: anywhere; }
.ledger-grid { display: grid; min-width: 0; grid-template-columns: minmax(155px, 1.25fr) minmax(150px, 1.25fr) minmax(90px, .7fr) 96px minmax(96px, .7fr) minmax(160px, 1fr) 112px; align-items: center; gap: 12px; }
.ledger-header { min-height: 42px; padding: 0 16px; border-bottom: 1px solid var(--na-border); background: var(--na-table-header); color: var(--na-muted-foreground); font-size: .6875rem; font-weight: 620; }
.ledger-header > span:last-child { text-align: center; }
.ledger-row { min-height: 62px; padding: 8px 16px; border-bottom: 1px solid var(--na-border); transition: background-color 160ms ease; }
.ledger-row:hover { background: var(--na-table-hover); }
.invoice-identity { display: flex; min-width: 0; flex-direction: column; gap: 4px; padding: 0; border: 0; background: transparent; color: var(--na-foreground); text-align: left; }
.invoice-identity:hover strong { color: var(--na-primary); }
.invoice-identity strong, .seller-cell strong { overflow: hidden; font-size: .8125rem; text-overflow: ellipsis; white-space: nowrap; }
.invoice-identity span, .seller-cell span { overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; text-overflow: ellipsis; white-space: nowrap; }
.invoice-identity .mobile-details { display: none; overflow: hidden; color: var(--na-muted-foreground); font-size: .6875rem; font-weight: 400; text-overflow: ellipsis; white-space: nowrap; }
.seller-cell { display: flex; min-width: 0; flex-direction: column; gap: 4px; }
.category-cell { display: flex; min-width: 0; align-items: center; gap: 7px; }
.category-cell i { width: 7px; height: 7px; flex: 0 0 7px; border-radius: 50%; }
.category-cell span { overflow: hidden; font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.ledger-row time { color: var(--na-muted-foreground); font-size: .75rem; }
.money-cell { overflow: hidden; font-size: .75rem; font-variant-numeric: tabular-nums; text-overflow: ellipsis; white-space: nowrap; }
.status-stack { display: flex; min-width: 0; flex-wrap: wrap; align-items: center; gap: 4px; }
.status-stack :deep(.el-tag) { max-width: 100%; }
.row-actions { display: flex; min-width: 0; align-items: center; justify-content: center; gap: 2px; white-space: nowrap; }
.row-actions :deep(.el-button) { width: 30px; min-width: 30px; min-height: 30px; padding: 0; }
.row-actions :deep(.el-button + .el-button) { margin-left: 0; }

@media (max-width: 1280px) {
  .filter-grid { grid-template-columns: repeat(2, minmax(180px, 1fr)); }
  .ledger-grid { grid-template-columns: minmax(150px, 1.3fr) minmax(140px, 1.2fr) minmax(90px, .7fr) minmax(90px, .7fr) minmax(160px, 1fr) 112px; }
  .ledger-grid > :nth-child(4) { display: none; }
}
@media (max-width: 1040px) {
  .ledger-grid { grid-template-columns: minmax(150px, 1.4fr) minmax(140px, 1.2fr) minmax(90px, .7fr) minmax(160px, 1fr) 112px; }
  .ledger-grid > :nth-child(3) { display: none; }
}
@media (max-width: 760px) {
  .filter-grid { grid-template-columns: 1fr; }
  .filter-actions { padding-bottom: 16px; }
  .ledger-header { display: none; }
  .ledger-row { grid-template-columns: minmax(0, 1fr) auto; grid-template-rows: auto auto auto; gap: 5px 10px; padding: 12px 14px; }
  .ledger-row > :nth-child(2), .ledger-row > :nth-child(3), .ledger-row > :nth-child(4) { display: none; }
  .invoice-identity { grid-column: 1; grid-row: 1; }
  .invoice-identity .mobile-details { display: block; }
  .money-cell { grid-column: 1; grid-row: 2; color: var(--na-primary); }
  .status-stack { grid-column: 2; grid-row: 1; justify-content: flex-end; }
  .row-actions { grid-column: 2; grid-row: 2; }
  .na-pagination { overflow: hidden; padding-inline: 8px; }
  .na-pagination :deep(.el-pagination__total) { display: none; }
}
@media (pointer: coarse) {
  .row-actions :deep(.el-button) { width: 44px; min-width: 44px; min-height: 44px; }
}
</style>
