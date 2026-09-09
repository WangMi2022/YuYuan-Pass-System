<template><span :aria-label="formattedValue">{{ displayedValue }}</span></template>

<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { finiteNumber } from './data'

const props = defineProps({
  value: { type: Number, default: 0 },
  format: { type: Function, default: (value) => Math.round(value).toLocaleString('zh-CN') },
  animated: { type: Boolean, default: true }
})
const displayed = ref(finiteNumber(props.value))
const formattedValue = computed(() => props.format(finiteNumber(props.value)))
const displayedValue = computed(() => props.format(displayed.value))
let frame = 0
watch([() => props.value, () => props.animated], ([value, animated]) => {
  cancelAnimationFrame(frame)
  const target = finiteNumber(value)
  if (!animated) { displayed.value = target; return }
  const start = displayed.value
  const startedAt = performance.now()
  const tick = (now) => {
    const progress = Math.min(1, (now - startedAt) / 650)
    displayed.value = start + (target - start) * (1 - Math.pow(1 - progress, 3))
    if (progress < 1) frame = requestAnimationFrame(tick)
  }
  frame = requestAnimationFrame(tick)
})
onBeforeUnmount(() => cancelAnimationFrame(frame))
</script>
