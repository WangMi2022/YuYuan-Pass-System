import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as THREE from 'three'
import { useDocumentVisibility, usePreferredReducedMotion } from '@vueuse/core'

/**
 * useThreeScene manages Three.js scene lifecycle, resize events, WebGL fallback,
 * and visibility/motion state.
 */
export function useThreeScene(containerRef, options = {}) {
  const isAvailable = ref(false)
  const isRunning = ref(false)
  const visibility = useDocumentVisibility()
  const motionPreference = usePreferredReducedMotion()

  let scene = null
  let camera = null
  let renderer = null
  let animationFrameId = null
  let resizeObserver = null
  let clock = new THREE.Clock()

  const {
    fov = 55,
    near = 1,
    far = 1000,
    cameraPos = [0, 45, 110],
    lookAt = [0, 0, 0],
    alpha = true,
    antialias = true,
    onInit = null,
    onRender = null,
    onResize = null
  } = options

  function checkWebGLSupport() {
    try {
      const canvas = document.createElement('canvas')
      return Boolean(window.WebGLRenderingContext && (canvas.getContext('webgl2') || canvas.getContext('webgl')))
    } catch {
      return false
    }
  }

  function init() {
    if (!containerRef.value || !checkWebGLSupport()) return false

    const container = containerRef.value
    const width = container.clientWidth || 300
    const height = container.clientHeight || 150

    scene = new THREE.Scene()
    camera = new THREE.PerspectiveCamera(fov, width / Math.max(height, 1), near, far)
    camera.position.set(...cameraPos)
    camera.lookAt(...lookAt)

    try {
      renderer = new THREE.WebGLRenderer({
        alpha,
        antialias,
        powerPreference: 'high-performance'
      })
      renderer.setSize(width, height)
      renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
      container.appendChild(renderer.domElement)
    } catch (e) {
      console.warn('Three.js WebGL initialization skipped:', e)
      return false
    }

    if (typeof onInit === 'function') {
      onInit({ scene, camera, renderer, THREE, width, height })
    }

    resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width: newW, height: newH } = entry.contentRect
        if (newW > 0 && newH > 0 && camera && renderer) {
          camera.aspect = newW / newH
          camera.updateProjectionMatrix()
          renderer.setSize(newW, newH)
          if (typeof onResize === 'function') {
            onResize({ width: newW, height: newH, camera, renderer })
          }
        }
      }
    })
    resizeObserver.observe(container)

    isAvailable.value = true
    return true
  }

  function tick() {
    if (!isRunning.value) return
    const delta = clock.getDelta()
    const elapsed = clock.getElapsedTime()

    if (typeof onRender === 'function') {
      onRender({ scene, camera, renderer, delta, elapsed, THREE, clock })
    }

    if (renderer && scene && camera) {
      renderer.render(scene, camera)
    }

    animationFrameId = requestAnimationFrame(tick)
  }

  function start() {
    if (!isAvailable.value || isRunning.value) return
    if (motionPreference.value === 'reduce') {
      // If motion is reduced, render a single frame and do not loop
      if (renderer && scene && camera) renderer.render(scene, camera)
      return
    }
    isRunning.value = true
    clock.start()
    animationFrameId = requestAnimationFrame(tick)
  }

  function stop() {
    isRunning.value = false
    if (animationFrameId) {
      cancelAnimationFrame(animationFrameId)
      animationFrameId = null
    }
  }

  function renderSingleFrame() {
    if (renderer && scene && camera) {
      if (typeof onRender === 'function') {
        onRender({ scene, camera, renderer, delta: 0, elapsed: 0, THREE, clock })
      }
      renderer.render(scene, camera)
    }
  }

  function cleanup() {
    stop()
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
    if (renderer) {
      if (renderer.domElement && renderer.domElement.parentNode) {
        renderer.domElement.parentNode.removeChild(renderer.domElement)
      }
      renderer.dispose()
      renderer = null
    }
    if (scene) {
      scene.clear()
      scene = null
    }
    camera = null
    isAvailable.value = false
  }

  onMounted(() => {
    if (init()) {
      if (visibility.value === 'visible') {
        start()
      }
    }
  })

  onBeforeUnmount(() => {
    cleanup()
  })

  watch(visibility, (state) => {
    if (state === 'visible') {
      start()
    } else {
      stop()
    }
  })

  watch(motionPreference, (pref) => {
    if (pref === 'reduce') {
      stop()
      renderSingleFrame()
    } else if (visibility.value === 'visible') {
      start()
    }
  })

  return {
    isAvailable,
    isRunning,
    start,
    stop,
    renderSingleFrame,
    getScene: () => scene,
    getCamera: () => camera,
    getRenderer: () => renderer
  }
}
