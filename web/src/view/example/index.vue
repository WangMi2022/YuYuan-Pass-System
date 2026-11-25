<template>
  <div>
    <router-view v-slot="{ Component }">
      <transition mode="out-in" :name="appStore.config.transition_type || 'fade'">
        <keep-alive :include="routerStore.keepAliveRouters">
          <component :is="Component" />
        </keep-alive>
      </transition>
    </router-view>
  </div>
</template>

<script setup>
  import { useAppStore } from '@/pinia'
  import { useRouterStore } from '@/pinia/modules/router'
  const appStore = useAppStore()
  const routerStore = useRouterStore()
  defineOptions({
    name: 'Example'
  })
</script>
