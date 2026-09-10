<template>
  <div
    ref="containerRef"
    class="hero-canvas-container"
    aria-hidden="true"
  />
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as THREE from 'three'
import { useThreeScene } from './useThreeScene'
import { useAppStore } from '@/pinia/modules/app'

const props = defineProps({
  particleCountX: { type: Number, default: 52 },
  particleCountY: { type: Number, default: 24 },
  separation: { type: Number, default: 7.8 }
})

const appStore = useAppStore()
const containerRef = ref(null)

let geometry = null
let material = null
let pointsMesh = null
let lineGeometry = null
let lineMaterial = null
let lineMesh = null
let motesGeometry = null
let motesMaterial = null
let motesMesh = null
let texture = null
let hostElement = null

const pointer = { x: 0, y: 0, targetX: 0, targetY: 0, active: false }

function createBloomTexture() {
  const canvas = document.createElement('canvas')
  canvas.width = 64
  canvas.height = 64
  const ctx = canvas.getContext('2d')
  if (!ctx) return null

  const gradient = ctx.createRadialGradient(32, 32, 0, 32, 32, 32)
  gradient.addColorStop(0, 'rgba(255, 255, 255, 1)')
  gradient.addColorStop(0.18, 'rgba(255, 255, 255, 0.95)')
  gradient.addColorStop(0.42, 'rgba(255, 255, 255, 0.45)')
  gradient.addColorStop(0.72, 'rgba(255, 255, 255, 0.12)')
  gradient.addColorStop(1, 'rgba(255, 255, 255, 0)')

  ctx.fillStyle = gradient
  ctx.beginPath()
  ctx.arc(32, 32, 32, 0, Math.PI * 2)
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
  const color = getThemeColor()
  const dark = isDarkMode.value

  if (material) {
    material.color.set(color)
    material.opacity = dark ? 0.88 : 0.65
    material.size = dark ? 5.8 : 4.8
  }

  if (lineMaterial) {
    lineMaterial.color.set(color)
    lineMaterial.opacity = dark ? 0.26 : 0.15
  }

  if (motesMaterial) {
    motesMaterial.color.set(color)
    motesMaterial.opacity = dark ? 0.85 : 0.6
    motesMaterial.size = dark ? 6.5 : 5.2
  }
}

function handlePointerMove(e) {
  if (!containerRef.value) return
  const rect = containerRef.value.getBoundingClientRect()
  if (rect.width <= 0 || rect.height <= 0) return
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
  fov: 46,
  cameraPos: [0, 56, 92],
  lookAt: [0, 2, 0],
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

    texture = createBloomTexture()
    material = new THREE.PointsMaterial({
      size: isDarkMode.value ? 5.8 : 4.8,
      map: texture,
      transparent: true,
      opacity: isDarkMode.value ? 0.88 : 0.65,
      depthWrite: false,
      blending: THREE.NormalBlending
    })

    pointsMesh = new THREE.Points(geometry, material)
    scene.add(pointsMesh)

    // Wireframe cyber grid lines linking adjacent nodes
    const lineIndices = []
    for (let ix = 0; ix < countX; ix++) {
      for (let iy = 0; iy < countY; iy++) {
        const i = ix * countY + iy
        if (ix < countX - 1) {
          lineIndices.push(i, (ix + 1) * countY + iy)
        }
        if (iy < countY - 1 && iy % 2 === 0) {
          lineIndices.push(i, ix * countY + (iy + 1))
        }
      }
    }

    lineGeometry = new THREE.BufferGeometry()
    lineGeometry.setAttribute('position', geometry.getAttribute('position'))
    lineGeometry.setIndex(lineIndices)

    lineMaterial = new THREE.LineBasicMaterial({
      color: new THREE.Color(getThemeColor()),
      transparent: true,
      opacity: isDarkMode.value ? 0.26 : 0.15,
      depthWrite: false
    })

    lineMesh = new THREE.LineSegments(lineGeometry, lineMaterial)
    scene.add(lineMesh)

    // Floating ambient sparkles
    const moteCount = 28
    const motePositions = new Float32Array(moteCount * 3)
    for (let m = 0; m < moteCount * 3; m += 3) {
      motePositions[m] = (Math.random() - 0.5) * offsetX * 1.8
      motePositions[m + 1] = 4 + Math.random() * 22
      motePositions[m + 2] = (Math.random() - 0.5) * offsetZ * 1.6
    }
    motesGeometry = new THREE.BufferGeometry()
    motesGeometry.setAttribute('position', new THREE.BufferAttribute(motePositions, 3))

    motesMaterial = new THREE.PointsMaterial({
      size: isDarkMode.value ? 6.5 : 5.2,
      map: texture,
      transparent: true,
      opacity: isDarkMode.value ? 0.85 : 0.6,
      depthWrite: false,
      blending: THREE.NormalBlending
    })
    motesMesh = new THREE.Points(motesGeometry, motesMaterial)
    scene.add(motesMesh)

    updateMaterialStyle()
  },
  onRender({ elapsed }) {
    if (!geometry || !material) return

    pointer.x += (pointer.targetX - pointer.x) * 0.06
    pointer.y += (pointer.targetY - pointer.y) * 0.06

    const positions = geometry.attributes.position.array
    const countX = props.particleCountX
    const countY = props.particleCountY
    const time = elapsed * 1.1

    let idx = 0
    for (let ix = 0; ix < countX; ix++) {
      for (let iy = 0; iy < countY; iy++) {
        const u = ix / countX
        const v = iy / countY

        // Harmonious cyber wave undulating
        const waveA = Math.sin(u * 7.5 + time * 1.4) * 5.4
        const waveB = Math.cos(v * 5.2 + time * 1.2) * 4.2
        const waveC = Math.sin((u + v) * 4.0 + time * 0.9) * 3.2

        let pointerShift = 0
        if (pointer.active || Math.abs(pointer.x) > 0.01 || Math.abs(pointer.y) > 0.01) {
          const px = (u - 0.5) * 2.2
          const py = (v - 0.5) * 2.2
          const distSq = (px - pointer.x) * (px - pointer.x) + (py - pointer.y) * (py - pointer.y)
          if (distSq < 1.2) {
            pointerShift = Math.cos(Math.sqrt(distSq) * Math.PI * 0.5) * 8.0
          }
        }

        positions[idx + 1] = waveA + waveB + waveC + pointerShift
        idx += 3
      }
    }
    geometry.attributes.position.needsUpdate = true

    // Animate floating sparkles
    if (motesGeometry) {
      const motesPos = motesGeometry.attributes.position.array
      for (let m = 0; m < motesPos.length; m += 3) {
        motesPos[m + 1] += Math.sin(time * 1.5 + m) * 0.04
      }
      motesGeometry.attributes.position.needsUpdate = true
    }
  }
})

onMounted(() => {
  // Attach pointer listener to parent card so moving cursor anywhere on the card ripples the wave
  hostElement = containerRef.value?.parentElement
  if (hostElement) {
    hostElement.addEventListener('pointermove', handlePointerMove, { passive: true })
    hostElement.addEventListener('pointerleave', handlePointerLeave, { passive: true })
  }
})

onBeforeUnmount(() => {
  if (hostElement) {
    hostElement.removeEventListener('pointermove', handlePointerMove)
    hostElement.removeEventListener('pointerleave', handlePointerLeave)
    hostElement = null
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
  pointer-events: none;
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
    opacity: 0.22;
  }
}
</style>
