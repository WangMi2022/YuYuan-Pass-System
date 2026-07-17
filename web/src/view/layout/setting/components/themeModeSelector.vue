<template>
  <div class="theme-mode-grid" role="radiogroup" aria-label="主题模式" :style="{ '--mode-accent': primaryColor }">
    <button
      v-for="mode in themeModes"
      :key="mode.value"
      type="button"
      class="theme-mode-option"
      :class="{ 'is-active': modelValue === mode.value }"
      role="radio"
      :aria-checked="modelValue === mode.value"
      @click="handleModeChange(mode.value)"
    >
      <span class="theme-mode-preview" :class="`is-${mode.value}`" aria-hidden="true">
        <i class="theme-mode-preview__header" />
        <i class="theme-mode-preview__sidebar" />
        <span class="theme-mode-preview__content"><i /><i /><i /></span>
      </span>
      <span class="theme-mode-option__label">
        <el-icon><component :is="mode.icon" /></el-icon>
        <strong>{{ mode.label }}</strong>
      </span>
      <small>{{ mode.description }}</small>
      <span v-if="modelValue === mode.value" class="theme-mode-option__check" aria-hidden="true">
        <el-icon><Check /></el-icon>
      </span>
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { Check, Monitor, Moon, Sunny } from '@element-plus/icons-vue'
import { useAppStore } from '@/pinia'

defineOptions({
  name: 'ThemeModeSelector'
})

defineProps({
  modelValue: {
    type: String,
    default: 'auto'
  }
})

const emit = defineEmits(['update:modelValue'])
const appStore = useAppStore()
const { config } = storeToRefs(appStore)
const primaryColor = computed(() => config.value.primaryColor)

const themeModes = [
  { value: 'light', label: '浅色', description: '明亮工作区', icon: Sunny },
  { value: 'dark', label: '深色', description: '低亮度界面', icon: Moon },
  { value: 'auto', label: '跟随系统', description: '自动切换', icon: Monitor }
]

const handleModeChange = (mode) => {
  emit('update:modelValue', mode)
}
</script>

<style scoped lang="scss">
.theme-mode-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; }

.theme-mode-option {
  position: relative;
  min-width: 0;
  min-height: 142px;
  padding: 10px;
  border: 1px solid var(--na-border);
  border-radius: 8px;
  background: var(--na-card);
  color: var(--na-foreground);
  text-align: left;
  transition: border-color 160ms ease, background-color 160ms ease, box-shadow 160ms ease;
}

.theme-mode-option:hover { border-color: var(--na-border-strong); background: color-mix(in srgb, var(--na-muted) 55%, var(--na-card)); }
.theme-mode-option.is-active { border-color: var(--mode-accent); box-shadow: 0 0 0 2px color-mix(in srgb, var(--mode-accent) 16%, transparent); }

.theme-mode-preview {
  position: relative;
  display: block;
  height: 70px;
  overflow: hidden;
  margin-bottom: 9px;
  border: 1px solid #e4e4e7;
  border-radius: 5px;
  background: #f4f4f5;
}

.theme-mode-preview__header { position: absolute; inset: 0 0 auto 0; height: 14px; border-bottom: 1px solid #e4e4e7; background: #fff; }
.theme-mode-preview__sidebar { position: absolute; inset: 14px auto 0 0; width: 24px; border-right: 1px solid #e4e4e7; background: #fff; }
.theme-mode-preview__content { position: absolute; inset: 22px 8px 8px 32px; display: grid; grid-template-columns: 1fr 1fr; gap: 5px; }
.theme-mode-preview__content i { display: block; border-radius: 2px; background: #fff; }
.theme-mode-preview__content i:first-child { grid-column: 1 / -1; }

.theme-mode-preview.is-dark { border-color: #3f3f46; background: #101012; }
.theme-mode-preview.is-dark .theme-mode-preview__header,
.theme-mode-preview.is-dark .theme-mode-preview__sidebar,
.theme-mode-preview.is-dark .theme-mode-preview__content i { border-color: #2c2c30; background: #1f1f23; }
.theme-mode-preview.is-dark .theme-mode-preview__content i:first-child { background: color-mix(in srgb, var(--mode-accent) 48%, #1f1f23); }

.theme-mode-preview.is-light .theme-mode-preview__content i:first-child,
.theme-mode-preview.is-auto .theme-mode-preview__content i:first-child { background: color-mix(in srgb, var(--mode-accent) 30%, #fff); }
.theme-mode-preview.is-auto .theme-mode-preview__sidebar { background: #1f1f23; }
.theme-mode-preview.is-auto .theme-mode-preview__content i:last-child { background: #2c2c30; }

.theme-mode-option__label { display: flex; align-items: center; gap: 6px; }
.theme-mode-option__label .el-icon { color: var(--na-muted-foreground); font-size: 14px; }
.theme-mode-option__label strong { overflow: hidden; font-size: 12px; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.theme-mode-option > small { display: block; margin: 2px 0 0 20px; color: var(--na-muted-foreground); font-size: 9px; }
.theme-mode-option__check { position: absolute; top: 7px; right: 7px; display: grid; width: 18px; height: 18px; place-items: center; border-radius: 50%; background: var(--mode-accent); color: var(--na-on-primary); font-size: 11px; }

@media (max-width: 480px) {
  .theme-mode-grid { gap: 7px; }
  .theme-mode-option { min-height: 120px; padding: 7px; }
  .theme-mode-preview { height: 58px; }
  .theme-mode-option > small { display: none; }
  .theme-mode-option__label { justify-content: center; }
}
</style>
