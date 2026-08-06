<template>
  <div
    class="na-app-shell"
    :class="{ 'has-bottom-navigation': device === 'mobile' }"
  >
    <el-watermark
      v-if="config.show_watermark"
      :font="font"
      :z-index="30"
      :gap="[180, 150]"
      class="!absolute !inset-0 !pointer-events-none"
      :content="userStore.userInfo.nickName"
    />
    <gva-header />
    <div class="na-app-body">
      <gva-aside
        v-if="
          device !== 'mobile' &&
          (config.side_mode === 'normal' || config.side_mode === 'sidebar')
        "
      />
      <gva-aside
        v-if="config.side_mode === 'combination' && device !== 'mobile'"
        mode="normal"
      />
      <div class="na-main-column">
        <gva-tabs v-if="config.showTabs" />
        <div class="na-page-scroll">
          <router-view v-if="reloadFlag" v-slot="{ Component, route }">
            <div
              id="gva-base-load-dom"
              class="na-page-host"
            >
              <transition
                mode="out-in"
                :name="route.meta.transitionType || config.transition_type"
              >
                <keep-alive :include="routerStore.keepAliveRouters">
                  <component :is="Component" :key="route.fullPath" />
                </keep-alive>
              </transition>
            </div>
          </router-view>
        </div>
      </div>
    </div>
    <bottom-navigation
      v-if="device === 'mobile'"
      :items="mobileMenuItems"
      :active="mobileActiveMenu"
      @select="selectMobileMenu"
    />
  </div>
</template>

<script setup>
  import GvaAside from '@/view/layout/aside/index.vue'
  import GvaHeader from '@/view/layout/header/index.vue'
  import BottomNavigation from '@/components/navigation/BottomNavigation.vue'
  import useResponsive from '@/hooks/responsive'
  import GvaTabs from './tabs/index.vue'
  import { emitter } from '@/utils/bus.js'
  import {
    computed,
    ref,
    onMounted,
    nextTick,
    reactive,
    watchEffect
  } from 'vue'
  import { useRouter, useRoute } from 'vue-router'
  import { useRouterStore } from '@/pinia/modules/router'
  import { useUserStore } from '@/pinia/modules/user'
  import { useAppStore } from '@/pinia'
  import { storeToRefs } from 'pinia'
  import '@/style/transition.scss'
  const appStore = useAppStore()
  const { config, isDark, device } = storeToRefs(appStore)

  defineOptions({
    name: 'GvaLayout'
  })

  useResponsive(true)
  const font = reactive({
    color: 'rgba(0, 0, 0, .15)'
  })

  watchEffect(() => {
    font.color = isDark.value ? 'rgba(255,255,255, .15)' : 'rgba(0, 0, 0, .15)'
  })

  const router = useRouter()
  const route = useRoute()
  const routerStore = useRouterStore()
  const mobileMenuItems = computed(() =>
    (routerStore.asyncRouters[0]?.children || []).filter((item) => !item.hidden)
  )
  const mobileActiveMenu = computed(
    () => route.meta.activeName || route.name || ''
  )

  const selectMobileMenu = (item) => {
    const index = item?.name
    if (!index || index === route.name) return
    if (/^https?:\/\//.test(index)) {
      window.open(index, '_blank', 'noopener,noreferrer')
      return
    }

    const query = {}
    const params = {}
    routerStore.routeMap[index]?.parameters?.forEach((parameter) => {
      const target = parameter.type === 'query' ? query : params
      target[parameter.key] = parameter.value
    })
    router.push({ name: index, query, params })
  }

  onMounted(() => {
    // 挂载一些通用的事件
    emitter.on('reload', reload)
    if (userStore.loadingInstance) {
      userStore.loadingInstance.close()
    }
  })

  const userStore = useUserStore()

  const reloadFlag = ref(true)
  let reloadTimer = null
  const reload = async () => {
    if (reloadTimer) {
      window.clearTimeout(reloadTimer)
    }
    reloadTimer = window.setTimeout(async () => {
      if (route.meta.keepAlive) {
        reloadFlag.value = false
        await nextTick()
        reloadFlag.value = true
      } else {
        const title = route.meta.title
        router.push({ name: 'Reload', params: { title } })
      }
    }, 400)
  }
</script>

<style lang="scss"></style>
