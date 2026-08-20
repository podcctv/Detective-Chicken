<script setup lang="ts">
import { computed, ref } from 'vue'
import { Bot, CheckCircle2, Columns, Tv, X, XCircle } from '@lucide/vue'
import type { Node, ServiceCategory } from '../types'
import StatusBadge from './StatusBadge.vue'

const props = defineProps<{
  nodes: Node[]
  compareIds: string[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'removeNode', id: string): void
}>()

const activeTab = ref<ServiceCategory | 'all'>('all')

const comparedNodes = computed(() => {
  return props.nodes.filter((n) => props.compareIds.includes(n.id))
})

const serviceList: { id: string; name: string; category: ServiceCategory }[] = [
  // AI
  { id: 'chatgpt', name: 'ChatGPT / OpenAI', category: 'ai' },
  { id: 'claude', name: 'Claude (Anthropic)', category: 'ai' },
  { id: 'gemini', name: 'Google Gemini', category: 'ai' },
  { id: 'midjourney', name: 'Midjourney', category: 'ai' },
  { id: 'copilot', name: 'Microsoft Copilot', category: 'ai' },
  { id: 'grok', name: 'xAI Grok', category: 'ai' },
  { id: 'perplexity', name: 'Perplexity AI', category: 'ai' },
  { id: 'github_cop', name: 'GitHub Copilot', category: 'ai' },
  { id: 'deepseek', name: 'DeepSeek', category: 'ai' },
  { id: 'huggingface', name: 'HuggingFace', category: 'ai' },

  // Streaming
  { id: 'netflix', name: 'Netflix', category: 'streaming' },
  { id: 'disney', name: 'Disney+', category: 'streaming' },
  { id: 'youtube', name: 'YouTube Premium', category: 'streaming' },
  { id: 'spotify', name: 'Spotify', category: 'streaming' },
  { id: 'prime', name: 'Prime Video', category: 'streaming' },
  { id: 'hbo', name: 'Max (HBO)', category: 'streaming' },
  { id: 'hulu', name: 'Hulu', category: 'streaming' },
  { id: 'bilibili', name: 'Bilibili (港澳台)', category: 'streaming' },
  { id: 'tiktok', name: 'TikTok', category: 'streaming' },
  { id: 'appletv', name: 'Apple TV+', category: 'streaming' },
]

const filteredServices = computed(() => {
  if (activeTab.value === 'all') return serviceList
  return serviceList.filter((s) => s.category === activeTab.value)
})

const getUnlock = (node: Node, serviceId: string) => {
  return node.unlocks?.streaming?.[serviceId] ?? node.unlocks?.ai?.[serviceId] ?? null
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <div class="compare-modal-card" role="dialog" aria-modal="true" aria-label="多节点横向解锁对比">
      <div class="compare-head">
        <div class="head-title">
          <Columns :size="18" class="text-info" />
          <div>
            <h2>节点横向对比分析</h2>
            <p>已选择 {{ comparedNodes.length }} 个节点进行 AI 工具与流媒体解锁能力全景横向核验</p>
          </div>
        </div>
        <button class="icon-btn" aria-label="关闭" @click="emit('close')">
          <X :size="18" />
        </button>
      </div>

      <!-- Tab Switcher -->
      <div class="compare-tabs">
        <button
          class="tab-pill"
          :class="{ active: activeTab === 'all' }"
          @click="activeTab = 'all'"
        >
          全量服务 ({{ serviceList.length }})
        </button>
        <button
          class="tab-pill"
          :class="{ active: activeTab === 'ai' }"
          @click="activeTab = 'ai'"
        >
          <Bot :size="14" />
          AI 生产力工具 (10)
        </button>
        <button
          class="tab-pill"
          :class="{ active: activeTab === 'streaming' }"
          @click="activeTab = 'streaming'"
        >
          <Tv :size="14" />
          全球流媒体 (10)
        </button>
      </div>

      <!-- Comparison Content Grid -->
      <div class="compare-content-scroll">
        <table class="compare-table">
          <thead>
            <tr>
              <th class="feature-col">评估维度 / 服务</th>
              <th v-for="node in comparedNodes" :key="node.id" class="node-col">
                <div class="node-col-head">
                  <div class="col-head-info">
                    <span class="country-tag">{{ node.country_code }}</span>
                    <strong>{{ node.name }}</strong>
                  </div>
                  <button
                    class="remove-btn"
                    title="移除对比"
                    aria-label="移除对比"
                    @click="emit('removeNode', node.id)"
                  >
                    <X :size="13" />
                  </button>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <!-- General Info Rows -->
            <tr class="section-divider">
              <td :colspan="comparedNodes.length + 1">网络基本特征与安全评级</td>
            </tr>
            <tr>
              <td class="feature-col">综合风险评分</td>
              <td v-for="node in comparedNodes" :key="node.id" class="node-col">
                <strong
                  class="risk-score"
                  :class="node.risk >= 60 ? 'risk-high' : node.risk >= 35 ? 'risk-mid' : 'risk-low'"
                >
                  {{ node.risk }} 分
                </strong>
              </td>
            </tr>
            <tr>
              <td class="feature-col">服务商 / 归属 ASN</td>
              <td v-for="node in comparedNodes" :key="node.id" class="node-col">
                <span>{{ node.provider }}</span>
                <code class="d-block">AS{{ node.asn }}</code>
              </td>
            </tr>
            <tr>
              <td class="feature-col">脱敏 IP / 协议族</td>
              <td v-for="node in comparedNodes" :key="node.id" class="node-col">
                <code>{{ node.masked_ip }}</code>
                <small class="d-block">IPv{{ node.family }}</small>
              </td>
            </tr>
            <tr>
              <td class="feature-col">DNSBL 黑名单命中</td>
              <td v-for="node in comparedNodes" :key="node.id" class="node-col">
                <span class="dnsbl-tag" :class="{ hit: node.dnsbl > 0 }">
                  {{ node.dnsbl }} 处命中
                </span>
              </td>
            </tr>

            <!-- Service Unlocks Rows -->
            <tr class="section-divider">
              <td :colspan="comparedNodes.length + 1">AI 工具与流媒体解锁能力</td>
            </tr>
            <tr v-for="svc in filteredServices" :key="svc.id">
              <td class="feature-col">
                <div class="svc-name-row">
                  <Bot v-if="svc.category === 'ai'" :size="14" class="text-muted" />
                  <Tv v-else :size="14" class="text-muted" />
                  <span>{{ svc.name }}</span>
                </div>
              </td>
              <td v-for="node in comparedNodes" :key="node.id" class="node-col">
                <div v-if="getUnlock(node, svc.id)" class="unlock-compare-cell">
                  <StatusBadge :value="getUnlock(node, svc.id)!.status" />
                  <small v-if="getUnlock(node, svc.id)?.quality" class="quality-text">
                    {{ getUnlock(node, svc.id)?.quality }}
                  </small>
                  <code v-if="getUnlock(node, svc.id)?.latency_ms" class="latency-text">
                    {{ getUnlock(node, svc.id)?.latency_ms }}ms
                  </code>
                </div>
                <span v-else class="text-muted">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="compare-foot">
        <button class="secondary-btn" @click="emit('close')">完成并返回</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 120;
  background: rgba(10, 14, 20, 0.7);
  backdrop-filter: blur(6px);
  display: grid;
  place-items: center;
  padding: 18px;
}

.compare-modal-card {
  width: min(940px, 100%);
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  background: var(--surface, #1e242b);
  border: 1px solid var(--border, #343c45);
  border-radius: 10px;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

.compare-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border, #343c45);
}
.head-title {
  display: flex;
  align-items: center;
  gap: 12px;
}
.head-title h2 {
  margin: 0;
  font-size: 16px;
  color: var(--text, #f8fafc);
}
.head-title p {
  margin: 2px 0 0;
  font-size: 11px;
  color: var(--muted, #94a3b8);
}

.compare-tabs {
  display: flex;
  gap: 8px;
  padding: 12px 20px;
  background: var(--surface-2, #242b33);
  border-bottom: 1px solid var(--border, #343c45);
}
.tab-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 9999px;
  border: 1px solid var(--border, #343c45);
  background: var(--surface, #1e242b);
  color: var(--muted, #94a3b8);
  font-size: 11px;
  font-weight: 600;
  transition: all 0.15s ease;
}
.tab-pill:hover {
  color: var(--text, #f8fafc);
  border-color: rgba(56, 189, 248, 0.4);
}
.tab-pill.active {
  background: rgba(14, 165, 233, 0.15);
  color: #38bdf8;
  border-color: #38bdf8;
}

.compare-content-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.compare-table {
  width: 100%;
  border-collapse: collapse;
}

.compare-table th,
.compare-table td {
  padding: 10px 14px;
  border-bottom: 1px solid var(--border, #343c45);
  font-size: 11.5px;
  vertical-align: middle;
}

.feature-col {
  width: 200px;
  min-width: 180px;
  font-weight: 600;
  color: var(--muted, #94a3b8);
  background: var(--surface-2, #242b33);
}

.node-col {
  text-align: center;
  min-width: 160px;
}

.node-col-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--surface-2, #242b33);
  padding: 6px 8px;
  border-radius: 6px;
}
.col-head-info {
  display: flex;
  align-items: center;
  gap: 6px;
}
.country-tag {
  padding: 1px 4px;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  border-radius: 3px;
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  font-weight: 700;
}
.remove-btn {
  background: transparent;
  border: 0;
  color: var(--muted, #94a3b8);
  padding: 2px;
  display: grid;
  place-items: center;
  border-radius: 4px;
}
.remove-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.section-divider td {
  background: rgba(56, 189, 248, 0.08);
  color: #38bdf8;
  font-weight: 700;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 8px 14px;
}

.risk-score {
  font-family: 'Fira Code', monospace;
  font-size: 14px;
  font-weight: 700;
}
.risk-high { color: #ef4444; }
.risk-mid { color: #f59e0b; }
.risk-low { color: #10b981; }

.d-block { display: block; }
.text-muted { color: var(--muted, #94a3b8); }
.text-info { color: #38bdf8; }

.dnsbl-tag {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Fira Code', monospace;
  font-size: 10px;
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}
.dnsbl-tag.hit {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.svc-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.unlock-compare-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.quality-text {
  font-size: 9.5px;
  color: var(--muted, #94a3b8);
}
.latency-text {
  font-size: 8.5px;
  color: #38bdf8;
}

.compare-foot {
  display: flex;
  justify-content: flex-end;
  padding: 12px 20px;
  border-top: 1px solid var(--border, #343c45);
  background: var(--surface, #1e242b);
}
</style>
