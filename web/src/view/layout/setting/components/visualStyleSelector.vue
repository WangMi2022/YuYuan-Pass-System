<template>
  <div class="visual-style-grid" role="radiogroup" aria-label="界面质感" :style="{ '--style-accent': primaryColor }">
    <button
      v-for="style in visualStyles"
      :key="style.value"
      type="button"
      class="visual-style-option"
      :class="{ 'is-active': modelValue === style.value }"
      role="radio"
      :aria-checked="modelValue === style.value"
      @click="emit('update:modelValue', style.value)"
    >
      <span class="visual-style-preview" :class="`is-${style.value}`" aria-hidden="true">
        <i class="preview-header" />
        <i class="preview-sidebar" />
        <span class="preview-content"><i /><i /><i /></span>
      </span>
      <span class="visual-style-copy">
        <span><el-icon><component :is="style.icon" /></el-icon><strong>{{ style.label }}</strong></span>
        <small>{{ style.description }}</small>
      </span>
      <span v-if="modelValue === style.value" class="visual-style-check" aria-hidden="true">
        <el-icon><Check /></el-icon>
      </span>
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { Check, Grid, Operation } from '@element-plus/icons-vue'
import { useAppStore } from '@/pinia'

defineOptions({ name: 'VisualStyleSelector' })

defineProps({
  modelValue: {
    type: String,
    default: 'default'
  }
})

const emit = defineEmits(['update:modelValue'])
const appStore = useAppStore()
const { config } = storeToRefs(appStore)
const primaryColor = computed(() => config.value.primaryColor)

const visualStyles = [
  { value: 'default', label: '标准界面', description: '清晰边界与轻量层级', icon: Grid },
  { value: 'neumorphism', label: '新拟态', description: '柔和浮雕与触感反馈', icon: Operation }
]
</script>

<style scoped lang="scss">
.visual-style-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
.visual-style-option {
  position: relative;
  display: grid;
  min-width: 0;
  grid-template-columns: 118px minmax(0, 1fr);
  align-items: center;
  gap: 14px;
  min-height: 112px;
  padding: 12px;
  border: 1px solid var(--na-border);
  border-radius: 8px;
  background: var(--na-card);
  color: var(--na-foreground);
  text-align: left;
  transition: border-color 160ms ease, background-color 160ms ease, box-shadow 160ms ease, transform 140ms ease;
}
.visual-style-option:hover { border-color: var(--na-border-strong); background: color-mix(in srgb, var(--na-muted) 55%, var(--na-card)); }
.visual-style-option:active { transform: translateY(1px); }
.visual-style-option.is-active { border-color: var(--style-accent); box-shadow: 0 0 0 2px color-mix(in srgb, var(--style-accent) 16%, transparent); }
.visual-style-preview { position: relative; display: block; width: 118px; height: 76px; overflow: hidden; border: 1px solid #dfe2e8; border-radius: 7px; background: #f2f4f7; }
.preview-header { position: absolute; inset: 0 0 auto; height: 15px; border-bottom: 1px solid #dfe2e8; background: #fff; }
.preview-sidebar { position: absolute; inset: 15px auto 0 0; width: 26px; border-right: 1px solid #dfe2e8; background: #fff; }
.preview-content { position: absolute; inset: 24px 9px 9px 35px; display: grid; grid-template-columns: 1fr 1fr; gap: 7px; }
.preview-content i { display: block; border-radius: 3px; background: #fff; }
.preview-content i:first-child { grid-column: 1 / -1; background: color-mix(in srgb, var(--style-accent) 28%, #fff); }
.visual-style-preview.is-neumorphism { border-color: transparent; background: #e9edf4; box-shadow: inset 2px 2px 5px rgb(151 161 180 / 24%), inset -2px -2px 5px rgb(255 255 255 / 88%); }
.visual-style-preview.is-neumorphism .preview-header { border: 0; background: #e9edf4; box-shadow: 0 3px 7px rgb(151 161 180 / 20%); }
.visual-style-preview.is-neumorphism .preview-sidebar { border: 0; background: #e9edf4; box-shadow: 3px 0 7px rgb(151 161 180 / 18%); }
.visual-style-preview.is-neumorphism .preview-content i { background: #e9edf4; box-shadow: -2px -2px 4px rgb(255 255 255 / 90%), 2px 2px 4px rgb(151 161 180 / 28%); }
.visual-style-preview.is-neumorphism .preview-content i:first-child { background: color-mix(in srgb, var(--style-accent) 14%, #e9edf4); box-shadow: inset 2px 2px 4px rgb(151 161 180 / 22%), inset -2px -2px 4px rgb(255 255 255 / 72%); }
.visual-style-copy { display: flex; min-width: 0; flex-direction: column; gap: 5px; }
.visual-style-copy > span { display: flex; min-width: 0; align-items: center; gap: 7px; }
.visual-style-copy .el-icon { color: var(--na-muted-foreground); font-size: 15px; }
.visual-style-copy strong { overflow: hidden; font-size: 12px; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.visual-style-copy small { color: var(--na-muted-foreground); font-size: 9px; line-height: 1.45; }
.visual-style-check { position: absolute; top: 8px; right: 8px; display: grid; width: 18px; height: 18px; place-items: center; border-radius: 50%; background: var(--style-accent); color: var(--na-on-primary); font-size: 11px; }

@media (max-width: 520px) {
  .visual-style-grid { grid-template-columns: 1fr; }
  .visual-style-option { grid-template-columns: 108px minmax(0, 1fr); }
  .visual-style-preview { width: 108px; }
}
@media (prefers-reduced-motion: reduce) {
  .visual-style-option { transition: none; }
}
</style>
