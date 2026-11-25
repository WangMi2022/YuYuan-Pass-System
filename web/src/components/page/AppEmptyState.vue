<template>
  <section
    class="na-empty-state"
    :class="{ 'is-compact': compact }"
    role="region"
    :aria-label="title"
  >
    <span class="na-empty-state__icon" aria-hidden="true">
      <el-icon><Document /></el-icon>
    </span>
    <div class="na-empty-state__content">
      <h3>{{ title }}</h3>
      <p v-if="description">{{ description }}</p>
      <ul v-if="highlights.length" aria-label="当前状态">
        <li v-for="item in highlights" :key="item">
          <el-icon><CircleCheck /></el-icon>
          <span>{{ item }}</span>
        </li>
      </ul>
    </div>
    <div v-if="$slots.actions" class="na-empty-state__actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<script setup>
import { CircleCheck, Document } from '@element-plus/icons-vue'

defineOptions({ name: 'AppEmptyState' })

defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' },
  highlights: { type: Array, default: () => [] },
  compact: { type: Boolean, default: false }
})
</script>

<style scoped lang="scss">
.na-empty-state {
  display: grid;
  min-width: 0;
  min-height: 220px;
  grid-template-columns: 48px minmax(0, 1fr);
  align-items: center;
  gap: 16px;
  padding: 24px 28px;
  background: color-mix(in srgb, var(--na-primary) 3%, var(--na-card));
}

.na-empty-state.is-compact {
  min-height: 148px;
  padding: 18px 20px;
}

.na-empty-state__icon {
  display: grid;
  width: 48px;
  height: 48px;
  place-items: center;
  border-radius: 8px;
  background: var(--na-primary-soft);
  color: var(--na-primary);
  font-size: 22px;
}

.na-empty-state__content {
  min-width: 0;
}

.na-empty-state h3 {
  margin: 0;
  color: var(--na-foreground);
  font-size: .9375rem;
  font-weight: 660;
  line-height: 1.4;
}

.na-empty-state p {
  max-width: 68ch;
  margin: 5px 0 0;
  color: var(--na-muted-foreground);
  font-size: .75rem;
  line-height: 1.6;
  text-wrap: pretty;
}

.na-empty-state ul {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 12px;
  margin: 11px 0 0;
  padding: 0;
  list-style: none;
}

.na-empty-state li {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--na-muted-foreground);
  font-size: .6875rem;
}

.na-empty-state li .el-icon {
  color: var(--na-success);
}

.na-empty-state__actions {
  display: flex;
  grid-column: 2;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
}

@media (max-width: 640px) {
  .na-empty-state,
  .na-empty-state.is-compact {
    min-height: 0;
    grid-template-columns: 40px minmax(0, 1fr);
    gap: 12px;
    padding: 18px 16px;
  }

  .na-empty-state__icon {
    width: 40px;
    height: 40px;
    font-size: 19px;
  }

  .na-empty-state__actions {
    grid-column: 1 / -1;
    justify-content: stretch;
  }

  .na-empty-state__actions :deep(.el-button) {
    min-height: 44px;
    flex: 1;
  }
}
</style>
