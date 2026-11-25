<template>
  <component :is="chartComponent" v-bind="props" />
</template>

<script setup>
import { computed, defineAsyncComponent } from 'vue'

defineOptions({ inheritAttrs: false })

const props = defineProps({
  options: {
    type: Object,
    default: () => ({})
  },
  autoResize: {
    type: Boolean,
    default: true
  },
  width: {
    type: String,
    default: '100%'
  },
  height: {
    type: String,
    default: '100%'
  }
})

const chartModules = {
  bar: defineAsyncComponent(() => import('./bar.vue')),
  line: defineAsyncComponent(() => import('./line.vue')),
  pie: defineAsyncComponent(() => import('./pie.vue'))
}

const chartType = computed(() => {
  const series = Array.isArray(props.options?.series) ? props.options.series : []
  const type = series.find((item) => item?.type)?.type
  return chartModules[type] ? type : 'line'
})
const chartComponent = computed(() => chartModules[chartType.value])
</script>
