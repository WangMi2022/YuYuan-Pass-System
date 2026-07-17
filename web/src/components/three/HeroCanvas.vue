<template>
  <div ref="host" class="hero-canvas" :class="{ 'is-failed': failed }" aria-hidden="true" />
</template>

<script setup>
  import { ref, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useAppStore } from '@/pinia'
  import { useThreeScene } from './useThreeScene'

  defineOptions({ name: 'HeroCanvas' })

  const host = ref(null)
  const appStore = useAppStore()
  const { config, isDark } = storeToRefs(appStore)

  const COLS = 110
  const ROWS = 30
  const GAP = 0.42

  const { failed, api, renderOnce } = useThreeScene(host, {
    setup({ THREE, scene, camera }) {
      camera.fov = 42
      camera.position.set(0, 5.2, 10.5)
      camera.lookAt(0, -0.6, 0)

      const count = COLS * ROWS
      const positions = new Float32Array(count * 3)
      let i = 0
      for (let r = 0; r < ROWS; r++) {
        for (let c = 0; c < COLS; c++) {
          positions[i++] = (c - (COLS - 1) / 2) * GAP
          positions[i++] = 0
          positions[i++] = (r - (ROWS - 1) / 2) * GAP
        }
      }
      const geometry = new THREE.BufferGeometry()
      geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3))

      const material = new THREE.PointsMaterial({
        color: new THREE.Color(config.value.primaryColor),
        size: 0.045,
        sizeAttenuation: true,
        transparent: true,
        opacity: isDark.value ? 0.6 : 0.42,
        depthWrite: false
      })
      scene.add(new THREE.Points(geometry, material))

      return {
        setTheme(color, dark) {
          material.color.set(color)
          material.opacity = dark ? 0.6 : 0.42
        },
        wave(elapsed) {
          const pos = geometry.attributes.position
          const arr = pos.array
          for (let p = 0; p < count; p++) {
            const x = arr[p * 3]
            const z = arr[p * 3 + 2]
            arr[p * 3 + 1] =
              Math.sin(x * 0.55 + elapsed * 0.7) * 0.32 +
              Math.cos((z + x * 0.4) * 0.62 + elapsed * 0.45) * 0.28
          }
          pos.needsUpdate = true
        }
      }
    },
    onFrame({ elapsed }) {
      api.value?.wave(elapsed)
    }
  })

  watch([() => config.value.primaryColor, isDark], ([color, dark]) => {
    api.value?.setTheme(color, dark)
    renderOnce()
  })
</script>

<style scoped lang="scss">
  .hero-canvas {
    position: absolute;
    inset: 0;
    overflow: hidden;
    pointer-events: none;
    mask-image: linear-gradient(to bottom, rgb(0 0 0 / 85%) 30%, transparent 96%);

    &.is-failed {
      background:
        radial-gradient(120% 130% at 88% -8%, var(--na-primary-soft), transparent 55%),
        radial-gradient(90% 120% at 12% 110%, var(--na-accent), transparent 60%);
      mask-image: none;
    }
  }
</style>
