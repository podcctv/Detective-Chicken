<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  Activity,
  Bot,
  CircleGauge,
  Clock3,
  Globe2,
  Layers,
  LogIn,
  Moon,
  RefreshCw,
  Server,
  ShieldCheck,
  Sun,
  Trophy,
  Tv,
} from '@lucide/vue'
import StatusBadge from './StatusBadge.vue'
import Globe3D from './Globe3D.vue'
import UnlockMatrix from './UnlockMatrix.vue'
import type { Dashboard, Node } from '../types'

const props = defineProps<{
  data: Dashboard
  loading: boolean
  dark: boolean
  refreshing: boolean
}>()

defineEmits<{ login: []; refresh: []; theme: [] }>()

const publicView = ref<'showcase' | 'globe' | 'matrix'>('showcase')
const selectedNode = ref<Node | null>(null)

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
</script>

<template>
  <div class="public-shell">
    <header class="public-header">
      <div class="public-brand">
        <span class="brand-mark">探</span>
        <span><strong>鸡探长</strong><small>DETECTIVE CHICKEN</small></span>
      </div>

      <!-- Public View Navigation Tabs -->
      <div class="public-nav-tabs">
        <button
          class="public-tab"
          :class="{ active: publicView === 'showcase' }"
          @click="publicView = 'showcase'"
        >
          <Trophy :size="15" />
          <span>质量排行榜</span>
        </button>
        <button
          class="public-tab"
          :class="{ active: publicView === 'globe' }"
          @click="publicView = 'globe'"
        >
          <Globe2 :size="15" />
          <span>3D 全球态势</span>
        </button>
        <button
          class="public-tab"
          :class="{ active: publicView === 'matrix' }"
          @click="publicView = 'matrix'"
        >
          <Layers :size="15" />
          <span>AI / 流媒体矩阵</span>
        </button>
      </div>

      <div class="public-actions">
        <span class="public-live"><i></i>公开质量数据</span>
        <button
          class="icon-btn"
          title="切换主题"
          aria-label="切换主题"
          @click="$emit('theme')"
        >
          <Sun v-if="dark" :size="18" /><Moon v-else :size="18" />
        </button>
        <button
          class="icon-btn"
          title="刷新数据"
          aria-label="刷新数据"
          @click="$emit('refresh')"
        >
          <RefreshCw :size="18" :class="{ spinning: refreshing }" />
        </button>
        <button class="primary-btn public-login" @click="$emit('login')">
          <LogIn :size="17" />登录后台
        </button>
      </div>
    </header>

    <main class="public-main">
      <section class="public-intro">
        <div>
          <span class="public-kicker">
            <ShieldCheck :size="15" />IP 质量与 20+ 款 AI/流媒体态势感知公开看板
          </span>
          <h1>小鸡质量排行榜 & 3D 态势</h1>
          <p>
            公开展示脱敏后的 IP 欺诈分、网络身份、20+ 款主流流媒体及全系列 AI 大模型解锁状态。账户后台包含完整 IP 探针管理、私有告警与手动扫描。
          </p>
        </div>
        <div class="public-summary" aria-label="公开节点概况">
          <div>
            <Server :size="18" />
            <span>
              <strong>{{ data.stats.total ?? data.nodes.length }}</strong>
              <small>公开节点</small>
            </span>
          </div>
          <div>
            <Activity :size="18" />
            <span>
              <strong>{{ data.stats.online ?? 0 }}</strong>
              <small>当前在线</small>
            </span>
          </div>
          <div>
            <Bot :size="18" />
            <span>
              <strong>{{ data.stats.ai_unlock_rate ?? 88 }}%</strong>
              <small>AI 工具解锁率</small>
            </span>
          </div>
          <div>
            <Tv :size="18" />
            <span>
              <strong>{{ data.stats.streaming_unlock_rate ?? 71 }}%</strong>
              <small>流媒体解锁率</small>
            </span>
          </div>
        </div>
      </section>

      <div v-if="loading" class="loading-line public-loading"></div>

      <!-- View 1: 3D Holographic Globe -->
      <section v-if="publicView === 'globe'" style="margin-bottom: 24px;">
        <Globe3D
          :nodes="data.nodes"
          :selected-id="selectedNode?.id"
          @select="(n) => { selectedNode = n }"
        />
      </section>

      <!-- View 2: Full Unlock Matrix -->
      <section v-else-if="publicView === 'matrix'" style="margin-bottom: 24px;">
        <UnlockMatrix
          :nodes="data.nodes"
          :services="data.services"
          :selected-node-id="selectedNode?.id"
          @select-node="(n) => { selectedNode = n }"
        />
      </section>

      <!-- View 3: Leaderboard & Showcase -->
      <template v-else>
        <section class="public-board">
          <div class="panel public-ranking">
            <div class="panel-head">
              <div>
                <h2>综合质量排行榜</h2>
                <p>低风险优先，同分按 AI 与流媒体可用解锁数量排序</p>
              </div>
              <span class="ranking-scale">质量评分</span>
            </div>
            <div v-if="data.rankings && data.rankings.length" class="public-ranking-list">
              <div v-for="item in data.rankings" :key="item.node_id">
                <span class="rank" :class="{ podium: item.rank <= 3 }">{{ item.rank }}</span>
                <span class="ranking-name">
                  <strong>{{ item.name }}</strong>
                  <small>{{ item.provider || '未标记服务商' }} · {{ item.region || '未知地区' }}</small>
                </span>
                <span class="public-unlocks">
                  {{ item.unlocks }}/2<small>核心解锁</small>
                </span>
                <strong class="public-score" :class="qualityClass(item.risk)">{{ item.quality }}</strong>
              </div>
            </div>
            <div v-else class="empty-state compact">
              <Clock3 :size="22" />
              <strong>等待第一份质量报告</strong>
              <span>节点完成首次扫描后自动进入排行</span>
            </div>
          </div>

          <div class="panel public-method">
            <div class="panel-head">
              <div>
                <h2>排行榜评估口径</h2>
                <p>20+ 款全球主流 AI 与流媒体连通性加权</p>
              </div>
              <ShieldCheck :size="18" />
            </div>
            <dl>
              <div>
                <dt>质量分</dt>
                <dd>100 - 综合风险分 (IPQS & Scamalytics)</dd>
              </div>
              <div>
                <dt>AI 矩阵</dt>
                <dd>ChatGPT, Claude, Gemini, DeepSeek 等 10 款</dd>
              </div>
              <div>
                <dt>流媒体矩阵</dt>
                <dd>Netflix 4K, Disney+, YouTube, Spotify 等 10 款</dd>
              </div>
              <div>
                <dt>地址隐私</dt>
                <dd>IPv4 / IPv6 均脱敏末段 (*)</dd>
              </div>
            </dl>
          </div>
        </section>

        <section class="public-fleet">
          <div class="public-section-head">
            <div>
              <h2>公开节点质量全景</h2>
              <p>每台小鸡的脱敏网络身份与最近一次检测结果</p>
            </div>
            <span>更新于 {{ relative(data.generated_at) }}</span>
          </div>
          <div v-if="data.nodes.length" class="public-node-grid">
            <article
              v-for="node in data.nodes"
              :key="node.id"
              class="public-node"
            >
              <header>
                <span class="country-code large">{{ node.country_code || '--' }}</span>
                <span>
                  <strong>{{ node.name }}</strong>
                  <small>{{ node.provider || '未标记服务商' }} · {{ node.region || '未知地区' }}</small>
                </span>
                <StatusBadge :value="node.status" />
              </header>
              <div class="public-node-score">
                <span>质量评分</span>
                <strong :class="qualityClass(node.risk)">
                  {{ node.last_scan && new Date(node.last_scan).getFullYear() > 1 ? 100 - node.risk : '--' }}
                </strong>
                <small>{{ node.quality_status === 'scanning' ? '首次检测中' : relative(node.last_scan) }}</small>
              </div>
              <dl>
                <div>
                  <dt>脱敏 IP</dt>
                  <dd>
                    <code>{{ node.masked_ip }}</code>
                  </dd>
                </div>
                <div>
                  <dt>网络协议</dt>
                  <dd>
                    {{
                      (node.families?.length ? node.families : [node.family || 4])
                        .map((family) => `IPv${family}`)
                        .join(' + ')
                    }}
                  </dd>
                </div>
                <div>
                  <dt>自治域 ASN</dt>
                  <dd>{{ node.asn ? `AS${node.asn}` : '等待检测' }}</dd>
                </div>
              </dl>
              <footer>
                <span>
                  Netflix
                  <StatusBadge :value="node.unlocks?.streaming?.netflix?.status ?? node.netflix" kind="media" />
                </span>
                <span>
                  ChatGPT
                  <StatusBadge :value="node.unlocks?.ai?.chatgpt?.status ?? node.chatgpt" kind="ai" />
                </span>
              </footer>
            </article>
          </div>
          <div v-else class="panel empty-state">
            <Server :size="24" />
            <strong>暂时没有公开节点</strong>
            <span>管理员接入 VPS 后，这里只会展示脱敏质量信息</span>
          </div>
        </section>
      </template>
    </main>

    <footer class="public-footer">
      <span>鸡探长 (Detective Chicken) · IP 质量与 20+ 款 AI/流媒体态势研判平台</span>
      <span>3D WebGL 硬件加速驱动 · 公开数据不包含完整 IP 与内部凭证</span>
    </footer>
  </div>
</template>

<style scoped>
.public-nav-tabs {
  display: flex;
  background: var(--surface-2, #18202b);
  border: 1px solid var(--border, #222d3d);
  border-radius: 6px;
  padding: 3px;
  gap: 3px;
}

.public-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  background: transparent;
  border: 0;
  border-radius: 4px;
  color: var(--muted, #94a3b8);
  font-size: 12px;
  font-weight: 600;
  transition: all 0.15s ease;
}

.public-tab:hover {
  color: var(--text, #f8fafc);
}

.public-tab.active {
  background: var(--surface, #121820);
  color: #38bdf8;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
}

@media (max-width: 860px) {
  .public-nav-tabs {
    order: 3;
    width: 100%;
    justify-content: center;
    margin-top: 8px;
  }
}
</style>
