/**
 * 本地静态头像与图片缓存管理工具
 * 1. 内存高速缓存 + localStorage 本地持久化缓存
 * 2. 自动预加载并离线缓存头像静态资源，避免重复网络请求或内网/OSS 签名失效
 * 3. 限制单项尺寸与总条数，防止 localStorage 溢出
 */

const STORAGE_KEY = 'mit_avatar_local_cache_v1'
const MAX_CACHE_ITEMS = 60
const memoryCache = new Map()

// 获取本地缓存字典
const getStorageMap = () => {
  if (typeof window === 'undefined' || !window.localStorage) return {}
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch (err) {
    void err
    return {}
  }
}

// 写入本地持久化
const saveStorageMap = (map) => {
  if (typeof window === 'undefined' || !window.localStorage) return
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(map))
  } catch (err) {
    void err
    try {
      const keys = Object.keys(map)
      if (keys.length > 10) {
        const keepKeys = keys.slice(Math.floor(keys.length / 2))
        const cleanMap = {}
        keepKeys.forEach((k) => {
          cleanMap[k] = map[k]
        })
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(cleanMap))
      }
    } catch (cleanupErr) {
      void cleanupErr
    }
  }
}

/**
 * 获取本地已缓存的头像（同步返回，无网络延迟）
 * @param {string} url 原始头像地址
 * @returns {string|null} 本地 Base64 或缓存地址
 */
export const getLocalCachedAvatar = (url) => {
  if (!url || typeof url !== 'string') return null
  const cleanUrl = url.trim()
  if (!cleanUrl) return null

  // 1. 内存缓存
  if (memoryCache.has(cleanUrl)) {
    return memoryCache.get(cleanUrl)
  }

  // 2. 本地持久化缓存
  const storageMap = getStorageMap()
  if (storageMap[cleanUrl]) {
    const cachedData = storageMap[cleanUrl]
    memoryCache.set(cleanUrl, cachedData)
    return cachedData
  }

  return null
}

/**
 * 缓存头像数据到本地
 * @param {string} url 原始头像地址
 * @param {string} dataUrl Base64 数据或 Blob URL
 */
export const setLocalCachedAvatar = (url, dataUrl) => {
  if (!url || !dataUrl || typeof url !== 'string' || typeof dataUrl !== 'string') return
  const cleanUrl = url.trim()
  if (!cleanUrl) return

  memoryCache.set(cleanUrl, dataUrl)

  const storageMap = getStorageMap()
  const keys = Object.keys(storageMap)
  if (keys.length >= MAX_CACHE_ITEMS) {
    delete storageMap[keys[0]]
  }
  storageMap[cleanUrl] = dataUrl
  saveStorageMap(storageMap)
}

/**
 * 异步下载并将头像缓存到本地（离线持久化）
 * @param {string} url 原始头像地址
 * @returns {Promise<string>}
 */
export const fetchAndCacheAvatar = async (url) => {
  if (!url || typeof url !== 'string') return url
  const cleanUrl = url.trim()
  if (!cleanUrl) return cleanUrl

  if (cleanUrl.startsWith('data:') || cleanUrl.startsWith('blob:')) {
    setLocalCachedAvatar(cleanUrl, cleanUrl)
    return cleanUrl
  }

  const cached = getLocalCachedAvatar(cleanUrl)
  if (cached) return cached

  if (typeof window === 'undefined') return cleanUrl

  try {
    return await new Promise((resolve) => {
      const img = new Image()
      img.setAttribute('crossOrigin', 'anonymous')
      img.onload = () => {
        try {
          const canvas = document.createElement('canvas')
          const size = Math.min(img.width || 120, 160)
          canvas.width = size
          canvas.height = size
          const ctx = canvas.getContext('2d')
          ctx.drawImage(img, 0, 0, size, size)
          const base64 = canvas.toDataURL('image/png', 0.88)
          setLocalCachedAvatar(cleanUrl, base64)
          resolve(base64)
        } catch {
          resolve(cleanUrl)
        }
      }
      img.onerror = () => {
        resolve(cleanUrl)
      }
      img.src = cleanUrl
    })
  } catch {
    return cleanUrl
  }
}
