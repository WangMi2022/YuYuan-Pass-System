<template>
  <PdfPreview :source="pdf" @load="renderedHandler" @error="errorHandler" />
</template>
<script>
  export default {
    name: 'Pdf'
  }
</script>
<script setup>
  import { ref, watch } from 'vue'

  import PdfPreview from './PdfPreview.vue'
  const props = defineProps({
    modelValue: {
      type: String,
      default: () => ''
    }
  })
  const pdf = ref(null)
  watch(
    () => props.modelValue,
    (val) => (pdf.value = val),
    { immediate: true }
  )
  const renderedHandler = () => {}
  const errorHandler = () => {}
</script>
