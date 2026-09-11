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
  separation: { type: Number, default: 7.8 },
  paused: { type: Boolean, default: false }
})

const appStore = useAppStore()
const containerRef = ref(null)

let cameraRef = null
let hostElement = null

// Meshes & Geometries references for lifecycle disposal
let floorPointsMesh = null
let floorLinesMesh = null
let floorGeometry = null
let floorLineGeometry = null
let floorMaterial = null
let floorLineMaterial = null

let hubPointsMesh = null
let hubGeometry = null
let hubMaterial = null

let nodePointsMesh = null
let nodeGeometry = null
let nodeMaterial = null

let linkMesh = null
let linkGeometry = null
let linkMaterial = null

let pulseMesh = null
let pulseGeometry = null
let pulseMaterial = null

let motesMesh = null
let motesGeometry = null
let motesMaterial = null

const radarRings = []

let hubTexture = null
let nodeTexture = null
let pulseTexture = null

const pointer = {
  x: 0,
  y: 0,
  targetX: 0,
  targetY: 0,
  active: false
}

// Deterministic pseudo-random helper
function seededRandom(seed) {
  const x = Math.sin(seed) * 10000
  return x - Math.floor(x)
}

// Procedural textures for high-performance glowing bloom
function createBloomTexture({ innerColor = 'rgba(255, 255, 255, 1)', midColor = 'rgba(165, 180, 252, 0.75)', outerColor = 'rgba(99, 102, 241, 0.25)', size = 128 } = {}) {
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')
  if (!ctx) return null

  const center = size / 2
  const gradient = ctx.createRadialGradient(center, center, 0, center, center, center)
  gradient.addColorStop(0, innerColor)
  gradient.addColorStop(0.18, 'rgba(255, 255, 255, 0.95)')
  gradient.addColorStop(0.4, midColor)
  gradient.addColorStop(0.72, outerColor)
  gradient.addColorStop(1, 'rgba(0, 0, 0, 0)')

  ctx.fillStyle = gradient
  ctx.beginPath()
  ctx.arc(center, center, center, 0, Math.PI * 2)
  ctx.fill()

  const tex = new THREE.CanvasTexture(canvas)
  tex.needsUpdate = true
  return tex
}

function createRadarRingGeometry(segments = 72) {
  const positions = new Float32Array(segments * 3)
  for (let i = 0; i < segments; i++) {
    const theta = (i / segments) * Math.PI * 2
    positions[i * 3 + 0] = Math.cos(theta)
    positions[i * 3 + 1] = 0
    positions[i * 3 + 2] = Math.sin(theta)
  }
  const geo = new THREE.BufferGeometry()
  geo.setAttribute('position', new THREE.BufferAttribute(positions, 3))
  return geo
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
  const themeHex = getThemeColor()
  const primary = new THREE.Color(themeHex)
  const bright = primary.clone().lerp(new THREE.Color('#FFFFFF'), 0.65)
  const dark = isDarkMode.value

  if (hubMaterial) {
    hubMaterial.color.copy(bright)
    hubMaterial.opacity = dark ? 0.96 : 0.78
    hubMaterial.size = dark ? 15.0 : 12.0
  }

  if (nodeMaterial) {
    nodeMaterial.color.copy(primary)
    nodeMaterial.opacity = dark ? 0.88 : 0.68
    nodeMaterial.size = dark ? 7.8 : 6.2
  }

  if (pulseMaterial) {
    pulseMaterial.color.copy(bright)
    pulseMaterial.opacity = dark ? 0.98 : 0.82
    pulseMaterial.size = dark ? 9.5 : 7.6
  }

  if (linkMaterial) {
    linkMaterial.color.copy(primary)
    linkMaterial.opacity = dark ? 0.28 : 0.16
  }

  if (floorMaterial) {
    floorMaterial.color.copy(primary)
    floorMaterial.opacity = dark ? 0.62 : 0.38
    floorMaterial.size = dark ? 4.6 : 3.8
  }

  if (floorLineMaterial) {
    floorLineMaterial.color.copy(primary)
    floorLineMaterial.opacity = dark ? 0.15 : 0.08
  }

  radarRings.forEach((ring) => {
    if (ring.material) {
      ring.material.color.copy(bright)
    }
  })

  if (motesMaterial) {
    motesMaterial.color.copy(bright)
    motesMaterial.opacity = dark ? 0.85 : 0.55
    motesMaterial.size = dark ? 6.5 : 5.0
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

// 6 Core Asset Management Hubs coordinates in 3D Space
const HUB_SPECS = [
  { name: 'HQ', pos: [0, 8, -4], scale: 1.45, phase: 0 },
  { name: 'Equipment', pos: [-36, 12, -18], scale: 1.2, phase: 1.2 },
  { name: 'IT_Cloud', pos: [34, 15, -14], scale: 1.25, phase: 2.1 },
  { name: 'Logistics', pos: [-26, 3, 16], scale: 1.15, phase: 3.4 },
  { name: 'Finance', pos: [28, 5, 18], scale: 1.18, phase: 4.6 },
  { name: 'RiskRadar', pos: [-2, 22, -26], scale: 1.3, phase: 5.5 }
]

// Simulation states
const constellationNodes = []
const constellationLinks = []
const nodeToLinks = []
const telemetryPulses = []

function buildAssetConstellation() {
  const hubCount = HUB_SPECS.length
  const hubIdxList = []

  // 1. Add Primary Hubs
  HUB_SPECS.forEach((hub, idx) => {
    const node = {
      x: hub.pos[0],
      y: hub.pos[1],
      z: hub.pos[2],
      baseX: hub.pos[0],
      baseY: hub.pos[1],
      baseZ: hub.pos[2],
      isHub: true,
      hubIdx: idx,
      scale: hub.scale,
      phase: hub.phase,
      speed: 0.75 + idx * 0.12
    }
    hubIdxList.push(constellationNodes.length)
    constellationNodes.push(node)
  })

  // 2. Add Satellites around each Hub
  HUB_SPECS.forEach((hub, hIdx) => {
    const satCount = 5 + (hIdx % 2)
    for (let s = 0; s < satCount; s++) {
      const seed = hIdx * 10 + s + 1
      const angle = (s / satCount) * Math.PI * 2 + seededRandom(seed * 3) * 0.55
      const dist = 9.5 + seededRandom(seed * 7) * 8.5
      const yOff = (seededRandom(seed * 13) - 0.5) * 9.0

      const sx = hub.pos[0] + Math.cos(angle) * dist
      const sy = Math.max(1, hub.pos[1] + yOff)
      const sz = hub.pos[2] + Math.sin(angle) * (dist * 0.85)

      constellationNodes.push({
        x: sx,
        y: sy,
        z: sz,
        baseX: sx,
        baseY: sy,
        baseZ: sz,
        isHub: false,
        hubIdx: hIdx,
        scale: 0.85 + seededRandom(seed * 17) * 0.35,
        phase: seed * 1.5,
        speed: 0.85 + seededRandom(seed * 19) * 0.45
      })
    }
  })

  // Initialize adjacency table
  for (let i = 0; i < constellationNodes.length; i++) {
    nodeToLinks.push([])
  }

  function addLink(fromIdx, toIdx) {
    if (fromIdx >= 0 && fromIdx < constellationNodes.length && toIdx >= 0 && toIdx < constellationNodes.length && fromIdx !== toIdx) {
      const linkIdx = constellationLinks.length
      constellationLinks.push({ from: fromIdx, to: toIdx })
      nodeToLinks[fromIdx].push(linkIdx)
      nodeToLinks[toIdx].push(linkIdx)
    }
  }

  // Inter-Hub Primary Backbone Arteries
  const backbonePairs = [
    [0, 1], [0, 2], [0, 3], [0, 4], [0, 5],
    [1, 3], [2, 4], [1, 5], [2, 5], [3, 4]
  ]
  backbonePairs.forEach(([a, b]) => addLink(hubIdxList[a], hubIdxList[b]))

  // Hub-to-satellite & satellite ring connections
  let curClusterStart = hubCount
  HUB_SPECS.forEach((hub, hIdx) => {
    const satCount = 5 + (hIdx % 2)
    for (let s = 0; s < satCount; s++) {
      const sIdx = curClusterStart + s
      addLink(hubIdxList[hIdx], sIdx)
      if (s > 0) {
        addLink(sIdx - 1, sIdx)
      }
    }
    addLink(curClusterStart + satCount - 1, curClusterStart)
    curClusterStart += satCount
  })

  // Inter-cluster bridges (Equipment <-> Logistics, IT <-> Finance, HQ <-> Satellites)
  addLink(hubCount + 1, hubCount + 8)
  addLink(hubCount + 14, hubCount + 22)
  addLink(hubCount + 3, hubCount + 28)

  // Initialize telemetry flowing data pulses
  const pulseCount = 30
  for (let p = 0; p < pulseCount; p++) {
    const linkIdx = Math.floor(seededRandom(p * 17) * constellationLinks.length)
    telemetryPulses.push({
      linkIdx,
      progress: seededRandom(p * 23),
      speed: 0.28 + seededRandom(p * 31) * 0.42,
      forward: seededRandom(p * 47) > 0.5
    })
  }
}

const { isAvailable, start, stop } = useThreeScene(containerRef, {
  fov: 48,
  cameraPos: [0, 44, 108],
  lookAt: [0, 4, -4],
  onInit({ scene, camera }) {
    cameraRef = camera

    // Build Constellation graph
    constellationNodes.length = 0
    constellationLinks.length = 0
    nodeToLinks.length = 0
    telemetryPulses.length = 0
    buildAssetConstellation()

    // Create Textures
    hubTexture = createBloomTexture({
      innerColor: 'rgba(255, 255, 255, 1)',
      midColor: 'rgba(224, 231, 255, 0.95)',
      outerColor: 'rgba(99, 102, 241, 0.38)',
      size: 128
    })

    nodeTexture = createBloomTexture({
      innerColor: 'rgba(255, 255, 255, 1)',
      midColor: 'rgba(199, 210, 254, 0.85)',
      outerColor: 'rgba(99, 102, 241, 0.22)',
      size: 96
    })

    pulseTexture = createBloomTexture({
      innerColor: 'rgba(255, 255, 255, 1)',
      midColor: 'rgba(255, 255, 255, 0.98)',
      outerColor: 'rgba(129, 140, 248, 0.45)',
      size: 96
    })

    // 1. Asset Hub Points Mesh
    const hubPositions = new Float32Array(HUB_SPECS.length * 3)
    for (let h = 0; h < HUB_SPECS.length; h++) {
      hubPositions[h * 3 + 0] = constellationNodes[h].x
      hubPositions[h * 3 + 1] = constellationNodes[h].y
      hubPositions[h * 3 + 2] = constellationNodes[h].z
    }
    hubGeometry = new THREE.BufferGeometry()
    hubGeometry.setAttribute('position', new THREE.BufferAttribute(hubPositions, 3))
    hubMaterial = new THREE.PointsMaterial({
      size: isDarkMode.value ? 15.0 : 12.0,
      map: hubTexture,
      transparent: true,
      opacity: isDarkMode.value ? 0.96 : 0.78,
      depthWrite: false,
      blending: THREE.NormalBlending
    })
    hubPointsMesh = new THREE.Points(hubGeometry, hubMaterial)
    scene.add(hubPointsMesh)

    // 2. Satellite Nodes Mesh
    const satCount = constellationNodes.length - HUB_SPECS.length
    const satPositions = new Float32Array(satCount * 3)
    for (let s = 0; s < satCount; s++) {
      const node = constellationNodes[HUB_SPECS.length + s]
      satPositions[s * 3 + 0] = node.x
      satPositions[s * 3 + 1] = node.y
      satPositions[s * 3 + 2] = node.z
    }
    nodeGeometry = new THREE.BufferGeometry()
    nodeGeometry.setAttribute('position', new THREE.BufferAttribute(satPositions, 3))
    nodeMaterial = new THREE.PointsMaterial({
      size: isDarkMode.value ? 7.8 : 6.2,
      map: nodeTexture,
      transparent: true,
      opacity: isDarkMode.value ? 0.88 : 0.68,
      depthWrite: false,
      blending: THREE.NormalBlending
    })
    nodePointsMesh = new THREE.Points(nodeGeometry, nodeMaterial)
    scene.add(nodePointsMesh)

    // 3. Topology Conduit Lines Mesh
    const linkPositions = new Float32Array(constellationLinks.length * 6)
    for (let l = 0; l < constellationLinks.length; l++) {
      const link = constellationLinks[l]
      const fNode = constellationNodes[link.from]
      const tNode = constellationNodes[link.to]
      linkPositions[l * 6 + 0] = fNode.x
      linkPositions[l * 6 + 1] = fNode.y
      linkPositions[l * 6 + 2] = fNode.z
      linkPositions[l * 6 + 3] = tNode.x
      linkPositions[l * 6 + 4] = tNode.y
      linkPositions[l * 6 + 5] = tNode.z
    }
    linkGeometry = new THREE.BufferGeometry()
    linkGeometry.setAttribute('position', new THREE.BufferAttribute(linkPositions, 3))
    linkMaterial = new THREE.LineBasicMaterial({
      color: new THREE.Color(getThemeColor()),
      transparent: true,
      opacity: isDarkMode.value ? 0.28 : 0.16,
      depthWrite: false
    })
    linkMesh = new THREE.LineSegments(linkGeometry, linkMaterial)
    scene.add(linkMesh)

    // 4. Flowing Telemetry Pulses Mesh
    const pulsePositions = new Float32Array(telemetryPulses.length * 3)
    pulseGeometry = new THREE.BufferGeometry()
    pulseGeometry.setAttribute('position', new THREE.BufferAttribute(pulsePositions, 3))
    pulseMaterial = new THREE.PointsMaterial({
      size: isDarkMode.value ? 9.5 : 7.6,
      map: pulseTexture,
      transparent: true,
      opacity: isDarkMode.value ? 0.98 : 0.82,
      depthWrite: false,
      blending: THREE.AdditiveBlending
    })
    pulseMesh = new THREE.Points(pulseGeometry, pulseMaterial)
    scene.add(pulseMesh)

    // 5. Concentric Radar Waves (Cyber Posture Scan)
    const radarCenter = HUB_SPECS[0].pos
    const ringOffsets = [0, 24, 48]
    radarRings.length = 0
    const ringBaseGeo = createRadarRingGeometry(72)
    ringOffsets.forEach((offset) => {
      const ringMat = new THREE.LineBasicMaterial({
        color: new THREE.Color(getThemeColor()),
        transparent: true,
        opacity: 0.35,
        depthWrite: false
      })
      const ringMesh = new THREE.LineLoop(ringBaseGeo, ringMat)
      ringMesh.position.set(radarCenter[0], radarCenter[1] - 0.5, radarCenter[2])
      scene.add(ringMesh)
      radarRings.push({ mesh: ringMesh, material: ringMat, offset })
    })

    // 6. Base Spatial Holographic Coordinate Floor Matrix
    const countX = props.particleCountX
    const countY = props.particleCountY
    const sep = props.separation
    const totalFloor = countX * countY
    const floorPos = new Float32Array(totalFloor * 3)
    const offsetX = ((countX - 1) * sep) / 2
    const offsetZ = ((countY - 1) * sep) / 2
    const floorBaseY = -18

    let fIdx = 0
    for (let ix = 0; ix < countX; ix++) {
      for (let iy = 0; iy < countY; iy++) {
        floorPos[fIdx] = ix * sep - offsetX
        floorPos[fIdx + 1] = floorBaseY
        floorPos[fIdx + 2] = iy * sep - offsetZ
        fIdx += 3
      }
    }
    floorGeometry = new THREE.BufferGeometry()
    floorGeometry.setAttribute('position', new THREE.BufferAttribute(floorPos, 3))
    floorMaterial = new THREE.PointsMaterial({
      size: isDarkMode.value ? 4.6 : 3.8,
      map: nodeTexture,
      transparent: true,
      opacity: isDarkMode.value ? 0.62 : 0.38,
      depthWrite: false,
      blending: THREE.NormalBlending
    })
    floorPointsMesh = new THREE.Points(floorGeometry, floorMaterial)
    scene.add(floorPointsMesh)

    // Floor lines connecting adjacent coordinate cells
    const floorLineIndices = []
    for (let ix = 0; ix < countX; ix++) {
      for (let iy = 0; iy < countY; iy++) {
        const i = ix * countY + iy
        if (ix < countX - 1) {
          floorLineIndices.push(i, (ix + 1) * countY + iy)
        }
        if (iy < countY - 1 && iy % 2 === 0) {
          floorLineIndices.push(i, ix * countY + (iy + 1))
        }
      }
    }
    floorLineGeometry = new THREE.BufferGeometry()
    floorLineGeometry.setAttribute('position', floorGeometry.getAttribute('position'))
    floorLineGeometry.setIndex(floorLineIndices)
    floorLineMaterial = new THREE.LineBasicMaterial({
      color: new THREE.Color(getThemeColor()),
      transparent: true,
      opacity: isDarkMode.value ? 0.15 : 0.08,
      depthWrite: false
    })
    floorLinesMesh = new THREE.LineSegments(floorLineGeometry, floorLineMaterial)
    scene.add(floorLinesMesh)

    // 7. Ambient Telemetry Motes
    const moteCount = 36
    const motePositions = new Float32Array(moteCount * 3)
    for (let m = 0; m < moteCount * 3; m += 3) {
      motePositions[m] = (seededRandom(m + 1) - 0.5) * offsetX * 1.7
      motePositions[m + 1] = 6 + seededRandom(m + 3) * 28
      motePositions[m + 2] = (seededRandom(m + 5) - 0.5) * offsetZ * 1.5
    }
    motesGeometry = new THREE.BufferGeometry()
    motesGeometry.setAttribute('position', new THREE.BufferAttribute(motePositions, 3))
    motesMaterial = new THREE.PointsMaterial({
      size: isDarkMode.value ? 6.5 : 5.0,
      map: hubTexture,
      transparent: true,
      opacity: isDarkMode.value ? 0.85 : 0.55,
      depthWrite: false,
      blending: THREE.NormalBlending
    })
    motesMesh = new THREE.Points(motesGeometry, motesMaterial)
    scene.add(motesMesh)

    updateMaterialStyle()
  },
  onRender({ elapsed, delta }) {
    if (!hubGeometry || !floorGeometry) return

    const isPaused = props.paused
    const dt = isPaused ? 0 : Math.min(delta || 0.016, 0.05)
    const time = elapsed * 1.05

    // Pointer smooth damping
    pointer.x += (pointer.targetX - pointer.x) * 0.05
    pointer.y += (pointer.targetY - pointer.y) * 0.05

    // Parallax camera tilt
    if (cameraRef) {
      cameraRef.position.x = pointer.x * 14
      cameraRef.position.y = 44 + pointer.y * 6
      cameraRef.lookAt(pointer.x * 4, 4 + pointer.y * 2, -4)
    }

    // 1. Animate Asset Constellation Nodes (Breathing & Subtle 3D Orbit Floating)
    for (let i = 0; i < constellationNodes.length; i++) {
      const node = constellationNodes[i]
      if (!isPaused) {
        node.phase += dt * node.speed
      }
      const floatY = Math.sin(node.phase) * (node.isHub ? 1.3 : 0.8)
      const floatX = Math.cos(node.phase * 0.8) * (node.isHub ? 0.6 : 0.4)
      node.y = node.baseY + floatY
      node.x = node.baseX + floatX
    }

    // Update Hub Points
    const hubPosArray = hubGeometry.attributes.position.array
    for (let h = 0; h < HUB_SPECS.length; h++) {
      hubPosArray[h * 3 + 0] = constellationNodes[h].x
      hubPosArray[h * 3 + 1] = constellationNodes[h].y
      hubPosArray[h * 3 + 2] = constellationNodes[h].z
    }
    hubGeometry.attributes.position.needsUpdate = true

    // Update Satellite Nodes
    const satPosArray = nodeGeometry.attributes.position.array
    const satCount = constellationNodes.length - HUB_SPECS.length
    for (let s = 0; s < satCount; s++) {
      const n = constellationNodes[HUB_SPECS.length + s]
      satPosArray[s * 3 + 0] = n.x
      satPosArray[s * 3 + 1] = n.y
      satPosArray[s * 3 + 2] = n.z
    }
    nodeGeometry.attributes.position.needsUpdate = true

    // Update Topology Links
    const linkPosArray = linkGeometry.attributes.position.array
    for (let l = 0; l < constellationLinks.length; l++) {
      const link = constellationLinks[l]
      const fNode = constellationNodes[link.from]
      const tNode = constellationNodes[link.to]
      linkPosArray[l * 6 + 0] = fNode.x
      linkPosArray[l * 6 + 1] = fNode.y
      linkPosArray[l * 6 + 2] = fNode.z
      linkPosArray[l * 6 + 3] = tNode.x
      linkPosArray[l * 6 + 4] = tNode.y
      linkPosArray[l * 6 + 5] = tNode.z
    }
    linkGeometry.attributes.position.needsUpdate = true

    // 2. Animate Flowing Telemetry Pulses
    const pulsePosArray = pulseGeometry.attributes.position.array
    for (let p = 0; p < telemetryPulses.length; p++) {
      const pulse = telemetryPulses[p]
      if (!isPaused) {
        pulse.progress += pulse.speed * dt
        if (pulse.progress >= 1) {
          pulse.progress = 0
          const curLink = constellationLinks[pulse.linkIdx]
          const targetNode = pulse.forward ? curLink.to : curLink.from
          const nextLinkCandidates = nodeToLinks[targetNode]
          if (nextLinkCandidates && nextLinkCandidates.length > 0) {
            const pick = nextLinkCandidates[Math.floor(seededRandom(p + time) * nextLinkCandidates.length)]
            pulse.linkIdx = pick
            pulse.forward = constellationLinks[pick].from === targetNode
          } else {
            pulse.forward = !pulse.forward
          }
        }
      }

      const activeLink = constellationLinks[pulse.linkIdx]
      const fNode = constellationNodes[activeLink.from]
      const tNode = constellationNodes[activeLink.to]
      const t = pulse.forward ? pulse.progress : (1 - pulse.progress)

      pulsePosArray[p * 3 + 0] = fNode.x + (tNode.x - fNode.x) * t
      pulsePosArray[p * 3 + 1] = fNode.y + (tNode.y - fNode.y) * t
      pulsePosArray[p * 3 + 2] = fNode.z + (tNode.z - fNode.z) * t
    }
    pulseGeometry.attributes.position.needsUpdate = true

    // 3. Animate Radar Scanner Rings (Expanding concentric cyber sweep)
    const maxRadarRadius = 86
    const dark = isDarkMode.value
    radarRings.forEach((ring) => {
      if (!isPaused) {
        ring.offset += dt * 14.0
        if (ring.offset >= maxRadarRadius) {
          ring.offset -= maxRadarRadius
        }
      }
      const r = ring.offset + 2.5
      ring.mesh.scale.set(r, 1, r)
      ring.mesh.rotation.y = time * 0.08

      const alphaProgress = Math.max(0, 1 - r / maxRadarRadius)
      ring.material.opacity = Math.pow(alphaProgress, 1.35) * (dark ? 0.42 : 0.22)
    })

    // 4. Animate Base Spatial Coordinate Floor
    const floorPositions = floorGeometry.attributes.position.array
    const countX = props.particleCountX
    const countY = props.particleCountY
    const floorBaseY = -18

    let idx = 0
    for (let ix = 0; ix < countX; ix++) {
      for (let iy = 0; iy < countY; iy++) {
        const u = ix / countX
        const v = iy / countY
        const px = floorPositions[idx]
        const pz = floorPositions[idx + 2]
        const distCenter = Math.sqrt(px * px + pz * pz)

        // Expansive digital ripple
        const ripple = Math.sin(distCenter * 0.09 - time * 1.3) * 1.6
        const waveX = Math.cos(u * 5.0 + time * 0.8) * 0.9
        const waveZ = Math.sin(v * 4.2 + time * 0.7) * 0.7

        let pointerShift = 0
        if (pointer.active || Math.abs(pointer.x) > 0.01 || Math.abs(pointer.y) > 0.01) {
          const normPx = (u - 0.5) * 2.2
          const normPy = (v - 0.5) * 2.2
          const distSq = (normPx - pointer.x) * (normPx - pointer.x) + (normPy - pointer.y) * (normPy - pointer.y)
          if (distSq < 1.2) {
            pointerShift = Math.cos(Math.sqrt(distSq) * Math.PI * 0.5) * 6.5
          }
        }

        floorPositions[idx + 1] = floorBaseY + ripple + waveX + waveZ + pointerShift
        idx += 3
      }
    }
    floorGeometry.attributes.position.needsUpdate = true

    // 5. Animate Floating Telemetry Motes
    if (motesGeometry) {
      const motesPos = motesGeometry.attributes.position.array
      for (let m = 0; m < motesPos.length; m += 3) {
        if (!isPaused) {
          motesPos[m + 1] += Math.sin(time * 1.2 + m) * 0.035
        }
      }
      motesGeometry.attributes.position.needsUpdate = true
    }
  }
})

onMounted(() => {
  hostElement = containerRef.value?.parentElement
  if (hostElement) {
    hostElement.addEventListener('pointermove', handlePointerMove, { passive: true })
    hostElement.addEventListener('pointerleave', handlePointerLeave, { passive: true })
  }
  if (props.paused) {
    stop()
  }
})

onBeforeUnmount(() => {
  if (hostElement) {
    hostElement.removeEventListener('pointermove', handlePointerMove)
    hostElement.removeEventListener('pointerleave', handlePointerLeave)
    hostElement = null
  }

  // Safe disposal of Three.js objects
  hubGeometry?.dispose()
  nodeGeometry?.dispose()
  linkGeometry?.dispose()
  pulseGeometry?.dispose()
  floorGeometry?.dispose()
  floorLineGeometry?.dispose()
  motesGeometry?.dispose()

  hubMaterial?.dispose()
  nodeMaterial?.dispose()
  linkMaterial?.dispose()
  pulseMaterial?.dispose()
  floorMaterial?.dispose()
  floorLineMaterial?.dispose()
  motesMaterial?.dispose()

  radarRings.forEach((ring) => {
    ring.mesh?.geometry?.dispose()
    ring.material?.dispose()
  })
  radarRings.length = 0

  hubTexture?.dispose()
  nodeTexture?.dispose()
  pulseTexture?.dispose()
})

watch(() => props.paused, (isPaused) => {
  if (isPaused) {
    stop()
  } else {
    start()
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
