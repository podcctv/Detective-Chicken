# 鸡探长（Detective Chicken）

鸡探长（Detective Chicken）是一套面向 VPS 舰队的 IP 质量持续监测与自动上报平台。它持续巡查每台“小鸡”的公网 IP 身份、风险评分、流媒体与 AI 解锁能力及 DNSBL 变化，并把一次性终端检测升级为可追踪、可告警的长期 IP 资产档案。

平台将 `IP.Check.Place` 作为可替换的外部检测器，通过自有 Agent、Canonical JSON、Ed25519 设备身份和变化型告警构成完整监测闭环。

![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-475569)

## 当前可用能力

- 舰队总览：总节点、在线、异常、高风险、IP 变更、解锁下降和 DNSBL 新增。
- 账户权限：首个注册账户自动成为管理员；管理员控制后续注册、用户角色和一次性密码重置链接。
- 节点资产：按创建者隔离，IPv4/IPv6 默认隐藏最后两段；节点所有者和管理员可按需查看完整 IP。
- 节点详情：质量评分、风险因子、解锁矩阵、趋势、告警、网络身份和最近采集器信息。
- 质量首页：小鸡质量排行榜、Netflix/ChatGPT 解锁矩阵、风险趋势和变化型告警。
- 公开看板：完成首个管理员注册后，未登录访客可查看脱敏 IP、质量排名和核心解锁结果；完整 IP、Agent、告警和控制操作仍仅限登录账户。
- 变化追踪：风险分趋势、IP/ASN/解锁能力变化和变化型告警。
- Agent 身份：一次性注册凭证，本机生成 Ed25519 私钥，请求摘要、时间戳、nonce 和重放拦截。
- IPQuality 适配器：自动识别可用 IPv4/IPv6 并行执行，默认 `-j -p`；只要上游输出可解析 JSON，即使上游返回非零状态也保留有效报告。
- 兼容安装：按 Debian/Alpine/RHEL/Arch、AMD64/ARM64/ARMv7 和 PVE/独服/LXC/Docker/Podman/Incus 生成脚本，自动选择 systemd、OpenRC/cron 或容器循环。
- 持久化：默认将账户、会话、节点、Agent、报告与设置保存到权限受限的 JSON 快照；PostgreSQL/TimescaleDB 契约保留为规模化升级路径。
- 数据契约：OpenAPI 3.1、JSON Schema 2020-12、PostgreSQL/TimescaleDB 初始化脚本。

## 本地运行

需要 Go 1.24+ 与 Node.js 22+。

```bash
go run ./cmd/server
```

另一个终端：

```bash
cd web
npm ci
npm run dev
```

打开 `http://localhost:4173`。Vite 会把 `/api` 代理到 `http://127.0.0.1:8080`；API 不可用时，前端会明确显示连接错误，不会用演示数据冒充实时状态。

首次打开控制台时创建的第一个账户会自动成为管理员，随后注册默认关闭。需要展示内置演示节点时，可在启动 API 前设置 `DETECTIVE_CHICKEN_SEED_DEMO=true`；默认不注入演示数据。

## Docker Compose

```bash
docker compose up --build
```

控制台位于 `http://localhost:8088`。PostgreSQL/TimescaleDB 生产数据契约是可选 profile：

```bash
docker compose --profile production-data up --build
```

Compose 默认使用 `detective_data` 卷保存 `/data/state.json`。备份该文件即可迁移单实例数据；规模化部署应按 [架构说明](docs/architecture.md) 切换到 PostgreSQL repository 和分布式 nonce/session cache。数据库迁移已经固定账户、核心采样表、hypertable、RLS 与 retention 基线。

### OVH + Caddy

OVH 部署使用独立 Compose 文件，不发布宿主机端口；Web 容器通过外部 Docker 网络 `pt-suite_default` 接入已有 Caddy：

```bash
docker compose -f compose.ovh.yml up -d --build
```

Caddy 站点配置：

```caddyfile
detective.428048.xyz {
	import common_secure
	reverse_proxy detective-chicken-web:80
}
```

部署前需要确认 `pt-suite_default` 已存在，且域名 A/AAAA 记录指向服务器。配置变更后先执行 `caddy validate`，再热加载 Caddy。

## Agent 流程

先在控制台“添加 VPS”创建一次性 token，然后在目标 Linux VPS 上安装 Agent。也可以手动构建和注册：

```bash
go build -o detective-chicken-agent ./cmd/agent
./detective-chicken-agent --server https://detective-chicken.example.com --token 'et_xxx' enroll
./detective-chicken-agent heartbeat
./detective-chicken-agent --family auto scan
./detective-chicken-agent --family 4 scan
./detective-chicken-agent --family 6 scan
```

Agent 配置默认写入 `/etc/detective-chicken/agent.json`，权限为 `0600`。安装脚本会立即完成第一次心跳并并行扫描可用的 IPv4/IPv6，实测通常约 1–3 分钟，单次最多 8 分钟。心跳固定每 2 分钟检查一次服务器指令；质量扫描默认每 6 小时，可在创建安装命令时选择，也可在节点详情中修改为 1 小时到 1 周。服务器调度会在下一次心跳生效，因此“立即扫描”通常在 2 分钟内开始。

## API 与测试

- OpenAPI: [`openapi/openapi.yaml`](openapi/openapi.yaml)
- Canonical Schema: [`schemas/ip-quality-report-v1.schema.json`](schemas/ip-quality-report-v1.schema.json)
- 数据库迁移: [`migrations/001_init.sql`](migrations/001_init.sql)

```bash
go test ./...
go vet ./...
cd web && npm run build
```

## 许可证与合规边界

本仓库代码使用 MIT License。IP.Check.Place/IPQuality 是外部独立程序，其上游仓库采用 AGPL-3.0；本项目没有复制或修改其源码。公开或商业部署前仍需评估上游许可证、第三方风险数据库与媒体服务条款、IP 数据最小化、保存期限以及所在地区的备案与地图展示要求。
