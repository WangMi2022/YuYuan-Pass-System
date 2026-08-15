<template>
  <div class="dictionary-detail-page">
    <section class="na-panel dictionary-detail-panel">
      <header class="dictionary-detail-panel__header">
        <div class="dictionary-detail-panel__title-group">
          <span class="dictionary-detail-panel__title">字典详细内容</span>
          <span class="dictionary-detail-panel__count">{{ detailCount }} 个条目</span>
        </div>
        <div class="dictionary-detail-panel__actions">
          <el-input
            placeholder="搜索展示值"
            v-model="searchName"
            clearable
            class="dictionary-detail-panel__search"
            @clear="clearSearchInput"
            :prefix-icon="Search"
            v-click-outside="handleCloseSearchInput"
            @keydown="handleInputKeyDown"
          >
            <template #append>
              <el-button
                :type="searchName ? 'primary' : 'info'"
                @click="applySearch"
                >搜索</el-button
              >
            </template>
          </el-input>
          <el-button
            type="primary"
            :icon="Plus"
            :disabled="!props.sysDictionaryID"
            @click="openDrawer"
          >
            新增字典项
          </el-button>
        </div>
      </header>
      <div class="dictionary-detail-panel__body">
        <el-table
          class="dictionary-detail-table"
          :data="displayTreeData"
          style="width: 100%"
          tooltip-effect="dark"
          :tree-props="{ children: 'children'}"
          row-key="ID"
          default-expand-all
        >
          <el-table-column type="selection" width="52" />

          <el-table-column align="left" label="展示值" prop="label" min-width="140">
            <template #default="scope">
              <span class="dictionary-detail-table__label">{{ scope.row.label }}</span>
            </template>
          </el-table-column>

          <el-table-column align="left" label="字典值" prop="value" min-width="120">
            <template #default="scope">
              <code class="dictionary-detail-table__code">{{ scope.row.value }}</code>
            </template>
          </el-table-column>

          <el-table-column align="left" label="扩展值" prop="extend" min-width="120">
            <template #default="scope">
              <span class="dictionary-detail-table__muted">{{ scope.row.extend || '无' }}</span>
            </template>
          </el-table-column>

        <el-table-column align="left" label="层级" prop="level" width="80" />

        <el-table-column
          align="left"
          label="启用状态"
          prop="status"
          width="100"
        >
          <template #default="scope">
              <el-tag
                :type="scope.row.status ? 'success' : 'info'"
                effect="light"
                round
              >
                {{ formatBoolean(scope.row.status) }}
              </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          align="left"
          label="排序标记"
          prop="sort"
          width="100"
        />

        <el-table-column
          align="left"
          label="操作"
          :min-width="appStore.operateMinWith"
          fixed="right"
        >
          <template #default="scope">
              <div class="dictionary-detail-table__actions">
            <el-button
              type="primary"
              link
                  :icon="Plus"
              @click="addChildNode(scope.row)"
            >
              添加子项
            </el-button>
            <el-button
              type="primary"
              link
                  :icon="Edit"
              @click="updateSysDictionaryDetailFunc(scope.row)"
            >
              变更
            </el-button>
            <el-button
                  type="danger"
              link
                  :icon="Delete"
              @click="deleteSysDictionaryDetailFunc(scope.row)"
            >
              删除
            </el-button>
              </div>
          </template>
        </el-table-column>
          <template #empty>
            <AppEmptyState
              compact
              :title="detailEmptyTitle"
              :description="detailEmptyDescription"
              :highlights="['支持父子层级', '可配置排序与启停状态']"
            >
              <template #actions>
                <el-button v-if="searchName" @click="clearSearchInput">清除筛选</el-button>
                <el-button
                  v-else-if="props.sysDictionaryID"
                  type="primary"
                  :icon="Plus"
                  @click="openDrawer"
                >
                  新增字典项
                </el-button>
              </template>
            </AppEmptyState>
          </template>
      </el-table>
      </div>
    </section>

    <el-drawer
      v-model="drawerFormVisible"
      :size="appStore.drawerSize"
      :show-close="false"
      :before-close="closeDrawer"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{
            type === 'create' ? '添加字典项' : '修改字典项'
          }}</span>
          <div>
            <el-button @click="closeDrawer"> 取 消 </el-button>
            <el-button type="primary" @click="enterDrawer"> 确 定 </el-button>
          </div>
        </div>
      </template>
      <el-form
        ref="drawerForm"
        :model="formData"
        :rules="rules"
        label-width="110px"
      >
        <el-form-item label="父级字典项" prop="parentID">
          <el-cascader
            v-model="formData.parentID"
            :options="[rootOption,...treeData]"
            :props="cascadeProps"
            placeholder="请选择父级字典项（可选）"
            clearable
            filterable
            :style="{ width: '100%' }"
            @change="handleParentChange"
          />
        </el-form-item>
        <el-form-item label="展示值" prop="label">
          <el-input
            v-model="formData.label"
            placeholder="请输入展示值"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
        <el-form-item label="字典值" prop="value">
          <el-input
            v-model="formData.value"
            placeholder="请输入字典值"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
        <el-form-item label="扩展值" prop="extend">
          <el-input
            v-model="formData.extend"
            placeholder="请输入扩展值"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
        <el-form-item label="启用状态" prop="status" required>
          <el-switch
            v-model="formData.status"
            active-text="开启"
            inactive-text="停用"
          />
        </el-form-item>
        <el-form-item label="排序标记" prop="sort">
          <el-input-number
            v-model.number="formData.sort"
            placeholder="排序标记"
          />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createSysDictionaryDetail,
    deleteSysDictionaryDetail,
    updateSysDictionaryDetail,
    findSysDictionaryDetail,
    getDictionaryTreeList
  } from '@/api/sysDictionaryDetail' // 此处请自行替换地址
  import { computed, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatBoolean } from '@/utils/format'
  import { useAppStore } from '@/pinia'
  import { Delete, Edit, Plus, Search } from '@element-plus/icons-vue'
  import AppEmptyState from '@/components/page/AppEmptyState.vue'

  defineOptions({
    name: 'SysDictionaryDetail'
  })

  const appStore = useAppStore()
  const searchName = ref('')

  const props = defineProps({
    sysDictionaryID: {
      type: Number,
      default: 0
    }
  })

  const formData = ref({
    label: null,
    value: null,
    status: true,
    sort: null,
    parentID: null
  })

  const rules = ref({
    label: [
      {
        required: true,
        message: '请输入展示值',
        trigger: 'blur'
      }
    ],
    value: [
      {
        required: true,
        message: '请输入字典值',
        trigger: 'blur'
      }
    ],
    sort: [
      {
        required: true,
        message: '排序标记',
        trigger: 'blur'
      }
    ]
  })

  const treeData = ref([])
  const displayTreeData = ref([])
  const detailCount = computed(() => displayTreeData.value.length)
  const detailEmptyTitle = computed(() => {
    if (searchName.value) return '未找到匹配的字典项'
    return props.sysDictionaryID ? '当前字典还没有条目' : '尚未选择字典'
  })
  const detailEmptyDescription = computed(() => {
    if (searchName.value) return `没有包含“${searchName.value}”的展示值，可清除筛选查看全部条目。`
    if (props.sysDictionaryID) return '新增根级条目后，可继续建立子项并设置排序与启用状态。'
    return '从左侧选择一个字典后，这里会显示对应的层级条目。'
  })

  // 级联选择器配置
  const cascadeProps = {
    value: 'ID',
    label: 'label',
    children: 'children',
    checkStrictly: true, // 允许选择任意级别
    emitPath: false // 只返回选中节点的值
  }


  const normalizeSearch = (value) => (value ?? '').toString().toLowerCase()

  const filterTree = (nodes, keyword) => {
    const trimmed = normalizeSearch(keyword).trim()
    if (!trimmed) {
      return nodes
    }
    const walk = (list) => {
      const result = []
      for (const node of list) {
        const label = normalizeSearch(node.label)
        const children = Array.isArray(node.children) ? walk(node.children) : []
        if (label.includes(trimmed) || children.length > 0) {
          result.push({
            ...node,
            children
          })
        }
      }
      return result
    }
    return walk(nodes)
  }

  const applySearch = () => {
    displayTreeData.value = filterTree(treeData.value, searchName.value)
  }

  // 获取树形数据
  const getTreeData = async () => {
    if (!props.sysDictionaryID) return
    try {
      const res = await getDictionaryTreeList({
        sysDictionaryID: props.sysDictionaryID
      })
      if (res.code === 0) {
        treeData.value = res.data.list || []
        applySearch()
      }
    } catch (error) {
      console.error('获取树形数据失败:', error)
      ElMessage.error('获取层级数据失败')
    }
  }

  const rootOption = {
    ID: null,
    label: '无父级（根级）'
  }


  // 初始加载
  getTreeData()

  const type = ref('')
  const drawerFormVisible = ref(false)

  const updateSysDictionaryDetailFunc = async (row) => {
    drawerForm.value && drawerForm.value.clearValidate()
    const res = await findSysDictionaryDetail({ ID: row.ID })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data.reSysDictionaryDetail
      drawerFormVisible.value = true
    }
  }

  // 添加子节点
  const addChildNode = (parentNode) => {
    type.value = 'create'
    formData.value = {
      label: null,
      value: null,
      status: true,
      sort: null,
      parentID: parentNode.ID,
      sysDictionaryID: props.sysDictionaryID
    }
    drawerForm.value && drawerForm.value.clearValidate()
    drawerFormVisible.value = true
  }

  // 处理父级选择变化
  const handleParentChange = (value) => {
    formData.value.parentID = value
  }

  const closeDrawer = () => {
    drawerFormVisible.value = false
    formData.value = {
      label: null,
      value: null,
      status: true,
      sort: null,
      parentID: null,
      sysDictionaryID: props.sysDictionaryID
    }
  }

  const deleteSysDictionaryDetailFunc = async (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteSysDictionaryDetail({ ID: row.ID })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        await getTreeData() // 重新加载数据
      }
    })
  }

  const drawerForm = ref(null)
  const enterDrawer = async () => {
    drawerForm.value.validate(async (valid) => {
      formData.value.sysDictionaryID = props.sysDictionaryID
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createSysDictionaryDetail(formData.value)
          break
        case 'update':
          res = await updateSysDictionaryDetail(formData.value)
          break
        default:
          res = await createSysDictionaryDetail(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '创建/更改成功'
        })
        closeDrawer()
        await getTreeData() // 重新加载数据
      }
    })
  }

  const openDrawer = () => {
    type.value = 'create'
    formData.value.parentID = null
    drawerForm.value && drawerForm.value.clearValidate()
    drawerFormVisible.value = true
  }

  const clearSearchInput = () => {
    searchName.value = ''
    applySearch()
  }

  const handleCloseSearchInput = () => {
    // 处理搜索输入框关闭
  }

  const handleInputKeyDown = (e) => {
    if (e.key === 'Enter') {
      applySearch()
    }
  }

  watch(
    () => props.sysDictionaryID,
    () => {
      getTreeData()
    }
  )
</script>

<style scoped>
  .dictionary-detail-page {
    height: 100%;
    min-height: 0;
  }

  .dictionary-detail-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    overflow: hidden;
  }

  .dictionary-detail-panel__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 14px 16px;
    border-bottom: 1px solid var(--na-border);
  }

  .dictionary-detail-panel__title-group {
    min-width: 0;
  }

  .dictionary-detail-panel__title {
    display: block;
    color: var(--na-foreground);
    font-size: 14px;
    font-weight: 650;
    line-height: 1.35;
  }

  .dictionary-detail-panel__count {
    display: block;
    margin-top: 2px;
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 1.4;
  }

  .dictionary-detail-panel__actions {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    min-width: 0;
  }

  .dictionary-detail-panel__search {
    width: min(320px, 36vw);
  }

  .dictionary-detail-panel__body {
    flex: 1 1 auto;
    min-height: 0;
    padding: 16px;
    overflow: auto;
  }

  .dictionary-detail-table {
    min-width: 860px;
  }

  .dictionary-detail-table__label {
    color: var(--na-foreground);
    font-weight: 560;
  }

  .dictionary-detail-table__code {
    display: inline-flex;
    max-width: 100%;
    align-items: center;
    min-height: 24px;
    padding: 2px 7px;
    border: 1px solid var(--na-border);
    border-radius: 7px;
    background: color-mix(in srgb, var(--na-muted) 70%, var(--na-card));
    color: color-mix(in srgb, var(--na-foreground) 88%, var(--na-primary));
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    line-height: 1.3;
    word-break: break-all;
  }

  .dictionary-detail-table__muted {
    color: var(--na-muted-foreground);
  }

  .dictionary-detail-table__actions {
    display: flex;
    align-items: center;
    gap: 2px;
    white-space: nowrap;
  }

  .dictionary-detail-table__actions :deep(.el-button.is-link) {
    min-height: 28px;
    padding: 5px 7px;
  }

  @media (max-width: 980px) {
    .dictionary-detail-panel__header {
      align-items: stretch;
      flex-direction: column;
    }

    .dictionary-detail-panel__actions {
      justify-content: flex-start;
      width: 100%;
    }

    .dictionary-detail-panel__search {
      width: 100%;
      max-width: 360px;
    }
  }

  @media (max-width: 640px) {
    .dictionary-detail-panel__actions {
      flex-wrap: wrap;
    }

    .dictionary-detail-panel__search {
      max-width: none;
    }
  }
</style>
