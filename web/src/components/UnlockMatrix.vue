<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  AlertTriangle,
  Bot,
  CheckCircle2,
  Columns,
  Cpu,
  Filter,
  Grid,
  Layers,
  LayoutGrid,
  Search,
  ShieldAlert,
  ShieldCheck,
  Sparkles,
  Table,
  Tv,
  XCircle,
  Zap,
} from '@lucide/vue'

import type { Node, ServiceCategory, ServiceStat, UnlockInfo } from '../types'
import MetalBadge from './MetalBadge.vue'

const props = defineProps<{
  nodes: Node[]
  services?: ServiceStat[]
  selectedNodeId?: string | null
}>()

const emit = defineEmits<{
  (e: 'selectNode', node: Node): void
  (e: 'inspectUnlock', payload: { node: Node; unlock: UnlockInfo }): void
  (e: 'compareNodes', nodeIds: string[]): void
}>()

const matrixMode = ref<'table' | 'cards'>('table')
const categoryFilter = ref<ServiceCategory | 'all'>('all')
const statusFilter = ref<string>('all')
const searchQuery = ref('')
const selectedForCompare = ref<string[]>([])
const activeTooltip = ref<{ node: Node; unlock: UnlockInfo; x: number; y: number } | null>(null)

// Comprehensive service catalog (20+ items)
const serviceCatalog: { id: string; name: string; category: ServiceCategory; hint: string }[] = [
  // AI Tools (10)
  { id: 'chatgpt', name: 'ChatGPT', category: 'ai', hint: 'OpenAI GPT-4o / Web & API' },
  { id: 'claude', name: 'Claude', category: 'ai', hint: 'Anthropic Claude 3.5 Sonnet' },
  { id: 'gemini', name: 'Gemini', category: 'ai', hint: 'Google Gemini Advanced / AI Studio' },
  { id: 'deepseek', name: 'DeepSeek', category: 'ai', hint: 'DeepSeek R1/V3 推理服务' },
  { id: 'midjourney', name: 'Midjourney', category: 'ai', hint: 'Web / Discord AI 绘画' },
  { id: 'copilot', name: 'Copilot', category: 'ai', hint: 'Microsoft Copilot / Bing AI' },
  { id: 'grok', name: 'Grok', category: 'ai', hint: 'xAI Grok-2 / X 平台' },
  { id: 'perplexity', name: 'Perplexity', category: 'ai', hint: 'Pro 搜索 / 实时推理' },
  { id: 'github_cop', name: 'GitHub Cop', category: 'ai', hint: 'GitHub Copilot IDE 补全' },
  { id: 'reddit', name: 'Reddit', category: 'ai', hint: 'Reddit 社区讨论与搜索' },

  // Streaming Media (11)
  { id: 'netflix', name: 'Netflix', category: 'streaming', hint: '原生 4K / 自制剧判定' },
  { id: 'disney', name: 'Disney+', category: 'streaming', hint: 'Star / IMAX Enhanced' },
  { id: 'youtube', name: 'YouTube Prem', category: 'streaming', hint: '免广告 / YouTube Music' },
  { id: 'prime', name: 'Prime Video', category: 'streaming', hint: 'Amazon 影视全库' },
  { id: 'max', name: 'Max (HBO)', category: 'streaming', hint: '华纳兄弟 / 4K 电影' },
  { id: 'spotify', name: 'Spotify', category: 'streaming', hint: '无损音质 / 动态歌词' },
  { id: 'hulu', name: 'Hulu', category: 'streaming', hint: 'Live TV / 原生美区' },
  { id: 'bahamut', name: '巴哈姆特', category: 'streaming', hint: '台湾动画疯 1080P' },
  { id: 'abema', name: 'AbemaTV', category: 'streaming', hint: '日本全量动漫与直播' },
  { id: 'tiktok', name: 'TikTok', category: 'streaming', hint: '免拔卡视频流' },
  { id: 'dazn', name: 'DAZN', category: 'streaming', hint: '全赛事体育直播' },
]

// Displayed columns based on category
const displayedServices = computed(() => {
  return serviceCatalog.filter((s) => {
    if (categoryFilter.value === 'all') return true
    return s.category === categoryFilter.value
  })
})

// Filtered nodes
const filteredNodes = computed(() => {
  return props.nodes.filter((node) => {
    const q = searchQuery.value.trim().toLowerCase()
    const matchesSearch =
      !q ||
      `${node.name} ${node.provider} ${node.region} ${node.masked_ip} ${node.organization}`.toLowerCase().includes(q)

    if (!matchesSearch) return false
    if (statusFilter.value === 'all') return true

    const allUnlocks = [
      ...Object.values(node.unlocks?.streaming ?? {}),
      ...Object.values(node.unlocks?.ai ?? {}),
    ]
    return allUnlocks.some((u) => u.status === statusFilter.value)
  })
})

// Dynamic summary metrics computed from real nodes
const dynamicSummary = computed(() => {
  const nodes = props.nodes || []
  let totalAI = 0
  let availableAI = 0
  let totalStream = 0
  let availableStream = 0

  const goldenNodes: string[] = []
  const blockedNodes: string[] = []

  for (const n of nodes) {
    let nodeHasBlocked = false

    const streamUnlocks = Object.values(n.unlocks?.streaming ?? {})
    const aiUnlocks = Object.values(n.unlocks?.ai ?? {})

    for (const u of aiUnlocks) {
      if (u.status !== 'untested') {
        totalAI++
        if (u.status === 'available') availableAI++
        else if (u.status === 'blocked') nodeHasBlocked = true
      }
    }
    for (const u of streamUnlocks) {
      if (u.status !== 'untested') {
        totalStream++
        if (u.status === 'available') availableStream++
        else if (u.status === 'blocked') nodeHasBlocked = true
      }
    }

    if (n.status === 'online' && n.risk <= 30 && (n.netflix === 'available' || n.chatgpt === 'available')) {
      goldenNodes.push(n.name)
    }
    if (nodeHasBlocked || n.risk >= 60 || n.status === 'alert') {
      blockedNodes.push(n.name)
    }
  }

  const aiRate = totalAI > 0 ? Math.round((availableAI / totalAI) * 100) : 100
  const streamRate = totalStream > 0 ? Math.round((availableStream / totalStream) * 100) : 100

  return {
    aiRate,
    streamRate,
    goldenNodesText: goldenNodes.slice(0, 3).join(' · ') || (nodes.length > 0 ? nodes[0].name : '暂无节点'),
    blockedNodesText: blockedNodes.slice(0, 3).join(' · ') || '无阻断告警节点',
    hasBlocked: blockedNodes.length > 0,
  }
})

// Helper to get unlock info for a node and service
const getUnlock = (node: Node, serviceId: string): UnlockInfo => {
  const found = node.unlocks?.streaming?.[serviceId] ?? node.unlocks?.ai?.[serviceId]
  if (found) return found

  const meta = serviceCatalog.find((s) => s.id === serviceId)
  return {
    id: serviceId,
    name: meta?.name || serviceId,
    category: meta?.category || 'ai',
    status: 'untested',
    region: '',
    quality: '未检测',
    detail: '当前节点未扫描该项服务',
  }
}

// Compute pass rates per service
const servicePassRates = computed(() => {
  const map: Record<string, { total: number; testedCount: number; available: number; limited: number; blocked: number; rate: number }> = {}
  for (const s of serviceCatalog) {
    let testedCount = 0
    let available = 0
    let limited = 0
    let blocked = 0
    for (const n of props.nodes) {
      const u = getUnlock(n, s.id)
      if (u && u.status !== 'untested' && u.status !== 'unknown') {
        testedCount++
        if (u.status === 'available') available++
        else if (u.status === 'limited') limited++
        else if (u.status === 'blocked') blocked++
      }
    }
    const rate = testedCount > 0 ? Math.round((available / testedCount) * 100) : 0
    map[s.id] = { total: props.nodes.length, testedCount, available, limited, blocked, rate }
  }
  return map
})

const toggleCompareSelect = (id: string) => {
  const idx = selectedForCompare.value.indexOf(id)
  if (idx > -1) {
    selectedForCompare.value.splice(idx, 1)
  } else {
    if (selectedForCompare.value.length >= 4) {
      selectedForCompare.value.shift()
    }
    selectedForCompare.value.push(id)
  }
}

const showCellTooltip = (event: MouseEvent, node: Node, unlock: UnlockInfo) => {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  activeTooltip.value = {
    node,
    unlock,
    x: rect.left + rect.width / 2,
    y: rect.top - 10,
  }
}

const hideCellTooltip = () => {
  activeTooltip.value = null
}
</script>

<template>
  <div class="unlock-matrix-deck">
    <!-- Top Holographic Metrics Bar -->
    <div class="matrix-banner">
      <div class="banner-summary">
        <div class="summary-metric">
          <div class="metric-icon ai"><Bot :size="20" /></div>
          <div>
            <span>AI 大模型解锁率</span>
            <strong class="text-good">{{ dynamicSummary.aiRate }}%</strong>
            <small>ChatGPT / Claude / Gemini / DeepSeek</small>
          </div>
        </div>
        <div class="summary-metric">
          <div class="metric-icon streaming"><Tv :size="20" /></div>
          <div>
            <span>流媒体综合解锁率</span>
            <strong class="text-good">{{ dynamicSummary.streamRate }}%</strong>
            <small>Netflix / Disney+ / YouTube / Max</small>
          </div>
        </div>
        <div class="summary-metric">
          <div class="metric-icon best"><CheckCircle2 :size="20" /></div>
          <div>
            <span>全解锁金牌节点</span>
            <strong>{{ dynamicSummary.goldenNodesText }}</strong>
            <small>低风险 · 连通畅通</small>
          </div>
        </div>
        <div class="summary-metric">
          <div class="metric-icon" :class="dynamicSummary.hasBlocked ? 'warn' : 'best'">
            <AlertTriangle v-if="dynamicSummary.hasBlocked" :size="20" />
            <ShieldCheck v-else :size="20" />
          </div>
          <div>
            <span>封锁阻断节点</span>
            <strong :class="dynamicSummary.hasBlocked ? 'text-danger' : 'text-good'">{{ dynamicSummary.blockedNodesText }}</strong>
            <small>{{ dynamicSummary.hasBlocked ? '需关注封禁与风险' : '全网节点运行稳定' }}</small>
          </div>
        </div>
      </div>
    </div>

    <!-- Filter & View Controls Toolbar -->
    <div class="matrix-toolbar">
      <div class="category-tabs" role="tablist">
        <button
          class="tab-btn"
          :class="{ active: categoryFilter === 'all' }"
          role="tab"
          @click="categoryFilter = 'all'"
        >
          <Layers :size="15" />
          <span>全量服务 ({{ serviceCatalog.length }})</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: categoryFilter === 'ai' }"
          role="tab"
          @click="categoryFilter = 'ai'"
        >
          <Bot :size="15" />
          <span>AI 生产力 (10)</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: categoryFilter === 'streaming' }"
          role="tab"
          @click="categoryFilter = 'streaming'"
        >
          <Tv :size="15" />
          <span>流媒体娱乐 (11)</span>
        </button>
      </div>

      <div class="toolbar-right">
        <!-- Search Box -->
        <div class="search-box">
          <Search :size="15" />
          <input v-model="searchQuery" type="search" placeholder="搜索节点、服务或地区..." aria-label="搜索矩阵节点" />
        </div>

        <!-- Status Filter -->
        <div class="filter-dropdown">
          <Filter :size="15" />
          <select v-model="statusFilter" aria-label="按状态过滤">
            <option value="all">全部状态</option>
            <option value="available">仅可用 (Available)</option>
            <option value="limited">仅受限 (Limited)</option>
            <option value="blocked">仅阻断 (Blocked)</option>
          </select>
        </div>

        <!-- View Mode Switcher -->
        <div class="mode-toggle-group">
          <button
            class="mode-btn"
            :class="{ active: matrixMode === 'table' }"
            title="矩阵大表视图"
            @click="matrixMode = 'table'"
          >
            <Table :size="15" />
          </button>
          <button
            class="mode-btn"
            :class="{ active: matrixMode === 'cards' }"
            title="徽章卡片视图"
            @click="matrixMode = 'cards'"
          >
            <LayoutGrid :size="15" />
          </button>
        </div>

        <!-- Compare Trigger -->
        <button
          class="compare-btn"
          :disabled="selectedForCompare.length < 2"
          @click="emit('compareNodes', selectedForCompare)"
        >
          <Columns :size="15" />
          <span>对比 ({{ selectedForCompare.length }})</span>
        </button>
      </div>
    </div>

    <!-- VIEW 1: Holographic Metallic Matrix Table -->
    <div v-if="matrixMode === 'table'" class="matrix-table-wrap">
      <table class="matrix-table">
        <thead>
          <tr>
            <th class="col-sticky-node">
              <div class="node-th-content">
                <span>节点资产 ({{ filteredNodes.length }})</span>
                <small>点击查看完整诊断</small>
              </div>
            </th>
            <th v-for="service in displayedServices" :key="service.id" class="col-service-th">
              <div class="service-th-content">
                <span class="svc-name">{{ service.name }}</span>
                <div class="pass-bar-wrap" :title="`可用率: ${servicePassRates[service.id]?.rate ?? 0}% (${servicePassRates[service.id]?.available ?? 0}/${servicePassRates[service.id]?.testedCount ?? 0})`">
                  <div
                    class="pass-bar"
                    :style="{
                      width: `${servicePassRates[service.id]?.rate ?? 0}%`,
                      backgroundColor:
                        (servicePassRates[service.id]?.rate ?? 0) >= 80
                          ? '#10b981'
                          : (servicePassRates[service.id]?.rate ?? 0) >= 50
                            ? '#f59e0b'
                            : '#ef4444',
                    }"
                  ></div>
                </div>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="node in filteredNodes" :key="node.id" class="matrix-row" :class="{ selected: selectedNodeId === node.id }">
            <!-- Node Profile Sticky Column -->
            <td class="col-sticky-node" @click="emit('selectNode', node)">
              <div class="node-cell-flex">
                <input
                  type="checkbox"
                  class="compare-checkbox"
                  :checked="selectedForCompare.includes(node.id)"
                  :title="`勾选对比 ${node.name}`"
                  @click.stop="toggleCompareSelect(node.id)"
                />
                <span class="country-pill">{{ node.country_code || node.region }}</span>
                <div class="node-cell-meta">
                  <div class="node-title-row">
                    <strong>{{ node.name }}</strong>
                    <span class="risk-mini-badge" :class="node.risk >= 60 ? 'risk-high' : node.risk >= 35 ? 'risk-mid' : 'risk-low'">
                      {{ node.risk }}
                    </span>
                  </div>
                  <small>{{ node.provider }} · {{ node.region }}</small>
                </div>
              </div>
            </td>

            <!-- Service Cells with MetalBadge Component -->
            <td
              v-for="service in displayedServices"
              :key="service.id"
              class="matrix-cell"
              @click="
                () => {
                  const u = getUnlock(node, service.id)
                  emit('inspectUnlock', { node, unlock: u })
                }
              "
              @mouseenter="(e) => {
                const u = getUnlock(node, service.id)
                showCellTooltip(e, node, u)
              }"
              @mouseleave="hideCellTooltip"
            >
              <MetalBadge
                :service-id="service.id"
                :name="service.name"
                :status="getUnlock(node, service.id).status"
                :region="getUnlock(node, service.id).region"
                :quality="getUnlock(node, service.id).quality"
                :latency-ms="getUnlock(node, service.id).latency_ms"
                size="sm"
                :show-label="false"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- VIEW 2: Metal Badge Cards Grid View -->
    <div v-else class="matrix-cards-deck">
      <div v-for="node in filteredNodes" :key="node.id" class="node-badges-card" @click="emit('selectNode', node)">
        <header class="card-head">
          <div class="card-node-id">
            <span class="card-flag">{{ node.country_code || node.region }}</span>
            <div>
              <strong>{{ node.name }}</strong>
              <small>{{ node.provider }} · AS{{ node.asn }}</small>
            </div>
          </div>
          <div class="card-risk-pill" :class="node.risk >= 60 ? 'risk-high' : 'risk-low'">
            {{ 100 - node.risk }} 分
          </div>
        </header>

        <div class="card-badges-flex">
          <MetalBadge
            v-for="service in displayedServices"
            :key="service.id"
            :service-id="service.id"
            :name="service.name"
            :status="getUnlock(node, service.id).status"
            :region="getUnlock(node, service.id).region"
            :quality="getUnlock(node, service.id).quality"
            :latency-ms="getUnlock(node, service.id).latency_ms"
            size="md"
            @click="
              () => {
                const u = getUnlock(node, service.id)
                emit('inspectUnlock', { node, unlock: u })
              }
            "
          />
        </div>
      </div>
    </div>

    <!-- Hover Detail Popover -->
    <Teleport to="body">
      <div
        v-if="activeTooltip"
        class="matrix-popover"
        :style="{ left: `${activeTooltip.x}px`, top: `${activeTooltip.y}px` }"
      >
        <div class="popover-head">
          <strong>{{ activeTooltip.unlock.name }}</strong>
          <span class="popover-status-badge" :class="activeTooltip.unlock.status">
            {{ activeTooltip.unlock.status === 'available' ? '解锁' : activeTooltip.unlock.status === 'limited' ? '受限' : activeTooltip.unlock.status === 'blocked' ? '封锁' : '未检测' }}
          </span>
        </div>
        <div class="popover-body">
          <div class="popover-row">
            <span>节点</span>
            <strong>{{ activeTooltip.node.name }} ({{ activeTooltip.node.country_code }})</strong>
          </div>
          <div v-if="activeTooltip.unlock.latency_ms" class="popover-row">
            <span>探测延迟</span>
            <strong class="text-sky">{{ activeTooltip.unlock.latency_ms }} ms</strong>
          </div>
          <div v-if="activeTooltip.unlock.quality" class="popover-row">
            <span>质量等级</span>
            <span>{{ activeTooltip.unlock.quality }}</span>
          </div>
          <div v-if="activeTooltip.unlock.detail" class="popover-desc">
            {{ activeTooltip.unlock.detail }}
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.unlock-matrix-deck {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 14px;
}

/* Banner */
.matrix-banner {
  background: linear-gradient(145deg, #182029 0%, #0e141b 100%);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.banner-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.summary-metric {
  display: flex;
  align-items: center;
  gap: 12px;
}

.metric-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.metric-icon.ai { background: rgba(168, 85, 247, 0.15); color: #c084fc; }
.metric-icon.streaming { background: rgba(56, 189, 248, 0.15); color: #38bdf8; }
.metric-icon.best { background: rgba(16, 185, 129, 0.15); color: #34d399; }
.metric-icon.warn { background: rgba(239, 68, 68, 0.15); color: #f87171; }

.summary-metric span {
  display: block;
  font-size: 12px;
  color: #a8b4c2;
}
.summary-metric strong {
  display: block;
  font-family: 'Fira Code', monospace;
  font-size: 16px;
  font-weight: 700;
  color: #f8fafc;
  margin: 1px 0;
}
.summary-metric small {
  display: block;
  font-size: 11px;
  color: #8794a6;
  line-height: 1.4;
}

.text-good { color: #10b981 !important; }
.text-danger { color: #ef4444 !important; }
.text-sky { color: #38bdf8 !important; }

/* Toolbar */
.matrix-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.category-tabs {
  display: flex;
  background: rgba(0, 0, 0, 0.35);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 3px;
  gap: 3px;
}
.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: transparent;
  border: 0;
  border-radius: 6px;
  color: #94a3b8;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}
.tab-btn.active {
  background: #1e293b;
  color: #38bdf8;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(0, 0, 0, 0.35);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 6px 10px;
  color: #94a3b8;
}
.search-box input {
  background: transparent;
  border: 0;
  color: #f8fafc;
  font-size: 12px;
  outline: none;
  width: 170px;
}

.filter-dropdown {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(0, 0, 0, 0.35);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 6px 10px;
  color: #94a3b8;
}
.filter-dropdown select {
  background: transparent;
  border: 0;
  color: #f8fafc;
  font-size: 12px;
  outline: none;
  cursor: pointer;
}

.mode-toggle-group {
  display: flex;
  background: rgba(0, 0, 0, 0.35);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 2px;
}
.mode-btn {
  padding: 6px 8px;
  background: transparent;
  border: 0;
  border-radius: 6px;
  color: #94a3b8;
  cursor: pointer;
}
.mode-btn.active {
  background: #1e293b;
  color: #38bdf8;
}

.compare-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #94a3b8;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}
.compare-btn:not(:disabled) {
  background: #0284c7;
  color: #fff;
  border-color: #38bdf8;
}

/* Table View */
.matrix-table-wrap {
  width: 100%;
  max-width: 100%;
  overflow-x: auto;
  overscroll-behavior-inline: contain;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  background: #0c1117;
}

.matrix-table {
  width: 100%;
  min-width: max-content;
  border-collapse: collapse;
}

.matrix-table th {
  padding: 10px 8px;
  background: #131922;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  border-right: 1px solid rgba(255, 255, 255, 0.04);
  text-align: center;
  font-size: 12px;
  color: #cbd5e1;
}

.col-sticky-node {
  position: sticky;
  left: 0;
  z-index: 10;
  background: #131922;
  min-width: 236px;
  text-align: left !important;
}

.matrix-row {
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  transition: background 0.15s ease;
}
.matrix-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.matrix-row td {
  padding: 9px 7px;
  text-align: center;
  vertical-align: middle;
  border-right: 1px solid rgba(255, 255, 255, 0.03);
}
.col-service-th,
.matrix-cell {
  min-width: 48px;
}
.node-th-content,
.service-th-content {
  line-height: 1.35;
}
.node-th-content small,
.node-cell-meta small {
  display: block;
  margin-top: 2px;
  color: #8794a6;
  font-size: 11px;
}
.svc-name {
  display: block;
  max-width: 76px;
  overflow: hidden;
  color: #dbe5ef;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-cell-flex {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  cursor: pointer;
}
.country-pill {
  font-family: 'Fira Code', monospace;
  font-size: 11px;
  font-weight: 800;
  padding: 2px 5px;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  border-radius: 4px;
}
.node-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}
.node-title-row strong {
  font-size: 13px;
  font-weight: 650;
  color: #f8fafc;
}
.risk-mini-badge {
  font-family: 'Fira Code', monospace;
  font-size: 10.5px;
  font-weight: 800;
  padding: 0 4px;
  border-radius: 3px;
}
.risk-low { color: #10b981; background: rgba(16, 185, 129, 0.15); }
.risk-mid { color: #f59e0b; background: rgba(245, 158, 11, 0.15); }
.risk-high { color: #ef4444; background: rgba(239, 68, 68, 0.15); }

.pass-bar-wrap {
  width: 44px;
  height: 3px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  margin: 3px auto 0;
  overflow: hidden;
}
.pass-bar {
  height: 100%;
  border-radius: 2px;
}

/* Cards View */
.matrix-cards-deck {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(460px, 1fr));
  gap: 16px;
}
.node-badges-card {
  background: linear-gradient(145deg, #161e27, #0c1117);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 14px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  cursor: pointer;
  transition: transform 0.15s ease, border-color 0.15s ease;
}
.node-badges-card:hover {
  transform: translateY(-2px);
  border-color: rgba(56, 189, 248, 0.35);
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  padding-bottom: 8px;
}
.card-node-id {
  display: flex;
  align-items: center;
  gap: 8px;
}
.card-flag {
  font-family: 'Fira Code', monospace;
  font-size: 11px;
  font-weight: 800;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  padding: 2px 6px;
  border-radius: 4px;
}
.card-node-id strong {
  display: block;
  font-size: 13px;
  color: #f8fafc;
}
.card-node-id small {
  font-size: 10.5px;
  color: #64748b;
}

.card-risk-pill {
  font-family: 'Fira Code', monospace;
  font-size: 11px;
  font-weight: 800;
  padding: 2px 8px;
  border-radius: 12px;
}

.card-badges-flex {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  overflow: hidden;
}

/* Hover Popover */
.matrix-popover {
  position: fixed;
  z-index: 1000;
  transform: translate(-50%, -100%);
  width: 248px;
  padding: 12px 14px;
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(56, 189, 248, 0.35);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  pointer-events: none;
  font-size: 13px;
  line-height: 1.45;
}
.popover-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: 6px;
  margin-bottom: 6px;
}
.popover-status-badge {
  font-size: 11px;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 4px;
}
.popover-status-badge.available { color: #10b981; background: rgba(16, 185, 129, 0.2); }
.popover-status-badge.limited { color: #f59e0b; background: rgba(245, 158, 11, 0.2); }
.popover-status-badge.blocked { color: #ef4444; background: rgba(239, 68, 68, 0.2); }

.popover-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.popover-row span { color: #b0bbc8; }
.popover-desc {
  font-size: 12px;
  color: #98a6b6;
  margin-top: 6px;
  border-top: 1px dashed rgba(255, 255, 255, 0.08);
  padding-top: 6px;
}

@media (max-width: 1440px) {
  .banner-summary { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .summary-metric { min-height: 48px; }
  .matrix-toolbar { align-items: flex-start; }
  .matrix-table th { padding-block: 11px; }
  .matrix-row td { padding-block: 10px; }
}

@media (max-width: 720px) {
  .banner-summary { grid-template-columns: 1fr; gap: 12px; }
  .matrix-banner { padding: 14px; }
  .toolbar-right, .search-box { width: 100%; }
  .search-box input { width: 100%; min-width: 0; }
  .matrix-cards-deck { grid-template-columns: minmax(0, 1fr); }
}
</style>
