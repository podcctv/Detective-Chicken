<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  Activity,
  Bot,
  CreditCard,
  Globe2,
  Layers,
  LogIn,
  Moon,
  Radar,
  RefreshCw,
  Server,
  ShieldCheck,
  Sparkles,
  Sun,
  Trophy,
  Tv,
  X,
  Zap,
} from '@lucide/vue'

import CreditCardNode from './CreditCardNode.vue'
import MetalBadge from './MetalBadge.vue'
import Globe3D from './Globe3D.vue'
import RadarScanner3D from './RadarScanner3D.vue'
import UnlockMatrix from './UnlockMatrix.vue'
import type { Dashboard, Node } from '../types'

const props = defineProps<{
  data: Dashboard
  loading: boolean
  dark: boolean
  refreshing: boolean
}>()

defineEmits<{ login: []; refresh: []; theme: [] }>()

const publicView = ref<'cards' | 'globe' | 'matrix' | 'ranking'>('cards')
const selectedNode = ref<Node | null>(null)
const inspectNode = ref<Node | null>(null)
const activeRegionFilter = ref<string>('all')

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
          <small class="brand-sub">DETECTIVE CHICKEN · 节点质量全息研判平台</small>
        </div>
      </div>

      <!-- Segmented Metallic View Tabs -->
      <div class="public-nav-tabs" role="tablist" aria-label="视图模式切换">
        <button
          class="public-tab"
          :class="{ active: publicView === 'cards' }"
          role="tab"
          @click="publicView = 'cards'"
        >
          <CreditCard :size="15" />
          <span>金属卡片展厅</span>
        </button>
        <button
          class="public-tab"
          :class="{ active: publicView === 'globe' }"
          role="tab"
          @click="publicView = 'globe'"
        >
          <Globe2 :size="15" />
          <span>3D 全球态势</span>
        </button>
        <button
          class="public-tab"
          :class="{ active: publicView === 'matrix' }"
          role="tab"
          @click="publicView = 'matrix'"
        >
          <Layers :size="15" />
          <span>20+ 款品牌矩阵</span>
        </button>
        <button
          class="public-tab"
          :class="{ active: publicView === 'ranking' }"
          role="tab"
          @click="publicView = 'ranking'"
        >
          <Trophy :size="15" />
          <span>天梯排行榜</span>
        </button>
      </div>

      <!-- Top Action Bar -->
      <div class="public-actions">
        <div class="telemetry-live-pill">
          <span class="pulse-emerald"></span>
          <span>100% 真实探针实时同步</span>
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
      <!-- Luxury Titanium Hero Command Bar -->
      <section class="hero-command-deck">
        <div class="hero-intro">
          <div class="hero-badge">
            <ShieldCheck :size="14" />
            <span>ENTERPRISE GRADE IP QUALITY & MULTI-MODEL DISCOVERY</span>
          </div>
          <h1 class="hero-title">全球 VPS 算力资产质量与解锁全息看板</h1>
          <p class="hero-desc">
            无缝洞悉脱敏后全球节点的 IP 纯净度、权威风险数据库评分（Scamalytics/AbuseIPDB/IPQS）与 20+ 款主流流媒体（Netflix/Disney+/YouTube）及 AI 模型（ChatGPT/Claude/Gemini/DeepSeek）的真实解锁生态。
          </p>
        </div>

        <!-- 4 Luxury Titanium Metric Gauges -->
        <div class="hero-gauges-grid">
          <div class="metal-gauge-card">
            <div class="gauge-metal-layer"></div>
            <div class="gauge-icon fleet"><Server :size="20" /></div>
            <div class="gauge-meta">
              <span class="gauge-label">全网公开资产</span>
              <strong class="gauge-value">{{ data.stats.total ?? data.nodes.length }} <small>NODES</small></strong>
              <span class="gauge-sub">{{ data.stats.online ?? data.nodes.length }} 台实时在线</span>
            </div>
          </div>

          <div class="metal-gauge-card">
            <div class="gauge-metal-layer"></div>
            <div class="gauge-icon ai"><Bot :size="20" /></div>
            <div class="gauge-meta">
              <span class="gauge-label">AI 模型全域解锁率</span>
              <strong class="gauge-value text-emerald">{{ data.stats.ai_unlock_rate ?? 100 }}%</strong>
              <span class="gauge-sub">Claude / Gemini / GPT-4o</span>
            </div>
          </div>

          <div class="metal-gauge-card">
            <div class="gauge-metal-layer"></div>
            <div class="gauge-icon streaming"><Tv :size="20" /></div>
            <div class="gauge-meta">
              <span class="gauge-label">流媒体原生贯通率</span>
              <strong class="gauge-value text-sky">{{ data.stats.streaming_unlock_rate ?? 100 }}%</strong>
              <span class="gauge-sub">4K 原生 / 免跨区限制</span>
            </div>
          </div>

          <div class="metal-gauge-card">
            <div class="gauge-metal-layer"></div>
            <div class="gauge-icon purity"><ShieldCheck :size="20" /></div>
            <div class="gauge-meta">
              <span class="gauge-label">全网综合纯净度</span>
              <strong class="gauge-value text-gold">{{ averagePurityScore }} <small>PTS</small></strong>
              <span class="gauge-sub">多源欺诈数据库聚合</span>
            </div>
          </div>
        </div>
      </section>

      <div v-if="loading" class="loading-line public-loading"></div>

      <!-- VIEW 1: Credit Card Fleet Gallery (默认金属卡片展厅) -->
      <section v-if="publicView === 'cards'" class="public-cards-view">
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
            <span>🖱️ 悬浮卡片体验 3D 陀螺仪视差与全息反光</span>
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

      <!-- VIEW 2: 3D Holographic Globe & Radar Command Hub -->
      <section v-else-if="publicView === 'globe'" class="public-3d-hub">
        <div class="tactical-split-layout">
          <div class="globe-main-panel">
            <Globe3D
              :nodes="data.nodes"
              :selected-id="selectedNode?.id"
              @select="(n) => { selectedNode = n; inspectNode = n }"
            />
          </div>
          <div class="radar-side-panel">
            <RadarScanner3D
              :nodes="data.nodes"
              :active-node="selectedNode || data.nodes[0]"
            />
          </div>
        </div>
      </section>

      <!-- VIEW 3: Full 20+ Brand Unlock Matrix -->
      <section v-else-if="publicView === 'matrix'" class="public-matrix-view">
        <UnlockMatrix
          :nodes="data.nodes"
          :services="data.services"
          :selected-node-id="selectedNode?.id"
          @select-node="(n) => { selectedNode = n; inspectNode = n }"
        />
      </section>

      <!-- VIEW 4: Power Rankings Leaderboard -->
      <section v-else-if="publicView === 'ranking'" class="public-ranking-view">
        <div class="ranking-split-grid">
          <div class="metal-panel">
            <div class="panel-head-metallic">
              <Trophy :size="18" class="text-gold" />
              <div>
                <h2>小鸡综合战力天梯榜</h2>
                <small>基于 IP 纯净度、AI/流媒体全项真实验证及网络延迟综合评定</small>
              </div>
            </div>

            <div v-if="data.rankings && data.rankings.length" class="metal-ranking-table">
              <div
                v-for="item in data.rankings"
                :key="item.node_id"
                class="ranking-row"
                :class="{ 'podium-gold': item.rank === 1, 'podium-silver': item.rank === 2, 'podium-bronze': item.rank === 3 }"
                @click="onSelectNode(data.nodes.find((n) => n.id === item.node_id) || data.nodes[0])"
              >
                <div class="rank-badge-wrap">
                  <span class="rank-pos">{{ item.rank }}</span>
                </div>
                <div class="rank-node-info">
                  <strong>{{ item.name }}</strong>
                  <small>{{ item.provider || 'VPS' }} · {{ item.region || 'GL' }}</small>
                </div>
                <div class="rank-services-count">
                  <span>{{ item.unlocks }} 项核心畅通</span>
                </div>
                <div class="rank-score-capsule" :class="qualityClass(item.risk)">
                  <span>{{ item.quality }} 分</span>
                </div>
              </div>
            </div>
          </div>

          <div class="metal-panel">
            <div class="panel-head-metallic">
              <Zap :size="18" class="text-sky" />
              <div>
                <h2>天梯榜量化算法与口径</h2>
                <small>双引擎轻量级握手 + 权威数据库严谨核验</small>
              </div>
            </div>
            <div class="criteria-list">
              <div class="criteria-item">
                <strong>IP 纯净度评分 (40%)</strong>
                <p>实时调用 Scamalytics、AbuseIPDB 与 IPQS，阻断高欺诈机房 IP。</p>
              </div>
              <div class="criteria-item">
                <strong>AI 生产力大模型解锁 (30%)</strong>
                <p>ChatGPT、Claude 3.5、Gemini、DeepSeek 等 10 款 AI 平台免验证直连。</p>
              </div>
              <div class="criteria-item">
                <strong>流媒体原生 4K/HDR (30%)</strong>
                <p>Netflix 原生非自制全库、Disney+、YouTube Premium、Spotify 音质畅通。</p>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Node Diagnostic Inspection Modal -->
    <Teleport to="body">
      <div v-if="inspectNode" class="modal-backdrop" @click="inspectNode = null">
        <div class="metal-inspect-modal" @click.stop>
          <div class="modal-foil"></div>
          <header class="modal-head">
            <div class="modal-title-wrap">
              <span class="modal-flag">{{ inspectNode.country_code || inspectNode.region }}</span>
              <div>
                <h3>{{ inspectNode.name }}</h3>
                <small>{{ inspectNode.provider }} · {{ inspectNode.organization || 'Direct Network' }}</small>
              </div>
            </div>
            <button class="metal-close-btn" @click="inspectNode = null">
              <X :size="18" />
            </button>
          </header>

          <div class="modal-body">
            <div class="modal-metric-ribbon">
              <div>
                <span>纯净度评分</span>
                <strong :class="qualityClass(inspectNode.risk)">{{ 100 - inspectNode.risk }} <small>分</small></strong>
              </div>
              <div>
                <span>网络协议</span>
                <strong>{{ (inspectNode.families?.length ? inspectNode.families : [inspectNode.family || 4]).map((f) => `IPv${f}`).join(' + ') }}</strong>
              </div>
              <div>
                <span>自治域</span>
                <strong>AS{{ inspectNode.asn }}</strong>
              </div>
              <div>
                <span>脱敏出口</span>
                <code class="modal-ip">{{ inspectNode.masked_ip }}</code>
              </div>
            </div>

            <div class="modal-section-title">
              <Sparkles :size="14" />
              <span>全量 20+ 款品牌 AI 与流媒体解锁矩阵 (含精准延迟)</span>
            </div>

            <div class="modal-badges-grid">
              <!-- AI Category -->
              <div class="badge-category-group">
                <span class="group-label">AI 生产力大模型</span>
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
                <span class="group-label">流媒体与娱乐矩阵</span>
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
        <span>WebGL 3D 硬件加速 · 公开展示已实施末段脱敏保护</span>
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

/* 3D Hub */
.tactical-split-layout {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: 20px;
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
  max-width: 780px;
  max-height: 90vh;
  overflow-y: auto;
  background: linear-gradient(145deg, #161e27, #0c1117);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.8), 0 0 32px rgba(56, 189, 248, 0.2);
  padding: 24px;
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
.modal-title-wrap small {
  font-size: 11px;
  color: #94a3b8;
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
  .tactical-split-layout,
  .ranking-split-grid {
    grid-template-columns: 1fr;
  }
  .credit-cards-grid {
    grid-template-columns: 1fr;
  }
}
</style>
