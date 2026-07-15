<template>
  <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
    <button
      v-for="item in banners"
      :key="item.title"
      type="button"
      class="group overflow-hidden rounded-xl border border-black/10 bg-white/80 p-4 text-left shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-[var(--el-color-primary)] hover:shadow-md dark:border-white/10 dark:bg-white/[0.03]"
      @click="openCard(item)"
    >
      <div class="mb-4 h-1.5 w-16 rounded-full" :class="item.barClass" />
      <div class="text-base font-semibold text-black/85 dark:text-white/85">{{ item.title }}</div>
      <p class="mt-2 min-h-10 text-sm leading-5 text-black/55 dark:text-white/55">{{ item.desc }}</p>
      <div class="mt-4 text-xs font-medium text-[var(--el-color-primary)] opacity-80 group-hover:opacity-100">
        {{ item.action }} →
      </div>
    </button>
  </div>
</template>

<script setup>
  import { useRouter } from 'vue-router'

  const router = useRouter()

  const banners = [
    {
      title: '资产可视化大屏',
      desc: '汇总资产数量、价值、状态与分类结构，快速掌握资产运行情况。',
      action: '查看大屏',
      route: 'assetDashboard',
      barClass: 'bg-sky-500'
    },
    {
      title: '资产档案管理',
      desc: '统一维护资产编号、分类、位置、责任人、照片和价值信息。',
      action: '进入档案',
      route: 'assetInventory',
      barClass: 'bg-emerald-500'
    },
    {
      title: '文档管理',
      desc: '上传并在线维护 MD、Word、Excel、PDF 等项目资料。',
      action: '查看文档',
      route: 'documentViewer',
      barClass: 'bg-violet-500'
    }
  ]

  const openCard = (item) => {
    if (item.route) router.push({ name: item.route })
  }
</script>

<style scoped lang="scss"></style>
