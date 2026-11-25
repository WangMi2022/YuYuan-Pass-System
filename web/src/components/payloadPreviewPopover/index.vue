<template>
  <el-popover
    v-if="hasContent"
    placement="left-start"
    :width="520"
    trigger="click"
    popper-class="payload-preview-popover"
  >
    <div class="payload-preview">
      <header class="payload-preview__header">
        <div class="payload-preview__heading">
          <span class="payload-preview__icon" aria-hidden="true">
            <el-icon><Document /></el-icon>
          </span>
          <div>
            <strong>{{ title }}</strong>
            <span>{{ contentType }}</span>
          </div>
        </div>
        <el-button
          class="payload-preview__copy"
          text
          :icon="CopyDocument"
          :aria-label="`复制${title}`"
          @click="copyContent"
        >
          复制
        </el-button>
      </header>

      <pre class="payload-preview__content"><code>{{ formattedContent }}</code></pre>
    </div>

    <template #reference>
      <el-button
        class="payload-preview__trigger"
        type="primary"
        link
        :icon="View"
        :aria-label="`查看${title}`"
      >
        查看
      </el-button>
    </template>
  </el-popover>
  <span v-else class="payload-preview__empty">无</span>
</template>

<script setup>
  import { computed } from 'vue'
  import { ElMessage } from 'element-plus'
  import { CopyDocument, Document, View } from '@element-plus/icons-vue'

  defineOptions({
    name: 'PayloadPreviewPopover'
  })

  const props = defineProps({
    title: {
      type: String,
      default: '数据详情'
    },
    value: {
      type: [String, Number, Boolean, Object, Array],
      default: ''
    }
  })

  const parsedContent = computed(() => {
    if (typeof props.value !== 'string') {
      return { isJson: true, value: props.value }
    }

    const content = props.value.trim()
    if (!content) return { isJson: false, value: '' }

    try {
      return { isJson: true, value: JSON.parse(content) }
    } catch {
      return { isJson: false, value: props.value }
    }
  })

  const hasContent = computed(() => {
    const value = props.value
    return value !== null && value !== undefined && String(value).trim() !== ''
  })

  const formattedContent = computed(() => {
    if (!parsedContent.value.isJson) return String(parsedContent.value.value)

    try {
      return JSON.stringify(parsedContent.value.value, null, 2)
    } catch {
      return String(props.value)
    }
  })

  const contentType = computed(() => parsedContent.value.isJson ? 'JSON 数据' : '文本数据')

  const fallbackCopy = (content) => {
    const textarea = document.createElement('textarea')
    textarea.value = content
    textarea.setAttribute('readonly', '')
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    const copied = document.execCommand('copy')
    textarea.remove()
    if (!copied) throw new Error('copy failed')
  }

  const copyContent = async () => {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(formattedContent.value)
      } else {
        fallbackCopy(formattedContent.value)
      }
      ElMessage.success(`${props.title}已复制`)
    } catch {
      ElMessage.error('复制失败，请手动选择内容')
    }
  }
</script>

<style scoped>
  :global(.payload-preview-popover.el-popper) {
    width: min(520px, calc(100vw - 24px)) !important;
    overflow: hidden;
    padding: 0;
    border-color: var(--na-border);
    background: var(--na-popover);
    color: var(--na-foreground);
  }

  .payload-preview__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 12px 14px;
    border-bottom: 1px solid var(--na-border);
    background: var(--na-popover);
  }

  .payload-preview__heading {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 10px;
  }

  .payload-preview__heading > div {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 1px;
  }

  .payload-preview__heading strong {
    overflow: hidden;
    color: var(--na-foreground);
    font-size: 13px;
    font-weight: 650;
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .payload-preview__heading span:not(.payload-preview__icon) {
    color: var(--na-muted-foreground);
    font-size: 11px;
    line-height: 1.35;
  }

  .payload-preview__icon {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: 9px;
    background: var(--na-primary-soft);
    color: var(--na-primary);
    font-size: 15px;
  }

  .payload-preview__copy {
    flex: 0 0 auto;
    color: var(--na-muted-foreground);
  }

  .payload-preview__copy:hover,
  .payload-preview__copy:focus-visible {
    color: var(--na-primary);
  }

  .payload-preview__content {
    max-height: min(52vh, 420px);
    margin: 0;
    overflow: auto;
    padding: 15px 16px 17px;
    background: var(--na-muted);
    color: var(--na-foreground);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
      'Liberation Mono', monospace;
    font-size: 12px;
    line-height: 1.65;
    tab-size: 2;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .payload-preview__trigger {
    min-height: 28px;
    padding: 0 4px;
  }

  .payload-preview__empty {
    color: var(--na-muted-foreground);
    font-size: 12px;
  }

  @media (max-width: 640px) {
    :global(.payload-preview-popover.el-popper) {
      width: calc(100vw - 16px) !important;
    }

    .payload-preview__content {
      max-height: 46vh;
      padding: 13px 14px 15px;
    }
  }
</style>
