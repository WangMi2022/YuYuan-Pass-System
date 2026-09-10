<template>
  <div
    ref="containerRef"
    class="hero-canvas-container"
    aria-hidden="true"
    @pointermove="handlePointerMove"
    @pointerleave="handlePointerLeave"
  />
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import * as THREE from 'three'
import { useThreeScene } from './useThreeScene'
import { useAppStore } from '@/pinia/modules/app'

const props = defineProps({
  particleCountX: { type: Number, default: 46 },
  particleCountY: { type: Number, default: 22 },
  separation: { type: Number, default: 8.5 }
})

const appStore = useAppStore()
const containerRef = ref(null)

let geometry = null
let material = null
let pointsMesh = null
let texture = null

const pointer = { x: 0, y: 0, targetX: 0, targetY: 0, active: false }

function createCircleTexture() {
  const canvas = document.createElement('canvas')
  canvas.width = 32
  canvas.height = 32
  const ctx = canvas.getContext('2d')
  if (!ctx) return null

  const gradient = ctx.createRadialGradient(16, 16, 0, 16, 16, 16)
  gradient.addColorStop(0, 'rgba(255, 255, 255, 1)')
  gradient.addColorStop(0.25, 'rgba(255, 255, 255, 0.85)')
  gradient.addColorStop(0.6, 'rgba(255, 255, 255, 0.25)')
  gradient.addColorStop(1, 'rgba(255, 255, 255, 0)')

  ctx.fillStyle = gradient
  ctx.beginPath()
  ctx.arc(16, 16, 16, 0, Math.PI * 2)
  ctx.fill()

  const tex = new THREE.CanvasTexture(canvas)
  tex.needsUpdate = true
  return tex
}

const isDarkMode = computed(() => {
  if (typeof document !== 'undefined') {
    return document.documentElement.classList.contains('dark') || appStore.config.darkMode === 'dark'
  }
  return false
})

function getThemeColor() {
  if (appStore.config.primaryColor) {
    return appStore.config.primaryColor
  }
  if (typeof window !== 'undefined') {
    const val = getComputedStyle(document.documentElement).getPropertyValue('--na-primary').trim()
    if (val) return val
  }
  return '#6366F1'
}

function updateMaterialStyle() {
  if (!material) return
  const color = getThemeColor()
  material.color.set(color)
  material.opacity = isDarkMode.value ? 0.65 : 0.32
  material.size = isDarkMode.value ? 4.8 : 4.0
}

function handlePointerMove(e) {
  if (!containerRef.value) return
  const rect = containerRef.value.getBoundingClientRect()
  pointer.targetX = ((e.clientX - rect.left) / rect.width) * 2 - 1
  pointer.targetY = -(((e.clientY - rect.top) / rect.height) * 2 - 1)
  pointer.active = true
}

function handlePointerLeave() {
  pointer.active = false
  pointer.targetX = 0
  pointer.targetY = 0
}

const { isAvailable } = useThreeScene(containerRef, {
  fov: 48,
  cameraPos: [0, 52, 98],
  lookAt: [0, 4, 0],
  onInit({ scene }) {
    const countX = props.particleCountX
    const countY = props.particleCountY
    const sep = props.separation
    const total = countX * countY

    const positions = new Float32Array(total * 3)
    const offsetX = ((countX - 1) * sep) / 2
    const offsetZ = ((countY - 1) * sep) / 2

    let idx = 0
    for (let ix = 0; ix < countX; ix++) {
      for (let iy = 0; iy < countY; iy++) {
        positions[idx] = ix * sep - offsetX
        positions[idx + 1] = 0
        positions[idx + 2] = iy * sep - offsetZ
        idx += 3
      }
    }

    geometry = new THREE.BufferGeometry()
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3))

    texture = createCircleTexture()
    material = new THREE.PointsMaterial({
      size: isDarkMode.value ? 4.8 : 4.0,
      map: texture,
      transparent: true,
      opacity: isDarkMode.value ? 0.65 : 0.32,
      depthWrite: false,
      blending: THREE.NormalBlending
    })
    updateMaterialStyle()

    pointsMesh = new THREE.Points(geometry, material)
    scene.add(pointsMesh)
  },
  onRender({ elapsed }) {
    if (!geometry || !material) return

    // Smooth pointer damping
    pointer.x += (pointer.targetX - pointer.x) * 0.05
    pointer.y += (pointer.targetY - pointer.y) * 0.05

    const positions = geometry.attributes.position.array
    const countX = props.particleCountX
    const countY = props.particleCountY
    const time = elapsed * 0.95

    let idx = 0
    for (let ix = 0; ix < countX; ix++) {
      for (let iy = 0; iy < countY; iy++) {
        const u = ix / countX
        const v = iy / countY

        // Harmonious 3D wave interference
        const waveA = Math.sin(u * 6.2 + time * 1.3) * 4.6
        const waveB = Math.cos(v * 4.8 + time * 1.1) * 3.8
        const waveC = Math.sin((u + v) * 3.6 + time * 0.8) * 2.8

        // Subtle interactive disturbance from pointer
        let pointerShift = 0
        if (pointer.active || Math.abs(pointer.x) > 0.01 || Math.abs(pointer.y) > 0.01) {
          const px = (u - 0.5) * 2
          const py = (v - 0.5) * 2
          const distSq = (px - pointer.x) * (px - pointer.x) + (py - pointer.y) * (py - pointer.y)
          if (distSq < 1.0) {
            pointerShift = Math.cos(Math.sqrt(distSq) * Math.PI * 0.5) * 6.5
          }
        }

        positions[idx + 1] = waveA + waveB + waveC + pointerShift
        idx += 3
      }
    }
    geometry.attributes.position.needsUpdate = true
  }
})

watch(() => appStore.config.primaryColor, () => {
  updateMaterialStyle()
})

watch(() => appStore.config.darkMode, () => {
  updateMaterialStyle()
})

defineExpose({ isAvailable })
</script>

<style scoped>
.hero-canvas-container {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  pointer-events: auto;
  z-index: 0;
}
.hero-canvas-container :deep(canvas) {
  display: block;
  width: 100% !important;
  height: 100% !important;
  outline: none;
}
@media (prefers-reduced-motion: reduce) {
  .hero-canvas-container {
    opacity: 0.18;
  }
}
</style>
