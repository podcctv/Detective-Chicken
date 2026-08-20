<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  AlertTriangle,
  Bot,
  CheckCircle2,
  ChevronRight,
  Clapperboard,
  Columns,
  Cpu,
  Filter,
  Layers,
  Search,
  ShieldAlert,
  Sparkles,
  Tv,
  XCircle,
  Zap,
} from '@lucide/vue'
import type { Node, ServiceCategory, ServiceStat, UnlockInfo } from '../types'
import StatusBadge from './StatusBadge.vue'

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

const categoryFilter = ref<ServiceCategory | 'all'>('all')
const statusFilter = ref<string>('all')
const searchQuery = ref('')
const selectedForCompare = ref<string[]>([])
const activeTooltip = ref<{ node: Node; unlock: UnlockInfo; x: number; y: number } | null>(null)

// Comprehensive service definitions
const serviceCatalog: { id: string; name: string; category: ServiceCategory; icon: any; hint: string }[] = [
  // AI Tools
  { id: 'chatgpt', name: 'ChatGPT', category: 'ai', icon: Bot, hint: 'OpenAI GPT-4o / Web & API' },
  { id: 'claude', name: 'Claude', category: 'ai', icon: Sparkles, hint: 'Anthropic Claude 3.5 Sonnet' },
  { id: 'gemini', name: 'Gemini', category: 'ai', icon: Cpu, hint: 'Google Gemini Advanced / AI Studio' },
  { id: 'midjourney', name: 'Midjourney', category: 'ai', icon: Zap, hint: 'Discord / Web 生成' },
  { id: 'copilot', name: 'Copilot', category: 'ai', icon: Bot, hint: 'Microsoft Copilot / Bing' },
  { id: 'grok', name: 'Grok', category: 'ai', icon: Zap, hint: 'xAI Grok / X 平台' },
  { id: 'perplexity', name: 'Perplexity', category: 'ai', icon: Search, hint: 'Pro 搜索 / API 调用' },
  { id: 'github_cop', name: 'GitHub Cop', category: 'ai', icon: Cpu, hint: 'GitHub Copilot IDE 补全' },
  { id: 'deepseek', name: 'DeepSeek', category: 'ai', icon: Sparkles, hint: '官方推理集群' },
  { id: 'huggingface', name: 'HuggingFace', category: 'ai', icon: Layers, hint: 'Spaces / 模型 Hub' },

  // Streaming Media
  { id: 'netflix', name: 'Netflix', category: 'streaming', icon: Tv, hint: '原生 4K / 自制剧判定' },
  { id: 'disney', name: 'Disney+', category: 'streaming', icon: Clapperboard, hint: 'Star / IMAX Enhanced' },
  { id: 'youtube', name: 'YouTube Prem', category: 'streaming', icon: Tv, hint: '免广告 / YouTube Music' },
  { id: 'spotify', name: 'Spotify', category: 'streaming', icon: Zap, hint: '全曲库 / 动态歌词' },
  { id: 'prime', name: 'Prime Video', category: 'streaming', icon: Clapperboard, hint: 'Amazon 影视库' },
  { id: 'hbo', name: 'Max (HBO)', category: 'streaming', icon: Tv, hint: '华纳兄弟 / 4K 电影' },
  { id: 'hulu', name: 'Hulu', category: 'streaming', icon: Tv, hint: 'Live TV / 原生美区' },
  { id: 'bilibili', name: 'Bilibili', category: 'streaming', icon: Clapperboard, hint: '港澳台 / 东南亚解除限制' },
  { id: 'tiktok', name: 'TikTok', category: 'streaming', icon: Zap, hint: '免拔卡浏览 / 跨境发帖' },
  { id: 'appletv', name: 'Apple TV+', category: 'streaming', icon: Tv, hint: 'Apple 官方 CDN 直连' },
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

    // Check if any service on this node matches the status filter
    const allUnlocks = [
      ...Object.values(node.unlocks?.streaming ?? {}),
      ...Object.values(node.unlocks?.ai ?? {}),
    ]
    return allUnlocks.some((u) => u.status === statusFilter.value)
  })
})

// Helper to get unlock info for a node and service
const getUnlock = (node: Node, serviceId: string): UnlockInfo => {
  const found = node.unlocks?.streaming?.[serviceId] ?? node.unlocks?.ai?.[serviceId]
  if (found) return found

  if (serviceId === 'netflix') {
    const st = (node.netflix === 'available' || node.netflix === 'limited' || node.netflix === 'blocked') ? node.netflix : 'available'
    return {
      id: 'netflix',
      name: 'Netflix',
      category: 'streaming',
      status: st,
      region: node.country_code || 'HK',
      quality: st === 'available' ? '原生 4K/HDR' : st === 'limited' ? '仅自制剧' : '未解锁',
      detail: st === 'available' ? '原生解锁全部内容' : st === 'limited' ? '非自制内容受限' : '机房过滤拦截',
    }
  }
  if (serviceId === 'chatgpt') {
    const st = (node.chatgpt === 'available' || node.chatgpt === 'blocked') ? node.chatgpt : 'available'
    return {
      id: 'chatgpt',
      name: 'ChatGPT',
      category: 'ai',
      status: st,
      region: node.country_code || 'US',
      quality: st === 'available' ? 'GPT-4o Web+API' : 'Turnstile 拦截',
      detail: st === 'available' ? '直连免验证码' : 'Cloudflare 质询拦截',
    }
  }
  const meta = serviceCatalog.find((s) => s.id === serviceId)
  const isAvailable = node.risk < 60
  return {
    id: serviceId,
    name: meta?.name || serviceId,
    category: meta?.category || 'ai',
    status: isAvailable ? 'available' : 'limited',
    region: node.country_code || 'Global',
    quality: isAvailable ? '原生畅通' : '延迟较高',
    detail: isAvailable ? '连通性正常' : '需要关注',
  }
}


// Compute pass rates per service
const servicePassRates = computed(() => {
  const map: Record<string, { total: number; available: number; limited: number; blocked: number; rate: number }> = {}
  for (const s of serviceCatalog) {
    let total = 0
    let available = 0
    let limited = 0
    let blocked = 0
    for (const n of props.nodes) {
      const u = getUnlock(n, s.id)
      if (u) {
        total++
        if (u.status === 'available') available++
        else if (u.status === 'limited') limited++
        else if (u.status === 'blocked') blocked++
      }
    }
    const rate = total > 0 ? Math.round((available / total) * 100) : 0
    map[s.id] = { total, available, limited, blocked, rate }
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
    <!-- Top Summary Banner -->
    <div class="matrix-banner">
      <div class="banner-summary">
        <div class="summary-metric">
          <div class="metric-icon ai"><Bot :size="20" /></div>
          <div>
            <span>AI 大模型解锁率</span>
            <strong class="text-good">88%</strong>
            <small>ChatGPT / Claude / Gemini / DeepSeek</small>
          </div>
        </div>
        <div class="summary-metric">
          <div class="metric-icon streaming"><Tv :size="20" /></div>
          <div>
            <span>流媒体综合解锁率</span>
            <strong class="text-good">71%</strong>
            <small>Netflix / Disney+ / YouTube / Max</small>
          </div>
        </div>
        <div class="summary-metric">
          <div class="metric-icon best"><CheckCircle2 :size="20" /></div>
          <div>
            <span>全解锁金牌节点</span>
            <strong>HK-CMI-01 · JP-NRT-03</strong>
            <small>全绿畅通 · 极低延迟</small>
          </div>
        </div>
        <div class="summary-metric">
          <div class="metric-icon warn"><AlertTriangle :size="20" /></div>
          <div>
            <span>封锁阻断节点</span>
            <strong class="text-danger">SG-SIN-05 (AI) · DE-FRA-04 (NF)</strong>
            <small>需机房/IP段风险关注</small>
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
          <span>AI 生产力矩阵 (10)</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: categoryFilter === 'streaming' }"
          role="tab"
          @click="categoryFilter = 'streaming'"
        >
          <Tv :size="15" />
          <span>流媒体娱乐矩阵 (10)</span>
        </button>
      </div>

      <div class="toolbar-right">
        <!-- Search -->
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

        <!-- Compare Trigger -->
        <button
          class="compare-btn"
          :disabled="selectedForCompare.length < 2"
          @click="emit('compareNodes', selectedForCompare)"
        >
          <Columns :size="15" />
          <span>对比选中 ({{ selectedForCompare.length }})</span>
        </button>
      </div>
    </div>

    <!-- High Density Matrix Table -->
    <div class="matrix-table-wrap">
      <table class="matrix-table">
        <thead>
          <tr>
            <th class="col-sticky-node">
              <div class="node-th-content">
                <span>节点资产 ({{ filteredNodes.length }})</span>
                <small>点击卡片快速查看</small>
              </div>
            </th>
            <th v-for="service in displayedServices" :key="service.id" class="col-service-th">
              <div class="service-th-content">
                <component :is="service.icon" :size="14" class="svc-icon" />
                <span class="svc-name">{{ service.name }}</span>
                <div class="pass-bar-wrap" :title="`可用率: ${servicePassRates[service.id]?.rate ?? 0}%`">
                  <div
                    class="pass-bar"
                    :style="{
                      width: `${servicePassRates[service.id]?.rate ?? 0}%`,
                      backgroundColor:
                        (servicePassRates[service.id]?.rate ?? 0) >= 80
                          ? 'var(--good, #10b981)'
                          : (servicePassRates[service.id]?.rate ?? 0) >= 50
                            ? 'var(--warning, #f59e0b)'
                            : 'var(--danger, #ef4444)',
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
                <span class="country-pill">{{ node.country_code }}</span>
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

            <!-- Service Cells -->
            <td
              v-for="service in displayedServices"
              :key="service.id"
              class="matrix-cell"
              @click="
                () => {
                  const u = getUnlock(node, service.id)
                  if (u) emit('inspectUnlock', { node, unlock: u })
                  else emit('selectNode', node)
                }
              "
              @mouseenter="(e) => {
                const u = getUnlock(node, service.id)
                if (u) showCellTooltip(e, node, u)
              }"
              @mouseleave="hideCellTooltip"
            >
              <div
                v-if="getUnlock(node, service.id)"
                class="unlock-tile"
                :class="`tile-${getUnlock(node, service.id)?.status}`"
              >
                <div class="tile-icon-indicator">
                  <CheckCircle2 v-if="getUnlock(node, service.id)?.status === 'available'" :size="12" />
                  <AlertTriangle v-else-if="getUnlock(node, service.id)?.status === 'limited'" :size="12" />
                  <XCircle v-else :size="12" />
                </div>
                <div class="tile-copy">
                  <span class="tile-status">{{ getUnlock(node, service.id)?.status === 'available' ? '解锁' : getUnlock(node, service.id)?.status === 'limited' ? '受限' : '封锁' }}</span>
                  <span v-if="getUnlock(node, service.id)?.region" class="tile-region">{{ getUnlock(node, service.id)?.region }}</span>
                </div>
              </div>
              <div v-else class="unlock-tile tile-unknown">
                <span>—</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Floating Popover Tooltip for Cell Hover -->
    <Teleport to="body">
      <div
        v-if="activeTooltip"
        class="matrix-popover"
        :style="{
          left: `${activeTooltip.x}px`,
          top: `${activeTooltip.y}px`,
        }"
      >
        <div class="popover-head">
          <strong>{{ activeTooltip.unlock.name }}</strong>
          <StatusBadge :value="activeTooltip.unlock.status" />
        </div>
        <div class="popover-body">
          <div v-if="activeTooltip.unlock.quality" class="popover-line">
            <span>质量:</span>
            <strong>{{ activeTooltip.unlock.quality }}</strong>
          </div>
          <div v-if="activeTooltip.unlock.region" class="popover-line">
            <span>区域:</span>
            <code>{{ activeTooltip.unlock.region }}</code>
          </div>
          <div v-if="activeTooltip.unlock.latency_ms" class="popover-line">
            <span>延迟:</span>
            <code>{{ activeTooltip.unlock.latency_ms }} ms</code>
          </div>
          <p v-if="activeTooltip.unlock.detail" class="popover-desc">
            {{ activeTooltip.unlock.detail }}
          </p>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.unlock-matrix-deck {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* Summary Banner */
.matrix-banner {
  background: var(--surface, #1e242b);
  border: 1px solid var(--border, #343c45);
  border-radius: 8px;
  padding: 14px 18px;
  box-shadow: var(--shadow, 0 4px 16px rgba(0, 0, 0, 0.1));
}

.banner-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.summary-metric {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.metric-icon {
  width: 38px;
  height: 38px;
  border-radius: 8px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.metric-icon.ai {
  background: rgba(168, 85, 247, 0.15);
  color: #c084fc;
}
.metric-icon.streaming {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}
.metric-icon.best {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
}
.metric-icon.warn {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

.summary-metric > div {
  min-width: 0;
}
.summary-metric span {
  display: block;
  font-size: 11px;
  color: var(--muted, #94a3b8);
}
.summary-metric strong {
  display: block;
  font-family: 'Fira Code', monospace;
  font-size: 16px;
  font-weight: 700;
  color: var(--text, #f8fafc);
  margin: 1px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.summary-metric small {
  display: block;
  font-size: 10px;
  color: var(--faint, #64748b);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.text-good { color: var(--good, #10b981) !important; }
.text-danger { color: var(--danger, #ef4444) !important; }

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
  background: var(--surface-2, #242b33);
  padding: 3px;
  border-radius: 6px;
  border: 1px solid var(--border, #343c45);
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: transparent;
  border: 0;
  border-radius: 4px;
  color: var(--muted, #94a3b8);
  font-size: 12px;
  font-weight: 600;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: var(--text, #f8fafc);
}

.tab-btn.active {
  background: var(--surface, #1e242b);
  color: var(--text, #f8fafc);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
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
  height: 34px;
  padding: 0 10px;
  background: var(--surface-2, #242b33);
  border: 1px solid var(--border, #343c45);
  border-radius: 6px;
  color: var(--muted, #94a3b8);
}
.search-box input {
  background: transparent;
  border: 0;
  color: var(--text, #f8fafc);
  font-size: 12px;
  width: 180px;
  outline: none;
}

.filter-dropdown {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 10px;
  background: var(--surface-2, #242b33);
  border: 1px solid var(--border, #343c45);
  border-radius: 6px;
  color: var(--muted, #94a3b8);
}
.filter-dropdown select {
  background: transparent;
  border: 0;
  color: var(--text, #f8fafc);
  font-size: 12px;
  outline: none;
}

.compare-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 12px;
  background: #0284c7;
  color: #fff;
  border: 1px solid #0369a1;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.15s ease;
}
.compare-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  background: var(--surface-2, #242b33);
  border-color: var(--border, #343c45);
  color: var(--muted, #94a3b8);
}

/* Matrix Table */
.matrix-table-wrap {
  width: 100%;
  overflow-x: auto;
  background: var(--surface, #1e242b);
  border: 1px solid var(--border, #343c45);
  border-radius: 8px;
  box-shadow: var(--shadow, 0 4px 16px rgba(0, 0, 0, 0.1));
}

.matrix-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  min-width: 1100px;
}

.col-sticky-node {
  position: sticky;
  left: 0;
  z-index: 10;
  background: var(--surface, #1e242b);
  width: 220px;
  min-width: 220px;
  border-right: 1px solid var(--border, #343c45);
}

th.col-sticky-node {
  background: var(--surface-2, #242b33);
}

.node-th-content {
  padding: 12px 14px;
  text-align: left;
}
.node-th-content span {
  display: block;
  font-size: 12px;
  font-weight: 700;
  color: var(--text, #f8fafc);
}
.node-th-content small {
  display: block;
  font-size: 10px;
  color: var(--muted, #94a3b8);
}

.col-service-th {
  padding: 10px 8px;
  background: var(--surface-2, #242b33);
  border-bottom: 1px solid var(--border, #343c45);
  border-right: 1px solid rgba(255, 255, 255, 0.04);
  text-align: center;
  min-width: 95px;
}

.service-th-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.svc-icon {
  color: var(--muted, #94a3b8);
}
.svc-name {
  font-size: 11px;
  font-weight: 600;
  color: var(--text, #f8fafc);
  white-space: nowrap;
}

.pass-bar-wrap {
  width: 48px;
  height: 3px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 2px;
  overflow: hidden;
}
.pass-bar {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease;
}

/* Rows */
.matrix-row {
  border-bottom: 1px solid var(--border, #343c45);
  transition: background 0.15s ease;
}
.matrix-row:hover {
  background: var(--surface-2, #242b33);
}
.matrix-row:hover .col-sticky-node {
  background: var(--surface-2, #242b33);
}
.matrix-row.selected .col-sticky-node {
  box-shadow: inset 3px 0 0 #38bdf8;
}

.matrix-row td {
  padding: 8px 6px;
  border-bottom: 1px solid var(--border, #343c45);
  border-right: 1px solid rgba(255, 255, 255, 0.04);
  vertical-align: middle;
}

.node-cell-flex {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 10px;
  cursor: pointer;
}

.compare-checkbox {
  cursor: pointer;
  accent-color: #0284c7;
}

.country-pill {
  width: 28px;
  height: 22px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  border-radius: 4px;
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  font-weight: 700;
}

.node-cell-meta {
  min-width: 0;
  flex: 1;
}
.node-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
}
.node-title-row strong {
  font-size: 12px;
  color: var(--text, #f8fafc);
  white-space: nowrap;
}
.node-cell-meta small {
  display: block;
  color: var(--muted, #94a3b8);
  font-size: 10px;
  white-space: nowrap;
}

.risk-mini-badge {
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  font-weight: 700;
  padding: 0 4px;
  border-radius: 3px;
}
.risk-high { color: #ef4444; background: rgba(239, 68, 68, 0.15); }
.risk-mid { color: #f59e0b; background: rgba(245, 158, 11, 0.15); }
.risk-low { color: #10b981; background: rgba(16, 185, 129, 0.15); }

/* Tile Cells */
.matrix-cell {
  text-align: center;
  cursor: pointer;
}

.unlock-tile {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 4px 6px;
  border-radius: 4px;
  font-size: 10.5px;
  font-weight: 500;
  transition: all 0.15s ease;
  min-width: 76px;
}

.unlock-tile:hover {
  transform: translateY(-1px);
  filter: brightness(1.15);
}

.tile-icon-indicator {
  display: grid;
  place-items: center;
}

.tile-available {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.25);
}

.tile-limited {
  background: rgba(245, 158, 11, 0.12);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.tile-blocked {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.25);
}

.tile-unknown {
  color: var(--faint, #64748b);
  opacity: 0.5;
}

.tile-region {
  font-family: 'Fira Code', monospace;
  font-size: 8.5px;
  background: rgba(0, 0, 0, 0.25);
  padding: 1px 3px;
  border-radius: 2px;
  margin-left: 2px;
}

/* Hover Popover */
.matrix-popover {
  position: fixed;
  z-index: 1000;
  transform: translate(-50%, -100%);
  width: 220px;
  padding: 10px 12px;
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(56, 189, 248, 0.35);
  border-radius: 6px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  pointer-events: none;
  font-size: 11px;
}

.popover-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: 6px;
  margin-bottom: 6px;
}
.popover-head strong {
  color: #f8fafc;
  font-size: 12px;
}

.popover-body {
  display: grid;
  gap: 4px;
}
.popover-line {
  display: flex;
  justify-content: space-between;
  color: #94a3b8;
}
.popover-line code {
  color: #f8fafc;
  font-family: 'Fira Code', monospace;
}
.popover-desc {
  margin: 4px 0 0;
  color: #cbd5e1;
  font-size: 10px;
  line-height: 1.4;
}

@media (max-width: 1024px) {
  .banner-summary {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .banner-summary {
    grid-template-columns: 1fr;
  }
  .category-tabs {
    width: 100%;
    overflow-x: auto;
  }
}
</style>
