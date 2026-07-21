<template>
  <header class="na-page-header" :aria-labelledby="headingId">
    <div>
      <p v-if="kicker" class="na-page-kicker">{{ kicker }}</p>
      <h1 :id="headingId" class="na-page-title">{{ title }}</h1>
      <p v-if="description || $slots.description" class="na-page-description">
        <slot name="description">{{ description }}</slot>
      </p>
    </div>
    <div v-if="$slots.actions" class="na-page-actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup>
import { computed, useId } from 'vue'

defineOptions({ name: 'AppPageHeader' })

const props = defineProps({
  title: { type: String, required: true },
  titleId: { type: String, default: '' },
  kicker: { type: String, default: '' },
  description: { type: String, default: '' }
})

const generatedId = useId()
const headingId = computed(() => props.titleId || `page-title-${generatedId}`)
</script>
