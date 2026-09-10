<script setup lang="ts">
import { computed, ref } from 'vue'
import { Camera, Check, Loader2, Sparkles, Cpu } from '@lucide/vue'
import { toBlob } from 'html-to-image'
import type { Node, UnlockInfo } from '../types'
import MetalBadge from './MetalBadge.vue'

const props = withDefaults(defineProps<{ node: Node; selected?: boolean; interactive?: boolean }>(), {
  selected: false,
  interactive: true,
})
const emit = defineEmits<{
  (e: 'select', node: Node): void
  (e: 'inspectService', payload: { node: Node; serviceId: string }): void
}>()

const cardRef = ref<HTMLElement | null>(null)
const isHovering = ref(false)
const flagImgError = ref(false)
const copyStatus = ref<'idle' | 'copying' | 'copied' | 'error'>('idle')

const copyStatusText = computed(() => {
  if (copyStatus.value === 'copying') return '生成图片中...'
  if (copyStatus.value === 'copied') return '已复制到剪贴板'
  if (copyStatus.value === 'error') return '复制失败'
  return '复制分析'
})

const hasTelemetry = computed(() => Boolean(
  props.node.masked_ipv4 || props.node.masked_ipv6 || props.node.masked_ip || props.node.asn > 0,
))
const hasQualityResult = computed(() => {
  if (['pending', 'scanning'].includes(props.node.quality_status || '')) return false
  const scannedAt = new Date(props.node.last_scan || '')
  return ['ready', 'partial'].includes(props.node.quality_status || '') ||
    (Number.isFinite(scannedAt.getTime()) && scannedAt.getFullYear() > 1) ||
    props.node.asn > 0 || Boolean(props.node.organization)
})
const riskCategory = computed(() => {
  if (!hasTelemetry.value || !hasQualityResult.value) {
    return { label: 'WAITING', class: 'risk-waiting', score: '—' }
  }
  if (props.node.risk <= 20) return { label: '极净优质', class: 'risk-pure', score: String(props.node.risk) }
  if (props.node.risk <= 50) return { label: '低度风险', class: 'risk-low', score: String(props.node.risk) }
  if (props.node.risk <= 75) return { label: '中度注意', class: 'risk-mid', score: String(props.node.risk) }
  return { label: '高危关注', class: 'risk-high', score: String(props.node.risk) }
})
const keyServices = computed(() => {
  const ids = ['chatgpt', 'claude', 'deepseek', 'netflix', 'disney', 'youtube', 'prime', 'spotify']
  return ids.map((id) => {
    const found: UnlockInfo | undefined = props.node.unlocks?.streaming?.[id] || props.node.unlocks?.ai?.[id]
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
const normalizedCountryCode = computed(() => {
  const code = (props.node.country_code || '').trim().toLowerCase()
  if (/^[a-z]{2}$/.test(code)) return code === 'uk' ? 'gb' : code
  const region = (props.node.region || '').toLowerCase()
  const map: Record<string, string> = {
    香港: 'hk', 'hong kong': 'hk', 美国: 'us', 'united states': 'us', 日本: 'jp', japan: 'jp',
    新加坡: 'sg', singapore: 'sg', 台湾: 'tw', taiwan: 'tw', 德国: 'de', germany: 'de',
    英国: 'gb', 'united kingdom': 'gb', 法国: 'fr', france: 'fr', 韩国: 'kr', korea: 'kr',
    加拿大: 'ca', canada: 'ca', 荷兰: 'nl', netherlands: 'nl', 澳大利亚: 'au', australia: 'au',
  }
  return Object.entries(map).find(([key]) => region.includes(key))?.[1] || ''
})
const usageLabel = computed(() => props.node.usage_type || 'WAITING')
const locationUsageBadge = computed(() => {
  const labels = [props.node.country_code || props.node.region || '—', usageLabel.value]
  if (props.node.ip_type) labels.push(props.node.ip_type)
  return labels.join(' · ')
})
const families = computed(() => {
  const values = props.node.families?.length ? props.node.families : [props.node.family]
  return values.filter((family, index): family is number =>
    (family === 4 || family === 6) && values.indexOf(family) === index,
  )
})
const ipRows = computed(() => {
  const rows: Array<{ family: number; address: string; warp: boolean }> = []
  if (families.value.includes(4)) rows.push({
    family: 4,
    address: props.node.masked_ipv4 || (families.value.length === 1 ? props.node.masked_ip : '') || '—',
    warp: Boolean(props.node.warp4 || (props.node.is_warp && !props.node.warp6)),
  })
  if (families.value.includes(6)) rows.push({
    family: 6,
    address: props.node.masked_ipv6 || (families.value.length === 1 ? props.node.masked_ip : '') || '—',
    warp: Boolean(props.node.warp6 || (props.node.is_warp && !props.node.warp4 && !families.value.includes(4))),
  })
  return rows.length ? rows : [{ family: 0, address: 'WAITING', warp: false }]
})
const asnDisplay = computed(() => props.node.asn > 0 ? `AS${props.node.asn}` : '—')
const organizationDisplay = computed(() => props.node.organization || '—')
const statusDisplay = computed(() => ({
  online: '在线', warning: '注意', alert: '告警', offline: '离线', pending: '待接入',
}[props.node.status] || '未知'))

const selectCard = () => props.interactive && emit('select', props.node)

// Copy card face as image directly to clipboard
const copyCardImage = async () => {
  if (!cardRef.value || copyStatus.value === 'copying') return
  copyStatus.value = 'copying'

  try {
    const blob = await toBlob(cardRef.value, {
      pixelRatio: 2, // Retina HD
      cacheBust: true,
      backgroundColor: '#0d1217',
      filter: (node) => {
        if (node instanceof HTMLElement && node.classList.contains('no-export')) {
          return false
        }
        return true
      },
    })

    if (!blob) throw new Error('Blob generation failed')

    // Copy directly into system clipboard
    if (navigator.clipboard && window.ClipboardItem) {
      await navigator.clipboard.write([
        new ClipboardItem({ 'image/png': blob }),
      ])
      copyStatus.value = 'copied'
    } else {
      // Fallback: Download file
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${props.node.name || 'node'}-card.png`
      a.click()
      URL.revokeObjectURL(url)
      copyStatus.value = 'copied'
    }
  } catch (err) {
    console.warn('Clipboard write failed, triggering fallback download', err)
    try {
      const blob = await toBlob(cardRef.value, {
        pixelRatio: 2,
        cacheBust: true,
        backgroundColor: '#0d1217',
        filter: (node) => !(node instanceof HTMLElement && node.classList.contains('no-export')),
      })
      if (blob) {
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `${props.node.name || 'node'}-card.png`
        a.click()
        URL.revokeObjectURL(url)
        copyStatus.value = 'copied'
      } else {
        copyStatus.value = 'error'
      }
    } catch {
      copyStatus.value = 'error'
    }
  } finally {
    setTimeout(() => {
      copyStatus.value = 'idle'
    }, 2400)
  }
}
</script>

<template>
  <article
    class="credit-card-wrap"
    :class="{ 'is-selected': selected, 'is-disabled': !interactive }"
    :tabindex="interactive ? 0 : -1"
    role="button"
    :aria-label="`查看 ${node.name} 深度检测报告`"
    :aria-pressed="selected"
    @mouseenter="isHovering = true"
    @mouseleave="isHovering = false"
    @click="selectCard"
    @keydown.enter.prevent="selectCard"
    @keydown.space.prevent="selectCard"
  >
    <div ref="cardRef" class="metal-credit-card" :class="{ hovering: isHovering }">
      <!-- HUD Tech Corners -->
      <div class="hud-corner top-left" aria-hidden="true"></div>
      <div class="hud-corner top-right" aria-hidden="true"></div>
      <div class="hud-corner bottom-left" aria-hidden="true"></div>
      <div class="hud-corner bottom-right" aria-hidden="true"></div>

      <!-- Sci-fi Laser Grid & Holographic Sheen -->
      <div class="card-circuit-grid" aria-hidden="true"></div>
      <div class="card-metal-sheen" aria-hidden="true"></div>
      <div class="card-border-glow" aria-hidden="true"></div>

      <!-- Watermarked Flag -->
      <div class="card-flag-container" aria-hidden="true">
        <img
          v-if="normalizedCountryCode && !flagImgError"
          :src="`https://flagcdn.com/w640/${normalizedCountryCode}.png`"
          crossorigin="anonymous"
          class="card-flag-img"
          alt=""
          width="640"
          height="426"
          loading="lazy"
          @error="flagImgError = true"
        />
      </div>

      <!-- Hologram Cyber Bar -->
      <div class="holo-security-bar" aria-hidden="true">
        <span class="holo-text">DETECTIVE TITANIUM · CHIP VERIFIED</span>
      </div>

      <!-- Card Header -->
      <header class="card-header">
        <div class="card-issuer-row">
          <span class="region-usage-pill">{{ locationUsageBadge }}</span>
          <strong class="issuer-title">{{ node.provider || 'VPS NODE' }}</strong>
          <span v-if="node.region && node.region !== node.country_code" class="issuer-region-text">{{ node.region }}</span>
        </div>

        <div class="header-right-deck">
          <!-- Copy Analysis Button (filtered out during export via .no-export) -->
          <button
            class="copy-analysis-btn no-export"
            :class="{ 'is-copying': copyStatus === 'copying', 'is-copied': copyStatus === 'copied' }"
            :disabled="copyStatus === 'copying'"
            :title="copyStatus === 'copied' ? '已复制，可直接粘贴到微信/AI等' : '一键复制卡片正面高清图片'"
            @click.stop="copyCardImage"
          >
            <Loader2 v-if="copyStatus === 'copying'" :size="11" class="spin-icon" />
            <Check v-else-if="copyStatus === 'copied'" :size="11" class="check-icon" />
            <Camera v-else :size="11" class="camera-icon" />
            <span>{{ copyStatusText }}</span>
          </button>

          <!-- Risk / Quality HUD Badge -->
          <div class="risk-medal" :class="riskCategory.class" :aria-label="`风险值 ${riskCategory.score}，${riskCategory.label}`">
            <span class="risk-number">{{ riskCategory.score }}</span>
            <span class="risk-text">{{ riskCategory.label }}</span>
          </div>
        </div>
      </header>

      <!-- Identity & Cyber EMV Chip Section -->
      <section class="card-identity-block">
        <div class="identity-top-row">
          <h3 class="node-name">{{ node.name }}</h3>

          <!-- Realistic Cyber EMV Gold Smartcard Chip -->
          <div class="emv-chip-container" aria-hidden="true" title="Titanium Contact Smartchip">
            <div class="emv-gold-chip">
              <div class="emv-contact-cut"></div>
              <div class="emv-center-plate"></div>
            </div>
            <div class="contactless-waves">
              <span class="wave w1"></span>
              <span class="wave w2"></span>
              <span class="wave w3"></span>
            </div>
          </div>
        </div>

        <div class="ip-stack" :class="{ 'is-dual': ipRows.length > 1 }">
          <div v-for="row in ipRows" :key="row.family" class="ip-row">
            <span class="ip-version">{{ row.family ? `IPv${row.family}` : 'IP' }}</span>
            <code class="ip-digits">{{ row.address }}</code>
            <span v-if="row.warp" class="warp-badge">WARP</span>
          </div>
        </div>
      </section>

      <!-- Network Subsystems HUD Row -->
      <section class="card-network-row" aria-label="网络归属信息">
        <div class="meta-item">
          <span class="meta-label">AUTONOMOUS SYSTEM</span>
          <span class="meta-val">{{ asnDisplay }}</span>
        </div>
        <div class="meta-item organization-item">
          <span class="meta-label">ORGANIZATION</span>
          <span class="meta-val org-val" :title="organizationDisplay">{{ organizationDisplay }}</span>
        </div>
        <div class="meta-item status-item">
          <span class="meta-label">STATUS</span>
          <span class="meta-val status-val" :class="node.status">
            <span class="live-dot" aria-hidden="true"></span>
            {{ statusDisplay }}
          </span>
        </div>
      </section>

      <!-- Service Availability Dock -->
      <footer class="card-unlocks-footer">
        <div class="unlocks-head">
          <span class="unlocks-title">SERVICE AVAILABILITY</span>
          <span class="tech-spec-label">8-WAY PROBE MATRIX</span>
        </div>
        <div class="service-dock" aria-label="核心服务检测状态">
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
      </footer>
    </div>
  </article>
</template>

<style scoped>
.credit-card-wrap {
  position: relative;
  min-width: 0;
  border-radius: 16px;
  cursor: pointer;
  user-select: none;
  outline: none;
  transition: transform 0.16s ease;
}
.credit-card-wrap:active:not(.is-disabled) {
  transform: scale(0.985);
}
.credit-card-wrap.is-disabled {
  cursor: default;
}
.credit-card-wrap:focus-visible .metal-credit-card {
  border-color: rgba(125, 211, 252, .8);
  box-shadow: 0 0 0 3px rgba(56, 189, 248, .28), 0 18px 42px rgba(0, 0, 0, .54);
}

.metal-credit-card {
  position: relative;
  isolation: isolate;
  min-height: 298px;
  height: 100%;
  padding: 18px 20px 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 12px;
  border: 1px solid rgba(125, 211, 252, 0.18);
  border-radius: 16px;
  background:
    radial-gradient(ellipse at 85% 0%, rgba(56, 189, 248, 0.12), transparent 45%),
    radial-gradient(ellipse at 15% 100%, rgba(139, 92, 246, 0.08), transparent 50%),
    linear-gradient(135deg, #182029 0%, #11171f 38%, #0a0e13 72%, #121922 100%);
  box-shadow:
    0 16px 36px rgba(0, 0, 0, 0.58),
    inset 0 1px 0 rgba(255, 255, 255, 0.2),
    inset 0 -1px 0 rgba(0, 0, 0, 0.85);
  transition: transform 220ms ease, box-shadow 220ms ease, border-color 220ms ease;
}

.metal-credit-card.hovering {
  transform: translateY(-2px);
  border-color: rgba(56, 189, 248, 0.45);
  box-shadow:
    0 22px 46px rgba(0, 0, 0, 0.65),
    0 0 24px rgba(56, 189, 248, 0.16),
    inset 0 1px 0 rgba(186, 230, 253, 0.3);
}
.is-selected .metal-credit-card {
  border-color: rgba(56, 189, 248, 0.6);
  box-shadow:
    0 20px 44px rgba(0, 0, 0, 0.65),
    0 0 28px rgba(56, 189, 248, 0.24),
    inset 0 1px 0 rgba(186, 230, 253, 0.35);
}

/* Sci-fi HUD Corner Brackets */
.hud-corner {
  position: absolute;
  width: 9px;
  height: 9px;
  pointer-events: none;
  z-index: 5;
  border-color: rgba(56, 189, 248, 0.55);
  border-style: solid;
}
.hud-corner.top-left { top: 6px; left: 6px; border-width: 2px 0 0 2px; }
.hud-corner.top-right { top: 6px; right: 6px; border-width: 2px 2px 0 0; }
.hud-corner.bottom-left { bottom: 6px; left: 6px; border-width: 0 0 2px 2px; }
.hud-corner.bottom-right { bottom: 6px; right: 6px; border-width: 0 2px 2px 0; }

/* Micro-Circuit Dotted Laser Grid */
.card-circuit-grid {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
  opacity: 0.18;
  background-image:
    radial-gradient(circle, rgba(125, 211, 252, 0.3) 1px, transparent 1px);
  background-size: 14px 14px;
}

/* Hologram Cyber Security Strip */
.holo-security-bar {
  position: absolute;
  top: 10px;
  right: -24px;
  transform: rotate(24deg);
  background: linear-gradient(90deg, rgba(244, 114, 182, 0.12), rgba(56, 189, 248, 0.2), rgba(250, 204, 21, 0.12));
  border-block: 1px solid rgba(255, 255, 255, 0.12);
  padding: 1px 30px;
  pointer-events: none;
  z-index: 2;
  opacity: 0.45;
}
.holo-text {
  font: 700 6.5px 'Fira Code', monospace;
  color: #e2e8f0;
  letter-spacing: 1.2px;
  text-transform: uppercase;
}

.card-metal-sheen {
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
  background:
    repeating-linear-gradient(0deg, rgba(255,255,255,.018) 0, rgba(255,255,255,.018) 1px, transparent 1px, transparent 4px),
    linear-gradient(115deg, rgba(255,255,255,.07), transparent 28%, transparent 74%, rgba(255,255,255,.03));
}

.card-flag-container {
  position: absolute;
  inset: 0 0 0 auto;
  z-index: 0;
  width: 44%;
  opacity: 0.13;
  overflow: hidden;
  pointer-events: none;
  mask-image: linear-gradient(90deg, transparent 0%, rgba(0,0,0,.34) 22%, rgba(0,0,0,.95) 100%);
  -webkit-mask-image: linear-gradient(90deg, transparent 0%, rgba(0,0,0,.34) 22%, rgba(0,0,0,.95) 100%);
}
.card-flag-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  filter: grayscale(25%) saturate(85%) contrast(105%) brightness(1.05);
}

.card-border-glow {
  position: absolute;
  inset: 0;
  z-index: 4;
  border-radius: inherit;
  pointer-events: none;
  box-shadow: inset 0 0 0 1px rgba(255,255,255,.04);
}

.card-header, .card-identity-block, .card-network-row, .card-unlocks-footer {
  position: relative;
  z-index: 3;
}

/* Card Header */
.card-header {
  min-height: 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.card-issuer-row {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 7px;
  flex: 1;
}
.region-usage-pill {
  flex: none;
  padding: 3px 8px;
  border: 1px solid rgba(125, 211, 252, .38);
  border-radius: 4px;
  background: rgba(56, 189, 248, .12);
  color: #7dd3fc;
  font: 700 10px/1.3 'Fira Code', monospace;
  letter-spacing: 0.2px;
  box-shadow: 0 0 10px rgba(56, 189, 248, 0.15);
}
.issuer-title {
  min-width: 0;
  overflow: hidden;
  color: #f1f5f9;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: .05em;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}
.issuer-region-text {
  overflow: hidden;
  color: #8da2b5;
  font-size: 10px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-right-deck {
  flex: none;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Copy Analysis Image Button */
.copy-analysis-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 9px;
  border-radius: 6px;
  background: linear-gradient(135deg, rgba(56, 189, 248, 0.15), rgba(14, 165, 233, 0.25));
  border: 1px solid rgba(125, 211, 252, 0.4);
  color: #bae6fd;
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  font-weight: 650;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
  transition: all 0.18s cubic-bezier(0.16, 1, 0.3, 1);
}
.copy-analysis-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, rgba(56, 189, 248, 0.3), rgba(14, 165, 233, 0.45));
  border-color: rgba(125, 211, 252, 0.7);
  color: #fff;
  box-shadow: 0 0 12px rgba(56, 189, 248, 0.4);
  transform: translateY(-1px);
}
.copy-analysis-btn:active:not(:disabled) {
  transform: scale(0.96);
}
.copy-analysis-btn.is-copying {
  opacity: 0.8;
  cursor: wait;
  color: #7dd3fc;
}
.copy-analysis-btn.is-copied {
  background: rgba(16, 185, 129, 0.2);
  border-color: rgba(16, 185, 129, 0.6);
  color: #34d399;
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.35);
}
.spin-icon {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Risk / Quality Medal */
.risk-medal {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-height: 24px;
  padding: 2px 8px 2px 4px;
  border: 1px solid rgba(148,163,184,.22);
  border-radius: 999px;
  background: rgba(3, 7, 12, 0.6);
  color: #94a3b8;
  backdrop-filter: blur(4px);
}
.risk-number {
  min-width: 18px;
  height: 18px;
  display: grid;
  place-items: center;
  border: 1px solid currentColor;
  border-radius: 50%;
  font: 700 9.5px/1 'Fira Code', monospace;
  font-variant-numeric: tabular-nums;
}
.risk-text {
  font-size: 10px;
  font-weight: 650;
}
.risk-pure { color: #34d399; border-color: rgba(52, 211, 153, 0.4); box-shadow: 0 0 8px rgba(52, 211, 153, 0.15); }
.risk-low { color: #38bdf8; border-color: rgba(56, 189, 248, 0.4); }
.risk-mid { color: #fbbf24; border-color: rgba(251, 191, 36, 0.4); }
.risk-high { color: #f87171; border-color: rgba(248, 113, 113, 0.4); }
.risk-waiting { color: #7c8794; border-style: dashed; }

/* Identity & Smartchip Block */
.card-identity-block {
  min-height: 76px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.identity-top-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.node-name {
  margin: 0;
  overflow: hidden;
  color: #f8fafc;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: .02em;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Realistic Cyber EMV Smartchip */
.emv-chip-container {
  display: flex;
  align-items: center;
  gap: 5px;
  flex: none;
}
.emv-gold-chip {
  position: relative;
  width: 32px;
  height: 23px;
  border-radius: 4px;
  background: linear-gradient(135deg, #fcd34d 0%, #b45309 48%, #fef3c7 76%, #d97706 100%);
  box-shadow: inset 0 1px 1px rgba(255, 255, 255, 0.7), 0 2px 6px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  border: 1px solid rgba(180, 83, 9, 0.6);
}
.emv-contact-cut {
  position: absolute;
  inset: 3px;
  border: 1px solid rgba(120, 53, 15, 0.5);
  border-radius: 2px;
}
.emv-contact-cut::before,
.emv-contact-cut::after {
  content: '';
  position: absolute;
  background: rgba(120, 53, 15, 0.5);
}
.emv-contact-cut::before {
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
}
.emv-contact-cut::after {
  left: 50%;
  top: 0;
  bottom: 0;
  width: 1px;
}
.emv-center-plate {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 12px;
  height: 9px;
  border-radius: 2px;
  background: rgba(254, 243, 199, 0.8);
  border: 1px solid rgba(120, 53, 15, 0.6);
}

/* Contactless Wave icon */
.contactless-waves {
  display: flex;
  flex-direction: column;
  gap: 2px;
  opacity: 0.55;
}
.wave {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: #7dd3fc;
}
.wave.w2 { width: 5px; height: 5px; }
.wave.w3 { width: 7px; height: 7px; }

/* IP Address Rows */
.ip-stack {
  width: min(100%, 360px);
  display: grid;
  gap: 4px;
}
.ip-row {
  min-width: 0;
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
}
.ip-version {
  color: #94a3b8;
  font: 700 9px/1.3 'Fira Code', monospace;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}
.ip-digits {
  min-width: 0;
  color: #e2e8f0;
  font: 600 12.5px/1.35 'Fira Code', monospace;
  letter-spacing: 0;
  overflow-wrap: anywhere;
  white-space: normal;
  font-variant-numeric: tabular-nums;
  text-shadow: 0 0 10px rgba(56, 189, 248, 0.15);
}
.warp-badge {
  padding: 1px 5px;
  border: 1px solid rgba(245, 158, 11, 0.4);
  border-radius: 3px;
  font-size: 8.5px;
  font-weight: 700;
  line-height: 1.4;
  color: #fbbf24;
  background: rgba(245, 158, 11, 0.12);
}

/* Network Subsystems HUD Box */
.card-network-row {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(76px, .8fr) minmax(0, 1.45fr) minmax(58px, auto);
  gap: 9px;
  padding: 7px 10px;
  border: 1px solid rgba(125, 211, 252, 0.12);
  border-radius: 8px;
  background: rgba(4, 8, 14, 0.45);
  box-shadow: inset 0 1px 6px rgba(0, 0, 0, 0.45);
}
.meta-item {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.status-item {
  align-items: flex-end;
}
.meta-label {
  color: #7b8fa3;
  font-size: 8.5px;
  font-weight: 700;
  letter-spacing: 0.3px;
}
.meta-val {
  color: #e1e7ed;
  font: 600 11px/1.4 'Fira Code', monospace;
  font-variant-numeric: tabular-nums;
}
.org-val {
  overflow: hidden;
  font-family: inherit;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.status-val {
  display: flex;
  align-items: center;
  gap: 4px;
}
.status-val.online { color: #34d399; }
.status-val.warning { color: #fbbf24; }
.status-val.alert { color: #f87171; }
.status-val.offline { color: #64748b; }
.status-val.pending { color: #38bdf8; }

.live-dot {
  width: 5px;
  height: 5px;
  flex: none;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 6px currentColor;
}

/* Service Dock */
.card-unlocks-footer {
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid rgba(125, 211, 252, 0.1);
}
.unlocks-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 7px;
}
.unlocks-title {
  color: #7b8fa3;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.4px;
}
.tech-spec-label {
  color: #475569;
  font: 600 8px 'Fira Code', monospace;
  letter-spacing: 0.5px;
}
.service-dock {
  display: grid;
  grid-template-columns: repeat(8, 30px);
  align-items: center;
  gap: 8px;
}
.service-dock :deep(.metal-badge) {
  width: 30px;
  height: 30px;
  padding: 0;
  overflow: visible;
  display: grid;
  place-items: center;
  border: 1px solid rgba(226, 232, 240, 0.12);
  border-radius: 6px;
  background: rgba(3, 7, 12, 0.45);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
  opacity: 1;
  transform: none;
  transition: border-color 0.15s ease, transform 0.15s ease;
}
.service-dock :deep(.metal-badge:hover) {
  border-color: rgba(56, 189, 248, 0.45);
  box-shadow: 0 0 8px rgba(56, 189, 248, 0.25);
  transform: translateY(-1px);
}
.service-dock :deep(.metal-foil), .service-dock :deep(.metal-light-sweep) {
  display: none;
}
.service-dock :deep(.brand-logo-wrap) {
  width: 22px;
  height: 22px;
  border: 0;
  background: transparent;
  box-shadow: none;
}
.service-dock :deep(.brand-svg) {
  width: 19px;
  height: 19px;
}
.service-dock :deep(.brand-generic-char) {
  font-size: 8px;
}
.service-dock :deep(.compact-cn-tag) {
  right: -5px;
  bottom: -4px;
  box-shadow: none;
}

@media (max-width: 480px) {
  .metal-credit-card { min-height: 304px; padding: 16px; }
  .node-name { font-size: 18px; }
  .issuer-region-text, .risk-text { display: none; }
  .risk-medal { padding-right: 5px; }
  .card-network-row { grid-template-columns: minmax(66px,.8fr) minmax(0,1.25fr) auto; gap: 7px; padding-inline: 8px; }
  .meta-label { font-size: 8px; }
  .service-dock { gap: 6px; }
  .copy-analysis-btn span { display: none; }
  .copy-analysis-btn { padding: 4px 6px; }
}

@media (prefers-reduced-motion: reduce) {
  .metal-credit-card { transition: none; }
  .metal-credit-card.hovering { transform: none; }
}
</style>
