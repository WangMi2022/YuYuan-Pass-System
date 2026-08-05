<template>
  <section class="media-category-tree" aria-label="媒体分类">
    <div class="media-category-tree__header">
      <div class="media-category-tree__heading">
        <span class="media-category-tree__heading-icon" aria-hidden="true">
          <el-icon><FolderOpened /></el-icon>
        </span>
        <div>
          <h2 class="media-category-tree__title">媒体分类</h2>
          <span class="media-category-tree__count">{{ categoryCount }} 个分类</span>
        </div>
      </div>
      <el-tooltip content="新建一级分类" placement="top">
        <el-button
          class="media-category-tree__create"
          circle
          text
          size="small"
          aria-label="新建一级分类"
          @click="emit('add', rootCategory)"
        >
          <el-icon><Plus /></el-icon>
        </el-button>
      </el-tooltip>
    </div>

    <div class="media-category-tree__body">
      <button
        type="button"
        class="media-category-tree__all"
        :class="{ 'is-active': activeId === rootCategory.ID }"
        :aria-current="activeId === rootCategory.ID ? 'page' : undefined"
        @click="emit('select', rootCategory)"
      >
        <span class="media-category-tree__all-icon" aria-hidden="true">
          <el-icon><Files /></el-icon>
        </span>
        <span class="media-category-tree__all-label">全部分类</span>
        <el-icon v-if="activeId === rootCategory.ID" class="media-category-tree__check" aria-hidden="true">
          <Check />
        </el-icon>
      </button>

      <div class="media-category-tree__section-label">
        <span>分类列表</span>
        <span>{{ categoryCount }}</span>
      </div>

      <el-scrollbar class="media-category-tree__scrollbar" :max-height="scrollHeight">
        <el-tree
          :data="treeCategories"
          node-key="ID"
          :props="defaultProps"
          :current-node-key="activeId"
          :empty-text="emptyText"
          :indent="16"
          highlight-current
          default-expand-all
          :expand-on-click-node="false"
          class="media-category-tree__tree"
          @node-click="emit('select', $event)"
        >
          <template #default="{ data }">
            <div
              class="media-category-tree__node"
              :class="{
                'is-active': activeId === data.ID
              }"
            >
              <el-icon class="media-category-tree__node-icon" aria-hidden="true">
                <FolderOpened v-if="data.children?.length" />
                <Folder v-else />
              </el-icon>
              <span class="media-category-tree__label" :title="data.name">{{ data.name }}</span>

              <el-tooltip v-if="data.ID > 0" content="分类操作" placement="top">
                <el-dropdown trigger="click">
                  <el-button
                    class="media-category-tree__action"
                    circle
                    text
                    size="small"
                    :aria-label="`${data.name} 分类操作`"
                    @click.stop
                  >
                    <el-icon><MoreFilled /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="emit('add', data)">
                        <el-icon><Plus /></el-icon>
                        添加子分类
                      </el-dropdown-item>
                      <el-dropdown-item @click="emit('edit', data)">
                        <el-icon><EditPen /></el-icon>
                        编辑分类
                      </el-dropdown-item>
                      <el-dropdown-item divided class="media-category-tree__delete-item" @click="emit('delete', data.ID)">
                        <el-icon><Delete /></el-icon>
                        删除分类
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </el-tooltip>
            </div>
          </template>
        </el-tree>
      </el-scrollbar>
    </div>
  </section>
</template>

<script setup>
  import { computed } from 'vue'
  import { Check, Delete, EditPen, Files, Folder, FolderOpened, MoreFilled, Plus } from '@element-plus/icons-vue'

  const props = defineProps({
    categories: {
      type: Array,
      default: () => []
    },
    activeId: {
      type: Number,
      default: 0
    },
    scrollHeight: {
      type: String,
      default: 'calc(100vh - 300px)'
    },
    emptyText: {
      type: String,
      default: '暂无分类'
    }
  })

  const emit = defineEmits(['select', 'add', 'edit', 'delete'])

  const defaultProps = {
    children: 'children',
    label: 'name',
    value: 'ID'
  }

  const rootCategory = computed(() => {
    return props.categories.find((item) => item.ID === 0) || {
      ID: 0,
      name: '全部分类',
      pid: 0,
      children: []
    }
  })

  const treeCategories = computed(() => {
    return props.categories.filter((item) => item.ID !== 0)
  })

  const categoryCount = computed(() => {
    const count = (items = []) => items.reduce((total, item) => {
      const children = Array.isArray(item.children) ? item.children : []
      return total + (item.ID > 0 ? 1 : 0) + count(children)
    }, 0)

    return count(props.categories)
  })
</script>

<style scoped lang="scss">
.media-category-tree {
  min-width: 0;
  overflow: hidden;
  border-radius: inherit;
  color: var(--na-foreground);
}

.media-category-tree__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 68px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--na-border);
  background: color-mix(in srgb, var(--na-primary) 3%, var(--na-card));
}

.media-category-tree__heading {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  color: var(--na-foreground);
}

.media-category-tree__heading-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  border-radius: 9px;
  background: var(--na-primary-soft);
  color: var(--na-primary);
  font-size: 17px;
}

.media-category-tree__title {
  margin: 0 0 4px;
  color: var(--na-foreground);
  font-size: 14px;
  font-weight: 680;
  line-height: 1;
}

.media-category-tree__count {
  display: block;
  color: var(--na-muted-foreground);
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
}

.media-category-tree__create.el-button,
.media-category-tree__action.el-button {
  width: 30px;
  min-width: 30px;
  min-height: 30px;
  height: 30px;
  padding: 0;
  color: var(--na-muted-foreground);
}

.media-category-tree__create.el-button {
  border: 1px solid color-mix(in srgb, var(--na-primary) 22%, var(--na-border));
  background: var(--na-primary-soft);
  color: var(--na-primary);
}

.media-category-tree__create.el-button:hover,
.media-category-tree__action.el-button:hover {
  border-color: color-mix(in srgb, var(--na-primary) 35%, var(--na-border));
  background: color-mix(in srgb, var(--na-primary) 14%, var(--na-card));
  color: var(--na-primary);
}

.media-category-tree__body {
  padding: 10px;
}

.media-category-tree__all {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: 44px;
  gap: 10px;
  padding: 6px 10px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--na-foreground);
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition: color 180ms cubic-bezier(.22, 1, .36, 1), background-color 180ms cubic-bezier(.22, 1, .36, 1), box-shadow 180ms cubic-bezier(.22, 1, .36, 1);
}

.media-category-tree__all:hover {
  background: var(--na-muted);
}

.media-category-tree__all:focus-visible {
  outline: 3px solid var(--na-ring);
  outline-offset: 1px;
}

.media-category-tree__all.is-active {
  background: var(--na-primary-soft);
  color: var(--na-primary);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--na-primary) 18%, transparent);
}

.media-category-tree__all-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  flex: 0 0 28px;
  border-radius: 7px;
  background: var(--na-muted);
  color: var(--na-muted-foreground);
  font-size: 15px;
}

.media-category-tree__all.is-active .media-category-tree__all-icon {
  background: color-mix(in srgb, var(--na-primary) 15%, var(--na-card));
  color: var(--na-primary);
}

.media-category-tree__all-label {
  min-width: 0;
  overflow: hidden;
  flex: 1;
  font-size: 13px;
  font-weight: 620;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.media-category-tree__check {
  flex: 0 0 auto;
  font-size: 14px;
}

.media-category-tree__section-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin: 12px 7px 5px;
  color: var(--na-muted-foreground);
  font-size: 11px;
  font-weight: 560;
  line-height: 1;
}

.media-category-tree__scrollbar {
  min-width: 0;
}

.media-category-tree__tree {
  --el-tree-node-hover-bg-color: transparent;

  min-width: 0;
  padding-bottom: 2px;
  background: transparent;
}

.media-category-tree :deep(.el-tree-node__content) {
  height: 36px;
  margin: 1px 0;
  padding-right: 4px;
  border-radius: 8px;
  color: var(--na-foreground);
  transition: background-color 180ms cubic-bezier(.22, 1, .36, 1), color 180ms cubic-bezier(.22, 1, .36, 1);
}

.media-category-tree :deep(.el-tree-node__content:hover) {
  background: var(--na-muted);
}

.media-category-tree :deep(.el-tree-node__content:focus-visible) {
  position: relative;
  z-index: 1;
  outline: 3px solid var(--na-ring);
  outline-offset: 1px;
}

.media-category-tree :deep(.el-tree--highlight-current .el-tree-node.is-current > .el-tree-node__content) {
  background: var(--na-primary-soft);
  color: var(--na-primary);
}

.media-category-tree :deep(.el-tree-node__expand-icon) {
  width: 20px;
  margin-left: 2px;
  color: var(--na-muted-foreground);
  font-size: 13px;
}

.media-category-tree :deep(.el-tree-node__expand-icon.is-leaf) {
  color: transparent;
}

.media-category-tree__node {
  display: flex;
  align-items: center;
  min-width: 0;
  width: 100%;
  gap: 8px;
}

.media-category-tree__node-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  flex: 0 0 24px;
  border-radius: 7px;
  background: color-mix(in srgb, var(--na-muted) 72%, var(--na-card));
  color: var(--na-muted-foreground);
  font-size: 14px;
}

.media-category-tree__node.is-active .media-category-tree__node-icon {
  background: color-mix(in srgb, var(--na-primary) 14%, var(--na-card));
  color: var(--na-primary);
}

.media-category-tree__label {
  min-width: 0;
  overflow: hidden;
  color: inherit;
  font-size: 13px;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.media-category-tree__node.is-active .media-category-tree__label {
  color: var(--na-primary);
  font-weight: 650;
}

.media-category-tree__action.el-button {
  flex: 0 0 auto;
  margin-left: auto;
  opacity: 0;
}

.media-category-tree :deep(.el-tree-node__content:hover .media-category-tree__action.el-button),
.media-category-tree__node.is-active .media-category-tree__action.el-button,
.media-category-tree__action.el-button:focus-visible {
  opacity: 1;
}

:deep(.media-category-tree__delete-item) {
  color: var(--na-action-danger);
}

:deep(.media-category-tree__delete-item:hover) {
  background: var(--na-danger-soft);
  color: var(--na-action-danger);
}

@media (prefers-reduced-motion: reduce) {
  .media-category-tree__all,
  .media-category-tree :deep(.el-tree-node__content) {
    transition: none;
  }
}
</style>
