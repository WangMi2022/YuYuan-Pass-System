import service from '@/utils/request'

export const uploadInvoice = (file) => {
  const data = new FormData()
  data.append('file', file)
  return service({
    url: '/invoice/upload',
    method: 'post',
    data,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export const uploadInvoices = (files) => {
  const data = new FormData()
  files.forEach((file) => data.append('files', file))
  return service({
    url: '/invoice/upload',
    method: 'post',
    data,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

const withoutEmptyParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null)
)

export const getInvoiceList = (params) => service({
  url: '/invoice/list',
  method: 'get',
  params: withoutEmptyParams(params)
})
export const getInvoiceDetail = (params) => service({ url: '/invoice/detail', method: 'get', params })
export const getInvoiceCapabilities = () => service({ url: '/invoice/capabilities', method: 'get' })
export const downloadInvoiceFile = (params) => service({
  url: '/invoice/file',
  method: 'get',
  params,
  responseType: 'arraybuffer',
  donNotShowLoading: true
})
export const updateInvoice = (data) => service({ url: '/invoice/update', method: 'put', data })
export const confirmInvoice = (data) => service({ url: '/invoice/confirm', method: 'put', data })
export const reopenInvoice = (params) => service({ url: '/invoice/reopen', method: 'put', params })
export const retryInvoice = (params) => service({ url: '/invoice/retry', method: 'put', params })
export const recheckInvoice = (params) => service({ url: '/invoice/recheck', method: 'post', params })
export const verifyInvoice = (params) => service({ url: '/invoice/verify', method: 'post', params })
export const getInvoiceVerificationHistory = (params) => service({ url: '/invoice/verificationHistory', method: 'get', params })
export const deleteInvoice = (params) => service({ url: '/invoice/delete', method: 'delete', params })
export const getInvoiceDashboard = () => service({ url: '/invoice/dashboard', method: 'get' })
export const getInvoiceCategoryOptions = () => service({ url: '/invoice/categoryOptions', method: 'get' })

export const createInvoiceCategory = (data) => service({ url: '/invoiceCategory/create', method: 'post', data })
export const updateInvoiceCategory = (data) => service({ url: '/invoiceCategory/update', method: 'put', data })
export const deleteInvoiceCategory = (params) => service({ url: '/invoiceCategory/delete', method: 'delete', params })
export const getInvoiceCategoryList = (params) => service({ url: '/invoiceCategory/list', method: 'get', params })

export const createInvoiceRule = (data) => service({ url: '/invoiceRule/create', method: 'post', data })
export const updateInvoiceRule = (data) => service({ url: '/invoiceRule/update', method: 'put', data })
export const deleteInvoiceRule = (params) => service({ url: '/invoiceRule/delete', method: 'delete', params })
export const getInvoiceRuleList = (params) => service({ url: '/invoiceRule/list', method: 'get', params })
