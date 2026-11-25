<template>
  <div
    ref="viewport"
    class="wallboard-ticker"
    :style="{ '--ticker-row-size': `max(var(--wb-ticker-row-size, 44px), ${viewportHeight / visibleRows}px)` }"
    role="region"
    :aria-label="label"
    :tabindex="canLoop ? 0 : undefined"
    @mouseenter="hovered = true"
    @mouseleave="hovered = false"
    @focusin="focused = true"
    @focusout="onFocusOut"
    @touchstart.passive="touching = true"
    @touchend.passive="endTouch"
    @touchcancel.passive="endTouch"
    @wheel.passive="pauseForInteraction"
    @scroll.passive="onScroll"
  >
    <div ref="content" class="ticker-group" role="list">
      <div v-for="(item, index) in items" :key="item[itemKey]" class="ticker-row" role="listitem">
        <slot :item="item" :index="index" />
      </div>
    </div>
    <div v-if="canLoop" class="ticker-group" aria-hidden="true" inert>
      <div v-for="(item, index) in items" :key="item[itemKey]" class="ticker-row">
        <slot :item="item" :index="index" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useElementSize, useIntersectionObserver, useRafFn } from '@vueuse/core'

const props = defineProps({
  items: { type: Array, default: () => [] },
  itemKey: { type: String, required: true },
  label: { type: String, required: true },
  visibleRows: { type: Number, default: 5 },
  animated: { type: Boolean, default: true }
})
const viewport = ref(null)
const content = ref(null)
const hovered = ref(false)
const focused = ref(false)
const touching = ref(false)
const interacting = ref(false)
const visible = ref(false)
const { height: viewportHeight } = useElementSize(viewport)
const { height: contentHeight } = useElementSize(content)
const canLoop = computed(() => props.items.length > 1 && contentHeight.value > viewportHeight.value + 1 && viewportHeight.value > 0)
const running = computed(() => props.animated && canLoop.value && visible.value && !hovered.value && !focused.value && !touching.value && !interacting.value)
let position = 0
let interactionTimer

useIntersectionObserver(viewport, ([entry]) => { visible.value = entry.isIntersecting })

// Keep fractional pixels between frames: native scrollTop rounds on some displays.
// One row takes 3.2 seconds, regardless of the wallboard's display density.
const { pause, resume } = useRafFn(({ delta }) => {
  const distance = contentHeight.value / props.items.length * Math.min(delta, 64) / 3200
  position = (position + distance) % contentHeight.value
  viewport.value.scrollTop = position
}, { immediate: false })

watch(running, (value) => {
  if (value) {
    position = viewport.value.scrollTop % contentHeight.value
    resume()
  } else pause()
}, { immediate: true })

watch(canLoop, () => {
  position = 0
  if (viewport.value) viewport.value.scrollTop = 0
})

function onScroll() {
  if (!running.value) position = viewport.value.scrollTop
}

function onFocusOut(event) {
  focused.value = viewport.value.contains(event.relatedTarget)
}

function pauseForInteraction() {
  interacting.value = true
  window.clearTimeout(interactionTimer)
  interactionTimer = window.setTimeout(() => { interacting.value = false }, 1800)
}

function endTouch() {
  touching.value = false
  pauseForInteraction()
}

onBeforeUnmount(() => window.clearTimeout(interactionTimer))
</script>

<style scoped>
.wallboard-ticker {
  flex: 1;
  min-height: 0;
  margin: 4px var(--wb-pad) 8px;
  overflow: hidden auto;
  overscroll-behavior-y: contain;
  overflow-anchor: none;
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
  mask-image: linear-gradient(to bottom, transparent, black 6px, black calc(100% - 6px), transparent);
}
.wallboard-ticker:hover, .wallboard-ticker:focus-within {
  scrollbar-color: var(--wb-border) transparent;
}
.wallboard-ticker:focus-visible {
  outline: 2px solid var(--wb-accent);
  outline-offset: 2px;
  mask-image: none;
}
.ticker-row {
  display: flex;
  align-items: center;
  height: var(--ticker-row-size);
}
.ticker-row > :deep(*) { width: 100%; }
</style>
