<template>
  <div class="h-full space-y-5">
    <div>
      <div class="mb-2 text-xs tracking-wide text-black/55 dark:text-white/55">常用入口</div>
      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
        <button
          v-for="(item, index) in shortcuts"
          :key="index"
          class="group flex w-full items-center gap-3 rounded-lg border border-black/10 bg-white/70 p-2.5 text-left transition-all duration-200 hover:border-[var(--el-color-primary)] hover:shadow-sm dark:border-white/10 dark:bg-white/[0.02]"
          type="button"
          @click="toPath(item)"
        >
          <span
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-slate-100 text-slate-700 transition-colors group-hover:bg-[var(--el-color-primary)] group-hover:text-white dark:bg-slate-800 dark:text-slate-200"
          >
            <el-icon><component :is="item.icon" /></el-icon>
          </span>
          <span class="min-w-0 text-sm text-black/75 dark:text-white/75">{{ item.title }}</span>
        </button>
      </div>
    </div>

    <div>
      <div class="mb-2 text-xs tracking-wide text-black/55 dark:text-white/55">技术文档</div>
      <div class="space-y-2">
        <button
          v-for="(item, index) in recentVisits"
          :key="index"
          class="flex w-full items-center justify-between rounded-lg border border-black/10 bg-white/70 px-3 py-2 text-left transition-all duration-200 hover:border-[var(--el-color-primary)] hover:shadow-sm dark:border-white/10 dark:bg-white/[0.02]"
          type="button"
          @click="openLink(item)"
        >
          <span class="flex items-center gap-2 text-sm text-black/75 dark:text-white/75">
            <el-icon><component :is="item.icon" /></el-icon>
            {{ item.title }}
          </span>
          <span class="text-xs text-black/45 dark:text-white/45">打开</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
  import {
    Menu,
    Link,
    User,
    Service,
    Document,
    Files,
    Memo,
    DataAnalysis
  } from '@element-plus/icons-vue'
  import { useRouter } from 'vue-router'

  const router = useRouter()

  const toPath = (item) => {
    router.push({ name: item.path })
  }

  const openLink = (item) => {
    window.open(item.path, '_blank', 'noopener,noreferrer')
  }

  const shortcuts = [
    { icon: DataAnalysis, title: '资产大屏', path: 'assetDashboard' },
    { icon: Files, title: '资产档案', path: 'assetInventory' },
    { icon: Document, title: '文档查看', path: 'documentViewer' },
    { icon: Menu, title: '菜单管理', path: 'menu' },
    { icon: Service, title: '角色管理', path: 'authority' },
    { icon: User, title: '用户管理', path: 'user' }
  ]

  const recentVisits = [
    { icon: Link, title: 'Vue 文档', path: 'https://cn.vuejs.org/' },
    { icon: Link, title: 'Element Plus', path: 'https://element-plus.org/zh-CN/' },
    { icon: Memo, title: 'Gin 文档', path: 'https://gin-gonic.com/zh-cn/docs/' }
  ]
</script>

<style scoped lang="scss"></style>
