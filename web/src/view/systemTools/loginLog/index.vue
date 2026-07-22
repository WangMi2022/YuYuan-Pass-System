<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="用户名">
          <el-input v-model="searchInfo.username" placeholder="搜索用户名" />
        </el-form-item>
        <el-form-item label="状态">
             <el-select v-model="searchInfo.status" placeholder="请选择" clearable>
                 <el-option label="成功" :value="true" />
                 <el-option label="失败" :value="false" />
             </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button
          icon="delete"
          :disabled="!multipleSelection.length"
          @click="onDelete"
        >
          删除选中
        </el-button>
        <LogClearButton
          log-name="登录日志"
          :count-request="getLoginLogList"
          :clear-request="clearLoginLogs"
          @cleared="handleLogsCleared"
        />
      </div>
      <el-table
        ref="multipleTable"
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        tooltip-effect="dark"
        row-key="ID"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column align="left" label="ID" prop="ID" width="80" />
        <el-table-column align="left" label="用户名" prop="username" width="150" />
        <el-table-column align="left" label="登录IP" prop="ip" width="150" />
        <el-table-column align="left" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status ? 'success' : 'danger'">
              {{ scope.row.status ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="详情" show-overflow-tooltip>
             <template #default="scope">
                 {{ scope.row.status ? '登录成功' : scope.row.errorMessage }}
             </template>
        </el-table-column>
        <el-table-column align="left" label="浏览器/设备" prop="agent" show-overflow-tooltip />
        <el-table-column align="left" label="登录时间" width="180">
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column align="left" label="操作" width="120">
          <template #default="scope">
            <el-popover v-model:visible="scope.row.visible" placement="top" width="160">
              <p>确定要删除吗？</p>
              <div style="text-align: right; margin: 0">
                <el-button size="small" type="primary" link @click="scope.row.visible = false">取消</el-button>
                <el-button size="small" type="primary" @click="deleteRow(scope.row)">确定</el-button>
              </div>
              <template #reference>
                <el-button icon="delete" type="primary" link @click="scope.row.visible = true">删除</el-button>
              </template>
            </el-popover>
          </template>
        </el-table-column>
      </el-table>
      <div class="gva-pagination">
        <el-pagination
          :current-page="searchInfo.page"
          :page-size="searchInfo.pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="changePage"
          @size-change="changePageSize"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import {
  clearLoginLogs,
  getLoginLogList,
  deleteLoginLog,
  deleteLoginLogByIds
} from '@/api/sysLoginLog'
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatDate } from '@/utils/format'
import { usePagedList } from '@/hooks/usePagedList'
import LogClearButton from '@/components/logClearButton/index.vue'

const multipleSelection = ref([])

const {
  search: searchInfo,
  items: tableData,
  total,
  loading,
  load: getTableData,
  submit: onSubmit,
  reset: onReset,
  changePage,
  changePageSize,
  reloadAfterRemoval
} = usePagedList({
  defaults: { page: 1, pageSize: 10, username: '', status: undefined },
  request: getLoginLogList
})

const handleSelectionChange = (val) => {
  multipleSelection.value = val
}

const handleLogsCleared = () => {
  multipleSelection.value = []
  searchInfo.value.page = 1
  getTableData()
}

const deleteRow = async (row) => {
  row.visible = false
  const res = await deleteLoginLog(row)
  if (res.code === 0) {
    ElMessage({
      type: 'success',
      message: '删除成功'
    })
    reloadAfterRemoval()
  }
}

const onDelete = async() => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(async() => {
        const ids = multipleSelection.value.map(item => item.ID)
        const res = await deleteLoginLogByIds({ ids })
        if (res.code === 0) {
            ElMessage({
                type: 'success',
                message: '删除成功'
            })
            reloadAfterRemoval(ids.length)
        }
    })
}

// 首次加载
getTableData()
</script>

<style scoped>
</style>
