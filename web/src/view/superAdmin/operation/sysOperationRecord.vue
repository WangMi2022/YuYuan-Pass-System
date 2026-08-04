<template>
  <main class="na-page na-page--list audit-page">
    <AppPageHeader
      title-id="operation-history-title"
      title="操作历史"
      description="追踪系统接口调用、请求数据和响应结果，定位异常操作。"
    >
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="getTableData">刷新</el-button>
        <LogClearButton
          log-name="操作历史"
          :count-request="getSysOperationRecordList"
          :clear-request="clearSysOperationRecords"
          @cleared="handleLogsCleared"
        />
      </template>
    </AppPageHeader>

    <section class="audit-overview" aria-label="操作历史概览">
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

    <section class="na-panel audit-filter" aria-label="操作历史筛选">
      <el-form :model="searchInfo" label-position="top" @submit.prevent="onSubmit">
        <div class="audit-filter__grid">
        <el-form-item label="请求方法">
          <el-select v-model="searchInfo.method" clearable placeholder="全部方法">
            <el-option v-for="method in httpMethods" :key="method" :label="method" :value="method" />
          </el-select>
        </el-form-item>
        <el-form-item label="请求路径">
          <el-input v-model="searchInfo.path" clearable placeholder="输入接口路径" />
        </el-form-item>
        <el-form-item label="结果状态码">
          <el-input v-model="searchInfo.status" clearable placeholder="例如 200、404" />
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
          <strong>调用记录</strong>
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
        <el-table-column align="center" type="selection" width="48" />
        <el-table-column align="left" label="操作人" min-width="150">
          <template #default="scope">
            <div>
              {{ scope.row.user.userName }}({{ scope.row.user.nickName }})
            </div>
          </template>
        </el-table-column>
        <el-table-column align="left" label="时间" width="168">
          <template #default="scope">{{
            formatDate(scope.row.CreatedAt)
          }}</template>
        </el-table-column>
        <el-table-column align="center" label="状态" prop="status" width="88">
          <template #default="scope">
            <el-tag :type="statusTagType(scope.row.status)" effect="light">
              {{ scope.row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="请求 IP" prop="ip" width="128" />
        <el-table-column
          align="left"
          label="请求方法"
          prop="method"
          width="88"
        />
        <el-table-column
          align="left"
          label="请求路径"
          prop="path"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column align="center" label="请求" width="74">
          <template #default="scope">
            <PayloadPreviewPopover title="请求数据" :value="scope.row.body" />
          </template>
        </el-table-column>
        <el-table-column align="center" label="响应" width="74">
          <template #default="scope">
            <PayloadPreviewPopover title="响应数据" :value="scope.row.resp" />
          </template>
        </el-table-column>
        <el-table-column align="center" label="操作" width="72">
          <template #default="scope">
            <el-tooltip content="删除记录" placement="top">
              <el-button
                :icon="Delete"
                type="danger"
                text
                aria-label="删除操作记录"
                @click="deleteSysOperationRecordFunc(scope.row)"
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
    clearSysOperationRecords,
    deleteSysOperationRecord,
    getSysOperationRecordList,
    deleteSysOperationRecordByIds
  } from '@/api/sysOperationRecord' // 此处请自行替换地址
  import { formatDate } from '@/utils/format'
  import { computed, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { Delete, Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
  import { usePagedList } from '@/hooks/usePagedList'
  import AppPageHeader from '@/components/page/AppPageHeader.vue'
  import LogClearButton from '@/components/logClearButton/index.vue'
  import PayloadPreviewPopover from '@/components/payloadPreviewPopover/index.vue'

  defineOptions({
    name: 'SysOperationRecord'
  })

  const httpMethods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']

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
    defaults: { page: 1, pageSize: 10, method: '', path: '', status: '' },
    request: (params) => getSysOperationRecordList({
      ...params,
      status: params.status === '' ? null : params.status
    })
  })

  getTableData()

  const multipleSelection = ref([])
  const activeFilterCount = computed(() => [
    searchInfo.value.method,
    searchInfo.value.path,
    searchInfo.value.status
  ].filter((value) => value !== '' && value !== null && value !== undefined).length)
  const statusTagType = (status) => {
    const code = Number(status)
    if (code >= 500) return 'danger'
    if (code >= 400) return 'warning'
    if (code >= 300) return 'info'
    return code >= 200 ? 'success' : 'info'
  }
  const handleSelectionChange = (val) => {
    multipleSelection.value = val
  }
  const handleLogsCleared = () => {
    multipleSelection.value = []
    searchInfo.value.page = 1
    getTableData()
  }
  const onDelete = async () => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const ids = []
      multipleSelection.value &&
        multipleSelection.value.forEach((item) => {
          ids.push(item.ID)
        })
      const res = await deleteSysOperationRecordByIds({ ids })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        reloadAfterRemoval(ids.length)
      }
    })
  }
  const deleteSysOperationRecordFunc = async (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteSysOperationRecord({ ID: row.ID })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        reloadAfterRemoval()
      }
    })
  }
</script>
