<template>
  <main class="na-page na-page--list system-settings-page">
    <section class="na-page-header settings-heading" aria-labelledby="settings-title">
      <div>
        <p class="eyebrow">平台标识与登录体验</p>
        <h1 id="settings-title">品牌外观</h1>
        <p>统一管理系统名称、Logo 和登录背景，保存后将在所有终端生效。</p>
      </div>
    </section>

    <section class="na-panel setting-card" aria-labelledby="brand-identity-title">
      <header class="na-panel-header setting-card-header">
        <div>
          <h2 id="brand-identity-title">系统标识</h2>
          <p>系统名称用于侧边栏、登录页和浏览器标题；副标题可留空。</p>
        </div>
      </header>

      <div class="brand-editor">
        <el-form
          ref="brandFormRef"
          :model="brandForm"
          :rules="brandRules"
          label-position="top"
          class="brand-fields"
        >
          <el-form-item label="系统名称" prop="systemName">
            <el-input v-model="brandForm.systemName" maxlength="80" show-word-limit placeholder="例如：MIT 资产管理平台" />
          </el-form-item>
          <el-form-item label="品牌副标题" prop="subtitle">
            <el-input v-model="brandForm.subtitle" maxlength="120" show-word-limit placeholder="可选，例如：ASSET CONTROL" />
          </el-form-item>
          <div class="brand-actions">
            <el-button type="primary" :loading="brandingSaving" @click="saveBrandIdentity">保存并应用</el-button>
            <el-button :disabled="brandingSaving" @click="restoreDefaultBrand">恢复默认名称</el-button>
          </div>
        </el-form>

        <aside class="brand-preview" aria-label="系统品牌预览">
          <Logo :size="3" />
          <span>
            <strong>{{ brandForm.systemName || '系统名称' }}</strong>
            <small v-if="brandForm.subtitle">{{ brandForm.subtitle }}</small>
          </span>
        </aside>
      </div>
    </section>

    <section class="na-panel setting-card" aria-labelledby="login-logo-title">
      <header class="na-panel-header setting-card-header">
        <div>
          <h2 id="login-logo-title">系统 Logo</h2>
          <p>Logo 将同步用于系统顶栏、登录页和浏览器图标，建议上传清晰的正方形图片。</p>
        </div>
      </header>

      <div v-loading="logoLoading" class="current-logo">
        <div class="logo-preview" aria-label="当前系统Logo预览">
          <img
            v-if="currentLogo?.url && !logoPreviewFailed"
            :src="currentLogo.url"
            alt="当前系统Logo"
            @error="logoPreviewFailed = true"
          />
          <Logo v-else :size="2.5" />
        </div>
        <div class="current-info">
          <span class="current-label">当前 Logo</span>
          <strong>{{ currentLogo?.name || '系统默认 Logo' }}</strong>
          <p>{{ currentLogo ? '图片已保存至 OSS，并应用到所有品牌位置。' : '当前使用系统内置的默认 Logo。' }}</p>
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
          >恢复默认 Logo</el-button>
        </div>
      </div>
    </section>

    <section class="na-panel setting-card" aria-labelledby="login-background-title">
      <header class="na-panel-header setting-card-header">
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
import { computed, onMounted, reactive, ref } from 'vue'
import { Check, Delete, Edit, Picture, RefreshLeft, Upload } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getBaseUrl } from '@/utils/format'
import { useUserStore } from '@/pinia/modules/user'
import { useBrandingStore } from '@/pinia'
import defaultBackground from '@/assets/login_background.jpg'
import Logo from '@/components/logo/index.vue'
import {
  activateLoginBackground,
  createLoginBackground,
  deleteLoginBackground,
  getCurrentLoginLogo,
  getLoginBackgrounds,
  resetLoginLogo,
  saveBranding,
  saveLoginLogo
} from '@/api/systemSettings'

defineOptions({ name: 'SystemSettings' })

const userStore = useUserStore()
const brandingStore = useBrandingStore()
const brandFormRef = ref(null)
const brandForm = reactive({ systemName: 'mit-assets-admin', subtitle: 'ASSET CONTROL' })
const brandRules = {
  systemName: [{ required: true, message: '请输入系统名称', trigger: 'blur' }]
}
const brandingSaving = ref(false)
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
      brandForm.systemName = res.data?.systemName || 'mit-assets-admin'
      brandForm.subtitle = res.data?.subtitle ?? 'ASSET CONTROL'
      brandingStore.setBranding(res.data || {})
      logoPreviewFailed.value = false
    }
  } finally {
    logoLoading.value = false
  }
}

const saveBrandIdentity = async () => {
  const valid = await brandFormRef.value?.validate().catch(() => false)
  if (!valid) return
  brandingSaving.value = true
  try {
    const res = await saveBranding({
      systemName: brandForm.systemName.trim(),
      subtitle: brandForm.subtitle.trim()
    })
    if (res.code === 0) {
      await loadLoginLogo()
      ElMessage.success('系统品牌已更新')
    }
  } finally {
    brandingSaving.value = false
  }
}

const restoreDefaultBrand = async () => {
  brandForm.systemName = 'mit-assets-admin'
  brandForm.subtitle = 'ASSET CONTROL'
  await saveBrandIdentity()
}

const beforeLogoUpload = (file) => {
  const allowed = ['image/jpeg', 'image/png', 'image/webp'].includes(file.type?.toLowerCase())
  if (!allowed) {
    ElMessage.error('系统 Logo 仅支持 JPG、PNG、WebP 图片')
    return false
  }
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.error('系统 Logo 不能超过 2MB')
    return false
  }
  logoUploading.value = true
  return true
}

const logoUploadSuccess = async (response, uploadFile) => {
  try {
    if (response?.code !== 0 || !response?.data?.file?.url) {
      ElMessage.error(response?.msg || '系统 Logo 上传失败')
      return
    }
    const file = response.data.file
    const res = await saveLoginLogo({
      name: uploadFile?.name || file.name || '登录图标',
      url: file.url
    })
    if (res.code === 0) {
      await loadLoginLogo()
      ElMessage.success('系统 Logo 已更新')
    }
  } finally {
    logoUploading.value = false
  }
}

const logoUploadError = () => {
  logoUploading.value = false
  ElMessage.error('系统 Logo 上传失败')
}

const restoreDefaultLogo = async () => {
  if (!currentLogo.value) return
  try {
    await ElMessageBox.confirm('确定恢复系统默认 Logo 吗？系统名称不会改变。', '恢复默认 Logo', {
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
    brandingStore.useDefaultLogo()
    ElMessage.success('已恢复默认系统 Logo')
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
.system-settings-page { overflow-x: hidden; color: var(--na-foreground); }
.settings-heading { margin-bottom: var(--na-space-lg); }
.eyebrow { margin: 0 0 var(--na-space-xs); color: var(--na-primary); font-size: .75rem; font-weight: 600; line-height: 1.4; }
.settings-heading h1 { margin: 0; font-size: 1.5rem; font-weight: 700; line-height: 1.25; }
.settings-heading p:last-child { max-width: 65ch; margin: var(--na-space-xs) 0 0; color: var(--na-muted-foreground); font-size: .8125rem; line-height: 1.5; }
.setting-card { overflow: hidden; }
.setting-card-header { gap: var(--na-space-lg); }
.setting-card-header h2 { margin: 0; font-size: 1rem; font-weight: 600; line-height: 1.35; }
.setting-card-header p { max-width: 65ch; margin: var(--na-space-xs) 0 0; color: var(--na-muted-foreground); font-size: .75rem; line-height: 1.5; }
.brand-editor { display: grid; grid-template-columns: minmax(320px, 1fr) minmax(260px, .72fr); align-items: center; gap: var(--na-space-xl); padding: var(--na-space-lg); }
.brand-fields { min-width: 0; }
.brand-actions { display: flex; flex-wrap: wrap; gap: var(--na-space-xs); }
.brand-preview { display: flex; min-width: 0; align-items: center; gap: var(--na-space-md); padding: var(--na-space-lg); border: 1px solid var(--na-border); border-radius: var(--na-radius-sm); background: var(--na-muted); }
.brand-preview > span { display: flex; min-width: 0; flex-direction: column; gap: var(--na-space-2xs); }
.brand-preview strong, .brand-preview small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.brand-preview strong { font-size: 1rem; font-weight: 700; line-height: 1.35; }
.brand-preview small { color: var(--na-muted-foreground); font-size: .75rem; line-height: 1.4; }
.current-background { display: grid; grid-template-columns: minmax(260px, 420px) minmax(220px, 1fr); align-items: center; gap: var(--na-space-lg); padding: var(--na-space-lg); }
.current-preview { position: relative; overflow: hidden; aspect-ratio: 16 / 7; border: 1px solid var(--el-border-color-lighter); border-radius: 10px; background: var(--el-fill-color-light); }
.current-preview img { width: 100%; height: 100%; object-fit: cover; }
.current-badge { position: absolute; top: 10px; left: 10px; border-radius: 999px; background: rgb(15 23 42 / 86%); padding: 4px 9px; color: #fff; font-size: .6875rem; font-weight: 600; }
.current-info { min-width: 0; }
.current-label { display: block; margin-bottom: 6px; color: var(--na-muted-foreground); font-size: .75rem; }
.current-info strong { display: block; overflow: hidden; font-size: 1rem; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.current-info p { max-width: 65ch; margin: var(--na-space-xs) 0 0; color: var(--na-muted-foreground); font-size: .75rem; line-height: 1.6; }
.current-logo { display: grid; grid-template-columns: 80px minmax(200px, 1fr) auto; align-items: center; gap: var(--na-space-lg); min-height: 120px; padding: var(--na-space-lg); }
.logo-preview { display: grid; place-items: center; width: 72px; height: 72px; overflow: hidden; border: 1px solid var(--el-border-color-lighter); border-radius: 14px; background: var(--el-fill-color-extra-light); }
.logo-preview img { width: 48px; height: 48px; object-fit: contain; }
.logo-actions { display: flex; align-items: center; gap: var(--na-space-xs); }
.background-manager { border-top: 1px solid var(--el-border-color-lighter); background: var(--el-fill-color-extra-light); padding: 18px 20px 20px; }
.manager-toolbar { display: flex; align-items: center; justify-content: space-between; gap: var(--na-space-lg); margin-bottom: var(--na-space-md); }
.manager-toolbar h3 { margin: 0; font-size: .875rem; font-weight: 600; }
.manager-toolbar p { max-width: 65ch; margin: var(--na-space-2xs) 0 0; color: var(--na-muted-foreground); font-size: .75rem; line-height: 1.5; }
.background-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 12px; min-height: 144px; }
.background-option { position: relative; min-width: 0; overflow: hidden; padding: 0; border: 1px solid var(--el-border-color); border-radius: 10px; background: var(--el-bg-color); color: inherit; text-align: left; cursor: pointer; transition: border-color 180ms ease, box-shadow 180ms ease; }
.background-option:hover { border-color: var(--el-color-primary-light-5); }
.background-option:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 2px; }
.background-option.selected { border-color: var(--el-color-primary); box-shadow: 0 0 0 2px var(--el-color-primary-light-9); }
.thumbnail-wrap { position: relative; display: block; aspect-ratio: 16 / 8; overflow: hidden; background: var(--el-fill-color-light); }
.thumbnail-wrap img { width: 100%; height: 100%; object-fit: cover; }
.selected-mark { position: absolute; right: 9px; bottom: 9px; display: grid; place-items: center; width: 24px; height: 24px; border-radius: 50%; background: var(--el-color-primary); color: #fff; box-shadow: 0 2px 8px rgb(0 0 0 / 18%); }
.option-info { display: flex; min-width: 0; flex-direction: column; gap: 3px; padding: 10px 42px 11px 12px; }
.option-info strong { overflow: hidden; font-size: .8125rem; text-overflow: ellipsis; white-space: nowrap; }
.option-info span { color: var(--el-text-color-secondary); font-size: .6875rem; }
.delete-background { position: absolute; right: 7px; bottom: 6px; width: 32px; height: 32px; padding: 0; }
.empty-gallery { grid-column: 1 / -1; display: flex; min-height: 144px; align-items: center; justify-content: center; flex-direction: column; gap: 6px; border: 1px dashed var(--el-border-color); border-radius: 10px; color: var(--el-text-color-secondary); }
.empty-gallery .el-icon { font-size: 1.875rem; }
.empty-gallery strong { color: var(--el-text-color-primary); font-size: .8125rem; }
.empty-gallery span { font-size: .75rem; }
.manager-actions { display: flex; align-items: center; justify-content: space-between; gap: 18px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--el-border-color-lighter); }
.manager-actions > span { overflow: hidden; color: var(--el-text-color-secondary); font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.manager-actions > div { display: flex; flex: 0 0 auto; gap: 8px; }
@media (max-width: 720px) {
  .system-settings-page { padding: 12px; }
  .setting-card-header,
  .manager-toolbar,
  .manager-actions { align-items: stretch; flex-direction: column; }
  .brand-editor, .current-background { grid-template-columns: 1fr; gap: var(--na-space-md); padding: var(--na-space-md); }
  .current-logo { grid-template-columns: 72px minmax(0, 1fr); gap: 14px; padding: 16px; }
  .logo-actions { grid-column: 1 / -1; flex-wrap: wrap; }
  .background-manager { padding: 16px; }
  .background-grid { grid-template-columns: 1fr; }
  .manager-actions > div { justify-content: flex-end; }
}
@media (prefers-reduced-motion: reduce) { .background-option { transition: none; } }
</style>
