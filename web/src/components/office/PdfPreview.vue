<template>
  <iframe
    v-if="objectUrl"
    :src="objectUrl"
    class="pdf-preview-frame"
    :title="title"
    loading="lazy"
    @load="$emit('load')"
    @error="$emit('error')"
  />
</template>

<script setup>
import { onBeforeUnmount, ref, watch } from 'vue'

defineOptions({ name: 'PdfPreview' })

const props = defineProps({
  source: { type: [String, Object], default: null },
  title: { type: String, default: 'PDF 预览' }
})

defineEmits(['load', 'error'])

const objectUrl = ref('')
let ownedUrl = ''

const revokeOwnedUrl = () => {
  if (!ownedUrl) return
  URL.revokeObjectURL(ownedUrl)
  ownedUrl = ''
}

const sourceToUrl = (source) => {
  revokeOwnedUrl()
  if (!source) return ''
  if (typeof source === 'string') return source

  let blob = source
  if (source instanceof ArrayBuffer || ArrayBuffer.isView(source)) {
    blob = new Blob([source], { type: 'application/pdf' })
  }
  if (!(blob instanceof Blob)) return ''

  ownedUrl = URL.createObjectURL(blob)
  return ownedUrl
}

watch(() => props.source, (source) => {
  objectUrl.value = sourceToUrl(source)
}, { immediate: true })

onBeforeUnmount(() => {
  revokeOwnedUrl()
})
</script>

<style scoped>
.pdf-preview-frame {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 520px;
  border: 0;
  background: var(--na-card, #fff);
}
</style>
