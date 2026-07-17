<template>
  <div class="settings-module general-settings">
    <header class="settings-module__header">
      <p class="settings-module__eyebrow">GENERAL</p>
      <h3>通用设置</h3>
      <p>查看运行环境，备份配置或恢复系统默认设置。</p>
    </header>

    <section class="settings-section" aria-labelledby="system-info-title">
      <div class="settings-section__header">
        <h4 id="system-info-title">系统信息</h4>
        <p>当前前端运行环境与基础技术信息。</p>
      </div>
      <dl class="system-info-grid">
        <div v-for="item in systemInfo" :key="item.label">
          <dt>{{ item.label }}</dt>
          <dd>{{ item.value }}</dd>
        </div>
      </dl>
    </section>

    <section class="settings-section" aria-labelledby="config-management-title">
      <div class="settings-section__header">
        <h4 id="config-management-title">配置管理</h4>
        <p>配置文件仅包含界面偏好，不包含业务数据和登录凭据。</p>
      </div>
      <div class="settings-action-list">
        <div class="settings-action-row">
          <span class="settings-action-row__icon" aria-hidden="true"><el-icon><Download /></el-icon></span>
          <div>
            <strong>导出配置</strong>
            <p>下载当前配置的 JSON 备份文件</p>
          </div>
          <el-button :icon="Download" @click="handleExportConfig">导出</el-button>
        </div>

        <div class="settings-action-row">
          <span class="settings-action-row__icon" aria-hidden="true"><el-icon><Upload /></el-icon></span>
          <div>
            <strong>导入配置</strong>
            <p>从 JSON 文件恢复界面偏好</p>
          </div>
          <el-upload ref="uploadRef" :auto-upload="false" :show-file-list="false" accept="application/json,.json" @change="handleImportConfig">
            <el-button :icon="Upload">导入</el-button>
          </el-upload>
        </div>

        <div class="settings-action-row is-danger">
          <span class="settings-action-row__icon" aria-hidden="true"><el-icon><RefreshLeft /></el-icon></span>
          <div>
            <strong>恢复默认配置</strong>
            <p>清除当前外观与布局偏好</p>
          </div>
          <el-button type="danger" plain :icon="RefreshLeft" @click="emit('reset')">恢复</el-button>
        </div>
      </div>
    </section>

    <section class="settings-section" aria-labelledby="about-title">
      <div class="settings-section__header">
        <h4 id="about-title">关于系统</h4>
      </div>
      <div class="about-system">
        <span class="about-system__logo"><Logo /></span>
        <div>
          <strong>资产管理中心</strong>
          <p>面向资产全生命周期管理的业务平台，覆盖档案、入库、领用、调拨、归还、维修、报废与盘点流程。</p>
          <span>当前版本 v2.9.2</span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Download, RefreshLeft, Upload } from '@element-plus/icons-vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/pinia'
import Logo from '@/components/logo/index.vue'

defineOptions({
  name: 'GeneralSettings'
})

const emit = defineEmits(['reset'])
const appStore = useAppStore()
const { config } = storeToRefs(appStore)
const uploadRef = ref()
const browserInfo = ref('检测中')
const screenResolution = ref('检测中')

const systemInfo = computed(() => [
  { label: '系统版本', value: 'v2.9.2' },
  { label: '前端框架', value: 'Vue 3' },
  { label: '组件库', value: 'Element Plus' },
  { label: '构建工具', value: 'Vite' },
  { label: '浏览器', value: browserInfo.value },
  { label: '屏幕分辨率', value: screenResolution.value }
])

onMounted(() => {
  const userAgent = navigator.userAgent
  if (/Edg\//.test(userAgent)) browserInfo.value = 'Microsoft Edge'
  else if (/Chrome\//.test(userAgent)) browserInfo.value = 'Chrome'
  else if (/Firefox\//.test(userAgent)) browserInfo.value = 'Firefox'
  else if (/Safari\//.test(userAgent)) browserInfo.value = 'Safari'
  else browserInfo.value = '其他浏览器'

  screenResolution.value = `${screen.width} x ${screen.height}`
})

const handleExportConfig = () => {
  const configData = JSON.stringify(config.value, null, 2)
  const blob = new Blob([configData], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')

  link.href = url
  link.download = `asset-center-config-${new Date().toISOString().split('T')[0]}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  ElMessage.success('配置已导出')
}

const handleImportConfig = (file) => {
  const reader = new FileReader()
  reader.onload = (event) => {
    try {
      const importedConfig = JSON.parse(event.target.result)
      if (!importedConfig || Array.isArray(importedConfig) || typeof importedConfig !== 'object') {
        throw new Error('配置内容无效')
      }

      Object.keys(importedConfig).forEach((key) => {
        if (key in config.value) config.value[key] = importedConfig[key]
      })
      ElMessage.success('配置已导入')
    } catch {
      ElMessage.error('配置文件格式错误')
    } finally {
      uploadRef.value?.clearFiles()
    }
  }
  reader.readAsText(file.raw)
}
</script>

<style scoped lang="scss">
.system-info-grid {
  display: grid;
  overflow: hidden;
  margin: 0;
  grid-template-columns: repeat(3, 1fr);
  border: 1px solid var(--na-border);
  border-radius: 8px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
}

.system-info-grid > div { min-width: 0; padding: 14px; border-right: 1px solid var(--na-border); border-bottom: 1px solid var(--na-border); }
.system-info-grid > div:nth-child(3n) { border-right: 0; }
.system-info-grid > div:nth-last-child(-n + 3) { border-bottom: 0; }
.system-info-grid dt { color: var(--na-muted-foreground); font-size: 9px; }
.system-info-grid dd { overflow: hidden; margin: 4px 0 0; color: var(--na-foreground); font-size: 11px; font-weight: 620; text-overflow: ellipsis; white-space: nowrap; }

.settings-action-list { overflow: hidden; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-card); box-shadow: var(--na-shadow-sm); }
.settings-action-row { display: grid; min-height: 76px; grid-template-columns: 36px minmax(0, 1fr) auto; align-items: center; gap: 12px; padding: 12px 14px; border-bottom: 1px solid var(--na-border); }
.settings-action-row:last-child { border-bottom: 0; }
.settings-action-row__icon { display: grid; width: 36px; height: 36px; place-items: center; border-radius: 7px; background: var(--na-primary-soft); color: var(--na-primary); font-size: 16px; }
.settings-action-row > div { min-width: 0; }
.settings-action-row strong { display: block; color: var(--na-foreground); font-size: 12px; font-weight: 650; }
.settings-action-row p { margin: 3px 0 0; color: var(--na-muted-foreground); font-size: 10px; }
.settings-action-row.is-danger .settings-action-row__icon { background: color-mix(in srgb, var(--na-danger) 10%, var(--na-card)); color: var(--na-danger); }

.about-system { display: flex; align-items: flex-start; gap: 14px; padding: 2px 0; }
.about-system__logo { display: grid; width: 46px; height: 46px; place-items: center; flex: 0 0 46px; border: 1px solid var(--na-border); border-radius: 8px; background: var(--na-card); }
.about-system__logo :deep(img), .about-system__logo :deep(svg) { max-width: 30px; max-height: 30px; }
.about-system > div { min-width: 0; }
.about-system strong { color: var(--na-foreground); font-size: 13px; font-weight: 680; }
.about-system p { max-width: 520px; margin: 5px 0; color: var(--na-muted-foreground); font-size: 10px; line-height: 1.7; }
.about-system > div > span { color: var(--na-primary); font-size: 10px; font-weight: 600; }

@media (max-width: 520px) {
  .system-info-grid { grid-template-columns: repeat(2, 1fr); }
  .system-info-grid > div:nth-child(3n) { border-right: 1px solid var(--na-border); }
  .system-info-grid > div:nth-child(2n) { border-right: 0; }
  .system-info-grid > div:nth-last-child(-n + 3) { border-bottom: 1px solid var(--na-border); }
  .system-info-grid > div:nth-last-child(-n + 2) { border-bottom: 0; }
  .settings-action-row { grid-template-columns: 32px minmax(0, 1fr) auto; gap: 9px; padding: 11px; }
  .settings-action-row__icon { width: 32px; height: 32px; }
  .settings-action-row p { display: none; }
}
</style>
