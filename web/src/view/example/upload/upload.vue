<template>
  <div class="na-page na-page--list upload-page">
    <div class="upload-layout min-w-0 gap-4 pt-2">
      <div class="na-panel upload-sidebar">
        <MediaCategoryTree
          :categories="categories"
          :active-id="search.classId"
          scroll-height="min(480px, calc(100vh - 330px))"
          @select="handleNodeClick"
          @add="addCategoryFun"
          @edit="editCategory"
          @delete="deleteCategoryFun"
        />
      </div>
      <div
        class="upload-main min-w-0 flex-1"
      >
        <div class="upload-table-box gva-table-box mt-0 mb-0">
          <warning-bar
            title="媒体库仅支持图片上传，上传文件会写入 OSS；点击“文件名”可以编辑，选择的类别即是上传的类别。"
          />
          <div class="upload-toolbar gva-btn-list gap-3">
            <upload-common
              :image-common="imageCommon"
              :classId="search.classId"
              @on-success="onSuccess"
            />
            <cropper-image :classId="search.classId" @on-success="onSuccess" />
            <QRCodeUpload :classId="search.classId" @on-success="onSuccess" />
            <upload-image
              :image-url="imageUrl"
              :file-size="512"
              :max-w-h="1080"
              :classId="search.classId"
              @on-success="onSuccess"
            />
            <el-button type="primary" icon="upload" :loading="importing" @click="importUrlFunc">
              导入URL
            </el-button>
            <el-input
              v-model="search.keyword"
              class="upload-search-input"
              placeholder="请输入文件名或备注"
            />
            <el-button type="primary" icon="search" :loading="tableLoading" @click="onSubmit"
              >查询
            </el-button>
          </div>

          <el-table v-loading="tableLoading && !tableLoaded" :data="tableData" class="upload-file-table" table-layout="fixed">
            <el-table-column align="left" label="预览" width="100">
              <template #default="scope">
                <CustomPic pic-type="file" :pic-src="scope.row.previewUrl || scope.row.url" preview />
              </template>
            </el-table-column>
            <el-table-column
              align="left"
              label="日期"
              prop="UpdatedAt"
              width="180"
            >
              <template #default="scope">
                <div>{{ formatDate(scope.row.UpdatedAt) }}</div>
              </template>
            </el-table-column>
            <el-table-column
              align="left"
              label="文件名/备注"
              prop="name"
              min-width="160"
            >
              <template #default="scope">
                <div
                  class="file-name-cell cursor-pointer"
                  @click="editFileNameFunc(scope.row)"
                >
                  {{ scope.row.name }}
                </div>
              </template>
            </el-table-column>
            <el-table-column
              align="left"
              label="链接"
              prop="url"
              min-width="220"
              show-overflow-tooltip
            >
              <template #default="scope">
                <span class="url-cell">{{ scope.row.url }}</span>
              </template>
            </el-table-column>
            <el-table-column align="left" label="标签" prop="tag" width="100">
              <template #default="scope">
                <el-tag
                  :type="
                    scope.row.tag?.toLowerCase() === 'jpg' ? 'info' : 'success'
                  "
                  disable-transitions
                  >{{ scope.row.tag }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column align="left" label="操作" width="160">
              <template #default="scope">
                <el-button
                  icon="download"
                  type="primary"
                  link
                  @click="downloadFile(scope.row)"
                  >下载
                </el-button>
                <el-button
                  icon="delete"
                  type="primary"
                  link
                  :loading="deletingId === scope.row.ID"
                  @click="deleteFileFunc(scope.row)"
                  >删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="upload-pagination gva-pagination">
            <el-pagination
              :current-page="page"
              :page-size="pageSize"
              :page-sizes="[10, 30, 50, 100]"
              :total="total"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="handleCurrentChange"
              @size-change="handleSizeChange"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 添加分类弹窗 -->
    <el-dialog
      v-model="categoryDialogVisible"
      @close="closeAddCategoryDialog"
      width="520"
      :title="(categoryFormData.ID === 0 ? '添加' : '编辑') + '分类'"
      draggable
    >
      <el-form
        ref="categoryForm"
        :rules="rules"
        :model="categoryFormData"
        label-width="80px"
      >
        <el-form-item label="上级分类">
          <el-tree-select
            v-model="categoryFormData.pid"
            :data="categories"
            check-strictly
            :props="defaultProps"
            :render-after-expand="false"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="分类名称" prop="name">
          <el-input
            v-model.trim="categoryFormData.name"
            placeholder="分类名称"
          ></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeAddCategoryDialog">取消</el-button>
        <el-button type="primary" :loading="savingCategory" @click="confirmAddCategory">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    getFileList,
    deleteFile,
    editFileName,
    importURL
  } from '@/api/fileUploadAndDownload'
  import { downloadImage } from '@/utils/downloadImg'
  import CustomPic from '@/components/customPic/index.vue'
  import UploadImage from '@/components/upload/image.vue'
  import UploadCommon from '@/components/upload/common.vue'
  import { CreateUUID, formatDate } from '@/utils/format'
  import WarningBar from '@/components/warningBar/warningBar.vue'

  import {
    onActivated,
    onBeforeUnmount,
    onDeactivated,
    onMounted,
    ref
  } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    addCategory,
    deleteCategory,
    getCategoryList
  } from '@/api/attachmentCategory'
  import CropperImage from '@/components/upload/cropper.vue'
  import QRCodeUpload from '@/components/upload/QR-code.vue'
  import MediaCategoryTree from '@/components/mediaCategoryTree/index.vue'

  defineOptions({
    name: 'Upload'
  })

  const path = ref(import.meta.env.VITE_BASE_API)

  const imageUrl = ref('')
  const imageCommon = ref('')

  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const search = ref({
    keyword: null,
    classId: 0
  })
  const tableData = ref([])
  const tableLoading = ref(false)
  const tableLoaded = ref(false)
  const deletingId = ref(0)
  const importing = ref(false)
  const savingCategory = ref(false)
  const mediaPreviewRefreshInterval = 10 * 60 * 1000
  let mediaPreviewRefreshTimer = null

  // 分页
  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  const onSubmit = () => {
    search.value.classId = 0
    page.value = 1
    getTableData()
  }

  // 查询
  const getTableData = async () => {
    tableLoading.value = true
    try {
      const table = await getFileList({
        page: page.value,
        pageSize: pageSize.value,
        ...search.value,
        fileType: 'image'
      })
      if (table.code === 0) {
        tableData.value = table.data.list
        total.value = table.data.total
        page.value = table.data.page
        pageSize.value = table.data.pageSize
      }
    } finally {
      tableLoading.value = false
      tableLoaded.value = true
    }
  }
  const refreshMediaPreviewURLs = () => {
    const hasSignedPreview = tableData.value.some(
      (item) => item.previewUrl && item.previewUrl !== item.url
    )
    if (document.visibilityState === 'visible' && hasSignedPreview) {
      getTableData()
    }
  }

  const startMediaPreviewRefresh = () => {
    if (mediaPreviewRefreshTimer !== null) return false
    mediaPreviewRefreshTimer = window.setInterval(
      refreshMediaPreviewURLs,
      mediaPreviewRefreshInterval
    )
    document.addEventListener('visibilitychange', refreshMediaPreviewURLs)
    return true
  }

  const stopMediaPreviewRefresh = () => {
    if (mediaPreviewRefreshTimer !== null) {
      window.clearInterval(mediaPreviewRefreshTimer)
      mediaPreviewRefreshTimer = null
    }
    document.removeEventListener('visibilitychange', refreshMediaPreviewURLs)
  }

  onMounted(() => {
    getTableData()
    startMediaPreviewRefresh()
  })
  onActivated(() => {
    if (startMediaPreviewRefresh()) {
      refreshMediaPreviewURLs()
    }
  })
  onDeactivated(stopMediaPreviewRefresh)
  onBeforeUnmount(stopMediaPreviewRefresh)

  const deleteFileFunc = async (row) => {
    ElMessageBox.confirm('此操作将永久删除文件, 是否继续?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        deletingId.value = row.ID
        try {
          const res = await deleteFile(row)
          if (res.code === 0) {
            ElMessage({
              type: 'success',
              message: '删除成功!'
            })
            if (tableData.value.length === 1 && page.value > 1) {
              page.value--
            }
            await getTableData()
          }
        } finally {
          deletingId.value = 0
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: '已取消删除'
        })
      })
  }

  const downloadFile = (row) => {
    const downloadUrl = row.previewUrl || row.url
    if (downloadUrl.indexOf('http://') > -1 || downloadUrl.indexOf('https://') > -1) {
      downloadImage(downloadUrl, row.name)
    } else {
      downloadImage(path.value + '/' + downloadUrl, row.name)
    }
  }

  /**
   * 编辑文件名或者备注
   * @param row
   * @returns {Promise<void>}
   */
  const editFileNameFunc = async (row) => {
    ElMessageBox.prompt('请输入文件名或者备注', '编辑', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /\S/,
      inputErrorMessage: '不能为空',
      inputValue: row.name
    })
      .then(async ({ value }) => {
        row.name = value
        // console.log(row)
        const res = await editFileName(row)
        if (res.code === 0) {
          ElMessage({
            type: 'success',
            message: '编辑成功!'
          })
          await getTableData()
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: '取消修改'
        })
      })
  }

  /**
   * 导入URL
   */
  const importUrlFunc = () => {
    ElMessageBox.prompt('仅支持图片 URL，格式：文件名|链接或者仅链接。', '导入', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputType: 'textarea',
      inputPlaceholder:
        '我的图片|https://my-oss.com/my.png\nhttps://my-oss.com/my_1.png',
      inputPattern: /\S/,
      inputErrorMessage: '不能为空'
    })
      .then(async ({ value }) => {
        let data = value.split('\n')
        let importData = []
        const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg', 'avif']
        data.forEach((item) => {
          let oneData = item.trim().split('|')
          let url, name
          if (oneData.length > 1) {
            name = oneData[0].trim()
            url = oneData[1]
          } else {
            url = oneData[0].trim()
            let str = url.substring(url.lastIndexOf('/') + 1)
            name = str.substring(0, str.lastIndexOf('.'))
          }
          const tag = url.substring(url.lastIndexOf('.') + 1).split('?')[0].split('#')[0].toLowerCase()
          if (url && imageExts.includes(tag)) {
            importData.push({
              name: name,
              url: url,
              classId: search.value.classId,
              tag,
              key: CreateUUID()
            })
          }
        })
        if (!importData.length) {
          ElMessage.warning('未检测到有效图片 URL')
          return
        }

        importing.value = true
        try {
          const res = await importURL(importData)
          if (res.code === 0) {
            ElMessage({
              type: 'success',
              message: '导入成功!'
            })
            await getTableData()
          }
        } finally {
          importing.value = false
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: '取消导入'
        })
      })
  }

  const onSuccess = () => {
    search.value.keyword = null
    page.value = 1
    getTableData()
  }

  const defaultProps = {
    children: 'children',
    label: 'name',
    value: 'ID'
  }

  const categories = ref([])
  const fetchCategories = async () => {
    const res = await getCategoryList()
    let data = {
      name: '全部分类',
      ID: 0,
      pid: 0,
      children: []
    }
    if (res.code === 0) {
      categories.value = res.data || []
      categories.value.unshift(data)
    }
  }

  const handleNodeClick = (node) => {
    search.value.keyword = null
    search.value.classId = node.ID
    page.value = 1
    getTableData()
  }

  const categoryDialogVisible = ref(false)
  const categoryFormData = ref({
    ID: 0,
    pid: 0,
    name: ''
  })

  const categoryForm = ref(null)
  const rules = ref({
    name: [
      { required: true, message: '请输入分类名称', trigger: 'blur' },
      { max: 20, message: '最多20位字符', trigger: 'blur' }
    ]
  })

  const addCategoryFun = (category) => {
    categoryDialogVisible.value = true
    categoryFormData.value.ID = 0
    categoryFormData.value.pid = category.ID
  }

  const editCategory = (category) => {
    categoryFormData.value = {
      ID: category.ID,
      pid: category.pid,
      name: category.name
    }
    categoryDialogVisible.value = true
  }

  const deleteCategoryFun = async (id) => {
    const res = await deleteCategory({ id: id })
    if (res.code === 0) {
      ElMessage.success({ type: 'success', message: '删除成功' })
      await fetchCategories()
    }
  }

  const confirmAddCategory = async () => {
    categoryForm.value.validate(async (valid) => {
      if (valid) {
        savingCategory.value = true
        try {
          const res = await addCategory(categoryFormData.value)
          if (res.code === 0) {
            ElMessage({ type: 'success', message: '操作成功' })
            await fetchCategories()
            closeAddCategoryDialog()
          }
        } finally {
          savingCategory.value = false
        }
      }
    })
  }

  const closeAddCategoryDialog = () => {
    categoryDialogVisible.value = false
    categoryFormData.value = {
      ID: 0,
      pid: 0,
      name: ''
    }
  }

  fetchCategories()
</script>


<style scoped lang="scss">
.upload-page {
  width: 100%;
  max-width: 100%;
  overflow-x: hidden;
}

.upload-layout,
.upload-main,
.upload-table-box {
  max-width: 100%;
  overflow-x: hidden;
}

.upload-layout {
  display: grid;
  grid-template-columns: 272px minmax(0, 1fr);
  align-items: start;
}

.upload-sidebar {
  position: sticky;
  top: 12px;
  align-self: start;
}

.upload-table-box {
  padding-right: 0;
}

.upload-toolbar {
  max-width: 100%;
  flex-wrap: wrap;
  align-items: center;
}

.upload-search-input {
  width: min(288px, 100%);
}

.upload-file-table {
  width: 100%;
  max-width: 100%;
}

.upload-file-table :deep(.el-table__inner-wrapper),
.upload-file-table :deep(.el-table__body-wrapper),
.upload-file-table :deep(.el-scrollbar),
.upload-file-table :deep(.el-scrollbar__wrap) {
  max-width: 100%;
}

.upload-file-table :deep(.cell) {
  min-width: 0;
}

.file-name-cell {
  max-width: 100%;
  overflow-wrap: anywhere;
  line-height: 1.45;
}

.url-cell {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload-pagination {
  display: flex;
  justify-content: flex-end;
  max-width: 100%;
  overflow-x: hidden;
  padding: 18px 0 4px;
}

.upload-pagination :deep(.el-pagination) {
  flex-wrap: wrap;
  justify-content: flex-end;
  row-gap: 8px;
  max-width: 100%;
}

@media (max-width: 1200px) {
  .upload-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .upload-sidebar {
    position: static;
    width: 100%;
  }
}
</style>
