<script setup lang="ts">
import { computed, ref } from 'vue'
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

const isHovering = ref(false)
const flagImgError = ref(false)
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
    <div class="metal-credit-card" :class="{ hovering: isHovering }">
      <div class="card-flag-container" aria-hidden="true">
        <img
          v-if="normalizedCountryCode && !flagImgError"
          :src="`https://flagcdn.com/w640/${normalizedCountryCode}.png`"
          class="card-flag-img"
          alt=""
          width="640"
          height="426"
          loading="lazy"
          @error="flagImgError = true"
        />
      </div>
      <div class="card-metal-sheen" aria-hidden="true"></div>
      <div class="card-border-glow" aria-hidden="true"></div>

      <header class="card-header">
        <div class="card-issuer-row">
          <span class="region-usage-pill">{{ locationUsageBadge }}</span>
          <strong class="issuer-title">{{ node.provider || 'VPS NODE' }}</strong>
          <span v-if="node.region && node.region !== node.country_code" class="issuer-region-text">{{ node.region }}</span>
        </div>
        <div class="risk-medal" :class="riskCategory.class" :aria-label="`风险值 ${riskCategory.score}，${riskCategory.label}`">
          <span class="risk-number">{{ riskCategory.score }}</span><span class="risk-text">{{ riskCategory.label }}</span>
        </div>
      </header>

      <section class="card-identity-block">
        <h3 class="node-name">{{ node.name }}</h3>
        <div class="ip-stack" :class="{ 'is-dual': ipRows.length > 1 }">
          <div v-for="row in ipRows" :key="row.family" class="ip-row">
            <span class="ip-version">{{ row.family ? `IPv${row.family}` : 'IP' }}</span>
            <code class="ip-digits">{{ row.address }}</code>
            <span v-if="row.warp" class="warp-badge">WARP</span>
          </div>
        </div>
      </section>

      <section class="card-network-row" aria-label="网络归属信息">
        <div class="meta-item">
          <span class="meta-label">AUTONOMOUS SYSTEM</span><span class="meta-val">{{ asnDisplay }}</span>
        </div>
        <div class="meta-item organization-item">
          <span class="meta-label">ORGANIZATION</span><span class="meta-val org-val" :title="organizationDisplay">{{ organizationDisplay }}</span>
        </div>
        <div class="meta-item status-item">
          <span class="meta-label">STATUS</span>
          <span class="meta-val status-val" :class="node.status"><span class="live-dot" aria-hidden="true"></span>{{ statusDisplay }}</span>
        </div>
      </section>

      <footer class="card-unlocks-footer">
        <span class="unlocks-title">SERVICE AVAILABILITY</span>
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
.credit-card-wrap { position: relative; min-width: 0; border-radius: 16px; cursor: pointer; user-select: none; outline: none; }
.credit-card-wrap.is-disabled { cursor: default; }
.credit-card-wrap:focus-visible .metal-credit-card { border-color: rgba(125, 211, 252, .72); box-shadow: 0 0 0 3px rgba(56, 189, 248, .24), 0 18px 42px rgba(0, 0, 0, .54); }
.metal-credit-card {
  position: relative; isolation: isolate; min-height: 292px; height: 100%; padding: 17px 19px 16px; overflow: hidden;
  display: flex; flex-direction: column; gap: 13px; border: 1px solid rgba(226, 232, 240, .15); border-radius: 16px;
  background: linear-gradient(145deg, #1a2026 0%, #12171c 38%, #0c1014 70%, #11161b 100%);
  box-shadow: 0 14px 34px rgba(0, 0, 0, .52), inset 0 1px 0 rgba(255, 255, 255, .16), inset 0 -1px 0 rgba(0, 0, 0, .78);
  transition: transform 220ms ease, box-shadow 220ms ease, border-color 220ms ease;
}
.metal-credit-card.hovering { transform: translateY(-2px); border-color: rgba(226, 232, 240, .22); box-shadow: 0 19px 42px rgba(0, 0, 0, .58), inset 0 1px 0 rgba(255, 255, 255, .18); }
.is-selected .metal-credit-card { border-color: rgba(56, 189, 248, .56); box-shadow: 0 18px 42px rgba(0, 0, 0, .62), inset 0 1px 0 rgba(186, 230, 253, .2); }
.card-metal-sheen { position: absolute; inset: 0; z-index: 0; pointer-events: none; background: repeating-linear-gradient(0deg, rgba(255,255,255,.018) 0, rgba(255,255,255,.018) 1px, transparent 1px, transparent 4px), linear-gradient(115deg, rgba(255,255,255,.055), transparent 24%, transparent 74%, rgba(255,255,255,.025)); }
.card-flag-container { position: absolute; inset: 0 0 0 auto; z-index: -1; width: 42%; opacity: .12; overflow: hidden; pointer-events: none; mask-image: linear-gradient(90deg, transparent 0%, rgba(0,0,0,.34) 24%, rgba(0,0,0,.9) 100%); -webkit-mask-image: linear-gradient(90deg, transparent 0%, rgba(0,0,0,.34) 24%, rgba(0,0,0,.9) 100%); }
.card-flag-img { width: 100%; height: 100%; object-fit: cover; filter: grayscale(35%) saturate(70%) contrast(100%) brightness(1.03); }
.card-border-glow { position: absolute; inset: 0; z-index: 4; border-radius: inherit; pointer-events: none; box-shadow: inset 0 0 0 1px rgba(255,255,255,.035); }
.card-header, .card-identity-block, .card-network-row, .card-unlocks-footer { position: relative; z-index: 2; }
.card-header { min-height: 28px; display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.card-issuer-row { min-width: 0; display: flex; align-items: center; gap: 7px; }
.region-usage-pill { flex: none; padding: 3px 7px; border: 1px solid rgba(125,211,252,.32); border-radius: 4px; background: rgba(56,189,248,.1); color: #9cdef9; font: 650 10px/1.35 'Fira Code', monospace; letter-spacing: 0; }
.issuer-title { min-width: 0; overflow: hidden; color: #e5eaf0; font-size: 12px; font-weight: 650; letter-spacing: .045em; text-overflow: ellipsis; text-transform: uppercase; white-space: nowrap; }
.issuer-region-text { overflow: hidden; color: #7f8a98; font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.risk-medal { flex: none; display: inline-flex; align-items: center; gap: 5px; min-height: 25px; padding: 3px 8px 3px 4px; border: 1px solid rgba(148,163,184,.18); border-radius: 999px; background: rgba(4,7,10,.42); color: #94a3b8; }
.risk-number { min-width: 19px; height: 19px; display: grid; place-items: center; border: 1px solid currentColor; border-radius: 50%; font: 650 10px/1 'Fira Code', monospace; }
.risk-text { font-size: 10px; font-weight: 600; }
.risk-pure { color: #47cfa0; } .risk-low { color: #7dd3fc; } .risk-mid { color: #f5bd57; } .risk-high { color: #f17878; }
.risk-waiting { color: #7c8794; border-style: dashed; }
.card-identity-block { min-height: 78px; }
.node-name { margin: 0 0 7px; overflow: hidden; color: #f3f5f7; font-size: 21px; font-weight: 650; letter-spacing: .01em; line-height: 1.12; text-overflow: ellipsis; white-space: nowrap; }
.ip-stack { width: min(100%, 340px); display: grid; gap: 3px; }
.ip-row { min-width: 0; display: grid; grid-template-columns: 30px minmax(0, 1fr) auto; align-items: start; gap: 7px; }
.ip-version { color: #919daa; font: 550 9px/1.3 'Fira Code', monospace; letter-spacing: 0; text-transform: uppercase; }
.ip-digits { min-width: 0; color: #d5dce4; font: 600 13px/1.4 'Fira Code', monospace; letter-spacing: 0; overflow-wrap: anywhere; white-space: normal; }
.warp-badge { padding: 1px 5px; border: 1px solid rgba(148,163,184,.22); border-radius: 3px; font-size: 9px; font-weight: 650; line-height: 1.45; }
.warp-badge { color: #f2be62; background: rgba(245,158,11,.09); border-color: rgba(245,158,11,.28); }
.card-network-row { min-width: 0; display: grid; grid-template-columns: minmax(76px,.8fr) minmax(0,1.45fr) minmax(58px,auto); gap: 9px; padding: 8px 10px; border: 1px solid rgba(226,232,240,.065); border-radius: 7px; background: rgba(1,4,7,.31); box-shadow: inset 0 1px 5px rgba(0,0,0,.32); }
.meta-item { min-width: 0; display: flex; flex-direction: column; gap: 2px; } .status-item { align-items: flex-end; }
.meta-label { color: #85919f; font-size: 8.5px; font-weight: 650; letter-spacing: 0; }
.meta-val { color: #e1e7ed; font: 600 11px/1.4 'Fira Code', monospace; }
.org-val { overflow: hidden; font-family: inherit; text-overflow: ellipsis; white-space: nowrap; }
.status-val { display: flex; align-items: center; gap: 4px; }
.status-val.online { color: #52cfa3; } .status-val.warning { color: #efb957; } .status-val.alert { color: #ef7777; } .status-val.offline { color: #7b8794; } .status-val.pending { color: #7cbfdd; }
.live-dot { width: 5px; height: 5px; flex: none; border-radius: 50%; background: currentColor; }
.card-unlocks-footer { margin-top: auto; padding-top: 9px; border-top: 1px solid rgba(226,232,240,.07); }
.unlocks-title { display: block; margin-bottom: 8px; color: #87929f; font-size: 9px; font-weight: 650; letter-spacing: 0; }
.service-dock { display: grid; grid-template-columns: repeat(8,30px); align-items: center; gap: 8px; }
.service-dock :deep(.metal-badge) { width: 30px; height: 30px; padding: 0; overflow: visible; display: grid; place-items: center; border: 1px solid rgba(226,232,240,.1); border-radius: 6px; background: rgba(3,6,9,.32); box-shadow: inset 0 1px 0 rgba(255,255,255,.03); opacity: 1; transform: none; }
.service-dock :deep(.metal-badge:hover) { border-color: rgba(226,232,240,.2); box-shadow: none; transform: none; }
.service-dock :deep(.metal-foil), .service-dock :deep(.metal-light-sweep) { display: none; }
.service-dock :deep(.brand-logo-wrap) { width: 22px; height: 22px; border: 0; background: transparent; box-shadow: none; }
.service-dock :deep(.brand-svg) { width: 19px; height: 19px; }
.service-dock :deep(.brand-generic-char) { font-size: 8px; }
.service-dock :deep(.compact-cn-tag) { right: -5px; bottom: -4px; box-shadow: none; }
@media (max-width: 480px) {
  .metal-credit-card { min-height: 302px; padding: 16px; }
  .node-name { font-size: 19px; }
  .issuer-region-text, .risk-text { display: none; }
  .risk-medal { padding-right: 6px; }
  .card-network-row { grid-template-columns: minmax(66px,.8fr) minmax(0,1.25fr) auto; gap: 7px; padding-inline: 8px; }
  .meta-label { font-size: 8px; }
  .service-dock { gap: 6px; }
}
@media (prefers-reduced-motion: reduce) { .metal-credit-card { transition: none; } .metal-credit-card.hovering { transform: none; } }
</style>
