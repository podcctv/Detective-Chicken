<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Node, UnlockInfo } from '../types'
import MetalBadge from './MetalBadge.vue'
import { Shield, ShieldAlert, Cpu, Radio, Zap, ArrowUpRight } from '@lucide/vue'


const props = withDefaults(
  defineProps<{
    node: Node
    selected?: boolean
    interactive?: boolean
  }>(),
  {
    selected: false,
    interactive: true,
  }
)

const emit = defineEmits<{
  (e: 'select', node: Node): void
  (e: 'inspectService', payload: { node: Node; serviceId: string }): void
}>()

// 3D Parallax Tilt state
const cardRef = ref<HTMLElement | null>(null)
const rotX = ref(0)
const rotY = ref(0)
const glareX = ref(50)
const glareY = ref(50)
const isHovering = ref(false)

const onMouseMove = (e: MouseEvent) => {
  if (!props.interactive || !cardRef.value) return
  const rect = cardRef.value.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top
  const centerX = rect.width / 2
  const centerY = rect.height / 2

  rotX.value = -((y - centerY) / centerY) * 12
  rotY.value = ((x - centerX) / centerX) * 14
  glareX.value = (x / rect.width) * 100
  glareY.value = (y / rect.height) * 100
}

const onMouseEnter = () => {
  isHovering.value = true
}

const onMouseLeave = () => {
  isHovering.value = false
  rotX.value = 0
  rotY.value = 0
  glareX.value = 50
  glareY.value = 50
}

// Risk Level and Style
const riskCategory = computed(() => {
  if (props.node.risk <= 20) return { label: '极净优质', class: 'risk-pure', color: '#10b981' }
  if (props.node.risk <= 50) return { label: '低度风险', class: 'risk-low', color: '#38bdf8' }
  if (props.node.risk <= 75) return { label: '中度注意', class: 'risk-mid', color: '#f59e0b' }
  return { label: '高危关注', class: 'risk-high', color: '#ef4444' }
})

// Top Key Services to display on card front
const keyServices = computed(() => {
  const list = ['chatgpt', 'claude', 'deepseek', 'netflix', 'disney', 'youtube', 'prime', 'spotify']
  return list.map((id) => {
    const stream = props.node.unlocks?.streaming?.[id]
    const ai = props.node.unlocks?.ai?.[id]
    const found: UnlockInfo | undefined = stream || ai
    return {
      id,
      name: found?.name || id,
      status: found?.status || 'untested',
      region: found?.region || props.node.country_code || '',
      quality: found?.quality || '',
      latency: found?.latency_ms || 0,
    }
  })
})
</script>

<template>
  <div
    ref="cardRef"
    class="credit-card-wrap"
    :class="{ 'is-selected': selected }"
    @mousemove="onMouseMove"
    @mouseenter="onMouseEnter"
    @mouseleave="onMouseLeave"
    @click="emit('select', node)"
  >
    <div
      class="metal-credit-card"
      :style="{
        transform: `perspective(1000px) rotateX(${rotX}deg) rotateY(${rotY}deg) ${isHovering ? 'scale3d(1.02, 1.02, 1.02)' : 'scale3d(1, 1, 1)'}`,
      }"
    >
      <!-- Metallic Foil, Brushed Shading & Glare Reflection -->
      <div class="card-metal-sheen"></div>
      <div
        class="card-dynamic-glare"
        :style="{
          background: `radial-gradient(circle at ${glareX}% ${glareY}%, rgba(255, 255, 255, 0.22) 0%, rgba(255, 255, 255, 0.05) 40%, transparent 75%)`,
          opacity: isHovering ? 1 : 0.2,
        }"
      ></div>
      <div class="card-border-glow"></div>

      <!-- Top Header Row: EMV Microchip, Provider, Region Pill -->
      <div class="card-header">
        <div class="emv-chip-wrap" title="安全探测晶片">
          <div class="emv-chip">
            <div class="chip-circuit"></div>
          </div>
          <div class="nfc-waves" title="双栈实时流光探测">
            <Radio :size="16" class="nfc-icon" />
          </div>
        </div>

        <div class="card-issuer">
          <div class="issuer-title">{{ node.provider || 'VPS 节点资产' }}</div>
          <div class="issuer-region">
            <span class="region-flag-pill">{{ node.country_code || node.region || 'GL' }}</span>
            <span class="region-name">{{ node.region }}</span>
          </div>
        </div>

        <!-- Node Risk / Quality Holographic Medal -->
        <div class="risk-medal" :class="riskCategory.class">
          <div class="medal-ring">
            <span class="risk-number">{{ node.risk }}</span>
          </div>
          <span class="risk-text">{{ riskCategory.label }}</span>
        </div>
      </div>

      <!-- Card Core: Embossed Node Name & Masked IP (Credit Card Number Style) -->
      <div class="card-identity-block">
        <div class="node-embossed-name">{{ node.name }}</div>
        <div class="node-embossed-ip">
          <span class="ip-digits">{{ node.masked_ip || '0.0.0.0' }}</span>
          <span class="ip-family-badge">
            {{ (node.families?.length ? node.families : [node.family || 4]).map((f) => `IPv${f}`).join(' + ') }}
          </span>
        </div>
      </div>

      <!-- Card Lower Details: ASN, Organization, Coordinates -->
      <div class="card-network-row">
        <div class="meta-item">
          <span class="meta-label">AUTONOMOUS SYSTEM</span>
          <span class="meta-val">AS{{ node.asn || '00000' }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">ORGANIZATION</span>
          <span class="meta-val org-val" :title="node.organization">{{ node.organization || 'Direct Carrier' }}</span>
        </div>
        <div class="meta-item right">
          <span class="meta-label">STATUS</span>
          <span class="meta-val status-val" :class="node.status">
            <span class="live-dot"></span>
            {{ node.status === 'online' ? '在线' : node.status === 'warning' ? '注意' : '告警' }}
          </span>
        </div>
      </div>

      <!-- Card Footer: Metallic Unlock Badges Cluster -->
      <div class="card-unlocks-footer">
        <div class="unlocks-title-row">
          <span>REAL-TIME STREAMING & AI UNLOCKS</span>
          <span class="action-hint">点击卡片展开详情 <ArrowUpRight :size="12" /></span>
        </div>
        <div class="badges-scroll-row">
          <MetalBadge
            v-for="svc in keyServices"
            :key="svc.id"
            :service-id="svc.id"
            :name="svc.name"
            :status="svc.status"
            :region="svc.region"
            :quality="svc.quality"
            :latency-ms="svc.latency"
            size="sm"
            :show-label="false"
            @click.stop="emit('inspectService', { node, serviceId: svc.id })"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.credit-card-wrap {
  position: relative;
  border-radius: 16px;
  perspective: 1200px;
  cursor: pointer;
  user-select: none;
}

.metal-credit-card {
  position: relative;
  border-radius: 16px;
  padding: 18px 22px;
  background: linear-gradient(135deg, #182029 0%, #10161d 50%, #0a0e13 100%);
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow:
    0 16px 36px rgba(0, 0, 0, 0.6),
    0 4px 12px rgba(0, 0, 0, 0.4),
    inset 0 1px 1px rgba(255, 255, 255, 0.25),
    inset 0 -1px 2px rgba(0, 0, 0, 0.8);
  overflow: hidden;
  transition: transform 0.15s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.25s ease, border-color 0.25s ease;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.is-selected .metal-credit-card {
  border-color: #38bdf8;
  box-shadow:
    0 20px 48px rgba(0, 0, 0, 0.8),
    0 0 24px rgba(56, 189, 248, 0.35),
    inset 0 1px 2px rgba(255, 255, 255, 0.4);
}

/* Metallic Shader Layers */
.card-metal-sheen {
  position: absolute;
  inset: 0;
  background:
    repeating-linear-gradient(
      60deg,
      rgba(255, 255, 255, 0.015) 0px,
      rgba(255, 255, 255, 0.015) 2px,
      transparent 2px,
      transparent 6px
    ),
    radial-gradient(ellipse at 20% 0%, rgba(255, 255, 255, 0.08), transparent 70%);
  pointer-events: none;
}

.card-dynamic-glare {
  position: absolute;
  inset: 0;
  pointer-events: none;
  mix-blend-mode: overlay;
  transition: opacity 0.25s ease;
}

.card-border-glow {
  position: absolute;
  inset: 0;
  border-radius: 16px;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.08);
  pointer-events: none;
}

/* Header Row */
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.emv-chip-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* EMV Chip Simulation */
.emv-chip {
  width: 38px;
  height: 28px;
  background: linear-gradient(135deg, #d4af37 0%, #aa8010 50%, #f3e5ab 100%);
  border-radius: 5px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  box-shadow:
    inset 0 1px 2px rgba(255, 255, 255, 0.6),
    inset 0 -1px 2px rgba(0, 0, 0, 0.5),
    0 2px 6px rgba(0, 0, 0, 0.4);
  position: relative;
  overflow: hidden;
}
.chip-circuit {
  position: absolute;
  inset: 3px;
  border: 1px solid rgba(0, 0, 0, 0.35);
  border-radius: 3px;
}
.chip-circuit::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background: rgba(0, 0, 0, 0.35);
}

.nfc-waves {
  color: #94a3b8;
  opacity: 0.7;
}

.card-issuer {
  flex: 1;
}
.issuer-title {
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 1px;
}
.issuer-region {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 2px;
}
.region-flag-pill {
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  font-weight: 800;
  padding: 1px 5px;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  border: 1px solid rgba(56, 189, 248, 0.35);
  border-radius: 4px;
}
.region-name {
  font-size: 11px;
  color: #f1f5f9;
  font-weight: 600;
}

/* Risk Holographic Medal */
.risk-medal {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 20px;
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.1);
}
.medal-ring {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-family: 'Fira Code', monospace;
  font-size: 11px;
  font-weight: 800;
}
.risk-pure .medal-ring {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid #10b981;
}
.risk-low .medal-ring {
  background: rgba(56, 189, 248, 0.2);
  color: #38bdf8;
  border: 1px solid #38bdf8;
}
.risk-mid .medal-ring {
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
  border: 1px solid #f59e0b;
}
.risk-high .medal-ring {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  border: 1px solid #ef4444;
}
.risk-text {
  font-size: 10px;
  font-weight: 600;
  color: #f8fafc;
}

/* Card Identity (Raised Embossed Text) */
.card-identity-block {
  margin: 4px 0;
}
.node-embossed-name {
  font-family: 'Outfit', sans-serif;
  font-size: 19px;
  font-weight: 800;
  color: #f8fafc;
  letter-spacing: 0.5px;
  text-shadow:
    0 1px 2px rgba(0, 0, 0, 0.8),
    0 -1px 0 rgba(255, 255, 255, 0.2);
}
.node-embossed-ip {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}
.ip-digits {
  font-family: 'Fira Code', monospace;
  font-size: 13.5px;
  font-weight: 600;
  color: #cbd5e1;
  letter-spacing: 1.2px;
  text-shadow: 0 1px 1px #000;
}
.ip-family-badge {
  font-size: 9px;
  font-weight: 700;
  background: rgba(255, 255, 255, 0.08);
  color: #94a3b8;
  padding: 1px 5px;
  border-radius: 3px;
  border: 1px solid rgba(255, 255, 255, 0.12);
}

/* Network Row */
.card-network-row {
  display: grid;
  grid-template-columns: 1fr 1.6fr 1fr;
  gap: 8px;
  padding: 8px 10px;
  background: rgba(0, 0, 0, 0.25);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.04);
}
.meta-item {
  display: flex;
  flex-direction: column;
}
.meta-item.right {
  align-items: flex-end;
}
.meta-label {
  font-size: 8px;
  font-weight: 700;
  color: #64748b;
  letter-spacing: 0.5px;
}
.meta-val {
  font-family: 'Fira Code', monospace;
  font-size: 11px;
  font-weight: 600;
  color: #f1f5f9;
  margin-top: 1px;
}
.org-val {
  font-family: inherit;
  font-size: 11px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.status-val {
  display: flex;
  align-items: center;
  gap: 4px;
}
.status-val.online { color: #10b981; }
.status-val.warning { color: #f59e0b; }
.status-val.alert { color: #ef4444; }

.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 6px currentColor;
}

/* Unlocks Footer */
.card-unlocks-footer {
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding-top: 10px;
}
.unlocks-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 8.5px;
  font-weight: 700;
  color: #64748b;
  letter-spacing: 0.8px;
  margin-bottom: 8px;
}
.action-hint {
  display: flex;
  align-items: center;
  gap: 2px;
  color: #38bdf8;
  opacity: 0;
  transition: opacity 0.2s ease;
}
.credit-card-wrap:hover .action-hint {
  opacity: 1;
}

.badges-scroll-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
</style>
