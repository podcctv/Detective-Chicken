<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { Activity, AlertTriangle, Bell, ChevronRight, CircleGauge, Cloud, Copy, Eye, EyeOff, KeyRound, LayoutDashboard, ListFilter, LogOut, Menu, Moon, Network, Plus, RefreshCw, ScanLine, Search, Server, Settings, ShieldAlert, Sun, UserCog, X } from '@lucide/vue'
import StatusBadge from './components/StatusBadge.vue'
import type { AuthStatus, Dashboard, Enrollment, Node, NodeDetail, User } from './types'

const TrendChart = defineAsyncComponent(() => import('./components/TrendChart.vue'))
const emptyDashboard = (): Dashboard => ({ generated_at: new Date().toISOString(), stats: {}, trend: [], nodes: [], rankings: [], alerts: [], regions: {} })

const auth = ref<AuthStatus | null>(null)
const data = ref<Dashboard>(emptyDashboard())
const loading = ref(true)
const refreshing = ref(false)
const dark = ref(false)
const menuOpen = ref(false)
const search = ref('')
const filter = ref('all')
const toast = ref('')
const fatalError = ref('')

const authMode = ref<'login' | 'register'>('login')
const authBusy = ref(false)
const authError = ref('')
const authForm = ref({ username: '', display_name: '', password: '' })
const resetToken = ref('')
const resetPassword = ref('')

const selected = ref<NodeDetail | null>(null)
const detailLoading = ref(false)
const showFullIP = ref(false)
const enrollOpen = ref(false)
const enrollment = ref<Enrollment | null>(null)
const enrollForm = ref({ name: '', provider: '', region: '', os_family: 'auto', platform: 'auto', arch: 'auto' })

const adminOpen = ref(false)
const users = ref<User[]>([])
const registrationEnabled = ref(false)
const resetLink = ref('')
const passwordOpen = ref(false)
const passwordForm = ref({ current_password: '', new_password: '' })

const api = async <T>(path: string, options: RequestInit = {}): Promise<T> => {
  const response = await fetch(path, {
    ...options,
    credentials: 'include',
    headers: { ...(options.body ? { 'Content-Type': 'application/json' } : {}), ...(options.headers ?? {}) },
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body?.error?.message || `请求失败 (${response.status})`)
  }
  if (response.status === 204) return undefined as T
  return response.json()
}

const boot = async () => {
  loading.value = true
  fatalError.value = ''
  try {
    auth.value = await api<AuthStatus>('/api/v1/auth/status')
    if (!auth.value.settings.bootstrapped) authMode.value = 'register'
    if (auth.value.authenticated) await loadDashboard()
  } catch (error) {
    fatalError.value = error instanceof Error ? error.message : '无法连接控制面 API'
  } finally {
    loading.value = false
  }
}

const submitAuth = async () => {
  authBusy.value = true
  authError.value = ''
  try {
    await api(`/api/v1/auth/${authMode.value}`, { method: 'POST', body: JSON.stringify(authForm.value) })
    authForm.value.password = ''
    await boot()
  } catch (error) {
    authError.value = error instanceof Error ? error.message : '认证失败'
  } finally {
    authBusy.value = false
  }
}

const completeReset = async () => {
  authBusy.value = true
  authError.value = ''
  try {
    await api('/api/v1/auth/password-reset/complete', { method: 'POST', body: JSON.stringify({ token: resetToken.value, new_password: resetPassword.value }) })
    location.hash = ''
    resetToken.value = ''
    resetPassword.value = ''
    authMode.value = 'login'
    showToast('密码已重置，请重新登录')
  } catch (error) {
    authError.value = error instanceof Error ? error.message : '密码重置失败'
  } finally {
    authBusy.value = false
  }
}

const logout = async () => {
  await api('/api/v1/auth/logout', { method: 'POST' }).catch(() => undefined)
  selected.value = null
  data.value = emptyDashboard()
  await boot()
}

const loadDashboard = async (manual = false) => {
  if (manual) refreshing.value = true
  try {
    data.value = await api<Dashboard>('/api/v1/dashboard')
    data.value.rankings ??= []
  } catch (error) {
    fatalError.value = error instanceof Error ? error.message : '总览加载失败'
  } finally {
    refreshing.value = false
  }
}

const statCards = computed(() => [
  { key: 'total', label: '总节点', value: data.value.stats.total ?? 0, hint: '已纳管资产', icon: Server, tone: 'neutral' },
  { key: 'online', label: '在线', value: data.value.stats.online ?? 0, hint: '心跳正常', icon: Activity, tone: 'good' },
  { key: 'abnormal', label: '异常', value: data.value.stats.abnormal ?? 0, hint: '需要关注', icon: AlertTriangle, tone: 'warn' },
  { key: 'high_risk', label: '高风险 IP', value: data.value.stats.high_risk ?? 0, hint: '风险 ≥ 60', icon: ShieldAlert, tone: 'danger' },
  { key: 'media_degraded', label: '解锁异常', value: data.value.stats.media_degraded ?? 0, hint: '流媒体或 AI', icon: Cloud, tone: 'violet' },
  { key: 'dnsbl_added', label: 'DNSBL 命中', value: data.value.stats.dnsbl_added ?? 0, hint: '邮件信誉风险', icon: Bell, tone: 'danger' },
])

const filteredNodes = computed(() => data.value.nodes.filter((node) => {
  const matchesText = `${node.name} ${node.provider} ${node.region} ${node.masked_ip} ${node.organization}`.toLowerCase().includes(search.value.toLowerCase())
  return matchesText && (filter.value === 'all' || node.status === filter.value)
}))

const relative = (input?: string) => {
  if (!input || new Date(input).getFullYear() <= 1) return '尚未上报'
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(input).getTime()) / 1000))
  if (seconds < 60) return `${seconds} 秒前`
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} 小时前`
  return `${Math.floor(seconds / 86400)} 天前`
}

const riskClass = (risk: number) => risk >= 60 ? 'risk-high' : risk >= 35 ? 'risk-mid' : 'risk-low'
const riskLabel = (risk: number) => risk >= 60 ? '高' : risk >= 35 ? '中' : '低'
const showToast = (message: string) => { toast.value = message; window.setTimeout(() => { toast.value = '' }, 3200) }

const openNode = async (node: Node, fullIP = false) => {
  detailLoading.value = true
  showFullIP.value = fullIP
  try {
    selected.value = await api<NodeDetail>(`/api/v1/nodes/${node.id}${fullIP ? '?full_ip=true' : ''}`)
  } catch (error) {
    showToast(error instanceof Error ? error.message : '节点详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const toggleFullIP = async () => {
  if (!selected.value) return
  await openNode(selected.value, !showFullIP.value)
}

const scan = async (node: Node) => {
  try {
    await api(`/api/v1/nodes/${node.id}/scan`, { method: 'POST' })
    showToast(`已向 ${node.name} 下发扫描任务`)
  } catch (error) {
    showToast(error instanceof Error ? error.message : '扫描任务下发失败')
  }
}

const createEnrollment = async () => {
  try {
    enrollment.value = await api<Enrollment>('/api/v1/enrollment-tokens', { method: 'POST', body: JSON.stringify(enrollForm.value) })
  } catch (error) {
    showToast(error instanceof Error ? error.message : '创建安装命令失败')
  }
}

const copyText = async (value: string, message: string) => {
  await navigator.clipboard.writeText(value)
  showToast(message)
}

const closeEnrollment = () => {
  enrollOpen.value = false
  enrollment.value = null
  enrollForm.value = { name: '', provider: '', region: '', os_family: 'auto', platform: 'auto', arch: 'auto' }
}

const openAdmin = async () => {
  adminOpen.value = true
  resetLink.value = ''
  try {
    const [userResponse, settings] = await Promise.all([
      api<{ items: User[] }>('/api/v1/admin/users'),
      api<{ registration_enabled: boolean }>('/api/v1/admin/settings'),
    ])
    users.value = userResponse.items
    registrationEnabled.value = settings.registration_enabled
  } catch (error) {
    showToast(error instanceof Error ? error.message : '管理设置加载失败')
  }
}

const saveRegistration = async () => {
  try {
    await api('/api/v1/admin/settings', { method: 'PATCH', body: JSON.stringify({ registration_enabled: registrationEnabled.value }) })
    if (auth.value) auth.value.settings.registration_enabled = registrationEnabled.value
    showToast(registrationEnabled.value ? '已允许新用户注册' : '已关闭新用户注册')
  } catch (error) {
    showToast(error instanceof Error ? error.message : '设置保存失败')
  }
}

const updateRole = async (user: User, role: string) => {
  try {
    const updated = await api<User>(`/api/v1/admin/users/${user.id}`, { method: 'PATCH', body: JSON.stringify({ role }) })
    Object.assign(user, updated)
    showToast(`${user.display_name} 已设为${role === 'admin' ? '管理员' : '普通用户'}`)
  } catch (error) {
    showToast(error instanceof Error ? error.message : '角色更新失败')
    await openAdmin()
  }
}

const createReset = async (user: User) => {
  try {
    const response = await api<{ token: string }>(`/api/v1/admin/users/${user.id}/password-reset`, { method: 'POST' })
    resetLink.value = `${location.origin}/#reset=${response.token}`
    await copyText(resetLink.value, '重置链接已复制，30 分钟内有效')
  } catch (error) {
    showToast(error instanceof Error ? error.message : '重置链接生成失败')
  }
}

const changePassword = async () => {
  try {
    await api('/api/v1/auth/password', { method: 'POST', body: JSON.stringify(passwordForm.value) })
    passwordOpen.value = false
    passwordForm.value = { current_password: '', new_password: '' }
    showToast('密码已修改，请重新登录')
    await boot()
  } catch (error) {
    showToast(error instanceof Error ? error.message : '密码修改失败')
  }
}

const mediaRows = computed(() => {
  if (!selected.value) return []
  const media = selected.value.latest_quality?.media
  if (media && Object.keys(media).length) {
    return Object.entries(media).map(([name, raw]) => ({ name, status: typeof raw === 'object' && raw !== null && 'status' in raw ? String((raw as Record<string, unknown>).status) : String(raw) }))
  }
  return [{ name: 'Netflix', status: selected.value.netflix }, { name: 'ChatGPT', status: selected.value.chatgpt }]
})

const scoreRows = computed(() => Object.entries(selected.value?.latest_quality?.scores ?? {}).map(([name, value]) => ({ name, value: typeof value === 'number' ? value : Number(value) || 0 })))
const factorRows = computed(() => Object.entries(selected.value?.latest_quality?.factors ?? {}).slice(0, 10).map(([name, value]) => ({ name, value: typeof value === 'boolean' ? (value ? '是' : '否') : String(value) })))
const toggleDark = () => { dark.value = !dark.value; document.documentElement.dataset.theme = dark.value ? 'dark' : 'light'; localStorage.setItem('detective-chicken-theme', dark.value ? 'dark' : 'light'); nextTick(() => window.dispatchEvent(new Event('resize'))) }
const closeOverlay = (event: KeyboardEvent) => { if (event.key === 'Escape') { selected.value = null; enrollOpen.value = false; adminOpen.value = false; passwordOpen.value = false; menuOpen.value = false } }

onMounted(() => {
  dark.value = localStorage.getItem('detective-chicken-theme') === 'dark'
  document.documentElement.dataset.theme = dark.value ? 'dark' : 'light'
  const match = location.hash.match(/^#reset=(.+)$/)
  if (match) resetToken.value = decodeURIComponent(match[1])
  window.addEventListener('keydown', closeOverlay)
  boot()
})
onBeforeUnmount(() => window.removeEventListener('keydown', closeOverlay))
</script>

<template>
  <div v-if="!auth || !auth.authenticated" class="auth-shell">
    <div class="auth-brand"><div class="brand-mark">探</div><div><strong>鸡探长</strong><span>DETECTIVE CHICKEN</span></div></div>
    <section class="auth-panel" aria-labelledby="auth-title">
      <template v-if="resetToken">
        <div class="auth-heading"><KeyRound :size="22" /><div><h1 id="auth-title">重置密码</h1><p>重置链接只能使用一次，密码至少 10 位。</p></div></div>
        <form @submit.prevent="completeReset"><label>新密码<input v-model="resetPassword" type="password" minlength="10" required autocomplete="new-password" /></label><p v-if="authError" class="form-error">{{ authError }}</p><button class="primary-btn auth-submit" :disabled="authBusy">确认重置</button></form>
      </template>
      <template v-else>
        <div class="auth-heading"><ShieldAlert :size="22" /><div><h1 id="auth-title">{{ auth?.settings.bootstrapped ? '登录控制台' : '创建首位管理员' }}</h1><p>{{ auth?.settings.bootstrapped ? '管理 VPS IP 质量、解锁能力与 Agent。' : '第一个注册账户将自动获得管理员权限。' }}</p></div></div>
        <div v-if="auth?.settings.bootstrapped && auth.settings.registration_enabled" class="segmented"><button :class="{ active: authMode === 'login' }" @click="authMode = 'login'">登录</button><button :class="{ active: authMode === 'register' }" @click="authMode = 'register'">注册</button></div>
        <form @submit.prevent="submitAuth">
          <label>用户名<input v-model="authForm.username" minlength="3" maxlength="64" required autocomplete="username" placeholder="例如 flanker" /></label>
          <label v-if="authMode === 'register'">显示名称<input v-model="authForm.display_name" maxlength="64" autocomplete="name" placeholder="控制台显示名称" /></label>
          <label>密码<input v-model="authForm.password" type="password" minlength="10" maxlength="128" required :autocomplete="authMode === 'login' ? 'current-password' : 'new-password'" /></label>
          <p v-if="authError" class="form-error">{{ authError }}</p><p v-if="fatalError" class="form-error">{{ fatalError }}</p>
          <button class="primary-btn auth-submit" :disabled="authBusy">{{ authMode === 'login' ? '登录' : '创建账户' }}</button>
        </form>
        <p v-if="auth?.settings.bootstrapped && !auth.settings.registration_enabled" class="auth-footnote">当前实例已关闭新用户注册，请联系管理员开放。</p>
      </template>
    </section>
  </div>

  <div v-else class="app-shell">
    <a class="skip-link" href="#main-content">跳到主要内容</a>
    <aside class="sidebar" :class="{ open: menuOpen }">
      <div class="brand"><div class="brand-mark">探</div><div><strong>鸡探长</strong><span>DETECTIVE CHICKEN</span></div><button class="icon-btn mobile-close" aria-label="关闭菜单" @click="menuOpen = false"><X :size="19" /></button></div>
      <nav aria-label="主导航"><a class="nav-item active" href="#dashboard"><LayoutDashboard :size="18" /><span>质量总览</span></a><a class="nav-item" href="#fleet"><Server :size="18" /><span>节点资产</span><small>{{ data.stats.total ?? 0 }}</small></a><a class="nav-item" href="#alerts"><Bell :size="18" /><span>告警中心</span><small class="danger-count">{{ data.alerts.length }}</small></a></nav>
      <div class="sidebar-section">账户与系统</div>
      <nav aria-label="管理导航"><button class="nav-item nav-button" @click="passwordOpen = true"><KeyRound :size="18" /><span>修改密码</span></button><button v-if="auth.user?.role === 'admin'" class="nav-item nav-button" @click="openAdmin"><UserCog :size="18" /><span>用户管理</span></button><button class="nav-item nav-button" @click="logout"><LogOut :size="18" /><span>退出登录</span></button></nav>
      <div class="sidebar-foot"><div class="health-dot"></div><div><strong>{{ auth.user?.display_name }}</strong><span>{{ auth.user?.role === 'admin' ? 'ADMINISTRATOR' : 'MEMBER' }}</span></div></div>
    </aside>
    <button v-if="menuOpen" class="sidebar-scrim" aria-label="关闭菜单" @click="menuOpen = false"></button>

    <main id="main-content" class="main-content">
      <header class="topbar"><div class="page-heading"><button class="icon-btn menu-btn" aria-label="打开菜单" @click="menuOpen = true"><Menu :size="20" /></button><div><h1>小鸡质量总览</h1><p>优先查看 IP 质量、流媒体与 AI 解锁状态</p></div></div><div class="top-actions"><span class="env-badge"><i></i>实时数据</span><button class="icon-btn" title="切换主题" @click="toggleDark"><Sun v-if="dark" :size="18" /><Moon v-else :size="18" /></button><button class="icon-btn" title="刷新" @click="loadDashboard(true)"><RefreshCw :size="18" :class="{ spinning: refreshing }" /></button><button class="primary-btn" @click="enrollOpen = true"><Plus :size="17" />添加 VPS</button></div></header>
      <div v-if="loading" class="loading-line"></div>
      <div v-if="fatalError" class="page-error"><AlertTriangle :size="17" />{{ fatalError }}</div>

      <section id="dashboard" class="kpi-grid" aria-label="关键指标"><article v-for="card in statCards" :key="card.key" class="kpi-card" :class="`tone-${card.tone}`"><div class="kpi-icon"><component :is="card.icon" :size="18" /></div><div><span>{{ card.label }}</span><strong>{{ card.value }}</strong><small>{{ card.hint }}</small></div></article></section>

      <section class="quality-grid">
        <div class="panel ranking-panel"><div class="panel-head"><div><h2>小鸡质量排行榜</h2><p>低风险优先，同分时按解锁数量排序</p></div><CircleGauge :size="19" /></div><div v-if="data.rankings.length" class="ranking-list"><button v-for="item in data.rankings.slice(0, 8)" :key="item.node_id" @click="openNode(data.nodes.find((n) => n.id === item.node_id)!)"><span class="rank" :class="{ podium: item.rank <= 3 }">{{ item.rank }}</span><span class="ranking-name"><strong>{{ item.name }}</strong><small>{{ item.provider }} · {{ item.region }}</small></span><span class="unlock-count">{{ item.unlocks }}/2<small>解锁</small></span><strong :class="riskClass(item.risk)">{{ item.quality }}</strong><ChevronRight :size="16" /></button></div><div v-else class="empty-state compact"><Server :size="22" /><strong>还没有节点</strong><span>添加第一台 VPS 后生成质量排名</span></div></div>
        <div class="panel unlock-panel"><div class="panel-head"><div><h2>解锁能力矩阵</h2><p>快速定位流媒体与 AI 服务限制</p></div><Cloud :size="19" /></div><div v-if="data.nodes.length" class="unlock-list"><button v-for="node in data.nodes.slice(0, 8)" :key="node.id" @click="openNode(node)"><span><strong>{{ node.name }}</strong><small>{{ node.masked_ip }}</small></span><StatusBadge :value="node.netflix" kind="media" /><StatusBadge :value="node.chatgpt" kind="media" /><ChevronRight :size="16" /></button></div><div v-else class="empty-state compact"><Cloud :size="22" /><strong>暂无解锁数据</strong><span>首次扫描完成后在这里展示</span></div></div>
      </section>

      <section class="analytics-grid"><div class="panel trend-panel"><div class="panel-head"><div><h2>风险与信誉趋势</h2><p>{{ data.nodes[0]?.name ?? '等待节点' }} · 最近采样</p></div><span class="trend-change">{{ data.nodes[0]?.risk ?? 0 }}<small>当前风险</small></span></div><TrendChart :points="data.trend" /><div class="chart-foot"><span><i class="anomaly-dot"></i>综合风险 / IPQS / Scamalytics</span><span>更新于 {{ relative(data.generated_at) }}</span></div></div><div id="alerts" class="panel alert-panel"><div class="panel-head"><div><h2>近期告警</h2><p>变化型规则优先</p></div><Bell :size="18" /></div><div v-if="data.alerts.length" class="alert-list"><button v-for="alert in data.alerts.slice(0, 4)" :key="alert.id" class="alert-row" @click="openNode(data.nodes.find((n) => n.id === alert.node_id)!)"><span class="alert-mark" :class="alert.severity"><ShieldAlert v-if="alert.severity === 'critical'" :size="16" /><AlertTriangle v-else :size="16" /></span><span class="alert-copy"><strong>{{ alert.title }}</strong><span>{{ alert.node_name }} · {{ alert.detail }}</span><small>{{ relative(alert.created_at) }}</small></span><ChevronRight :size="16" /></button></div><div v-else class="empty-state compact"><Activity :size="22" /><strong>当前没有告警</strong><span>节点状态保持稳定</span></div></div></section>

      <section id="fleet" class="panel fleet-panel"><div class="fleet-head"><div><h2>节点资产</h2><p>{{ filteredNodes.length }} 个结果，默认隐藏 IP 后两段</p></div><div class="fleet-tools"><label class="search-field"><Search :size="16" /><input v-model="search" type="search" placeholder="搜索节点、地区或 ASN" /></label><label class="filter-field"><ListFilter :size="16" /><select v-model="filter"><option value="all">全部状态</option><option value="online">在线</option><option value="warning">注意</option><option value="alert">告警</option><option value="offline">离线</option><option value="pending">待接入</option></select></label></div></div><div class="table-wrap"><table><thead><tr><th>节点</th><th>IP 地址</th><th>ASN / 地区</th><th>质量风险</th><th>Netflix</th><th>ChatGPT</th><th>DNSBL</th><th>上次上报</th><th>状态</th><th><span class="sr-only">操作</span></th></tr></thead><tbody><tr v-for="node in filteredNodes" :key="node.id" tabindex="0" @click="openNode(node)" @keydown.enter="openNode(node)"><td><div class="node-name"><span class="country-code">{{ node.country_code || '--' }}</span><div><strong>{{ node.name }}</strong><small>{{ node.provider || '未标记服务商' }}</small></div></div></td><td><code>{{ node.masked_ip }}</code><small class="family">IPv{{ node.family || 4 }}</small></td><td><strong class="asn">AS{{ node.asn || '—' }}</strong><small>{{ node.region || node.organization }}</small></td><td><div class="risk-cell"><strong :class="riskClass(node.risk)">{{ node.risk }}</strong><span>{{ riskLabel(node.risk) }}风险</span></div></td><td><StatusBadge :value="node.netflix" kind="media" /></td><td><StatusBadge :value="node.chatgpt" kind="media" /></td><td><span class="dnsbl" :class="{ hit: node.dnsbl > 0 }">{{ node.dnsbl }}</span></td><td><span class="last-seen">{{ relative(node.last_seen) }}</span></td><td><StatusBadge :value="node.status" /></td><td><button class="row-action" title="查看详情" @click.stop="openNode(node)"><ChevronRight :size="17" /></button></td></tr></tbody></table><div v-if="filteredNodes.length === 0" class="empty-state"><Search :size="24" /><strong>{{ data.nodes.length ? '没有匹配的节点' : '还没有纳管 VPS' }}</strong><span>{{ data.nodes.length ? '调整搜索词或状态筛选' : '点击右上角添加 VPS 生成兼容安装脚本' }}</span></div></div></section>
      <footer><span>鸡探长 · 数据按账户隔离，IP 默认隐藏后两段</span><span>心跳 2 min · 完整扫描 6 h + jitter</span></footer>
    </main>

    <Transition name="drawer"><div v-if="selected || detailLoading" class="drawer-backdrop" @click.self="selected = null"><aside class="detail-drawer wide" aria-label="节点详情"><div v-if="detailLoading && !selected" class="drawer-loading">正在读取节点报告…</div><template v-if="selected"><div class="drawer-head"><div><span class="country-code large">{{ selected.country_code || '--' }}</span><div><h2>{{ selected.name }}</h2><p>{{ selected.provider }} · {{ selected.region }}</p></div></div><button class="icon-btn" aria-label="关闭详情" @click="selected = null"><X :size="19" /></button></div><div class="drawer-summary"><div><span>综合风险</span><strong :class="riskClass(selected.risk)">{{ selected.risk }}</strong></div><div><span>当前状态</span><StatusBadge :value="selected.status" /></div><div><span>质量评分</span><strong>{{ 100 - selected.risk }}</strong></div></div><div class="detail-section"><div class="section-title"><h3>网络身份</h3><button v-if="selected.can_view_full_ip" class="text-btn" @click="toggleFullIP"><EyeOff v-if="showFullIP" :size="15" /><Eye v-else :size="15" />{{ showFullIP ? '隐藏完整 IP' : '显示完整 IP' }}</button></div><dl><div><dt>公网 IP</dt><dd><code>{{ selected.ip_address || selected.masked_ip }}</code></dd></div><div><dt>ASN</dt><dd>AS{{ selected.asn || '—' }} · {{ selected.organization || '等待检测' }}</dd></div><div><dt>网络族</dt><dd>IPv{{ selected.family || 4 }}</dd></div><div><dt>上次扫描</dt><dd>{{ relative(selected.last_scan) }}</dd></div></dl></div><div class="detail-section"><h3>解锁能力</h3><div class="media-detail-grid"><div v-for="media in mediaRows" :key="media.name"><span>{{ media.name }}</span><StatusBadge :value="media.status" kind="media" /></div></div></div><div v-if="scoreRows.length" class="detail-section"><h3>IP 质量评分</h3><div class="score-grid"><div v-for="score in scoreRows" :key="score.name"><span>{{ score.name }}</span><strong :class="riskClass(score.value)">{{ score.value }}</strong><i><b :style="{ width: `${Math.min(100, score.value)}%` }"></b></i></div></div></div><div v-if="factorRows.length" class="detail-section"><h3>风险因子</h3><dl><div v-for="factor in factorRows" :key="factor.name"><dt>{{ factor.name }}</dt><dd>{{ factor.value }}</dd></div></dl></div><div class="detail-section"><h3>风险趋势</h3><TrendChart :points="selected.series" /></div><div class="detail-section"><h3>最近变化与告警</h3><div v-if="selected.alerts.length" class="change-list"><div v-for="alert in selected.alerts" :key="alert.id"><AlertTriangle :size="17" /><span><strong>{{ alert.title }}</strong><small>{{ alert.detail }} · {{ relative(alert.created_at) }}</small></span></div></div><div v-else class="quiet-state"><Activity :size="17" />未发现显著变化</div></div><div class="drawer-actions"><button class="secondary-btn" @click="selected = null">关闭</button><button class="primary-btn" @click="scan(selected)"><ScanLine :size="17" />立即扫描</button></div></template></aside></div></Transition>

    <Transition name="modal"><div v-if="enrollOpen" class="modal-backdrop" @click.self="closeEnrollment"><section class="modal large-modal" role="dialog" aria-modal="true"><div class="modal-head"><div><h2>添加 VPS</h2><p>根据目标系统、架构与运行环境生成安装脚本</p></div><button class="icon-btn" @click="closeEnrollment"><X :size="19" /></button></div><form v-if="!enrollment" @submit.prevent="createEnrollment"><label>节点名称<input v-model="enrollForm.name" required placeholder="例如 HK-CMI-01" /></label><div class="form-grid"><label>服务商<input v-model="enrollForm.provider" placeholder="例如 OVH" /></label><label>地区标签<input v-model="enrollForm.region" placeholder="例如 加拿大" /></label></div><div class="form-grid three"><label>系统<select v-model="enrollForm.os_family"><option value="auto">自动识别</option><option value="debian">Debian / Ubuntu</option><option value="alpine">Alpine</option><option value="rhel">RHEL / Rocky</option><option value="arch">Arch Linux</option></select></label><label>环境<select v-model="enrollForm.platform"><option value="auto">自动识别</option><option value="pve">PVE 宿主机</option><option value="baremetal">独立服务器</option><option value="lxc">LXC</option><option value="docker">Docker</option><option value="podman">Podman</option><option value="incus">Incus</option></select></label><label>架构<select v-model="enrollForm.arch"><option value="auto">自动识别</option><option value="amd64">AMD64</option><option value="arm64">ARM64</option><option value="armv7">ARMv7</option></select></label></div><div class="form-note"><ShieldAlert :size="17" /><span>凭证 10 分钟内有效且只能使用一次。脚本会优先配置 systemd，否则使用 OpenRC/cron；无 init 容器会启动兼容循环。</span></div><div class="modal-actions"><button type="button" class="secondary-btn" @click="closeEnrollment">取消</button><button class="primary-btn" type="submit">生成安装命令<ChevronRight :size="17" /></button></div></form><div v-else class="enrollment-result"><div class="step-status"><span>1</span><div><strong>兼容安装脚本已生成</strong><small>{{ new Date(enrollment.expires_at).toLocaleTimeString('zh-CN') }} 前有效</small></div></div><label>在目标 VPS 上执行</label><div class="command-box"><code>{{ enrollment.install_command }}</code><button class="icon-btn" title="复制" @click="copyText(enrollment.install_command, '安装命令已复制')"><Copy :size="17" /></button></div><div class="installation-steps"><div class="done"><i></i>环境检测</div><div><i></i>Agent 注册</div><div><i></i>心跳启动</div><div><i></i>首次检测</div></div><div class="modal-actions"><button class="secondary-btn" @click="enrollment = null">返回</button><button class="primary-btn" @click="closeEnrollment">完成</button></div></div></section></div></Transition>

    <Transition name="modal"><div v-if="adminOpen" class="modal-backdrop" @click.self="adminOpen = false"><section class="modal admin-modal"><div class="modal-head"><div><h2>用户与权限</h2><p>管理注册开关、账户角色和密码重置</p></div><button class="icon-btn" @click="adminOpen = false"><X :size="19" /></button></div><div class="admin-settings"><label class="toggle-row"><span><strong>允许新用户注册</strong><small>开启后，新注册账户默认为普通用户</small></span><input v-model="registrationEnabled" type="checkbox" @change="saveRegistration" /></label></div><div class="user-table"><div v-for="user in users" :key="user.id" class="user-row"><span class="user-avatar">{{ user.display_name.slice(0, 1).toUpperCase() }}</span><span><strong>{{ user.display_name }}</strong><small>@{{ user.username }} · {{ relative(user.created_at) }}</small></span><select :value="user.role" :disabled="user.id === auth.user?.id" @change="updateRole(user, ($event.target as HTMLSelectElement).value)"><option value="user">普通用户</option><option value="admin">管理员</option></select><button class="icon-btn" title="生成密码重置链接" @click="createReset(user)"><KeyRound :size="16" /></button></div></div><div v-if="resetLink" class="reset-result"><span>密码重置链接（30 分钟有效）</span><code>{{ resetLink }}</code><button class="secondary-btn" @click="copyText(resetLink, '重置链接已复制')"><Copy :size="15" />复制</button></div></section></div></Transition>

    <Transition name="modal"><div v-if="passwordOpen" class="modal-backdrop" @click.self="passwordOpen = false"><section class="modal"><div class="modal-head"><div><h2>修改密码</h2><p>修改后其他登录会话会被撤销</p></div><button class="icon-btn" @click="passwordOpen = false"><X :size="19" /></button></div><form @submit.prevent="changePassword"><label>当前密码<input v-model="passwordForm.current_password" type="password" required autocomplete="current-password" /></label><label>新密码<input v-model="passwordForm.new_password" type="password" minlength="10" required autocomplete="new-password" /></label><div class="modal-actions"><button type="button" class="secondary-btn" @click="passwordOpen = false">取消</button><button class="primary-btn">更新密码</button></div></form></section></div></Transition>
    <Transition name="toast"><div v-if="toast" class="toast" role="status"><Activity :size="17" />{{ toast }}</div></Transition>
  </div>
</template>
