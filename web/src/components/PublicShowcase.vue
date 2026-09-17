<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Bot,
  LogIn,
  Moon,
  RefreshCw,
  Server,
  Sparkles,
  Sun,
  Tv,
  X,
} from '@lucide/vue'

import CreditCardNode from './CreditCardNode.vue'
import MetalBadge from './MetalBadge.vue'
import type { Dashboard, Node } from '../types'

const props = defineProps<{
  data: Dashboard
  loading: boolean
  dark: boolean
  refreshing: boolean
}>()

const emit = defineEmits<{ login: []; refresh: []; theme: [] }>()

const selectedNode = ref<Node | null>(null)
const inspectNode = ref<Node | null>(null)
const activeRegionFilter = ref<string>('all')

const qualityClass = (risk: number) =>
  risk >= 60 ? 'risk-high' : risk >= 35 ? 'risk-mid' : 'risk-low'

// Available Regions from nodes
const availableRegions = computed(() => {
  const set = new Set<string>()
  props.data.nodes.forEach((n) => {
    if (n.country_code) set.add(n.country_code)
    else if (n.region) set.add(n.region)
  })
  return Array.from(set)
})

// Filtered nodes based on active region
const filteredNodes = computed(() => {
  if (activeRegionFilter.value === 'all') return props.data.nodes
  return props.data.nodes.filter(
    (n) => n.country_code === activeRegionFilter.value || n.region === activeRegionFilter.value
  )
})

const onSelectNode = (node: Node) => {
  selectedNode.value = node
  inspectNode.value = node
}

const onInspectService = ({ node }: { node: Node; serviceId: string }) => {
  selectedNode.value = node
  inspectNode.value = node
}

const detectedUsage = (node: Node) => node.usage_type || '待检测'
const detectedIPType = (node: Node) => node.ip_type || '待检测'
const detectedASN = (node: Node) => node.asn > 0 ? `AS${node.asn}` : 'ASN 待检测'
const detectedOrganization = (node: Node) => node.organization || '归属组织待检测'
const hasQualityResult = (node: Node) => {
  if (['pending', 'scanning'].includes(node.quality_status || '')) return false
  const scannedAt = new Date(node.last_scan || '')
  return ['ready', 'partial'].includes(node.quality_status || '') ||
    (Number.isFinite(scannedAt.getTime()) && scannedAt.getFullYear() > 1) ||
    node.asn > 0 || Boolean(node.organization)
}
const riskLabel = (node: Node) => {
  if (!hasQualityResult(node)) return '等待检测'
  if (node.risk <= 20) return '极净优质'
  if (node.risk <= 50) return '低风险'
  if (node.risk <= 75) return '中度注意'
  return '高风险'
}
const relative = (input?: string) => {
  if (!input || new Date(input).getFullYear() <= 1) return '尚未上报'
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(input).getTime()) / 1000))
  if (seconds < 60) return `${seconds} 秒前`
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} 小时前`
  return `${Math.floor(seconds / 86400)} 天前`
}
const maskedAddresses = (node: Node) => {
  const addresses = [node.masked_ipv4, node.masked_ipv6].filter(Boolean)
  return addresses.length ? addresses.join(' · ') : (node.masked_ip || 'IP 待检测')
}
const protocolLabel = (node: Node) => {
  const families = (node.families?.length ? node.families : [node.family]).filter(
    (family) => family === 4 || family === 6,
  )
  const labels: string[] = []
  if (families.includes(4)) labels.push(node.warp4 || (node.is_warp && !node.warp6) ? 'IPv4(WARP)' : 'IPv4')
  if (families.includes(6)) labels.push(node.warp6 || (node.is_warp && !node.warp4 && !families.includes(4)) ? 'IPv6(WARP)' : 'IPv6')
  return labels.length ? labels.join(' + ') : '协议待检测'
}
const aiServices = computed(() => Object.values(inspectNode.value?.unlocks?.ai ?? {}))
const streamingServices = computed(() => Object.values(inspectNode.value?.unlocks?.streaming ?? {}))
const serviceCount = computed(() => aiServices.value.length + streamingServices.value.length)

// UI Handbook: Lock body scroll when modal is open & handle Esc dismissal
watch(inspectNode, (node) => {
  if (typeof document !== 'undefined') {
    document.body.style.overflow = node ? 'hidden' : ''
  }
})

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && inspectNode.value) {
    inspectNode.value = null
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
  if (typeof document !== 'undefined') {
    document.body.style.overflow = ''
  }
})
</script>

<template>
  <div class="public-shell">
    <!-- Metallic Luxury Header -->
    <header class="public-header">
      <div class="public-brand">
        <div class="brand-metal-emblem">
          <Sparkles :size="16" class="emblem-spark" />
          <span class="brand-char">探</span>
        </div>
        <div class="brand-titles">
          <span class="brand-main">鸡探长<span class="brand-tag">TITANIUM</span></span>
          <small class="brand-sub">DETECTIVE CHICKEN · 全球 VPS 质量与 AI/流媒体全景研判</small>
        </div>
      </div>

      <!-- 未登录状态不展示视图切换 tabs -->

      <!-- Top Action Bar -->
      <div class="public-actions">
        <div class="telemetry-live-pill">
          <span class="pulse-emerald"></span>
          <span>100% 真实探针并发采集</span>
        </div>
        <button
          class="metal-icon-btn"
          title="切换主题"
          aria-label="切换主题"
          @click="$emit('theme')"
        >
          <Sun v-if="dark" :size="16" /><Moon v-else :size="16" />
        </button>
        <button
          class="metal-icon-btn"
          title="刷新数据"
          aria-label="刷新数据"
          @click="$emit('refresh')"
        >
          <RefreshCw :size="16" :class="{ spinning: refreshing }" />
        </button>
        <button class="primary-metal-btn public-login" @click="$emit('login')">
          <LogIn :size="15" /><span>登录控制台</span>
        </button>
      </div>
    </header>

    <main class="public-main">
      <div v-if="loading" class="loading-line public-loading"></div>

      <!-- 金属卡片展厅 (未登录唯一视图) -->
      <section class="public-cards-view">
        <div class="fleet-toolbar">
          <div class="region-filter-chips">
            <button
              class="chip-btn"
              :class="{ active: activeRegionFilter === 'all' }"
              @click="activeRegionFilter = 'all'"
            >
              全部地区 ({{ data.nodes.length }})
            </button>
            <button
              v-for="region in availableRegions"
              :key="region"
              class="chip-btn"
              :class="{ active: activeRegionFilter === region }"
              @click="activeRegionFilter = region"
            >
              {{ region }}
            </button>
          </div>

          <div class="fleet-meta-hint">
            <span>选择节点查看完整出口 IP 质量、AI 与流媒体可用性报告</span>
          </div>
        </div>

        <div v-if="filteredNodes.length" class="credit-cards-grid">
          <CreditCardNode
            v-for="node in filteredNodes"
            :key="node.id"
            :node="node"
            :selected="selectedNode?.id === node.id"
            @select="onSelectNode"
            @inspect-service="onInspectService"
          />
        </div>

        <div v-else class="metal-empty-deck">
          <Server :size="32" />
          <strong>当前筛选无节点</strong>
          <span>请切换筛选条件或等待探针上报</span>
        </div>
      </section>

    </main>

    <!-- Card Back: 3D Flip & Expand Detailed Inspection Modal -->
    <Teleport to="body">
      <div v-if="inspectNode" class="modal-backdrop" @click="inspectNode = null">
        <div
          class="metal-inspect-modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="inspection-report-title"
          @click.stop
        >
          <div class="modal-foil"></div>
          <header class="modal-head">
            <div class="modal-title-wrap">
              <span class="report-eyebrow">DEEP INSPECTION REPORT</span>
              <div class="report-title-line">
                <h3 id="inspection-report-title">{{ inspectNode.name }}</h3>
                <span class="report-location">{{ inspectNode.country_code || inspectNode.region || '—' }} · {{ detectedUsage(inspectNode) }}</span>
                <span class="hud-status-tag" :class="inspectNode.status">
                  {{ inspectNode.status === 'offline' ? '已离线' : inspectNode.status === 'alert' ? '告警' : inspectNode.status === 'warning' ? '注意' : inspectNode.status === 'pending' ? '待接入' : '在线' }}
                </span>
              </div>
              <small>{{ inspectNode.provider || 'VPS NODE' }} · {{ inspectNode.region || '地区待检测' }}</small>
            </div>
            <div class="head-actions">
              <button class="metal-close-btn" title="关闭报告" aria-label="关闭深度检测报告" @click="inspectNode = null">
                <X :size="18" />
              </button>
            </div>
          </header>

          <div class="modal-body">
            <div class="modal-metric-ribbon">
              <div class="summary-metric">
                <span>纯净度评分</span>
                <strong :class="hasQualityResult(inspectNode) ? qualityClass(inspectNode.risk) : 'text-muted'">
                  {{ hasQualityResult(inspectNode) ? 100 - inspectNode.risk : '—' }} <small v-if="hasQualityResult(inspectNode)">分</small>
                </strong>
              </div>
              <div class="summary-metric">
                <span>欺诈风险级别</span>
                <strong :class="hasQualityResult(inspectNode) ? qualityClass(inspectNode.risk) : 'text-muted'">
                  {{ riskLabel(inspectNode) }}
                </strong>
              </div>
              <div class="summary-metric">
                <span>属性 / 宽带类型</span>
                <strong class="text-sky">{{ detectedIPType(inspectNode) }} · {{ detectedUsage(inspectNode) }}</strong>
              </div>
              <div class="summary-metric">
                <span>DNSBL 邮件信誉</span>
                <strong :class="!hasQualityResult(inspectNode) ? 'text-muted' : inspectNode.dnsbl > 0 ? 'text-danger' : 'text-emerald'">
                  {{ !hasQualityResult(inspectNode) ? '等待检测' : inspectNode.dnsbl > 0 ? `命中 ${inspectNode.dnsbl} 项` : '未命中 (安全)' }}
                </strong>
              </div>
            </div>

            <div class="network-hud-bar">
              <div class="hud-item">
                <span class="hud-label">在线状态</span>
                <span class="hud-val" :class="inspectNode.status === 'offline' ? 'text-muted' : 'text-emerald'">
                  {{ inspectNode.status === 'offline' ? '已离线' : '在线运行' }}
                </span>
                <span class="hud-sub">心跳: {{ relative(inspectNode.last_seen) }}</span>
              </div>
              <div class="hud-item">
                <span class="hud-label">公网脱敏 IP</span>
                <code class="modal-ip">{{ maskedAddresses(inspectNode) }}</code>
              </div>
              <div class="hud-item">
                <span class="hud-label">协议栈 & WARP</span>
                <span class="hud-val" :class="(inspectNode.is_warp || inspectNode.warp4 || inspectNode.warp6) ? 'text-gold' : 'text-sky'">
                  {{ protocolLabel(inspectNode) }}
                </span>
              </div>
              <div class="hud-item hud-asn">
                <span class="hud-label">自治系统 ASN</span>
                <strong class="hud-val">{{ detectedASN(inspectNode) }}</strong>
                <span class="hud-sub" :title="detectedOrganization(inspectNode)">{{ detectedOrganization(inspectNode) }}</span>
              </div>
              <div class="hud-item">
                <span class="hud-label">地区归属</span>
                <span class="hud-val">{{ inspectNode.region }} ({{ inspectNode.country_code }})</span>
              </div>
            </div>

            <div class="modal-section-title">
              <span>SERVICE AVAILABILITY</span>
              <small>{{ serviceCount }} SERVICES</small>
            </div>

            <div class="modal-badges-grid">
              <div class="badge-category-group">
                <div class="group-label"><Bot :size="14" /><span>AI MODELS</span><small>{{ aiServices.length }}</small></div>
                <div class="group-badges">
                  <MetalBadge
                    v-for="svc in aiServices"
                    :key="svc.id"
                    :service-id="svc.id"
                    :name="svc.name"
                    :status="svc.status"
                    :region="svc.region"
                    :quality="svc.quality"
                    :latency-ms="svc.latency_ms"
                    :node-region="inspectNode.country_code || inspectNode.region"
                    size="md"
                    :interactive="false"
                  />
                </div>
              </div>

              <div class="badge-category-group">
                <div class="group-label"><Tv :size="14" /><span>STREAMING</span><small>{{ streamingServices.length }}</small></div>
                <div class="group-badges">
                  <MetalBadge
                    v-for="svc in streamingServices"
                    :key="svc.id"
                    :service-id="svc.id"
                    :name="svc.name"
                    :status="svc.status"
                    :region="svc.region"
                    :quality="svc.quality"
                    :latency-ms="svc.latency_ms"
                    :node-region="inspectNode.country_code || inspectNode.region"
                    size="md"
                    :interactive="false"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Luxury Metallic Footer -->
    <footer class="public-footer">
      <div class="footer-meta">
        <span>鸡探长 (Detective Chicken) · 100% 真实探针驱动 · 20+ 款主流服务态势研判平台</span>
        <span>公开展示已实施末段脱敏保护 · 账号后台可享完整探针配置与告警管理</span>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.public-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg);
  color: var(--text);
  font-family: 'Outfit', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  transition: background-color 0.2s ease, color 0.2s ease;
}

/* Header */
.public-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 28px;
  background: var(--surface);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 100;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.public-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-metal-emblem {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #38bdf8 0%, #1d4ed8 100%);
  display: grid;
  place-items: center;
  position: relative;
  box-shadow:
    0 0 16px rgba(56, 189, 248, 0.4),
    inset 0 1px 1px rgba(255, 255, 255, 0.6);
}
.brand-char {
  font-family: 'Fira Code', monospace;
  font-weight: 900;
  font-size: 16px;
  color: #fff;
}
.emblem-spark {
  position: absolute;
  top: -4px;
  right: -4px;
  color: #facc15;
}

.brand-titles {
  display: flex;
  flex-direction: column;
}
.brand-main {
  font-size: 17px;
  font-weight: 800;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 6px;
}
.brand-tag {
  font-size: 9px;
  font-weight: 900;
  background: linear-gradient(90deg, #38bdf8, #818cf8);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  border: 1px solid rgba(56, 189, 248, 0.4);
  padding: 1px 4px;
  border-radius: 4px;
}
.brand-sub {
  font-size: 10px;
  color: var(--muted);
  letter-spacing: 0.5px;
}

/* Nav Tabs */
.public-nav-tabs {
  display: flex;
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 4px;
  gap: 4px;
}

.public-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  background: transparent;
  border: 0;
  border-radius: 7px;
  color: var(--muted);
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}
.public-tab:hover {
  color: var(--text);
  background: var(--surface-3);
}
.public-tab.active {
  background: var(--surface);
  color: var(--primary);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
}

/* Top Actions */
.public-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.telemetry-live-pill {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 20px;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.25);
  font-size: 11px;
  color: #34d399;
  font-weight: 500;
}
.pulse-emerald {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}

.metal-icon-btn {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: var(--surface-2);
  border: 1px solid var(--border);
  color: var(--text);
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: all 0.15s ease;
}
.metal-icon-btn:hover {
  background: var(--surface-3);
  color: var(--text);
}
.metal-icon-btn:active {
  transform: scale(0.97);
}

.primary-metal-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  background: linear-gradient(135deg, #0284c7 0%, #0369a1 100%);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  color: #fff;
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(2, 132, 199, 0.35);
  transition: all 0.2s ease;
}
.primary-metal-btn:hover {
  filter: brightness(1.15);
  transform: translateY(-1px);
}
.primary-metal-btn:active {
  transform: scale(0.97);
}

/* UI Handbook: Strict tabular numbers for metrics and stats */
.gauge-value,
.modal-metric-ribbon strong,
.modal-ip,
.hud-val,
.rank-score-capsule,
.rank-badge-wrap {
  font-variant-numeric: tabular-nums;
  font-feature-settings: 'tnum';
}

/* Main */
.public-main {
  flex: 1;
  max-width: 1480px;
  width: 100%;
  margin: 0 auto;
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Hero Command Bar */
.hero-command-deck {
  display: flex;
  flex-direction: column;
  gap: 20px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 24px 28px;
  box-shadow: var(--shadow);
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 20px;
  background: rgba(56, 189, 248, 0.12);
  color: var(--primary);
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.8px;
  border: 1px solid rgba(56, 189, 248, 0.3);
  margin-bottom: 8px;
}

.hero-title {
  font-size: 26px;
  font-weight: 800;
  color: var(--text);
  letter-spacing: -0.5px;
}

.hero-desc {
  font-size: 13px;
  color: var(--muted);
  max-width: 860px;
  line-height: 1.6;
  margin-top: 6px;
}

/* Gauges Grid */
.hero-gauges-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.metal-gauge-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
}
.gauge-metal-layer {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 0% 0%, rgba(255, 255, 255, 0.04), transparent 60%);
  pointer-events: none;
}

.gauge-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.gauge-icon.fleet { background: rgba(56, 189, 248, 0.15); color: #38bdf8; }
.gauge-icon.ai { background: rgba(168, 85, 247, 0.15); color: #c084fc; }
.gauge-icon.streaming { background: rgba(16, 185, 129, 0.15); color: #34d399; }
.gauge-icon.purity { background: rgba(245, 158, 11, 0.15); color: #fbbf24; }

.gauge-meta {
  display: flex;
  flex-direction: column;
}
.gauge-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--muted);
}
.gauge-value {
  font-family: 'Fira Code', monospace;
  font-size: 20px;
  font-weight: 800;
  color: var(--text);
  margin: 1px 0;
}
.gauge-value small {
  font-size: 11px;
  color: var(--faint);
}
.gauge-sub {
  font-size: 10px;
  color: var(--muted);
}

.text-emerald { color: #10b981 !important; }
.text-sky { color: #38bdf8 !important; }
.text-gold { color: #f59e0b !important; }
.text-danger { color: #ef7777 !important; }
.text-muted { color: #7c8794 !important; }

/* Fleet Toolbar & Cards View */
.public-cards-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 20px;
}

.fleet-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.region-filter-chips {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.chip-btn {
  padding: 5px 12px;
  border-radius: 20px;
  background: var(--surface-2);
  border: 1px solid var(--border);
  color: var(--muted);
  font-size: 11.5px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}
.chip-btn:hover {
  background: var(--surface-3);
  color: var(--text);
}
.chip-btn.active {
  background: rgba(56, 189, 248, 0.18);
  color: var(--primary);
  border-color: rgba(56, 189, 248, 0.4);
}

.fleet-meta-hint {
  font-size: 11.5px;
  color: var(--muted);
}

.credit-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 420px), 440px));
  grid-auto-rows: 1fr;
  justify-content: start;
  align-items: stretch;
  gap: 20px;
}

.metal-empty-deck {
  padding: 48px;
  text-align: center;
  background: var(--surface-2);
  border: 1px dashed var(--border);
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--muted);
}

/* Rankings View */
.ranking-split-grid {
  display: grid;
  grid-template-columns: 1.5fr 1fr;
  gap: 20px;
}

.metal-panel {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 20px;
  box-shadow: var(--shadow);
}

.panel-head-metallic {
  display: flex;
  align-items: center;
  gap: 10px;
  border-bottom: 1px solid var(--border);
  padding-bottom: 12px;
  margin-bottom: 14px;
}
.panel-head-metallic h2 {
  font-size: 16px;
  font-weight: 700;
  color: var(--text);
}
.panel-head-metallic small {
  font-size: 11px;
  color: var(--muted);
}

.metal-ranking-table {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ranking-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.ranking-row:hover {
  background: var(--surface-3);
  border-color: var(--primary);
  transform: translateX(3px);
}

.rank-badge-wrap {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: var(--surface-3);
  display: grid;
  place-items: center;
  font-family: 'Fira Code', monospace;
  font-weight: 800;
  font-size: 12px;
  color: var(--text);
}
.podium-gold .rank-badge-wrap { background: #eab308; color: #000; }
.podium-silver .rank-badge-wrap { background: #cbd5e1; color: #000; }
.podium-bronze .rank-badge-wrap { background: #d97706; color: #fff; }

.rank-node-info {
  flex: 1;
}
.rank-node-info strong {
  display: block;
  font-size: 13px;
  color: var(--text);
}
.rank-node-info small {
  font-size: 11px;
  color: var(--muted);
}

.rank-services-count {
  font-size: 11px;
  color: var(--muted);
}

.rank-score-capsule {
  font-family: 'Fira Code', monospace;
  font-weight: 800;
  font-size: 13px;
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--surface-3);
}
.rank-score-capsule.risk-low { color: #10b981; }
.rank-score-capsule.risk-mid { color: #f59e0b; }
.rank-score-capsule.risk-high { color: #ef4444; }

.criteria-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.criteria-item {
  padding: 10px 12px;
  background: var(--surface-2);
  border-radius: 8px;
  border: 1px solid var(--border);
}
.criteria-item strong {
  font-size: 12.5px;
  color: var(--primary);
}
.criteria-item p {
  font-size: 11.5px;
  color: var(--muted);
  margin-top: 2px;
  line-height: 1.4;
}

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(5px);
  z-index: 1000;
  display: grid;
  place-items: center;
  padding: 20px;
}

.metal-inspect-modal {
  position: relative;
  width: 100%;
  max-width: 940px;
  max-height: min(92vh, 900px);
  overflow-y: auto;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  box-shadow: var(--shadow-lg);
  padding: 24px;
  animation: cardFlipExpand 0.28s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  transform-origin: center center;
}

@keyframes cardFlipExpand {
  0% {
    opacity: 0;
    transform: translateY(8px) scale(0.985);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-foil {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 0%, rgba(255, 255, 255, 0.06), transparent 70%);
  pointer-events: none;
}

.modal-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  border-bottom: 1px solid var(--border);
  padding-bottom: 16px;
  gap: 12px;
}
.modal-title-wrap {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  min-width: 0;
}
.report-eyebrow { color: var(--muted); font: 600 9px/1.3 'Fira Code', monospace; letter-spacing: 0.16em; }
.report-title-line { display: flex; align-items: center; gap: 9px; min-width: 0; }
.modal-title-wrap h3 {
  margin: 0;
  overflow: hidden;
  color: var(--text);
  font-size: 20px;
  font-weight: 650;
  letter-spacing: 0.015em;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.report-location {
  flex: none;
  padding: 2px 6px;
  border: 1px solid rgba(125, 211, 252, 0.24);
  border-radius: 4px;
  color: #8bd8f8;
  font: 600 9px/1.4 'Fira Code', monospace;
}
.hud-status-tag {
  flex: none;
  padding: 1px 7px;
  border-radius: 9999px;
  font-size: 9.5px;
  font-weight: 600;
  line-height: 1.4;
}
.hud-status-tag.online { background: rgba(16, 185, 129, 0.15); color: #10b981; }
.hud-status-tag.warning { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }
.hud-status-tag.alert { background: rgba(239, 68, 68, 0.15); color: #ef4444; }
.hud-status-tag.offline { background: rgba(100, 116, 139, 0.18); color: #64748b; }
.hud-status-tag.pending { background: rgba(56, 189, 248, 0.15); color: #38bdf8; }
.modal-title-wrap small {
  color: var(--muted);
  font-size: 10px;
}

.head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.metal-close-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--surface-2);
  border: 1px solid var(--border);
  color: var(--text);
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: background 180ms ease, color 180ms ease, border-color 180ms ease;
}
.metal-close-btn:hover {
  background: var(--surface-3);
  border-color: var(--primary);
  color: var(--text);
}
.metal-close-btn:focus-visible { outline: 2px solid #7dd3fc; outline-offset: 2px; }

.modal-body {
  margin-top: 18px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-metric-ribbon {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  border-block: 1px solid var(--border);
}
.summary-metric { min-width: 0; padding: 11px 14px; }
.summary-metric + .summary-metric { border-left: 1px solid var(--border); }
.modal-metric-ribbon span {
  display: block;
  color: var(--muted);
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.04em;
}
.modal-metric-ribbon strong {
  display: block;
  overflow: hidden;
  margin-top: 3px;
  color: var(--text);
  font-family: 'Fira Code', monospace;
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.modal-ip {
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  color: #8bd8f8;
  overflow-wrap: anywhere;
}

.network-hud-bar {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 0;
  padding: 3px 0;
  background: var(--surface-2);
  border-radius: 8px;
  border: 1px solid var(--border);
}
.hud-item {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 9px 13px;
}
.hud-item + .hud-item { border-left: 1px solid var(--border); }
.hud-label {
  font-size: 9px;
  font-weight: 600;
  color: var(--muted);
  letter-spacing: 0.5px;
}
.hud-val {
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  font-weight: 550;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.hud-sub { overflow: hidden; color: var(--muted); font-size: 9px; line-height: 1.35; text-overflow: ellipsis; white-space: nowrap; }

.modal-section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text);
  font: 600 10px/1.4 'Fira Code', monospace;
  letter-spacing: 0.12em;
}
.modal-section-title small { color: var(--muted); font-size: 8px; letter-spacing: 0.08em; }

.modal-badges-grid {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.badge-category-group {
  display: flex;
  flex-direction: column;
  gap: 7px;
}
.group-label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--muted);
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.08em;
}
.group-label small { color: var(--faint); font: 500 8px/1 'Fira Code', monospace; }
.group-badges {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 7px;
}
.group-badges :deep(.metal-badge) {
  width: 100%;
  min-width: 0;
  min-height: 52px;
  padding: 7px 9px;
  border-color: var(--border);
  border-radius: 7px;
  background: var(--surface-2);
}
.group-badges :deep(.brand-logo-wrap) { width: 30px; height: 30px; background: transparent; box-shadow: none; }
.group-badges :deep(.brand-svg) { width: 21px; height: 21px; }
.group-badges :deep(.metal-badge-meta) { min-width: 0; }
.group-badges :deep(.badge-title), .group-badges :deep(.status-desc) { overflow: hidden; text-overflow: ellipsis; }
.group-badges :deep(.badge-title) { font-size: 10px; font-weight: 550; }
.group-badges :deep(.status-desc) { max-width: 110px; font-size: 8.5px; }
.group-badges :deep(.metal-foil), .group-badges :deep(.metal-light-sweep) { opacity: .25; }

/* Footer */
.public-footer {
  margin-top: auto;
  border-top: 1px solid var(--border);
  padding: 18px 28px;
  background: var(--surface);
}
.footer-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: var(--muted);
}

@media (max-width: 980px) {
  .hero-gauges-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .ranking-split-grid {
    grid-template-columns: 1fr;
  }
  .credit-cards-grid {
    grid-template-columns: minmax(0, 440px);
  }
}

@media (max-width: 768px) {
  .public-header { align-items: flex-start; padding: 12px 18px; }
  .brand-sub, .telemetry-live-pill { display: none; }
  .public-actions { gap: 7px; }
  .public-main { padding: 18px; }
  .group-badges { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .modal-metric-ribbon, .network-hud-bar { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .summary-metric:nth-child(3), .summary-metric:nth-child(4) { border-top: 1px solid var(--border); }
  .summary-metric:nth-child(3) { border-left: 0; }
  .hud-item:nth-child(3), .hud-item:nth-child(4) { border-top: 1px solid var(--border); }
  .hud-item:nth-child(3) { border-left: 0; }
  .footer-meta { align-items: flex-start; flex-direction: column; gap: 5px; }
}

@media (max-width: 520px) {
  .public-header { align-items: center; padding: 10px 12px; }
  .brand-metal-emblem { width: 32px; height: 32px; }
  .brand-main { font-size: 15px; }
  .brand-tag { display: none; }
  .metal-icon-btn { width: 36px; height: 36px; }
  .public-login { width: 36px; height: 36px; justify-content: center; padding: 0; }
  .public-login span { display: none; }
  .public-main { padding: 14px 12px 20px; }
  .public-cards-view { padding-top: 8px; }
  .fleet-toolbar { align-items: flex-start; }
  .fleet-meta-hint { line-height: 1.45; }
  .credit-cards-grid { grid-template-columns: minmax(0, 1fr); gap: 14px; }
  .modal-backdrop { align-items: start; padding: 8px; overflow-y: auto; }
  .metal-inspect-modal { max-height: none; padding: 18px 14px; border-radius: 12px; }
  .report-title-line { align-items: flex-start; flex-direction: column; gap: 5px; }
  .modal-title-wrap h3 { font-size: 18px; }
  .modal-metric-ribbon, .network-hud-bar, .group-badges { grid-template-columns: 1fr; }
  .summary-metric + .summary-metric, .hud-item + .hud-item { border-left: 0; border-top: 1px solid var(--border); }
  .group-badges :deep(.metal-badge) { min-height: 50px; }
  .public-footer { padding: 14px 12px; }
}

@media (prefers-reduced-motion: reduce) {
  .metal-inspect-modal { animation: none; }
  .spinning { animation: none; }
}
</style>
