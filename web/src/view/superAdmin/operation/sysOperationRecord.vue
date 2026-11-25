<template>
  <main class="na-page na-page--list audit-page audit-page--operation">
    <AppPageHeader
      title-id="operation-history-title"
      title="操作历史"
      description="系统级接口调用审计中心，追踪接口请求、请求参数、响应报文及耗时性能，定位异常操作与安全风险。"
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

    <!-- 1. 顶部四大核心运营与审计态势看板 -->
    <section class="audit-overview" aria-label="操作历史态势概览">
      <div class="audit-overview__primary">
        <div class="overview-header">
          <span>总调用记录</span>
          <el-icon class="overview-icon"><Operation /></el-icon>
        </div>
        <strong>{{ total }}</strong>
        <small>{{ activeFilterCount ? `当前筛选结果（共 ${total} 条）` : '全量操作日志审计' }}</small>
      </div>

      <div class="overview-card--success">
        <div class="overview-header">
          <span>调用正常率</span>
          <el-icon class="overview-icon text-success"><Check /></el-icon>
        </div>
        <strong class="text-success">{{ successRate }}%</strong>
        <small>当前展示记录成功占比 (2xx/3xx)</small>
      </div>

      <div
        class="overview-card--warning"
        :class="{ 'is-clickable': exceptionCount > 0 }"
        @click="filterExceptions"
      >
        <div class="overview-header">
          <span>异常与告警</span>
          <el-icon class="overview-icon text-warning"><WarningFilled /></el-icon>
        </div>
        <strong :class="exceptionCount > 0 ? 'text-danger' : 'text-muted'">{{ exceptionCount }}</strong>
        <small>{{ exceptionCount > 0 ? '点击快速定位 4xx/5xx 异常' : '未检测到异常状态响应' }}</small>
      </div>

      <div class="overview-card--selection">
        <div class="overview-header">
          <span>选区管理</span>
          <el-icon class="overview-icon"><Delete /></el-icon>
        </div>
        <strong>{{ multipleSelection.length }} <small class="text-sub">项</small></strong>
        <small>{{ multipleSelection.length ? '已勾选，支持批量安全删除' : '勾选表格行可批量清理' }}</small>
      </div>
    </section>

    <!-- 2. 多维度专业审计筛选工具栏 -->
    <section class="na-panel audit-filter" aria-label="操作历史筛选">
      <!-- 快捷预设时段切换 -->
      <div class="audit-quick-presets">
        <span class="preset-label">快捷时间：</span>
        <button
          v-for="preset in timePresets"
          :key="preset.value"
          type="button"
          class="preset-chip"
          :class="{ 'is-active': activeTimePreset === preset.value }"
          @click="setTimePreset(preset.value)"
        >
          {{ preset.label }}
        </button>
      </div>

      <el-form :model="searchInfo" label-position="top" @submit.prevent="onSubmit">
        <div class="audit-filter__grid" style="--audit-filter-columns: minmax(260px, 1.4fr) minmax(130px, .7fr) minmax(140px, .8fr) minmax(200px, 1.2fr) auto">
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              @change="onDateRangeChange"
            />
          </el-form-item>

          <el-form-item label="请求方法">
            <el-select v-model="searchInfo.method" clearable placeholder="全部方法">
              <el-option v-for="method in httpMethods" :key="method" :label="method" :value="method">
                <span class="method-option-item">
                  <i :class="`dot-${method.toLowerCase()}`" />
                  <span>{{ method }}</span>
                </span>
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item label="结果状态码">
            <el-select
              v-model="searchInfo.status"
              clearable
              filterable
              allow-create
              placeholder="全部状态"
            >
              <el-option label="200 - 成功" value="200" />
              <el-option label="400 - 请求错误" value="400" />
              <el-option label="401 - 未登录/凭据无效" value="401" />
              <el-option label="403 - 无权限访问" value="403" />
              <el-option label="404 - 资源未找到" value="404" />
              <el-option label="500 - 服务器内部错误" value="500" />
            </el-select>
          </el-form-item>

          <el-form-item label="请求路径">
            <el-input
              v-model="searchInfo.path"
              clearable
              placeholder="模糊匹配接口路径，如 /user"
              :prefix-icon="Link"
            />
          </el-form-item>

          <div class="audit-filter__actions">
            <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
            <el-button native-type="submit" type="primary" :icon="Search" :loading="loading">查询</el-button>
          </div>
        </div>
      </el-form>
    </section>

    <!-- 3. 数据表格面板 -->
    <section class="na-panel audit-table-panel">
      <header class="na-panel-header audit-table-toolbar">
        <div class="toolbar-left">
          <strong>接口调用审计流水</strong>
          <span class="toolbar-count">共 {{ total }} 条记录</span>
          <span v-if="activeFilterCount" class="toolbar-filter-badge">已过滤</span>
        </div>
        <div class="toolbar-right">
          <el-button
            type="danger"
            plain
            :icon="Delete"
            :disabled="!multipleSelection.length || bulkDeleting"
            :loading="bulkDeleting"
            @click="onDelete"
          >
            删除选中<span v-if="multipleSelection.length">（{{ multipleSelection.length }}）</span>
          </el-button>
        </div>
      </header>

      <el-table
        ref="multipleTable"
        v-loading="loading && !loaded"
        :data="tableData"
        style="width: 100%"
        tooltip-effect="dark"
        row-key="ID"
        class="audit-custom-table"
        @selection-change="handleSelectionChange"
      >
        <el-table-column align="center" type="selection" width="44" />

        <el-table-column align="left" label="操作人员" min-width="170">
          <template #default="scope">
            <div class="operator-cell">
              <div class="operator-avatar" aria-hidden="true">
                {{ getAvatarInitials(scope.row.user?.nickName || scope.row.user?.userName) }}
              </div>
              <div class="operator-info">
                <strong>{{ scope.row.user?.userName || '系统操作' }}</strong>
                <small v-if="scope.row.user?.nickName">{{ scope.row.user.nickName }}</small>
                <small v-else-if="scope.row.user_id" class="text-muted">UID: {{ scope.row.user_id }}</small>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column align="center" label="方法" prop="method" width="90">
          <template #default="scope">
            <span :class="['method-badge', `method-${scope.row.method?.toLowerCase()}`]">
              {{ scope.row.method || 'GET' }}
            </span>
          </template>
        </el-table-column>

        <el-table-column align="center" label="状态" prop="status" width="95">
          <template #default="scope">
            <el-tooltip :content="getStatusDesc(scope.row.status)" placement="top">
              <el-tag :type="statusTagType(scope.row.status)" effect="light" class="status-tag">
                {{ scope.row.status }}
              </el-tag>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column align="center" label="耗时" width="95">
          <template #default="scope">
            <span :class="['latency-chip', getLatencyTone(scope.row.latency)]">
              {{ formatLatency(scope.row.latency) }}
            </span>
          </template>
        </el-table-column>

        <el-table-column align="left" label="请求路径" prop="path" min-width="240" show-overflow-tooltip>
          <template #default="scope">
            <div class="path-cell">
              <span class="path-text" :title="scope.row.path">{{ scope.row.path }}</span>
              <el-button
                class="copy-path-btn"
                link
                :icon="CopyDocument"
                aria-label="复制接口路径"
                @click.stop="copyToClipboard(scope.row.path, '接口路径')"
              />
            </div>
          </template>
        </el-table-column>

        <el-table-column align="left" label="客户端环境" min-width="170" show-overflow-tooltip>
          <template #default="scope">
            <div class="client-cell">
              <span class="ip-tag">{{ scope.row.ip || '—' }}</span>
              <span class="agent-tag" :title="scope.row.agent">
                {{ parseUserAgent(scope.row.agent).summary }}
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column align="left" label="操作时间" width="168">
          <template #default="scope">
            <span class="time-cell" :title="scope.row.CreatedAt">{{ formatDate(scope.row.CreatedAt) }}</span>
          </template>
        </el-table-column>

        <el-table-column align="center" label="报文" width="105">
          <template #default="scope">
            <div class="payload-actions">
              <PayloadPreviewPopover title="请求" :value="scope.row.body" />
              <PayloadPreviewPopover title="响应" :value="scope.row.resp" />
            </div>
          </template>
        </el-table-column>

        <el-table-column align="center" label="操作" width="95" fixed="right">
          <template #default="scope">
            <div class="table-actions">
              <el-tooltip content="查看结构化详情" placement="top">
                <el-button
                  :icon="View"
                  type="primary"
                  text
                  aria-label="查看操作审计详情"
                  @click="openDetail(scope.row)"
                />
              </el-tooltip>
              <el-tooltip content="删除记录" placement="top">
                <el-button
                  :icon="Delete"
                  type="danger"
                  text
                  aria-label="删除操作记录"
                  @click="deleteSysOperationRecordFunc(scope.row)"
                />
              </el-tooltip>
            </div>
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

    <!-- 4. 结构化操作审计详情抽屉 -->
    <el-drawer
      v-model="detailVisible"
      title="操作审计详情"
      direction="rtl"
      size="640px"
      destroy-on-close
    >
      <div v-if="currentRecord" class="audit-drawer-content">
        <!-- 头部状态指示区 -->
        <div class="drawer-headline">
          <div class="drawer-headline__tags">
            <span :class="['method-badge', `method-${currentRecord.method?.toLowerCase()}`]">
              {{ currentRecord.method }}
            </span>
            <el-tag :type="statusTagType(currentRecord.status)" effect="light">
              HTTP {{ currentRecord.status }} · {{ getStatusDesc(currentRecord.status) }}
            </el-tag>
            <span :class="['latency-chip', getLatencyTone(currentRecord.latency)]">
              耗时 {{ formatLatency(currentRecord.latency) }}
            </span>
          </div>
          <div class="drawer-headline__id">ID: #{{ currentRecord.ID }}</div>
        </div>

        <!-- 关键元数据面板 -->
        <div class="drawer-section">
          <h3><el-icon><InfoFilled /></el-icon>基础元数据</h3>
          <el-descriptions :column="2" border size="small" class="detail-descriptions">
            <el-descriptions-item label="操作用户">
              <span class="user-desc">
                <strong>{{ currentRecord.user?.userName || '系统用户' }}</strong>
                <small v-if="currentRecord.user?.nickName">（{{ currentRecord.user.nickName }}）</small>
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="操作时间">
              {{ formatDate(currentRecord.CreatedAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="客户端 IP">
              <code>{{ currentRecord.ip || '—' }}</code>
            </el-descriptions-item>
            <el-descriptions-item label="客户端环境">
              <span>{{ parseUserAgent(currentRecord.agent).full }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="请求路径" :span="2">
              <div class="drawer-path-box">
                <code>{{ currentRecord.path }}</code>
                <el-button
                  link
                  type="primary"
                  :icon="CopyDocument"
                  @click="copyToClipboard(currentRecord.path, '接口完整路径')"
                >复制</el-button>
              </div>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 异常报错追踪（若有） -->
        <div v-if="currentRecord.error_message" class="drawer-section error-section">
          <h3><el-icon class="text-danger"><WarningFilled /></el-icon>异常错误堆栈</h3>
          <pre class="code-block error-block"><code>{{ currentRecord.error_message }}</code></pre>
        </div>

        <!-- 请求 Body -->
        <div class="drawer-section">
          <div class="section-title-row">
            <h3><el-icon><Document /></el-icon>请求报文 (Request Body)</h3>
            <el-button
              v-if="currentRecord.body"
              link
              type="primary"
              :icon="CopyDocument"
              @click="copyToClipboard(currentRecord.body, '请求数据')"
            >
              复制
            </el-button>
          </div>
          <div v-if="currentRecord.body" class="code-container">
            <pre class="code-block"><code>{{ formatJsonString(currentRecord.body) }}</code></pre>
          </div>
          <p v-else class="empty-hint">（无请求体内容）</p>
        </div>

        <!-- 响应 Body -->
        <div class="drawer-section">
          <div class="section-title-row">
            <h3><el-icon><Document /></el-icon>响应报文 (Response Body)</h3>
            <el-button
              v-if="currentRecord.resp"
              link
              type="primary"
              :icon="CopyDocument"
              @click="copyToClipboard(currentRecord.resp, '响应数据')"
            >
              复制
            </el-button>
          </div>
          <div v-if="currentRecord.resp" class="code-container">
            <pre class="code-block"><code>{{ formatJsonString(currentRecord.resp) }}</code></pre>
          </div>
          <p v-else class="empty-hint">（无响应体数据）</p>
        </div>
      </div>

      <template #footer>
        <div class="drawer-footer">
          <el-button
            type="danger"
            plain
            :icon="Delete"
            @click="deleteSysOperationRecordFunc(currentRecord)"
          >
            删除此记录
          </el-button>
          <el-button @click="detailVisible = false">关闭</el-button>
        </div>
      </template>
    </el-drawer>
  </main>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Check,
  CopyDocument,
  Delete,
  Document,
  InfoFilled,
  Link,
  Operation,
  Refresh,
  RefreshLeft,
  Search,
  View,
  WarningFilled
} from '@element-plus/icons-vue'
import {
  clearSysOperationRecords,
  deleteSysOperationRecord,
  deleteSysOperationRecordByIds,
  getSysOperationRecordList
} from '@/api/sysOperationRecord'
import { formatDate } from '@/utils/format'
import { usePagedList } from '@/hooks/usePagedList'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import LogClearButton from '@/components/logClearButton/index.vue'
import PayloadPreviewPopover from '@/components/payloadPreviewPopover/index.vue'

defineOptions({
  name: 'SysOperationRecord'
})

const httpMethods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
const timePresets = [
  { label: '全部', value: -1 },
  { label: '今日', value: 0 },
  { label: '近3天', value: 3 },
  { label: '近7天', value: 7 },
  { label: '近30天', value: 30 }
]

const activeTimePreset = ref(-1)
const dateRange = ref([])
const multipleSelection = ref([])
const bulkDeleting = ref(false)

// 详情抽屉
const detailVisible = ref(false)
const currentRecord = ref(null)

const {
  search: searchInfo,
  items: tableData,
  total,
  loading,
  loaded,
  load: getTableData,
  submit: onSubmit,
  reset: onResetPagedList,
  changePage,
  changePageSize,
  reloadAfterRemoval
} = usePagedList({
  defaults: {
    page: 1,
    pageSize: 10,
    method: '',
    path: '',
    status: '',
    startTime: '',
    endTime: ''
  },
  request: (params) => {
    const cleanParams = {
      ...params,
      status: params.status === '' ? null : params.status
    }
    if (!cleanParams.startTime) delete cleanParams.startTime
    if (!cleanParams.endTime) delete cleanParams.endTime
    return getSysOperationRecordList(cleanParams)
  }
})

getTableData()

// 过滤项计数
const activeFilterCount = computed(() => [
  searchInfo.method,
  searchInfo.path,
  searchInfo.status,
  searchInfo.startTime,
  searchInfo.endTime
].filter((v) => v !== '' && v !== null && v !== undefined).length)

// 成功率统计
const successRate = computed(() => {
  if (!tableData.value || !tableData.value.length) return 100
  const successes = tableData.value.filter((item) => Number(item.status) < 400).length
  return Math.round((successes / tableData.value.length) * 100)
})

// 异常数统计
const exceptionCount = computed(() => {
  if (!tableData.value || !tableData.value.length) return 0
  return tableData.value.filter((item) => Number(item.status) >= 400).length
})

// 时间选择器变更
function onDateRangeChange(val) {
  if (val && val.length === 2) {
    searchInfo.startTime = val[0]
    searchInfo.endTime = val[1]
    activeTimePreset.value = null
  } else {
    searchInfo.startTime = ''
    searchInfo.endTime = ''
    activeTimePreset.value = -1
  }
}

// 快捷时间范围
function setTimePreset(days) {
  activeTimePreset.value = days
  if (days === -1) {
    dateRange.value = []
    searchInfo.startTime = ''
    searchInfo.endTime = ''
  } else if (days === 0) {
    const now = new Date()
    const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0)
    const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
    dateRange.value = [start.toISOString(), end.toISOString()]
    searchInfo.startTime = dateRange.value[0]
    searchInfo.endTime = dateRange.value[1]
  } else {
    const end = new Date()
    const start = new Date()
    start.setDate(start.getDate() - days)
    dateRange.value = [start.toISOString(), end.toISOString()]
    searchInfo.startTime = dateRange.value[0]
    searchInfo.endTime = dateRange.value[1]
  }
  onSubmit()
}

// 一键定位异常
function filterExceptions() {
  searchInfo.status = '400'
  onSubmit()
}

// 重置
function onReset() {
  dateRange.value = []
  activeTimePreset.value = -1
  onResetPagedList()
}

// 状态码标签
function statusTagType(status) {
  const code = Number(status)
  if (code >= 500) return 'danger'
  if (code >= 400) return 'warning'
  if (code >= 300) return 'info'
  return code >= 200 ? 'success' : 'info'
}

function getStatusDesc(status) {
  const code = Number(status)
  const map = {
    200: 'OK · 请求成功',
    201: 'Created · 创建成功',
    204: 'No Content · 无内容',
    400: 'Bad Request · 客户端参数错误',
    401: 'Unauthorized · 凭据无效/未登录',
    403: 'Forbidden · 权限不足，禁止访问',
    404: 'Not Found · 接口路径不存在',
    405: 'Method Not Allowed · 请求方法不支持',
    500: 'Internal Server Error · 服务器内部故障',
    502: 'Bad Gateway · 网关异常',
    503: 'Service Unavailable · 服务暂时不可用'
  }
  return map[code] || `HTTP ${code}`
}

// 耗时格式化与等级
function formatLatency(val) {
  if (val === undefined || val === null || val === '') return '—'
  if (typeof val === 'string') {
    if (val.endsWith('s') || val.endsWith('ms') || val.endsWith('µs') || val.endsWith('ns')) {
      return val
    }
    const num = Number(val)
    if (!Number.isNaN(num)) return formatNs(num)
    return val
  }
  if (typeof val === 'number') {
    return formatNs(val)
  }
  return String(val)
}

function formatNs(ns) {
  if (ns < 1000) return `${ns}ns`
  if (ns < 1000000) return `${(ns / 1000).toFixed(1)}µs`
  if (ns < 1000000000) return `${(ns / 1000000).toFixed(0)}ms`
  return `${(ns / 1000000000).toFixed(2)}s`
}

function getLatencyTone(val) {
  const ns = Number(val)
  if (!Number.isNaN(ns) && ns > 0) {
    if (ns < 150000000) return 'latency-fast' // < 150ms
    if (ns < 600000000) return 'latency-normal' // < 600ms
    if (ns < 1500000000) return 'latency-slow' // < 1.5s
    return 'latency-danger' // >= 1.5s
  }
  return 'latency-fast'
}

// 解析 User-Agent
function parseUserAgent(agent) {
  if (!agent) return { summary: '未知客户端', full: '未知客户端' }
  let os = 'OS'
  if (agent.includes('Windows')) os = 'Windows'
  else if (agent.includes('Macintosh') || agent.includes('Mac OS')) os = 'macOS'
  else if (agent.includes('Linux')) os = 'Linux'
  else if (agent.includes('Android')) os = 'Android'
  else if (agent.includes('iPhone') || agent.includes('iPad')) os = 'iOS'

  let browser = 'Client'
  if (agent.includes('Edg/')) browser = 'Edge'
  else if (agent.includes('Chrome/')) browser = 'Chrome'
  else if (agent.includes('Firefox/')) browser = 'Firefox'
  else if (agent.includes('Safari/') && !agent.includes('Chrome/')) browser = 'Safari'
  else if (agent.includes('Postman')) browser = 'Postman'
  else if (agent.includes('curl/')) browser = 'curl'

  return {
    summary: `${browser} · ${os}`,
    full: `${browser} on ${os} (${agent})`
  }
}

// 头像文字
function getAvatarInitials(name) {
  if (!name) return '系'
  return name.trim().charAt(0).toUpperCase()
}

// 格式化 JSON
function formatJsonString(str) {
  if (!str) return ''
  try {
    const obj = JSON.parse(str)
    return JSON.stringify(obj, null, 2)
  } catch {
    return str
  }
}

// 复制到剪贴板
async function copyToClipboard(text, label = '内容') {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(`${label}已复制到剪贴板`)
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    ElMessage.success(`${label}已复制到剪贴板`)
  }
}

// 查看详情
function openDetail(row) {
  currentRecord.value = row
  detailVisible.value = true
}

// 多选
function handleSelectionChange(val) {
  multipleSelection.value = val
}

function handleLogsCleared() {
  multipleSelection.value = []
  searchInfo.page = 1
  getTableData()
}

// 批量删除
async function onDelete() {
  if (!multipleSelection.value.length) return
  ElMessageBox.confirm(
    `确定要永久删除已选中的 ${multipleSelection.value.length} 条操作历史记录吗？此操作无法撤销。`,
    '安全删除确认',
    {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    bulkDeleting.value = true
    try {
      const ids = multipleSelection.value.map((item) => item.ID)
      const res = await deleteSysOperationRecordByIds({ ids })
      if (res.code === 0) {
        ElMessage.success('批量删除成功')
        reloadAfterRemoval(ids.length)
      }
    } finally {
      bulkDeleting.value = false
    }
  })
}

// 单行删除
async function deleteSysOperationRecordFunc(row) {
  if (!row?.ID) return
  ElMessageBox.confirm('确定要删除该条操作审计记录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    const res = await deleteSysOperationRecord({ ID: row.ID })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      if (detailVisible.value && currentRecord.value?.ID === row.ID) {
        detailVisible.value = false
      }
      reloadAfterRemoval()
    }
  })
}
</script>

<style scoped lang="scss">
.audit-overview {
  display: grid;
  grid-template-columns: minmax(220px, 1.3fr) repeat(3, minmax(160px, 1fr));
  gap: 14px;
  margin-bottom: 16px;
}

.audit-overview > div {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  min-height: 106px;
  padding: 16px 20px;
  border: 1px solid var(--na-border);
  border-radius: 12px;
  background: var(--na-card);
  box-shadow: var(--na-shadow-sm);
  transition: transform 0.2s ease, box-shadow 0.2s ease;

  &:hover {
    box-shadow: 0 4px 16px -2px rgba(0, 0, 0, 0.06);
  }
}

.overview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--na-muted-foreground);
  font-size: 12px;
  font-weight: 500;
}

.overview-icon {
  font-size: 18px;
  color: var(--na-muted-foreground);
}

.audit-overview strong {
  font-size: 26px;
  font-weight: 700;
  color: var(--na-foreground);
  font-variant-numeric: tabular-nums;
  line-height: 1.1;

  .text-sub {
    font-size: 13px;
    font-weight: 400;
    color: var(--na-muted-foreground);
    margin-left: 4px;
  }
}

.audit-overview small {
  font-size: 11px;
  color: var(--na-muted-foreground);
}

.overview-card--warning.is-clickable {
  cursor: pointer;
  &:hover {
    border-color: var(--na-warning);
    transform: translateY(-2px);
  }
}

.text-success { color: var(--na-success) !important; }
.text-warning { color: var(--na-warning) !important; }
.text-danger { color: var(--na-danger) !important; }
.text-muted { color: var(--na-muted-foreground) !important; }

/* 快捷预设时段切换 */
.audit-quick-presets {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 12px;
  margin-bottom: 12px;
  border-bottom: 1px dashed var(--na-border);
}

.preset-label {
  font-size: 12px;
  color: var(--na-muted-foreground);
}

.preset-chip {
  padding: 3px 10px;
  font-size: 12px;
  border: 1px solid var(--na-border);
  border-radius: 14px;
  background: transparent;
  color: var(--na-foreground);
  cursor: pointer;
  transition: all 0.16s ease;

  &:hover {
    border-color: var(--na-primary);
    color: var(--na-primary);
  }

  &.is-active {
    border-color: var(--na-primary);
    background: var(--na-primary-soft);
    color: var(--na-primary);
    font-weight: 600;
  }
}

.method-option-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;

  i {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }
  .dot-get { background: #3b82f6; }
  .dot-post { background: #10b981; }
  .dot-put { background: #f59e0b; }
  .dot-delete { background: #ef4444; }
  .dot-patch { background: #8b5cf6; }
}

/* 工具栏 */
.audit-table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  border-bottom: 1px solid var(--na-border);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;

  strong {
    font-size: 14px;
    font-weight: 650;
    color: var(--na-foreground);
  }
}

.toolbar-count {
  font-size: 12px;
  color: var(--na-muted-foreground);
}

.toolbar-filter-badge {
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 600;
  border-radius: 4px;
  background: var(--na-primary-soft);
  color: var(--na-primary);
}

/* 表格定制单元格 */
.operator-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.operator-avatar {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--na-primary-soft);
  color: var(--na-primary);
  font-size: 12px;
  font-weight: 700;
  display: grid;
  place-items: center;
}

.operator-info {
  display: flex;
  flex-direction: column;
  min-width: 0;

  strong {
    font-size: 13px;
    color: var(--na-foreground);
    line-height: 1.3;
  }
  small {
    font-size: 11px;
    color: var(--na-muted-foreground);
  }
}

/* HTTP Method Badge */
.method-badge {
  display: inline-block;
  padding: 2px 7px;
  font-size: 11px;
  font-weight: 700;
  border-radius: 4px;
  letter-spacing: 0.02em;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.method-get {
  background: rgba(59, 130, 246, 0.12);
  color: #2563eb;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.method-post {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.method-put {
  background: rgba(245, 158, 11, 0.12);
  color: #d97706;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.method-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #dc2626;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.method-patch {
  background: rgba(139, 92, 246, 0.12);
  color: #7c3aed;
  border: 1px solid rgba(139, 92, 246, 0.3);
}

/* 耗时 Chip */
.latency-chip {
  display: inline-block;
  padding: 1px 6px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 4px;
  font-variant-numeric: tabular-nums;
  font-family: ui-monospace, monospace;
}

.latency-fast {
  background: rgba(16, 185, 129, 0.08);
  color: #059669;
}

.latency-normal {
  background: rgba(59, 130, 246, 0.08);
  color: #2563eb;
}

.latency-slow {
  background: rgba(245, 158, 11, 0.12);
  color: #d97706;
}

.latency-danger {
  background: rgba(239, 68, 68, 0.12);
  color: #dc2626;
}

/* 路径单元格 */
.path-cell {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;

  .path-text {
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--na-foreground);
  }

  .copy-path-btn {
    opacity: 0;
    transition: opacity 0.16s ease;
  }

  &:hover .copy-path-btn {
    opacity: 1;
  }
}

/* 客户端单元格 */
.client-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;

  .ip-tag {
    font-family: ui-monospace, monospace;
    font-size: 12px;
    color: var(--na-foreground);
  }

  .agent-tag {
    font-size: 11px;
    color: var(--na-muted-foreground);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.time-cell {
  font-size: 12px;
  color: var(--na-muted-foreground);
  font-variant-numeric: tabular-nums;
}

.payload-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.table-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;

  .el-button {
    padding: 0;
    width: 28px;
    height: 28px;
    min-width: 28px;
  }
}

/* 抽屉样式 */
.audit-drawer-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 10px 4px 20px;
}

.drawer-headline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-radius: 8px;
  background: var(--na-card);
  border: 1px solid var(--na-border);

  &__tags {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  &__id {
    font-family: ui-monospace, monospace;
    font-size: 12px;
    color: var(--na-muted-foreground);
  }
}

.drawer-section {
  display: flex;
  flex-direction: column;
  gap: 10px;

  h3 {
    margin: 0;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    font-weight: 650;
    color: var(--na-foreground);
  }
}

.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.drawer-path-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;

  code {
    font-size: 12px;
    word-break: break-all;
    color: var(--na-primary);
  }
}

.code-container {
  border-radius: 6px;
  border: 1px solid var(--na-border);
  overflow: hidden;
}

.code-block {
  margin: 0;
  padding: 12px 14px;
  max-height: 280px;
  overflow: auto;
  background: var(--na-table-hover);
  color: var(--na-foreground);
  font: 12px/1.6 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.error-section {
  .error-block {
    background: rgba(239, 68, 68, 0.06);
    border: 1px solid rgba(239, 68, 68, 0.25);
    color: #dc2626;
    border-radius: 6px;
  }
}

.empty-hint {
  margin: 0;
  font-size: 12px;
  color: var(--na-muted-foreground);
  font-style: italic;
}

.drawer-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

@media (max-width: 1100px) {
  .audit-overview {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 720px) {
  .audit-overview {
    grid-template-columns: 1fr;
  }
  .audit-quick-presets {
    flex-wrap: wrap;
  }
}
</style>

