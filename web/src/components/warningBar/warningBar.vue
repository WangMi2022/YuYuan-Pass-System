<template>
  <component
    :is="href ? 'a' : 'div'"
    class="na-warning-bar"
    :class="href && 'is-link'"
    :href="href || undefined"
    :target="href ? '_blank' : undefined"
    :rel="href ? 'noopener noreferrer' : undefined"
  >
    <el-icon class="na-warning-bar__icon">
      <warning-filled />
    </el-icon>
    <span>
      {{ title }}
    </span>
  </component>
</template>
<script setup>
  import { WarningFilled } from '@element-plus/icons-vue'
  defineProps({
    title: {
      type: String,
      default: ''
    },
    href: {
      type: String,
      default: ''
    }
  })

</script>

<style lang="scss" scoped>
  .na-warning-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 2px 0 14px;
    padding: 9px 12px;
    border: 1px solid color-mix(in srgb, var(--na-warning) 28%, transparent);
    border-radius: var(--na-radius-sm);
    background: var(--na-warning-soft);
    color: var(--na-warning);
    font-size: 13px;
    line-height: 1.5;
    text-decoration: none;

    &.is-link { cursor: pointer; }

    &__icon { flex: 0 0 auto; font-size: 15px; }

    span { color: color-mix(in srgb, var(--na-warning) 72%, var(--na-foreground)); }
  }
</style>
