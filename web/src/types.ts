export type ServiceCategory = 'streaming' | 'ai' | 'social'
export type ServiceStatus = 'available' | 'limited' | 'blocked' | 'untested' | 'unknown'


export interface UnlockInfo {
  id: string
  name: string
  category: ServiceCategory
  status: ServiceStatus
  region?: string
  quality?: string
  latency_ms?: number
  detail?: string
  checked_at?: string
}

export interface NodeUnlocks {
  streaming: Record<string, UnlockInfo>
  ai: Record<string, UnlockInfo>
}

export interface ServiceStat {
  id: string
  name: string
  category: ServiceCategory
  total: number
  available: number
  limited: number
  blocked: number
  rate: number
}

export interface User {
  id: string
  username: string
  display_name: string
  role: 'admin' | 'user'
  created_at: string
  updated_at: string
}

export interface AuthStatus {
  authenticated: boolean
  user?: User
  settings: { registration_enabled: boolean; bootstrapped: boolean }
}

export interface Node {
  id: string
  name: string
  provider: string
  region: string
  family: number
  families?: number[]
  masked_ip: string
  ip_address?: string
  can_view_full_ip?: boolean
  asn: number
  organization: string
  country_code: string
  latitude?: number
  longitude?: number
  risk: number
  status: 'online' | 'warning' | 'alert' | 'offline' | 'pending'
  netflix: string
  chatgpt: string
  unlocks?: NodeUnlocks
  dnsbl: number
  ip_changed: boolean
  last_seen: string
  last_scan: string
  scan_interval_minutes?: number
  quality_status?: 'pending' | 'scanning' | 'ready' | 'partial' | 'failed' | ''
  last_scan_error?: string
  last_task?: TaskLog
}

export interface TaskLog {
  id: string
  node_id: string
  type: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  message: string
  error?: string
  created_at: string
  updated_at: string
}


export interface TrendPoint {
  at: string
  risk: number
  ipqs: number
  scamalytics: number
  dnsbl: number
}

export interface Alert {
  id: string
  node_id: string
  node_name: string
  type: string
  severity: 'critical' | 'warning' | 'info'
  title: string
  detail: string
  created_at: string
  acknowledged: boolean
}

export interface NodeRanking {
  rank: number
  node_id: string
  name: string
  provider: string
  region: string
  quality: number
  unlocks: number
  risk: number
  status: string
}

export interface Quality {
  asn: number
  organization: string
  country_code: string
  usage_type: string
  company_type: string
  scores: Record<string, unknown>
  factors: Record<string, unknown>
  media: Record<string, unknown>
  mail: Record<string, unknown>
}

export interface NetworkSnapshot {
  family: number
  masked_ip: string
  ip_address?: string
  asn: number
  organization: string
  country_code: string
  risk: number
  netflix: string
  chatgpt: string
  unlocks?: NodeUnlocks
  collected_at: string
  quality: Quality
}

export interface NodeDetail extends Node {
  series: TrendPoint[]
  alerts: Alert[]
  tasks?: TaskLog[]
  latest_quality?: Quality
  latest_collector?: {
    name: string
    adapter_version: string
    upstream_version?: string
  }
  report_time?: string
  networks: NetworkSnapshot[]
}


export interface Dashboard {
  generated_at: string
  stats: Record<string, number>
  trend: TrendPoint[]
  nodes: Node[]
  rankings?: NodeRanking[]
  alerts: Alert[]
  regions: Record<string, number>
  services?: ServiceStat[]
}

export interface Enrollment {
  token: string
  expires_at: string
  install_url: string
  install_command: string
}
