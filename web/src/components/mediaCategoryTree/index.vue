<template>
  <section class="media-category-tree" aria-label="媒体分类">
    <div class="media-category-tree__header">
      <div class="media-category-tree__heading">
        <span>分类</span>
        <span class="media-category-tree__count">{{ categoryCount }} 个</span>
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

    <el-scrollbar class="media-category-tree__scrollbar" :height="scrollHeight">
      <el-tree
        :data="categories"
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
              'is-active': activeId === data.ID,
              'is-root': data.ID === 0
            }"
          >
            <el-icon class="media-category-tree__node-icon" aria-hidden="true">
              <FolderOpened v-if="data.ID === 0 || data.children?.length" />
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
  </section>
</template>

<script setup>
  import { computed } from 'vue'
  import { Delete, EditPen, Folder, FolderOpened, MoreFilled, Plus } from '@element-plus/icons-vue'

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
  color: var(--na-foreground);
}

.media-category-tree__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 30px;
  margin: 0 2px 8px;
  padding: 0 3px 0 8px;
}

.media-category-tree__heading {
  display: inline-flex;
  align-items: baseline;
  gap: 7px;
  min-width: 0;
  color: var(--na-foreground);
  font-size: 12px;
  font-weight: 680;
  line-height: 1;
}

.media-category-tree__count {
  color: var(--na-muted-foreground);
  font-size: 11px;
  font-weight: 500;
}

.media-category-tree__create.el-button,
.media-category-tree__action.el-button {
  width: 28px;
  min-width: 28px;
  min-height: 28px;
  height: 28px;
  padding: 0;
  color: var(--na-muted-foreground);
}

.media-category-tree__create.el-button:hover,
.media-category-tree__action.el-button:hover {
  background: var(--na-primary-soft);
  color: var(--na-primary);
}

.media-category-tree__scrollbar {
  min-width: 0;
  padding: 0 2px;
}

.media-category-tree__tree {
  --el-tree-node-hover-bg-color: transparent;

  min-width: 0;
  padding-bottom: 2px;
  background: transparent;
}

.media-category-tree :deep(.el-tree-node__content) {
  height: 38px;
  margin: 2px 0;
  padding-right: 4px;
  border: 1px solid transparent;
  border-radius: 9px;
  color: var(--na-foreground);
  transition: background-color 180ms cubic-bezier(.22, 1, .36, 1), border-color 180ms cubic-bezier(.22, 1, .36, 1), box-shadow 180ms cubic-bezier(.22, 1, .36, 1);
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
  border-color: color-mix(in srgb, var(--na-primary) 24%, var(--na-border));
  background: color-mix(in srgb, var(--na-primary) 10%, var(--na-card));
  box-shadow: inset 0 1px 0 rgb(255 255 255 / 20%);
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
  flex: 0 0 auto;
  color: var(--na-muted-foreground);
  font-size: 15px;
}

.media-category-tree__node.is-root .media-category-tree__node-icon,
.media-category-tree__node.is-active .media-category-tree__node-icon {
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

.media-category-tree__node.is-root .media-category-tree__label,
.media-category-tree__node.is-active .media-category-tree__label {
  color: var(--na-primary);
  font-weight: 650;
}

.media-category-tree__action.el-button {
  flex: 0 0 auto;
  margin-left: auto;
  opacity: .68;
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
  .media-category-tree :deep(.el-tree-node__content) {
    transition: none;
  }
}
</style>
