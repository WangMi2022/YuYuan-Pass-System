<template>
  <div class="layout-mode-grid" role="radiogroup" aria-label="布局模式" :style="{ '--layout-accent': primaryColor }">
    <button
      v-for="layout in layoutModes"
      :key="layout.value"
      type="button"
      class="layout-mode-option"
      :class="{ 'is-active': modelValue === layout.value }"
      role="radio"
      :aria-checked="modelValue === layout.value"
      @click="handleLayoutChange(layout.value)"
    >
      <span class="layout-mode-preview" :class="`is-${layout.value}`" aria-hidden="true">
        <i class="layout-mode-preview__header" />
        <i class="layout-mode-preview__sidebar" />
        <span class="layout-mode-preview__content"><i /><i /><i /></span>
        <i class="layout-mode-preview__secondary" />
      </span>
      <span class="layout-mode-option__copy">
        <strong>{{ layout.label }}</strong>
        <small>{{ layout.description }}</small>
      </span>
      <span v-if="modelValue === layout.value" class="layout-mode-option__check" aria-hidden="true">
        <el-icon><Check /></el-icon>
      </span>
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { Check } from '@element-plus/icons-vue'
import { useAppStore } from '@/pinia'

defineOptions({
  name: 'LayoutModeCard'
})

defineProps({
  modelValue: {
    type: String,
    default: 'normal'
  }
})

const emit = defineEmits(['update:modelValue'])
const appStore = useAppStore()
const { config } = storeToRefs(appStore)
const primaryColor = computed(() => config.value.primaryColor)

const layoutModes = [
  { value: 'normal', label: '经典布局', description: '左侧导航与顶部工具栏' },
  { value: 'head', label: '顶部导航', description: '主菜单集中在顶部' },
  { value: 'combination', label: '混合布局', description: '顶部与侧栏组合导航' },
  { value: 'sidebar', label: '侧栏常驻', description: '二级菜单持续展开' }
]

const handleLayoutChange = (layout) => {
  emit('update:modelValue', layout)
}
</script>

<style scoped lang="scss">
.layout-mode-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }

.layout-mode-option {
  position: relative;
  display: grid;
  min-width: 0;
  min-height: 136px;
  grid-template-columns: 116px minmax(0, 1fr);
  align-items: center;
  gap: 13px;
  padding: 12px;
  border: 1px solid var(--na-border);
  border-radius: 8px;
  background: var(--na-card);
  color: var(--na-foreground);
  text-align: left;
  transition: border-color 160ms ease, background-color 160ms ease, box-shadow 160ms ease;
}

.layout-mode-option:hover { border-color: var(--na-border-strong); background: color-mix(in srgb, var(--na-muted) 55%, var(--na-card)); }
.layout-mode-option.is-active { border-color: var(--layout-accent); box-shadow: 0 0 0 2px color-mix(in srgb, var(--layout-accent) 16%, transparent); }

.layout-mode-preview {
  position: relative;
  display: block;
  width: 116px;
  height: 78px;
  overflow: hidden;
  border: 1px solid var(--na-border-strong);
  border-radius: 5px;
  background: var(--na-background);
}

.layout-mode-preview__header { position: absolute; top: 0; right: 0; left: 25px; height: 15px; border-bottom: 1px solid var(--na-border); background: var(--na-card); }
.layout-mode-preview__sidebar { position: absolute; inset: 0 auto 0 0; width: 25px; background: var(--layout-accent); }
.layout-mode-preview__content { position: absolute; inset: 23px 8px 8px 33px; display: grid; grid-template-columns: 1fr 1fr; gap: 5px; }
.layout-mode-preview__content i { border-radius: 2px; background: var(--na-card); }
.layout-mode-preview__content i:first-child { grid-column: 1 / -1; }
.layout-mode-preview__secondary { display: none; }

.layout-mode-preview.is-head .layout-mode-preview__header { left: 0; height: 20px; background: var(--layout-accent); }
.layout-mode-preview.is-head .layout-mode-preview__sidebar { display: none; }
.layout-mode-preview.is-head .layout-mode-preview__content { inset: 28px 8px 8px; }
.layout-mode-preview.is-combination .layout-mode-preview__header { left: 18px; background: var(--layout-accent); }
.layout-mode-preview.is-combination .layout-mode-preview__sidebar { width: 18px; background: color-mix(in srgb, var(--layout-accent) 55%, var(--na-card)); }
.layout-mode-preview.is-combination .layout-mode-preview__content { inset: 23px 19px 8px 26px; }
.layout-mode-preview.is-combination .layout-mode-preview__secondary { display: block; position: absolute; inset: 15px 0 0 auto; width: 11px; border-left: 1px solid var(--na-border); background: var(--na-card); }
.layout-mode-preview.is-sidebar .layout-mode-preview__sidebar { width: 36px; }
.layout-mode-preview.is-sidebar .layout-mode-preview__header { left: 36px; }
.layout-mode-preview.is-sidebar .layout-mode-preview__content { left: 44px; }

.layout-mode-option__copy { display: flex; min-width: 0; flex-direction: column; gap: 4px; }
.layout-mode-option__copy strong { font-size: 12px; font-weight: 650; }
.layout-mode-option__copy small { color: var(--na-muted-foreground); font-size: 10px; line-height: 1.5; }
.layout-mode-option__check { position: absolute; top: 8px; right: 8px; display: grid; width: 18px; height: 18px; place-items: center; border-radius: 50%; background: var(--layout-accent); color: var(--na-on-primary); font-size: 11px; }

@media (max-width: 560px) {
  .layout-mode-grid { grid-template-columns: 1fr; }
  .layout-mode-option { min-height: 112px; grid-template-columns: 120px minmax(0, 1fr); }
}
</style>
