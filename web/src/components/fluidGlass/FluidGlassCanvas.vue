<template>
  <canvas ref="canvas" class="fluid-glass-canvas" aria-hidden="true" />
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import vertexShaderSource from './shaders/canonical-vertex.glsl?raw'
import fragmentShaderSource from './shaders/canonical-fragment.glsl?raw'

const props = defineProps({
  speed: { type: Number, default: 0.52 },
  intensity: { type: Number, default: 0.86 },
  pointer: { type: Number, default: 0.62 }
})

const canvas = ref(null)
let animationFrame = 0
let gl = null
let program = null
let resizeObserver = null
let host = null
let onPointerMove = null
let onPointerLeave = null
let onVisibilityChange = null
let startedAt = 0
let lastFrame = 0
let reducedMotion = false
const uniforms = {}
const pointerState = {
  current: { x: 0.78, y: 0.5 },
  target: { x: 0.78, y: 0.5 },
  velocity: { x: 0, y: 0 }
}

function compileShader(type, source) {
  const shader = gl.createShader(type)
  gl.shaderSource(shader, source)
  gl.compileShader(shader)
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    const error = gl.getShaderInfoLog(shader)
    gl.deleteShader(shader)
    throw new Error(error || 'Fluid Glass shader compilation failed')
  }
  return shader
}

function createProgram() {
  const vertex = compileShader(gl.VERTEX_SHADER, vertexShaderSource)
  const fragment = compileShader(gl.FRAGMENT_SHADER, fragmentShaderSource)
  const nextProgram = gl.createProgram()
  gl.attachShader(nextProgram, vertex)
  gl.attachShader(nextProgram, fragment)
  gl.linkProgram(nextProgram)
  gl.deleteShader(vertex)
  gl.deleteShader(fragment)
  if (!gl.getProgramParameter(nextProgram, gl.LINK_STATUS)) {
    const error = gl.getProgramInfoLog(nextProgram)
    gl.deleteProgram(nextProgram)
    throw new Error(error || 'Fluid Glass shader linking failed')
  }
  return nextProgram
}

function hexToRgb(hex) {
  const value = Number.parseInt(hex.slice(1), 16)
  return [((value >> 16) & 255) / 255, ((value >> 8) & 255) / 255, (value & 255) / 255]
}

function resize() {
  if (!canvas.value || !gl) return
  const rect = canvas.value.getBoundingClientRect()
  const dpr = Math.min(window.devicePixelRatio || 1, 1.35)
  const width = Math.max(1, Math.round(rect.width * dpr))
  const height = Math.max(1, Math.round(rect.height * dpr))
  if (canvas.value.width === width && canvas.value.height === height) return
  canvas.value.width = width
  canvas.value.height = height
  gl.viewport(0, 0, width, height)
}

function setPointer(event) {
  if (!host) return
  const rect = host.getBoundingClientRect()
  if (!rect.width || !rect.height) return
  const nextX = Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width))
  const nextY = Math.min(1, Math.max(0, 1 - (event.clientY - rect.top) / rect.height))
  pointerState.velocity.x = Math.max(-1, Math.min(1, (nextX - pointerState.target.x) * 5))
  pointerState.velocity.y = Math.max(-1, Math.min(1, (nextY - pointerState.target.y) * 5))
  pointerState.target.x = nextX
  pointerState.target.y = nextY
}

function resetPointer() {
  pointerState.target.x = 0.78
  pointerState.target.y = 0.5
  pointerState.velocity.x = 0
  pointerState.velocity.y = 0
}

function draw(now) {
  if (!gl || !program || !canvas.value) return
  const elapsed = (now - startedAt) / 1000
  const dt = Math.min(0.08, Math.max(0.001, (now - lastFrame) / 1000))
  lastFrame = now
  pointerState.current.x += (pointerState.target.x - pointerState.current.x) * Math.min(1, dt * 5)
  pointerState.current.y += (pointerState.target.y - pointerState.current.y) * Math.min(1, dt * 5)
  pointerState.velocity.x *= Math.max(0, 1 - dt * 7)
  pointerState.velocity.y *= Math.max(0, 1 - dt * 7)

  gl.useProgram(program)
  gl.uniform2f(uniforms.resolution, canvas.value.width, canvas.value.height)
  gl.uniform2f(uniforms.mouse, pointerState.current.x, pointerState.current.y)
  gl.uniform2f(uniforms.mouseVelocity, pointerState.velocity.x, pointerState.velocity.y)
  gl.uniform1f(uniforms.mouseMix, reducedMotion ? 0.12 : 0.38)
  gl.uniform1f(uniforms.time, reducedMotion ? elapsed * 0.15 : elapsed)
  gl.uniform1f(uniforms.speed, props.speed)
  gl.uniform1f(uniforms.intensity, props.intensity)
  gl.uniform1f(uniforms.pointer, props.pointer)
  gl.uniform1f(uniforms.seed, 5.4)
  gl.uniform1f(uniforms.surfaceOpacity, 0.08)
  gl.uniform3fv(uniforms.colorA, hexToRgb('#00e7d2'))
  gl.uniform3fv(uniforms.colorB, hexToRgb('#3cc8ff'))
  gl.uniform3fv(uniforms.colorC, hexToRgb('#075f68'))
  gl.drawArrays(gl.TRIANGLES, 0, 6)
}

function animate(now) {
  if (document.hidden) {
    animationFrame = requestAnimationFrame(animate)
    return
  }
  draw(now)
  animationFrame = requestAnimationFrame(animate)
}

function activateFallback() {
  host?.classList.add('fluid-glass-fallback')
}

onMounted(() => {
  host = canvas.value?.parentElement
  reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  try {
    gl = canvas.value?.getContext('webgl', { alpha: true, premultipliedAlpha: false, antialias: true })
    if (!gl) throw new Error('WebGL unavailable')
    program = createProgram()
    for (const name of ['resolution', 'mouse', 'mouseVelocity', 'mouseMix', 'time', 'speed', 'intensity', 'pointer', 'seed', 'surfaceOpacity', 'colorA', 'colorB', 'colorC']) {
      uniforms[name] = gl.getUniformLocation(program, `u_${name}`)
    }
    const buffer = gl.createBuffer()
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer)
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 1, -1, -1, 1, -1, 1, 1, -1, 1, 1]), gl.STATIC_DRAW)
    const position = gl.getAttribLocation(program, 'a_position')
    gl.enableVertexAttribArray(position)
    gl.vertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0)
    gl.clearColor(0, 0, 0, 0)
    resize()
    resizeObserver = new ResizeObserver(resize)
    resizeObserver.observe(canvas.value)
    onPointerMove = (event) => setPointer(event)
    onPointerLeave = resetPointer
    onVisibilityChange = () => {
      if (!document.hidden) lastFrame = performance.now()
    }
    host?.addEventListener('pointermove', onPointerMove, { passive: true })
    host?.addEventListener('pointerleave', onPointerLeave, { passive: true })
    document.addEventListener('visibilitychange', onVisibilityChange)
    startedAt = performance.now()
    lastFrame = startedAt
    animationFrame = requestAnimationFrame(animate)
  } catch (error) {
    activateFallback()
    console.warn('[FluidGlassCanvas] falling back to CSS surface', error)
  }
})

onBeforeUnmount(() => {
  cancelAnimationFrame(animationFrame)
  resizeObserver?.disconnect()
  if (host && onPointerMove) host.removeEventListener('pointermove', onPointerMove)
  if (host && onPointerLeave) host.removeEventListener('pointerleave', onPointerLeave)
  if (onVisibilityChange) document.removeEventListener('visibilitychange', onVisibilityChange)
  if (gl && program) gl.deleteProgram(program)
  gl = null
  program = null
})
</script>

<style scoped>
.fluid-glass-canvas {
  display: block;
  width: 100%;
  height: 100%;
  pointer-events: none;
  opacity: .9;
  transform: scale(1.035);
  filter: saturate(1.1) contrast(1.04) blur(1.8px);
}

.fluid-glass-fallback .fluid-glass-canvas {
  display: none;
}
</style>
