<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import { Eye, EyeOff, Globe2, Maximize2, Pause, Play, RotateCcw, ShieldCheck, Zap } from '@lucide/vue'
import type { Node } from '../types'

const props = defineProps<{
  nodes: Node[]
  selectedId?: string | null
}>()

const emit = defineEmits<{
  (e: 'select', node: Node): void
}>()

const canvasRef = ref<HTMLCanvasElement | null>()
const containerRef = ref<HTMLDivElement | null>()
const hoveredNode = ref<Node | null>(null)
const mousePos = ref({ x: 0, y: 0, inside: false })

// 3D Globe State
const autoRotate = ref(true)
const showArcs = ref(true)
const showGrid = ref(true)
const showLabels = ref(true)
const zoomLevel = ref(1.0)

let animationFrameId: number | null = null
let rotY = 0.8
let rotX = 0.25
let targetRotY = 0.8
let targetRotX = 0.25
let isDragging = false
let lastMouseX = 0
let lastMouseY = 0
let dragDistance = 0

// Projected Node Cache for hit detection
interface ProjectedNode {
  node: Node
  x: number
  y: number
  z: number
  visible: boolean
  radius: number
}
let projectedNodes: ProjectedNode[] = []

// Pre-computed world landmass dots (approximate continent clusters)
const landmassPoints: { lat: number; lng: number }[] = []
const initLandmass = () => {
  if (landmassPoints.length > 0) return
  // North America
  for (let lat = 25; lat <= 60; lat += 3.5) {
    for (let lng = -125; lng <= -70; lng += 4.5) {
      if ((lat > 45 && lng < -120) || (lat < 30 && lng < -105)) continue
      landmassPoints.push({ lat, lng })
    }
  }
  // South America
  for (let lat = -50; lat <= 10; lat += 4) {
    for (let lng = -75; lng <= -35; lng += 4.5) {
      if (lat < -20 && lng > -55) continue
      landmassPoints.push({ lat, lng })
    }
  }
  // Europe
  for (let lat = 36; lat <= 65; lat += 3) {
    for (let lng = -10; lng <= 40; lng += 3.5) {
      landmassPoints.push({ lat, lng })
    }
  }
  // Asia
  for (let lat = 10; lat <= 65; lat += 3.5) {
    for (let lng = 40; lng <= 140; lng += 4) {
      if (lat < 20 && lng > 90 && lng < 100) continue
      landmassPoints.push({ lat, lng })
    }
  }
  // Africa
  for (let lat = -35; lat <= 35; lat += 4) {
    for (let lng = -15; lng <= 50; lng += 4.5) {
      if (lat > 20 && lng > 35) continue
      landmassPoints.push({ lat, lng })
    }
  }
  // Oceania / Australia
  for (let lat = -40; lat <= -12; lat += 3.5) {
    for (let lng = 112; lng <= 155; lng += 4) {
      landmassPoints.push({ lat, lng })
    }
  }
}

// Coordinate conversions
const toCartesian = (lat: number, lng: number, radius: number) => {
  const phi = (90 - lat) * (Math.PI / 180)
  const theta = (lng + 180) * (Math.PI / 180)
  const x = -(radius * Math.sin(phi) * Math.cos(theta))
  const z = radius * Math.sin(phi) * Math.sin(theta)
  const y = radius * Math.cos(phi)
  return { x, y, z }
}

const rotate3D = (x: number, y: number, z: number, rx: number, ry: number) => {
  // Rotate around X-axis
  const cosRx = Math.cos(rx)
  const sinRx = Math.sin(rx)
  const y1 = y * cosRx - z * sinRx
  const z1 = y * sinRx + z * cosRx

  // Rotate around Y-axis
  const cosRy = Math.cos(ry)
  const sinRy = Math.sin(ry)
  const x2 = x * cosRy + z1 * sinRy
  const z2 = -x * sinRy + z1 * cosRy

  return { x: x2, y: y1, z: z2 }
}

let tick = 0

const render = () => {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const width = canvas.width
  const height = canvas.height
  const cx = width / 2
  const cy = height / 2
  const baseRadius = Math.min(width, height) * 0.38 * zoomLevel.value

  ctx.clearRect(0, 0, width, height)

  // Smooth rotation
  if (autoRotate.value && !isDragging) {
    targetRotY += 0.0035
  }
  rotY += (targetRotY - rotY) * 0.1
  rotX += (targetRotX - rotX) * 0.1
  tick += 0.02

  // 1. Draw Globe Ambient Glow & Atmosphere
  const glowGrad = ctx.createRadialGradient(cx, cy, baseRadius * 0.7, cx, cy, baseRadius * 1.35)
  glowGrad.addColorStop(0, 'rgba(16, 185, 129, 0.03)')
  glowGrad.addColorStop(0.5, 'rgba(59, 130, 246, 0.08)')
  glowGrad.addColorStop(0.85, 'rgba(168, 85, 247, 0.05)')
  glowGrad.addColorStop(1, 'rgba(0, 0, 0, 0)')
  ctx.fillStyle = glowGrad
  ctx.beginPath()
  ctx.arc(cx, cy, baseRadius * 1.35, 0, Math.PI * 2)
  ctx.fill()

  // Globe Sphere Background
  const sphereGrad = ctx.createRadialGradient(
    cx - baseRadius * 0.3,
    cy - baseRadius * 0.3,
    baseRadius * 0.1,
    cx,
    cy,
    baseRadius
  )
  sphereGrad.addColorStop(0, 'rgba(20, 27, 38, 0.95)')
  sphereGrad.addColorStop(0.7, 'rgba(15, 20, 28, 0.98)')
  sphereGrad.addColorStop(1, 'rgba(8, 11, 16, 1)')
  ctx.fillStyle = sphereGrad
  ctx.beginPath()
  ctx.arc(cx, cy, baseRadius, 0, Math.PI * 2)
  ctx.fill()
  ctx.strokeStyle = 'rgba(56, 189, 248, 0.25)'
  ctx.lineWidth = 1.5
  ctx.stroke()

  // 2. Draw Latitude & Longitude Wireframe Grid
  if (showGrid.value) {
    ctx.strokeStyle = 'rgba(100, 116, 139, 0.15)'
    ctx.lineWidth = 0.8

    // Latitudes
    for (let lat = -60; lat <= 60; lat += 30) {
      ctx.beginPath()
      let first = true
      for (let lng = -180; lng <= 180; lng += 6) {
        const p = toCartesian(lat, lng, baseRadius)
        const r = rotate3D(p.x, p.y, p.z, rotX, rotY)
        if (r.z > 0) {
          if (first) {
            ctx.moveTo(cx + r.x, cy + r.y)
            first = false
          } else {
            ctx.lineTo(cx + r.x, cy + r.y)
          }
        } else {
          first = true
        }
      }
      ctx.stroke()
    }

    // Longitudes
    for (let lng = -180; lng < 180; lng += 45) {
      ctx.beginPath()
      let first = true
      for (let lat = -80; lat <= 80; lat += 5) {
        const p = toCartesian(lat, lng, baseRadius)
        const r = rotate3D(p.x, p.y, p.z, rotX, rotY)
        if (r.z > 0) {
          if (first) {
            ctx.moveTo(cx + r.x, cy + r.y)
            first = false
          } else {
            ctx.lineTo(cx + r.x, cy + r.y)
          }
        } else {
          first = true
        }
      }
      ctx.stroke()
    }
  }

  // 3. Draw Continents Point Cloud
  for (const pt of landmassPoints) {
    const p = toCartesian(pt.lat, pt.lng, baseRadius)
    const r = rotate3D(p.x, p.y, p.z, rotX, rotY)
    if (r.z > 0) {
      const alpha = Math.min(1, Math.max(0.1, r.z / baseRadius))
      ctx.fillStyle = `rgba(148, 163, 184, ${alpha * 0.45})`
      ctx.beginPath()
      ctx.arc(cx + r.x, cy + r.y, 1.2, 0, Math.PI * 2)
      ctx.fill()
    }
  }

  // 4. Project & Cache Nodes
  projectedNodes = []
  for (const n of props.nodes) {
    let lat = n.latitude
    let lng = n.longitude

    if (lat === undefined || lng === undefined || (lat === 0 && lng === 0)) {
      const text = `${n.region} ${n.country_code} ${n.name}`.toLowerCase()
      if (text.includes('hk') || text.includes('hong kong') || text.includes('香港')) {
        lat = 22.3193; lng = 114.1694
      } else if (text.includes('tw') || text.includes('taiwan') || text.includes('台湾')) {
        lat = 25.0330; lng = 121.5654
      } else if (text.includes('jp') || text.includes('tokyo') || text.includes('东京') || text.includes('日本')) {
        lat = 35.6762; lng = 139.6503
      } else if (text.includes('sg') || text.includes('singapore') || text.includes('新加坡')) {
        lat = 1.3521; lng = 103.8198
      } else if (text.includes('de') || text.includes('fra') || text.includes('germany') || text.includes('法兰克福') || text.includes('德国')) {
        lat = 50.1109; lng = 8.6821
      } else if (text.includes('uk') || text.includes('gb') || text.includes('london') || text.includes('伦敦') || text.includes('英国')) {
        lat = 51.5074; lng = -0.1278
      } else if (text.includes('nl') || text.includes('ams') || text.includes('amsterdam') || text.includes('阿姆斯特丹') || text.includes('荷兰')) {
        lat = 52.3676; lng = 4.9041
      } else if (text.includes('lax') || text.includes('los angeles') || text.includes('洛杉矶')) {
        lat = 34.0522; lng = -118.2437
      } else if (text.includes('us') || text.includes('america') || text.includes('美国')) {
        lat = 37.7749; lng = -122.4194
      } else if (text.includes('au') || text.includes('sydney') || text.includes('悉尼') || text.includes('澳大利亚')) {
        lat = -33.8688; lng = 151.2093
      } else if (text.includes('kr') || text.includes('seoul') || text.includes('首尔') || text.includes('韩国')) {
        lat = 37.5665; lng = 126.9780
      } else {
        lat = 20 + ((n.name.charCodeAt(0) || 10) % 30)
        lng = 80 + ((n.name.charCodeAt(1) || 20) % 60)
      }
    }

    const p = toCartesian(lat, lng, baseRadius)
    const r = rotate3D(p.x, p.y, p.z, rotX, rotY)
    projectedNodes.push({
      node: n,
      x: cx + r.x,
      y: cy + r.y,
      z: r.z,
      visible: r.z > -baseRadius * 0.2, // slightly visible over edge
      radius: 6,
    })
  }


  // 5. Draw Flightline Arcs between visible nodes
  if (showArcs.value) {
    const activeNodes = projectedNodes.filter((p) => p.visible && p.z > 0)
    for (let i = 0; i < activeNodes.length; i++) {
      for (let j = i + 1; j < activeNodes.length; j++) {
        // Connect select major hub lines
        const n1 = activeNodes[i]
        const n2 = activeNodes[j]
        const dist = Math.hypot(n1.x - n2.x, n1.y - n2.y)
        if (dist > 40 && dist < baseRadius * 1.6) {
          const midX = (n1.x + n2.x) / 2
          const midY = (n1.y + n2.y) / 2
          // Elevate arc control point outward from center
          const vecX = midX - cx
          const vecY = midY - cy
          const ctrlX = midX + vecX * 0.25
          const ctrlY = midY + vecY * 0.25

          // Arc gradient
          const arcGrad = ctx.createLinearGradient(n1.x, n1.y, n2.x, n2.y)
          arcGrad.addColorStop(0, 'rgba(56, 189, 248, 0.1)')
          arcGrad.addColorStop(0.5, 'rgba(168, 85, 247, 0.4)')
          arcGrad.addColorStop(1, 'rgba(56, 189, 248, 0.1)')

          ctx.strokeStyle = arcGrad
          ctx.lineWidth = 1.2
          ctx.beginPath()
          ctx.moveTo(n1.x, n1.y)
          ctx.quadraticCurveTo(ctrlX, ctrlY, n2.x, n2.y)
          ctx.stroke()

          // Moving photon particle along quadratic bezier
          const t = (tick * 0.4 + (i * 3 + j) * 0.15) % 1
          const px = (1 - t) * (1 - t) * n1.x + 2 * (1 - t) * t * ctrlX + t * t * n2.x
          const py = (1 - t) * (1 - t) * n1.y + 2 * (1 - t) * t * ctrlY + t * t * n2.y

          ctx.fillStyle = '#38bdf8'
          ctx.shadowColor = '#38bdf8'
          ctx.shadowBlur = 6
          ctx.beginPath()
          ctx.arc(px, py, 2, 0, Math.PI * 2)
          ctx.fill()
          ctx.shadowBlur = 0
        }
      }
    }
  }

  // 6. Draw Nodes (Beacons, Pulse Rings, Labels)
  for (const pn of projectedNodes) {
    if (!pn.visible) continue
    const n = pn.node
    const isFront = pn.z > 0
    const depthAlpha = isFront ? Math.min(1, Math.max(0.3, pn.z / baseRadius)) : 0.2
    const isHovered = hoveredNode.value?.id === n.id
    const isSelected = props.selectedId === n.id

    let color = '#10b981' // green
    if (n.status === 'alert' || n.risk >= 60) color = '#ef4444' // red
    else if (n.status === 'warning' || n.risk >= 35) color = '#f59e0b' // yellow
    else if (n.status === 'offline') color = '#64748b'

    // Pulsing Radar Rings around front nodes
    if (isFront) {
      const pulseProgress = (tick + n.risk * 0.1) % 2
      const ringRadius = 6 + pulseProgress * 14
      const ringAlpha = Math.max(0, (1 - pulseProgress / 2) * depthAlpha * 0.7)

      ctx.strokeStyle = color
      ctx.globalAlpha = ringAlpha
      ctx.lineWidth = 1.5
      ctx.beginPath()
      ctx.arc(pn.x, pn.y, ringRadius, 0, Math.PI * 2)
      ctx.stroke()
      ctx.globalAlpha = 1.0

      // Holographic Light Pillar / Beacon
      const pillarH = (isHovered || isSelected ? 34 : 22) * depthAlpha
      const beaconGrad = ctx.createLinearGradient(pn.x, pn.y, pn.x, pn.y - pillarH)
      beaconGrad.addColorStop(0, color)
      beaconGrad.addColorStop(1, 'rgba(255, 255, 255, 0)')

      ctx.strokeStyle = beaconGrad
      ctx.lineWidth = isHovered || isSelected ? 2.5 : 1.5
      ctx.beginPath()
      ctx.moveTo(pn.x, pn.y)
      ctx.lineTo(pn.x, pn.y - pillarH)
      ctx.stroke()

      // Beacon Top Light
      ctx.fillStyle = '#ffffff'
      ctx.shadowColor = color
      ctx.shadowBlur = 8
      ctx.beginPath()
      ctx.arc(pn.x, pn.y - pillarH, 2.5, 0, Math.PI * 2)
      ctx.fill()
      ctx.shadowBlur = 0
    }

    // Main Node Pin
    ctx.fillStyle = color
    ctx.shadowColor = color
    ctx.shadowBlur = isHovered || isSelected ? 12 : 6
    ctx.beginPath()
    ctx.arc(pn.x, pn.y, (isHovered || isSelected ? 5.5 : 4) * (isFront ? 1 : 0.7), 0, Math.PI * 2)
    ctx.fill()
    ctx.shadowBlur = 0

    // Label on Front
    if (isFront && showLabels.value && (isHovered || isSelected || depthAlpha > 0.6)) {
      ctx.font = '600 10px "Fira Code", monospace'
      ctx.fillStyle = isHovered || isSelected ? '#ffffff' : `rgba(226, 232, 240, ${depthAlpha})`
      ctx.textAlign = 'center'
      ctx.fillText(n.name, pn.x, pn.y + 16)

      ctx.font = '500 8.5px "Fira Sans", sans-serif'
      ctx.fillStyle = `rgba(148, 163, 184, ${depthAlpha * 0.9})`
      ctx.fillText(`${n.region} · AS${n.asn}`, pn.x, pn.y + 26)
    }
  }

  // 7. Raycast & Hit Detection on Mouse Move
  if (mousePos.value.inside) {
    let closest: Node | null = null
    let minDist = 22
    for (const pn of projectedNodes) {
      if (pn.z < 0) continue
      const d = Math.hypot(pn.x - mousePos.value.x, pn.y - mousePos.value.y)
      if (d < minDist) {
        minDist = d
        closest = pn.node
      }
    }
    hoveredNode.value = closest
  }

  animationFrameId = requestAnimationFrame(render)
}

// Mouse Drag & Inertia Handlers
const onMouseDown = (e: MouseEvent) => {
  isDragging = true
  lastMouseX = e.clientX
  lastMouseY = e.clientY
  dragDistance = 0
}

const onMouseMove = (e: MouseEvent) => {
  const rect = canvasRef.value?.getBoundingClientRect()
  if (!rect) return
  mousePos.value = {
    x: (e.clientX - rect.left) * (canvasRef.value!.width / rect.width),
    y: (e.clientY - rect.top) * (canvasRef.value!.height / rect.height),
    inside: true,
  }

  if (isDragging) {
    const dx = e.clientX - lastMouseX
    const dy = e.clientY - lastMouseY
    dragDistance += Math.abs(dx) + Math.abs(dy)
    targetRotY += dx * 0.006
    targetRotX = Math.max(-0.9, Math.min(0.9, targetRotX + dy * 0.006))
    lastMouseX = e.clientX
    lastMouseY = e.clientY
  }
}

const onMouseUp = () => {
  if (isDragging && dragDistance < 5 && hoveredNode.value) {
    emit('select', hoveredNode.value)
  }
  isDragging = false
}

const onMouseLeave = () => {
  isDragging = false
  mousePos.value.inside = false
  hoveredNode.value = null
}

const resetOrientation = () => {
  targetRotY = 0.8
  targetRotX = 0.25
  zoomLevel.value = 1.0
}

const adjustZoom = (delta: number) => {
  zoomLevel.value = Math.max(0.75, Math.min(1.45, zoomLevel.value + delta))
}

const handleResize = () => {
  const canvas = canvasRef.value
  const container = containerRef.value
  if (!canvas || !container) return
  const dpr = window.devicePixelRatio || 1
  const rect = container.getBoundingClientRect()
  canvas.width = rect.width * dpr
  canvas.height = rect.height * dpr
}

onMounted(() => {
  initLandmass()
  handleResize()
  window.addEventListener('resize', handleResize)
  render()
})

onBeforeUnmount(() => {
  if (animationFrameId) cancelAnimationFrame(animationFrameId)
  window.removeEventListener('resize', handleResize)
})

watch(() => props.selectedId, (id) => {
  if (!id) return
  const target = props.nodes.find((n) => n.id === id)
  if (target && target.latitude !== undefined && target.longitude !== undefined) {
    // Focus camera onto selected node
    targetRotY = -(target.longitude + 180) * (Math.PI / 180) + Math.PI / 2
    targetRotX = (target.latitude) * (Math.PI / 180) * 0.5
  }
})

const hudStyle = computed(() => {
  const dpr = typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1
  const width = canvasRef.value ? canvasRef.value.clientWidth : 600
  const height = canvasRef.value ? canvasRef.value.clientHeight : 450
  const left = Math.min(width - 240, Math.max(20, mousePos.value.x / dpr + 15))
  const top = Math.min(height - 180, Math.max(20, mousePos.value.y / dpr - 40))
  return {
    left: `${left}px`,
    top: `${top}px`,
  }
})
</script>


<template>
  <div ref="containerRef" class="globe-container" role="region" aria-label="3D 全球战情拓扑地球仪">
    <canvas
      ref="canvasRef"
      class="globe-canvas"
      @mousedown="onMouseDown"
      @mousemove="onMouseMove"
      @mouseup="onMouseUp"
      @mouseleave="onMouseLeave"
    ></canvas>

    <!-- Top Left Holographic Telemetry HUD -->
    <div class="hud-telemetry">
      <div class="hud-badge"><Zap :size="14" /><span>3D 全球拓扑全息雷达</span></div>
      <div class="hud-metric">
        <strong>{{ nodes.length }}</strong>
        <span>活跃战情节点</span>
      </div>
      <div class="hud-status-row">
        <span class="hud-dot live"></span>
        <small>实时流光信标正常 · 60 FPS</small>
      </div>
    </div>

    <!-- Top Right 3D Controls -->
    <div class="globe-controls">
      <button
        class="ctrl-btn"
        :class="{ active: autoRotate }"
        :title="autoRotate ? '暂停自动旋转' : '开启自动旋转'"
        aria-label="切换自动旋转"
        @click="autoRotate = !autoRotate"
      >
        <Pause v-if="autoRotate" :size="15" />
        <Play v-else :size="15" />
      </button>
      <button
        class="ctrl-btn"
        :class="{ active: showArcs }"
        title="切换航线光柱"
        aria-label="切换航线光柱"
        @click="showArcs = !showArcs"
      >
        <Zap :size="15" />
      </button>
      <button
        class="ctrl-btn"
        :class="{ active: showLabels }"
        title="切换节点标签"
        aria-label="切换节点标签"
        @click="showLabels = !showLabels"
      >
        <Eye v-if="showLabels" :size="15" />
        <EyeOff v-else :size="15" />
      </button>
      <button class="ctrl-btn" title="放大" aria-label="放大地球" @click="adjustZoom(0.1)">+</button>
      <button class="ctrl-btn" title="缩小" aria-label="缩小地球" @click="adjustZoom(-0.1)">-</button>
      <button class="ctrl-btn" title="重置视角" aria-label="重置地球视角" @click="resetOrientation">
        <RotateCcw :size="15" />
      </button>
    </div>

    <!-- Node Hover Holographic HUD Tooltip -->
    <Transition name="hud-fade">
      <div
        v-if="hoveredNode"
        class="node-hud-card"
        :style="hudStyle"
      >
        <div class="hud-card-head">

          <div>
            <span class="hud-country">{{ hoveredNode.country_code }}</span>
            <strong>{{ hoveredNode.name }}</strong>
          </div>
          <span class="hud-risk" :class="hoveredNode.risk >= 60 ? 'risk-high' : hoveredNode.risk >= 35 ? 'risk-mid' : 'risk-low'">
            {{ hoveredNode.risk }}
          </span>
        </div>
        <div class="hud-card-meta">
          <span>{{ hoveredNode.provider }} · {{ hoveredNode.region }}</span>
          <code>{{ hoveredNode.masked_ip }}</code>
        </div>
        <div class="hud-card-unlocks">
          <div class="unlock-mini-row">
            <span>Netflix:</span>
            <strong :class="`status-${hoveredNode.unlocks?.streaming?.netflix?.status ?? hoveredNode.netflix}`">
              {{ hoveredNode.unlocks?.streaming?.netflix?.quality ?? hoveredNode.netflix }}
            </strong>
          </div>
          <div class="unlock-mini-row">
            <span>ChatGPT:</span>
            <strong :class="`status-${hoveredNode.unlocks?.ai?.chatgpt?.status ?? hoveredNode.chatgpt}`">
              {{ hoveredNode.unlocks?.ai?.chatgpt?.quality ?? hoveredNode.chatgpt }}
            </strong>
          </div>
        </div>
        <div class="hud-card-foot">
          <small>点击查看完整 20+ 项解锁审计</small>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.globe-container {
  position: relative;
  width: 100%;
  height: 480px;
  min-height: 400px;
  background: radial-gradient(circle at center, #101620 0%, #090d13 100%);
  border: 1px solid rgba(56, 189, 248, 0.15);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: inset 0 0 40px rgba(0, 0, 0, 0.6), 0 8px 32px rgba(0, 0, 0, 0.35);
  user-select: none;
}

.globe-canvas {
  width: 100%;
  height: 100%;
  display: block;
  cursor: grab;
}

.globe-canvas:active {
  cursor: grabbing;
}

/* Telemetry HUD */
.hud-telemetry {
  position: absolute;
  top: 16px;
  left: 18px;
  pointer-events: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
  z-index: 10;
}

.hud-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: rgba(15, 23, 42, 0.75);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(56, 189, 248, 0.25);
  border-radius: 6px;
  color: #38bdf8;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.hud-metric {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-top: 2px;
}
.hud-metric strong {
  font-family: 'Fira Code', monospace;
  font-size: 26px;
  font-weight: 700;
  color: #f8fafc;
  line-height: 1;
}
.hud-metric span {
  font-size: 11px;
  color: #94a3b8;
}

.hud-status-row {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #64748b;
  font-size: 10px;
  font-family: 'Fira Code', monospace;
}
.hud-dot.live {
  width: 6px;
  height: 6px;
  background: #10b981;
  border-radius: 50%;
  box-shadow: 0 0 8px #10b981;
  animation: pulse-live 1.6s infinite ease-in-out;
}

@keyframes pulse-live {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.85); }
}

/* Controls */
.globe-controls {
  position: absolute;
  top: 16px;
  right: 18px;
  display: flex;
  gap: 6px;
  z-index: 10;
}

.ctrl-btn {
  width: 32px;
  height: 32px;
  padding: 0;
  display: inline-grid;
  place-items: center;
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #94a3b8;
  font-family: 'Fira Code', monospace;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.15s ease;
}

.ctrl-btn:hover {
  background: rgba(30, 41, 59, 0.9);
  color: #f8fafc;
  border-color: rgba(56, 189, 248, 0.4);
}

.ctrl-btn.active {
  background: rgba(14, 165, 233, 0.2);
  color: #38bdf8;
  border-color: rgba(56, 189, 248, 0.6);
  box-shadow: 0 0 10px rgba(56, 189, 248, 0.2);
}

/* Hover HUD Card */
.node-hud-card {
  position: absolute;
  z-index: 20;
  width: 230px;
  padding: 12px 14px;
  background: rgba(15, 23, 42, 0.92);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(56, 189, 248, 0.35);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45), 0 0 12px rgba(56, 189, 248, 0.15);
  pointer-events: none;
}

.hud-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding-bottom: 8px;
  margin-bottom: 8px;
}
.hud-card-head > div {
  display: flex;
  align-items: center;
  gap: 6px;
}
.hud-country {
  padding: 1px 4px;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  border-radius: 3px;
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  font-weight: 700;
}
.hud-card-head strong {
  font-size: 13px;
  color: #f8fafc;
}
.hud-risk {
  font-family: 'Fira Code', monospace;
  font-weight: 700;
  font-size: 14px;
}

.hud-card-meta {
  display: flex;
  justify-content: space-between;
  font-size: 10.5px;
  color: #94a3b8;
  margin-bottom: 8px;
}
.hud-card-meta code {
  color: #cbd5e1;
  font-family: 'Fira Code', monospace;
}

.hud-card-unlocks {
  display: grid;
  gap: 4px;
  background: rgba(0, 0, 0, 0.25);
  padding: 6px 8px;
  border-radius: 4px;
  margin-bottom: 6px;
}
.unlock-mini-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 10px;
}
.unlock-mini-row span {
  color: #94a3b8;
}
.unlock-mini-row strong {
  font-size: 9.5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 130px;
}

.status-available { color: #10b981; }
.status-limited { color: #f59e0b; }
.status-blocked { color: #ef4444; }

.hud-card-foot {
  text-align: center;
  color: #38bdf8;
  font-size: 9px;
  opacity: 0.85;
}

.hud-fade-enter-active,
.hud-fade-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}
.hud-fade-enter-from,
.hud-fade-leave-to {
  opacity: 0;
  transform: scale(0.96) translateY(4px);
}

@media (max-width: 768px) {
  .globe-container {
    height: 360px;
  }
  .hud-telemetry {
    top: 10px;
    left: 10px;
  }
  .globe-controls {
    top: 10px;
    right: 10px;
  }
}
</style>
