<template><div ref="host" class="wallboard-chart" role="img" :aria-label="label" /></template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { init, use } from 'echarts/core'
import { BarChart, LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([BarChart, LineChart, GridComponent, TooltipComponent, CanvasRenderer])
const props = defineProps({ option: { type: Object, required: true }, label: { type: String, required: true }, animated: { type: Boolean, default: true } })
const host = ref(null)
let chart
let observer
function update() {
  chart?.setOption({ ...props.option, animation: props.animated, animationDuration: 650, animationDurationUpdate: 500, animationEasing: 'cubicOut' }, { notMerge: true })
}
watch([() => props.option, () => props.animated], update, { deep: true, flush: 'post' })
onMounted(() => {
  chart = init(host.value, null, { devicePixelRatio: Math.min(window.devicePixelRatio || 1, 2) })
  update()
  observer = new ResizeObserver(() => chart?.resize())
  observer.observe(host.value)
})
onBeforeUnmount(() => { observer?.disconnect(); chart?.dispose(); chart = null })
</script>

<style scoped>
.wallboard-chart { width: 100%; height: 100%; min-width: 0; min-height: 0; }
</style>
