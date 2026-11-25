<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useBrandingStore } from '@/pinia'

const props = defineProps({
  size: {
    type: Number,
    default: 2
  }
})

const brandingStore = useBrandingStore()
const showTextPlaceholder = ref(false)
const initial = computed(() => brandingStore.systemName.trim().slice(0, 1).toUpperCase() || 'M')
const logoStyle = computed(() => {
  const pixels = props.size * 16
  return { width: `${pixels}px`, height: `${pixels}px` }
})

const handleLogoError = () => {
  if (brandingStore.customLogoUrl) {
    brandingStore.useDefaultLogo()
    return
  }
  showTextPlaceholder.value = true
}

watch(() => brandingStore.logoUrl, () => {
  showTextPlaceholder.value = false
})

onMounted(() => brandingStore.loadBranding())
</script>

<template>
  <img
    v-if="!showTextPlaceholder"
    :src="brandingStore.logoUrl"
    :alt="`${brandingStore.systemName} Logo`"
    class="brand-logo-image"
    :style="logoStyle"
    @error="handleLogoError"
  />
  <span
    v-else
    class="brand-logo-fallback"
    :style="logoStyle"
    aria-hidden="true"
  >{{ initial }}</span>
</template>

<style scoped>
.brand-logo-image {
  display: block;
  flex: 0 0 auto;
  object-fit: contain;
}

.brand-logo-fallback {
  display: inline-grid;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 9px;
  background: var(--na-primary);
  color: var(--na-on-primary);
  font-size: .75rem;
  font-weight: 700;
  line-height: 1;
}
</style>
