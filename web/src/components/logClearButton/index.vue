<template>
  <el-button
    type="danger"
    plain
    icon="delete-filled"
    :loading="clearing"
    :disabled="clearing"
    :aria-label="`清空全部${logName}`"
    @click="clearLogs"
  >
    清空日志
  </el-button>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({
    name: 'LogClearButton'
  })

  const props = defineProps({
    logName: {
      type: String,
      required: true
    },
    countRequest: {
      type: Function,
      required: true
    },
    clearRequest: {
      type: Function,
      required: true
    }
  })

  const emit = defineEmits(['cleared'])
  const clearing = ref(false)

  const clearLogs = async () => {
    if (clearing.value) return

    clearing.value = true
    try {
      const countRes = await props.countRequest({ page: 1, pageSize: 1 })
      if (countRes.code !== 0) return

      const total = Number(countRes.data.total) || 0
      if (total === 0) {
        ElMessage({ type: 'info', message: `暂无${props.logName}需要清空` })
        return
      }

      const confirmed = await ElMessageBox.confirm(
        `将永久删除全部 ${total} 条${props.logName}，包括当前筛选条件之外的记录。此操作不可恢复。`,
        `清空全部${props.logName}？`,
        {
          confirmButtonText: '确认清空',
          cancelButtonText: '取消',
          confirmButtonClass: 'el-button--danger',
          type: 'warning'
        }
      ).catch(() => false)

      if (!confirmed) return

      const clearRes = await props.clearRequest()
      if (clearRes.code === 0) {
        emit('cleared', clearRes.data?.deleted ?? total)
        ElMessage({
          type: 'success',
          message: clearRes.msg || `${props.logName}已清空`
        })
      }
    } catch {
      // 请求层已统一展示错误信息
    } finally {
      clearing.value = false
    }
  }
</script>
