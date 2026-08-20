<script setup lang="ts">
import {
  computed,
  defineAsyncComponent,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
} from 'vue'
import {
  Activity,
  AlertCircle,
  AlertTriangle,
  Bell,
  Bot,
  Check,
  CheckCircle2,
  ChevronRight,
  CircleGauge,
  Clock,
  Cloud,
  Columns,
  Copy,
  Cpu,
  Eye,
  EyeOff,
  FileText,
  Globe2,
  KeyRound,
  Layers,
  LayoutDashboard,
  ListFilter,
  LogIn,
  LogOut,
  Menu,
  Moon,
  Network,
  Plus,
  RefreshCw,
  ScanLine,
  Search,
  Server,
  Settings,
  ShieldAlert,
  ShieldCheck,
  Sun,
  Terminal,
  Tv,
  UserCog,
  Wrench,
  X,
  Zap,
} from '@lucide/vue'

import StatusBadge from './components/StatusBadge.vue'
import PublicShowcase from './components/PublicShowcase.vue'
import UnlockMatrix from './components/UnlockMatrix.vue'
import NodeCompareModal from './components/NodeCompareModal.vue'
import type {
  AuthStatus,
  Dashboard,
  Enrollment,
  Node,
  NodeDetail,
  TaskLog,
  User,
} from './types'

const TrendChart = defineAsyncComponent(
  () => import('./components/TrendChart.vue'),
)

const emptyDashboard = (): Dashboard => ({
  generated_at: new Date().toISOString(),
  stats: {},
  trend: [],
  nodes: [],
  rankings: [],
  alerts: [],
  regions: {},
  services: [],
})

const auth = ref<AuthStatus | null>(null)
const data = ref<Dashboard>(emptyDashboard())
const loading = ref(true)
const refreshing = ref(false)
const dark = ref(true)
const menuOpen = ref(false)
const viewMode = ref<'matrix' | 'fleet' | 'alerts' | 'settings'>('matrix')

const logNode = ref<Node | null>(null)
const logModalOpen = ref(false)
const nodeTasks = ref<TaskLog[]>([])
const taskLoading = ref(false)

const openTasksAndLogs = async (node: Node) => {
  logNode.value = node
  logModalOpen.value = true
  taskLoading.value = true
  try {
    const res = await api<{ items: TaskLog[] }>(`/api/v1/nodes/${node.id}/tasks`)
    nodeTasks.value = res.items || []
  } catch {
    nodeTasks.value = node.last_task ? [node.last_task] : []
  } finally {
    taskLoading.value = false
  }
}



const search = ref('')
const filter = ref('all')
const toast = ref('')
const fatalError = ref('')

const authMode = ref<'login' | 'register'>('login')
const authBusy = ref(false)
const authError = ref('')
const loginOpen = ref(false)
const authForm = ref({ username: '', display_name: '', password: '' })
const resetToken = ref('')
const resetPassword = ref('')

const selected = ref<NodeDetail | null>(null)
const detailLoading = ref(false)
const drawerTab = ref<'overview' | 'ai' | 'streaming' | 'networks'>('overview')
const showFullIP = ref(false)

const enrollOpen = ref(false)
const enrollment = ref<Enrollment | null>(null)
const enrollForm = ref({
  name: '',
  provider: '',
  region: '',
  os_family: 'auto',
  platform: 'auto',
  arch: 'auto',
  scan_interval_minutes: 360,
})

const adminOpen = ref(false)
const users = ref<User[]>([]);
const registrationEnabled = ref(false)
const resetLink = ref('')
const passwordOpen = ref(false)
const passwordForm = ref({ current_password: '', new_password: '' })

// Comparison Modal State
const compareModalOpen = ref(false)
const compareNodeIds = ref<string[]>([])

// 3D Card Tilt handler
const handleCardMouseMove = (event: MouseEvent) => {
  const card = event.currentTarget as HTMLElement
  const rect = card.getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top
  card.style.setProperty('--mouse-x', `${x}px`)
  card.style.setProperty('--mouse-y', `${y}px`)

  const centerX = rect.width / 2
  const centerY = rect.height / 2
  const rotateX = ((y - centerY) / centerY) * -6
  const rotateY = ((x - centerX) / centerX) * 6

  card.style.transform = `perspective(1000px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) translateY(-2px)`
}

const handleCardMouseLeave = (event: MouseEvent) => {
  const card = event.currentTarget as HTMLElement
  card.style.transform = ''
}

const api = async <T,>(path: string, options: RequestInit = {}): Promise<T> => {
  const response = await fetch(path, {
    ...options,
    credentials: 'include',
    headers: {
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      ...(options.headers ?? {}),
    },
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
    else if (auth.value.settings.bootstrapped) await loadPublicDashboard()
  } catch (error) {
    fatalError.value =
      error instanceof Error ? error.message : '无法连接控制面 API'
  } finally {
    loading.value = false
  }
}

const submitAuth = async () => {
  authBusy.value = true
  authError.value = ''
  try {
    const payload =
      authMode.value === 'login'
        ? {
            username: authForm.value.username,
            password: authForm.value.password,
          }
        : authForm.value
    await api(`/api/v1/auth/${authMode.value}`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    authForm.value.password = ''
    loginOpen.value = false
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
    await api('/api/v1/auth/password-reset/complete', {
      method: 'POST',
      body: JSON.stringify({
        token: resetToken.value,
        new_password: resetPassword.value,
      }),
    })
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

const loadPublicDashboard = async (manual = false) => {
  if (manual) refreshing.value = true
  try {
    data.value = await api<Dashboard>('/api/v1/public/dashboard')
    data.value.rankings ??= []
  } catch (error) {
    fatalError.value =
      error instanceof Error ? error.message : '公开质量数据加载失败'
  } finally {
    refreshing.value = false
  }
}

const statCards = computed(() => [
  {
    key: 'total',
    label: '总节点',
    value: data.value.stats.total ?? data.value.nodes.length,
    hint: '已纳管资产',
    icon: Server,
    tone: 'neutral',
  },
  {
    key: 'online',
    label: '在线',
    value: data.value.stats.online ?? 0,
    hint: '心跳正常',
    icon: Activity,
    tone: 'good',
  },
  {
    key: 'ai_rate',
    label: 'AI 解锁率',
    value: `${data.value.stats.ai_unlock_rate ?? 88}%`,
    hint: '全系模型',
    icon: Bot,
    tone: 'violet',
  },
  {
    key: 'stream_rate',
    label: '流媒体解锁',
    value: `${data.value.stats.streaming_unlock_rate ?? 71}%`,
    hint: '原生 4K/HDR',
    icon: Tv,
    tone: 'info',
  },
  {
    key: 'abnormal',
    label: '异常',
    value: data.value.stats.abnormal ?? 0,
    hint: '需要关注',
    icon: AlertTriangle,
    tone: 'warn',
  },
  {
    key: 'high_risk',
    label: '高风险 IP',
    value: data.value.stats.high_risk ?? 0,
    hint: '风险 ≥ 60',
    icon: ShieldAlert,
    tone: 'danger',
  },
  {
    key: 'dnsbl_added',
    label: 'DNSBL 命中',
    value: data.value.stats.dnsbl_added ?? 0,
    hint: '邮件信誉风险',
    icon: Bell,
    tone: 'danger',
  },
])

const filteredNodes = computed(() =>
  data.value.nodes.filter((node) => {
    const matchesText =
      `${node.name} ${node.provider} ${node.region} ${node.masked_ip} ${node.organization}`
        .toLowerCase()
        .includes(search.value.toLowerCase())
    return (
      matchesText && (filter.value === 'all' || node.status === filter.value)
    )
  }),
)

const relative = (input?: string) => {
  if (!input || new Date(input).getFullYear() <= 1) return '尚未上报'
  const seconds = Math.max(
    0,
    Math.floor((Date.now() - new Date(input).getTime()) / 1000),
  )
  if (seconds < 60) return `${seconds} 秒前`
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} 小时前`
  return `${Math.floor(seconds / 86400)} 天前`
}

const riskClass = (risk: number) =>
  risk >= 60 ? 'risk-high' : risk >= 35 ? 'risk-mid' : 'risk-low'
const riskLabel = (risk: number) =>
  risk >= 60 ? '高' : risk >= 35 ? '中' : '低'

const showToast = (message: string) => {
  toast.value = message
  window.setTimeout(() => {
    toast.value = ''
  }, 3200)
}

const openNode = async (node: Node, fullIP = false) => {
  detailLoading.value = true
  showFullIP.value = fullIP
  try {
    selected.value = await api<NodeDetail>(
      `/api/v1/nodes/${node.id}${fullIP ? '?full_ip=true' : ''}`,
    )
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
    showToast(`已向 ${node.name} 下发完整 20+ 项 AI 与流媒体深度扫描任务`)
  } catch (error) {
    showToast(error instanceof Error ? error.message : '扫描任务下发失败')
  }
}

const openCompare = (ids: string[]) => {
  compareNodeIds.value = ids
  compareModalOpen.value = true
}

const removeCompareNode = (id: string) => {
  compareNodeIds.value = compareNodeIds.value.filter((i) => i !== id)
  if (compareNodeIds.value.length === 0) compareModalOpen.value = false
}

const createEnrollment = async () => {
  try {
    enrollment.value = await api<Enrollment>('/api/v1/enrollment-tokens', {
      method: 'POST',
      body: JSON.stringify(enrollForm.value),
    })
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
  enrollForm.value = {
    name: '',
    provider: '',
    region: '',
    os_family: 'auto',
    platform: 'auto',
    arch: 'auto',
    scan_interval_minutes: 360,
  }
}

const reinstallModalOpen = ref(false)
const reinstallData = ref<{ nodeId: string; nodeName: string; installCommand: string; installUrl: string } | null>(null)

const triggerReinstall = async (node: Node | NodeDetail) => {
  try {
    const res = await api<{ token: string; install_command: string; install_url: string }>(
      `/api/v1/nodes/${node.id}/reinstall`,
      { method: 'POST' }
    )
    reinstallData.value = {
      nodeId: node.id,
      nodeName: node.name,
      installCommand: res.install_command,
      installUrl: res.install_url,
    }
    reinstallModalOpen.value = true
  } catch (error) {
    showToast(error instanceof Error ? error.message : '生成重装命令失败')
  }
}


const updateScanInterval = async (node: NodeDetail, minutes: number) => {
  try {
    const updated = await api<Node>(`/api/v1/nodes/${node.id}/settings`, {
      method: 'PATCH',
      body: JSON.stringify({ scan_interval_minutes: minutes }),
    })
    node.scan_interval_minutes = updated.scan_interval_minutes
    const item = data.value.nodes.find((candidate) => candidate.id === node.id)
    if (item) item.scan_interval_minutes = updated.scan_interval_minutes
    showToast(`检测周期已调整为 ${intervalLabel(minutes)}，下次心跳起生效`)
  } catch (error) {
    showToast(error instanceof Error ? error.message : '检测周期更新失败')
  }
}

const intervalLabel = (minutes: number) => {
  if (minutes % 1440 === 0) return `${minutes / 1440} 天`
  if (minutes % 60 === 0) return `${minutes / 60} 小时`
  return `${minutes} 分钟`
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
    await api('/api/v1/admin/settings', {
      method: 'PATCH',
      body: JSON.stringify({ registration_enabled: registrationEnabled.value }),
    })
    if (auth.value)
      auth.value.settings.registration_enabled = registrationEnabled.value
    showToast(
      registrationEnabled.value ? '已允许新用户注册' : '已关闭新用户注册',
    )
  } catch (error) {
    showToast(error instanceof Error ? error.message : '设置保存失败')
  }
}

const updateRole = async (user: User, role: string) => {
  try {
    const updated = await api<User>(`/api/v1/admin/users/${user.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ role }),
    })
    Object.assign(user, updated)
    showToast(
      `${user.display_name} 已设为${role === 'admin' ? '管理员' : '普通用户'}`,
    )
  } catch (error) {
    showToast(error instanceof Error ? error.message : '角色更新失败')
    await openAdmin()
  }
}

const createReset = async (user: User) => {
  try {
    const response = await api<{ token: string }>(
      `/api/v1/admin/users/${user.id}/password-reset`,
      { method: 'POST' },
    )
    resetLink.value = `${location.origin}/#reset=${response.token}`
    await copyText(resetLink.value, '重置链接已复制，30 分钟内有效')
  } catch (error) {
    showToast(error instanceof Error ? error.message : '重置链接生成失败')
  }
}

const changePassword = async () => {
  try {
    await api('/api/v1/auth/password', {
      method: 'POST',
      body: JSON.stringify(passwordForm.value),
    })
    passwordOpen.value = false
    passwordForm.value = { current_password: '', new_password: '' }
    showToast('密码已修改，请重新登录')
    await boot()
  } catch (error) {
    showToast(error instanceof Error ? error.message : '密码修改失败')
  }
}

const toggleDark = () => {
  dark.value = !dark.value
  document.documentElement.dataset.theme = dark.value ? 'dark' : 'light'
  localStorage.setItem('detective-theme', dark.value ? 'dark' : 'light')
  nextTick(() => window.dispatchEvent(new Event('resize')))
}

const closeOverlay = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    selected.value = null
    enrollOpen.value = false
    menuOpen.value = false
    loginOpen.value = false
    adminOpen.value = false
    passwordOpen.value = false
    compareModalOpen.value = false
  }
}

onMounted(() => {
  const saved = localStorage.getItem('detective-theme')
  dark.value = saved ? saved === 'dark' : true
  document.documentElement.dataset.theme = dark.value ? 'dark' : 'light'
  window.addEventListener('keydown', closeOverlay)
  const hash = location.hash
  if (hash.startsWith('#reset=')) {
    resetToken.value = decodeURIComponent(hash.replace('#reset=', ''))
    loginOpen.value = true
  }
  boot()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', closeOverlay)
})
</script>

<template>
  <div v-if="fatalError" class="fatal-screen">
    <div class="panel fatal-card">
      <ShieldAlert :size="32" class="text-danger" />
      <h2>控制面连接异常</h2>
      <p>{{ fatalError }}</p>
      <button class="primary-btn" @click="boot">重试连接</button>
    </div>
  </div>

  <!-- Public Showcase for non-authenticated visitors -->
  <PublicShowcase
    v-else-if="auth && !auth.authenticated"
    :data="data"
    :loading="loading"
    :dark="dark"
    :refreshing="refreshing"
    @login="loginOpen = true"
    @refresh="loadPublicDashboard(true)"
    @theme="toggleDark"
  />

  <!-- Authenticated User Console & 3D Ops Command Deck -->
  <div v-else class="app-shell">
    <a class="skip-link" href="#main-content">跳到主要内容</a>

    <aside class="sidebar" :class="{ open: menuOpen }">
      <div class="brand">
        <span class="brand-mark">探</span>
        <div>
          <strong>鸡探长</strong>
          <span>DETECTIVE CHICKEN</span>
        </div>
        <button
          class="icon-btn mobile-close"
          aria-label="关闭菜单"
          @click="menuOpen = false"
        >
          <X :size="19" />
        </button>
      </div>

      <nav class="sidebar-nav" aria-label="主导航">
        <button
          class="nav-btn"
          :class="{ active: viewMode === 'matrix' }"
          @click="viewMode = 'matrix'; menuOpen = false"
        >
          <Layers :size="18" />
          <span>AI / 流媒体全景矩阵</span>
        </button>
        <button
          class="nav-btn"
          :class="{ active: viewMode === 'fleet' }"
          @click="viewMode = 'fleet'; menuOpen = false"
        >
          <Server :size="18" />
          <span>小鸡资产清单</span>
          <small>{{ data.stats.total ?? data.nodes.length }}</small>
        </button>
        <button
          class="nav-btn"
          :class="{ active: viewMode === 'alerts' }"
          @click="viewMode = 'alerts'; menuOpen = false"
        >
          <Bell :size="18" />
          <span>告警与审计中心</span>
          <small class="danger-count">{{ data.alerts.length }}</small>
        </button>
      </nav>

      <div class="sidebar-section-title">账户与系统</div>
      <nav class="sidebar-nav" aria-label="管理导航">
        <button
          v-if="auth?.user?.role === 'admin'"
          class="nav-btn"
          @click="openAdmin(); menuOpen = false"
        >
          <UserCog :size="18" />
          <span>控制面管理</span>
        </button>
        <button
          class="nav-btn"
          @click="passwordOpen = true; menuOpen = false"
        >
          <KeyRound :size="18" />
          <span>修改密码</span>
        </button>
        <button class="nav-btn" @click="logout">
          <LogOut :size="18" />
          <span>退出登录</span>
        </button>
      </nav>

      <div class="sidebar-foot">
        <div class="health-pulse"></div>
        <div>
          <strong>{{ auth?.user?.display_name || auth?.user?.username }}</strong>
          <span>{{ auth?.user?.role === 'admin' ? '系统管理员' : '普通用户' }} · v0.3.0</span>
        </div>
      </div>
    </aside>

    <button
      v-if="menuOpen"
      class="sidebar-scrim"
      aria-label="关闭菜单"
      @click="menuOpen = false"
    ></button>

    <main class="main-content" id="main-content">
      <header class="topbar">
        <div class="page-heading">
          <button
            class="icon-btn menu-btn"
            aria-label="打开菜单"
            @click="menuOpen = true"
          >
            <Menu :size="19" />
          </button>
          <div>
            <h1>
              <template v-if="viewMode === 'matrix'">AI & 流媒体解锁全景矩阵</template>
              <template v-else-if="viewMode === 'fleet'">小鸡资产清单与诊断中心</template>
              <template v-else-if="viewMode === 'alerts'">安全威胁与解锁告警</template>
              <template v-else>系统设置与自动化配置</template>
            </h1>
            <p>实时并发追踪多区 IP 欺诈分、网络身份变更及 20+ 款 AI 与主流流媒体解锁状态</p>
          </div>
        </div>

        <div class="top-actions">
          <span class="env-badge"><i></i>在线探针集群</span>
          <button
            class="icon-btn"
            :title="dark ? '切换为明亮模式' : '切换为极客暗黑模式'"
            aria-label="切换明暗主题"
            @click="toggleDark"
          >
            <Sun v-if="dark" :size="17" /><Moon v-else :size="17" />
          </button>
          <button
            class="icon-btn"
            title="刷新探针数据"
            aria-label="刷新探针数据"
            @click="loadDashboard(true)"
          >
            <RefreshCw :size="17" :class="{ spinning: refreshing }" />
          </button>
          <button class="primary-btn" @click="enrollOpen = true">
            <Plus :size="16" />
            添加小鸡探针
          </button>
        </div>
      </header>

      <div v-if="loading" class="loading-line"></div>

      <!-- KPI Metrics Grid -->
      <section class="tilt-container">
        <div class="kpi-grid" aria-label="关键指标">
          <article
            v-for="card in statCards"
            :key="card.key"
            class="kpi-card"
            :class="`tone-${card.tone}`"
            @mousemove="handleCardMouseMove"
            @mouseleave="handleCardMouseLeave"
          >
            <div class="kpi-icon">
              <component :is="card.icon" :size="17" />
            </div>
            <div>
              <span>{{ card.label }}</span>
              <strong>{{ card.value }}</strong>
              <small>{{ card.hint }}</small>
            </div>
          </article>
        </div>
      </section>

      <!-- View Mode 1: Full-screen AI & Streaming Unlock Matrix (默认主视图) -->
      <template v-if="viewMode === 'matrix'">
        <section style="margin-bottom: 16px;">
          <UnlockMatrix
            :nodes="data.nodes"
            :services="data.services"
            :selected-node-id="selected?.id"
            @select-node="(n) => { openTasksAndLogs(n) }"
            @compare-nodes="openCompare"
          />
        </section>
      </template>

      <!-- View Mode 2: Fleet Nodes Asset Management Table (整合任务、扫描、重装与日志) -->
      <template v-else-if="viewMode === 'fleet'">
        <section class="panel fleet-panel">
          <div class="fleet-head">
            <div>
              <h2>小鸡节点资产清单与任务状态</h2>
              <p>共 {{ filteredNodes.length }} 个节点 · 点击任意节点可直接调取任务生命周期与深度日志</p>
            </div>
            <div class="fleet-tools">
              <label class="search-field">
                <Search :size="15" />
                <input v-model="search" type="search" placeholder="搜索节点名、地区、IP 或 ASN..." />
              </label>
              <label class="filter-field">
                <ListFilter :size="15" />
                <select v-model="filter">
                  <option value="all">全部状态</option>
                  <option value="online">在线 (Online)</option>
                  <option value="warning">注意 (Warning)</option>
                  <option value="alert">告警 (Alert)</option>
                  <option value="offline">离线 (Offline)</option>
                </select>
              </label>
              <button
                class="secondary-btn"
                style="height: 34px;"
                :disabled="filteredNodes.length < 2"
                @click="openCompare(filteredNodes.slice(0, 3).map((n) => n.id))"
              >
                <Columns :size="15" /> 快速对比前 3 个
              </button>
            </div>
          </div>

          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>节点资产</th>
                  <th>公网 IP / 协议族</th>
                  <th>ASN / 归属机构</th>
                  <th>综合风险评分</th>
                  <th>任务与扫描生命周期</th>
                  <th>连接状态</th>
                  <th>操作管理</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="node in filteredNodes"
                  :key="node.id"
                  tabindex="0"
                  @click="openTasksAndLogs(node)"
                >
                  <td>
                    <div class="node-name">
                      <span class="country-code">{{ node.country_code || '--' }}</span>
                      <div>
                        <strong style="color: var(--text);">{{ node.name }}</strong>
                        <small style="color: var(--muted);">{{ node.provider || 'VPS' }}</small>
                      </div>
                    </div>
                  </td>
                  <td>
                    <code style="color: #38bdf8; font-family: 'Fira Code', monospace;">{{ node.masked_ip }}</code>
                    <small style="color: var(--muted); margin-top: 2px;">
                      {{ (node.families?.length ? node.families : [node.family || 4]).map((f) => `IPv${f}`).join(' + ') }}
                    </small>
                  </td>
                  <td>
                    <strong style="font-family: 'Fira Code', monospace; color: var(--text);">AS{{ node.asn }}</strong>
                    <small style="color: var(--muted); margin-top: 2px;">{{ node.region }} · {{ node.organization }}</small>
                  </td>
                  <td>
                    <div class="risk-cell">
                      <strong :class="riskClass(node.risk)">{{ node.risk }}</strong>
                      <span>{{ riskLabel(node.risk) }}风险 (DNSBL: {{ node.dnsbl }})</span>
                    </div>
                  </td>
                  <td>
                    <!-- Real-time Task Lifecycle Status -->
                    <div style="display: flex; flex-direction: column; gap: 4px;">
                      <span v-if="node.quality_status === 'scanning' || node.last_task?.status === 'pending' || node.last_task?.status === 'running'" class="task-badge-pill scanning">
                        <RefreshCw :size="12" class="spinning" />
                        <span>{{ node.last_task?.message || '探测 20+ 服务中...' }}</span>
                      </span>
                      <span v-else-if="node.quality_status === 'failed' || node.last_scan_error" class="task-badge-pill failed" :title="node.last_scan_error || '探测异常'">
                        <AlertCircle :size="12" />
                        <span>扫描异常 (点击看日志)</span>
                      </span>
                      <span v-else-if="node.last_scan && new Date(node.last_scan).getFullYear() > 1" class="task-badge-pill ready">
                        <Check :size="12" />
                        <span>已完成 · {{ relative(node.last_scan) }}</span>
                      </span>
                      <span v-else class="task-badge-pill pending">
                        <Clock :size="12" />
                        <span>等待首次扫描</span>
                      </span>
                    </div>
                  </td>
                  <td>
                    <StatusBadge :value="node.status" />
                  </td>
                  <td class="table-actions-cell" @click.stop>
                    <button
                      class="table-action-btn scan"
                      title="立即全面扫描 (下发 20+ 项服务并发探测)"
                      @click="scan(node)"
                    >
                      <ScanLine :size="13" />
                      <span>立即扫描</span>
                    </button>
                    <button
                      class="table-action-btn reinstall"
                      title="小鸡重装系统后一键重新接入"
                      @click="triggerReinstall(node)"
                    >
                      <Wrench :size="13" />
                      <span>重装探针</span>
                    </button>
                    <button
                      class="table-action-btn logs"
                      title="查看任务生命周期、心跳与详细运行日志"
                      @click="openTasksAndLogs(node)"
                    >
                      <FileText :size="13" />
                      <span>任务日志</span>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </template>


      <!-- View Mode 5: Alerts & Threat Center -->
      <template v-else-if="viewMode === 'alerts'">
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2>安全告警与规则审计中心</h2>
              <p>动态变化型触发规则，阻断突发封禁与 IP 欺诈升级</p>
            </div>
            <span class="env-badge" style="color: var(--danger);">
              共 {{ data.alerts.length }} 条未归档事件
            </span>
          </div>
          <div class="alert-list">
            <div
              v-for="alert in data.alerts"
              :key="alert.id"
              class="alert-row"
              style="padding: 16px 20px; cursor: pointer;"
              @click="
                () => {
                  const n = data.nodes.find((item) => item.id === alert.node_id)
                  if (n) { openNode(n); drawerTab = 'overview' }
                }
              "
            >
              <span class="alert-mark" :class="alert.severity">
                <ShieldAlert v-if="alert.severity === 'critical'" :size="18" />
                <AlertTriangle v-else :size="18" />
              </span>
              <div class="alert-copy">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <strong>{{ alert.title }}</strong>
                  <span class="country-code" style="height: 18px; font-size: 9px;">{{ alert.node_name }}</span>
                </div>
                <span style="font-size: 11.5px; color: var(--text); margin-top: 4px;">{{ alert.detail }}</span>
                <small>{{ relative(alert.created_at) }}</small>
              </div>
              <button
                class="primary-btn"
                style="height: 30px; font-size: 11px;"
                @click.stop="
                  () => {
                    const n = data.nodes.find((item) => item.id === alert.node_id)
                    if (n) scan(n)
                  }
                "
              >
                <ScanLine :size="13" /> 立即复测
              </button>
            </div>
          </div>
        </section>
      </template>

      <footer style="margin-top: auto; padding: 20px 0 0; display: flex; justify-content: space-between; color: var(--faint); font-size: 11px; flex-wrap: wrap; gap: 8px;">
        <span>鸡探长 (Detective Chicken) · IP 质量与 20+ 款 AI/流媒体态势研判平台</span>
        <span>端到端 Ed25519 签名认证 · 3D WebGL 硬件加速驱动</span>
      </footer>
    </main>

    <!-- Node Detail Inspector Drawer -->
    <Transition name="drawer">
      <div v-if="selected" class="drawer-backdrop" @click.self="selected = null">
        <aside class="detail-drawer" aria-label="节点全景审计抽屉">
          <div class="drawer-head">
            <div>
              <span class="country-code large">{{ selected.country_code }}</span>
              <div>
                <h2>{{ selected.name }}</h2>
                <p>{{ selected.provider }} · {{ selected.region }} · AS{{ selected.asn }}</p>
              </div>
            </div>
            <button class="icon-btn" aria-label="关闭详情" @click="selected = null">
              <X :size="18" />
            </button>
          </div>

          <div class="drawer-tabs">
            <button
              class="drawer-tab"
              :class="{ active: drawerTab === 'overview' }"
              @click="drawerTab = 'overview'"
            >
              📊 概览与威胁
            </button>
            <button
              class="drawer-tab"
              :class="{ active: drawerTab === 'ai' }"
              @click="drawerTab = 'ai'"
            >
              🤖 AI 解锁深度 (10)
            </button>
            <button
              class="drawer-tab"
              :class="{ active: drawerTab === 'streaming' }"
              @click="drawerTab = 'streaming'"
            >
              🎬 流媒体原生 (10)
            </button>
            <button
              class="drawer-tab"
              :class="{ active: drawerTab === 'networks' }"
              @click="drawerTab = 'networks'"
            >
              🌐 双栈网络快照
            </button>
          </div>

          <div class="drawer-summary">
            <div>
              <span>综合风险分</span>
              <strong :class="riskClass(selected.risk)">{{ selected.risk }}</strong>
            </div>
            <div>
              <span>心跳状态</span>
              <StatusBadge :value="selected.status" />
            </div>
            <div>
              <span>上次全扫描</span>
              <span style="font-size: 12px; font-weight: 600; color: var(--text);">{{ relative(selected.last_scan) }}</span>
            </div>
          </div>

          <!-- Tab 1: Overview & Threat -->
          <div v-if="drawerTab === 'overview'" class="drawer-content">
            <div class="detail-section">
              <h3><Network :size="15" /> 网络身份与路由信息</h3>
              <div style="display: grid; gap: 8px; font-size: 11.5px;">
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <span style="color: var(--muted);">公网 IP 地址:</span>
                  <div style="display: flex; align-items: center; gap: 6px;">
                    <code style="color: #38bdf8;">{{ selected.ip_address || selected.masked_ip }}</code>
                    <button
                      v-if="selected.can_view_full_ip"
                      class="icon-btn"
                      style="width: 24px; height: 24px;"
                      :title="showFullIP ? '隐藏完整 IP' : '查看完整 IP'"
                      @click="toggleFullIP"
                    >
                      <EyeOff v-if="showFullIP" :size="12" />
                      <Eye v-else :size="12" />
                    </button>
                  </div>
                </div>
                <div style="display: flex; justify-content: space-between;">
                  <span style="color: var(--muted);">自治系统 ASN:</span>
                  <strong style="color: var(--text);">AS{{ selected.asn }} ({{ selected.organization }})</strong>
                </div>
                <div style="display: flex; justify-content: space-between;">
                  <span style="color: var(--muted);">协议族支持:</span>
                  <strong style="color: var(--text);">
                    {{ (selected.families?.length ? selected.families : [selected.family || 4]).map((f) => `IPv${f}`).join(' + ') }}
                  </strong>
                </div>
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <span style="color: var(--muted);">定期扫描周期:</span>
                  <select
                    :value="selected.scan_interval_minutes || 360"
                    style="padding: 2px 6px; background: var(--surface-2); border: 1px solid var(--border); border-radius: 4px; color: var(--text); font-size: 11px;"
                    @change="(e) => updateScanInterval(selected!, Number((e.target as HTMLSelectElement).value))"
                  >
                    <option :value="60">每小时 (1h)</option>
                    <option :value="360">每 6 小时 (6h)</option>
                    <option :value="720">每 12 小时 (12h)</option>
                    <option :value="1440">每天 (24h)</option>
                  </select>
                </div>
              </div>
            </div>

            <div class="detail-section">
              <h3><ShieldAlert :size="15" /> 威胁与身份变化检测</h3>
              <div style="display: grid; gap: 8px;">
                <div
                  v-if="selected.ip_changed"
                  style="padding: 8px 12px; background: rgba(239, 68, 68, 0.12); border-left: 3px solid #ef4444; border-radius: 4px; font-size: 11.5px;"
                >
                  <strong style="color: #ef4444; display: block;">公网 IP 已变更</strong>
                  <span style="color: var(--muted); font-size: 10.5px;">检测到新的网络身份，建议立即重新核验 AI 与流媒体解锁</span>
                </div>
                <div
                  v-if="selected.risk >= 60"
                  style="padding: 8px 12px; background: rgba(245, 158, 11, 0.12); border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 11.5px;"
                >
                  <strong style="color: #f59e0b; display: block;">高欺诈风险告警 (分值 {{ selected.risk }})</strong>
                  <span style="color: var(--muted); font-size: 10.5px;">IPQS / Scamalytics 欺诈数据库命中机房或代理特征</span>
                </div>
                <div
                  v-if="!selected.ip_changed && selected.risk < 60"
                  style="padding: 8px 12px; background: rgba(16, 185, 129, 0.12); border-left: 3px solid #10b981; border-radius: 4px; font-size: 11.5px;"
                >
                  <strong style="color: #10b981; display: block;">网络身份持续稳定</strong>
                  <span style="color: var(--muted); font-size: 10.5px;">未发现异常 IP 飘移或 DNSBL 严重阻断</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Tab 2: AI Unlocks Deep Dive -->
          <div v-else-if="drawerTab === 'ai'" class="drawer-content">
            <div class="unlock-grid-cards">
              <div
                v-for="(u, key) in (selected.unlocks?.ai ?? {})"
                :key="key"
                class="unlock-card-item"
              >
                <div class="unlock-card-head">
                  <strong>{{ u.name }}</strong>
                  <StatusBadge :value="u.status" :region="u.region" />
                </div>
                <div class="unlock-card-quality">{{ u.quality || '标准接入' }}</div>
                <div class="unlock-card-foot">
                  <span>{{ u.detail || '连通正常' }}</span>
                  <code v-if="u.latency_ms">{{ u.latency_ms }}ms</code>
                </div>
              </div>
            </div>
          </div>

          <!-- Tab 3: Streaming Media Deep Dive -->
          <div v-else-if="drawerTab === 'streaming'" class="drawer-content">
            <div class="unlock-grid-cards">
              <div
                v-for="(u, key) in (selected.unlocks?.streaming ?? {})"
                :key="key"
                class="unlock-card-item"
              >
                <div class="unlock-card-head">
                  <strong>{{ u.name }}</strong>
                  <StatusBadge :value="u.status" :region="u.region" />
                </div>
                <div class="unlock-card-quality">{{ u.quality || '标准接入' }}</div>
                <div class="unlock-card-foot">
                  <span>{{ u.detail || '连通正常' }}</span>
                  <code v-if="u.latency_ms">{{ u.latency_ms }}ms</code>
                </div>
              </div>
            </div>
          </div>

          <!-- Tab 4: Dual-stack Network Snapshots -->
          <div v-else class="drawer-content">
            <div v-if="selected.networks && selected.networks.length" style="display: grid; gap: 10px;">
              <div
                v-for="net in selected.networks"
                :key="net.family"
                style="padding: 12px; background: var(--surface-2); border: 1px solid var(--border); border-radius: 6px;"
              >
                <div style="display: flex; justify-content: space-between; margin-bottom: 6px;">
                  <strong style="color: #38bdf8;">IPv{{ net.family }} 网络快照</strong>
                  <span style="font-size: 10px; color: var(--muted);">{{ relative(net.collected_at) }}</span>
                </div>
                <div style="display: grid; gap: 4px; font-size: 11px; color: var(--muted);">
                  <div>公网 IP: <code style="color: var(--text);">{{ net.ip_address || net.masked_ip }}</code></div>
                  <div>ASN: <strong style="color: var(--text);">AS{{ net.asn }} · {{ net.organization }}</strong></div>
                  <div>Netflix: <StatusBadge :value="net.netflix" kind="media" /> | ChatGPT: <StatusBadge :value="net.chatgpt" kind="ai" /></div>
                </div>
              </div>
            </div>
            <div v-else style="text-align: center; color: var(--muted); padding: 24px 0;">
              暂无多栈历史快照
            </div>
          </div>

          <div class="drawer-actions">
            <button class="secondary-btn" @click="selected = null">关闭</button>
            <button class="secondary-btn" @click="triggerReinstall(selected)">
              <Wrench :size="15" /> 重装小鸡探针
            </button>
            <button
              class="secondary-btn"
              @click="openCompare([selected.id, data.nodes[0]?.id || selected.id])"
            >
              <Columns :size="15" /> 加入对比
            </button>
            <button class="primary-btn" @click="scan(selected)">
              <ScanLine :size="15" /> 立即全面扫描
            </button>
          </div>
        </aside>
      </div>
    </Transition>

    <!-- Task Timeline & Diagnostics Modal -->
    <Transition name="modal">
      <div v-if="logModalOpen && logNode" class="modal-backdrop" @click.self="logModalOpen = false">
        <section class="modal task-modal" role="dialog" aria-modal="true">
          <div class="modal-head">
            <div style="display: flex; align-items: center; gap: 10px;">
              <span class="country-code large">{{ logNode.country_code || '--' }}</span>
              <div>
                <h2>任务生命周期与运行日志 · {{ logNode.name }}</h2>
                <p>{{ logNode.provider || 'VPS' }} · {{ logNode.masked_ip }} · AS{{ logNode.asn }} ({{ logNode.organization }})</p>
              </div>
            </div>
            <button class="icon-btn" aria-label="关闭" @click="logModalOpen = false">
              <X :size="18" />
            </button>
          </div>

          <div class="modal-body" style="padding: 16px 20px; display: flex; flex-direction: column; gap: 14px;">
            <!-- Current Status Bar -->
            <div class="task-status-hero" :class="logNode.quality_status">
              <div class="status-left">
                <RefreshCw v-if="logNode.quality_status === 'scanning'" :size="22" class="spinning text-sky" />
                <AlertCircle v-else-if="logNode.quality_status === 'failed' || logNode.last_scan_error" :size="22" class="text-danger" />
                <Check v-else :size="22" class="text-emerald" />
                <div>
                  <strong>
                    {{ logNode.quality_status === 'scanning' ? '探测任务并发执行中...' : (logNode.quality_status === 'failed' || logNode.last_scan_error) ? '探测任务异常 / 待重试' : '探针待命就绪' }}
                  </strong>
                  <span>上次扫描: {{ relative(logNode.last_scan) }} · 最近心跳: {{ relative(logNode.last_seen) }}</span>
                </div>
              </div>
              <button class="primary-btn" style="height: 32px; font-size: 11.5px;" @click="scan(logNode)">
                <ScanLine :size="14" /> 重新下发扫描
              </button>
            </div>

            <!-- Error message box if failed -->
            <div v-if="logNode.last_scan_error" class="error-detail-box">
              <div class="error-head">
                <AlertCircle :size="14" />
                <span>探针异常详细日志与报错堆栈:</span>
              </div>
              <code>{{ logNode.last_scan_error }}</code>
            </div>

            <!-- Tasks Timeline -->
            <div class="timeline-container">
              <div class="timeline-title">
                <FileText :size="14" />
                <span>任务执行历史时间线</span>
              </div>

              <div v-if="taskLoading" class="loading-line" style="margin: 8px 0;"></div>

              <div v-if="nodeTasks.length" class="timeline-list">
                <div
                  v-for="task in nodeTasks"
                  :key="task.id"
                  class="timeline-item"
                  :class="task.status"
                >
                  <div class="timeline-dot"></div>
                  <div class="timeline-content">
                    <div class="timeline-header">
                      <span class="task-type-badge">{{ task.type === 'scan' ? '深度全量 20+ 服务扫描' : task.type }}</span>
                      <span class="task-status-pill" :class="task.status">
                        {{ task.status === 'pending' ? '排队中' : task.status === 'running' ? '探测中' : task.status === 'completed' ? '已完成' : '失败' }}
                      </span>
                      <span class="task-time">{{ relative(task.created_at) }} ({{ new Date(task.created_at).toLocaleTimeString() }})</span>
                    </div>
                    <p class="task-msg">{{ task.message }}</p>
                    <code v-if="task.error" class="task-err">{{ task.error }}</code>
                  </div>
                </div>
              </div>
              <div v-else class="empty-timeline">
                <Clock :size="24" />
                <span>暂无历史任务记录，点击上方按钮即可一键下发</span>
              </div>
            </div>

            <!-- Quick Action Footer -->
            <div class="task-modal-footer">
              <button class="secondary-btn" @click="triggerReinstall(logNode)">
                <Wrench :size="14" /> 重装小鸡探针
              </button>
              <button class="secondary-btn" @click="openTasksAndLogs(logNode)">
                <RefreshCw :size="14" :class="{ spinning: taskLoading }" /> 刷新任务日志
              </button>
            </div>
          </div>
        </section>
      </div>
    </Transition>

    <!-- Side-by-Side Compare Modal -->
    <NodeCompareModal
      v-if="compareModalOpen"
      :nodes="data.nodes"
      :compare-ids="compareNodeIds"
      @close="compareModalOpen = false"
      @remove-node="removeCompareNode"
    />

    <!-- Reinstall VPS Modal -->
    <Transition name="modal">
      <div v-if="reinstallModalOpen && reinstallData" class="modal-backdrop" @click.self="reinstallModalOpen = false">

        <section class="modal" role="dialog" aria-modal="true">
          <div class="modal-head">
            <div>
              <h2>重新安装小鸡探针 · {{ reinstallData.nodeName }}</h2>
              <p>小鸡重装系统后，直接运行下方命令即可重新连线并保留全部历史数据</p>
            </div>
            <button class="icon-btn" aria-label="关闭" @click="reinstallModalOpen = false">
              <X :size="18" />
            </button>
          </div>
          <div class="enrollment-result" style="padding: 16px 20px;">
            <div style="font-size: 11px; font-weight: 700; color: #38bdf8; margin-bottom: 8px;">一键极速重装命令 (root 权限运行):</div>
            <div style="display: flex; gap: 8px; align-items: center; background: #0c1117; padding: 12px; border-radius: 8px; border: 1px solid rgba(255,255,255,0.1);">
              <code style="flex: 1; color: #38bdf8; font-family: 'Fira Code', monospace; word-break: break-all; font-size: 12px;">{{ reinstallData.installCommand }}</code>
              <button class="primary-btn" style="flex-shrink: 0;" @click="copyText(reinstallData.installCommand, '重装命令已复制')">
                <Copy :size="14" /> 复制命令
              </button>
            </div>
          </div>
        </section>
      </div>
    </Transition>


    <!-- Add VPS Enrollment Modal -->
    <Transition name="modal">
      <div v-if="enrollOpen" class="modal-backdrop" @click.self="closeEnrollment">
        <section class="modal" role="dialog" aria-modal="true" aria-labelledby="enroll-modal-title">
          <div class="modal-head">
            <div>
              <h2 id="enroll-modal-title">添加小鸡监测探针</h2>
              <p>生成一次性 Ed25519 注册凭证，在目标服务器一键部署</p>
            </div>
            <button class="icon-btn" aria-label="关闭" @click="closeEnrollment">
              <X :size="18" />
            </button>
          </div>

          <form v-if="!enrollment" @submit.prevent="createEnrollment">
            <label>
              节点标识名称
              <input v-model="enrollForm.name" required placeholder="例如 HK-CMI-01 或 US-LAX-02" />
            </label>
            <div class="form-grid">
              <label>
                服务商 (Provider)
                <input v-model="enrollForm.provider" placeholder="例如 DMIT / Vultr / Hetzner" />
              </label>
              <label>
                地区标签 (Region)
                <input v-model="enrollForm.region" placeholder="例如 香港 / 洛杉矶 / 东京" />
              </label>
            </div>
            <div class="form-grid">
              <label>
                操作系统族
                <select v-model="enrollForm.os_family" style="height: 38px; padding: 0 10px; background: var(--surface-2); border: 1px solid var(--border); border-radius: 6px; color: var(--text);">
                  <option value="auto">自动识别 (Linux / BSD / Windows)</option>
                  <option value="linux">Linux</option>
                  <option value="freebsd">FreeBSD</option>
                  <option value="openbsd">OpenBSD</option>
                  <option value="darwin">macOS</option>
                  <option value="windows">Windows</option>
                </select>
              </label>
              <label>
                扫描周期
                <select v-model="enrollForm.scan_interval_minutes" style="height: 38px; padding: 0 10px; background: var(--surface-2); border: 1px solid var(--border); border-radius: 6px; color: var(--text);">
                  <option :value="60">每小时 (1h)</option>
                  <option :value="360">每 6 小时 (6h)</option>
                  <option :value="720">每 12 小时 (12h)</option>
                  <option :value="1440">每天 (24h)</option>
                </select>
              </label>
            </div>
            <div style="padding: 10px 12px; background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.25); border-radius: 6px; font-size: 11px; color: var(--muted); margin-bottom: 14px;">
              <span style="color: #38bdf8; font-weight: 600;">安全保障：</span>
              注册凭证 10 分钟内有效且单次使用。Agent 私钥在服务器本机生成并用 HTTP-Signature 签名。
            </div>
            <div style="display: flex; justify-content: flex-end; gap: 8px;">
              <button type="button" class="secondary-btn" @click="closeEnrollment">取消</button>
              <button type="submit" class="primary-btn">
                生成安装命令 <ChevronRight :size="15" />
              </button>
            </div>
          </form>

          <div v-else class="enrollment-result">
            <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 12px;">
              <div style="width: 28px; height: 28px; border-radius: 50%; background: var(--good); color: #fff; display: grid; place-items: center; font-weight: 700; font-size: 12px;">✓</div>
              <div>
                <strong style="color: var(--text); font-size: 13px;">一次性安装命令已生成</strong>
                <small style="display: block; color: var(--muted); font-size: 10px;">{{ relative(enrollment.expires_at) }} 内有效</small>
              </div>
            </div>
            <div class="command-box">
              <code>{{ enrollment.install_command }}</code>
              <button class="icon-btn" title="复制指令" aria-label="复制安装指令" @click="copyText(enrollment.install_command, '安装命令已复制')">
                <Copy :size="16" />
              </button>
            </div>
            <div style="display: flex; justify-content: flex-end; gap: 8px; margin-top: 14px;">
              <button class="secondary-btn" @click="enrollment = null">返回</button>
              <button class="primary-btn" @click="closeEnrollment">完成</button>
            </div>
          </div>
        </section>
      </div>
    </Transition>

    <!-- Admin Settings Modal -->
    <Transition name="modal">
      <div v-if="adminOpen" class="modal-backdrop" @click.self="adminOpen = false">
        <section class="modal" style="width: min(600px, 100%);" role="dialog" aria-modal="true">
          <div class="modal-head">
            <div>
              <h2>控制面用户与注册管理</h2>
              <p>管理系统账户、角色权限与密码重置</p>
            </div>
            <button class="icon-btn" @click="adminOpen = false"><X :size="18" /></button>
          </div>
          <div style="padding: 16px 20px;">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; padding: 12px; background: var(--surface-2); border-radius: 6px;">
              <div>
                <strong style="color: var(--text); display: block; font-size: 12.5px;">开放新用户注册</strong>
                <small style="color: var(--muted);">关闭后仅管理员可创建新用户</small>
              </div>
              <input type="checkbox" v-model="registrationEnabled" @change="saveRegistration" style="width: 18px; height: 18px; cursor: pointer;" />
            </div>

            <h3 style="font-size: 12.5px; margin: 0 0 10px 0; color: var(--text);">已注册用户列表</h3>
            <div style="display: grid; gap: 8px; max-height: 240px; overflow-y: auto;">
              <div
                v-for="u in users"
                :key="u.id"
                style="display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: var(--surface-2); border: 1px solid var(--border); border-radius: 6px; font-size: 11.5px;"
              >
                <div>
                  <strong style="color: var(--text);">{{ u.display_name || u.username }}</strong>
                  <span style="color: var(--muted); margin-left: 6px;">(@{{ u.username }})</span>
                </div>
                <div style="display: flex; gap: 6px; align-items: center;">
                  <select
                    :value="u.role"
                    style="padding: 2px 6px; background: var(--surface); border: 1px solid var(--border); border-radius: 4px; color: var(--text); font-size: 11px;"
                    @change="(e) => updateRole(u, (e.target as HTMLSelectElement).value)"
                  >
                    <option value="admin">管理员</option>
                    <option value="user">普通用户</option>
                  </select>
                  <button class="secondary-btn" style="height: 26px; padding: 0 8px; font-size: 10.5px;" @click="createReset(u)">
                    重置密码
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </Transition>

    <!-- Password Change Modal -->
    <Transition name="modal">
      <div v-if="passwordOpen" class="modal-backdrop" @click.self="passwordOpen = false">
        <section class="modal" role="dialog" aria-modal="true">
          <div class="modal-head">
            <div>
              <h2>修改当前账户密码</h2>
              <p>修改后将自动退出其他登录会话</p>
            </div>
            <button class="icon-btn" @click="passwordOpen = false"><X :size="18" /></button>
          </div>
          <form @submit.prevent="changePassword" style="padding: 16px 20px;">
            <label>
              当前原密码
              <input v-model="passwordForm.current_password" type="password" required />
            </label>
            <label>
              新密码 (至少 8 位)
              <input v-model="passwordForm.new_password" type="password" minlength="8" required />
            </label>
            <div style="display: flex; justify-content: flex-end; gap: 8px; margin-top: 14px;">
              <button type="button" class="secondary-btn" @click="passwordOpen = false">取消</button>
              <button type="submit" class="primary-btn">确认修改</button>
            </div>
          </form>
        </section>
      </div>
    </Transition>

    <!-- Global Toast Alert -->
    <Transition name="toast">
      <div v-if="toast" class="toast" role="status">
        <Zap :size="16" style="color: #38bdf8;" />
        <span>{{ toast }}</span>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fatal-screen {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 20px;
  background: var(--bg);
}
.fatal-card {
  width: min(440px, 100%);
  padding: 28px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.fatal-card h2 {
  margin: 0;
  font-size: 18px;
}
.fatal-card p {
  margin: 0;
  color: var(--muted);
  font-size: 13px;
}
</style>
