<template>
  <nav class="na-bottom-navigation" aria-label="主导航">
    <button
      v-for="item in primaryItems"
      :key="item.name"
      type="button"
      class="na-bottom-navigation__action"
      :class="{ 'is-active': containsActiveRoute(item, active) }"
      :disabled="isDisabled(item)"
      :aria-current="containsActiveRoute(item, active) ? 'page' : undefined"
      @click="openItem(item)"
    >
      <span class="na-bottom-navigation__icon" aria-hidden="true">
        <el-icon v-if="item.meta?.icon">
          <component :is="item.meta.icon" />
        </el-icon>
        <span v-else>{{ getMenuTitle(item).slice(0, 1) }}</span>
      </span>
      <span v-if="showLabels" class="na-bottom-navigation__label">
        {{ getMenuTitle(item) }}
      </span>
    </button>

    <button
      v-if="overflowItems.length"
      type="button"
      class="na-bottom-navigation__action"
      :class="{ 'is-active': overflowIsActive }"
      :aria-current="overflowIsActive ? 'page' : undefined"
      @click="openOverflow"
    >
      <span class="na-bottom-navigation__icon" aria-hidden="true">
        <el-icon><MoreFilled /></el-icon>
      </span>
      <span v-if="showLabels" class="na-bottom-navigation__label">更多</span>
    </button>
  </nav>

  <el-drawer
    v-model="sheetOpen"
    direction="btt"
    size="min(70vh, 520px)"
    :title="sheetTitle"
    append-to-body
    class="na-bottom-navigation-sheet"
  >
    <div class="na-bottom-navigation-sheet__list">
      <button
        v-for="entry in sheetItems"
        :key="entry.item.name"
        type="button"
        class="na-bottom-navigation-sheet__item"
        :class="{ 'is-active': containsActiveRoute(entry.item, active) }"
        :disabled="isDisabled(entry.item)"
        @click="selectSheetItem(entry.item)"
      >
        <span class="na-bottom-navigation-sheet__icon" aria-hidden="true">
          <el-icon v-if="entry.item.meta?.icon">
            <component :is="entry.item.meta.icon" />
          </el-icon>
          <span v-else>{{ getMenuTitle(entry.item).slice(0, 1) }}</span>
        </span>
        <span class="na-bottom-navigation-sheet__copy">
          <strong>{{ getMenuTitle(entry.item) }}</strong>
          <small v-if="entry.trail.length">{{ entry.trail.join(' / ') }}</small>
        </span>
      </button>
    </div>
  </el-drawer>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { MoreFilled } from '@element-plus/icons-vue'
  import {
    containsActiveRoute,
    flattenLeafMenuItems,
    getMenuTitle,
    getVisibleMenuItems,
    splitNavigationItems
  } from './bottomNavigation.js'

  defineOptions({
    name: 'BottomNavigation'
  })

  const props = defineProps({
    items: {
      type: Array,
      default: () => []
    },
    active: {
      type: [String, Symbol],
      default: ''
    },
    maxItems: {
      type: Number,
      default: 5
    },
    showLabels: {
      type: Boolean,
      default: true
    }
  })

  const emit = defineEmits(['select'])
  const sheetOpen = ref(false)
  const sheetTitle = ref('')
  const sheetItems = ref([])

  const navigationGroups = computed(() =>
    splitNavigationItems(props.items, props.maxItems)
  )
  const primaryItems = computed(() => navigationGroups.value.primaryItems)
  const overflowItems = computed(() => navigationGroups.value.overflowItems)
  const overflowIsActive = computed(() =>
    overflowItems.value.some((item) => containsActiveRoute(item, props.active))
  )

  const isDisabled = (item) => Boolean(item?.disabled || item?.meta?.disabled)

  const openItem = (item) => {
    const children = getVisibleMenuItems(item.children)
    if (!children.length) {
      emit('select', item)
      return
    }
    sheetTitle.value = getMenuTitle(item)
    sheetItems.value = flattenLeafMenuItems(children, [item])
    sheetOpen.value = true
  }

  const openOverflow = () => {
    sheetTitle.value = '更多导航'
    sheetItems.value = flattenLeafMenuItems(overflowItems.value)
    sheetOpen.value = true
  }

  const selectSheetItem = (item) => {
    sheetOpen.value = false
    emit('select', item)
  }
</script>

<style lang="scss">
  .na-bottom-navigation {
    position: fixed;
    z-index: 45;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    min-height: calc(var(--na-bottom-navigation-height) + env(safe-area-inset-bottom));
    align-items: flex-start;
    padding-bottom: env(safe-area-inset-bottom);
    overflow: hidden;
    background: var(--na-card);
    box-shadow:
      0 1px 8px rgb(0 0 0 / 12%),
      0 3px 4px rgb(0 0 0 / 14%),
      0 3px 3px -2px rgb(0 0 0 / 20%);
  }

  .na-bottom-navigation__action {
    display: flex;
    min-width: 0;
    height: var(--na-bottom-navigation-height);
    flex: 1 1 0;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 6px 12px 8px;
    border: 0;
    border-radius: 0;
    background: transparent;
    color: var(--na-muted-foreground);
    cursor: pointer;
    transition: color 140ms ease, background-color 140ms ease;
  }

  .na-bottom-navigation__action:hover:not(:disabled) {
    background: var(--na-muted);
    color: var(--na-foreground);
  }

  .na-bottom-navigation__action:focus-visible {
    position: relative;
    z-index: 1;
    outline: 2px solid var(--na-ring);
    outline-offset: -3px;
  }

  .na-bottom-navigation__action.is-active { color: var(--na-primary); }
  .na-bottom-navigation__action:disabled {
    color: color-mix(in srgb, var(--na-muted-foreground) 55%, transparent);
    cursor: not-allowed;
  }

  .na-bottom-navigation__icon {
    display: grid;
    width: 24px;
    height: 24px;
    place-items: center;
    flex: 0 0 24px;
    font-size: 16px;
    font-weight: 650;
  }
  .na-bottom-navigation__icon .el-icon { font-size: 20px; }

  .na-bottom-navigation__label {
    width: 100%;
    height: 16px;
    overflow: hidden;
    font-size: 12px;
    font-weight: 400;
    line-height: 16px;
    text-align: center;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .na-bottom-navigation-sheet {
    border-radius: 8px 8px 0 0;
  }
  .na-bottom-navigation-sheet .el-drawer__header {
    min-height: 52px;
    margin: 0;
    padding: 14px 16px;
  }
  .na-bottom-navigation-sheet .el-drawer__body {
    padding: 8px 12px calc(16px + env(safe-area-inset-bottom));
  }
  .na-bottom-navigation-sheet__list {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 4px;
  }
  .na-bottom-navigation-sheet__item {
    display: flex;
    min-width: 0;
    min-height: 52px;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border: 0;
    border-radius: 8px;
    background: transparent;
    color: var(--na-foreground);
    text-align: left;
  }
  .na-bottom-navigation-sheet__item:hover:not(:disabled),
  .na-bottom-navigation-sheet__item.is-active {
    background: var(--na-primary-soft);
    color: var(--na-primary);
  }
  .na-bottom-navigation-sheet__item:focus-visible {
    outline: 2px solid var(--na-ring);
    outline-offset: -2px;
  }
  .na-bottom-navigation-sheet__item:disabled { opacity: .45; }
  .na-bottom-navigation-sheet__icon {
    display: grid;
    width: 32px;
    height: 32px;
    place-items: center;
    flex: 0 0 32px;
    color: inherit;
    font-size: 14px;
  }
  .na-bottom-navigation-sheet__icon .el-icon { font-size: 19px; }
  .na-bottom-navigation-sheet__copy {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 2px;
  }
  .na-bottom-navigation-sheet__copy strong,
  .na-bottom-navigation-sheet__copy small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .na-bottom-navigation-sheet__copy strong { font-size: 13px; font-weight: 600; }
  .na-bottom-navigation-sheet__copy small {
    color: var(--na-muted-foreground);
    font-size: 11px;
    font-weight: 400;
  }

  @media (max-width: 420px) {
    .na-bottom-navigation-sheet__list { grid-template-columns: minmax(0, 1fr); }
  }
</style>
