<script setup lang="ts">
import { computed, ref } from 'vue'
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
  background: #090d12;
  color: #f8fafc;
  font-family: 'Outfit', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* Header */
.public-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 28px;
  background: rgba(13, 18, 24, 0.85);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
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
  color: #f8fafc;
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
  color: #64748b;
  letter-spacing: 0.5px;
}

/* Nav Tabs */
.public-nav-tabs {
  display: flex;
  background: rgba(0, 0, 0, 0.45);
  border: 1px solid rgba(255, 255, 255, 0.1);
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
  color: #94a3b8;
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}
.public-tab:hover {
  color: #f8fafc;
  background: rgba(255, 255, 255, 0.05);
}
.public-tab.active {
  background: linear-gradient(145deg, #1e293b, #0f172a);
  color: #38bdf8;
  border: 1px solid rgba(56, 189, 248, 0.35);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
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
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #cbd5e1;
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: all 0.15s ease;
}
.metal-icon-btn:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #f8fafc;
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
  background: linear-gradient(145deg, #131922 0%, #0d1219 100%);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 24px 28px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 20px;
  background: rgba(56, 189, 248, 0.12);
  color: #38bdf8;
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.8px;
  border: 1px solid rgba(56, 189, 248, 0.3);
  margin-bottom: 8px;
}

.hero-title {
  font-size: 26px;
  font-weight: 800;
  color: #f8fafc;
  letter-spacing: -0.5px;
}

.hero-desc {
  font-size: 13px;
  color: #94a3b8;
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
  background: rgba(0, 0, 0, 0.35);
  border: 1px solid rgba(255, 255, 255, 0.08);
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
  color: #94a3b8;
}
.gauge-value {
  font-family: 'Fira Code', monospace;
  font-size: 20px;
  font-weight: 800;
  color: #f8fafc;
  margin: 1px 0;
}
.gauge-value small {
  font-size: 11px;
  color: #64748b;
}
.gauge-sub {
  font-size: 10px;
  color: #64748b;
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
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #94a3b8;
  font-size: 11.5px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}
.chip-btn:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}
.chip-btn.active {
  background: rgba(56, 189, 248, 0.2);
  color: #38bdf8;
  border-color: rgba(56, 189, 248, 0.4);
}

.fleet-meta-hint {
  font-size: 11.5px;
  color: #64748b;
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
  background: rgba(0, 0, 0, 0.3);
  border: 1px dashed rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #64748b;
}

/* Rankings View */
.ranking-split-grid {
  display: grid;
  grid-template-columns: 1.5fr 1fr;
  gap: 20px;
}

.metal-panel {
  background: linear-gradient(145deg, #131922 0%, #0e141b 100%);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
}

.panel-head-metallic {
  display: flex;
  align-items: center;
  gap: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding-bottom: 12px;
  margin-bottom: 14px;
}
.panel-head-metallic h2 {
  font-size: 16px;
  font-weight: 700;
  color: #f8fafc;
}
.panel-head-metallic small {
  font-size: 11px;
  color: #64748b;
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
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.ranking-row:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(56, 189, 248, 0.3);
  transform: translateX(3px);
}

.rank-badge-wrap {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.08);
  display: grid;
  place-items: center;
  font-family: 'Fira Code', monospace;
  font-weight: 800;
  font-size: 12px;
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
  color: #f8fafc;
}
.rank-node-info small {
  font-size: 11px;
  color: #64748b;
}

.rank-services-count {
  font-size: 11px;
  color: #94a3b8;
}

.rank-score-capsule {
  font-family: 'Fira Code', monospace;
  font-weight: 800;
  font-size: 13px;
  padding: 2px 8px;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.4);
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
  background: rgba(0, 0, 0, 0.25);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.04);
}
.criteria-item strong {
  font-size: 12.5px;
  color: #38bdf8;
}
.criteria-item p {
  font-size: 11.5px;
  color: #94a3b8;
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
  background: linear-gradient(145deg, #151a20, #0b0f13);
  border: 1px solid rgba(226, 232, 240, 0.17);
  border-radius: 16px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.78), inset 0 1px 0 rgba(255, 255, 255, 0.12);
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
  border-bottom: 1px solid rgba(226, 232, 240, 0.1);
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
.report-eyebrow { color: #7c8794; font: 600 9px/1.3 'Fira Code', monospace; letter-spacing: 0.16em; }
.report-title-line { display: flex; align-items: center; gap: 9px; min-width: 0; }
.modal-title-wrap h3 {
  margin: 0;
  overflow: hidden;
  color: #f1f5f9;
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
.modal-title-wrap small {
  color: #7f8a98;
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
  background: rgba(255, 255, 255, 0.045);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #cbd5e1;
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: background 180ms ease, color 180ms ease, border-color 180ms ease;
}
.metal-close-btn:hover {
  background: rgba(255, 255, 255, 0.09);
  border-color: rgba(255, 255, 255, 0.16);
  color: #f8fafc;
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
  border-block: 1px solid rgba(226, 232, 240, 0.09);
}
.summary-metric { min-width: 0; padding: 11px 14px; }
.summary-metric + .summary-metric { border-left: 1px solid rgba(226, 232, 240, 0.08); }
.modal-metric-ribbon span {
  display: block;
  color: #707b88;
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.04em;
}
.modal-metric-ribbon strong {
  display: block;
  overflow: hidden;
  margin-top: 3px;
  color: #e6ebf0;
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
  grid-template-columns: repeat(4, 1fr);
  gap: 0;
  padding: 3px 0;
  background: rgba(0, 0, 0, 0.18);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.06);
}
.hud-item {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 9px 13px;
}
.hud-item + .hud-item { border-left: 1px solid rgba(226, 232, 240, 0.07); }
.hud-label {
  font-size: 9px;
  font-weight: 600;
  color: #6f7a87;
  letter-spacing: 0.5px;
}
.hud-val {
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  font-weight: 550;
  color: #dce2e8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.hud-sub { overflow: hidden; color: #929daa; font-size: 9px; line-height: 1.35; text-overflow: ellipsis; white-space: nowrap; }

.modal-section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #b9c2cc;
  font: 600 10px/1.4 'Fira Code', monospace;
  letter-spacing: 0.12em;
}
.modal-section-title small { color: #687481; font-size: 8px; letter-spacing: 0.08em; }

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
  color: #97a3af;
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.08em;
}
.group-label small { color: #687481; font: 500 8px/1 'Fira Code', monospace; }
.group-badges {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 7px;
}
.group-badges :deep(.metal-badge) { width: 100%; min-width: 0; min-height: 52px; padding: 7px 9px; border-color: rgba(226,232,240,.1); border-radius: 7px; background: rgba(4,7,10,.32); }
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
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding: 18px 28px;
  background: rgba(10, 14, 20, 0.9);
}
.footer-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: #64748b;
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
  .summary-metric:nth-child(3), .summary-metric:nth-child(4) { border-top: 1px solid rgba(226, 232, 240, 0.08); }
  .summary-metric:nth-child(3) { border-left: 0; }
  .hud-item:nth-child(3), .hud-item:nth-child(4) { border-top: 1px solid rgba(226, 232, 240, 0.07); }
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
  .summary-metric + .summary-metric, .hud-item + .hud-item { border-left: 0; border-top: 1px solid rgba(226, 232, 240, 0.08); }
  .group-badges :deep(.metal-badge) { min-height: 50px; }
  .public-footer { padding: 14px 12px; }
}

@media (prefers-reduced-motion: reduce) {
  .metal-inspect-modal { animation: none; }
  .spinning { animation: none; }
}
</style>
