<template>
  <div ref="host" class="ambient-canvas" :class="{ 'is-failed': failed }" aria-hidden="true" />
</template>

<script setup>
  import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useAppStore } from '@/pinia'
  import { useThreeScene } from './useThreeScene'

  defineOptions({ name: 'AmbientCanvas' })

  const host = ref(null)
  const appStore = useAppStore()
  const { config, isDark } = storeToRefs(appStore)

  const PRIMARY_COUNT = 150
  const NEUTRAL_COUNT = 110
  const SPREAD = { x: 17, y: 9, z: 6 }

  const pointer = { x: 0, y: 0 }
  const onPointerMove = (e) => {
    pointer.x = (e.clientX / window.innerWidth) * 2 - 1
    pointer.y = (e.clientY / window.innerHeight) * 2 - 1
  }
  onMounted(() => window.addEventListener('pointermove', onPointerMove, { passive: true }))
  onBeforeUnmount(() => window.removeEventListener('pointermove', onPointerMove))

  const buildCloud = (THREE, count, size) => {
    const positions = new Float32Array(count * 3)
    const speeds = new Float32Array(count)
    for (let i = 0; i < count; i++) {
      positions[i * 3] = (Math.random() * 2 - 1) * SPREAD.x
      positions[i * 3 + 1] = (Math.random() * 2 - 1) * SPREAD.y
      positions[i * 3 + 2] = (Math.random() * 2 - 1) * SPREAD.z
      speeds[i] = 0.12 + Math.random() * 0.3
    }
    const geometry = new THREE.BufferGeometry()
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3))
    const material = new THREE.PointsMaterial({
      size,
      sizeAttenuation: true,
      transparent: true,
      depthWrite: false
    })
    return { points: new THREE.Points(geometry, material), geometry, material, speeds, count }
  }

  const { failed, api, renderOnce } = useThreeScene(host, {
    setup({ THREE, scene, camera }) {
      camera.fov = 55
      camera.position.set(0, 0, 12)

      const primary = buildCloud(THREE, PRIMARY_COUNT, 0.075)
      const neutral = buildCloud(THREE, NEUTRAL_COUNT, 0.05)
      scene.add(primary.points, neutral.points)

      const applyTheme = (color, dark) => {
        primary.material.color.set(color)
        primary.material.opacity = dark ? 0.5 : 0.34
        neutral.material.color.set(dark ? '#52525b' : '#a1a1aa')
        neutral.material.opacity = dark ? 0.4 : 0.3
      }
      applyTheme(config.value.primaryColor, isDark.value)

      const drift = (cloud, delta) => {
        const arr = cloud.geometry.attributes.position.array
        for (let i = 0; i < cloud.count; i++) {
          arr[i * 3 + 1] += cloud.speeds[i] * delta
          if (arr[i * 3 + 1] > SPREAD.y) arr[i * 3 + 1] = -SPREAD.y
        }
        cloud.geometry.attributes.position.needsUpdate = true
      }

      return {
        setTheme: applyTheme,
        tick(delta) {
          drift(primary, delta)
          drift(neutral, delta * 0.75)
          camera.position.x += (pointer.x * 0.9 - camera.position.x) * 0.03
          camera.position.y += (-pointer.y * 0.55 - camera.position.y) * 0.03
          camera.lookAt(0, 0, 0)
        }
      }
    },
    onFrame({ delta }) {
      api.value?.tick(delta)
    }
  })

  watch([() => config.value.primaryColor, isDark], ([color, dark]) => {
    api.value?.setTheme(color, dark)
    renderOnce()
  })
</script>

<style scoped lang="scss">
  .ambient-canvas {
    position: absolute;
    inset: 0;
    overflow: hidden;
    pointer-events: none;

    &.is-failed {
      background:
        radial-gradient(90% 90% at 80% 0%, var(--na-primary-soft), transparent 60%),
        radial-gradient(70% 80% at 8% 100%, var(--na-accent), transparent 55%);
    }
  }
</style>
