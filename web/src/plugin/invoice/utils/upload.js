export const MAX_INVOICE_UPLOAD_BYTES = 10 * 1024 * 1024

const invoiceUploadTypes = new Set(['image/jpeg', 'image/png', 'application/pdf'])

export const invoiceUploadError = (file = {}) => {
  if (!invoiceUploadTypes.has(String(file.type || '').toLowerCase())) {
    return '仅支持 JPG、PNG 或 PDF 发票文件'
  }
  const size = Number(file.size || 0)
  if (size <= 0 || size > MAX_INVOICE_UPLOAD_BYTES) {
    return '发票文件大小必须在 10MB 以内'
  }
  return ''
}
