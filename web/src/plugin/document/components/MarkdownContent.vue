<template>
  <article v-html="html" />
</template>

<script setup>
import { computed } from 'vue'
import { marked } from 'marked'

const props = defineProps({
  source: {
    type: String,
    default: ''
  }
})

const html = computed(() => marked.parse(props.source || ''))
</script>
