<template>
  <main class="system-settings-page">
    <section class="settings-heading" aria-labelledby="settings-title">
      <div>
        <p class="eyebrow">SYSTEM APPEARANCE</p>
        <h1 id="settings-title">系统配置</h1>
        <p>统一管理系统外观配置，支持设置登录页图标和背景图片。</p>
      </div>
    </section>

    <section class="setting-card" aria-labelledby="login-logo-title">
      <header class="setting-card-header">
        <div>
          <h2 id="login-logo-title">登录页图标</h2>
          <p>图标显示在登录表单顶部和登录页品牌区域，建议上传清晰的正方形图片。</p>
        </div>
      </header>

      <div v-loading="logoLoading" class="current-logo">
        <div class="logo-preview" aria-label="当前登录页图标预览">
          <img
            v-if="currentLogo?.url && !logoPreviewFailed"
            :src="currentLogo.url"
            alt="当前登录页图标"
            @error="logoPreviewFailed = true"
          />
          <Logo v-else :size="2.5" />
        </div>
        <div class="current-info">
          <span class="current-label">当前图标</span>
          <strong>{{ currentLogo?.name || '系统默认图标' }}</strong>
          <p>{{ currentLogo ? '图片已保存至 OSS，并应用到登录页。' : '当前使用系统内置的默认图标。' }}</p>
        </div>
        <div class="logo-actions">
          <el-upload
            :action="`${getBaseUrl()}/fileUploadAndDownload/upload`"
            accept="image/jpeg,image/png,image/webp"
            :headers="{ 'x-token': userStore.token }"
            :show-file-list="false"
            :multiple="false"
            :disabled="logoUploading"
            :before-upload="beforeLogoUpload"
            :on-success="logoUploadSuccess"
            :on-error="logoUploadError"
          >
            <el-button type="primary" :icon="Upload" :loading="logoUploading">上传并应用</el-button>
          </el-upload>
          <el-button
            :icon="RefreshLeft"
            :disabled="!currentLogo || logoUploading"
            @click="restoreDefaultLogo"
          >恢复默认</el-button>
        </div>
      </div>
    </section>

    <section class="setting-card" aria-labelledby="login-background-title">
      <header class="setting-card-header">
        <div>
          <h2 id="login-background-title">登录页背景</h2>
          <p>登录页将使用当前启用的背景图片，并自动叠加遮罩保证表单可读性。</p>
        </div>
        <el-button v-if="!managing" type="primary" :icon="Edit" @click="startManaging">变更管理</el-button>
      </header>

      <div class="current-background">
        <div class="current-preview">
          <img :src="currentBackground?.url || defaultBackground" alt="当前登录页背景缩略图" @error="imageFallback" />
          <span class="current-badge">当前使用</span>
        </div>
        <div class="current-info">
          <span class="current-label">当前背景</span>
          <strong>{{ currentBackground?.name || '系统默认背景' }}</strong>
          <p>{{ currentBackground ? '图片已保存至 OSS，并在登录页实时生效。' : '当前使用系统内置的默认登录背景。' }}</p>
        </div>
      </div>

      <el-collapse-transition>
        <div v-if="managing" class="background-manager">
          <div class="manager-toolbar">
            <div>
              <h3>背景图库</h3>
              <p>上传图片后会生成缩略图。选中目标图片并保存，才会切换登录页背景。</p>
            </div>
            <el-upload
              :action="`${getBaseUrl()}/fileUploadAndDownload/upload`"
              accept="image/jpeg,image/png,image/webp"
              :headers="{ 'x-token': userStore.token }"
              :show-file-list="false"
              :multiple="false"
              :disabled="uploading"
              :before-upload="beforeUpload"
              :on-success="uploadSuccess"
              :on-error="uploadError"
            >
              <el-button type="primary" plain :icon="Upload" :loading="uploading">上传背景图片</el-button>
            </el-upload>
          </div>

          <div v-loading="loading" class="background-grid">
            <button
              v-for="item in backgrounds"
              :key="item.ID"
              type="button"
              class="background-option"
              :class="{ selected: selectedId === item.ID, active: item.isActive }"
              :aria-pressed="selectedId === item.ID"
              @click="selectedId = item.ID"
            >
              <span class="thumbnail-wrap">
                <img :src="item.url" :alt="`${item.name}缩略图`" loading="lazy" @error="imageFallback" />
                <span v-if="selectedId === item.ID" class="selected-mark"><el-icon><Check /></el-icon></span>
              </span>
              <span class="option-info">
                <strong :title="item.name">{{ item.name }}</strong>
                <span>{{ item.isActive ? '正在使用' : '点击选择' }}</span>
              </span>
              <el-button
                v-if="!item.isActive"
                class="delete-background"
                type="danger"
                text
                :icon="Delete"
                aria-label="删除背景图片"
                @click.stop="removeBackground(item)"
              />
            </button>

            <div v-if="!loading && !backgrounds.length" class="empty-gallery">
              <el-icon><Picture /></el-icon>
              <strong>图库暂无图片</strong>
              <span>请先上传一张 JPG、PNG 或 WebP 图片</span>
            </div>
          </div>

          <footer class="manager-actions">
            <span>{{ selectedBackground ? `已选择：${selectedBackground.name}` : '请选择要启用的背景图片' }}</span>
            <div>
              <el-button @click="cancelManaging">取消</el-button>
              <el-button type="primary" :loading="saving" :disabled="!selectedId || selectedId === currentBackground?.ID" @click="saveBackground">保存并应用</el-button>
            </div>
          </footer>
        </div>
      </el-collapse-transition>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { Check, Delete, Edit, Picture, RefreshLeft, Upload } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getBaseUrl } from '@/utils/format'
import { useUserStore } from '@/pinia/modules/user'
import defaultBackground from '@/assets/login_background.jpg'
import Logo from '@/components/logo/index.vue'
import {
  activateLoginBackground,
  createLoginBackground,
  deleteLoginBackground,
  getCurrentLoginLogo,
  getLoginBackgrounds,
  resetLoginLogo,
  saveLoginLogo
} from '@/api/systemSettings'

defineOptions({ name: 'SystemSettings' })

const userStore = useUserStore()
const backgrounds = ref([])
const selectedId = ref(0)
const managing = ref(false)
const loading = ref(false)
const uploading = ref(false)
const saving = ref(false)
const currentLogo = ref(null)
const logoLoading = ref(false)
const logoUploading = ref(false)
const logoPreviewFailed = ref(false)

const currentBackground = computed(() => backgrounds.value.find((item) => item.isActive))
const selectedBackground = computed(() => backgrounds.value.find((item) => item.ID === selectedId.value))

const loadLoginLogo = async () => {
  logoLoading.value = true
  try {
    const res = await getCurrentLoginLogo()
    if (res.code === 0) {
      currentLogo.value = res.data?.url ? res.data : null
      logoPreviewFailed.value = false
    }
  } finally {
    logoLoading.value = false
  }
}

const beforeLogoUpload = (file) => {
  const allowed = ['image/jpeg', 'image/png', 'image/webp'].includes(file.type?.toLowerCase())
  if (!allowed) {
    ElMessage.error('登录图标仅支持 JPG、PNG、WebP 图片')
    return false
  }
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.error('登录图标不能超过 2MB')
    return false
  }
  logoUploading.value = true
  return true
}

const logoUploadSuccess = async (response, uploadFile) => {
  try {
    if (response?.code !== 0 || !response?.data?.file?.url) {
      ElMessage.error(response?.msg || '登录图标上传失败')
      return
    }
    const file = response.data.file
    const res = await saveLoginLogo({
      name: uploadFile?.name || file.name || '登录图标',
      url: file.url
    })
    if (res.code === 0) {
      await loadLoginLogo()
      ElMessage.success('登录图标已更新')
    }
  } finally {
    logoUploading.value = false
  }
}

const logoUploadError = () => {
  logoUploading.value = false
  ElMessage.error('登录图标上传失败')
}

const restoreDefaultLogo = async () => {
  if (!currentLogo.value) return
  try {
    await ElMessageBox.confirm('确定恢复系统默认登录图标吗？', '恢复默认图标', {
      type: 'warning',
      confirmButtonText: '恢复默认',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  const res = await resetLoginLogo()
  if (res.code === 0) {
    await loadLoginLogo()
    ElMessage.success('已恢复默认登录图标')
  }
}

const loadBackgrounds = async () => {
  loading.value = true
  try {
    const res = await getLoginBackgrounds()
    if (res.code === 0) backgrounds.value = res.data || []
  } finally {
    loading.value = false
  }
}

const startManaging = () => {
  selectedId.value = currentBackground.value?.ID || 0
  managing.value = true
}

const cancelManaging = () => {
  selectedId.value = currentBackground.value?.ID || 0
  managing.value = false
}

const beforeUpload = (file) => {
  const allowed = ['image/jpeg', 'image/png', 'image/webp'].includes(file.type?.toLowerCase())
  if (!allowed) {
    ElMessage.error('仅支持 JPG、PNG、WebP 背景图片')
    return false
  }
  if (file.size > 10 * 1024 * 1024) {
    ElMessage.error('背景图片不能超过 10MB')
    return false
  }
  uploading.value = true
  return true
}

const uploadSuccess = async (response, uploadFile) => {
  try {
    if (response?.code !== 0 || !response?.data?.file?.url) {
      ElMessage.error(response?.msg || '背景图片上传失败')
      return
    }
    const file = response.data.file
    const res = await createLoginBackground({
      name: uploadFile?.name || file.name || '登录背景',
      url: file.url
    })
    if (res.code === 0) {
      await loadBackgrounds()
      selectedId.value = res.data.ID
      ElMessage.success('图片已上传，请确认选择后保存')
    }
  } finally {
    uploading.value = false
  }
}

const uploadError = () => {
  uploading.value = false
  ElMessage.error('背景图片上传失败')
}

const saveBackground = async () => {
  if (!selectedId.value) return
  saving.value = true
  try {
    const res = await activateLoginBackground({ id: selectedId.value })
    if (res.code === 0) {
      await loadBackgrounds()
      managing.value = false
      ElMessage.success('登录页背景已切换')
    }
  } finally {
    saving.value = false
  }
}

const removeBackground = async (item) => {
  try {
    await ElMessageBox.confirm(`确定从背景图库删除“${item.name}”吗？`, '删除背景图片', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  const res = await deleteLoginBackground({ id: item.ID })
  if (res.code === 0) {
    if (selectedId.value === item.ID) selectedId.value = currentBackground.value?.ID || 0
    await loadBackgrounds()
    ElMessage.success('背景图片已删除')
  }
}

const imageFallback = (event) => {
  if (!event.target.dataset.fallback) {
    event.target.dataset.fallback = 'true'
    event.target.src = defaultBackground
  }
}

onMounted(() => {
  loadLoginLogo()
  loadBackgrounds()
})
</script>

<style scoped lang="scss">
.system-settings-page { min-height: 100%; overflow-x: hidden; padding: 18px; background: var(--na-background, #f6f8fb); color: var(--el-text-color-primary); }
.settings-heading,
.setting-card { border: 1px solid var(--na-border, var(--el-border-color-light)); border-radius: 14px; background: var(--na-card, var(--el-bg-color)); box-shadow: 0 8px 24px rgb(15 23 42 / 4%); }
.settings-heading { margin-bottom: 14px; padding: 18px 20px; }
.setting-card + .setting-card { margin-top: 14px; }
.eyebrow { margin: 0 0 5px; color: var(--el-color-primary); font-size: 11px; font-weight: 750; letter-spacing: .12em; }
.settings-heading h1 { margin: 0; font-size: 19px; }
.settings-heading p:last-child { margin: 5px 0 0; color: var(--el-text-color-secondary); font-size: 13px; }
.setting-card { overflow: hidden; }
.setting-card-header { display: flex; align-items: center; justify-content: space-between; gap: 20px; padding: 18px 20px; border-bottom: 1px solid var(--el-border-color-lighter); }
.setting-card-header h2 { margin: 0; font-size: 16px; }
.setting-card-header p { margin: 5px 0 0; color: var(--el-text-color-secondary); font-size: 12px; }
.current-background { display: grid; grid-template-columns: minmax(260px, 420px) minmax(220px, 1fr); align-items: center; gap: 24px; padding: 20px; }
.current-preview { position: relative; overflow: hidden; aspect-ratio: 16 / 7; border: 1px solid var(--el-border-color-lighter); border-radius: 10px; background: var(--el-fill-color-light); }
.current-preview img { width: 100%; height: 100%; object-fit: cover; }
.current-badge { position: absolute; top: 10px; left: 10px; border-radius: 999px; background: rgb(15 23 42 / 78%); padding: 4px 9px; color: #fff; font-size: 11px; font-weight: 650; backdrop-filter: blur(4px); }
.current-info { min-width: 0; }
.current-label { display: block; margin-bottom: 6px; color: var(--el-text-color-secondary); font-size: 12px; }
.current-info strong { display: block; overflow: hidden; font-size: 17px; text-overflow: ellipsis; white-space: nowrap; }
.current-info p { margin: 8px 0 0; color: var(--el-text-color-secondary); font-size: 12px; line-height: 1.6; }
.current-logo { display: grid; grid-template-columns: 80px minmax(200px, 1fr) auto; align-items: center; gap: 20px; min-height: 120px; padding: 20px; }
.logo-preview { display: grid; place-items: center; width: 72px; height: 72px; overflow: hidden; border: 1px solid var(--el-border-color-lighter); border-radius: 14px; background: var(--el-fill-color-extra-light); }
.logo-preview img { width: 48px; height: 48px; object-fit: contain; }
.logo-actions { display: flex; align-items: center; gap: 8px; }
.background-manager { border-top: 1px solid var(--el-border-color-lighter); background: var(--el-fill-color-extra-light); padding: 18px 20px 20px; }
.manager-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 20px; margin-bottom: 16px; }
.manager-toolbar h3 { margin: 0; font-size: 15px; }
.manager-toolbar p { margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 12px; }
.background-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 12px; min-height: 144px; }
.background-option { position: relative; min-width: 0; overflow: hidden; padding: 0; border: 1px solid var(--el-border-color); border-radius: 10px; background: var(--el-bg-color); color: inherit; text-align: left; cursor: pointer; transition: border-color 180ms ease, box-shadow 180ms ease; }
.background-option:hover { border-color: var(--el-color-primary-light-5); }
.background-option:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 2px; }
.background-option.selected { border-color: var(--el-color-primary); box-shadow: 0 0 0 2px var(--el-color-primary-light-9); }
.thumbnail-wrap { position: relative; display: block; aspect-ratio: 16 / 8; overflow: hidden; background: var(--el-fill-color-light); }
.thumbnail-wrap img { width: 100%; height: 100%; object-fit: cover; }
.selected-mark { position: absolute; right: 9px; bottom: 9px; display: grid; place-items: center; width: 24px; height: 24px; border-radius: 50%; background: var(--el-color-primary); color: #fff; box-shadow: 0 2px 8px rgb(0 0 0 / 18%); }
.option-info { display: flex; min-width: 0; flex-direction: column; gap: 3px; padding: 10px 42px 11px 12px; }
.option-info strong { overflow: hidden; font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
.option-info span { color: var(--el-text-color-secondary); font-size: 11px; }
.delete-background { position: absolute; right: 7px; bottom: 6px; width: 32px; height: 32px; padding: 0; }
.empty-gallery { grid-column: 1 / -1; display: flex; min-height: 144px; align-items: center; justify-content: center; flex-direction: column; gap: 6px; border: 1px dashed var(--el-border-color); border-radius: 10px; color: var(--el-text-color-secondary); }
.empty-gallery .el-icon { font-size: 30px; }
.empty-gallery strong { color: var(--el-text-color-primary); font-size: 13px; }
.empty-gallery span { font-size: 12px; }
.manager-actions { display: flex; align-items: center; justify-content: space-between; gap: 18px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--el-border-color-lighter); }
.manager-actions > span { overflow: hidden; color: var(--el-text-color-secondary); font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.manager-actions > div { display: flex; flex: 0 0 auto; gap: 8px; }
@media (max-width: 720px) {
  .system-settings-page { padding: 12px; }
  .setting-card-header,
  .manager-toolbar,
  .manager-actions { align-items: stretch; flex-direction: column; }
  .current-background { grid-template-columns: 1fr; gap: 14px; padding: 16px; }
  .current-logo { grid-template-columns: 72px minmax(0, 1fr); gap: 14px; padding: 16px; }
  .logo-actions { grid-column: 1 / -1; flex-wrap: wrap; }
  .background-manager { padding: 16px; }
  .background-grid { grid-template-columns: 1fr; }
  .manager-actions > div { justify-content: flex-end; }
}
@media (prefers-reduced-motion: reduce) { .background-option { transition: none; } }
</style>
