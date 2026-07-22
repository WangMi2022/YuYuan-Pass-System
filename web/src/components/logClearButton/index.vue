<template>
  <el-button
    type="danger"
    plain
    icon="delete-filled"
    :aria-label="`按时间清理${logName}`"
    @click="openDialog"
  >
    清理日志
  </el-button>

  <el-dialog
    v-model="visible"
    class="log-clear-dialog"
    width="min(92vw, 580px)"
    append-to-body
    destroy-on-close
    :close-on-click-modal="!clearing"
    :close-on-press-escape="!clearing"
    :show-close="!clearing"
    @closed="resetDialog"
  >
    <template #header>
      <div class="log-clear-dialog__header">
        <span class="log-clear-dialog__icon" aria-hidden="true">
          <el-icon><DeleteFilled /></el-icon>
        </span>
        <div>
          <h2>清理{{ logName }}</h2>
          <p>选择需要永久删除的日志时间范围</p>
        </div>
      </div>
    </template>

    <section class="log-clear-dialog__section" aria-labelledby="log-clear-range-title">
      <div class="log-clear-dialog__section-heading">
        <h3 id="log-clear-range-title">清理范围</h3>
        <span>默认保留最近 30 天</span>
      </div>

      <div class="log-clear-dialog__presets" role="radiogroup" aria-label="日志清理时间范围">
        <button
          v-for="preset in presets"
          :key="preset.value"
          type="button"
          role="radio"
          :class="['log-clear-dialog__preset', { 'is-active': selectedMode === preset.value }]"
          :aria-checked="selectedMode === preset.value"
          @click="selectMode(preset.value)"
        >
          {{ preset.label }}
        </button>
      </div>

      <div v-if="selectedMode === 'custom'" class="log-clear-dialog__custom-range">
        <label for="log-clear-custom-range">自定义时间段</label>
        <el-date-picker
          id="log-clear-custom-range"
          v-model="customRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          :default-time="defaultTimes"
          :disabled-date="disableFutureDate"
          @change="refreshPreview"
        />
        <p v-if="customRangeError" class="log-clear-dialog__field-error" role="alert">
          {{ customRangeError }}
        </p>
      </div>
    </section>

    <div class="log-clear-dialog__summary" aria-live="polite">
      <div>
        <span>将要删除</span>
        <strong v-if="counting">正在统计…</strong>
        <strong v-else-if="previewTotal !== null">{{ previewTotal }} 条{{ logName }}</strong>
        <strong v-else>请选择完整时间范围</strong>
      </div>
      <p>{{ scopeDescription }}</p>
      <p v-if="previewError" class="log-clear-dialog__field-error" role="alert">
        {{ previewError }}
      </p>
    </div>

    <div class="log-clear-dialog__warning" role="note">
      <el-icon aria-hidden="true"><WarningFilled /></el-icon>
      <div>
        <strong>删除后无法恢复</strong>
        <p>只会清理上方选定时间内的记录，当前列表筛选条件不会影响清理范围。</p>
      </div>
    </div>

    <template #footer>
      <el-button :disabled="clearing" @click="visible = false">取消</el-button>
      <el-button
        type="danger"
        :loading="clearing"
        :disabled="!canClear"
        @click="clearLogs"
      >
        {{ confirmButtonText }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { DeleteFilled, WarningFilled } from '@element-plus/icons-vue'
  import {
    LOG_CLEAR_PRESETS,
    buildLogClearScope,
    buildLogCountParams
  } from './range.js'

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
  const presets = LOG_CLEAR_PRESETS
  const visible = ref(false)
  const selectedMode = ref('older30')
  const scopeReferenceTime = ref(new Date())
  const customRange = ref([])
  const previewTotal = ref(null)
  const previewError = ref('')
  const counting = ref(false)
  const clearing = ref(false)
  const requestSequence = ref(0)
  const defaultTimes = [
    new Date(2000, 0, 1, 0, 0, 0),
    new Date(2000, 0, 1, 23, 59, 59)
  ]

  const formatTime = (value) => new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  }).format(new Date(value))

  const currentScope = computed(() => buildLogClearScope(
    selectedMode.value,
    customRange.value,
    scopeReferenceTime.value
  ))

  const customRangeError = computed(() => {
    if (
      selectedMode.value !== 'custom' ||
      !Array.isArray(customRange.value) ||
      customRange.value.length === 0
    ) return ''
    return currentScope.value ? '' : '结束时间必须晚于开始时间'
  })

  const scopeDescription = computed(() => {
    const scope = currentScope.value
    if (!scope) return '请选择开始时间和结束时间'
    if (scope.clearAll) return `清理全部时间范围内的${props.logName}`
    if (scope.startTime && scope.endTime) {
      return `${formatTime(scope.startTime)} 至 ${formatTime(scope.endTime)}`
    }
    return `清理 ${formatTime(scope.endTime)} 及以前的${props.logName}`
  })

  const canClear = computed(() => (
    !counting.value &&
    !clearing.value &&
    currentScope.value !== null &&
    previewTotal.value > 0 &&
    !previewError.value
  ))

  const confirmButtonText = computed(() => {
    if (previewTotal.value > 0) return `确认清理 ${previewTotal.value} 条`
    return '确认清理'
  })

  const disableFutureDate = (date) => date.getTime() > Date.now()

  const refreshPreview = async () => {
    const scope = currentScope.value
    previewError.value = ''
    if (!scope) {
      previewTotal.value = null
      return
    }

    const sequence = ++requestSequence.value
    counting.value = true
    try {
      const response = await props.countRequest(buildLogCountParams(scope))
      if (sequence !== requestSequence.value) return
      if (response.code === 0) {
        previewTotal.value = Number(response.data?.total) || 0
      } else {
        previewTotal.value = null
        previewError.value = '暂时无法统计日志数量，请稍后重试'
      }
    } catch {
      if (sequence !== requestSequence.value) return
      previewTotal.value = null
      previewError.value = '暂时无法统计日志数量，请稍后重试'
    } finally {
      if (sequence === requestSequence.value) counting.value = false
    }
  }

  const selectMode = (mode) => {
    selectedMode.value = mode
    scopeReferenceTime.value = new Date()
    previewTotal.value = null
    previewError.value = ''
    if (mode === 'custom') {
      customRange.value = []
      requestSequence.value += 1
      counting.value = false
      return
    }
    refreshPreview()
  }

  const openDialog = () => {
    visible.value = true
    selectedMode.value = 'older30'
    scopeReferenceTime.value = new Date()
    customRange.value = []
    previewTotal.value = null
    previewError.value = ''
    refreshPreview()
  }

  const resetDialog = () => {
    requestSequence.value += 1
    selectedMode.value = 'older30'
    customRange.value = []
    previewTotal.value = null
    previewError.value = ''
    counting.value = false
    clearing.value = false
  }

  const clearLogs = async () => {
    if (!canClear.value) return

    clearing.value = true
    try {
      const response = await props.clearRequest(currentScope.value)
      if (response.code === 0) {
        const deleted = response.data?.deleted ?? previewTotal.value
        visible.value = false
        emit('cleared', deleted)
        ElMessage({
          type: 'success',
          message: response.msg || `已清理 ${deleted} 条${props.logName}`
        })
      }
    } catch {
      // 请求层已统一展示接口错误。
    } finally {
      clearing.value = false
    }
  }
</script>

<style scoped>
  :global(.log-clear-dialog) {
    overflow: hidden;
    border-radius: 16px;
  }

  :global(.log-clear-dialog .el-dialog__header) {
    padding: 20px 22px 16px;
  }

  :global(.log-clear-dialog .el-dialog__body) {
    padding: 20px 22px 22px;
  }

  :global(.log-clear-dialog .el-dialog__footer) {
    padding: 14px 22px;
  }

  .log-clear-dialog__header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding-right: 28px;
  }

  .log-clear-dialog__header h2,
  .log-clear-dialog__header p,
  .log-clear-dialog__section-heading h3,
  .log-clear-dialog__section-heading span,
  .log-clear-dialog__summary p,
  .log-clear-dialog__warning p {
    margin: 0;
  }

  .log-clear-dialog__header h2 {
    color: var(--na-foreground);
    font-size: 16px;
    font-weight: 680;
    line-height: 1.4;
  }

  .log-clear-dialog__header p {
    margin-top: 2px;
    color: var(--na-muted-foreground);
    font-size: 12px;
  }

  .log-clear-dialog__icon {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    border-radius: 12px;
    background: color-mix(in srgb, var(--na-danger) 10%, var(--na-card));
    color: var(--na-danger);
    font-size: 18px;
  }

  .log-clear-dialog__section-heading {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 10px;
  }

  .log-clear-dialog__section-heading h3 {
    color: var(--na-foreground);
    font-size: 13px;
    font-weight: 650;
  }

  .log-clear-dialog__section-heading span {
    color: var(--na-muted-foreground);
    font-size: 12px;
  }

  .log-clear-dialog__presets {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 8px;
  }

  .log-clear-dialog__preset {
    min-height: 38px;
    padding: 7px 8px;
    border: 1px solid var(--na-border-strong);
    border-radius: 10px;
    background: var(--na-card);
    color: var(--na-muted-foreground);
    cursor: pointer;
    font: inherit;
    font-size: 12px;
    transition: border-color 180ms ease-out, background-color 180ms ease-out, color 180ms ease-out;
  }

  .log-clear-dialog__preset:hover {
    border-color: color-mix(in srgb, var(--na-primary) 48%, var(--na-border-strong));
    color: var(--na-foreground);
  }

  .log-clear-dialog__preset:focus-visible {
    outline: 2px solid var(--na-primary);
    outline-offset: 2px;
  }

  .log-clear-dialog__preset.is-active {
    border-color: var(--na-primary);
    background: var(--na-primary-soft);
    color: var(--na-primary);
    font-weight: 620;
  }

  .log-clear-dialog__custom-range {
    margin-top: 14px;
  }

  .log-clear-dialog__custom-range label {
    display: block;
    margin-bottom: 7px;
    color: var(--na-foreground);
    font-size: 12px;
    font-weight: 600;
  }

  .log-clear-dialog__custom-range :deep(.el-date-editor) {
    width: 100%;
  }

  .log-clear-dialog__summary {
    margin-top: 18px;
    padding: 14px 16px;
    border-radius: 12px;
    background: var(--na-muted);
  }

  .log-clear-dialog__summary > div {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
  }

  .log-clear-dialog__summary span,
  .log-clear-dialog__summary p {
    color: var(--na-muted-foreground);
    font-size: 12px;
  }

  .log-clear-dialog__summary strong {
    color: var(--na-foreground);
    font-size: 15px;
    font-weight: 680;
  }

  .log-clear-dialog__summary p {
    margin-top: 5px;
    line-height: 1.5;
  }

  .log-clear-dialog__warning {
    display: flex;
    gap: 10px;
    margin-top: 14px;
    padding: 12px 14px;
    border-radius: 12px;
    background: color-mix(in srgb, var(--na-danger) 8%, var(--na-card));
    color: var(--na-danger);
  }

  .log-clear-dialog__warning > .el-icon {
    flex: 0 0 auto;
    margin-top: 2px;
    font-size: 16px;
  }

  .log-clear-dialog__warning strong {
    font-size: 12px;
    font-weight: 680;
  }

  .log-clear-dialog__warning p {
    margin-top: 3px;
    color: color-mix(in srgb, var(--na-danger) 68%, var(--na-foreground));
    font-size: 12px;
    line-height: 1.55;
  }

  .log-clear-dialog__field-error {
    margin-top: 6px !important;
    color: var(--na-danger) !important;
    font-size: 12px;
  }

  @media (max-width: 640px) {
    .log-clear-dialog__presets {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .log-clear-dialog__section-heading,
    .log-clear-dialog__summary > div {
      align-items: flex-start;
      flex-direction: column;
      gap: 4px;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .log-clear-dialog__preset {
      transition: none;
    }
  }
</style>
