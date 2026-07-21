<template>
  <main class="na-page na-page--list announcement-page">
    <AppPageHeader
      title-id="announcement-title"
      title="公告中心"
      description="发布与管理系统公告，在线用户实时提醒，离线用户登录后可见未读标记。"
    >
      <template #actions>
        <el-button type="primary" icon="plus" size="large" @click="openDialog">发布公告</el-button>
      </template>
    </AppPageHeader>

    <section class="na-panel filter-panel" aria-label="公告筛选">
      <el-form
        ref="elSearchFormRef"
        :model="searchInfo"
        :rules="searchRule"
        label-position="top"
        @keyup.enter="onSubmit"
      >
        <div class="filter-grid">
          <el-form-item prop="createdAt">
            <template #label>
              <span>
                创建日期
                <el-tooltip content="搜索范围是开始日期（包含）至结束日期（不包含）">
                  <el-icon><QuestionFilled /></el-icon>
                </el-tooltip>
              </span>
            </template>
            <div class="date-range">
              <el-date-picker
                v-model="searchInfo.startCreatedAt"
                type="datetime"
                placeholder="开始日期"
                :disabled-date="
                  (time) =>
                    searchInfo.endCreatedAt
                      ? time.getTime() > searchInfo.endCreatedAt.getTime()
                      : false
                "
              />
              <span class="date-sep">—</span>
              <el-date-picker
                v-model="searchInfo.endCreatedAt"
                type="datetime"
                placeholder="结束日期"
                :disabled-date="
                  (time) =>
                    searchInfo.startCreatedAt
                      ? time.getTime() < searchInfo.startCreatedAt.getTime()
                      : false
                "
              />
            </div>
          </el-form-item>
          <el-form-item label="发布状态">
            <el-select v-model="searchInfo.status" clearable placeholder="全部状态">
              <el-option label="已发布" value="published" />
              <el-option label="草稿" value="draft" />
            </el-select>
          </el-form-item>
          <div class="filter-actions">
            <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
            <el-button icon="refresh" @click="onReset">重置</el-button>
          </div>
        </div>
      </el-form>
    </section>

    <section class="na-panel table-panel">
      <header class="na-panel-header panel-header">
        <div>
          <h2>公告列表</h2>
          <span>共 {{ total }} 条公告</span>
        </div>
        <el-button
          icon="delete"
          text
          type="danger"
          :disabled="!multipleSelection.length"
          @click="onDelete"
        >
          删除所选{{ multipleSelection.length ? `（${multipleSelection.length}）` : '' }}
        </el-button>
      </header>

      <el-table
        ref="multipleTable"
        :data="tableData"
        row-key="ID"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column align="left" label="标题" prop="title" min-width="240" show-overflow-tooltip />
        <el-table-column align="left" label="状态" prop="status" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'published' ? 'success' : 'info'" effect="light">
              {{ scope.row.status === 'published' ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="作者" prop="userID" width="130">
          <template #default="scope">
            <span>{{ filterDataSource(dataSource.userID, scope.row.userID) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="创建日期" prop="CreatedAt" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="附件" prop="attachments" min-width="200">
          <template #default="scope">
            <div class="file-list">
              <el-tag
                v-for="file in scope.row.attachments"
                :key="file.uid"
                class="file-tag"
                @click="downloadFile(file.url)"
              >
                {{ file.name }}
              </el-tag>
              <span v-if="!scope.row.attachments?.length" class="file-empty">—</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column align="center" label="操作" fixed="right" width="160">
          <template #default="scope">
            <el-button type="primary" link icon="edit" @click="updateInfoFunc(scope.row)">变更</el-button>
            <el-button type="danger" link icon="delete" @click="deleteRow(scope.row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无公告">
            <el-button type="primary" @click="openDialog">发布第一条公告</el-button>
          </el-empty>
        </template>
      </el-table>

      <div class="na-pagination pagination-wrap">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </section>

    <el-drawer
      v-model="dialogFormVisible"
      destroy-on-close
      size="800"
      :show-close="false"
      :before-close="closeDialog"
    >
      <template #header>
        <div class="drawer-head">
          <div class="drawer-title">
            <span>{{ type === 'create' ? '发布公告' : '编辑公告' }}</span>
            <small>公告发布后，在线用户会立即收到顶部提醒</small>
          </div>
          <div class="drawer-actions">
            <el-button @click="enterDialog('draft')">{{ formData.status === 'published' ? '转为草稿' : '保存草稿' }}</el-button>
            <el-button type="primary" @click="enterDialog('published')">{{ formData.status === 'published' ? '保存修改' : '发布公告' }}</el-button>
            <el-button @click="closeDialog"> 取 消 </el-button>
          </div>
        </div>
      </template>

      <el-form
        ref="elFormRef"
        :model="formData"
        label-position="top"
        :rules="rule"
        label-width="80px"
      >
        <el-form-item label="标题:" prop="title">
          <el-input
            v-model="formData.title"
            :clearable="true"
            placeholder="请输入标题"
          />
        </el-form-item>
        <el-form-item label="内容:" prop="content">
          <RichEdit v-model="formData.content" />
        </el-form-item>
        <el-form-item label="附件:" prop="attachments">
          <SelectFile v-model="formData.attachments" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </main>
</template>

<script setup>
  import {
    getInfoDataSource,
    createInfo,
    deleteInfo,
    deleteInfoByIds,
    updateInfo,
    findInfo,
    getInfoList
  } from '@/plugin/announcement/api/info'
  import { getUrl } from '@/utils/image'
  // 富文本组件
  import RichEdit from '@/components/richtext/rich-edit.vue'
  // 文件选择组件
  import SelectFile from '@/components/selectFile/selectFile.vue'

  // 全量引入格式化工具 请按需保留
  import { formatDate, filterDataSource } from '@/utils/format'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref, reactive } from 'vue'
  import AppPageHeader from '@/components/page/AppPageHeader.vue'

  defineOptions({
    name: 'Info'
  })

  // 自动化生成的字典（可能为空）以及字段
  const formData = ref({
    title: '',
    content: '',
    userID: undefined,
    attachments: [],
    status: 'draft'
  })
  const dataSource = ref([])
  const getDataSourceFunc = async () => {
    const res = await getInfoDataSource()
    if (res.code === 0) {
      dataSource.value = res.data
    }
  }
  getDataSourceFunc()

  // 验证规则
  const rule = reactive({
    title: [{ required: true, message: '请输入公告标题', trigger: 'blur' }],
    content: [{ required: true, message: '请输入公告内容', trigger: 'change' }]
  })

  const searchRule = reactive({
    createdAt: [
      {
        validator: (rule, value, callback) => {
          if (
            searchInfo.value.startCreatedAt &&
            !searchInfo.value.endCreatedAt
          ) {
            callback(new Error('请填写结束日期'))
          } else if (
            !searchInfo.value.startCreatedAt &&
            searchInfo.value.endCreatedAt
          ) {
            callback(new Error('请填写开始日期'))
          } else if (
            searchInfo.value.startCreatedAt &&
            searchInfo.value.endCreatedAt &&
            (searchInfo.value.startCreatedAt.getTime() ===
              searchInfo.value.endCreatedAt.getTime() ||
              searchInfo.value.startCreatedAt.getTime() >
                searchInfo.value.endCreatedAt.getTime())
          ) {
            callback(new Error('开始日期应当早于结束日期'))
          } else {
            callback()
          }
        },
        trigger: 'change'
      }
    ]
  })

  const elFormRef = ref()
  const elSearchFormRef = ref()

  // =========== 表格控制部分 ===========
  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const searchInfo = ref({})

  // 重置
  const onReset = () => {
    searchInfo.value = {}
    getTableData()
  }

  // 搜索
  const onSubmit = () => {
    elSearchFormRef.value?.validate(async (valid) => {
      if (!valid) return
      page.value = 1
      getTableData()
    })
  }

  // 分页
  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  // 修改页面容量
  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  // 查询
  const getTableData = async () => {
    const table = await getInfoList({
      page: page.value,
      pageSize: pageSize.value,
      ...searchInfo.value
    })
    if (table.code === 0) {
      tableData.value = table.data.list
      total.value = table.data.total
      page.value = table.data.page
      pageSize.value = table.data.pageSize
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
      deleteInfoFunc(row)
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
      const res = await deleteInfoByIds({ IDs })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        if (tableData.value.length === IDs.length && page.value > 1) {
          page.value--
        }
        getTableData()
      }
    })
  }

  // 行为控制标记（弹窗内部需要增还是改）
  const type = ref('')

  // 更新行
  const updateInfoFunc = async (row) => {
    const res = await findInfo({ ID: row.ID })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data
      dialogFormVisible.value = true
    }
  }

  // 删除行
  const deleteInfoFunc = async (row) => {
    const res = await deleteInfo({ ID: row.ID })
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: '删除成功'
      })
      if (tableData.value.length === 1 && page.value > 1) {
        page.value--
      }
      getTableData()
    }
  }

  // 弹窗控制标记
  const dialogFormVisible = ref(false)

  // 打开弹窗
  const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
  }

  // 关闭弹窗
  const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
      title: '',
      content: '',
      userID: undefined,
      attachments: [],
      status: 'draft'
    }
  }
  // 弹窗确定
  const enterDialog = async (status) => {
    elFormRef.value?.validate(async (valid) => {
      if (!valid) return
      formData.value.status = status
      let res
      switch (type.value) {
        case 'create':
          res = await createInfo(formData.value)
          break
        case 'update':
          res = await updateInfo(formData.value)
          break
        default:
          res = await createInfo(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: status === 'published' ? '公告已发布' : '草稿已保存'
        })
        closeDialog()
        getTableData()
      }
    })
  }

  const downloadFile = (url) => {
    window.open(getUrl(url), '_blank')
  }
</script>

<style scoped lang="scss">
  .filter-panel { padding: 14px 16px 0; }
  .filter-grid { display: grid; grid-template-columns: minmax(320px, 1.6fr) minmax(160px, 1fr) auto; gap: 14px; align-items: end; }
  .date-range { display: flex; align-items: center; gap: 8px; }
  .date-sep { color: var(--na-muted-foreground); }
  .filter-actions { display: flex; gap: 8px; padding-bottom: 18px; }

  .table-panel { overflow: hidden; }
  .panel-header h2 { margin: 0 0 3px; font-size: 17px; }
  .panel-header span { color: var(--na-muted-foreground); font-size: 13px; }

  .file-list { display: flex; flex-wrap: wrap; gap: 4px; }
  .file-tag { cursor: pointer; }
  .file-empty { color: var(--na-muted-foreground); }

  .drawer-head { display: flex; flex: 1; align-items: center; justify-content: space-between; gap: 16px; }
  .drawer-title { display: flex; flex-direction: column; gap: 4px; }
  .drawer-title span { color: var(--na-foreground); font-size: 20px; font-weight: 700; }
  .drawer-title small { color: var(--na-muted-foreground); font-weight: 400; }
  .drawer-actions { display: flex; flex: 0 0 auto; gap: 8px; }

  @media (max-width: 1000px) { .filter-grid { grid-template-columns: 1fr; } .date-range { flex-wrap: wrap; } }
  @media (max-width: 767px) {
    .drawer-head { align-items: flex-start; flex-direction: column; }
  }
</style>
