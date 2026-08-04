<template>
  <main class="na-page na-page--list audit-page audit-page--error">
    <AppPageHeader
      title-id="error-log-title"
      title="错误日志"
      description="集中查看系统异常、处理进度和 AI 分析结果，便于快速恢复服务。"
    >
      <template #actions>
        <el-button :icon="Refresh" @click="getTableData">刷新</el-button>
        <LogClearButton
          log-name="错误日志"
          :count-request="getSysErrorList"
          :clear-request="clearSysError"
          @cleared="handleLogsCleared"
        />
      </template>
    </AppPageHeader>

    <section class="audit-overview" aria-label="错误日志概览">
      <div class="audit-overview__primary">
        <span>当前记录</span>
        <strong>{{ total }}</strong>
        <small>按当前筛选条件统计</small>
      </div>
      <div>
        <span>待处理</span>
        <strong>{{ pendingCount }}</strong>
        <small>当前页未完成记录</small>
      </div>
      <div>
        <span>已选择</span>
        <strong>{{ multipleSelection.length }}</strong>
        <small>可批量删除</small>
      </div>
    </section>

    <section class="na-panel audit-filter" aria-label="错误日志筛选">
      <el-form :model="searchInfo" label-position="top" @submit.prevent="onSubmit">
        <div class="audit-filter__grid" style="--audit-filter-columns: minmax(250px, 1.2fr) minmax(170px, .7fr) minmax(230px, 1fr) auto">
        <el-form-item prop="createdAtRange">
          <template #label>
            <span>
              创建日期
              <el-tooltip
                content="搜索范围是开始日期（包含）至结束日期（不包含）"
              >
                <el-icon><QuestionFilled /></el-icon>
              </el-tooltip>
            </span>
          </template>

          <el-date-picker
            v-model="searchInfo.createdAtRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
          />
        </el-form-item>

        <el-form-item label="错误来源" prop="form">
          <el-input v-model="searchInfo.form" clearable placeholder="输入模块或接口" />
        </el-form-item>

        <el-form-item label="错误内容" prop="info">
          <el-input v-model="searchInfo.info" clearable placeholder="输入错误关键词" />
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
          <strong>异常记录</strong>
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
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="ID"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" align="center" width="48" />

        <el-table-column
          sortable
          align="left"
          label="日期"
          prop="CreatedAt"
          width="168"
        >
          <template #default="scope">{{
            formatDate(scope.row.CreatedAt)
          }}</template>
        </el-table-column>

        <el-table-column
          align="left"
          label="错误来源"
          prop="form"
          min-width="150"
        />

        <el-table-column
          align="center"
          label="错误等级"
          prop="level"
          width="104"
        >
          <template #default="scope">
            <el-tag
              effect="light"
              :type="levelTagMap[scope.row.level] || 'info'"
            >
              {{ levelLabelMap[scope.row.level] || defaultLevelLabel }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          align="center"
          label="处理状态"
          prop="status"
          width="120"
        >
          <template #default="scope">
            <el-tag
              effect="light"
              :type="statusTagMap[scope.row.status] || 'info'"
            >
              {{ statusLabelMap[scope.row.status] || defaultStatusLabel }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          align="left"
          label="错误内容"
          prop="info"
          show-overflow-tooltip
          min-width="250"
        />

        <el-table-column
          align="left"
          label="解决方案"
          show-overflow-tooltip
          prop="solution"
          min-width="160"
        />

        <el-table-column
          align="center"
          label="操作"
          fixed="right"
          width="136"
        >
          <template #default="scope">
            <div class="error-row-actions">
              <el-tooltip v-if="scope.row.status !== '处理中'" content="使用 AI 分析解决方案" placement="top">
                <el-button
                  type="warning"
                  text
                  class="table-button"
                  @click="getSolution(scope.row.ID)"
                >
                  <el-icon><ai-gva /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="查看详情" placement="top">
                <el-button
                  :icon="InfoFilled"
                  type="primary"
                  text
                  class="table-button"
                  @click="getDetails(scope.row)"
                  aria-label="查看错误详情"
                />
              </el-tooltip>
              <el-tooltip content="删除记录" placement="top">
                <el-button
                  :icon="Delete"
                  type="danger"
                  text
                  class="table-button"
                  @click="deleteRow(scope.row)"
                  aria-label="删除错误日志"
                />
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="total > 0" class="na-pagination">
        <el-pagination
          :pager-count="5"
          layout="total, sizes, prev, pager, next"
          :current-page="searchInfo.page"
          :page-size="searchInfo.pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          @current-change="changePage"
          @size-change="changePageSize"
        />
      </div>
    </section>

    <el-drawer
      destroy-on-close
      :size="appStore.drawerSize"
      v-model="detailShow"
      :show-close="true"
      :before-close="closeDetailShow"
      title="查看"
    >
      <el-descriptions class="audit-detail" :column="2" border direction="vertical">
        <el-descriptions-item label="错误来源">
          {{ detailForm.form }}
        </el-descriptions-item>
        <el-descriptions-item label="错误等级">
          <el-tag
            effect="dark"
            :type="levelTagMap[detailForm.level] || 'info'"
          >
            {{ levelLabelMap[detailForm.level] || defaultLevelLabel }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="处理状态">
          <el-tag
            effect="light"
            :type="statusTagMap[detailForm.status] || 'info'"
          >
            {{ statusLabelMap[detailForm.status] || defaultStatusLabel }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="错误内容" :span="2">
          <pre class="whitespace-pre-wrap break-words">{{ detailForm.info }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="解决方案" :span="2">
          <pre class="whitespace-pre-wrap break-words">{{ detailForm.solution }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </main>
</template>

<script setup>
  import {
    clearSysError,
    deleteSysError,
    deleteSysErrorByIds,
    findSysError,
    getSysErrorList,
    getSysErrorSolution
  } from '@/api/system/sysError'

  import { formatDate } from '@/utils/format'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { computed, ref } from 'vue'
  import { Delete, Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
  import { useAppStore } from '@/pinia'
  import { usePagedList } from '@/hooks/usePagedList'
  import AppPageHeader from '@/components/page/AppPageHeader.vue'
  import LogClearButton from '@/components/logClearButton/index.vue'

  defineOptions({
    name: 'SysError'
  })

  const appStore = useAppStore()

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
    defaults: { page: 1, pageSize: 10, createdAtRange: undefined, form: '', info: '' },
    request: getSysErrorList
  })
  const pendingCount = computed(() => tableData.value.filter((item) => !['处理完成', '处理失败'].includes(item.status)).length)

  const getSolution = async (id) => {
    const confirmed = await ElMessageBox.confirm(
      '日志将通过 config.yaml 中配置的自有 AI 服务进行错误分析。是否确认进行 AI 处理？',
      '提示(Beta)',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).catch(() => false)
    if (!confirmed) return
    const res = await getSysErrorSolution({ id })
    if (res.code === 0) {
      ElMessage({ type: 'success', message: res.msg || '处理已提交，1分钟后完成' })
      getTableData()
    }
  }
  getTableData()

  // ============== 表格控制部分结束 ===============

  // 多选数据
  const multipleSelection = ref([])
  // 多选
  const handleSelectionChange = (val) => {
    multipleSelection.value = val
  }

  // 删除行
  const deleteRow = (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      deleteSysErrorFunc(row)
    })
  }

  // 多选删除
  const onDelete = async () => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const IDs = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: '请选择要删除的数据'
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map((item) => {
          IDs.push(item.ID)
        })
      const res = await deleteSysErrorByIds({ IDs })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        reloadAfterRemoval(IDs.length)
      }
    })
  }

  const handleLogsCleared = () => {
    searchInfo.page = 1
    multipleSelection.value = []
    getTableData()
  }

  // 删除行
  const deleteSysErrorFunc = async (row) => {
    const res = await deleteSysError({ ID: row.ID })
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: '删除成功'
      })
      reloadAfterRemoval()
    }
  }

  const detailForm = ref({})

  // 查看详情控制标记
  const detailShow = ref(false)

  // 打开详情弹窗
  const openDetailShow = () => {
    detailShow.value = true
  }

  // 打开详情
  const getDetails = async (row) => {
    // 打开弹窗
    const res = await findSysError({ ID: row.ID })
    if (res.code === 0) {
      detailForm.value = res.data
      openDetailShow()
    }
  }

  // 关闭详情弹窗
  const closeDetailShow = () => {
    detailShow.value = false
    detailForm.value = {}
  }

  const statusLabelMap = {
    未处理: '未处理',
    处理中: '处理中',
    处理完成: '处理完成',
    处理失败: '处理失败'
  }
  const statusTagMap = {
    未处理: 'info',
    处理中: 'warning',
    处理完成: 'success',
    处理失败: 'danger'
  }
  const defaultStatusLabel = '未处理'

  const levelLabelMap = {
    fatal: '致命错误',
    error: '一般错误'
  }
  const levelTagMap = {
    fatal: 'danger',
    error: 'warning'
  }
  const defaultLevelLabel = '一般错误'
</script>
