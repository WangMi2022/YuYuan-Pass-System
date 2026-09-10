<!-- 此路由可作为父类路由通用路由页面使用 如需自定义父类路由页面 请参考 @/view/superAdmin/index.vue -->
<template>
  <div>
    <router-view v-slot="{ Component }">
      <transition mode="out-in" :name="appStore.config.transition_type || 'fade'">
        <keep-alive :include="routerStore.keepAliveRouters">
          <div>
            <component :is="Component" />
          </div>
        </keep-alive>
      </transition>
    </router-view>
  </div>
</template>

<script setup>
  defineOptions({
    name: 'RouterHolder'
  })
  import { useAppStore } from '@/pinia'
  import { useRouterStore } from '@/pinia/modules/router'
  const appStore = useAppStore()
  const routerStore = useRouterStore()
</script>
