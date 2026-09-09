<template>
  <div ref="host" class="asset-orbit" :data-renderer="ready ? 'webgl' : 'fallback'" aria-hidden="true">
    <svg v-if="!ready" class="orbit-fallback" viewBox="0 0 400 300">
      <ellipse cx="200" cy="170" rx="164" ry="88" fill="none" :stroke="palette.primary" opacity=".15" />
      <g transform="translate(200 148) rotate(-90)">
        <circle r="100" fill="none" :stroke="palette.grid" stroke-width="14" />
        <circle v-for="arc in fallbackArcs" :key="arc.key" r="100" fill="none" :stroke="arc.color" stroke-width="14" :stroke-dasharray="arc.dash" :stroke-dashoffset="arc.offset" />
      </g>
    </svg>
    <canvas ref="canvas" :class="{ 'is-ready': ready }" />
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps({
  segments: { type: Array, default: () => [] },
  palette: { type: Object, required: true },
  animated: { type: Boolean, default: true }
})
const host = ref(null)
const canvas = ref(null)
const ready = ref(false)
const fallbackArcs = computed(() => {
  let offset = 0
  return props.segments.filter((item) => item.ratio > 0).map((item) => {
    const length = item.ratio / 100 * 628.318
    const arc = { key: item.key, color: props.palette[item.tone] || props.palette.primary, dash: `${Math.max(0, length - Math.min(3, length * .1))} 628.318`, offset: -offset }
    offset += length
    return arc
  })
})
let three, renderer, scene, camera, orbit, satellites, resizeObserver, intersectionObserver
let frame = 0
let disposed = false
let visible = true
let lost = false
let lastFrame = 0
let phase = 0

function disposeGroup(group) {
  group?.traverse((object) => {
    object.geometry?.dispose()
    if (Array.isArray(object.material)) object.material.forEach((material) => material.dispose())
    else object.material?.dispose()
  })
}

function rebuild() {
  if (!renderer || lost || disposed) return
  if (orbit) { scene.remove(orbit); disposeGroup(orbit) }
  orbit = new three.Group()
  orbit.rotation.x = -.38
  scene.add(orbit)
  let angle = 0
  const segments = props.segments.filter((item) => item.ratio > 0)
  if (!segments.length) segments.push({ ratio: 100, tone: 'grid' })
  for (const item of segments) {
    const arc = item.ratio / 100 * Math.PI * 2
    const material = new three.MeshStandardMaterial({ color: props.palette[item.tone] || props.palette.primary, metalness: .28, roughness: .32 })
    const mesh = new three.Mesh(new three.TorusGeometry(2, .115, 12, Math.max(8, Math.ceil(arc * 22)), Math.max(.001, arc - Math.min(.025, arc * .1))), material)
    mesh.rotation.z = angle
    orbit.add(mesh)
    angle += arc
  }
  for (const [radius, opacity] of [[2.3, .24], [2.48, .1], [1.68, .12]]) {
    const ring = new three.Mesh(new three.TorusGeometry(radius, .008, 4, 128), new three.MeshBasicMaterial({ color: props.palette.primary, transparent: true, opacity }))
    orbit.add(ring)
  }
  const ticks = []
  for (let i = 0; i < 80; i++) {
    const theta = i / 80 * Math.PI * 2
    const radius = i % 5 === 0 ? 2.63 : 2.57
    ticks.push(Math.cos(theta) * 2.52, Math.sin(theta) * 2.52, 0, Math.cos(theta) * radius, Math.sin(theta) * radius, 0)
  }
  const geometry = new three.BufferGeometry()
  geometry.setAttribute('position', new three.Float32BufferAttribute(ticks, 3))
  orbit.add(new three.LineSegments(geometry, new three.LineBasicMaterial({ color: props.palette.primary, transparent: true, opacity: .3 })))
  satellites = new three.Group()
  const points = []
  for (let i = 0; i < 54; i++) {
    const theta = i * 2.39996
    const radius = 2.78 + (i % 7) * .065
    points.push(Math.cos(theta) * radius, Math.sin(theta) * radius, Math.sin(i * .9) * .18)
  }
  const particles = new three.BufferGeometry()
  particles.setAttribute('position', new three.Float32BufferAttribute(points, 3))
  satellites.add(new three.Points(particles, new three.PointsMaterial({ color: props.palette.primary, size: .032, transparent: true, opacity: .6, sizeAttenuation: true })))
  orbit.add(satellites)
  draw()
}

function draw() {
  if (!renderer || lost || disposed) return
  renderer.render(scene, camera)
}
function tick(time) {
  frame = 0
  if (!props.animated || !visible || document.hidden || lost || disposed) return
  if (time - lastFrame >= 1000 / 30) {
    phase += Math.min(time - lastFrame, 60) / 1000
    lastFrame = time
    orbit.rotation.z = Math.sin(phase * .12) * .08
    orbit.rotation.x = -.38 + Math.sin(phase * .2) * .035
    satellites.rotation.z = phase * .07
    draw()
  }
  frame = requestAnimationFrame(tick)
}
function syncAnimation() {
  cancelAnimationFrame(frame)
  frame = 0
  if (renderer && !lost && props.animated && visible && !document.hidden && !disposed) {
    lastFrame = performance.now()
    frame = requestAnimationFrame(tick)
  }
}
function resize() {
  if (!renderer || disposed || lost) return
  const { width, height } = host.value.getBoundingClientRect()
  if (!width || !height) return
  renderer.setSize(width, height, false)
  camera.aspect = width / height
  camera.position.z = camera.aspect < 1.35 ? 9.8 / camera.aspect : 7.4
  camera.updateProjectionMatrix()
  draw()
}
function contextLost(event) {
  event.preventDefault()
  lost = true
  ready.value = false
  syncAnimation()
}
function contextRestored() {
  lost = false
  rebuild()
  resize()
  ready.value = true
  syncAnimation()
}

watch(() => props.animated, syncAnimation)
watch([() => props.segments, () => props.palette], rebuild, { deep: true })
onMounted(async () => {
  try {
    three = await import('three')
    if (disposed) return
    renderer = new three.WebGLRenderer({ canvas: canvas.value, alpha: true, antialias: true, powerPreference: 'low-power' })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 1.5))
    scene = new three.Scene()
    camera = new three.PerspectiveCamera(48, 1, .1, 30)
    camera.position.z = 7.4
    scene.add(new three.AmbientLight(0xffffff, 2.1))
    const light = new three.DirectionalLight(0xffffff, 3)
    light.position.set(2, 4, 5)
    scene.add(light)
    rebuild()
    resize()
    ready.value = true
    resizeObserver = new ResizeObserver(resize)
    resizeObserver.observe(host.value)
    intersectionObserver = new IntersectionObserver(([entry]) => { visible = entry.isIntersecting; syncAnimation() })
    intersectionObserver.observe(host.value)
    canvas.value.addEventListener('webglcontextlost', contextLost)
    canvas.value.addEventListener('webglcontextrestored', contextRestored)
    document.addEventListener('visibilitychange', syncAnimation)
    syncAnimation()
  } catch {
    ready.value = false
    renderer?.dispose()
    renderer = null
  }
})
onBeforeUnmount(() => {
  disposed = true
  cancelAnimationFrame(frame)
  resizeObserver?.disconnect()
  intersectionObserver?.disconnect()
  document.removeEventListener('visibilitychange', syncAnimation)
  canvas.value?.removeEventListener('webglcontextlost', contextLost)
  canvas.value?.removeEventListener('webglcontextrestored', contextRestored)
  disposeGroup(scene)
  renderer?.dispose()
  renderer?.forceContextLoss()
  renderer = null
})
</script>

<style scoped>
.asset-orbit { position: absolute; inset: 0; overflow: hidden; pointer-events: none; }
.asset-orbit canvas, .orbit-fallback { display: block; width: 100%; height: 100%; }
.asset-orbit canvas { opacity: 0; }
.asset-orbit canvas.is-ready { opacity: 1; }
.orbit-fallback { position: absolute; inset: 0; }
</style>
