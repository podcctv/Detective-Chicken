<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { Activity, AlertTriangle, Bell, ChevronRight, CircleGauge, Cloud, Copy, LayoutDashboard, ListFilter, Menu, Moon, Network, Plus, RefreshCw, ScanLine, Search, Server, Settings, ShieldAlert, Sun, X } from '@lucide/vue'
import StatusBadge from './components/StatusBadge.vue'
import { demoDashboard } from './demo'
import type { Dashboard, Node } from './types'

const data = ref<Dashboard>(demoDashboard)
const loading = ref(true)
const refreshing = ref(false)
const usingDemo = ref(false)
const dark = ref(false)
const menuOpen = ref(false)
const selected = ref<Node | null>(null)
const enrollOpen = ref(false)
const search = ref('')
const filter = ref('all')
const toast = ref('')
const form = ref({ name: '', provider: '', region: '' })
const enrollment = ref<{ token: string; expires_at: string } | null>(null)
const TrendChart = defineAsyncComponent(() => import('./components/TrendChart.vue'))

const statCards = computed(() => [
  { key: 'total', label: '总节点', value: data.value.stats.total ?? 0, hint: '全部资产', icon: Server, tone: 'neutral' },
  { key: 'online', label: '在线', value: data.value.stats.online ?? 0, hint: '心跳正常', icon: Activity, tone: 'good' },
  { key: 'abnormal', label: '异常', value: data.value.stats.abnormal ?? 0, hint: '需关注', icon: AlertTriangle, tone: 'warn' },
  { key: 'high_risk', label: '高风险 IP', value: data.value.stats.high_risk ?? 0, hint: '风险 ≥ 60', icon: ShieldAlert, tone: 'danger' },
  { key: 'ip_changes', label: '24h IP 变更', value: data.value.stats.ip_changes ?? 0, hint: '身份变化', icon: Network, tone: 'info' },
  { key: 'media_degraded', label: '解锁下降', value: data.value.stats.media_degraded ?? 0, hint: '媒体或 AI', icon: Cloud, tone: 'violet' },
  { key: 'dnsbl_added', label: '黑名单新增', value: data.value.stats.dnsbl_added ?? 0, hint: 'DNSBL 命中', icon: Bell, tone: 'danger' },
])

const filteredNodes = computed(() => data.value.nodes.filter((n) => {
  const matchesText = `${n.name} ${n.provider} ${n.region} ${n.masked_ip} ${n.organization}`.toLowerCase().includes(search.value.toLowerCase())
  const matchesFilter = filter.value === 'all' || n.status === filter.value
  return matchesText && matchesFilter
}))

const load = async (manual = false) => {
  if (manual) refreshing.value = true
  try {
    const response = await fetch('/api/v1/dashboard')
    if (!response.ok) throw new Error(response.statusText)
    data.value = await response.json()
    usingDemo.value = false
  } catch {
    data.value = demoDashboard
    usingDemo.value = true
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const relative = (input: string) => {
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(input).getTime()) / 1000))
  if (seconds < 60) return `${seconds} 秒前`
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
  return `${Math.floor(seconds / 3600)} 小时前`
}

const riskClass = (risk: number) => risk >= 60 ? 'risk-high' : risk >= 35 ? 'risk-mid' : 'risk-low'
const riskLabel = (risk: number) => risk >= 60 ? '高' : risk >= 35 ? '中' : '低'
const showToast = (message: string) => { toast.value = message; setTimeout(() => { toast.value = '' }, 2800) }

const scan = async (node: Node) => {
  try { await fetch(`/api/v1/nodes/${node.id}/scan`, { method: 'POST' }); showToast(`已向 ${node.name} 下发扫描任务`) }
  catch { showToast('演示模式：扫描任务已模拟下发') }
}

const createEnrollment = async () => {
  if (!form.value.name.trim()) return
  try {
    const response = await fetch('/api/v1/enrollment-tokens', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(form.value) })
    if (!response.ok) throw new Error()
    enrollment.value = await response.json()
  } catch {
    enrollment.value = { token: 'et_demo_9f8a7b6c5d4e', expires_at: new Date(Date.now() + 600_000).toISOString() }
  }
}

const installCommand = computed(() => enrollment.value ? `curl -fsSL https://agent.example.com/install.sh | sudo bash -s -- --server '${location.origin}' --enroll '${enrollment.value.token}'` : '')
const copyCommand = async () => { await navigator.clipboard.writeText(installCommand.value); showToast('安装命令已复制') }
const toggleDark = () => { dark.value = !dark.value; document.documentElement.dataset.theme = dark.value ? 'dark' : 'light'; localStorage.setItem('jijian-theme', dark.value ? 'dark' : 'light'); nextTick(() => window.dispatchEvent(new Event('resize'))) }

const closeOverlay = (event: KeyboardEvent) => { if (event.key === 'Escape') { selected.value = null; enrollOpen.value = false; menuOpen.value = false } }
onMounted(() => { dark.value = localStorage.getItem('jijian-theme') === 'dark'; document.documentElement.dataset.theme = dark.value ? 'dark' : 'light'; window.addEventListener('keydown', closeOverlay); load() })
onBeforeUnmount(() => window.removeEventListener('keydown', closeOverlay))
</script>

<template>
  <div class="app-shell">
    <a class="skip-link" href="#main-content">跳到主要内容</a>
    <aside class="sidebar" :class="{ open: menuOpen }">
      <div class="brand"><div class="brand-mark">鉴</div><div><strong>鸡鉴</strong><span>IP QUALITY OPS</span></div><button class="icon-btn mobile-close" aria-label="关闭菜单" @click="menuOpen = false"><X :size="19" /></button></div>
      <nav aria-label="主导航">
        <a class="nav-item active" href="#dashboard"><LayoutDashboard :size="18" /><span>总览</span></a>
        <a class="nav-item" href="#fleet"><Server :size="18" /><span>节点资产</span><small>{{ data.stats.total }}</small></a>
        <a class="nav-item" href="#alerts"><Bell :size="18" /><span>告警中心</span><small class="danger-count">{{ data.alerts.length }}</small></a>
        <a class="nav-item" href="#reports"><ScanLine :size="18" /><span>检测报告</span></a>
      </nav>
      <div class="sidebar-section"><span>管理</span></div>
      <nav aria-label="管理导航">
        <a class="nav-item" href="#settings"><Settings :size="18" /><span>系统设置</span></a>
      </nav>
      <div class="sidebar-foot"><div class="health-dot"></div><div><strong>控制面正常</strong><span>API v0.1.0</span></div></div>
    </aside>
    <button v-if="menuOpen" class="sidebar-scrim" aria-label="关闭菜单" @click="menuOpen = false"></button>

    <main class="main-content" id="main-content">
      <header class="topbar">
        <div class="page-heading"><button class="icon-btn menu-btn" aria-label="打开菜单" @click="menuOpen = true"><Menu :size="20" /></button><div><h1>VPS 舰队总览</h1><p>持续追踪公网 IP 身份、风险与解锁能力</p></div></div>
        <div class="top-actions">
          <span class="env-badge"><i></i>{{ usingDemo ? '演示数据' : '实时数据' }}</span>
          <button class="icon-btn" title="切换主题" aria-label="切换明暗主题" @click="toggleDark"><Sun v-if="dark" :size="18" /><Moon v-else :size="18" /></button>
          <button class="icon-btn" title="刷新数据" aria-label="刷新数据" @click="load(true)"><RefreshCw :size="18" :class="{ spinning: refreshing }" /></button>
          <button class="primary-btn" @click="enrollOpen = true"><Plus :size="17" />添加 VPS</button>
        </div>
      </header>

      <div v-if="loading" class="loading-line" aria-label="正在加载"></div>

      <section class="kpi-grid" aria-label="关键指标">
        <article v-for="card in statCards" :key="card.key" class="kpi-card" :class="`tone-${card.tone}`">
          <div class="kpi-icon"><component :is="card.icon" :size="18" /></div><div><span>{{ card.label }}</span><strong>{{ card.value }}</strong><small>{{ card.hint }}</small></div>
        </article>
      </section>

      <section class="analytics-grid">
        <div class="panel trend-panel">
          <div class="panel-head"><div><h2>风险分趋势</h2><p>US-LAX-02 · 最近 10 天 · 12 小时聚合</p></div><span class="trend-change">+31 <small>近 24h</small></span></div>
          <TrendChart :points="data.trend" />
          <div class="chart-foot"><span><i class="anomaly-dot"></i>异常上升点</span><span>最近更新 {{ relative(data.generated_at) }}</span></div>
        </div>
        <div class="panel alert-panel" id="alerts">
          <div class="panel-head"><div><h2>近期告警</h2><p>变化型规则优先</p></div><button class="text-btn">查看全部<ChevronRight :size="15" /></button></div>
          <div class="alert-list">
            <button v-for="alert in data.alerts" :key="alert.id" class="alert-row" @click="selected = data.nodes.find((n) => n.id === alert.node_id) ?? null">
              <span class="alert-mark" :class="alert.severity"><ShieldAlert v-if="alert.severity === 'critical'" :size="16" /><AlertTriangle v-else :size="16" /></span>
              <span class="alert-copy"><strong>{{ alert.title }}</strong><span>{{ alert.node_name }} · {{ alert.detail }}</span><small>{{ relative(alert.created_at) }}</small></span>
              <ChevronRight :size="16" />
            </button>
          </div>
        </div>
      </section>

      <section class="panel fleet-panel" id="fleet">
        <div class="fleet-head"><div><h2>节点资产</h2><p>{{ filteredNodes.length }} 个结果，按风险从高到低</p></div><div class="fleet-tools"><label class="search-field"><Search :size="16" /><input v-model="search" type="search" placeholder="搜索节点、地区或 ASN" aria-label="搜索节点" /></label><label class="filter-field"><ListFilter :size="16" /><select v-model="filter" aria-label="筛选节点状态"><option value="all">全部状态</option><option value="online">在线</option><option value="warning">注意</option><option value="alert">告警</option><option value="offline">离线</option></select></label></div></div>
        <div class="table-wrap">
          <table>
            <thead><tr><th>节点</th><th>IP 地址</th><th>ASN / 地区</th><th>综合风险</th><th>Netflix</th><th>ChatGPT</th><th>DNSBL</th><th>上次上报</th><th>状态</th><th><span class="sr-only">操作</span></th></tr></thead>
            <tbody><tr v-for="node in filteredNodes" :key="node.id" tabindex="0" @click="selected = node" @keydown.enter="selected = node">
              <td><div class="node-name"><span class="country-code">{{ node.country_code || '--' }}</span><div><strong>{{ node.name }}</strong><small>{{ node.provider }}</small></div></div></td>
              <td><code>{{ node.masked_ip }}</code><small class="family">IPv{{ node.family || 4 }}</small></td>
              <td><strong class="asn">AS{{ node.asn || '—' }}</strong><small>{{ node.region }}</small></td>
              <td><div class="risk-cell"><strong :class="riskClass(node.risk)">{{ node.risk }}</strong><span>{{ riskLabel(node.risk) }}风险</span></div></td>
              <td><StatusBadge :value="node.netflix" kind="media" /></td><td><StatusBadge :value="node.chatgpt" kind="media" /></td>
              <td><span class="dnsbl" :class="{ hit: node.dnsbl > 0 }">{{ node.dnsbl }}</span></td><td><span class="last-seen">{{ relative(node.last_seen) }}</span></td><td><StatusBadge :value="node.status" /></td>
              <td><button class="row-action" title="查看详情" aria-label="查看节点详情" @click.stop="selected = node"><ChevronRight :size="17" /></button></td>
            </tr></tbody>
          </table>
          <div v-if="filteredNodes.length === 0" class="empty-state"><Search :size="24" /><strong>没有匹配的节点</strong><span>调整搜索词或状态筛选</span></div>
        </div>
      </section>
      <footer><span>鸡鉴 · 数据默认仅租户可见，IP 已脱敏</span><span>心跳 2 min · 完整扫描 6 h + jitter</span></footer>
    </main>

    <Transition name="drawer"><div v-if="selected" class="drawer-backdrop" @click.self="selected = null"><aside class="detail-drawer" aria-label="节点详情">
      <div class="drawer-head"><div><span class="country-code large">{{ selected.country_code }}</span><div><h2>{{ selected.name }}</h2><p>{{ selected.provider }} · {{ selected.region }}</p></div></div><button class="icon-btn" aria-label="关闭详情" @click="selected = null"><X :size="19" /></button></div>
      <div class="drawer-summary"><div><span>综合风险</span><strong :class="riskClass(selected.risk)">{{ selected.risk }}</strong></div><div><span>当前状态</span><StatusBadge :value="selected.status" /></div><div><span>上次扫描</span><strong>{{ relative(selected.last_scan) }}</strong></div></div>
      <div class="detail-section"><h3>网络身份</h3><dl><div><dt>公网 IP</dt><dd><code>{{ selected.masked_ip }}</code></dd></div><div><dt>ASN</dt><dd>AS{{ selected.asn }} · {{ selected.organization }}</dd></div><div><dt>网络族</dt><dd>IPv{{ selected.family }}</dd></div></dl></div>
      <div class="detail-section"><h3>最近变化</h3><div class="change-list"><div v-if="selected.ip_changed"><Network :size="17" /><span><strong>公网 IP 已变更</strong><small>检测到新的网络身份，建议复核解锁状态</small></span></div><div v-if="selected.risk >= 60"><CircleGauge :size="17" /><span><strong>风险分 41 → {{ selected.risk }}</strong><small>超过变化阈值 20</small></span></div><div v-if="selected.chatgpt !== 'available'"><Cloud :size="17" /><span><strong>ChatGPT available → {{ selected.chatgpt }}</strong><small>解锁能力下降</small></span></div><div v-if="!selected.ip_changed && selected.risk < 60 && selected.chatgpt === 'available'"><Activity :size="17" /><span><strong>未发现显著变化</strong><small>网络身份与质量保持稳定</small></span></div></div></div>
      <div class="drawer-actions"><button class="secondary-btn" @click="selected = null">关闭</button><button class="primary-btn" @click="scan(selected)"><ScanLine :size="17" />立即扫描</button></div>
    </aside></div></Transition>

    <Transition name="modal"><div v-if="enrollOpen" class="modal-backdrop" @click.self="enrollOpen = false"><section class="modal" role="dialog" aria-modal="true" aria-labelledby="enroll-title">
      <div class="modal-head"><div><h2 id="enroll-title">添加 VPS</h2><p>创建一次性凭证，在目标服务器完成注册</p></div><button class="icon-btn" aria-label="关闭" @click="enrollOpen = false"><X :size="19" /></button></div>
      <form v-if="!enrollment" @submit.prevent="createEnrollment"><label>节点名称<input v-model="form.name" required placeholder="例如 HK-CMI-01" /></label><div class="form-grid"><label>服务商<input v-model="form.provider" placeholder="例如 DMIT" /></label><label>地区标签<input v-model="form.region" placeholder="例如 香港" /></label></div><div class="form-note"><ShieldAlert :size="17" /><span>注册凭证 10 分钟内有效且仅可使用一次。Agent 私钥始终在 VPS 本机生成。</span></div><div class="modal-actions"><button type="button" class="secondary-btn" @click="enrollOpen = false">取消</button><button class="primary-btn" type="submit">生成安装命令<ChevronRight :size="17" /></button></div></form>
      <div v-else class="enrollment-result"><div class="step-status"><span>1</span><div><strong>一次性凭证已创建</strong><small>{{ new Date(enrollment.expires_at).toLocaleTimeString('zh-CN') }} 前有效</small></div></div><label>在目标 VPS 上执行</label><div class="command-box"><code>{{ installCommand }}</code><button class="icon-btn" title="复制命令" aria-label="复制安装命令" @click="copyCommand"><Copy :size="17" /></button></div><div class="installation-steps"><div class="done"><i></i>等待安装</div><div><i></i>Agent 注册</div><div><i></i>首次心跳</div><div><i></i>完成检测</div></div><div class="modal-actions"><button class="secondary-btn" @click="enrollment = null">返回</button><button class="primary-btn" @click="enrollOpen = false; enrollment = null; form = { name: '', provider: '', region: '' }">完成</button></div></div>
    </section></div></Transition>
    <Transition name="toast"><div v-if="toast" class="toast" role="status"><Activity :size="17" />{{ toast }}</div></Transition>
  </div>
</template>
