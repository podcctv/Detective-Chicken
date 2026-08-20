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

// Hover state (no 3D tilt, just subtle lift)
const cardRef = ref<HTMLElement | null>(null)
const isHovering = ref(false)

const onMouseEnter = () => {
  isHovering.value = true
}

const onMouseLeave = () => {
  isHovering.value = false
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
const flagImgError = ref(false)

// Normalize country code for FlagCDN and Flag Emoji
const normalizedCountryCode = computed(() => {
  const code = (props.node.country_code || '').trim().toLowerCase()
  if (code.length === 2 && /^[a-z]{2}$/.test(code)) {
    if (code === 'uk') return 'gb'
    return code
  }
  const reg = (props.node.region || '').toLowerCase()
  const map: Record<string, string> = {
    '香港': 'hk', 'hong kong': 'hk', 'hongkong': 'hk', 'hk': 'hk',
    '美国': 'us', 'united states': 'us', 'usa': 'us', 'us': 'us',
    '日本': 'jp', 'japan': 'jp', 'jp': 'jp',
    '新加坡': 'sg', 'singapore': 'sg', 'sg': 'sg',
    '台湾': 'tw', 'taiwan': 'tw', 'tw': 'tw',
    '德国': 'de', 'germany': 'de', 'de': 'de',
    '英国': 'gb', 'united kingdom': 'gb', 'uk': 'gb', 'gb': 'gb',
    '法国': 'fr', 'france': 'fr', 'fr': 'fr',
    '韩国': 'kr', 'korea': 'kr', 'kr': 'kr',
    '加拿大': 'ca', 'canada': 'ca', 'ca': 'ca',
    '荷兰': 'nl', 'netherlands': 'nl', 'nl': 'nl',
    '澳大利亚': 'au', 'australia': 'au', 'au': 'au',
    '俄罗斯': 'ru', 'russia': 'ru', 'ru': 'ru',
  }
  for (const [k, v] of Object.entries(map)) {
    if (reg.includes(k)) return v
  }
  return ''
})

const countryFlagEmoji = computed(() => {
  const code = normalizedCountryCode.value
  if (!code || code.length !== 2) return '🌐'
  const offset = 127397
  return String.fromCodePoint(...[...code.toUpperCase()].map((c) => c.charCodeAt(0) + offset))
})
</script>

<template>
  <div
    ref="cardRef"
    class="credit-card-wrap"
    :class="{ 'is-selected': selected }"
    @mouseenter="onMouseEnter"
    @mouseleave="onMouseLeave"
    @click="emit('select', node)"
  >
    <div
      class="metal-credit-card"
      :class="{ hovering: isHovering }"
    >
      <!-- Metallic Foil, Brushed Shading -->
      <div class="card-metal-sheen"></div>
      <!-- Semi-transparent real national flag background covering right half -->
      <div class="card-flag-container" aria-hidden="true">
        <img
          v-if="normalizedCountryCode && !flagImgError"
          :src="`https://flagcdn.com/w640/${normalizedCountryCode}.png`"
          class="card-flag-img"
          alt=""
          loading="lazy"
          @error="flagImgError = true"
        />
        <span v-else class="card-flag-emoji">{{ countryFlagEmoji }}</span>
      </div>
      <div class="card-border-glow"></div>

      <!-- Top Header Row: Country Badge + Provider + Region in ONE line, Risk Medal on Right -->
      <div class="card-header">
        <div class="card-issuer-row">
          <span class="region-flag-pill">{{ node.country_code || node.region || 'GL' }}</span>
          <strong class="issuer-title">{{ node.provider || 'VPS 节点资产' }}</strong>
          <span v-if="node.region && node.region !== node.country_code" class="issuer-region-text">{{ node.region }}</span>
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
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.3s ease, border-color 0.3s ease;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.metal-credit-card.hovering {
  transform: translateY(-4px) scale(1.01);
  box-shadow:
    0 24px 48px rgba(0, 0, 0, 0.7),
    0 8px 16px rgba(0, 0, 0, 0.5),
    inset 0 1px 2px rgba(255, 255, 255, 0.3),
    inset 0 -1px 2px rgba(0, 0, 0, 0.8);
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

/* Semi-transparent country flag background covering right half */
.card-flag-container {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 52%;
  overflow: hidden;
  pointer-events: none;
  border-top-right-radius: 16px;
  border-bottom-right-radius: 16px;
  mask-image: radial-gradient(circle at 82% 50%, rgba(0, 0, 0, 0.85) 0%, rgba(0, 0, 0, 0.4) 50%, transparent 95%);
  -webkit-mask-image: radial-gradient(circle at 82% 50%, rgba(0, 0, 0, 0.85) 0%, rgba(0, 0, 0, 0.4) 50%, transparent 95%);
  opacity: 0.18;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.card-flag-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  filter: saturate(1.3) contrast(1.15);
  transform: scale(1.08);
}

.card-flag-emoji {
  font-size: 130px;
  line-height: 1;
  transform: translateX(15px);
  filter: saturate(1.3);
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

.card-issuer-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
}

.region-flag-pill {
  font-family: 'Fira Code', monospace;
  font-size: 10.5px;
  font-weight: 800;
  padding: 1px 6px;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  border: 1px solid rgba(56, 189, 248, 0.35);
  border-radius: 4px;
  flex-shrink: 0;
}

.issuer-title {
  font-size: 13px;
  font-weight: 800;
  color: #f1f5f9;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.issuer-region-text {
  font-size: 11.5px;
  color: #94a3b8;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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

/* Card Identity (Raised Embossed Text with premium typography) */
.card-identity-block {
  margin: 4px 0;
}
.node-embossed-name {
  font-family: 'Outfit', sans-serif;
  font-size: 21px;
  font-weight: 800;
  color: #f8fafc;
  letter-spacing: 1px;
  text-transform: uppercase;
  text-shadow:
    0 1px 3px rgba(0, 0, 0, 0.9),
    0 -1px 0 rgba(255, 255, 255, 0.15);
}
.node-embossed-ip {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}
.ip-digits {
  font-family: 'Fira Code', monospace;
  font-size: 14px;
  font-weight: 700;
  color: #e2e8f0;
  letter-spacing: 2px;
  text-shadow: 0 1px 2px #000;
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
  overflow: hidden;
}
</style>
