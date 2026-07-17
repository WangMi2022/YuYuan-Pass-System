<template>
  <div class="setting-row">
    <div class="setting-row__label">
      <strong>{{ label }}</strong>
      <p v-if="description">{{ description }}</p>
      <slot v-else name="suffix" />
    </div>
    <div class="setting-row__control">
      <slot />
    </div>
  </div>
</template>

<script setup>
defineOptions({
  name: 'SettingItem'
})

defineProps({
  label: {
    type: String,
    required: true
  },
  description: {
    type: String,
    default: ''
  }
})
</script>

<style scoped lang="scss">
.setting-row {
  display: flex;
  min-height: 68px;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--na-border);
}

.setting-row:last-child { border-bottom: 0; }
.setting-row__label { min-width: 0; }
.setting-row__label strong { display: block; color: var(--na-foreground); font-size: 12px; font-weight: 620; line-height: 1.5; }
.setting-row__label p { margin: 3px 0 0; color: var(--na-muted-foreground); font-size: 10px; line-height: 1.5; }
.setting-row__control { display: flex; min-width: 120px; flex: 0 0 auto; align-items: center; justify-content: flex-end; }

.setting-row__control :deep(.el-switch) {
  --el-switch-on-color: var(--na-primary);
}

.setting-row__control :deep(.el-select) { width: 148px; }
.setting-row__control :deep(.el-input-number) { width: 126px; }

@media (max-width: 480px) {
  .setting-row { min-height: 72px; gap: 12px; padding: 12px 14px; }
  .setting-row__label p { max-width: 180px; }
  .setting-row__control { min-width: auto; }
  .setting-row__control :deep(.el-select) { width: 124px; }
  .setting-row__control :deep(.el-input-number) { width: 112px; }
}
</style>
