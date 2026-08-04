<template>
  <main class="na-page na-page--list audit-page audit-page--login">
    <AppPageHeader
      title-id="login-log-title"
      title="登录日志"
      description="查看账户登录结果、终端信息和失败原因，及时发现异常访问。"
    >
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="getTableData">刷新</el-button>
        <LogClearButton
          log-name="登录日志"
          :count-request="getLoginLogList"
          :clear-request="clearLoginLogs"
          @cleared="handleLogsCleared"
        />
      </template>
    </AppPageHeader>

    <section class="audit-overview" aria-label="登录日志概览">
      <div class="audit-overview__primary">
        <span>当前记录</span>
        <strong>{{ total }}</strong>
        <small>按当前筛选条件统计</small>
      </div>
      <div>
        <span>已选择</span>
        <strong>{{ multipleSelection.length }}</strong>
        <small>可批量删除</small>
      </div>
      <div>
        <span>筛选状态</span>
        <strong>{{ activeFilterCount ? `${activeFilterCount} 项` : '全部' }}</strong>
        <small>{{ activeFilterCount ? '已应用查询条件' : '未限制查询范围' }}</small>
      </div>
    </section>

    <section class="na-panel audit-filter" aria-label="登录日志筛选">
      <el-form :model="searchInfo" label-position="top" @submit.prevent="onSubmit">
        <div class="audit-filter__grid" style="--audit-filter-columns: minmax(220px, 1.2fr) minmax(160px, .7fr) auto">
          <el-form-item label="用户名">
            <el-input v-model="searchInfo.username" clearable placeholder="输入用户名" />
          </el-form-item>
          <el-form-item label="登录状态">
            <el-select v-model="searchInfo.status" placeholder="全部状态" clearable>
              <el-option label="成功" :value="true" />
              <el-option label="失败" :value="false" />
            </el-select>
          </el-form-item>
          <div class="audit-filter__actions">
            <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
            <el-button native-type="submit" type="primary" :icon="Search">查询</el-button>
          </div>
        </div>
      </el-form>
    </section>

    <section class="na-panel audit-table-panel">
      <header class="na-panel-header audit-table-toolbar">
        <div>
          <strong>访问记录</strong>
          <span>共 {{ total }} 条</span>
        </div>
        <el-button
          type="danger"
          plain
          :icon="Delete"
          :disabled="!multipleSelection.length"
          @click="onDelete"
        >
          删除选中<span v-if="multipleSelection.length">（{{ multipleSelection.length }}）</span>
        </el-button>
      </header>
      <el-table
        ref="multipleTable"
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        tooltip-effect="dark"
        row-key="ID"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" align="center" width="48" />
        <el-table-column align="left" label="用户名" prop="username" min-width="140" />
        <el-table-column align="left" label="登录 IP" prop="ip" width="140" />
        <el-table-column align="center" label="状态" width="90">
          <template #default="scope">
            <el-tag :type="scope.row.status ? 'success' : 'danger'">
              {{ scope.row.status ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="结果详情" min-width="180" show-overflow-tooltip>
             <template #default="scope">
                 {{ scope.row.status ? '登录成功' : scope.row.errorMessage }}
             </template>
        </el-table-column>
        <el-table-column align="left" label="浏览器 / 设备" prop="agent" min-width="220" show-overflow-tooltip />
        <el-table-column align="left" label="登录时间" width="180">
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column align="center" label="操作" width="72">
          <template #default="scope">
            <el-tooltip content="删除记录" placement="top">
              <el-button
                :icon="Delete"
                type="danger"
                text
                aria-label="删除登录日志"
                @click="deleteRow(scope.row)"
              />
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="total > 0" class="na-pagination">
        <el-pagination
          :current-page="searchInfo.page"
          :page-size="searchInfo.pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          :pager-count="5"
          layout="total, sizes, prev, pager, next"
          @current-change="changePage"
          @size-change="changePageSize"
        />
      </div>
    </section>
  </main>
</template>

<script setup>
import {
  clearLoginLogs,
  getLoginLogList,
  deleteLoginLog,
  deleteLoginLogByIds
} from '@/api/sysLoginLog'
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatDate } from '@/utils/format'
import { usePagedList } from '@/hooks/usePagedList'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
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

const activeFilterCount = computed(() => [
  searchInfo.value.username,
  searchInfo.value.status
].filter((value) => value !== '' && value !== null && value !== undefined).length)

const handleSelectionChange = (val) => {
  multipleSelection.value = val
}

const handleLogsCleared = () => {
  multipleSelection.value = []
  searchInfo.value.page = 1
  getTableData()
}

const deleteRow = async (row) => {
  const confirmed = await ElMessageBox.confirm('确定删除这条登录日志吗？', '删除登录日志', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning'
  }).catch(() => false)
  if (!confirmed) return
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
