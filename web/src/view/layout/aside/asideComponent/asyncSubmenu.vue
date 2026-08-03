<template>
  <el-sub-menu
    ref="subMenu"
    :index="routerInfo.name"
    :style="menuStyleVars"
    class="gva-sub-menu relative"
  >
    <template #title>
      <el-icon v-if="routerInfo.meta.icon">
        <component :is="routerInfo.meta.icon" />
      </el-icon>
      <span>{{ routerInfo.meta.title }}</span>
    </template>
    <slot />
  </el-sub-menu>
</template>

<script setup>
  import { computed } from 'vue'
  import { useAppStore } from '@/pinia'
  import { storeToRefs } from 'pinia'
  const appStore = useAppStore()
  const { config } = storeToRefs(appStore)

  defineOptions({
    name: 'AsyncSubmenu'
  })

  defineProps({
    routerInfo: {
      default: function () {
        return null
      },
      type: Object
    }
  })

  const sideHeight = computed(() => {
    return config.value.layout_side_item_height + 'px'
  })

  const menuStyleVars = computed(() => ({
    '--gva-side-height': sideHeight.value
  }))
</script>

<style lang="scss">
  .gva-sub-menu {
    .el-sub-menu__title {
      height: var(--gva-side-height) !important;
    }
  }
</style>
