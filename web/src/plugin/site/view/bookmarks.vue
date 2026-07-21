<template>
  <main class="na-page na-page--list site-page">
    <AppPageHeader
      title-id="site-title"
      title="站点收藏"
      description="集中收藏工作中常用的 HTTP/HTTPS 网页站点，点击卡片即可新窗口跳转。"
    >
      <template #actions>
        <el-button type="primary" size="large" :icon="Plus" @click="openCreate">新增站点</el-button>
      </template>
    </AppPageHeader>

    <section class="na-panel filter-panel" aria-label="站点筛选">
      <el-form :model="search" @keyup.enter="submitSearch">
        <div class="filter-grid">
          <el-form-item label="关键词">
            <el-input v-model="search.keyword" clearable placeholder="搜索名称、地址、分类或说明" :prefix-icon="Search" />
          </el-form-item>
          <el-form-item label="分类">
            <el-select v-model="search.category" clearable placeholder="全部分类">
              <el-option v-for="item in categoryOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="search.enabled" clearable placeholder="全部状态">
              <el-option label="启用" :value="true" />
              <el-option label="停用" :value="false" />
            </el-select>
          </el-form-item>
          <div class="filter-actions">
            <el-button type="primary" :icon="Search" @click="submitSearch">查询</el-button>
            <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
          </div>
        </div>
      </el-form>
    </section>

    <section class="na-panel site-panel">
      <header class="na-panel-header panel-header">
        <div>
          <h2>工作站点</h2>
          <span>共 {{ total }} 个站点</span>
        </div>
        <el-button :icon="Refresh" text @click="loadSites">刷新</el-button>
      </header>

      <div v-loading="loading" class="site-grid">
        <article
          v-for="row in tableData"
          :key="row.ID"
          class="site-card"
          :class="{ disabled: !row.enabled }"
          role="button"
          :tabindex="row.enabled ? 0 : -1"
          :aria-disabled="!row.enabled"
          @click="openSite(row)"
          @keyup.enter="openSite(row)"
          @keyup.space.prevent="openSite(row)"
        >
          <header class="site-card-top">
            <div class="site-identity">
              <span class="site-icon" :style="{ background: row.color || '#6366f1' }">
                <el-icon><Link /></el-icon>
              </span>
              <h3 :title="row.name">{{ row.name }}</h3>
            </div>
            <span class="site-status" :class="row.enabled ? 'enabled' : 'disabled'">{{ row.enabled ? '启用' : '停用' }}</span>
          </header>
          <div class="site-main">
            <p class="site-url" :title="row.url">{{ row.url }}</p>
            <p class="site-desc">{{ row.description || '暂无说明，点击打开站点。' }}</p>
          </div>
          <footer class="site-card-footer">
            <div class="site-meta">
              <el-tag size="small" effect="plain">{{ row.category || '常用站点' }}</el-tag>
              <span class="visit-count"><strong>{{ row.visitCount || 0 }}</strong> 次访问</span>
            </div>
            <div class="site-actions" @click.stop>
              <el-button type="primary" plain :icon="Position" @click="openSite(row)">打开</el-button>
              <el-button plain :icon="Edit" @click="openEdit(row)">编辑</el-button>
              <el-button type="danger" plain :icon="Delete" @click="removeSite(row)">删除</el-button>
            </div>
          </footer>
        </article>

        <el-empty v-if="!loading && !tableData.length" class="site-empty" description="暂无收藏站点">
          <el-button type="primary" :icon="Plus" @click="openCreate">新增第一个站点</el-button>
        </el-empty>
      </div>

      <footer class="na-pagination pagination-wrap">
        <el-pagination
          v-model:current-page="search.page"
          v-model:page-size="search.pageSize"
          :page-sizes="[12, 24, 48, 96]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadSites"
          @size-change="sizeChanged"
        />
      </footer>
    </section>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑站点' : '新增站点'" width="min(92vw, 620px)" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="rules" label-position="top" class="site-form">
        <el-form-item label="站点名称" prop="name">
          <el-input v-model="formData.name" maxlength="120" placeholder="例如：监控大屏 / 内部系统 / 文档站" />
        </el-form-item>
        <el-form-item label="站点地址" prop="url">
          <el-input v-model="formData.url" maxlength="800" placeholder="https://example.com 或 http://example.com" />
        </el-form-item>
        <div class="form-grid">
          <el-form-item label="分类" prop="category">
            <el-input v-model="formData.category" maxlength="80" placeholder="例如：运维 / 文档 / 业务系统" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="formData.sort" :min="0" :max="9999" controls-position="right" />
          </el-form-item>
        </div>
        <div class="form-grid compact">
          <el-form-item label="标识颜色">
            <el-color-picker v-model="formData.color" show-alpha />
          </el-form-item>
          <el-form-item label="是否启用">
            <el-switch v-model="formData.enabled" active-text="启用" inactive-text="停用" />
          </el-form-item>
        </div>
        <el-form-item label="说明">
          <el-input v-model="formData.description" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="补充该站点用途、访问说明或负责人" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="saveSite">保存</el-button>
        </div>
      </template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Link, Plus, Position, Refresh, Search } from '@element-plus/icons-vue'
import { createSite, deleteSite, getSiteCategories, getSiteList, updateSite, visitSite } from '@/plugin/site/api/site'
import AppPageHeader from '@/components/page/AppPageHeader.vue'
import { usePagedList } from '@/hooks/usePagedList'

const saving = ref(false)
const dialogVisible = ref(false)
const editing = ref(false)
const formRef = ref(null)
const categories = ref([])
const formData = reactive({ ID: 0, name: '', url: '', category: '常用站点', description: '', color: '#6366f1', sort: 0, enabled: true })

const categoryOptions = computed(() => Array.from(new Set([...(categories.value || []), ...tableData.value.map((item) => item.category).filter(Boolean)])))

const validateHttpUrl = (_rule, value, callback) => {
  const text = String(value || '').trim()
  if (!text) return callback(new Error('请输入站点地址'))
  try {
    const parsed = new URL(text)
    if (!['http:', 'https:'].includes(parsed.protocol)) return callback(new Error('仅支持 http 或 https 地址'))
  } catch {
    return callback(new Error('站点地址格式不正确'))
  }
  callback()
}

const rules = {
  name: [{ required: true, message: '请输入站点名称', trigger: 'blur' }],
  url: [{ required: true, validator: validateHttpUrl, trigger: 'blur' }],
  category: [{ required: true, message: '请输入分类', trigger: 'blur' }]
}

const resetForm = () => {
  Object.assign(formData, { ID: 0, name: '', url: '', category: '常用站点', description: '', color: '#6366f1', sort: 0, enabled: true })
}

const loadCategories = async () => {
  const res = await getSiteCategories()
  if (res.code === 0) categories.value = res.data || []
}

const {
  search,
  items: tableData,
  total,
  loading,
  load: loadSites,
  submit: submitSearch,
  reset: resetSearch,
  changePageSize: sizeChanged
} = usePagedList({
  defaults: { page: 1, pageSize: 12, keyword: '', category: '', enabled: '' },
  request: (params) => {
    if (params.enabled === '') delete params.enabled
    return getSiteList(params)
  }
})

const openCreate = () => {
  editing.value = false
  resetForm()
  dialogVisible.value = true
}

const openEdit = (row) => {
  editing.value = true
  Object.assign(formData, {
    ID: row.ID,
    name: row.name || '',
    url: row.url || '',
    category: row.category || '常用站点',
    description: row.description || '',
    color: row.color || '#6366f1',
    sort: row.sort || 0,
    enabled: !!row.enabled
  })
  dialogVisible.value = true
}

const saveSite = async () => {
  await formRef.value?.validate()
  saving.value = true
  try {
    const payload = { ...formData }
    const res = editing.value ? await updateSite(payload) : await createSite(payload)
    if (res.code === 0) {
      ElMessage.success(editing.value ? '站点已更新' : '站点已添加')
      dialogVisible.value = false
      await Promise.all([loadSites(), loadCategories()])
    }
  } finally {
    saving.value = false
  }
}

const openSite = async (row) => {
  if (!row?.url || !row.enabled) return
  const win = window.open('about:blank', '_blank')
  try {
    await visitSite({ id: row.ID })
    row.visitCount = Number(row.visitCount || 0) + 1
  } finally {
    if (win) {
      win.location.href = row.url
    } else {
      window.open(row.url, '_blank')
    }
  }
}

const removeSite = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除站点“${row.name}”吗？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  const res = await deleteSite({ id: row.ID })
  if (res.code === 0) {
    ElMessage.success('站点已删除')
    await Promise.all([loadSites(), loadCategories()])
  }
}

onMounted(async () => {
  await Promise.all([loadSites(), loadCategories()])
})
</script>

<style scoped lang="scss">
.panel-header h2 {
  margin: 0;
  color: var(--el-text-color-primary);
}
.panel-header span {
  color: var(--el-text-color-secondary);
}

.filter-panel {
  padding: 12px 16px 0;
}

.filter-grid {
  display: grid;
  grid-template-columns: minmax(240px, 1.5fr) minmax(160px, 1fr) minmax(140px, 0.8fr) auto;
  gap: 10px;
  align-items: end;
}

.filter-panel :deep(.el-form-item) {
  margin-bottom: 12px;
}

.filter-actions {
  display: flex;
  gap: 8px;
  padding-bottom: 12px;
}

.site-panel {
  overflow: hidden;
}

.panel-header {
  gap: 12px;
}

.panel-header h2 {
  margin-bottom: 2px;
  font-size: 16px;
}

.site-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 14px;
  min-height: 240px;
  padding: 14px 16px;
}

.site-card {
  display: flex;
  min-width: 0;
  min-height: 202px;
  flex-direction: column;
  border: 1px solid var(--na-border-strong, var(--el-border-color));
  border-radius: 12px;
  background: var(--el-bg-color);
  padding: 14px;
  cursor: pointer;
  transition: border-color 150ms ease, background-color 150ms ease;
}

.site-card:hover,
.site-card:focus-visible {
  border-color: color-mix(in srgb, var(--na-primary) 34%, var(--na-border-strong));
  background: color-mix(in srgb, var(--na-primary) 2%, var(--na-card));
}

.site-card.disabled {
  opacity: 0.58;
  cursor: not-allowed;
}

.site-card-top,
.site-card-footer,
.site-identity,
.site-meta,
.site-actions {
  display: flex;
  align-items: center;
}

.site-card-top {
  justify-content: space-between;
  gap: 12px;
}

.site-identity {
  min-width: 0;
  gap: 10px;
}

.site-identity h3 {
  overflow: hidden;
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 16px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.site-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  color: #fff;
  font-size: 16px;
}

.site-status {
  flex: 0 0 auto;
  border-radius: 999px;
  padding: 3px 8px;
  font-size: 12px;
  font-weight: 650;
}

.site-status.enabled {
  background: var(--na-success-soft);
  color: var(--na-success);
}

.site-status.disabled {
  background: var(--na-muted);
  color: var(--na-muted-foreground);
}

.site-main {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  padding: 12px 0;
}

.site-url {
  overflow: hidden;
  margin: 0 0 8px;
  border-radius: 6px;
  background: var(--el-fill-color-lighter);
  padding: 6px 8px;
  color: var(--el-color-primary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.site-desc {
  display: -webkit-box;
  min-height: 38px;
  overflow: hidden;
  margin: 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.55;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.site-card-footer {
  align-items: stretch;
  flex-direction: column;
  gap: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.site-meta {
  justify-content: space-between;
  min-width: 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.visit-count strong {
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-variant-numeric: tabular-nums;
}

.site-actions {
  width: 100%;
  gap: 8px;
}

.site-actions :deep(.el-button) {
  min-width: 0;
  height: 32px;
  flex: 1;
  margin-left: 0;
  padding: 0 8px;
  font-size: 12px;
}

.site-empty {
  grid-column: 1 / -1;
}

.pagination-wrap {
  padding-bottom: 14px;
}

.form-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 160px;
  gap: 14px;
}

.form-grid.compact {
  grid-template-columns: 160px minmax(0, 1fr);
  align-items: center;
}

.site-form :deep(.el-input-number),
.site-form :deep(.el-select) {
  width: 100%;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 1100px) {
  .filter-grid {
    grid-template-columns: repeat(2, minmax(180px, 1fr));
  }
}

@media (max-width: 640px) {
  .panel-header {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-grid,
  .form-grid,
  .form-grid.compact {
    grid-template-columns: 1fr;
  }

  .site-grid {
    grid-template-columns: 1fr;
  }

}
</style>
