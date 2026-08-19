# 鸡鉴

鸡鉴是一个面向 VPS 舰队的 IP 质量持续监测与自动上报平台。它把 `IP.Check.Place` 作为可替换的外部检测器，通过自己的 Agent、Canonical JSON、Ed25519 设备身份和变化型告警，把一次性终端报告变成可追踪的 IP 资产档案。

![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-475569)

## 当前可用能力

- 舰队总览：总节点、在线、异常、高风险、IP 变更、解锁下降和 DNSBL 新增。
- 节点资产：IP 默认脱敏，支持风险、ASN、地区、媒体解锁、黑名单和心跳状态。
- 变化追踪：风险分趋势、IP/ASN/解锁能力变化和变化型告警。
- Agent 身份：一次性注册凭证，本机生成 Ed25519 私钥，请求摘要、时间戳、nonce 和重放拦截。
- IPQuality 适配器：IPv4/IPv6 分开执行，默认 `-j -p`，宽松解析并保留原始 JSON。
- 运维安装：心跳与完整扫描拆分的 systemd service/timer，扫描带随机抖动。
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

打开 `http://localhost:4173`。Vite 会把 `/api` 代理到 `http://127.0.0.1:8080`；API 不可用时，前端会明确显示并使用演示数据。

## Docker Compose

```bash
docker compose up --build
```

控制台位于 `http://localhost:8088`。PostgreSQL/TimescaleDB 生产数据契约是可选 profile：

```bash
docker compose --profile production-data up --build
```

当前 `v0.1.0` 为可运行 MVP，API 使用进程内演示存储；生产上线前应按 [架构说明](docs/architecture.md) 接入 PostgreSQL repository、用户认证和分布式 nonce cache。数据库迁移已经固定核心表、hypertable、RLS 与 retention 基线。

## Agent 流程

先在控制台“添加 VPS”创建一次性 token，然后在目标 Linux VPS 上安装 Agent。也可以手动构建和注册：

```bash
go build -o jijian-agent ./cmd/agent
./jijian-agent --server https://jijian.example.com --token 'et_xxx' enroll
./jijian-agent heartbeat
./jijian-agent --family 4 scan
./jijian-agent --family 6 scan
```

Agent 配置默认写入 `/etc/jijian/agent.json`，权限为 `0600`。完整扫描会访问多个第三方服务，不应高频执行；默认建议 6–24 小时并增加随机抖动，心跳则保持 1–5 分钟。

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
