<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Activity, Bot, Radar, ShieldAlert, Sparkles, Tv, Zap } from '@lucide/vue'
import type { Node } from '../types'

const props = defineProps<{
  nodes: Node[]
  activeNode?: Node | null
}>()

const emit = defineEmits<{
  (e: 'selectNode', node: Node): void
}>()

const canvasRef = ref<HTMLCanvasElement | null>()
const containerRef = ref<HTMLDivElement | null>()
let animationId: number | null = null
let angle = 0

const selectedNodeId = ref<string>(props.activeNode?.id || props.nodes[0]?.id || '')

const currentNode = computed(() => {
  return props.nodes.find((n) => n.id === selectedNodeId.value) || props.nodes[0]
})

watch(() => props.activeNode, (newNode) => {
  if (newNode) selectedNodeId.value = newNode.id
})

// Metrics computed for the 3D radar from real node data
const radarMetrics = computed(() => {
  const n = currentNode.value
  if (!n) return []
  const streamingEntries = Object.values(n.unlocks?.streaming ?? {})
  const streamingUnlocked = streamingEntries.filter((u) => u.status === 'available').length
  const streamingCount = streamingEntries.length
  const streamingRatio = streamingCount > 0 
    ? Math.round((streamingUnlocked / streamingCount) * 100)
    : (n.netflix === 'available' ? 100 : n.netflix === 'limited' ? 50 : 0)

  const aiEntries = Object.values(n.unlocks?.ai ?? {})
  const aiUnlocked = aiEntries.filter((u) => u.status === 'available').length
  const aiCount = aiEntries.length
  const aiRatio = aiCount > 0
    ? Math.round((aiUnlocked / aiCount) * 100)
    : (n.chatgpt === 'available' ? 100 : 0)

  return [
    { label: 'IP 纯净度', value: Math.max(0, 100 - n.risk), max: 100, color: '#10b981', unit: '分' },
    { label: 'AI 模型解锁', value: aiRatio, max: 100, color: '#a855f7', unit: '%' },
    { label: '流媒体原生', value: streamingRatio, max: 100, color: '#38bdf8', unit: '%' },
    { label: 'DNSBL 安全度', value: Math.max(0, 100 - n.dnsbl * 12), max: 100, color: '#f59e0b', unit: '分' },
    { label: '网络稳定性', value: n.status === 'online' ? 98 : n.status === 'warning' ? 75 : n.status === 'alert' ? 45 : 10, max: 100, color: '#06b6d4', unit: '%' },
  ]
})


const drawRadar = () => {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const width = canvas.width
  const height = canvas.height
  const cx = width / 2
  const cy = height / 2 + 10
  const radius = Math.min(width, height) * 0.4

  ctx.clearRect(0, 0, width, height)

  // Perspective 3D Ellipse factor (simulating 3D tilt plane)
  const scaleY = 0.55

  // 1. Draw Radar Base Grid Rings (3D tilted)
  ctx.save()
  ctx.translate(cx, cy)
  ctx.scale(1, scaleY)

  // Background glowing grid
  for (let r = radius * 0.25; r <= radius; r += radius * 0.25) {
    ctx.strokeStyle = 'rgba(56, 189, 248, 0.15)'
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.arc(0, 0, r, 0, Math.PI * 2)
    ctx.stroke()
  }

  // Cross lines
  ctx.strokeStyle = 'rgba(100, 116, 139, 0.2)'
  ctx.beginPath()
  ctx.moveTo(-radius * 1.1, 0)
  ctx.lineTo(radius * 1.1, 0)
  ctx.moveTo(0, -radius * 1.1)
  ctx.lineTo(0, radius * 1.1)
  ctx.stroke()

  // 2. Rotating Holographic Sweep Beam
  angle += 0.035
  const sweepGrad = ctx.createRadialGradient(0, 0, 0, 0, 0, radius)
  sweepGrad.addColorStop(0, 'rgba(56, 189, 248, 0.35)')
  sweepGrad.addColorStop(0.8, 'rgba(168, 85, 247, 0.15)')
  sweepGrad.addColorStop(1, 'rgba(0, 0, 0, 0)')

  ctx.fillStyle = sweepGrad
  ctx.beginPath()
  ctx.moveTo(0, 0)
  ctx.arc(0, 0, radius, angle, angle + Math.PI * 0.45)
  ctx.closePath()
  ctx.fill()

  // Sweep Leading Ray
  ctx.strokeStyle = 'rgba(56, 189, 248, 0.85)'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(0, 0)
  ctx.lineTo(Math.cos(angle + Math.PI * 0.45) * radius, Math.sin(angle + Math.PI * 0.45) * radius)
  ctx.stroke()

  // 3. Draw Polygon Radar Data Profile for Selected Node
  const metrics = radarMetrics.value
  if (metrics.length > 0) {
    ctx.beginPath()
    for (let i = 0; i < metrics.length; i++) {
      const theta = (i / metrics.length) * Math.PI * 2 - Math.PI / 2
      const valRatio = metrics[i].value / metrics[i].max
      const px = Math.cos(theta) * radius * valRatio
      const py = Math.sin(theta) * radius * valRatio
      if (i === 0) ctx.moveTo(px, py)
      else ctx.lineTo(px, py)
    }
    ctx.closePath()
    ctx.fillStyle = 'rgba(56, 189, 248, 0.22)'
    ctx.fill()
    ctx.strokeStyle = '#38bdf8'
    ctx.lineWidth = 2
    ctx.stroke()

    // Metric vertex points
    for (let i = 0; i < metrics.length; i++) {
      const theta = (i / metrics.length) * Math.PI * 2 - Math.PI / 2
      const valRatio = metrics[i].value / metrics[i].max
      const px = Math.cos(theta) * radius * valRatio
      const py = Math.sin(theta) * radius * valRatio

      ctx.fillStyle = metrics[i].color
      ctx.shadowColor = metrics[i].color
      ctx.shadowBlur = 8
      ctx.beginPath()
      ctx.arc(px, py, 4, 0, Math.PI * 2)
      ctx.fill()
      ctx.shadowBlur = 0
    }
  }

  // Draw other nodes as 3D radar blips
  props.nodes.forEach((n, idx) => {
    const theta = (idx / props.nodes.length) * Math.PI * 2 + 0.3
    const dist = (0.45 + ((idx * 7) % 5) * 0.1) * radius
    const px = Math.cos(theta) * dist
    const py = Math.sin(theta) * dist

    const isCurrent = n.id === currentNode.value?.id
    ctx.fillStyle = isCurrent ? '#38bdf8' : n.risk >= 60 ? '#ef4444' : n.risk >= 35 ? '#f59e0b' : '#10b981'
    ctx.beginPath()
    ctx.arc(px, py, isCurrent ? 5 : 3, 0, Math.PI * 2)
    ctx.fill()
  })

  ctx.restore()

  animationId = requestAnimationFrame(drawRadar)
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
  handleResize()
  window.addEventListener('resize', handleResize)
  drawRadar()
})

onBeforeUnmount(() => {
  if (animationId) cancelAnimationFrame(animationId)
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="radar-3d-card">
    <div class="radar-head">
      <div class="head-left">
        <Radar :size="17" class="radar-icon" />
        <div>
          <h3>3D 全息多维威胁与解锁雷达</h3>
          <p>实时多轴态势研判与 IP 纯净度评估</p>
        </div>
      </div>

      <!-- Node Selector -->
      <div class="node-select-wrap">
        <select v-model="selectedNodeId" @change="emit('selectNode', currentNode)">
          <option v-for="node in nodes" :key="node.id" :value="node.id">
            {{ node.name }} ({{ node.region }}) · 风险 {{ node.risk }}
          </option>
        </select>
      </div>
    </div>

    <div class="radar-body">
      <div ref="containerRef" class="radar-canvas-wrap">
        <canvas ref="canvasRef" class="radar-canvas"></canvas>
      </div>

      <!-- Metric Telemetry Bars -->
      <div class="radar-metrics-panel">
        <div v-for="m in radarMetrics" :key="m.label" class="metric-row">
          <div class="metric-labels">
            <span class="m-name">{{ m.label }}</span>
            <strong class="m-val" :style="{ color: m.color }">
              {{ m.value }} {{ m.unit }}
            </strong>
          </div>
          <div class="metric-track">
            <div
              class="metric-fill"
              :style="{ width: `${m.value}%`, backgroundColor: m.color, boxShadow: `0 0 8px ${m.color}` }"
            ></div>
          </div>
        </div>

        <div class="radar-node-brief">
          <div class="brief-item">
            <span>当前节点:</span>
            <strong>{{ currentNode?.name }}</strong>
          </div>
          <div class="brief-item">
            <span>服务商 / ASN:</span>
            <code>{{ currentNode?.provider }} · AS{{ currentNode?.asn }}</code>
          </div>
          <div class="brief-item">
            <span>IP 身份变化:</span>
            <strong :class="currentNode?.ip_changed ? 'text-danger' : 'text-good'">
              {{ currentNode?.ip_changed ? '近期曾变动' : '持续稳定' }}
            </strong>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.radar-3d-card {
  background: var(--surface, #1e242b);
  border: 1px solid var(--border, #343c45);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: var(--shadow, 0 4px 16px rgba(0, 0, 0, 0.1));
}

.radar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border, #343c45);
  gap: 12px;
  flex-wrap: wrap;
}

.head-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.radar-icon {
  color: #38bdf8;
}
.head-left h3 {
  margin: 0;
  font-size: 14px;
  color: var(--text, #f8fafc);
}
.head-left p {
  margin: 2px 0 0;
  font-size: 11px;
  color: var(--muted, #94a3b8);
}

.node-select-wrap select {
  height: 32px;
  padding: 0 10px;
  background: var(--surface-2, #242b33);
  color: var(--text, #f8fafc);
  border: 1px solid var(--border, #343c45);
  border-radius: 5px;
  font-size: 12px;
  font-family: 'Fira Code', monospace;
  outline: none;
}

.radar-body {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  padding: 14px;
  gap: 18px;
  align-items: center;
}

.radar-canvas-wrap {
  width: 100%;
  height: 250px;
  position: relative;
  border-radius: 6px;
  background: radial-gradient(circle at center, rgba(14, 165, 233, 0.05) 0%, rgba(15, 23, 42, 0.6) 100%);
  border: 1px solid rgba(56, 189, 248, 0.15);
}

.radar-canvas {
  width: 100%;
  height: 100%;
  display: block;
}

.radar-metrics-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.metric-row {
  display: grid;
  gap: 4px;
}

.metric-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
}
.m-name {
  color: var(--muted, #94a3b8);
}
.m-val {
  font-family: 'Fira Code', monospace;
  font-weight: 700;
}

.metric-track {
  height: 5px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 3px;
  overflow: hidden;
}
.metric-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.radar-node-brief {
  margin-top: 6px;
  padding: 10px 12px;
  background: var(--surface-2, #242b33);
  border: 1px solid var(--border, #343c45);
  border-radius: 6px;
  display: grid;
  gap: 6px;
  font-size: 11px;
}

.brief-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.brief-item span {
  color: var(--muted, #94a3b8);
}
.brief-item strong {
  color: var(--text, #f8fafc);
}
.brief-item code {
  color: #38bdf8;
  font-family: 'Fira Code', monospace;
}

.text-good { color: var(--good, #10b981) !important; }
.text-danger { color: var(--danger, #ef4444) !important; }

@media (max-width: 860px) {
  .radar-body {
    grid-template-columns: 1fr;
  }
}
</style>
