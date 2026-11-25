<template>
  <el-dialog
    v-model="dialogVisible"
    width="680px"
    class="command-menu-dialog"
    modal-class="command-menu-backdrop"
    :show-close="false"
    append-to-body
    @opened="focusInput"
    @closed="handleClosed"
  >
    <template #header>
      <div class="command-search">
        <el-icon class="command-search__icon"><Search /></el-icon>
        <input
          ref="commandInput"
          v-model="searchInput"
          type="search"
          role="combobox"
          aria-label="搜索菜单与功能"
          aria-autocomplete="list"
          aria-controls="command-menu-results"
          :aria-activedescendant="activeDescendant"
          :aria-expanded="dialogVisible"
          autocomplete="off"
          placeholder="搜索菜单与功能…"
          @keydown.down.prevent="moveSelection(1)"
          @keydown.up.prevent="moveSelection(-1)"
          @keydown.enter.prevent="executeSelected"
        />
        <kbd>Esc</kbd>
      </div>
    </template>

    <div
      id="command-menu-results"
      class="command-results"
      role="listbox"
      aria-label="快捷导航结果"
    >
      <template v-if="flatItems.length">
        <section
          v-for="option in options"
          :key="option.id"
          class="command-group"
          role="group"
          :aria-labelledby="`command-group-${option.id}`"
        >
          <header v-if="option.children.length" class="command-group__header">
            <span :id="`command-group-${option.id}`">{{ option.label }}</span>
            <small>{{ option.children.length }}</small>
          </header>

          <button
            v-for="item in option.children"
            :id="`command-option-${item.keyboardIndex}`"
            :key="item.id"
            type="button"
            tabindex="-1"
            class="command-item"
            :class="{
              'is-selected': item.keyboardIndex === selectedIndex,
              'is-current': item.isCurrent
            }"
            role="option"
            :aria-selected="item.keyboardIndex === selectedIndex"
            @mouseenter="selectedIndex = item.keyboardIndex"
            @click="item.func"
          >
            <span class="command-item__icon" aria-hidden="true">
              <el-icon><component :is="item.icon" /></el-icon>
            </span>
            <span class="command-item__copy">
              <strong>{{ item.label }}</strong>
              <small v-if="item.description">{{ item.description }}</small>
            </span>
            <span v-if="item.isCurrent" class="command-item__badge">当前页面</span>
            <el-icon v-else class="command-item__arrow"><ArrowRight /></el-icon>
          </button>
        </section>
      </template>

      <div v-else class="command-empty">
        <span class="command-empty__icon" aria-hidden="true">
          <el-icon><Search /></el-icon>
        </span>
        <strong>没有找到相关功能</strong>
        <span>换一个关键词试试</span>
      </div>
    </div>

    <template #footer>
      <div class="command-footer">
        <div class="command-footer__hints" aria-hidden="true">
          <span><kbd>↑</kbd><kbd>↓</kbd> 选择</span>
          <span><kbd>Enter</kbd> 打开</span>
          <span class="command-footer__optional"><kbd>Esc</kbd> 关闭</span>
        </div>
        <span>{{ flatItems.length }} 个结果</span>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
  import { computed, nextTick, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { useRouterStore } from '@/pinia/modules/router'
  import { useAppStore, useUserStore } from '@/pinia'

  defineOptions({
    name: 'CommandMenu'
  })

  const appStore = useAppStore()
  const userStore = useUserStore()
  const router = useRouter()
  const route = useRoute()
  const routerStore = useRouterStore()

  const dialogVisible = ref(false)
  const searchInput = ref('')
  const commandInput = ref()
  const selectedIndex = ref(0)

  const close = () => {
    dialogVisible.value = false
  }

  const changeRouter = (menu) => {
    const name = String(menu.name || '')
    if (name === route.name) {
      close()
      return
    }

    const query = {}
    const params = {}
    routerStore.routeMap[name]?.parameters?.forEach((item) => {
      if (item.type === 'query') {
        query[item.key] = item.value
      } else {
        params[item.key] = item.value
      }
    })

    if (name.startsWith('http://') || name.startsWith('https://')) {
      window.open(name, '_blank', 'noopener,noreferrer')
    } else {
      router.push({ name, query, params })
    }
    close()
  }

  const changeMode = (darkMode) => {
    appStore.toggleTheme(darkMode)
    close()
  }

  const logout = () => {
    close()
    userStore.LoginOut()
  }

  const normalizedSearch = computed(() => searchInput.value.trim().toLocaleLowerCase())

  const matchesSearch = (...values) => {
    if (!normalizedSearch.value) return true
    return values.some((value) =>
      String(value || '').toLocaleLowerCase().includes(normalizedSearch.value)
    )
  }

  const deepMenus = (menus, parentLabels = []) => {
    const items = []
    menus?.forEach((menu) => {
      if (menu.hidden) return

      const label = menu.meta?.title || ''
      if (menu.children?.length) {
        const nextParents = label ? [...parentLabels, label] : parentLabels
        items.push(...deepMenus(menu.children, nextParents))
        return
      }

      const description = parentLabels.join(' / ')
      if (label && matchesSearch(label, description)) {
        items.push({
          id: `route-${menu.name}`,
          label,
          description,
          icon: menu.meta?.icon || 'Menu',
          isCurrent: menu.name === route.name,
          func: () => changeRouter(menu)
        })
      }
    })
    return items
  }

  const quickActions = [
    {
      id: 'theme-light',
      label: '切换到亮色主题',
      description: '使用明亮界面外观',
      icon: 'Sunny',
      func: () => changeMode(false)
    },
    {
      id: 'theme-dark',
      label: '切换到深色主题',
      description: '使用深色界面外观',
      icon: 'Moon',
      func: () => changeMode(true)
    },
    {
      id: 'logout',
      label: '退出登录',
      description: '安全退出当前账号',
      icon: 'SwitchButton',
      func: logout
    }
  ]

  const options = computed(() => {
    let keyboardIndex = 0
    const groups = [
      {
        id: 'navigation',
        label: '页面跳转',
        children: deepMenus(routerStore.asyncRouters[0]?.children || [])
      },
      {
        id: 'actions',
        label: '快捷操作',
        children: quickActions.filter((item) =>
          matchesSearch(item.label, item.description)
        )
      }
    ]

    return groups
      .filter((group) => group.children.length)
      .map((group) => ({
        ...group,
        children: group.children.map((item) => ({
          ...item,
          keyboardIndex: keyboardIndex++
        }))
      }))
  })

  const flatItems = computed(() => options.value.flatMap((option) => option.children))
  const activeDescendant = computed(() =>
    flatItems.value.length ? `command-option-${selectedIndex.value}` : undefined
  )

  const moveSelection = (step) => {
    if (!flatItems.value.length) return
    selectedIndex.value =
      (selectedIndex.value + step + flatItems.value.length) % flatItems.value.length
  }

  const executeSelected = () => {
    flatItems.value[selectedIndex.value]?.func()
  }

  const focusInput = async () => {
    await nextTick()
    commandInput.value?.focus()
  }

  const open = () => {
    searchInput.value = ''
    selectedIndex.value = 0
    dialogVisible.value = true
  }

  const handleClosed = () => {
    searchInput.value = ''
    selectedIndex.value = 0
  }

  watch(searchInput, () => {
    selectedIndex.value = 0
  })

  defineExpose({ open })
</script>

<style lang="scss">
  .command-menu-backdrop {
    background: rgb(13 13 18 / 52%);
    backdrop-filter: blur(3px);
  }

  .command-menu-dialog {
    width: min(680px, calc(100vw - 32px)) !important;
    max-height: calc(100vh - 64px);
    margin-top: min(11vh, 96px);
    overflow: hidden;
    border: 0;
    border-radius: 16px;
    box-shadow: var(--na-shadow-md);

    .el-dialog__header {
      margin: 0;
      padding: 16px 16px 10px;
      border-bottom: 0;
    }

    .el-dialog__body {
      min-height: 0;
      padding: 0 10px !important;
      overflow: hidden;
    }

    .el-dialog__footer {
      padding: 10px 16px;
      background: color-mix(in srgb, var(--na-muted) 62%, var(--na-card));
    }
  }

  .command-search {
    display: flex;
    align-items: center;
    gap: 10px;
    height: 48px;
    padding: 0 12px;
    border: 1px solid var(--na-input);
    border-radius: 12px;
    background: color-mix(in srgb, var(--na-muted) 70%, var(--na-card));
    transition: border-color 150ms ease, box-shadow 150ms ease, background-color 150ms ease;

    &:focus-within {
      border-color: var(--na-primary);
      background: var(--na-card);
      box-shadow: 0 0 0 3px var(--na-ring);
    }

    input {
      min-width: 0;
      height: 100%;
      flex: 1;
      border: 0;
      outline: 0;
      background: transparent;
      color: var(--na-foreground);
      font: 500 15px/1 var(--na-font-sans);

      &::placeholder {
        color: var(--na-muted-foreground);
        opacity: 1;
      }

      &::-webkit-search-cancel-button { display: none; }
    }
  }

  .command-search__icon {
    flex: 0 0 auto;
    color: var(--na-muted-foreground);
    font-size: 18px;
  }

  .command-search kbd,
  .command-footer kbd {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 24px;
    height: 22px;
    padding: 0 6px;
    border: 1px solid var(--na-border-strong);
    border-radius: 6px;
    background: var(--na-card);
    color: var(--na-muted-foreground);
    box-shadow: 0 1px 0 var(--na-border-strong);
    font: 500 10px/1 var(--na-font-sans);
  }

  .command-results {
    min-height: 180px;
    max-height: min(56vh, 480px);
    overflow-y: auto;
    overscroll-behavior: contain;
    padding: 0 2px 10px;
  }

  .command-group__header {
    position: sticky;
    z-index: 1;
    top: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 10px 6px;
    background: var(--na-card);
    color: var(--na-muted-foreground);

    span {
      font-size: 11px;
      font-weight: 650;
    }

    small { font-size: 10px; }
  }

  .command-item {
    display: flex;
    width: 100%;
    min-height: 52px;
    align-items: center;
    gap: 10px;
    padding: 7px 10px;
    border: 0;
    border-radius: 10px;
    outline: 0;
    background: transparent;
    color: var(--na-foreground);
    text-align: left;
    cursor: pointer;
    transition: background-color 150ms cubic-bezier(.22, 1, .36, 1), color 150ms ease;

    &.is-selected,
    &:hover {
      background: var(--na-primary-soft);
    }

    &:focus-visible {
      box-shadow: inset 0 0 0 2px var(--na-primary);
    }

    &:active {
      background: color-mix(in srgb, var(--na-primary) 16%, var(--na-card));
    }
  }

  .command-item__icon,
  .command-empty__icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 auto;
    border-radius: 9px;
    background: var(--na-muted);
    color: var(--na-muted-foreground);
  }

  .command-item__icon {
    width: 34px;
    height: 34px;
    font-size: 16px;
  }

  .command-item.is-selected .command-item__icon,
  .command-item:hover .command-item__icon {
    background: color-mix(in srgb, var(--na-primary) 13%, var(--na-card));
    color: var(--na-primary);
  }

  .command-item__copy {
    display: flex;
    min-width: 0;
    flex: 1;
    flex-direction: column;
    gap: 3px;

    strong {
      overflow: hidden;
      color: inherit;
      font-size: 13px;
      font-weight: 580;
      line-height: 1.25;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    small {
      overflow: hidden;
      color: var(--na-muted-foreground);
      font-size: 11px;
      line-height: 1.25;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .command-item__badge {
    flex: 0 0 auto;
    padding: 4px 7px;
    border-radius: 999px;
    background: var(--na-primary-soft);
    color: var(--na-primary);
    font-size: 10px;
    font-weight: 600;
  }

  .command-item__arrow {
    flex: 0 0 auto;
    color: var(--na-muted-foreground);
    font-size: 13px;
    opacity: 0;
    translate: -4px 0;
    transition: opacity 150ms ease, translate 150ms cubic-bezier(.22, 1, .36, 1);
  }

  .command-item.is-selected .command-item__arrow,
  .command-item:hover .command-item__arrow {
    opacity: 1;
    translate: 0;
  }

  .command-empty {
    display: flex;
    min-height: 220px;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: 7px;
    color: var(--na-muted-foreground);
    text-align: center;

    strong {
      margin-top: 3px;
      color: var(--na-foreground);
      font-size: 14px;
      font-weight: 600;
    }

    > span:last-child { font-size: 12px; }
  }

  .command-empty__icon {
    width: 42px;
    height: 42px;
    color: var(--na-primary);
    font-size: 18px;
  }

  .command-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    color: var(--na-muted-foreground);
    font-size: 10px;
  }

  .command-footer__hints {
    display: flex;
    align-items: center;
    gap: 14px;

    span {
      display: inline-flex;
      align-items: center;
      gap: 4px;
    }
  }

  @media (max-width: 640px) {
    .command-menu-dialog {
      width: calc(100vw - 20px) !important;
      max-height: calc(100vh - 24px);
      margin-top: 12px;

      .el-dialog__header { padding: 12px 12px 8px; }
      .el-dialog__body { padding: 0 6px !important; }
      .el-dialog__footer { padding: 9px 12px; }
    }

    .command-results { max-height: calc(100vh - 168px); }
    .command-item { min-height: 56px; }
    .command-item__copy strong { font-size: 14px; }
    .command-footer__optional { display: none !important; }
  }

  @media (prefers-reduced-motion: reduce) {
    .command-search,
    .command-item,
    .command-item__arrow { transition: none; }
  }
</style>
