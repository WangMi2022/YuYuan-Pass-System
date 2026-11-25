<template>
  <VCharts :option="options" :autoresize="autoResize" :style="{ width, height }" />
</template>

<script setup>
import { BarChart } from 'echarts/charts'
import { registerChart, VCharts } from './base'

registerChart(BarChart)

defineProps({
  options: { type: Object, default: () => ({}) },
  autoResize: { type: Boolean, default: true },
  width: { type: String, default: '100%' },
  height: { type: String, default: '100%' }
})
</script>
