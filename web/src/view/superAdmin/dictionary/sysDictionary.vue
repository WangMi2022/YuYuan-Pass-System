<template>
  <div class="na-workspace na-workspace--flush dictionary-page">
    <warning-bar
      class="dictionary-warning"
      title="获取字典且缓存方法已在前端utils/dictionary 已经封装完成 不必自己书写 使用方法查看文件内注释"
    />
    <el-splitter class="dictionary-splitter">
      <el-splitter-panel size="320px" min="260px" max="460px" collapsible>
        <aside class="na-panel dictionary-sidebar">
          <header class="dictionary-sidebar__header">
            <div class="dictionary-sidebar__title-group">
              <span class="dictionary-sidebar__title">字典列表</span>
              <span class="dictionary-sidebar__count">{{ dictionaryData.length }} 个字典</span>
            </div>
            <div class="dictionary-sidebar__actions">
              <el-tooltip content="搜索" placement="top">
                <el-button
                  circle
                  :icon="Search"
                  :class="{ 'is-active': showSearchInput }"
                  @click="showSearchInputHandler"
                />
              </el-tooltip>
              <el-tooltip content="导入字典" placement="top">
                <el-button
                  circle
                  type="success"
                  plain
                  :icon="Upload"
                  @click="openImportDialog"
                />
              </el-tooltip>
              <el-tooltip content="AI 生成字典" placement="top">
                <el-button
                  class="dictionary-sidebar__ai-button"
                  type="warning"
                  plain
                  @click="openAiDialog"
                >
                  AI
                </el-button>
              </el-tooltip>
              <el-tooltip content="新建字典" placement="top">
                <el-button
                  circle
                  type="primary"
                  :icon="Plus"
                  @click="openDrawer"
                />
              </el-tooltip>
            </div>
          </header>
          <transition name="dictionary-search">
            <div
              v-if="showSearchInput"
              class="dictionary-sidebar__search"
            >
              <el-input
                v-model="searchName"
                placeholder="搜索"
                clearable
                :autofocus="showSearchInput"
                @clear="clearSearchInput"
                :prefix-icon="Search"
                v-click-outside="handleCloseSearchInput"
                @keydown="handleInputKeyDown"
              >
                <template #append>
                  <el-button
                    :type="searchName ? 'primary' : 'info'"
                    @click="getTableData"
                    >搜索</el-button
                  >
                </template>
              </el-input>
            </div>
          </transition>
          <el-scrollbar class="dictionary-sidebar__scroll">
            <div
              v-if="dictionaryData.length"
              class="dictionary-list"
            >
              <div
                v-for="dictionary in dictionaryData"
                :key="dictionary.ID"
                class="dictionary-list__item"
                :class="[
                  selectID === dictionary.ID ? 'is-active' : '',
                  dictionary.parentID ? 'is-child' : ''
                ]"
                role="button"
                tabindex="0"
                :aria-current="selectID === dictionary.ID ? 'true' : undefined"
                :title="`${dictionary.name}（${dictionary.type}）`"
                @click="toDetail(dictionary)"
                @keydown.enter.prevent="toDetail(dictionary)"
                @keydown.space.prevent="toDetail(dictionary)"
              >
                <div class="dictionary-list__content">
                  <span
                    v-if="dictionary.parentID"
                    class="dictionary-list__branch"
                    >└─</span
                  >
                  <span class="dictionary-list__name">{{ dictionary.name }}</span>
                  <span class="dictionary-list__type">{{ dictionary.type }}</span>
                </div>

                <div class="dictionary-list__actions">
                  <el-tooltip content="导出字典" placement="top">
                    <el-button
                      link
                      type="success"
                      :icon="Download"
                      aria-label="导出字典"
                      @click.stop="exportDictionary(dictionary)"
                    />
                  </el-tooltip>
                  <el-tooltip content="编辑字典" placement="top">
                    <el-button
                      link
                      type="primary"
                      :icon="Edit"
                      aria-label="编辑字典"
                      @click.stop="updateSysDictionaryFunc(dictionary)"
                    />
                  </el-tooltip>
                  <el-tooltip content="删除字典" placement="top">
                    <el-button
                      link
                      type="danger"
                      :icon="Delete"
                      aria-label="删除字典"
                      @click.stop="deleteSysDictionaryFunc(dictionary)"
                    />
                  </el-tooltip>
                </div>
              </div>
            </div>
            <AppEmptyState
              v-else
              compact
              title="字典列表为空"
              description="先创建基础字典，或导入已有 JSON；后续业务表单即可复用统一枚举值。"
              :highlights="['支持父子字典', '支持 JSON 导入导出', '可使用 AI 生成草稿']"
            >
              <template #actions>
                <el-button type="primary" :icon="Plus" @click="openDrawer">新建字典</el-button>
              </template>
            </AppEmptyState>
          </el-scrollbar>
        </aside>
      </el-splitter-panel>
      <el-splitter-panel :min="420">
        <main class="dictionary-detail">
          <sysDictionaryDetail :sys-dictionary-i-d="selectID" />
        </main>
      </el-splitter-panel>
    </el-splitter>

    <el-drawer
      v-model="drawerFormVisible"
      :size="appStore.drawerSize"
      :show-close="false"
      :before-close="closeDrawer"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{
            type === 'create' ? '添加字典' : '修改字典'
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
        <el-form-item label="父级字典" prop="parentID">
          <el-select
            v-model="formData.parentID"
            placeholder="请选择父级字典（可选）"
            clearable
            filterable
            :style="{ width: '100%' }"
          >
            <el-option
              v-for="dict in availableParentDictionaries"
              :key="dict.ID"
              :label="`${dict.name}（${dict.type}）`"
              :value="dict.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="字典名（中）" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入字典名（中）"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
        <el-form-item label="字典名（英）" prop="type">
          <el-input
            v-model="formData.type"
            placeholder="请输入字典名（英）"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status" required>
          <el-switch
            v-model="formData.status"
            active-text="开启"
            inactive-text="停用"
          />
        </el-form-item>
        <el-form-item label="描述" prop="desc">
          <el-input
            v-model="formData.desc"
            placeholder="请输入描述"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
      </el-form>
    </el-drawer>

    <!-- 导入字典抽屉 -->
    <el-drawer
      v-model="importDrawerVisible"
      :size="appStore.drawerSize"
      :show-close="false"
      :before-close="closeImportDrawer"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">导入字典JSON</span>
          <div>
            <el-button @click="closeImportDrawer"> 取 消 </el-button>
            <el-button type="primary" @click="handleImport" :loading="importing">
              确认导入
            </el-button>
          </div>
        </div>
      </template>
      
      <div class="import-drawer-content">
        <div class="mb-4">
          <el-alert
            title="请粘贴、编辑或拖拽JSON文件到下方区域"
            type="info"
            :closable="false"
            show-icon
          />
        </div>

        <!-- 拖拽上传区域 -->
        <div
          class="drag-upload-area"
          :class="{ 'is-dragging': isDragging }"
          @drop.prevent="handleDrop"
          @dragover.prevent="handleDragOver"
          @dragleave.prevent="handleDragLeave"
          @click="triggerFileInput"
        >
          <el-icon class="upload-icon"><Upload /></el-icon>
          <div class="upload-text">
            <p>将 JSON 文件拖到此处，或点击选择文件</p>
            <p class="upload-hint">也可以在下方文本框直接编辑</p>
          </div>
          <input
            ref="fileInputRef"
            type="file"
            accept=".json,application/json"
            style="display: none"
            @change="handleFileSelect"
          />
        </div>

        <div class="json-editor-container mt-4">
          <el-input
            v-model="importJsonText"
            type="textarea"
            :rows="15"
            placeholder='请输入JSON数据，例如：
{
  "name": "性别",
  "type": "gender",
  "status": true,
  "desc": "性别字典",
  "sysDictionaryDetails": [
    {
      "label": "男",
      "value": "1",
      "status": true,
      "sort": 1
    },
    {
      "label": "女",
      "value": "2",
      "status": true,
      "sort": 2
    }
  ]
}'
            class="json-textarea"
          />
        </div>

        <div class="mt-4" v-if="jsonPreviewError">
          <el-alert
            :title="jsonPreviewError"
            type="error"
            :closable="false"
            show-icon
          />
        </div>

    
      </div>
    </el-drawer>

    <!-- AI 对话框 -->
    <el-dialog
      v-model="aiDialogVisible"
      title="AI 生成字典"
      width="520px"
      :before-close="closeAiDialog"
    >
      <div class="relative">
        <el-input
          v-model="aiPrompt"
          type="textarea"
          :rows="6"
          :maxlength="2000"
          placeholder="请输入生成字典的描述，例如：生成一个用户状态字典（启用/禁用）\n支持粘贴或上传图片后识图生成。"
          resize="none"
          @keydown.ctrl.enter="handleAiGenerate"
          @paste="handlePaste"
          @focus="handleFocus"
          @blur="handleBlur"
        />

        <div class="flex absolute right-2 bottom-2">
          <el-tooltip effect="light">
            <template #content>
              <div>粘贴或上传图片后，识别图片内容生成字典。</div>
            </template>
            <el-button type="primary" @click="eyeFunc">
                <el-icon size="18">
                <ai-gva />
              </el-icon>
              识图
            </el-button>
          </el-tooltip>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeAiDialog">取 消</el-button>
          <el-button type="primary" @click="handleAiGenerate" :loading="aiGenerating">
            确 定
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    createSysDictionary,
    deleteSysDictionary,
    updateSysDictionary,
    findSysDictionary,
    getSysDictionaryList,
    exportSysDictionary,
    importSysDictionary
  } from '@/api/sysDictionary' // 此处请自行替换地址
  import { llmAuto } from '@/api/autoCode'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import AppEmptyState from '@/components/page/AppEmptyState.vue'
  import { ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'

  import sysDictionaryDetail from './sysDictionaryDetail.vue'
  import { Delete, Edit, Plus, Search, Download, Upload } from '@element-plus/icons-vue'
  import { useAppStore } from '@/pinia'

  defineOptions({
    name: 'SysDictionary'
  })

  const appStore = useAppStore()

  const selectID = ref(0)

  const formData = ref({
    name: null,
    type: null,
    status: true,
    desc: null,
    parentID: null
  })
  const searchName = ref('')
  const showSearchInput = ref(false)
  const rules = ref({
    name: [
      {
        required: true,
        message: '请输入字典名（中）',
        trigger: 'blur'
      }
    ],
    type: [
      {
        required: true,
        message: '请输入字典名（英）',
        trigger: 'blur'
      }
    ],
    desc: [
      {
        required: true,
        message: '请输入描述',
        trigger: 'blur'
      }
    ]
  })

  const dictionaryData = ref([])
  const availableParentDictionaries = ref([])

  // 导入相关
  const importDrawerVisible = ref(false)
  const importJsonText = ref('')
  const importing = ref(false)
  const jsonPreviewError = ref('')
  const jsonPreview = ref(null)
  const isDragging = ref(false)
  const fileInputRef = ref(null)

  // AI 相关
  const aiDialogVisible = ref(false)
  const aiPrompt = ref('')
  const aiGenerating = ref(false)

  // 图片上传/识别相关
  const focused = ref(false)

  const handleFocus = () => {
    focused.value = true
  }
  const handleBlur = () => {
    focused.value = false
  }

  const handlePaste = (event) => {
    const items = event.clipboardData.items;
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.indexOf('image') !== -1) {
        const file = items[i].getAsFile();
        const reader = new FileReader();
        reader.onload =async (e) => {
          const base64String = e.target.result;
          const res = await llmAuto({ _file_path: base64String, mode:"dictEye" })
          if (res.code === 0) {
            aiPrompt.value = res.data.text
          }
        };
        reader.readAsDataURL(file);
      }
    }
  };

  const eyeFunc = async () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/*';

    input.onchange = (event) => {
      const file = event.target.files[0];
      if (file) {
        const reader = new FileReader();
        reader.onload = async (e) => {
          const base64String = e.target.result;

          const res = await llmAuto({ _file_path: base64String, mode:"dictEye" })
          if (res.code === 0) {
            aiPrompt.value = res.data.text
          }
        };
        reader.readAsDataURL(file);
      }
    };

    input.click();
  }



  // 监听JSON文本变化，实时预览
  watch(importJsonText, (newVal) => {
    if (!newVal.trim()) {
      jsonPreview.value = null
      jsonPreviewError.value = ''
      return
    }
    try {
      jsonPreview.value = JSON.parse(newVal)
      jsonPreviewError.value = ''
    } catch (e) {
      jsonPreviewError.value = 'JSON格式错误: ' + e.message
      jsonPreview.value = null
    }
  })

  // 查询
  const getTableData = async () => {
    const res = await getSysDictionaryList({
      name: searchName.value.trim()
    })
    if (res.code === 0) {
      dictionaryData.value = res.data
      selectID.value = res.data[0]?.ID || 0
      // 更新可选父级字典列表
      updateAvailableParentDictionaries()
    }
  }

  // 更新可选父级字典列表
  const updateAvailableParentDictionaries = () => {
    // 如果是编辑模式，排除当前字典及其子字典
    if (type.value === 'update' && formData.value.ID) {
      availableParentDictionaries.value = dictionaryData.value.filter(
        (dict) => {
          return (
            dict.ID !== formData.value.ID &&
            !isChildDictionary(dict.ID, formData.value.ID)
          )
        }
      )
    } else {
      // 创建模式，显示所有字典
      availableParentDictionaries.value = [...dictionaryData.value]
    }
  }

  // 检查是否为子字典（防止循环引用）
  const isChildDictionary = (dictId, parentId) => {
    const dict = dictionaryData.value.find((d) => d.ID === dictId)
    if (!dict || !dict.parentID) return false
    if (dict.parentID === parentId) return true
    return isChildDictionary(dict.parentID, parentId)
  }

  getTableData()

  const toDetail = (row) => {
    selectID.value = row.ID
  }

  const drawerFormVisible = ref(false)
  const type = ref('')
  const updateSysDictionaryFunc = async (row) => {
    const res = await findSysDictionary({ ID: row.ID, status: row.status })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data.resysDictionary
      drawerFormVisible.value = true
      // 更新可选父级字典列表
      updateAvailableParentDictionaries()
    }
  }
  const closeDrawer = () => {
    drawerFormVisible.value = false
    formData.value = {
      name: null,
      type: null,
      status: true,
      desc: null,
      parentID: null
    }
  }
  const deleteSysDictionaryFunc = async (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteSysDictionary({ ID: row.ID })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        getTableData()
      }
    })
  }

  const drawerForm = ref(null)
  const enterDrawer = async () => {
    drawerForm.value.validate(async (valid) => {
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createSysDictionary(formData.value)
          break
        case 'update':
          res = await updateSysDictionary(formData.value)
          break
        default:
          res = await createSysDictionary(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage.success('操作成功')
        closeDrawer()
        getTableData()
      }
    })
  }
  const openDrawer = () => {
    type.value = 'create'
    drawerForm.value && drawerForm.value.clearValidate()
    drawerFormVisible.value = true
    // 更新可选父级字典列表
    updateAvailableParentDictionaries()
  }

  const clearSearchInput = () => {
    if (!showSearchInput.value) return
    searchName.value = ''
    showSearchInput.value = false
    getTableData()
  }
  const handleCloseSearchInput = () => {
    if (!showSearchInput.value || searchName.value.trim() != '') return
    showSearchInput.value = false
  }

  const showSearchInputHandler = () => {
    showSearchInput.value = true
  }

  const handleInputKeyDown = (e) => {
    if (e.key === 'Enter' && searchName.value.trim() !== '') {
      getTableData()
    }
  }

  // 导出字典
  const exportDictionary = async (row) => {
    try {
      const res = await exportSysDictionary({ ID: row.ID })
      if (res.code === 0) {
        // 将JSON数据转换为字符串并下载
        const jsonStr = JSON.stringify(res.data, null, 2)
        const blob = new Blob([jsonStr], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `${row.type}_${row.name}_dictionary.json`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(url)
        ElMessage.success('导出成功')
      }
    } catch (error) {
      ElMessage.error('导出失败: ' + error.message)
    }
  }

  // 打开导入抽屉
  const openImportDialog = () => {
    importDrawerVisible.value = true
    importJsonText.value = ''
    jsonPreview.value = null
    jsonPreviewError.value = ''
    isDragging.value = false
  }

  // 关闭导入抽屉
  const closeImportDrawer = () => {
    importDrawerVisible.value = false
    importJsonText.value = ''
    jsonPreview.value = null
    jsonPreviewError.value = ''
    isDragging.value = false
  }

  // 处理拖拽进入
  const handleDragOver = () => {
    isDragging.value = true
  }

  // 处理拖拽离开
  const handleDragLeave = () => {
    isDragging.value = false
  }
  // 处理文件拖拽
  const handleDrop = (e) => {
    isDragging.value = false
    const files = e.dataTransfer.files
    if (files.length === 0) return

    const file = files[0]
    readJsonFile(file)
  }

  // 触发文件选择
  const triggerFileInput = () => {
    fileInputRef.value?.click()
  }

  // 处理文件选择
  const handleFileSelect = (e) => {
    const files = e.target.files
    if (files.length === 0) return

    const file = files[0]
    readJsonFile(file)
    
    // 清空input，以便可以重复选择同一文件
    e.target.value = ''
  }

  // 读取JSON文件
  const readJsonFile = (file) => {
    // 检查文件类型
    if (!file.name.endsWith('.json')) {
      ElMessage.warning('请上传 JSON 文件')
      return
    }

    // 读取文件内容
    const reader = new FileReader()
    reader.onload = (event) => {
      try {
        const content = event.target.result
        // 验证是否为有效的 JSON
        JSON.parse(content)
        importJsonText.value = content
        ElMessage.success('文件读取成功')
      } catch (error) {
        ElMessage.error('文件内容不是有效的 JSON 格式')
      }
    }
    reader.onerror = () => {
      ElMessage.error('文件读取失败')
    }
    reader.readAsText(file)
  }

  // 处理导入
  const handleImport = async () => {
    if (!importJsonText.value.trim()) {
      ElMessage.warning('请输入JSON数据')
      return
    }

    if (jsonPreviewError.value) {
      ElMessage.error('JSON格式错误，请检查后重试')
      return
    }

    try {
      importing.value = true
      const res = await importSysDictionary({ json: importJsonText.value })
      if (res.code === 0) {
        ElMessage.success('导入成功')
        closeImportDrawer()
        getTableData()
      }
    } catch (error) {
      ElMessage.error('导入失败: ' + error.message)
    } finally {
      importing.value = false
    }
  }

  // 打开 AI 对话框
  const openAiDialog = () => {
    aiDialogVisible.value = true
    aiPrompt.value = ''
  }

  // 关闭 AI 对话框
  const closeAiDialog = () => {
    aiDialogVisible.value = false
    aiPrompt.value = ''
  }

  // 处理 AI 生成
  const handleAiGenerate = async () => {
    if (!aiPrompt.value.trim()) {
      ElMessage.warning('请输入描述内容')
      return
    }
    try {
      aiGenerating.value = true
      const aiRes = await llmAuto({
        prompt: aiPrompt.value,
        mode: 'dict'
      })
      if (aiRes && aiRes.code === 0) {
        ElMessage.success('AI 生成成功')
        try {
          // 将 AI 返回的数据填充到导入文本框（支持字符串或对象）
          if (typeof aiRes.data === 'string') {
            importJsonText.value = aiRes.data
          } else {
            importJsonText.value = JSON.stringify(aiRes.data, null, 2)
          }
          // 清除可能的解析错误并打开导入抽屉
          jsonPreviewError.value = ''
          importDrawerVisible.value = true
          closeAiDialog()
        } catch (e) {
          ElMessage.error('处理 AI 返回结果失败: ' + (e.message || e))
        }
      } 
    } catch (err) {
      ElMessage.error('AI 调用失败: ' + (err.message || err))
    } finally {
      aiGenerating.value = false
    }
  }
</script>

<style scoped>
  .dictionary-page {
    display: flex;
    flex-direction: column;
    gap: 12px;
    height: 100%;
    padding: 16px;
    background: var(--na-background);
  }

  .dictionary-warning {
    flex: 0 0 auto;
    margin: 0;
  }

  .dictionary-splitter {
    flex: 1 1 auto;
    min-height: 0;
  }

  .dictionary-splitter :deep(.el-splitter-panel) {
    min-width: 0;
  }

  .dictionary-sidebar,
  .dictionary-detail {
    height: 100%;
    min-height: 0;
  }

  .dictionary-sidebar {
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .dictionary-sidebar__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 14px 12px;
    border-bottom: 1px solid var(--na-border);
  }

  .dictionary-sidebar__title-group {
    min-width: 0;
  }

  .dictionary-sidebar__title {
    display: block;
    color: var(--na-foreground);
    font-size: 14px;
    font-weight: 650;
    line-height: 1.35;
  }

  .dictionary-sidebar__count {
    display: block;
    margin-top: 2px;
    color: var(--na-muted-foreground);
    font-size: 12px;
    line-height: 1.4;
  }

  .dictionary-sidebar__actions {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    gap: 6px;
  }

  .dictionary-sidebar__actions :deep(.el-button.is-circle) {
    width: 34px;
    min-width: 34px;
    height: 34px;
  }

  .dictionary-sidebar__actions :deep(.el-button.is-active) {
    border-color: color-mix(in srgb, var(--na-primary) 34%, var(--na-border));
    background: var(--na-primary-soft);
    color: var(--na-primary);
  }

  .dictionary-sidebar__ai-button {
    width: 34px;
    min-width: 34px;
    height: 34px;
    padding: 0;
  }

  .dictionary-sidebar__search {
    padding: 12px 14px 0;
  }

  .dictionary-search-enter-active,
  .dictionary-search-leave-active {
    transition: opacity 160ms cubic-bezier(.22, 1, .36, 1), transform 160ms cubic-bezier(.22, 1, .36, 1);
  }

  .dictionary-search-enter-from,
  .dictionary-search-leave-to {
    opacity: 0;
    transform: translateY(-4px);
  }

  .dictionary-sidebar__scroll {
    flex: 1 1 auto;
    min-height: 0;
    padding: 12px 10px 14px;
  }

  .dictionary-sidebar__scroll :deep(.el-scrollbar__view) {
    min-height: 100%;
  }

  .dictionary-list {
    display: grid;
    gap: 8px;
  }

  .dictionary-list__item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    min-height: 58px;
    padding: 10px 8px 10px 12px;
    border: 1px solid transparent;
    border-radius: 10px;
    background: color-mix(in srgb, var(--na-muted) 66%, var(--na-card));
    color: var(--na-foreground);
    cursor: pointer;
    transition:
      border-color 160ms cubic-bezier(.22, 1, .36, 1),
      background-color 160ms cubic-bezier(.22, 1, .36, 1),
      color 160ms cubic-bezier(.22, 1, .36, 1);
  }

  .dictionary-list__item:hover {
    border-color: color-mix(in srgb, var(--na-primary) 18%, var(--na-border));
    background: color-mix(in srgb, var(--na-primary) 7%, var(--na-card));
  }

  .dictionary-list__item:focus-visible {
    outline: 2px solid var(--na-primary);
    outline-offset: 2px;
  }

  .dictionary-list__item.is-active {
    border-color: color-mix(in srgb, var(--na-primary) 34%, var(--na-border));
    background: var(--na-primary-soft);
    color: var(--na-primary);
  }

  .dictionary-list__item.is-child:not(.is-active) {
    margin-left: 14px;
    background: color-mix(in srgb, var(--na-muted) 46%, var(--na-card));
  }

  .dictionary-list__content {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    grid-template-areas:
      'branch name'
      'branch type';
    column-gap: 6px;
    min-width: 0;
  }

  .dictionary-list__branch {
    grid-area: branch;
    align-self: center;
    color: var(--na-muted-foreground);
    font-size: 12px;
  }

  .dictionary-list__name {
    grid-area: name;
    overflow: hidden;
    color: currentColor;
    font-size: 13px;
    font-weight: 620;
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .dictionary-list__type {
    grid-area: type;
    overflow: hidden;
    margin-top: 3px;
    color: var(--na-muted-foreground);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 11px;
    line-height: 1.25;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .dictionary-list__item.is-active .dictionary-list__type {
    color: color-mix(in srgb, var(--na-primary) 70%, var(--na-foreground));
  }

  .dictionary-list__actions {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    gap: 1px;
    opacity: .74;
    transition: opacity 160ms ease-out;
  }

  .dictionary-list__item:hover .dictionary-list__actions,
  .dictionary-list__item:focus-within .dictionary-list__actions,
  .dictionary-list__item.is-active .dictionary-list__actions {
    opacity: 1;
  }

  .dictionary-list__actions :deep(.el-button.is-link) {
    min-height: 28px;
    padding: 5px;
  }

  .dictionary-detail {
    padding-left: 12px;
  }

  .dict-box {
    height: calc(100vh - 240px);
  }

  .active {
    background-color: var(--el-color-primary) !important;
    color: #fff;
  }

  .import-drawer-content {
    padding: 0 4px;
  }

  /* 拖拽上传区域 */
  .drag-upload-area {
    border: 2px dashed var(--na-border-strong);
    border-radius: 8px;
    padding: 40px 20px;
    text-align: center;
    background-color: var(--na-card);
    transition: border-color 0.3s ease, background-color 0.3s ease;
    cursor: pointer;
  }

  .drag-upload-area:hover {
    border-color: var(--na-primary);
    background-color: var(--na-primary-soft);
  }

  .drag-upload-area.is-dragging {
    border-color: var(--na-primary);
    background-color: var(--na-primary-soft);
    transform: scale(1.02);
  }

  .upload-icon {
    font-size: 48px;
    color: #8c939d;
    margin-bottom: 16px;
  }

  .drag-upload-area.is-dragging .upload-icon {
    color: #409eff;
  }

  .upload-text {
    color: #606266;
  }

  .upload-text p {
    margin: 4px 0;
  }

  .upload-hint {
    font-size: 12px;
    color: #909399;
  }

  .json-editor-container {
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    overflow: hidden;
  }

  .json-textarea :deep(.el-textarea__inner) {
    font-family: 'Courier New', Courier, monospace;
    font-size: 13px;
    line-height: 1.5;
  }

  .json-preview {
    background-color: #f5f7fa;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    padding: 16px;
    max-height: 400px;
    overflow: auto;
  }

  .json-preview pre {
    margin: 0;
    font-family: 'Courier New', Courier, monospace;
    font-size: 13px;
    line-height: 1.5;
    white-space: pre-wrap;
    word-wrap: break-word;
  }

  .dark .drag-upload-area {
    background-color: #1d1e1f;
    border-color: #414243;
  }

  .dark .drag-upload-area:hover,
  .dark .drag-upload-area.is-dragging {
    background-color: #1a3a52;
    border-color: #409eff;
  }

  .dark .json-preview {
    background-color: #1d1e1f;
    border-color: #414243;
  }
</style>
