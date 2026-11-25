<template>
  <div class="na-workspace na-workspace--tool form-designer-container">
    <div class="form-builder-toolbar">
      <div>
        <h2>表单结构</h2>
        <p>字段配置只在本地生成 Vue 模板，不加载第三方富文本设计器。</p>
      </div>
      <div class="form-builder-actions">
        <el-button :icon="Plus" @click="addField">添加字段</el-button>
        <el-button type="primary" @click="exportVueTemplate">生成 Vue 模板</el-button>
      </div>
    </div>

    <el-table :data="fieldRules" height="calc(100vh - 250px)" border>
      <el-table-column type="index" width="56" label="#" />
      <el-table-column label="字段标题" min-width="180">
        <template #default="{ row }"><el-input v-model="row.title" maxlength="40" /></template>
      </el-table-column>
      <el-table-column label="字段名" min-width="180">
        <template #default="{ row }"><el-input v-model="row.field" maxlength="60" /></template>
      </el-table-column>
      <el-table-column label="控件类型" min-width="170">
        <template #default="{ row }">
          <el-select v-model="row.type" class="w-full">
            <el-option v-for="option in fieldTypes" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="必填" width="90" align="center">
        <template #default="{ row }"><el-switch v-model="row.$required" /></template>
      </el-table-column>
      <el-table-column label="操作" width="80" align="center">
        <template #default="{ $index }">
          <el-button circle text type="danger" :icon="Delete" title="删除字段" @click="removeField($index)" />
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" title="生成的 Vue 模板代码" width="70%" top="5vh">
      <el-input 
        type="textarea" 
        :rows="25" 
        v-model="vueCode" 
        readonly 
        class="code-input"
        resize="none"
      />
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="copyCode">一键复制</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { Delete, Plus } from '@element-plus/icons-vue'

  defineOptions({
    name: 'FormGenerator'
  })

  const dialogVisible = ref(false)
  const vueCode = ref('')

  const fieldTypes = [
    { label: '单行输入', value: 'input' },
    { label: '数字输入', value: 'inputNumber' },
    { label: '下拉选择', value: 'select' },
    { label: '单选', value: 'radio' },
    { label: '多选', value: 'checkbox' },
    { label: '开关', value: 'switch' },
    { label: '日期', value: 'datePicker' },
    { label: '时间', value: 'timePicker' },
    { label: '滑块', value: 'slider' },
    { label: '颜色', value: 'colorPicker' }
  ]
  const fieldRules = ref([
    { type: 'input', field: 'name', title: '名称', value: '', props: { clearable: true }, $required: true }
  ])
  const formOptions = { form: { labelWidth: '120px', size: 'default', labelPosition: 'right' } }

  const addField = () => {
    const index = fieldRules.value.length + 1
    fieldRules.value.push({ type: 'input', field: `field${index}`, title: `字段 ${index}`, value: '', props: { clearable: true }, $required: false })
  }

  const removeField = (index) => fieldRules.value.splice(index, 1)

  const kebabCase = (str) => {
    return str.replace(/([A-Z])/g, '-$1').toLowerCase()
  }

  const generateVueCode = (rules, options) => {
    let formDataInit = []
    let formRules = []

    const parseRule = (rule) => {
      if (rule.type === 'row') {
        const propsStr = rule.props ? Object.entries(rule.props).map(([k, v]) => `:${k}="${v}"`).join(' ') : ''
        let childrenStr = rule.children ? rule.children.map(c => parseRule(c)).join('\n') : ''
        return `\n    <el-row ${propsStr}>${childrenStr}\n    </el-row>`
      }
      if (rule.type === 'col') {
        const propsStr = rule.props ? Object.entries(rule.props).map(([k, v]) => `:${k}="${v}"`).join(' ') : ''
        let childrenStr = rule.children ? rule.children.map(c => parseRule(c)).join('\n') : ''
        return `\n      <el-col ${propsStr}>${childrenStr}\n      </el-col>`
      }

      if (!rule.field) return ''

      let tag = rule.type
      
      const typeMap = {
        input: 'el-input',
        inputNumber: 'el-input-number',
        select: 'el-select',
        radio: 'el-radio-group',
        checkbox: 'el-checkbox-group',
        switch: 'el-switch',
        timePicker: 'el-time-picker',
        datePicker: 'el-date-picker',
        slider: 'el-slider',
        rate: 'el-rate',
        colorPicker: 'el-color-picker',
        cascader: 'el-cascader',
        upload: 'el-upload'
      }

      const elTag = typeMap[tag] || (tag.startsWith('el-') ? tag : `el-${tag}`)

      let propsStr = ''
      if (rule.props) {
        for (const [key, value] of Object.entries(rule.props)) {
          if (value === null || value === undefined) continue
          if (typeof value === 'boolean') {
            propsStr += value ? ` ${kebabCase(key)}` : ` :${kebabCase(key)}="false"`
          } else if (typeof value === 'string') {
            propsStr += ` ${kebabCase(key)}="${value}"`
          } else {
            propsStr += ` :${kebabCase(key)}='${JSON.stringify(value)}'`
          }
        }
      }

      let innerContent = ''
      if (rule.options && Array.isArray(rule.options)) {
        if (tag === 'select') {
          innerContent = rule.options.map(opt => `\n        <el-option label="${opt.label}" value="${opt.value}" />`).join('') + '\n      '
        } else if (tag === 'radio') {
          innerContent = rule.options.map(opt => `\n        <el-radio label="${opt.value}">${opt.label}</el-radio>`).join('') + '\n      '
        } else if (tag === 'checkbox') {
          innerContent = rule.options.map(opt => `\n        <el-checkbox label="${opt.value}">${opt.label}</el-checkbox>`).join('') + '\n      '
        }
      }

      let initVal = rule.value !== undefined ? rule.value : (tag === 'checkbox' ? [] : null)
      formDataInit.push(`  ${rule.field}: ${JSON.stringify(initVal)}`)

      if (rule.$required || (rule.effect && rule.effect.required)) {
        formRules.push(`  ${rule.field}: [{ required: true, message: '${rule.title}不能为空', trigger: 'blur' }]`)
      } else if (rule.validate) {
        formRules.push(`  ${rule.field}: ${JSON.stringify(rule.validate)}`)
      }

      return `
    <el-form-item label="${rule.title}" prop="${rule.field}">
      <${elTag} v-model="formData.${rule.field}"${propsStr}>${innerContent}</${elTag}>
    </el-form-item>`
    }

    const formItems = rules.map(parseRule).join('')

    const formConfig = options.form || {}
    let formPropsStr = []
    if (formConfig.labelWidth) formPropsStr.push(`label-width="${formConfig.labelWidth}"`)
    if (formConfig.size) formPropsStr.push(`size="${formConfig.size}"`)
    if (formConfig.labelPosition) formPropsStr.push(`label-position="${formConfig.labelPosition}"`)
    if (formConfig.hideRequiredAsterisk) formPropsStr.push(`hide-required-asterisk`)

    // 8. 拼装成标准的 <template> 和 <script setup> 闭环代码
    return `<template>
  <div>
    <el-form ref="formRef" :model="formData" :rules="rules" ${formPropsStr.join(' ')}>
${formItems}
      <el-form-item>
        <el-button type="primary" @click="submitForm">提交</el-button>
        <el-button @click="resetForm">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

const formRef = ref(null)

const formData = reactive({
${formDataInit.join(',\n')}
})

const rules = reactive({
${formRules.join(',\n')}
})

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      ElMessage.success('表单校验通过，准备提交')
      console.log('提交的数据: ', formData)
    } else {
      ElMessage.error('表单校验失败')
    }
  })
}

const resetForm = () => {
  if (!formRef.value) return
  formRef.value.resetFields()
}
</scr${'ipt'}>
`
  }

  const exportVueTemplate = () => {
    if (!fieldRules.value.length) {
      ElMessage.warning('请至少添加一个字段')
      return
    }
    const invalid = fieldRules.value.some((rule) => !rule.title?.trim() || !/^[A-Za-z_$][\w$]*$/.test(rule.field || ''))
    if (invalid) {
      ElMessage.warning('字段标题不能为空，字段名必须是合法 JavaScript 标识符')
      return
    }
    vueCode.value = generateVueCode(fieldRules.value, formOptions)
    dialogVisible.value = true
  }

  const copyCode = async () => {
    try {
      await navigator.clipboard.writeText(vueCode.value)
      ElMessage.success('代码已成功复制到剪贴板！')
      dialogVisible.value = false
    } catch (err) {
      ElMessage.error('复制失败，请手动选择复制')
    }
  }
</script>

<style scoped>
.form-builder-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.form-builder-toolbar h2 {
  margin: 0 0 4px;
  font-size: 18px;
}

.form-builder-toolbar p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.form-builder-actions {
  display: flex;
  gap: 8px;
}
</style>
