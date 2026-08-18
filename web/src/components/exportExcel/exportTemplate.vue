<template>
  <el-button type="primary" icon="download" :loading="downloading" @click="exportTemplateFunc"
    >下载模板</el-button
  >
</template>

<script setup>
  import { ElMessage } from 'element-plus'
  import {exportTemplate} from "@/api/exportTemplate";
  import { ref } from 'vue'

  const props = defineProps({
    templateId: {
      type: String,
      required: true
    }
  })
  const downloading = ref(false)


  const exportTemplateFunc = async () => {
    if (downloading.value) return
    if (props.templateId === '') {
      ElMessage.error('组件未设置模板ID')
      return
    }
    let baseUrl = import.meta.env.VITE_BASE_API
    if (baseUrl === "/"){
      baseUrl = ""
    }

    downloading.value = true
    try {
      const res = await exportTemplate({
        templateID: props.templateId
      })

      if(res.code === 0){
        ElMessage.success('创建导出任务成功，开始下载')
        const url = `${baseUrl}${res.data}`
        window.open(url, '_blank')
      }
    } finally {
      downloading.value = false
    }
  }
</script>
