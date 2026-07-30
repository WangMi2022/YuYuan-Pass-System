<template>
  <el-drawer
    :model-value="modelValue"
    :size="drawerSize"
    destroy-on-close
    :close-on-click-modal="false"
    class="invoice-review-drawer"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <template #header>
      <div class="review-heading">
        <div>
          <span>发票核对</span>
          <small>{{ invoice.fileName || '识别结果与原始凭证逐项核对' }}</small>
        </div>
        <InvoiceStatusTag :status="invoice.status" />
      </div>
    </template>

    <el-skeleton v-if="loading" :rows="9" animated />
    <el-result
      v-else-if="loadError"
      icon="error"
      title="发票加载失败"
      :sub-title="loadError"
      class="review-error"
    >
      <template #extra><el-button type="primary" :icon="RefreshRight" @click="loadInvoice">重新加载</el-button></template>
    </el-result>
    <div v-else class="review-workbench">
      <section class="evidence-pane" aria-label="发票原图">
        <div class="pane-heading">
          <div><h3>原始凭证</h3><p>文件仅通过当前登录权限读取</p></div>
          <el-button :icon="View" text :disabled="!invoice.ID" @click="previewVisible = true">查看大图</el-button>
        </div>
        <div class="invoice-image-wrap">
          <el-image
            v-if="invoice.ID"
            :src="fileUrl"
            fit="contain"
            :preview-src-list="[fileUrl]"
            preview-teleported
            alt="待核对的发票原图"
          >
            <template #error><el-empty description="原图加载失败" :image-size="72" /></template>
          </el-image>
        </div>
        <dl class="evidence-meta">
          <div><dt>识别方式</dt><dd>{{ invoice.recognitionProvider || '等待识别' }}</dd></div>
          <div><dt>整体置信度</dt><dd>{{ confidenceText }}</dd></div>
          <div><dt>分类依据</dt><dd>{{ invoice.classificationReason || '暂无规则命中' }}</dd></div>
        </dl>
        <el-alert
          v-if="invoice.recognitionError"
          :title="invoice.recognitionError"
          type="error"
          :closable="false"
          show-icon
        />
      </section>

      <section class="fields-pane" aria-label="结构化发票字段">
        <div class="pane-heading">
          <div>
            <h3>结构化字段</h3>
            <p v-if="readonly">该发票已确认并进入正式统计{{ canReopen ? '；重新打开后可继续编辑' : '' }}</p>
            <p v-else-if="lowConfidenceText" id="invoice-low-confidence">低置信度：{{ lowConfidenceText }}，建议重点核对</p>
            <p v-else>识别结果需经人工确认后才进入统计</p>
          </div>
          <div class="field-heading-actions">
            <span v-if="invoice.suggestedCategory" class="category-hint">
              建议：{{ invoice.suggestedCategory.name }}
            </span>
            <el-button
              v-if="!readonly"
              type="primary"
              plain
              :icon="RefreshRight"
              :loading="rechecking"
              :disabled="saving || confirming || rechecking || verifying || !!loadError"
              title="重新读取原图并用多模态模型核对，结果只回填当前表单"
              @click="recheck"
            >
              重新核对
            </el-button>
          </div>
        </div>

        <section class="verification-band" aria-label="发票权威查验">
          <div class="verification-summary">
            <div>
              <span class="verification-title">权威查验</span>
              <InvoiceVerificationTag :status="invoice.verificationStatus" />
            </div>
            <el-button
              v-if="!readonly"
              type="success"
              plain
              :icon="CircleCheck"
              :loading="verifying"
              :disabled="saving || confirming || rechecking || verifying || reopening || !!loadError"
              @click="verifyCurrent"
            >
              保存并查验
            </el-button>
          </div>
          <p class="verification-message">{{ verificationMessage }}</p>
          <ul v-if="latestVerification?.differences?.length" class="verification-differences">
            <li v-for="difference in latestVerification.differences" :key="difference.field">
              <strong>{{ difference.label }}</strong>
              <span>当前：{{ difference.localValue || '空' }}</span>
              <span>税局：{{ difference.officialValue || '空' }}</span>
            </li>
          </ul>
          <details v-if="verificationHistory.length" class="verification-history">
            <summary>查验记录 <span>{{ verificationHistory.length }}</span></summary>
            <ol>
              <li v-for="attempt in verificationHistory" :key="attempt.ID">
                <div>
                  <InvoiceVerificationTag :status="attempt.status" />
                  <time>{{ historyTime(attempt.completedAt || attempt.createdAt) }}</time>
                </div>
                <p>{{ attempt.verifyMessage || '未返回查验说明' }}</p>
                <small v-if="attempt.providerLogId">Log ID：{{ attempt.providerLogId }}</small>
              </li>
            </ol>
          </details>
        </section>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          :disabled="readonly || rechecking"
          :aria-describedby="lowConfidenceText ? 'invoice-low-confidence' : undefined"
        >
          <div class="field-grid">
            <el-form-item label="流水方向" prop="direction">
              <el-segmented v-model="form.direction" :options="directionOptions" block />
            </el-form-item>
            <el-form-item label="发票分类" prop="categoryId" :class="confidenceClass('categoryId')">
              <el-select v-model="form.categoryId" placeholder="选择分类" filterable>
                <el-option v-for="item in categories" :key="item.ID" :label="item.name" :value="item.ID" />
              </el-select>
            </el-form-item>
            <el-form-item label="发票类型" :class="confidenceClass('invoiceType')">
              <el-input v-model="form.invoiceType" maxlength="60" aria-label="发票类型" />
            </el-form-item>
            <el-form-item label="验真票种" prop="verificationType" :class="confidenceClass('verificationType')">
              <el-select v-model="form.verificationType" filterable placeholder="选择验真票种">
                <el-option v-for="item in verificationTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
            <el-form-item
              v-if="form.verificationType === 'motor_vehicle_invoice'"
              label="机动车验真口径"
              prop="verificationAmountMode"
              class="verification-amount-mode"
            >
              <el-segmented v-model="form.verificationAmountMode" :options="verificationAmountModeOptions" block />
            </el-form-item>
            <el-form-item label="开票日期" prop="issueDate" :class="confidenceClass('issueDate')">
              <el-date-picker v-model="form.issueDate" type="date" placeholder="选择开票日期" />
            </el-form-item>
            <el-form-item label="发票代码" :class="confidenceClass('invoiceCode')">
              <el-input v-model="form.invoiceCode" maxlength="80" aria-label="发票代码" />
            </el-form-item>
            <el-form-item label="发票号码" prop="invoiceNumber" :class="confidenceClass('invoiceNumber')">
              <el-input v-model="form.invoiceNumber" maxlength="80" aria-label="发票号码" />
            </el-form-item>
            <el-form-item label="校验码" :class="confidenceClass('checkCode')">
              <el-input v-model="form.checkCode" maxlength="80" aria-label="发票校验码" />
            </el-form-item>
            <el-form-item label="销售方名称" prop="sellerName" :class="confidenceClass('sellerName')">
              <el-input v-model="form.sellerName" maxlength="200" aria-label="销售方名称" />
            </el-form-item>
            <el-form-item label="销售方税号" :class="confidenceClass('sellerTaxNo')">
              <el-input v-model="form.sellerTaxNo" maxlength="80" aria-label="销售方税号" />
            </el-form-item>
            <el-form-item label="购买方名称" :class="confidenceClass('buyerName')">
              <el-input v-model="form.buyerName" maxlength="200" aria-label="购买方名称" />
            </el-form-item>
            <el-form-item label="购买方税号" :class="confidenceClass('buyerTaxNo')">
              <el-input v-model="form.buyerTaxNo" maxlength="80" aria-label="购买方税号" />
            </el-form-item>
          </div>

          <div class="amount-strip">
            <el-form-item label="不含税金额（元）" prop="amount" :class="confidenceClass('amountCents')">
              <el-input-number v-model="form.amount" :min="0" :precision="2" :controls="false" />
            </el-form-item>
            <el-form-item label="税额（元）" prop="tax" :class="confidenceClass('taxCents')">
              <el-input-number v-model="form.tax" :min="0" :precision="2" :controls="false" />
            </el-form-item>
            <el-form-item label="价税合计（元）" prop="total" :class="confidenceClass('totalCents')">
              <el-input-number v-model="form.total" :min="0" :precision="2" :controls="false" />
            </el-form-item>
          </div>

          <div class="items-heading">
            <div><h4>发票明细</h4><span>商品或服务将参与分类规则评分</span></div>
            <el-button v-if="!readonly" :icon="Plus" text @click="addItem">添加明细</el-button>
          </div>
          <div v-if="form.items.length" class="item-list">
            <div v-for="(item, index) in form.items" :key="item.key" class="item-row">
              <el-input v-model="item.name" placeholder="商品或服务名称" aria-label="商品或服务名称" />
              <el-input v-model="item.specification" placeholder="规格" aria-label="规格" />
              <el-input v-model="item.quantityText" placeholder="数量" aria-label="数量" />
              <el-input-number v-model="item.amount" :min="0" :precision="2" :controls="false" placeholder="金额" aria-label="明细金额" />
              <el-button v-if="!readonly" :icon="Delete" type="danger" text aria-label="删除明细" @click="removeItem(index)" />
            </div>
          </div>
          <el-empty v-else description="暂未识别到明细，可按需补充" :image-size="62" />

          <el-form-item label="核对备注" class="review-notes">
            <el-input v-model="form.reviewNotes" type="textarea" :rows="3" maxlength="1000" show-word-limit />
          </el-form-item>
        </el-form>
      </section>
    </div>

    <template #footer>
      <div class="drawer-actions">
        <el-button @click="$emit('update:modelValue', false)">关闭</el-button>
        <el-button
          v-if="readonly && canReopen"
          type="primary"
          plain
          :icon="EditPen"
          :loading="reopening"
          :disabled="reopening || !!loadError"
          @click="reopen"
        >
          重新打开
        </el-button>
        <template v-if="!readonly">
          <el-button :loading="saving" :disabled="saving || confirming || rechecking || verifying || reopening || !!loadError" @click="save(false)">保存核对</el-button>
          <el-tooltip :disabled="verificationReady" content="请先完成当前字段的权威查验" placement="top">
            <span>
              <el-button type="primary" :loading="confirming" :disabled="saving || confirming || rechecking || verifying || reopening || !verificationReady || !!loadError" @click="save(true)">保存并确认</el-button>
            </span>
          </el-tooltip>
        </template>
      </div>
    </template>
  </el-drawer>

  <el-dialog v-model="previewVisible" title="发票原图" width="min(94vw, 1080px)" append-to-body>
    <img class="preview-image" :src="fileUrl" alt="发票原图大图" />
  </el-dialog>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { CircleCheck, Delete, EditPen, Plus, RefreshRight, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { confirmInvoice, getInvoiceCategoryOptions, getInvoiceDetail, getInvoiceVerificationHistory, recheckInvoice, reopenInvoice, updateInvoice, verifyInvoice } from '@/plugin/invoice/api/invoice'
import { centsToYuan, invoiceFileUrl, yuanToCents } from '@/plugin/invoice/utils/invoice'
import InvoiceStatusTag from '@/plugin/invoice/components/InvoiceStatusTag.vue'
import InvoiceVerificationTag from '@/plugin/invoice/components/InvoiceVerificationTag.vue'
import { useAppStore } from '@/pinia/modules/app'
import { useUserStore } from '@/pinia/modules/user'

defineOptions({ name: 'InvoiceReviewDrawer' })

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  invoiceId: { type: Number, default: 0 }
})

const emit = defineEmits(['update:modelValue', 'saved', 'confirmed', 'reopened'])
const appStore = useAppStore()
const userStore = useUserStore()
const formRef = ref()
const loading = ref(false)
const saving = ref(false)
const confirming = ref(false)
const rechecking = ref(false)
const verifying = ref(false)
const reopening = ref(false)
const previewVisible = ref(false)
const invoice = ref({})
const categories = ref([])
const verificationHistory = ref([])
const verifiedFormSignature = ref('')
const loadError = ref('')
let itemKey = 0
let loadRequestId = 0
let recheckRequestId = 0

const emptyForm = () => ({
  ID: 0, direction: 'expense', invoiceType: '', verificationType: '', verificationAmountMode: '', invoiceCode: '', invoiceNumber: '', checkCode: '', issueDate: null,
  buyerName: '', buyerTaxNo: '', sellerName: '', sellerTaxNo: '', amount: 0, tax: 0, total: 0,
  categoryId: undefined, reviewNotes: '', items: []
})
const form = reactive(emptyForm())
const directionOptions = [{ label: '支出', value: 'expense' }, { label: '收入', value: 'income' }]
const verificationTypeOptions = [
  { value: 'special_vat_invoice', label: '增值税专用发票' },
  { value: 'elec_special_vat_invoice', label: '增值税电子专用发票' },
  { value: 'normal_invoice', label: '增值税普通发票' },
  { value: 'elec_normal_invoice', label: '增值税普通发票（电子）' },
  { value: 'roll_normal_invoice', label: '增值税普通发票（卷式）' },
  { value: 'toll_elec_normal_invoice', label: '通行费增值税电子普通发票' },
  { value: 'blockchain_invoice', label: '区块链电子发票' },
  { value: 'elec_invoice_special', label: '全电发票（专用发票）' },
  { value: 'elec_invoice_normal', label: '全电发票（普通发票）' },
  { value: 'special_freight_transport_invoice', label: '货物运输业增值税专用发票' },
  { value: 'motor_vehicle_invoice', label: '机动车销售统一发票' },
  { value: 'used_vehicle_invoice', label: '二手车销售统一发票' },
  { value: 'elec_flight_itinerary_invoice', label: '航空运输电子客票行程单' },
  { value: 'elec_train_ticket_invoice', label: '铁路电子客票' },
  { value: 'elec_toll_invoice', label: '全电发票（含通行费标识）' }
]
const verificationAmountModeOptions = [
  { label: '纸质票 · 不含税金额', value: 'amount' },
  { label: '电子票 · 价税合计', value: 'total' }
]
const verificationCheckCodeTypes = new Set([
  'elec_special_vat_invoice', 'normal_invoice', 'elec_normal_invoice',
  'roll_normal_invoice', 'blockchain_invoice', 'toll_elec_normal_invoice'
])
const verificationAmountTypes = new Set([
  'special_vat_invoice', 'elec_special_vat_invoice', 'blockchain_invoice', 'special_freight_transport_invoice'
])
const verificationTotalTypes = new Set([
  'used_vehicle_invoice', 'elec_invoice_special', 'elec_invoice_normal',
  'elec_train_ticket_invoice', 'elec_flight_itinerary_invoice', 'elec_toll_invoice'
])
const verificationInvoiceCodeOptionalTypes = new Set([
  'elec_invoice_special', 'elec_invoice_normal', 'used_vehicle_invoice',
  'elec_train_ticket_invoice', 'elec_flight_itinerary_invoice', 'elec_toll_invoice'
])
const rules = {
  direction: [{ required: true, message: '请选择流水方向', trigger: 'change' }],
  categoryId: [{ required: true, message: '请选择发票分类', trigger: 'change' }],
  verificationType: [{ required: true, message: '请选择验真票种', trigger: 'change' }],
  verificationAmountMode: [{ required: true, message: '请选择机动车发票验真口径', trigger: 'change' }],
  issueDate: [{ required: true, message: '请选择开票日期', trigger: 'change' }],
  invoiceNumber: [{ required: true, message: '请输入发票号码', trigger: 'blur' }],
  sellerName: [{ required: true, message: '请输入销售方名称', trigger: 'blur' }],
  total: [{ required: true, message: '请输入价税合计', trigger: 'change' }]
}

const drawerSize = computed(() => appStore.drawerSize === '100%' ? '100%' : 'min(94vw, 1040px)')
const readonly = computed(() => invoice.value.status === 'confirmed')
const canReopen = computed(() => Number(userStore.userInfo.authorityId) === 888)
const fileUrl = computed(() => invoice.value.ID ? invoiceFileUrl(invoice.value.ID) : '')
const confidenceText = computed(() => {
  const value = Number(invoice.value.recognitionConfidence || 0)
  return value > 0 ? `${Math.round(value * 100)}%` : '待人工核对'
})

const confidenceLabels = {
  categoryId: '发票分类', invoiceType: '发票类型', verificationType: '验真票种', verificationAmountMode: '机动车验真口径', checkCode: '校验码', issueDate: '开票日期',
  invoiceCode: '发票代码', invoiceNumber: '发票号码', sellerName: '销售方名称',
  sellerTaxNo: '销售方税号', buyerName: '购买方名称', buyerTaxNo: '购买方税号',
  amountCents: '不含税金额', taxCents: '税额', totalCents: '价税合计'
}

const isLowConfidence = (field) => {
  const fields = invoice.value.fieldConfidences || {}
  if (Object.prototype.hasOwnProperty.call(fields, field)) return Number(fields[field]) < 0.75
  const overall = Number(invoice.value.recognitionConfidence || 0)
  return overall > 0 && overall < 0.65
}

const lowConfidenceText = computed(() => Object.entries(confidenceLabels)
  .filter(([field]) => isLowConfidence(field))
  .map(([, label]) => label)
  .join('、'))

const confidenceClass = (field) => isLowConfidence(field) ? 'is-low-confidence' : ''

const formVerificationSignature = () => JSON.stringify({
  invoiceType: form.invoiceType,
  verificationType: form.verificationType,
  verificationAmountMode: form.verificationAmountMode,
  invoiceCode: form.invoiceCode,
  invoiceNumber: form.invoiceNumber,
  checkCode: form.checkCode,
  issueDate: form.issueDate ? new Date(form.issueDate).toISOString().slice(0, 10) : '',
  buyerName: form.buyerName,
  buyerTaxNo: form.buyerTaxNo,
  sellerName: form.sellerName,
  sellerTaxNo: form.sellerTaxNo,
  amount: Number(form.amount || 0),
  tax: Number(form.tax || 0),
  total: Number(form.total || 0),
  items: form.items.map((item) => ({
    name: item.name, specification: item.specification, unit: item.unit,
    quantityText: item.quantityText, amount: Number(item.amount || 0),
    unitPriceCents: Number(item.unitPriceCents || 0), taxRate: item.taxRate,
    taxCents: Number(item.taxCents || 0)
  }))
})

const verificationReady = computed(() =>
  invoice.value.verificationStatus === 'verified_valid' &&
  Boolean(verifiedFormSignature.value) &&
  verifiedFormSignature.value === formVerificationSignature()
)
const latestVerification = computed(() => verificationHistory.value[0] || null)
const verificationMessage = computed(() => {
  if (invoice.value.verificationMessage) return invoice.value.verificationMessage
  if (invoice.value.verificationStatus === 'verified_valid') return '税局信息与当前发票字段一致'
  if (invoice.value.verificationStatus === 'verifying') return '正在连接税局查验'
  return '尚未取得与当前字段一致的权威查验结果'
})
const historyTime = (value) => value
  ? new Date(value).toLocaleString('zh-CN', { hour12: false })
  : '处理中'

const fillForm = (data) => {
  Object.assign(form, emptyForm(), {
    ...data,
    issueDate: data.issueDate ? new Date(data.issueDate) : null,
    amount: centsToYuan(data.amountCents),
    tax: centsToYuan(data.taxCents),
    total: centsToYuan(data.totalCents),
    categoryId: data.categoryId || data.suggestedCategoryId || undefined,
    items: (data.items || []).map((item) => ({
      ...item,
      key: `${item.ID || 0}-${itemKey++}`,
      amount: centsToYuan(item.amountCents)
    }))
  })
  verifiedFormSignature.value = data.verificationStatus === 'verified_valid'
    ? formVerificationSignature()
    : ''
}

const recheckedAmount = (data, field, currentValue) => {
  const value = Number(data[field] || 0)
  const hasModelValue = value > 0 || Object.prototype.hasOwnProperty.call(data.fieldConfidences || {}, field)
  return hasModelValue ? centsToYuan(value) : currentValue
}

const fillRecheckResult = (data) => {
  Object.assign(form, {
    invoiceType: data.invoiceType || form.invoiceType,
    verificationType: data.verificationType || form.verificationType,
    verificationAmountMode: data.verificationAmountMode || form.verificationAmountMode,
    invoiceCode: data.invoiceCode || form.invoiceCode,
    invoiceNumber: data.invoiceNumber || form.invoiceNumber,
    checkCode: data.checkCode || form.checkCode,
    issueDate: data.issueDate ? new Date(data.issueDate) : form.issueDate,
    buyerName: data.buyerName || form.buyerName,
    buyerTaxNo: data.buyerTaxNo || form.buyerTaxNo,
    sellerName: data.sellerName || form.sellerName,
    sellerTaxNo: data.sellerTaxNo || form.sellerTaxNo,
    amount: recheckedAmount(data, 'amountCents', form.amount),
    tax: recheckedAmount(data, 'taxCents', form.tax),
    total: recheckedAmount(data, 'totalCents', form.total),
    items: (data.items || []).length
      ? data.items.map((item) => ({
        ...item,
        key: `recheck-${itemKey++}`,
        amount: centsToYuan(item.amountCents)
      }))
      : form.items
  })
  invoice.value = {
    ...invoice.value,
    recognitionProvider: data.provider || invoice.value.recognitionProvider,
    recognitionConfidence: Number(data.confidence || 0),
    fieldConfidences: data.fieldConfidences || {}
  }
  formRef.value?.clearValidate()
}

const loadCategories = async () => {
  if (categories.value.length) return
  const res = await getInvoiceCategoryOptions()
  if (res.code === 0) categories.value = res.data || []
}

const loadInvoice = async () => {
  if (!props.invoiceId) return
  const requestId = ++loadRequestId
  const requestedInvoiceId = Number(props.invoiceId)
  invoice.value = {}
  verificationHistory.value = []
  verifiedFormSignature.value = ''
  Object.assign(form, emptyForm())
  formRef.value?.clearValidate()
  previewVisible.value = false
  loadError.value = ''
  loading.value = true
  try {
    const [res, historyRes] = await Promise.all([
      getInvoiceDetail({ id: requestedInvoiceId }),
      getInvoiceVerificationHistory({ id: requestedInvoiceId }),
      loadCategories()
    ])
    if (requestId !== loadRequestId || requestedInvoiceId !== Number(props.invoiceId)) return
    if (res.code === 0) {
      invoice.value = res.data || {}
      verificationHistory.value = historyRes?.code === 0 ? (historyRes.data || []) : []
      fillForm(invoice.value)
    } else {
      loadError.value = res.msg || '无法读取这张发票，请检查权限或稍后重试'
    }
  } catch (error) {
    if (requestId !== loadRequestId) return
    loadError.value = error?.message || '无法读取这张发票，请稍后重试'
  } finally {
    if (requestId === loadRequestId) loading.value = false
  }
}

const addItem = () => form.items.push({ key: `new-${itemKey++}`, name: '', specification: '', unit: '', quantityText: '', amount: 0, taxRate: '', taxCents: 0, unitPriceCents: 0 })
const removeItem = (index) => form.items.splice(index, 1)

const recheck = async () => {
  if (rechecking.value || verifying.value || reopening.value || !form.ID || readonly.value) return
  const requestId = ++recheckRequestId
  const invoiceId = Number(form.ID)
  rechecking.value = true
  try {
    const res = await recheckInvoice({ id: invoiceId })
    if (requestId !== recheckRequestId || invoiceId !== Number(props.invoiceId) || !props.modelValue) return
    if (res.code === 0) {
      fillRecheckResult(res.data || {})
      ElMessage.success('模型核对完成，识别字段已回填；保存前请再次确认')
    }
  } finally {
    if (requestId === recheckRequestId) rechecking.value = false
  }
}

const validateVerificationFields = async () => {
  try {
    await formRef.value?.validateField(['verificationType', 'invoiceNumber', 'issueDate'])
  } catch {
    return false
  }
  const type = form.verificationType
  let amountMode = ''
  if (verificationAmountTypes.has(type)) amountMode = 'amount'
  if (verificationTotalTypes.has(type)) amountMode = 'total'
  if (type === 'motor_vehicle_invoice') {
    amountMode = form.verificationAmountMode
    if (!amountMode) {
      ElMessage.warning('请选择机动车发票是纸质票还是电子票')
      return false
    }
  }
  const invoiceCodeOptional = verificationInvoiceCodeOptionalTypes.has(type) ||
    (type === 'motor_vehicle_invoice' && amountMode === 'total')
  if (!invoiceCodeOptional && !String(form.invoiceCode || '').trim()) {
    ElMessage.warning('该票种验真需要填写发票代码')
    return false
  }
  if (verificationCheckCodeTypes.has(type) && String(form.checkCode || '').trim().slice(-6).length !== 6) {
    ElMessage.warning('该票种验真需要填写校验码后 6 位')
    return false
  }
  if (amountMode === 'amount' && Number(form.amount || 0) <= 0) {
    ElMessage.warning('该票种验真需要填写不含税金额')
    return false
  }
  if (amountMode === 'total' && Number(form.total || 0) <= 0) {
    ElMessage.warning('该票种验真需要填写价税合计')
    return false
  }
  return true
}

const verifyCurrent = async () => {
  if (verifying.value || saving.value || confirming.value || rechecking.value || reopening.value || !form.ID || readonly.value) return
  const valid = await validateVerificationFields()
  if (!valid) return
  verifying.value = true
  try {
    const saved = await updateInvoice(buildPayload())
    if (saved.code !== 0) return
    invoice.value = saved.data || invoice.value
    emit('saved', invoice.value)
    const result = await verifyInvoice({ id: form.ID })
    if (result.code !== 0) return
    invoice.value = result.data?.invoice || invoice.value
    fillForm(invoice.value)
    if (result.data?.attempt) {
      verificationHistory.value = [result.data.attempt, ...verificationHistory.value.filter((item) => item.ID !== result.data.attempt.ID)]
    }
    if (invoice.value.verificationStatus === 'verified_valid') {
      ElMessage.success('发票查验一致，当前字段可以确认入账')
    } else {
      ElMessage.warning(invoice.value.verificationMessage || '发票未通过权威查验')
    }
  } finally {
    verifying.value = false
  }
}

const reopen = async () => {
  if (reopening.value || !form.ID || !readonly.value || !canReopen.value) return
  try {
    await ElMessageBox.confirm(
      '重新打开后，这张发票将暂时移出正式统计。完成修改后需要再次确认，是否继续？',
      '重新打开发票',
      { type: 'warning', confirmButtonText: '重新打开', cancelButtonText: '取消' }
    )
  } catch (action) {
    if (action !== 'cancel' && action !== 'close') ElMessage.error(action?.message || '无法重新打开发票')
    return
  }

  const invoiceId = Number(form.ID)
  reopening.value = true
  try {
    const res = await reopenInvoice({ id: invoiceId })
    if (res.code !== 0) return
    emit('reopened', res.data || {})
    if (invoiceId === Number(props.invoiceId) && props.modelValue) {
      invoice.value = res.data || invoice.value
      fillForm(invoice.value)
      formRef.value?.clearValidate()
    }
    ElMessage.success('发票已重新打开，可以继续编辑')
  } finally {
    reopening.value = false
  }
}

const buildPayload = () => ({
  ID: form.ID,
  direction: form.direction,
  invoiceType: form.invoiceType,
  verificationType: form.verificationType,
  verificationAmountMode: form.verificationAmountMode,
  invoiceCode: form.invoiceCode,
  invoiceNumber: form.invoiceNumber,
  checkCode: form.checkCode,
  issueDate: form.issueDate ? new Date(form.issueDate).toISOString() : null,
  buyerName: form.buyerName,
  buyerTaxNo: form.buyerTaxNo,
  sellerName: form.sellerName,
  sellerTaxNo: form.sellerTaxNo,
  amountCents: yuanToCents(form.amount),
  taxCents: yuanToCents(form.tax),
  totalCents: yuanToCents(form.total),
  categoryId: form.categoryId,
  reviewNotes: form.reviewNotes,
  items: form.items.map((item) => ({
    name: item.name,
    specification: item.specification,
    unit: item.unit || '',
    quantityText: item.quantityText,
    unitPriceCents: Number(item.unitPriceCents || 0),
    amountCents: yuanToCents(item.amount),
    taxRate: item.taxRate || '',
    taxCents: Number(item.taxCents || 0)
  }))
})

const save = async (andConfirm) => {
  if (saving.value || confirming.value || reopening.value) return
  if (loadError.value || Number(form.ID) !== Number(props.invoiceId)) {
    ElMessage.error('当前发票尚未正确加载，请重新打开后再操作')
    return
  }
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = !andConfirm
  confirming.value = andConfirm
  try {
    const res = await updateInvoice(buildPayload())
    if (res.code !== 0) return
    invoice.value = res.data || invoice.value
    emit('saved', invoice.value)
    if (!andConfirm) {
      ElMessage.success('核对信息已保存')
      await loadInvoice()
      return
    }
    if (invoice.value.verificationStatus !== 'verified_valid') {
      verifiedFormSignature.value = ''
      ElMessage.warning('当前字段尚未通过权威查验，请先执行保存并查验')
      return
    }
    const confirmRes = await confirmInvoice({ id: form.ID })
    if (confirmRes.code === 0) {
      invoice.value = confirmRes.data || invoice.value
      ElMessage.success('发票已确认并纳入统计')
      emit('confirmed', invoice.value)
      emit('update:modelValue', false)
    }
  } finally {
    saving.value = false
    confirming.value = false
  }
}

watch(() => [props.modelValue, props.invoiceId], ([visible, id]) => {
  if (visible && id) {
    recheckRequestId++
    rechecking.value = false
    verifying.value = false
    reopening.value = false
    loadInvoice()
  } else if (!visible) {
    loadRequestId++
    recheckRequestId++
    loading.value = false
    rechecking.value = false
    verifying.value = false
    reopening.value = false
    verificationHistory.value = []
    verifiedFormSignature.value = ''
    loadError.value = ''
  }
}, { immediate: true })

watch(() => form.verificationType, (type) => {
  if (type !== 'motor_vehicle_invoice') form.verificationAmountMode = ''
})
</script>

<style scoped lang="scss">
.review-heading { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 14px; padding-right: 30px; }
.review-heading > div { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.review-heading span { color: var(--na-foreground); font-size: 1.125rem; font-weight: 650; }
.review-heading small { overflow: hidden; color: var(--na-muted-foreground); font-size: .75rem; font-weight: 400; text-overflow: ellipsis; white-space: nowrap; }
.review-error { min-height: 420px; }
.review-workbench { display: grid; min-width: 0; grid-template-columns: minmax(300px, .86fr) minmax(430px, 1.14fr); gap: 16px; }
.evidence-pane, .fields-pane { min-width: 0; }
.evidence-pane { align-self: start; padding: 14px; border: 1px solid var(--na-border); border-radius: 12px; background: var(--na-muted); }
.fields-pane { padding: 2px 4px; }
.pane-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
.pane-heading h3, .items-heading h4 { margin: 0; color: var(--na-foreground); font-size: .9375rem; font-weight: 650; }
.pane-heading p, .items-heading span { margin: 3px 0 0; color: var(--na-muted-foreground); font-size: .75rem; }
.invoice-image-wrap { display: grid; min-height: 460px; overflow: hidden; place-items: center; border: 1px solid var(--na-border); border-radius: 9px; background: var(--na-card); }
.invoice-image-wrap :deep(.el-image) { width: 100%; height: 460px; }
.evidence-meta { display: grid; gap: 0; margin: 12px 0; }
.evidence-meta > div { display: grid; grid-template-columns: 84px minmax(0, 1fr); gap: 8px; padding: 8px 0; border-bottom: 1px solid var(--na-border); }
.evidence-meta dt { color: var(--na-muted-foreground); font-size: .75rem; }
.evidence-meta dd { margin: 0; overflow-wrap: anywhere; color: var(--na-foreground); font-size: .75rem; }
.category-hint { padding: 5px 8px; border-radius: 7px; background: var(--na-primary-soft); color: var(--na-primary); font-size: .75rem; }
.field-heading-actions { display: flex; flex: 0 0 auto; flex-wrap: wrap; align-items: center; justify-content: flex-end; gap: 8px; }
.verification-band { margin: 4px 0 18px; padding: 12px 0; border-block: 1px solid var(--na-border); }
.verification-summary { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.verification-summary > div { display: flex; min-width: 0; align-items: center; gap: 9px; }
.verification-title { color: var(--na-foreground); font-size: .8125rem; font-weight: 650; }
.verification-message { margin: 7px 0 0; color: var(--na-muted-foreground); font-size: .75rem; line-height: 1.55; }
.verification-differences { display: grid; gap: 0; margin: 10px 0 0; padding: 0; list-style: none; border-top: 1px solid var(--na-border); }
.verification-differences li { display: grid; grid-template-columns: 88px minmax(0, 1fr) minmax(0, 1fr); gap: 8px; padding: 8px 0; border-bottom: 1px solid var(--na-border); font-size: .75rem; }
.verification-differences strong { color: var(--na-danger); }
.verification-differences span { min-width: 0; overflow-wrap: anywhere; color: var(--na-foreground); }
.verification-history { margin-top: 10px; border-top: 1px solid var(--na-border); }
.verification-history summary { display: flex; align-items: center; gap: 6px; padding: 9px 0 0; color: var(--na-foreground); font-size: .75rem; font-weight: 600; cursor: pointer; }
.verification-history summary span { color: var(--na-muted-foreground); font-variant-numeric: tabular-nums; }
.verification-history ol { display: grid; gap: 0; max-height: 240px; margin: 8px 0 0; padding: 0; overflow: auto; list-style: none; }
.verification-history li { padding: 9px 0; border-top: 1px solid var(--na-border); }
.verification-history li > div { display: flex; align-items: center; justify-content: space-between; gap: 10px; }
.verification-history time, .verification-history small { color: var(--na-muted-foreground); font-size: .6875rem; }
.verification-history p { margin: 5px 0 2px; color: var(--na-foreground); font-size: .75rem; line-height: 1.5; }
.field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0 14px; }
.field-grid :deep(.el-select), .field-grid :deep(.el-date-editor), .amount-strip :deep(.el-input-number) { width: 100%; }
.verification-amount-mode { grid-column: 1 / -1; }
.verification-amount-mode :deep(.el-segmented) { width: 100%; }
.is-low-confidence :deep(.el-input__wrapper), .is-low-confidence :deep(.el-select__wrapper), .is-low-confidence :deep(.el-date-editor) { border-color: color-mix(in srgb, var(--na-warning) 58%, var(--na-border)); background: var(--na-warning-soft); }
.is-low-confidence :deep(.el-form-item__label) { color: var(--na-warning); font-weight: 600; }
.amount-strip { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin: 4px 0 18px; padding: 14px; border: 1px solid var(--na-border); border-radius: 10px; background: var(--na-muted); }
.amount-strip :deep(.el-form-item) { margin-bottom: 0; }
.items-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin: 2px 0 10px; }
.items-heading > div { min-width: 0; }
.item-list { display: grid; gap: 7px; }
.item-row { display: grid; min-width: 0; grid-template-columns: minmax(130px, 1.5fr) minmax(90px, 1fr) 70px 100px 32px; gap: 7px; }
.item-row :deep(.el-input-number) { width: 100%; }
.review-notes { margin-top: 18px; }
.drawer-actions { display: flex; justify-content: flex-end; gap: 8px; }
.preview-image { display: block; width: 100%; max-height: 76vh; object-fit: contain; background: var(--na-muted); }

@media (max-width: 900px) {
  .review-workbench { grid-template-columns: 1fr; }
  .invoice-image-wrap, .invoice-image-wrap :deep(.el-image) { min-height: 340px; height: 340px; }
}
@media (max-width: 620px) {
  .field-grid, .amount-strip { grid-template-columns: 1fr; }
  .item-row { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 44px; padding: 10px; border: 1px solid var(--na-border); border-radius: 9px; background: var(--na-muted); }
  .item-row > :first-child { grid-column: 1 / -1; }
  .item-row > :nth-child(4) { grid-column: 1 / span 2; }
  .item-row > :last-child { grid-column: 3; grid-row: 2 / span 2; align-self: stretch; min-height: 44px; }
  .verification-summary { align-items: flex-start; flex-direction: column; }
  .verification-differences li { grid-template-columns: 72px minmax(0, 1fr); }
  .verification-differences li span:last-child { grid-column: 2; }
}
@media (pointer: coarse) {
  .item-row > :last-child { min-width: 44px; min-height: 44px; }
}
</style>
