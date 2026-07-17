<template>
  <div class="theme-color-picker">
    <div class="theme-color-grid" role="radiogroup" aria-label="预设主题颜色">
      <button
        v-for="colorItem in presetColors"
        :key="colorItem.color"
        type="button"
        class="theme-color-option"
        :class="{ 'is-active': normalizedValue === colorItem.color.toLowerCase() }"
        role="radio"
        :aria-checked="normalizedValue === colorItem.color.toLowerCase()"
        :aria-label="`使用${colorItem.name}主题色`"
        @click="handleColorChange(colorItem.color)"
      >
        <span class="theme-color-option__swatch" :style="{ backgroundColor: colorItem.color }">
          <el-icon v-if="normalizedValue === colorItem.color.toLowerCase()"><Check /></el-icon>
        </span>
        <span>{{ colorItem.name }}</span>
      </button>
    </div>

    <div class="theme-color-custom">
      <div>
        <strong>自定义主题色</strong>
        <p>从色板中选择任意品牌颜色</p>
      </div>
      <div class="theme-color-custom__control">
        <code>{{ modelValue.toUpperCase() }}</code>
        <el-color-picker
          v-model="customColor"
          size="large"
          :predefine="presetColors.map(item => item.color)"
          aria-label="选择自定义主题色"
          @change="handleCustomColorChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { Check } from '@element-plus/icons-vue'

defineOptions({
  name: 'ThemeColorPicker'
})

const props = defineProps({
  modelValue: {
    type: String,
    default: '#6366f1'
  }
})

const emit = defineEmits(['update:modelValue'])
const customColor = ref(props.modelValue)
const normalizedValue = computed(() => props.modelValue.toLowerCase())

const presetColors = [
  { color: '#6366f1', name: '靛蓝' },
  { color: '#8b5cf6', name: '雪青' },
  { color: '#3b82f6', name: '蔚蓝' },
  { color: '#0ea5e9', name: '晴空' },
  { color: '#06b6d4', name: '青碧' },
  { color: '#14b8a6', name: '孔雀' },
  { color: '#10b981', name: '翡翠' },
  { color: '#84cc16', name: '嫩芽' },
  { color: '#f59e0b', name: '琥珀' },
  { color: '#f97316', name: '暖橙' },
  { color: '#f43f5e', name: '绯红' },
  { color: '#ec4899', name: '樱粉' },
  { color: '#20aaa6', name: '经典青' }
]

const handleColorChange = (color) => {
  customColor.value = color
  emit('update:modelValue', color)
}

const handleCustomColorChange = (color) => {
  if (color) emit('update:modelValue', color)
}

watch(() => props.modelValue, (newValue) => {
  customColor.value = newValue
})
</script>

<style scoped lang="scss">
.theme-color-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 8px; }

.theme-color-option {
  display: flex;
  min-width: 0;
  min-height: 48px;
  align-items: center;
  gap: 8px;
  padding: 7px 9px;
  border: 1px solid var(--na-border);
  border-radius: 6px;
  background: var(--na-card);
  color: var(--na-foreground);
  font-size: 10px;
  font-weight: 580;
  text-align: left;
  transition: border-color 160ms ease, background-color 160ms ease, box-shadow 160ms ease;
}

.theme-color-option:hover { border-color: var(--na-border-strong); background: var(--na-muted); }
.theme-color-option.is-active { border-color: var(--na-primary); box-shadow: 0 0 0 2px var(--na-ring); }
.theme-color-option > span:last-child { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.theme-color-option__swatch {
  display: grid;
  width: 28px;
  height: 28px;
  place-items: center;
  flex: 0 0 28px;
  border: 2px solid rgb(255 255 255 / 78%);
  border-radius: 6px;
  color: var(--na-on-primary);
  font-size: 12px;
  box-shadow: 0 0 0 1px var(--na-border-strong);
}

.theme-color-custom {
  display: flex;
  min-height: 70px;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin-top: 12px;
  padding: 12px 14px;
  border: 1px solid var(--na-border);
  border-radius: 8px;
  background: var(--na-card);
}

.theme-color-custom strong { display: block; color: var(--na-foreground); font-size: 12px; font-weight: 650; }
.theme-color-custom p { margin: 3px 0 0; color: var(--na-muted-foreground); font-size: 10px; }
.theme-color-custom__control { display: flex; align-items: center; gap: 9px; }
.theme-color-custom code { padding: 6px 8px; border: 1px solid var(--na-border); border-radius: 5px; background: var(--na-muted); color: var(--na-foreground); font-size: 10px; }
.theme-color-custom :deep(.el-color-picker__trigger) { border-color: var(--na-border-strong); border-radius: 6px; }

@media (max-width: 520px) {
  .theme-color-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .theme-color-custom { align-items: flex-start; flex-direction: column; }
  .theme-color-custom__control { width: 100%; justify-content: space-between; }
}
</style>
