<template>
  <div>
    <el-table :data="tableData" stripe style="width: 100%" @row-click="openModule">
      <el-table-column prop="name" label="模块名称" show-overflow-tooltip width="180">
        <template #default="{ row }">
          <el-link type="primary" :underline="false" @click.stop="openModule(row)">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="resume" label="说明" show-overflow-tooltip />
      <el-table-column prop="status" label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.type" effect="light">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
  import { useRouter } from 'vue-router'

  const router = useRouter()

  const tableData = [
    { name: '资产大屏', resume: '资产数量、价值、分类、状态和位置排行统计。', status: '已启用', type: 'success', route: 'assetDashboard' },
    { name: '资产档案', resume: '资产全生命周期档案维护，支持照片、责任人和位置管理。', status: '已启用', type: 'success', route: 'assetInventory' },
    { name: '资产分类', resume: '维护资产分类和分类统计口径。', status: '已启用', type: 'success', route: 'assetCategories' },
    { name: '文档查看', resume: '查看已上传文档并维护在线编辑内容。', status: '已启用', type: 'success', route: 'documentViewer' }
  ]

  const openModule = (row) => {
    if (row.route) router.push({ name: row.route })
  }
</script>

<style scoped lang="scss"></style>
