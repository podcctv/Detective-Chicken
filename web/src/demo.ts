import type { Dashboard, Node } from './types'

const now = Date.now()
const ago = (minutes: number) => new Date(now - minutes * 60_000).toISOString()
const node = (values: Partial<Node> & Pick<Node, 'id' | 'name'>): Node => ({
  provider: 'Vultr', region: '东京', family: 4, masked_ip: '203.0.113.*', asn: 20473,
  organization: 'The Constant Company', country_code: 'JP', risk: 11, status: 'online',
  netflix: 'available', chatgpt: 'available', dnsbl: 0, ip_changed: false,
  last_seen: ago(1), last_scan: ago(240), ...values,
})

export const demoDashboard: Dashboard = {
  generated_at: new Date(now).toISOString(),
  stats: { total: 6, online: 5, abnormal: 2, high_risk: 1, ip_changes: 1, media_degraded: 3, dnsbl_added: 3 },
  trend: Array.from({ length: 20 }, (_, i) => ({
    at: new Date(now - (19 - i) * 12 * 3600_000).toISOString(),
    risk: [34, 36, 35, 38, 37, 41, 39, 43, 42, 40, 44, 46, 45, 49, 48, 54, 57, 63, 68, 72][i],
    ipqs: [29, 31, 30, 33, 32, 36, 34, 38, 37, 35, 39, 41, 40, 44, 43, 49, 52, 58, 63, 67][i],
    scamalytics: [18, 20, 19, 22, 21, 25, 23, 27, 26, 24, 28, 30, 29, 33, 32, 38, 41, 47, 52, 58][i],
    dnsbl: i > 16 ? 8 : i > 13 ? 3 : 0,
  })),
  nodes: [
    node({ id: 'node_lax_02', name: 'US-LAX-02', provider: 'RackNerd', region: '洛杉矶', masked_ip: '198.51.100.*', asn: 62240, organization: 'Clouvider Limited', country_code: 'US', risk: 72, status: 'alert', netflix: 'limited', dnsbl: 8, ip_changed: true, last_seen: ago(2), last_scan: ago(28) }),
    node({ id: 'node_sin_05', name: 'SG-SIN-05', provider: 'LightNode', region: '新加坡', masked_ip: '172.104.51.*', asn: 63949, organization: 'Akamai Connected Cloud', country_code: 'SG', risk: 48, status: 'warning', chatgpt: 'blocked', dnsbl: 3, last_seen: ago(4), last_scan: ago(660) }),
    node({ id: 'node_fra_04', name: 'DE-FRA-04', provider: 'Hetzner', region: '法兰克福', family: 6, masked_ip: '2a01:4f8:c2c:*', asn: 24940, organization: 'Hetzner Online', country_code: 'DE', risk: 34, netflix: 'blocked', dnsbl: 1, last_scan: ago(240) }),
    node({ id: 'node_ams_06', name: 'NL-AMS-06', provider: 'GreenCloud', region: '阿姆斯特丹', masked_ip: '185.22.153.*', asn: 49544, organization: 'iFog GmbH', country_code: 'NL', risk: 22, status: 'offline', last_seen: ago(19), last_scan: ago(780) }),
    node({ id: 'node_hkg_01', name: 'HK-CMI-01', provider: 'DMIT', region: '香港', masked_ip: '103.145.12.*', asn: 906, organization: 'DMIT Cloud Services', country_code: 'HK', risk: 18, last_seen: ago(1), last_scan: ago(120) }),
    node({ id: 'node_nrt_03', name: 'JP-NRT-03' }),
  ],
  alerts: [
    { id: 'alert_01', node_id: 'node_lax_02', node_name: 'US-LAX-02', type: 'risk_spike', severity: 'critical', title: '综合风险分显著上升', detail: '24 小时内从 41 上升到 72，超过变化阈值 20。', created_at: ago(26), acknowledged: false },
    { id: 'alert_02', node_id: 'node_sin_05', node_name: 'SG-SIN-05', type: 'media_degraded', severity: 'warning', title: 'ChatGPT 解锁能力下降', detail: '状态由 available 变为 blocked。', created_at: ago(120), acknowledged: false },
    { id: 'alert_03', node_id: 'node_ams_06', node_name: 'NL-AMS-06', type: 'heartbeat_missing', severity: 'warning', title: '节点心跳超时', detail: '已连续 19 分钟未收到心跳。', created_at: ago(9), acknowledged: false },
  ],
  regions: { HK: 1, US: 1, JP: 1, DE: 1, SG: 1, NL: 1 },
}
