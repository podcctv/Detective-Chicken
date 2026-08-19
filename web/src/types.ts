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
  masked_ip: string
  ip_address?: string
  can_view_full_ip: boolean
  asn: number
  organization: string
  country_code: string
  risk: number
  status: 'online' | 'warning' | 'alert' | 'offline' | 'pending'
  netflix: string
  chatgpt: string
  dnsbl: number
  ip_changed: boolean
  last_seen: string
  last_scan: string
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

export interface NodeDetail extends Node {
  series: TrendPoint[]
  alerts: Alert[]
  latest_quality?: Quality
  latest_collector?: { name: string; adapter_version: string; upstream_version?: string }
  report_time?: string
}

export interface Dashboard {
  generated_at: string
  stats: Record<string, number>
  trend: TrendPoint[]
  nodes: Node[]
  rankings: NodeRanking[]
  alerts: Alert[]
  regions: Record<string, number>
}

export interface Enrollment {
  token: string
  expires_at: string
  install_url: string
  install_command: string
}
