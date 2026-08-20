import type { Dashboard, Node, NodeUnlocks, ServiceStat } from './types'

const now = Date.now()
const ago = (minutes: number) => new Date(now - minutes * 60_000).toISOString()

const makeUnlocks = (overrides?: {
  streaming?: Record<string, Partial<NodeUnlocks['streaming'][string]>>
  ai?: Record<string, Partial<NodeUnlocks['ai'][string]>>
}): NodeUnlocks => {
  const baseStreaming: NodeUnlocks['streaming'] = {
    netflix: { id: 'netflix', name: 'Netflix', category: 'streaming', status: 'available', region: 'Global', quality: '原生 4K 解锁', latency_ms: 36, detail: '支持全量版权库与 HDR 杜比视界' },
    disney: { id: 'disney', name: 'Disney+', category: 'streaming', status: 'available', region: 'Global', quality: '原生解锁', latency_ms: 40, detail: 'Star / IMAX Enhanced 完整支持' },
    youtube: { id: 'youtube', name: 'YouTube Premium', category: 'streaming', status: 'available', region: 'Global', quality: '原生支持', latency_ms: 25, detail: '后台播放与无损音频正常' },
    spotify: { id: 'spotify', name: 'Spotify', category: 'streaming', status: 'available', region: 'Global', quality: '全功能', latency_ms: 28, detail: '曲库与动态歌词正常' },
    prime: { id: 'prime', name: 'Prime Video', category: 'streaming', status: 'available', region: 'Global', quality: '原生解锁', latency_ms: 42, detail: 'Amazon 区域解锁正常' },
    hbo: { id: 'hbo', name: 'Max (HBO)', category: 'streaming', status: 'available', region: 'Global', quality: '原生解锁', latency_ms: 48, detail: 'Max 完整电影库' },
    hulu: { id: 'hulu', name: 'Hulu', category: 'streaming', status: 'available', region: 'US', quality: '原生支持', latency_ms: 65, detail: 'Live TV / 点播支持' },
    bilibili: { id: 'bilibili', name: 'Bilibili (港澳台)', category: 'streaming', status: 'available', region: 'HK/MO/TW', quality: '解除限制', latency_ms: 20, detail: '支持仅限港澳台番剧' },
    tiktok: { id: 'tiktok', name: 'TikTok', category: 'streaming', status: 'available', region: 'Global', quality: '原生解锁', latency_ms: 30, detail: '短视频浏览与发帖无异常' },
    appletv: { id: 'appletv', name: 'Apple TV+', category: 'streaming', status: 'available', region: 'Global', quality: '原生支持', latency_ms: 28, detail: 'Apple 官方 CDN 直连' },
  }

  const baseAI: NodeUnlocks['ai'] = {
    chatgpt: { id: 'chatgpt', name: 'ChatGPT / OpenAI', category: 'ai', status: 'available', region: 'Global', quality: 'Web + API', latency_ms: 38, detail: '无 Cloudflare 验证，全模型畅通' },
    claude: { id: 'claude', name: 'Claude (Anthropic)', category: 'ai', status: 'available', region: 'Global', quality: 'Web + API', latency_ms: 42, detail: 'Claude 3.5 Sonnet / 控制台正常' },
    gemini: { id: 'gemini', name: 'Google Gemini', category: 'ai', status: 'available', region: 'Global', quality: 'Web + API', latency_ms: 26, detail: 'Gemini Advanced & AI Studio 正常' },
    midjourney: { id: 'midjourney', name: 'Midjourney', category: 'ai', status: 'available', region: 'Global', quality: 'Web + Discord', latency_ms: 45, detail: 'Discord 代理通信极速' },
    copilot: { id: 'copilot', name: 'Microsoft Copilot', category: 'ai', status: 'available', region: 'Global', quality: '原生支持', latency_ms: 28, detail: 'Bing AI / Edge Copilot 极速响应' },
    grok: { id: 'grok', name: 'xAI Grok', category: 'ai', status: 'available', region: 'Global', quality: '原生支持', latency_ms: 55, detail: 'X 平台 Grok 畅通' },
    perplexity: { id: 'perplexity', name: 'Perplexity AI', category: 'ai', status: 'available', region: 'Global', quality: '原生支持', latency_ms: 35, detail: 'Pro 搜索与 API 调用免验证' },
    github_cop: { id: 'github_cop', name: 'GitHub Copilot', category: 'ai', status: 'available', region: 'Global', quality: 'IDE 直连', latency_ms: 24, detail: '代码补全实时响应' },
    deepseek: { id: 'deepseek', name: 'DeepSeek', category: 'ai', status: 'available', region: 'Global', quality: 'Web + API', latency_ms: 22, detail: '官方推理集群直连' },
    huggingface: { id: 'huggingface', name: 'HuggingFace', category: 'ai', status: 'available', region: 'Global', quality: 'Spaces / Hub', latency_ms: 36, detail: '模型仓库与推理节点通畅' },
  }

  if (overrides?.streaming) {
    for (const [k, v] of Object.entries(overrides.streaming)) {
      if (baseStreaming[k]) baseStreaming[k] = { ...baseStreaming[k], ...v }
    }
  }
  if (overrides?.ai) {
    for (const [k, v] of Object.entries(overrides.ai)) {
      if (baseAI[k]) baseAI[k] = { ...baseAI[k], ...v }
    }
  }
  return { streaming: baseStreaming, ai: baseAI }
}

const node = (values: Partial<Node> & Pick<Node, 'id' | 'name'>): Node => {
  const defaultUnlocks = makeUnlocks()
  return {
    provider: 'Vultr', region: '东京', family: 4, families: [4], masked_ip: '203.0.113.*',
    masked_ipv4: '203.0.113.*', asn: 20473,
    organization: 'The Constant Company', country_code: 'JP', usage_type: '机房', ip_type: '原生',
    latitude: 35.6762, longitude: 139.6503,
    risk: 11, status: 'online', netflix: 'available', chatgpt: 'available',
    unlocks: defaultUnlocks, dnsbl: 0, ip_changed: false, can_view_full_ip: false,
    scan_interval_minutes: 360, quality_status: 'ready',
    last_seen: ago(1), last_scan: ago(240), ...values,
  }
}


export const demoNodes: Node[] = [
  node({
    id: 'node_hkg_01', name: 'HK-CMI-01', provider: 'DMIT', region: '香港', masked_ip: '103.145.12.*',
    masked_ipv4: '103.145.12.*', masked_ipv6: '2402:4e00:1000::*', families: [4, 6],
    asn: 906, organization: 'DMIT Cloud Services', country_code: 'HK', usage_type: '机房', ip_type: '原生',
    latitude: 22.3193, longitude: 114.1694,
    risk: 18, status: 'online', netflix: 'available', chatgpt: 'available', dnsbl: 0,
    last_seen: ago(1), last_scan: ago(120),
    unlocks: makeUnlocks({
      streaming: {
        netflix: { region: 'HK', quality: '原生 4K 解锁 (港区)', latency_ms: 22, detail: '支持港区全量影视与繁体中文字幕' },
        disney: { region: 'HK', quality: '原生解锁', latency_ms: 28, detail: 'Star+ 内容完整' },
        youtube: { region: 'HK', quality: '原生无广告', latency_ms: 18 },
        spotify: { region: 'HK', quality: '全功能', latency_ms: 20 },
        bilibili: { region: 'HK/MO/TW', quality: '解除限制', latency_ms: 15, detail: '港澳台独播动漫' },
        tiktok: { region: 'Global', quality: '原生解锁', latency_ms: 24 },
        hulu: { status: 'limited', region: 'US', quality: '需配合北美 DNS', latency_ms: 130 },
      },
      ai: {
        chatgpt: { region: 'US/Global', quality: 'Web + API (极速)', latency_ms: 28, detail: '无 Turnstile 质询' },
        claude: { region: 'Global', quality: 'Web + API', latency_ms: 32 },
        gemini: { region: 'Global', quality: 'Web + API', latency_ms: 22 },
        deepseek: { region: 'Global', quality: '官方直连', latency_ms: 14 },
      },
    }),
  }),
  node({
    id: 'node_lax_02', name: 'US-LAX-02', provider: 'RackNerd', region: '洛杉矶', masked_ip: '198.51.100.*',
    asn: 62240, organization: 'Clouvider Limited', country_code: 'US', latitude: 34.0522, longitude: -118.2437,
    risk: 72, status: 'alert', netflix: 'limited', chatgpt: 'available', dnsbl: 8, ip_changed: true,
    last_seen: ago(2), last_scan: ago(28),
    unlocks: makeUnlocks({
      streaming: {
        netflix: { status: 'limited', region: 'US', quality: '仅自制剧 (Originals Only)', latency_ms: 85, detail: '机房 ASN 命中 Netflix 黑名单，非自制剧受限' },
        disney: { region: 'US', quality: '原生美区', latency_ms: 70 },
        hbo: { region: 'US', quality: '原生 4K Max', latency_ms: 62 },
        hulu: { region: 'US', quality: '原生支持', latency_ms: 60 },
        bilibili: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0, detail: '海外非港澳台地区' },
        tiktok: { status: 'limited', region: 'US', quality: '发帖审查', latency_ms: 95, detail: 'IP 权重偏低，发帖需审核' },
        prime: { status: 'limited', region: 'US', quality: '二次验证', latency_ms: 90 },
      },
      ai: {
        chatgpt: { status: 'limited', region: 'US', quality: 'API 正常 / Web 弹窗验证', latency_ms: 110, detail: '高风险 IP 触发 Cloudflare 验证码' },
        claude: { status: 'blocked', region: 'US', quality: '403 封禁 (Blocked)', latency_ms: 0, detail: 'Anthropic 机房策略阻断' },
        gemini: { region: 'US', quality: '原生支持', latency_ms: 48 },
        perplexity: { status: 'limited', region: 'US', quality: '频繁人机验证', latency_ms: 88 },
      },
    }),
  }),
  node({
    id: 'node_nrt_03', name: 'JP-NRT-03', provider: 'Vultr', region: '东京', masked_ip: '45.76.201.*',
    asn: 20473, organization: 'The Constant Company', country_code: 'JP', latitude: 35.6762, longitude: 139.6503,
    risk: 11, status: 'online', netflix: 'available', chatgpt: 'available', dnsbl: 0,
    last_seen: ago(1), last_scan: ago(240),
    unlocks: makeUnlocks({
      streaming: {
        netflix: { region: 'JP', quality: '原生 4K 日区', latency_ms: 24, detail: '日漫新番、特摄及日剧独占库' },
        disney: { region: 'JP', quality: '原生解锁', latency_ms: 28 },
        hulu: { region: 'JP', quality: 'Hulu Japan 原生', latency_ms: 26 },
        hbo: { status: 'limited', region: 'JP', quality: 'U-NEXT 合作', latency_ms: 55 },
        bilibili: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
        tiktok: { region: 'JP', quality: '日区流量推荐', latency_ms: 20 },
      },
      ai: {
        chatgpt: { region: 'JP', quality: 'Web + API (极速)', latency_ms: 25, detail: '东京机房超低延迟' },
        claude: { region: 'JP', quality: 'Web + API', latency_ms: 30 },
        gemini: { region: 'JP', quality: 'Web + API', latency_ms: 20 },
        deepseek: { region: 'Global', quality: '直连稳定', latency_ms: 24 },
      },
    }),
  }),
  node({
    id: 'node_fra_04', name: 'DE-FRA-04', provider: 'Hetzner', region: '法兰克福', family: 6, masked_ip: '2a01:4f8:c2c:*',
    asn: 24940, organization: 'Hetzner Online', country_code: 'DE', latitude: 50.1109, longitude: 8.6821,
    risk: 34, status: 'online', netflix: 'blocked', chatgpt: 'available', dnsbl: 1,
    last_seen: ago(1), last_scan: ago(240),
    unlocks: makeUnlocks({
      streaming: {
        netflix: { status: 'blocked', region: 'DE', quality: 'IPv6 拦截', latency_ms: 0, detail: 'Hetzner IPv6 段被 Netflix 识别为机房代理' },
        disney: { status: 'blocked', region: 'DE', quality: '区域阻断', latency_ms: 0, detail: '错误码 73' },
        hbo: { status: 'blocked', region: 'DE', quality: '德国未开通', latency_ms: 0 },
        hulu: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
        bilibili: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
        youtube: { region: 'DE', quality: '原生欧区', latency_ms: 28 },
        spotify: { region: 'DE', quality: '全功能', latency_ms: 30 },
      },
      ai: {
        chatgpt: { region: 'DE/EU', quality: 'IPv6 端到端直连', latency_ms: 38 },
        claude: { region: 'EU', quality: 'Web + API', latency_ms: 42 },
        gemini: { region: 'DE', quality: 'Web + API', latency_ms: 28 },
        huggingface: { region: 'Global', quality: '欧洲骨干直连', latency_ms: 22 },
      },
    }),
  }),
  node({
    id: 'node_sin_05', name: 'SG-SIN-05', provider: 'LightNode', region: '新加坡', masked_ip: '172.104.51.*',
    asn: 63949, organization: 'Akamai Connected Cloud', country_code: 'SG', latitude: 1.3521, longitude: 103.8198,
    risk: 48, status: 'warning', netflix: 'available', chatgpt: 'blocked', dnsbl: 3,
    last_seen: ago(4), last_scan: ago(660),
    unlocks: makeUnlocks({
      streaming: {
        netflix: { region: 'SG', quality: '原生 4K 新区', latency_ms: 30, detail: '含东南亚华语与独家剧集' },
        disney: { region: 'SG', quality: '原生解锁', latency_ms: 35 },
        bilibili: { region: 'SEA', quality: '东南亚解除限制', latency_ms: 28, detail: '支持泰语/印尼语/英文字幕' },
        tiktok: { region: 'SG', quality: '东南亚电商流量', latency_ms: 22 },
        hulu: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
      },
      ai: {
        chatgpt: { status: 'blocked', region: 'SG', quality: '403 Forbidden', latency_ms: 0, detail: 'Akamai IP 段被 OpenAI 判定为代理封锁' },
        claude: { status: 'blocked', region: 'SG', quality: '403 Access Denied', latency_ms: 0, detail: 'Anthropic 机房黑名单' },
        gemini: { region: 'SG', quality: '原生支持', latency_ms: 20 },
        deepseek: { region: 'Global', quality: '极速 18ms', latency_ms: 18 },
        perplexity: { status: 'limited', region: 'SG', quality: '偶发人机验证', latency_ms: 65 },
      },
    }),
  }),
  node({
    id: 'node_ams_06', name: 'NL-AMS-06', provider: 'GreenCloud', region: '阿姆斯特丹', masked_ip: '185.22.153.*',
    asn: 49544, organization: 'iFog GmbH', country_code: 'NL', latitude: 52.3676, longitude: 4.9041,
    risk: 22, status: 'offline', netflix: 'available', chatgpt: 'available', dnsbl: 0,
    last_seen: ago(19), last_scan: ago(780),
    unlocks: makeUnlocks({
      streaming: {
        netflix: { region: 'NL', quality: '原生欧区', latency_ms: 32 },
        disney: { region: 'NL', quality: '原生解锁', latency_ms: 34 },
        hbo: { region: 'NL', quality: '原生 Max 荷兰', latency_ms: 40 },
        hulu: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
        bilibili: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
      },
      ai: {
        chatgpt: { region: 'NL/EU', quality: 'Web + API', latency_ms: 38 },
        claude: { region: 'EU', quality: 'Web + API', latency_ms: 42 },
        gemini: { region: 'NL', quality: 'Web + API', latency_ms: 28 },
      },
    }),
  }),
  node({
    id: 'node_lhr_07', name: 'UK-LON-07', provider: 'Linode', region: '伦敦', masked_ip: '176.58.102.*',
    asn: 63949, organization: 'Akamai International', country_code: 'GB', latitude: 51.5074, longitude: -0.1278,
    risk: 16, status: 'online', netflix: 'available', chatgpt: 'available', dnsbl: 0,
    last_seen: ago(1), last_scan: ago(180),
    unlocks: makeUnlocks({
      streaming: {
        netflix: { region: 'UK', quality: '原生 4K 英区', latency_ms: 28 },
        disney: { region: 'UK', quality: '原生解锁', latency_ms: 32 },
        hbo: { status: 'limited', region: 'UK', quality: 'Sky Atlantic 独占', latency_ms: 50 },
        hulu: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
        bilibili: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
      },
      ai: {
        chatgpt: { region: 'UK', quality: 'Web + API', latency_ms: 32 },
        claude: { region: 'UK', quality: 'Web + API', latency_ms: 35 },
        gemini: { region: 'UK', quality: 'Web + API', latency_ms: 25 },
      },
    }),
  }),
  node({
    id: 'node_syd_08', name: 'AU-SYD-08', provider: 'OVHcloud', region: '悉尼', masked_ip: '139.99.144.*',
    asn: 16276, organization: 'OVH SAS', country_code: 'AU', latitude: -33.8688, longitude: 151.2093,
    risk: 19, status: 'online', netflix: 'available', chatgpt: 'available', dnsbl: 0,
    last_seen: ago(2), last_scan: ago(310),
    unlocks: makeUnlocks({
      streaming: {
        netflix: { region: 'AU', quality: '原生 4K 澳区', latency_ms: 35 },
        disney: { region: 'AU', quality: '原生解锁', latency_ms: 38 },
        hbo: { status: 'limited', region: 'AU', quality: 'Binge 联运', latency_ms: 55 },
        hulu: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
        bilibili: { status: 'blocked', region: 'N/A', quality: '不可用', latency_ms: 0 },
      },
      ai: {
        chatgpt: { region: 'AU', quality: 'Web + API', latency_ms: 40 },
        claude: { region: 'AU', quality: 'Web + API', latency_ms: 42 },
        gemini: { region: 'AU', quality: 'Web + API', latency_ms: 30 },
      },
    }),
  }),
]

export const demoServices: ServiceStat[] = [
  { id: 'chatgpt', name: 'ChatGPT / OpenAI', category: 'ai', total: 8, available: 6, limited: 1, blocked: 1, rate: 75 },
  { id: 'claude', name: 'Claude (Anthropic)', category: 'ai', total: 8, available: 6, limited: 0, blocked: 2, rate: 75 },
  { id: 'gemini', name: 'Google Gemini', category: 'ai', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
  { id: 'midjourney', name: 'Midjourney', category: 'ai', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
  { id: 'copilot', name: 'Microsoft Copilot', category: 'ai', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
  { id: 'grok', name: 'xAI Grok', category: 'ai', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
  { id: 'perplexity', name: 'Perplexity AI', category: 'ai', total: 8, available: 6, limited: 2, blocked: 0, rate: 75 },
  { id: 'github_cop', name: 'GitHub Copilot', category: 'ai', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
  { id: 'deepseek', name: 'DeepSeek', category: 'ai', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
  { id: 'huggingface', name: 'HuggingFace', category: 'ai', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },

  { id: 'netflix', name: 'Netflix', category: 'streaming', total: 8, available: 6, limited: 1, blocked: 1, rate: 75 },
  { id: 'disney', name: 'Disney+', category: 'streaming', total: 8, available: 7, limited: 0, blocked: 1, rate: 88 },
  { id: 'youtube', name: 'YouTube Premium', category: 'streaming', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
  { id: 'spotify', name: 'Spotify', category: 'streaming', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
  { id: 'prime', name: 'Prime Video', category: 'streaming', total: 8, available: 7, limited: 1, blocked: 0, rate: 88 },
  { id: 'hbo', name: 'Max (HBO)', category: 'streaming', total: 8, available: 4, limited: 3, blocked: 1, rate: 50 },
  { id: 'hulu', name: 'Hulu', category: 'streaming', total: 8, available: 2, limited: 1, blocked: 5, rate: 25 },
  { id: 'bilibili', name: 'Bilibili (港澳台)', category: 'streaming', total: 8, available: 2, limited: 0, blocked: 6, rate: 25 },
  { id: 'tiktok', name: 'TikTok', category: 'streaming', total: 8, available: 7, limited: 1, blocked: 0, rate: 88 },
  { id: 'appletv', name: 'Apple TV+', category: 'streaming', total: 8, available: 8, limited: 0, blocked: 0, rate: 100 },
]

export const demoDashboard: Dashboard = {
  generated_at: new Date(now).toISOString(),
  stats: {
    total: 8,
    online: 7,
    abnormal: 2,
    high_risk: 1,
    ip_changes: 1,
    media_degraded: 3,
    dnsbl_added: 2,
    ai_unlock_rate: 88,
    streaming_unlock_rate: 71,
  },
  trend: Array.from({ length: 20 }, (_, i) => ({
    at: new Date(now - (19 - i) * 12 * 3600_000).toISOString(),
    risk: [34, 36, 35, 38, 37, 41, 39, 43, 42, 40, 44, 46, 45, 49, 48, 54, 57, 63, 68, 72][i],
    ipqs: [29, 31, 30, 33, 32, 36, 34, 38, 37, 35, 39, 41, 40, 44, 43, 49, 52, 58, 63, 67][i],
    scamalytics: [18, 20, 19, 22, 21, 25, 23, 27, 26, 24, 28, 30, 29, 33, 32, 38, 41, 47, 52, 58][i],
    dnsbl: i > 16 ? 8 : i > 13 ? 3 : 0,
  })),
  nodes: demoNodes,
  rankings: [
    { rank: 1, node_id: 'node_nrt_03', name: 'JP-NRT-03', provider: 'Vultr', region: '东京', quality: 89, unlocks: 2, risk: 11, status: 'online' },
    { rank: 2, node_id: 'node_hkg_01', name: 'HK-CMI-01', provider: 'DMIT', region: '香港', quality: 82, unlocks: 2, risk: 18, status: 'online' },
    { rank: 3, node_id: 'node_lon_07', name: 'UK-LON-07', provider: 'OVHcloud', region: '伦敦', quality: 81, unlocks: 2, risk: 19, status: 'online' },
    { rank: 4, node_id: 'node_fra_04', name: 'DE-FRA-04', provider: 'Hetzner', region: '法兰克福', quality: 66, unlocks: 1, risk: 34, status: 'online' },
  ],
  alerts: [

    { id: 'alert_01', node_id: 'node_lax_02', node_name: 'US-LAX-02', type: 'risk_spike', severity: 'critical', title: '综合风险分显著上升', detail: '24 小时内从 41 上升到 72，超过变化阈值 20。', created_at: ago(26), acknowledged: false },
    { id: 'alert_02', node_id: 'node_sin_05', node_name: 'SG-SIN-05', type: 'media_degraded', severity: 'warning', title: 'ChatGPT / Claude 遭到封禁', detail: '机房 ASN 被 OpenAI/Anthropic 标记为代理拦截。', created_at: ago(120), acknowledged: false },
    { id: 'alert_03', node_id: 'node_fra_04', node_name: 'DE-FRA-04', type: 'media_degraded', severity: 'warning', title: 'Netflix IPv6 过滤拦截', detail: 'IPv6 机房段被判定为机房代理，Disney+ 亦受阻断。', created_at: ago(240), acknowledged: false },
    { id: 'alert_04', node_id: 'node_ams_06', node_name: 'NL-AMS-06', type: 'heartbeat_missing', severity: 'warning', title: '节点心跳超时', detail: '已连续 19 分钟未收到心跳。', created_at: ago(9), acknowledged: false },
  ],
  regions: { HK: 1, US: 1, JP: 1, DE: 1, SG: 1, NL: 1, GB: 1, AU: 1 },
  services: demoServices,
}
