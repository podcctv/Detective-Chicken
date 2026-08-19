export interface Node {
  id: string
  name: string
  provider: string
  region: string
  family: number
  masked_ip: string
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

export interface Dashboard {
  generated_at: string
  stats: Record<string, number>
  trend: TrendPoint[]
  nodes: Node[]
  alerts: Alert[]
  regions: Record<string, number>
}
