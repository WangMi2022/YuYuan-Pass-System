/*
 * useThreeScene — shared lifecycle harness for decorative three.js scenes.
 *
 * Loads three via dynamic import (kept out of the main bundle), creates a
 * transparent low-power renderer inside `hostRef`, then hands scene/camera
 * to the caller's `setup`. Handles: ResizeObserver sizing, DPR cap, RAF loop
 * with visibilitychange pause, prefers-reduced-motion (renders a single
 * static frame), WebGL-missing fallback (`failed` -> caller shows CSS
 * gradient), and full disposal on unmount.
 */
import { onMounted, onBeforeUnmount, ref, shallowRef } from 'vue'

export function useThreeScene(hostRef, { setup, onFrame } = {}) {
  const ready = ref(false)
  const failed = ref(false)
  const api = shallowRef(null)

  let disposed = false
  let rafId = 0
  let renderer = null
  let scene = null
  let camera = null
  let timer = null
  let observer = null
  let reduceMotion = false
  let running = false

  const renderFrame = () => {
    if (!renderer) return
    timer.update()
    const delta = timer.getDelta()
    const elapsed = timer.getElapsed()
    onFrame?.({ elapsed, delta, scene, camera })
    renderer.render(scene, camera)
  }

  const loop = () => {
    renderFrame()
    if (running) rafId = requestAnimationFrame(loop)
  }

  const startLoop = () => {
    if (running || disposed || !renderer) return
    if (reduceMotion) {
      renderFrame()
      return
    }
    running = true
    rafId = requestAnimationFrame(loop)
  }

  const stopLoop = () => {
    running = false
    cancelAnimationFrame(rafId)
  }

  const onVisibility = () => {
    if (document.hidden) stopLoop()
    else startLoop()
  }

  // reduced-motion 下渲染单帧；主题变色后由组件调用以刷新静态画面
  const renderOnce = () => {
    if (!running) renderFrame()
  }

  const start = async () => {
    const host = hostRef.value
    if (!host || disposed) return

    let THREE
    try {
      THREE = await import('three')
    } catch {
      failed.value = true
      return
    }
    if (disposed || !hostRef.value) return

    try {
      renderer = new THREE.WebGLRenderer({
        antialias: true,
        alpha: true,
        powerPreference: 'low-power'
      })
    } catch {
      failed.value = true
      return
    }

    reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

    renderer.setClearColor(0x000000, 0)
    renderer.domElement.style.cssText =
      'position:absolute;inset:0;width:100%;height:100%;display:block;'
    host.appendChild(renderer.domElement)

    scene = new THREE.Scene()
    camera = new THREE.PerspectiveCamera(50, 1, 0.1, 200)
    timer = new THREE.Timer()
    timer.connect(document)

    const resize = () => {
      const w = host.clientWidth
      const h = host.clientHeight
      if (!w || !h) return
      renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
      renderer.setSize(w, h, false)
      camera.aspect = w / h
      camera.updateProjectionMatrix()
      if (reduceMotion) renderOnce()
    }

    api.value = setup?.({ THREE, scene, camera, renderer, host }) || null

    resize()
    observer = new ResizeObserver(resize)
    observer.observe(host)
    document.addEventListener('visibilitychange', onVisibility)

    ready.value = true
    startLoop()
  }

  const dispose = () => {
    disposed = true
    stopLoop()
    observer?.disconnect()
    observer = null
    document.removeEventListener('visibilitychange', onVisibility)
    if (scene) {
      scene.traverse((obj) => {
        obj.geometry?.dispose?.()
        const mats = Array.isArray(obj.material) ? obj.material : [obj.material]
        mats.forEach((m) => {
          if (!m) return
          m.map?.dispose?.()
          m.dispose?.()
        })
      })
      scene.clear()
    }
    if (renderer) {
      renderer.dispose()
      renderer.forceContextLoss?.()
      renderer.domElement?.remove()
    }
    renderer = null
    scene = null
    camera = null
    timer?.dispose()
    timer = null
    api.value = null
  }

  onMounted(start)
  onBeforeUnmount(dispose)

  return { ready, failed, api, renderOnce }
}
