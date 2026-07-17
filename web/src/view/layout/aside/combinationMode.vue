<template>
  <div class="h-full">
    <div
      v-if="mode === 'head'"
      class="na-head-nav"
    >
      <el-menu
        :default-active="routerStore.topActive"
        mode="horizontal"
        class="!border-r-0 border-b-0 w-full flex gap-1 items-center box-border h-[calc(100%-1px)]"
        unique-opened
        @select="(index, _, ele) => selectMenuItem(index, _, ele, true)"
      >
        <template v-for="item in routerStore.topMenu">
          <aside-component
            v-if="!item.hidden"
            :key="item.name"
            :router-info="item"
            mode="horizontal"
          />
        </template>
      </el-menu>
    </div>
    <div
      v-if="mode === 'normal'"
      class="na-sidebar"
      :class="{ 'has-sidebar-tip': !isCollapse && device !== 'mobile' }"
      :style="{
        width: layoutSideWidth + 'px'
      }"
    >
      <div class="na-sidebar-heading" :class="{ 'is-collapsed': isCollapse }">
        <span class="na-sidebar-heading__icon" aria-hidden="true">
          <el-icon><Menu /></el-icon>
        </span>
        <span v-if="!isCollapse" class="na-sidebar-heading__label">主菜单</span>
      </div>
      <el-scrollbar>
        <el-menu
          :collapse="isCollapse"
          :collapse-transition="false"
          :default-active="active"
          class="!border-r-0 w-full"
          @select="(index, _, ele) => selectMenuItem(index, _, ele, false)"
        >
          <template v-for="item in routerStore.leftMenu">
            <aside-component
              v-if="!item.hidden"
              :key="item.name"
              :router-info="item"
            />
          </template>
        </el-menu>
      </el-scrollbar>
      <sidebar-tip-card v-if="!isCollapse && device !== 'mobile'" />
      <button
        type="button"
        class="na-sidebar-toggle"
        :class="{ 'is-collapsed': isCollapse }"
        :aria-label="isCollapse ? '展开侧边栏' : '收起侧边栏'"
        @click="toggleCollapse"
      >
        <el-icon v-if="!isCollapse">
          <DArrowLeft />
        </el-icon>
        <el-icon v-else>
          <DArrowRight />
        </el-icon>
        <span v-if="!isCollapse">收起导航</span>
      </button>
    </div>
  </div>
</template>
<script setup>
  import AsideComponent from '@/view/layout/aside/asideComponent/index.vue'
  import SidebarTipCard from './SidebarTipCard.vue'
  import { ref, provide, watchEffect, computed } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { useRouterStore } from '@/pinia/modules/router'
  import { useAppStore } from '@/pinia'
  import { storeToRefs } from 'pinia'
  const appStore = useAppStore()
  const { device, config } = storeToRefs(appStore)

  defineOptions({
    name: 'GvaAside'
  })

  defineProps({
    mode: {
      type: String,
      default: 'normal'
    }
  })

  const route = useRoute()
  const router = useRouter()
  const routerStore = useRouterStore()
  const isCollapse = ref(false)
  const active = ref('')
  const layoutSideWidth = computed(() => {
    if (!isCollapse.value) {
      return config.value.layout_side_width
    } else {
      return config.value.layout_side_collapsed_width
    }
  })
  watchEffect(() => {
    active.value = route.meta.activeName || route.name
  })

  watchEffect(() => {
    if (device.value === 'mobile') {
      isCollapse.value = true
    } else {
      isCollapse.value = false
    }
  })

  provide('isCollapse', isCollapse)

  const selectMenuItem = (index, _, ele, top) => {
    const query = {}
    const params = {}
    routerStore.routeMap[index]?.parameters &&
      routerStore.routeMap[index]?.parameters.forEach((item) => {
        if (item.type === 'query') {
          query[item.key] = item.value
        } else {
          params[item.key] = item.value
        }
      })
    if (index === route.name) return
    if (index.indexOf('http://') > -1 || index.indexOf('https://') > -1) {
        window.open(index, '_blank')
        return
    }

      if (!top) {
        router.push({ name: index, query, params })
        return
      }
      const leftMenu = routerStore.setLeftMenu(index)
      if (!leftMenu) {
        router.push({ name: index, query, params })
        return;
      }
      const firstMenu = leftMenu.find((item) => !item.hidden && item.path.indexOf("http://") === -1 && item.path.indexOf("https://") === -1)
      router.push({ name: firstMenu.name, query, params })

  }

  const toggleCollapse = () => {
    isCollapse.value = !isCollapse.value
  }
</script>
