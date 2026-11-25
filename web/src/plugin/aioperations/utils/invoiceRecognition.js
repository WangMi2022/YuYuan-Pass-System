const providerKeys = ['baidu', 'public-ocr', 'verification', 'multimodal']

export const defaultInvoiceRecognition = () => ({
  'fallback-threshold': 0.82,
  'allow-private-endpoints': false,
  baidu: { enabled: false, 'api-key': '', 'secret-key': '', 'api-key-configured': false, 'secret-key-configured': false, 'clear-api-key': false, 'clear-secret-key': false, 'timeout-seconds': 30 },
  'public-ocr': { enabled: false, provider: '', protocol: '', endpoint: '', 'api-key': '', 'api-key-configured': false, 'clear-api-key': false, 'timeout-seconds': 30 },
  verification: { enabled: false, provider: '', protocol: '', endpoint: '', 'api-key': '', 'secret-key': '', 'api-key-configured': false, 'secret-key-configured': false, 'clear-api-key': false, 'clear-secret-key': false, 'timeout-seconds': 30 },
  multimodal: { enabled: false, 'base-url': '', 'api-key': '', 'api-key-configured': false, 'clear-api-key': false, model: '', protocol: '', 'timeout-seconds': 45 }
})

export const invoiceRecognitionFormValue = (value = {}) => {
  const defaults = defaultInvoiceRecognition()
  const result = { ...defaults, ...value }
  for (const key of providerKeys) result[key] = { ...defaults[key], ...(value[key] || {}) }
  return result
}

export const invoiceRecognitionPayload = (value = {}) => JSON.parse(JSON.stringify(value))
