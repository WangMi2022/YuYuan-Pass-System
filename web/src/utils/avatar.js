const SAFE_BROWSER_URL = /^(?:https?:|\/\/|data:|blob:)/i

const firstNonEmpty = (...values) => {
  for (const value of values) {
    const normalized = String(value ?? '').trim()
    if (normalized) {
      return normalized
    }
  }
  return ''
}

export const normalizeAvatarUrl = (value, fileBaseUrl = '/') => {
  const url = String(value ?? '').trim()
  if (!url || SAFE_BROWSER_URL.test(url)) {
    return url
  }

  const base = String(fileBaseUrl ?? '').trim().replace(/\/+$/, '')
  const relative = url.replace(/^\/+/, '')
  if (!base || base === '') {
    return `/${relative}`
  }
  return `${base}/${relative}`
}

export const resolveAvatarUrl = ({
  picSrc = '',
  headerImgPreviewUrl = '',
  headerImg = '',
  fileBaseUrl = '/'
} = {}) => normalizeAvatarUrl(
  firstNonEmpty(picSrc, headerImgPreviewUrl, headerImg),
  fileBaseUrl
)
