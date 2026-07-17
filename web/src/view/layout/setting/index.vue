<template>
  <el-drawer
    v-model="drawer"
    direction="rtl"
    :size="width"
    :show-close="false"
    append-to-body
    class="gva-theme-drawer"
  >
    <template #header="{ close, titleId, titleClass }">
      <header class="settings-header">
        <div class="settings-header__identity">
          <span class="settings-header__icon" aria-hidden="true">
            <el-icon><Operation /></el-icon>
          </span>
          <div>
            <h2 :id="titleId" :class="titleClass">系统配置</h2>
            <p>界面偏好与工作区设置</p>
          </div>
        </div>

        <div class="settings-header__actions">
          <span class="settings-save-state" :class="`is-${saveState}`" aria-live="polite">
            <i aria-hidden="true" />
            {{ saveStateText }}
          </span>
          <el-button :icon="RefreshLeft" @click="resetConfig">恢复默认</el-button>
          <el-tooltip content="关闭设置" placement="bottom">
            <el-button class="settings-close-button" :icon="Close" circle aria-label="关闭设置" @click="close" />
          </el-tooltip>
        </div>
      </header>
    </template>

    <div class="settings-workspace">
      <aside class="settings-sidebar">
        <nav class="settings-nav" aria-label="系统配置分类">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            type="button"
            class="settings-nav__item"
            :class="{ 'is-active': activeTab === tab.key }"
            :aria-current="activeTab === tab.key ? 'page' : undefined"
            @click="activeTab = tab.key"
          >
            <el-icon><component :is="tab.icon" /></el-icon>
            <span>
              <strong>{{ tab.label }}</strong>
              <small>{{ tab.description }}</small>
            </span>
          </button>
        </nav>

        <div class="settings-sidebar__summary">
          <span class="settings-sidebar__swatch" :style="{ backgroundColor: config.primaryColor }" />
          <span>
            <small>当前主题</small>
            <strong>{{ currentModeLabel }}</strong>
          </span>
        </div>
      </aside>

      <main class="settings-content" tabindex="-1">
        <div class="settings-content__inner">
          <Transition name="settings-view" mode="out-in">
            <AppearanceSettings v-if="activeTab === 'appearance'" key="appearance" />
            <LayoutSettings v-else-if="activeTab === 'layout'" key="layout" />
            <GeneralSettings v-else key="general" @reset="resetConfig" />
          </Transition>
        </div>
      </main>
    </div>
  </el-drawer>
</template>

<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Brush, Close, Grid, Operation, RefreshLeft, Setting } from '@element-plus/icons-vue'
import { useAppStore } from '@/pinia'
import { setSelfSetting } from '@/api/user'
import AppearanceSettings from './modules/appearance/index.vue'
import LayoutSettings from './modules/layout/index.vue'
import GeneralSettings from './modules/general/index.vue'

defineOptions({
  name: 'GvaSetting'
})

const appStore = useAppStore()
const { config, device } = storeToRefs(appStore)

const activeTab = ref('appearance')
const saveState = ref('saved')
let saveTimer = null

const tabs = [
  { key: 'appearance', label: '外观', description: '主题、色彩与显示', icon: Brush },
  { key: 'layout', label: '布局', description: '导航与界面尺寸', icon: Grid },
  { key: 'general', label: '通用', description: '配置与系统信息', icon: Setting }
]

const modeLabels = {
  light: '浅色模式',
  dark: '深色模式',
  auto: '跟随系统'
}

const currentModeLabel = computed(() => modeLabels[config.value.darkMode] || '自定义主题')
const saveStateText = computed(() => ({
  saved: '已保存',
  saving: '保存中',
  error: '保存失败'
}[saveState.value]))

const width = computed(() => device.value === 'mobile' ? '100%' : 'min(820px, 94vw)')

const drawer = defineModel('drawer', {
  default: true,
  type: Boolean
})

const saveConfig = async () => {
  saveState.value = 'saving'
  try {
    const res = await setSelfSetting(config.value)
    if (res.code !== 0) {
      throw new Error(res.msg || '保存失败')
    }
    localStorage.setItem('originSetting', JSON.stringify(config.value))
    saveState.value = 'saved'
  } catch (error) {
    saveState.value = 'error'
    ElMessage.error(error.message || '配置保存失败')
  }
}

const scheduleSave = () => {
  saveState.value = 'saving'
  window.clearTimeout(saveTimer)
  saveTimer = window.setTimeout(saveConfig, 450)
}

const resetConfig = async () => {
  try {
    await ElMessageBox.confirm(
      '将外观、布局和通用选项恢复为系统默认值。',
      '恢复默认配置',
      {
        confirmButtonText: '恢复默认',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    appStore.resetConfig()
    ElMessage.success('已恢复默认配置')
  } catch {
    // User cancelled.
  }
}

watch(config, scheduleSave, { deep: true })

onBeforeUnmount(() => {
  window.clearTimeout(saveTimer)
})
</script>

<style lang="scss">
.gva-theme-drawer.el-drawer,
.gva-theme-drawer .el-drawer {
  border-left: 1px solid var(--na-border);
  background: var(--na-background);
  box-shadow: -18px 0 48px rgb(35 53 65 / 14%);
}

.gva-theme-drawer .el-drawer__header {
  height: 72px;
  margin: 0;
  padding: 0;
  border-bottom: 1px solid var(--na-border);
  background: var(--na-card);
}

.gva-theme-drawer .el-drawer__body {
  height: calc(100% - 72px);
  overflow: hidden;
  padding: 0;
}

.settings-header {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 0 20px 0 22px;
}

.settings-header__identity,
.settings-header__actions,
.settings-header__identity > div {
  display: flex;
  align-items: center;
}

.settings-header__identity { min-width: 0; gap: 12px; }
.settings-header__identity > div { min-width: 0; align-items: flex-start; flex-direction: column; }
.settings-header__identity h2 { margin: 0; color: var(--na-foreground); font-size: 17px; font-weight: 650; line-height: 1.4; }
.settings-header__identity p { margin: 1px 0 0; color: var(--na-muted-foreground); font-size: 11px; line-height: 1.4; }

.settings-header__icon {
  display: inline-grid;
  width: 36px;
  height: 36px;
  place-items: center;
  flex: 0 0 36px;
  border-radius: 8px;
  background: var(--na-primary-soft);
  color: var(--na-accent-foreground);
  font-size: 18px;
}

.settings-header__actions { flex: 0 0 auto; gap: 8px; }
.settings-header__actions .el-button + .el-button { margin-left: 0; }
.settings-close-button { width: 36px; height: 36px; }

.settings-save-state {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-right: 2px;
  color: var(--na-muted-foreground);
  font-size: 11px;
  white-space: nowrap;
}

.settings-save-state i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--na-success);
}

.settings-save-state.is-saving i { background: var(--na-warning); animation: settings-saving 900ms ease-in-out infinite alternate; }
.settings-save-state.is-error { color: var(--na-danger); }
.settings-save-state.is-error i { background: var(--na-danger); }

.settings-workspace {
  display: grid;
  height: 100%;
  min-height: 0;
  grid-template-columns: 192px minmax(0, 1fr);
}

.settings-sidebar {
  display: flex;
  min-height: 0;
  flex-direction: column;
  justify-content: space-between;
  border-right: 1px solid var(--na-border);
  background: var(--na-card);
}

.settings-nav { display: grid; gap: 4px; padding: 18px 12px; }
.settings-nav__item {
  display: grid;
  min-height: 60px;
  grid-template-columns: 28px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--na-muted-foreground);
  text-align: left;
  transition: color 160ms ease, background-color 160ms ease;
}

.settings-nav__item:hover { background: var(--na-muted); color: var(--na-foreground); }
.settings-nav__item.is-active { background: var(--na-primary-soft); color: var(--na-accent-foreground); }
.settings-nav__item > .el-icon { width: 28px; font-size: 17px; }
.settings-nav__item > span { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.settings-nav__item strong { font-size: 13px; font-weight: 650; }
.settings-nav__item small { overflow: hidden; color: var(--na-muted-foreground); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.settings-nav__item.is-active small { color: color-mix(in srgb, var(--na-accent-foreground) 72%, var(--na-muted-foreground)); }

.settings-sidebar__summary {
  display: flex;
  align-items: center;
  gap: 9px;
  margin: 0 12px 14px;
  padding: 12px 10px;
  border-top: 1px solid var(--na-border);
}

.settings-sidebar__swatch { width: 22px; height: 22px; flex: 0 0 22px; border: 3px solid var(--na-card); border-radius: 6px; box-shadow: 0 0 0 1px var(--na-border-strong); }
.settings-sidebar__summary > span:last-child { display: flex; min-width: 0; flex-direction: column; }
.settings-sidebar__summary small { color: var(--na-muted-foreground); font-size: 9px; }
.settings-sidebar__summary strong { overflow: hidden; color: var(--na-foreground); font-size: 11px; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }

.settings-content {
  min-width: 0;
  min-height: 0;
  overflow: auto;
  background: var(--na-background);
  overscroll-behavior: contain;
}

.settings-content__inner { width: 100%; max-width: 660px; margin: 0 auto; padding: 28px; }

.settings-module { color: var(--na-foreground); }
.settings-module__header { margin-bottom: 28px; }
.settings-module__eyebrow { margin: 0 0 5px; color: var(--na-accent-foreground); font-size: 10px; font-weight: 650; }
.settings-module__header h3 { margin: 0; color: var(--na-foreground); font-size: 22px; font-weight: 680; line-height: 1.35; }
.settings-module__header p:last-child { margin: 6px 0 0; color: var(--na-muted-foreground); font-size: 12px; line-height: 1.6; }

.settings-section { margin-bottom: 28px; padding-bottom: 28px; border-bottom: 1px solid var(--na-border); }
.settings-section:last-child { margin-bottom: 0; padding-bottom: 0; border-bottom: 0; }
.settings-section__header { margin-bottom: 14px; }
.settings-section__header h4 { margin: 0; color: var(--na-foreground); font-size: 14px; font-weight: 650; }
.settings-section__header p { margin: 4px 0 0; color: var(--na-muted-foreground); font-size: 11px; line-height: 1.55; }
.settings-section__meta { color: var(--na-muted-foreground); font-size: 10px; }

.settings-panel {
  overflow: hidden;
  border: 1px solid var(--na-border);
  border-radius: 8px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
}

.settings-view-enter-active,
.settings-view-leave-active { transition: opacity 140ms ease, transform 140ms ease; }
.settings-view-enter-from { opacity: 0; transform: translateY(4px); }
.settings-view-leave-to { opacity: 0; transform: translateY(-2px); }

@keyframes settings-saving { from { opacity: .42; } to { opacity: 1; } }

@media (max-width: 700px) {
  .gva-theme-drawer .el-drawer__header { height: 64px; }
  .gva-theme-drawer .el-drawer__body { height: calc(100% - 64px); }
  .settings-header { padding: 0 12px 0 14px; }
  .settings-header__identity p,
  .settings-save-state,
  .settings-header__actions > .el-button:not(.settings-close-button) { display: none; }
  .settings-header__icon { width: 34px; height: 34px; flex-basis: 34px; }
  .settings-workspace { display: flex; flex-direction: column; }
  .settings-sidebar { flex: 0 0 auto; border-right: 0; border-bottom: 1px solid var(--na-border); }
  .settings-nav { display: grid; grid-template-columns: repeat(3, 1fr); gap: 6px; padding: 8px 10px; }
  .settings-nav__item { min-height: 46px; grid-template-columns: 20px auto; justify-content: center; gap: 6px; padding: 6px 8px; text-align: center; }
  .settings-nav__item > .el-icon { width: 20px; font-size: 15px; }
  .settings-nav__item > span { display: block; }
  .settings-nav__item strong { font-size: 12px; }
  .settings-nav__item small,
  .settings-sidebar__summary { display: none; }
  .settings-content { flex: 1 1 auto; }
  .settings-content__inner { padding: 22px 16px 32px; }
  .settings-module__header { margin-bottom: 22px; }
  .settings-module__header h3 { font-size: 20px; }
}

@media (prefers-reduced-motion: reduce) {
  .settings-view-enter-active,
  .settings-view-leave-active { transition: none; }
  .settings-save-state.is-saving i { animation: none; }
}
</style>
