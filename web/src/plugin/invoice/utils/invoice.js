import { formatCurrency, formatDateText } from '@/utils/format'

export const invoiceStatuses = [
  { value: 'uploaded', label: '等待识别', type: 'info' },
  { value: 'recognizing', label: '识别中', type: 'primary' },
  { value: 'pending_review', label: '待核对', type: 'warning' },
  { value: 'confirmed', label: '已确认', type: 'success' },
  { value: 'recognition_failed', label: '识别失败', type: 'danger' }
]

export const invoiceStatusMeta = (value) =>
  invoiceStatuses.find((item) => item.value === value) || { value, label: value || '未知', type: 'info' }

export const invoiceVerificationStatuses = [
  { value: 'unverified', label: '未查验', type: 'info' },
  { value: 'verifying', label: '查验中', type: 'primary' },
  { value: 'verified_valid', label: '查验一致', type: 'success' },
  { value: 'verified_voided', label: '已作废', type: 'danger' },
  { value: 'verified_red', label: '已红冲', type: 'danger' },
  { value: 'inconsistent', label: '信息不一致', type: 'danger' },
  { value: 'not_found', label: '税局无此票', type: 'danger' },
  { value: 'deferred', label: '次日可查', type: 'warning' },
  { value: 'expired', label: '超过查验期', type: 'warning' },
  { value: 'unavailable', label: '暂不可查', type: 'warning' }
]

export const invoiceVerificationStatusMeta = (value) =>
  invoiceVerificationStatuses.find((item) => item.value === value) || { value, label: value || '未查验', type: 'info' }

export const centsToCurrency = (value) => formatCurrency(Number(value || 0) / 100)
export const centsToYuan = (value) => Number((Number(value || 0) / 100).toFixed(2))
export const yuanToCents = (value) => Math.round(Number(value || 0) * 100)
export const invoiceDateText = formatDateText

export const invoiceFileUrl = (id) => {
  const baseURL = String(import.meta.env.VITE_BASE_API || '').replace(/\/$/, '')
  return `${baseURL}/invoice/file?id=${encodeURIComponent(id)}`
}
