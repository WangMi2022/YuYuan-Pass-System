<template>
  <div
    class="bg-background text-foreground w-screen h-screen"
  >
    <iframe
        v-if="reloadFlag"
        id="gva-base-load-dom"
        class="gva-body-h bg-background w-full border-t border-line"
        :src="url"
    ></iframe>
  </div>
</template>

<script setup>
  import useResponsive from '@/hooks/responsive'
  import { emitter } from '@/utils/bus.js'
  import { ref, onMounted, nextTick, reactive, watchEffect } from 'vue'
  import { useRouter, useRoute } from 'vue-router'
  import { useUserStore } from '@/pinia/modules/user'
  import { useAppStore } from '@/pinia'
  import { storeToRefs } from 'pinia'
  const appStore = useAppStore()
  const { isDark } = storeToRefs(appStore)


  defineOptions({
    name: 'GvaLayoutIframe'
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

  const url = route.query.url || 'about:blank'

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
