<template>
  <div>
    <el-table :data="tableData" stripe style="width: 100%">
      <el-table-column prop="ranking" label="序号" width="80" align="center" />
      <el-table-column prop="message" label="更新内容" show-overflow-tooltip />
      <el-table-column prop="author" label="负责人" width="140" />
      <el-table-column prop="date" label="时间" width="180" />
    </el-table>
  </div>
</template>

<script setup>
  const tableData = [
    { ranking: 1, message: '资产可视化大屏替换为首页仪表盘。', author: '资产管理组', date: '2026-07-14 20:10:00' },
    { ranking: 2, message: '文档管理支持上传、预览、下载与在线编辑。', author: '平台开发组', date: '2026-07-14 19:35:00' },
    { ranking: 3, message: 'Excel 导入导出文件统一接入 OSS 存储。', author: '平台开发组', date: '2026-07-14 18:20:00' },
    { ranking: 4, message: '分页导航增加一键清理，便于快速整理标签页。', author: '前端开发组', date: '2026-07-14 17:40:00' },
    { ranking: 5, message: '生成资产测试数据，覆盖多类资产场景。', author: '资产管理组', date: '2026-07-14 16:30:00' }
  ]
</script>

<style scoped lang="scss"></style>
