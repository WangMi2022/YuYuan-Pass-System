const statusLabels = {
  open: '待处理',
  acknowledged: '已确认',
  resolved: '已解决',
  ignored: '已忽略',
  pending_inbound: '待入库',
  idle: '闲置',
  in_use: '在用',
  maintenance: '维修中',
  retired: '已报废'
}

const severityLabels = {
  low: '低风险',
  medium: '中风险',
  high: '高风险',
  critical: '严重风险'
}

const categoryLabels = {
  status: '状态异常',
  value: '价值异常',
  return: '归还与调拨',
  maintenance: '维修异常',
  warranty: '质保风险',
  duplicate: '重复数据'
}

const internalColumns = new Set([
  'ID', 'id', 'CreatedAt', 'createdAt', 'UpdatedAt', 'updatedAt', 'DeletedAt', 'deletedAt',
  'categoryId', 'CategoryID', 'userId', 'UserID', 'fingerprint', 'assetId', 'ruleCode', 'ruleVersion',
  'evidence', 'assignedTo', 'assignedToName', 'handledBy', 'handledByName', 'lastScanRunId',
  'needsClarification'
])

const columnLabels = {
  assetCode: '资产编号',
  assetName: '资产名称',
  name: '资产名称',
  brand: '品牌',
  model: '型号',
  serialNumber: '序列号',
  status: '状态',
  severity: '风险等级',
  category: '风险类别',
  custodian: '保管人',
  title: '风险问题',
  description: '问题说明',
  recommendation: '处理建议',
  warrantyEndDate: '质保到期日',
  lastDetectedAt: '最近发现时间',
  firstDetectedAt: '首次发现时间',
  handledAt: '处理时间',
  purchaseDate: '购置日期',
  productionDate: '生产日期',
  quantity: '数量',
  currentValue: '当前估值',
  originalValue: '资产原值',
  unitPrice: '采购单价'
}

const currencyColumns = new Set(['currentValue', 'originalValue', 'unitPrice'])
const dateColumns = new Set(['lastDetectedAt', 'firstDetectedAt', 'handledAt', 'warrantyEndDate', 'purchaseDate', 'productionDate'])

export function isClarificationMessage(item) {
  return item?.intent === 'clarification' || item?.Intent === 'clarification' || item?.clarification === true || item?.data?.needsClarification === true || item?.Data?.needsClarification === true
}

export function messageStatus(item) {
  if (isClarificationMessage(item)) return { label: '待确认', className: 'message-status--clarification' }
  if (item?.partial) return { label: '部分结果', className: 'message-status--partial' }
  return { label: '已完成', className: 'message-status--complete' }
}

export function shouldHideBusinessColumn(key) {
  return internalColumns.has(key)
}

export function businessColumnLabel(key) {
  return columnLabels[key] || key
}

function formatShanghaiDate(value) {
  if (/^\d{4}-\d{2}-\d{2}$/.test(String(value))) return String(value).replaceAll('-', '/')
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  const parts = new Intl.DateTimeFormat('zh-CN', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric', month: 'numeric', day: 'numeric',
    hour: 'numeric', minute: '2-digit', second: '2-digit', hour12: false
  }).formatToParts(date)
  const values = Object.fromEntries(parts.filter((part) => part.type !== 'literal').map((part) => [part.type, part.value]))
  return `${values.year}/${values.month}/${values.day} ${values.hour}:${values.minute}:${values.second}`
}

export function businessColumnValue(key, value) {
  if (value === undefined || value === null || value === '') return '—'
  if (key === 'status') return statusLabels[value] || String(value)
  if (key === 'severity') return severityLabels[value] || String(value)
  if (key === 'category') return categoryLabels[value] || String(value)
  if (dateColumns.has(key)) return formatShanghaiDate(value)
  if (currencyColumns.has(key) && Number.isFinite(Number(value))) {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY' }).format(Number(value))
  }
  if (typeof value === 'number') return new Intl.NumberFormat('zh-CN').format(value)
  return String(value)
}
