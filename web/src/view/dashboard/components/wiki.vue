<template>
  <div class="grid grid-cols-2 gap-2">
    <a
      v-for="item in wikis"
      :key="item.url"
      :href="item.url"
      class="text-sm text-black/70 dark:text-white/70 no-underline hover:text-[var(--el-color-primary)] dark:hover:text-white"
      target="_blank"
      rel="noopener noreferrer"
    >
      {{ item.title }}
    </a>
  </div>
</template>

<script setup>
  const wikis = [
    { title: 'Vue3', url: 'https://cn.vuejs.org/' },
    { title: 'Element Plus', url: 'https://element-plus.org/zh-CN/' },
    { title: 'Gin 文档', url: 'https://gin-gonic.com/zh-cn/docs/' },
    { title: 'ECharts', url: 'https://echarts.apache.org/zh/index.html' },
    { title: 'Vite', url: 'https://cn.vite.dev/' }
  ]
</script>

<style scoped lang="scss"></style>
