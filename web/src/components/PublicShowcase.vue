<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  Activity,
  Bot,
  Check,
  Copy,
  CreditCard,
  Layers,
  LogIn,
  Moon,
  RefreshCw,
  Server,
  ShieldCheck,
  Sparkles,
  Sun,
  Terminal,
  Trophy,
  Tv,
  Wrench,
  X,
  Zap,
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

// 未登录状态只展示卡片视图
const publicView = ref<'cards'>('cards')
const selectedNode = ref<Node | null>(null)
const inspectNode = ref<Node | null>(null)
const activeRegionFilter = ref<string>('all')
const copiedCommand = ref(false)

const qualityClass = (risk: number) =>
  risk >= 60 ? 'risk-high' : risk >= 35 ? 'risk-mid' : 'risk-low'

const relative = (input?: string) => {
  if (!input || new Date(input).getFullYear() <= 1) return '等待首次检测'
  const seconds = Math.max(
    0,
    Math.floor((Date.now() - new Date(input).getTime()) / 1000),
  )
  if (seconds < 60) return `${seconds} 秒前`
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} 小时前`
  return `${Math.floor(seconds / 86400)} 天前`
}

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

// Overall average IP purity score
const averagePurityScore = computed(() => {
  if (!props.data.nodes.length) return 100
  const sum = props.data.nodes.reduce((acc, n) => acc + (100 - n.risk), 0)
  return Math.round(sum / props.data.nodes.length)
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
      <section class="public-cards-view" style="padding-top: 20px;">
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
            <span>💳 点击卡片 3D 翻转查看完整 AI 与流媒体解锁及 IP 纯净度研判</span>
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
        <div class="metal-inspect-modal" @click.stop>
          <div class="modal-foil"></div>
          <header class="modal-head">
            <div class="modal-title-wrap">
              <span class="modal-flag">{{ inspectNode.country_code || inspectNode.region || '未知地区' }}-{{ detectedUsage(inspectNode) }}</span>
              <div>
                <div style="display: flex; align-items: center; gap: 8px;">
                  <h3>{{ inspectNode.name }}</h3>
                  <span class="card-back-tag">卡片背面 · 深度研判报告</span>
                </div>
                <small>{{ inspectNode.provider }} · {{ detectedOrganization(inspectNode) }} · {{ detectedASN(inspectNode) }}</small>
              </div>
            </div>
            <div class="head-actions">
              <button class="metal-close-btn" title="关闭背面" @click="inspectNode = null">
                <X :size="18" />
              </button>
            </div>
          </header>

          <div class="modal-body">
            <!-- Terminal-style IP Quality & Threat Ribbon (参考检测脚本) -->
            <div class="modal-metric-ribbon">
              <div>
                <span>纯净度评分</span>
                <strong :class="qualityClass(inspectNode.risk)">{{ 100 - inspectNode.risk }} <small>分</small></strong>
              </div>
              <div>
                <span>欺诈风险级别</span>
                <strong :class="qualityClass(inspectNode.risk)">
                  {{ inspectNode.risk <= 20 ? '极净优质' : inspectNode.risk <= 50 ? '低风险' : '中度注意' }}
                </strong>
              </div>
              <div>
                <span>属性 / 宽带类型</span>
                <strong style="color: #38bdf8;">{{ detectedIPType(inspectNode) }} · {{ detectedUsage(inspectNode) }}</strong>
              </div>
              <div>
                <span>DNSBL 邮件信誉</span>
                <strong :class="inspectNode.dnsbl > 0 ? 'text-danger' : 'text-emerald'">
                  {{ inspectNode.dnsbl > 0 ? `命中 ${inspectNode.dnsbl} 项` : '未命中 (安全)' }}
                </strong>
              </div>
            </div>

            <!-- Network & Routing Info Bar -->
            <div class="network-hud-bar">
              <div class="hud-item">
                <span class="hud-label">公网脱敏 IP</span>
                <code class="modal-ip">{{ maskedAddresses(inspectNode) }}</code>
              </div>
              <div class="hud-item">
                <span class="hud-label">协议栈 & WARP</span>
                <span class="hud-val" :style="{ color: (inspectNode.is_warp || inspectNode.warp4 || inspectNode.warp6) ? '#f59e0b' : '#38bdf8' }">
                  {{ protocolLabel(inspectNode) }}
                </span>
              </div>
              <div class="hud-item">
                <span class="hud-label">自治系统 ASN</span>
                <span class="hud-val">{{ detectedASN(inspectNode) }} ({{ detectedOrganization(inspectNode) }})</span>
              </div>
              <div class="hud-item">
                <span class="hud-label">地区归属</span>
                <span class="hud-val">{{ inspectNode.region }} ({{ inspectNode.country_code }})</span>
              </div>
            </div>

            <div class="modal-section-title">
              <Sparkles :size="14" />
              <span>全量 20+ 款品牌 AI 与主流流媒体解锁全景 (含精准延迟与质量)</span>
            </div>

            <div class="modal-badges-grid">
              <!-- AI Category -->
              <div class="badge-category-group">
                <span class="group-label">🤖 AI 生产力大模型 (10 款)</span>
                <div class="group-badges">
                  <MetalBadge
                    v-for="svc in Object.values(inspectNode.unlocks?.ai ?? {})"
                    :key="svc.id"
                    :service-id="svc.id"
                    :name="svc.name"
                    :status="svc.status"
                    :region="svc.region"
                    :quality="svc.quality"
                    :latency-ms="svc.latency_ms"
                    size="md"
                  />
                </div>
              </div>

              <!-- Streaming Category -->
              <div class="badge-category-group">
                <span class="group-label">🎬 流媒体与娱乐矩阵 (11 款)</span>
                <div class="group-badges">
                  <MetalBadge
                    v-for="svc in Object.values(inspectNode.unlocks?.streaming ?? {})"
                    :key="svc.id"
                    :service-id="svc.id"
                    :name="svc.name"
                    :status="svc.status"
                    :region="svc.region"
                    :quality="svc.quality"
                    :latency-ms="svc.latency_ms"
                    size="md"
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

/* Fleet Toolbar & Cards View */
.public-cards-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
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
  grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
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
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  z-index: 1000;
  display: grid;
  place-items: center;
  padding: 20px;
}

.metal-inspect-modal {
  position: relative;
  width: 100%;
  max-width: 820px;
  max-height: 90vh;
  overflow-y: auto;
  background: linear-gradient(145deg, #161e27, #0c1117);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.8), 0 0 32px rgba(56, 189, 248, 0.2);
  padding: 24px;
  animation: cardFlipExpand 0.38s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  transform-origin: center center;
}

@keyframes cardFlipExpand {
  0% {
    opacity: 0;
    transform: perspective(1000px) rotateY(-60deg) scale(0.85);
  }
  60% {
    transform: perspective(1000px) rotateY(6deg) scale(1.01);
  }
  100% {
    opacity: 1;
    transform: perspective(1000px) rotateY(0deg) scale(1);
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
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: 14px;
  gap: 12px;
}
.modal-title-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
}
.modal-flag {
  font-family: 'Fira Code', monospace;
  font-size: 14px;
  font-weight: 800;
  background: rgba(56, 189, 248, 0.2);
  color: #38bdf8;
  padding: 4px 8px;
  border-radius: 6px;
  border: 1px solid rgba(56, 189, 248, 0.4);
}
.modal-title-wrap h3 {
  font-size: 18px;
  font-weight: 800;
  color: #f8fafc;
}
.card-back-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 12px;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  border: 1px solid rgba(56, 189, 248, 0.35);
}
.modal-title-wrap small {
  font-size: 11px;
  color: #94a3b8;
}

.head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.metal-close-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.08);
  border: 0;
  color: #cbd5e1;
  display: grid;
  place-items: center;
  cursor: pointer;
}
.metal-close-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.modal-body {
  margin-top: 18px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.modal-metric-ribbon {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  padding: 12px 14px;
  background: rgba(0, 0, 0, 0.35);
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.06);
}
.modal-metric-ribbon span {
  display: block;
  font-size: 10px;
  color: #64748b;
  font-weight: 700;
}
.modal-metric-ribbon strong {
  display: block;
  font-family: 'Fira Code', monospace;
  font-size: 14px;
  color: #f8fafc;
  margin-top: 2px;
}
.modal-ip {
  font-family: 'Fira Code', monospace;
  font-size: 11px;
  color: #38bdf8;
}

.network-hud-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  padding: 10px 14px;
  background: rgba(0, 0, 0, 0.25);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}
.hud-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.hud-label {
  font-size: 9px;
  font-weight: 700;
  color: #64748b;
  letter-spacing: 0.5px;
}
.hud-val {
  font-family: 'Fira Code', monospace;
  font-size: 11px;
  font-weight: 600;
  color: #f1f5f9;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.modal-section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 700;
  color: #38bdf8;
  letter-spacing: 0.5px;
}

.modal-badges-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.badge-category-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.group-label {
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
}
.group-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

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
    grid-template-columns: 1fr;
  }
}
</style>
