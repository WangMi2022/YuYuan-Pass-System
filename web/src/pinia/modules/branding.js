import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import config from '@/core/config'
import { getCurrentLoginLogo } from '@/api/systemSettings'

const DEFAULT_SYSTEM_NAME = 'mit-assets-admin'
const DEFAULT_SUBTITLE = 'ASSET CONTROL'
const DEFAULT_LOGO = '/asset-logo.svg'

export const useBrandingStore = defineStore('branding', () => {
  const systemName = ref(DEFAULT_SYSTEM_NAME)
  const subtitle = ref(DEFAULT_SUBTITLE)
  const customLogoUrl = ref('')
  const logoName = ref('')
  const loaded = ref(false)
  let loadingPromise = null

  const logoUrl = computed(() => customLogoUrl.value || DEFAULT_LOGO)

  const applyToDocument = (previousName = config.appName) => {
    config.appName = systemName.value
    if (typeof document === 'undefined') return

    const currentTitle = document.title || ''
    if (!currentTitle || currentTitle === previousName || currentTitle === '资产管理中心') {
      document.title = systemName.value
    } else if (previousName && currentTitle.endsWith(` - ${previousName}`)) {
      document.title = `${currentTitle.slice(0, -previousName.length)}${systemName.value}`
    }

    let favicon = document.querySelector('link[rel~="icon"]')
    if (!favicon) {
      favicon = document.createElement('link')
      favicon.rel = 'icon'
      document.head.appendChild(favicon)
    }
    favicon.href = logoUrl.value

    const loadingLogo = document.querySelector('#loading img')
    if (loadingLogo) {
      loadingLogo.src = logoUrl.value
      loadingLogo.alt = systemName.value
    }
  }

  const setBranding = (data = {}) => {
    const previousName = config.appName
    systemName.value = String(data.systemName || DEFAULT_SYSTEM_NAME).trim() || DEFAULT_SYSTEM_NAME
    subtitle.value = data.subtitle == null ? DEFAULT_SUBTITLE : String(data.subtitle).trim()
    customLogoUrl.value = String(data.url || '').trim()
    logoName.value = String(data.name || '').trim()
    loaded.value = true
    applyToDocument(previousName)
  }

  const loadBranding = async (force = false) => {
    if (loaded.value && !force) return
    if (loadingPromise && !force) return loadingPromise
    loadingPromise = getCurrentLoginLogo()
      .then((res) => {
        if (res.code === 0) setBranding(res.data || {})
      })
      .catch(() => {
        if (!loaded.value) setBranding()
      })
      .finally(() => {
        loadingPromise = null
      })
    return loadingPromise
  }

  const useDefaultLogo = () => {
    customLogoUrl.value = ''
    logoName.value = ''
    applyToDocument(config.appName)
  }

  return {
    systemName,
    subtitle,
    logoUrl,
    customLogoUrl,
    logoName,
    loaded,
    loadBranding,
    setBranding,
    useDefaultLogo
  }
})
