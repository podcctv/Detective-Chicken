<script setup lang="ts">
import {
  Activity,
  CircleGauge,
  Clock3,
  LogIn,
  Moon,
  RefreshCw,
  Server,
  ShieldCheck,
  Sun,
} from "@lucide/vue";
import StatusBadge from "./StatusBadge.vue";
import type { Dashboard } from "../types";

defineProps<{
  data: Dashboard;
  loading: boolean;
  dark: boolean;
  refreshing: boolean;
}>();
defineEmits<{ login: []; refresh: []; theme: [] }>();

const qualityClass = (risk: number) =>
  risk >= 60 ? "risk-high" : risk >= 35 ? "risk-mid" : "risk-low";
const relative = (input?: string) => {
  if (!input || new Date(input).getFullYear() <= 1) return "等待首次检测";
  const seconds = Math.max(
    0,
    Math.floor((Date.now() - new Date(input).getTime()) / 1000),
  );
  if (seconds < 60) return `${seconds} 秒前`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} 小时前`;
  return `${Math.floor(seconds / 86400)} 天前`;
};
</script>

<template>
  <div class="public-shell">
    <header class="public-header">
      <div class="public-brand">
        <span class="brand-mark">探</span
        ><span><strong>鸡探长</strong><small>DETECTIVE CHICKEN</small></span>
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
          title="刷新排行"
          aria-label="刷新排行"
          @click="$emit('refresh')"
        >
          <RefreshCw :size="18" :class="{ spinning: refreshing }" />
        </button>
        <button class="primary-btn public-login" @click="$emit('login')">
          <LogIn :size="17" />登录
        </button>
      </div>
    </header>

    <main class="public-main">
      <section class="public-intro">
        <div>
          <span class="public-kicker"
            ><ShieldCheck :size="15" />IP 质量公开看板</span
          >
          <h1>小鸡质量排行榜</h1>
          <p>
            公开展示脱敏后的 IP 信誉、网络身份与流媒体解锁结果。账户后台包含完整
            IP、Agent 管理、告警和手动扫描。
          </p>
        </div>
        <div class="public-summary" aria-label="公开节点概况">
          <div>
            <Server :size="18" /><span
              ><strong>{{ data.stats.total ?? 0 }}</strong
              ><small>公开节点</small></span
            >
          </div>
          <div>
            <Activity :size="18" /><span
              ><strong>{{ data.stats.online ?? 0 }}</strong
              ><small>当前在线</small></span
            >
          </div>
          <div>
            <CircleGauge :size="18" /><span
              ><strong>{{ data.stats.scanned ?? 0 }}</strong
              ><small>已有报告</small></span
            >
          </div>
        </div>
      </section>

      <div v-if="loading" class="loading-line public-loading"></div>

      <section class="public-board">
        <div class="panel public-ranking">
          <div class="panel-head">
            <div>
              <h2>综合质量榜</h2>
              <p>低风险优先，同分按可用解锁数量排序</p>
            </div>
            <span class="ranking-scale">质量分</span>
          </div>
          <div v-if="data.rankings.length" class="public-ranking-list">
            <div v-for="item in data.rankings" :key="item.node_id">
              <span class="rank" :class="{ podium: item.rank <= 3 }">{{
                item.rank
              }}</span>
              <span class="ranking-name"
                ><strong>{{ item.name }}</strong
                ><small
                  >{{ item.provider || "未标记服务商" }} ·
                  {{ item.region || "未知地区" }}</small
                ></span
              >
              <span class="public-unlocks"
                >{{ item.unlocks }}/2<small>核心解锁</small></span
              >
              <strong class="public-score" :class="qualityClass(item.risk)">{{
                item.quality
              }}</strong>
            </div>
          </div>
          <div v-else class="empty-state compact">
            <Clock3 :size="22" /><strong>等待第一份质量报告</strong
            ><span>节点完成首次扫描后自动进入排行</span>
          </div>
        </div>

        <div class="panel public-method">
          <div class="panel-head">
            <div>
              <h2>排行榜口径</h2>
              <p>只展示已经完成检测的节点</p>
            </div>
            <ShieldCheck :size="18" />
          </div>
          <dl>
            <div>
              <dt>质量分</dt>
              <dd>100 - 综合风险</dd>
            </div>
            <div>
              <dt>核心解锁</dt>
              <dd>Netflix + ChatGPT</dd>
            </div>
            <div>
              <dt>地址隐私</dt>
              <dd>IPv4 / IPv6 均隐藏末两段</dd>
            </div>
            <div>
              <dt>更新方式</dt>
              <dd>按节点设定周期自动检测</dd>
            </div>
          </dl>
        </div>
      </section>

      <section class="public-fleet">
        <div class="public-section-head">
          <div>
            <h2>公开节点质量</h2>
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
              <span class="country-code large">{{
                node.country_code || "--"
              }}</span
              ><span
                ><strong>{{ node.name }}</strong
                ><small
                  >{{ node.provider || "未标记服务商" }} ·
                  {{ node.region || "未知地区" }}</small
                ></span
              ><StatusBadge :value="node.status" />
            </header>
            <div class="public-node-score">
              <span>质量评分</span
              ><strong :class="qualityClass(node.risk)">{{
                node.last_scan && new Date(node.last_scan).getFullYear() > 1
                  ? 100 - node.risk
                  : "--"
              }}</strong
              ><small>{{
                node.quality_status === "scanning"
                  ? "首次检测中"
                  : relative(node.last_scan)
              }}</small>
            </div>
            <dl>
              <div>
                <dt>脱敏 IP</dt>
                <dd>
                  <code>{{ node.masked_ip }}</code>
                </dd>
              </div>
              <div>
                <dt>网络</dt>
                <dd>
                  {{
                    (node.families?.length ? node.families : [node.family || 4])
                      .map((family) => `IPv${family}`)
                      .join(" + ")
                  }}
                </dd>
              </div>
              <div>
                <dt>ASN</dt>
                <dd>{{ node.asn ? `AS${node.asn}` : "等待检测" }}</dd>
              </div>
            </dl>
            <footer>
              <span
                >Netflix
                <StatusBadge :value="node.netflix" kind="media" /></span
              ><span
                >ChatGPT <StatusBadge :value="node.chatgpt" kind="media"
              /></span>
            </footer>
          </article>
        </div>
        <div v-else class="panel empty-state">
          <Server :size="24" /><strong>暂时没有公开节点</strong
          ><span>管理员接入 VPS 后，这里只会展示脱敏质量信息</span>
        </div>
      </section>
    </main>

    <footer class="public-footer">
      <span>鸡探长 · Detective Chicken</span
      ><span>公开数据不包含完整 IP、账户信息、Agent 凭证和内部告警</span>
    </footer>
  </div>
</template>
