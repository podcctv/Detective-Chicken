# Architecture

Detective Chicken（鸡探长）treats IP.Check.Place as an external collector process, not as the platform identity or data model. The platform continuously investigates VPS public IP identity, risk and unlock-capability changes while keeping its own stable Agent and canonical data contracts.

```text
IP.Check.Place -4/-6 -j -p
  -> tolerant adapter + raw upstream payload
  -> canonical report v1
  -> Ed25519 signed HTTPS request
  -> ingest validation + nonce/replay guard
  -> time-series samples
  -> fleet dashboard + change-based alerts
```

## Identity boundary

- The first registered user becomes the initial administrator and public registration closes immediately.
- Administrators control registration, roles and one-use password reset links. Non-admin users can only access nodes they enrolled.
- The anonymous public dashboard exposes only masked addresses, aggregate quality scores, rankings and unlock status. Node detail, complete addresses, alerts and control operations require a session.
- A node is the business asset.
- An agent is one installation of the collector.
- The agent generates its Ed25519 key pair locally.
- A ten-minute, one-use enrollment token registers only the public key.
- Every ingest request binds method, path, content digest, timestamp, nonce and key ID.
- The server accepts five minutes of clock skew and remembers nonces for ten minutes.

## Reporting schedule and dual stack

- Heartbeats run every two minutes and return server-side scan directives and queued manual scans.
- Each node stores its own quality-scan interval. The default is six hours; enrollment and authenticated node settings accept 30 minutes through one week.
- The Agent detects usable local IPv4 and global IPv6 interfaces and scans available families concurrently. The server keeps the latest report for each family in node detail.
- A newly enrolled node is immediately due. Installation waits for the first scan and prints a 1–3 minute expectation instead of hiding the result in a background log.
- The adapter parses stdout before considering the upstream process exit code because IP.Check.Place can emit a complete JSON report and still exit with status 1.

## Storage path

The single-instance runtime persists users, hashed passwords, hashed sessions/reset tokens, nodes, Agent keys, reports and settings in an atomic JSON snapshot. Docker mounts it at `/data/state.json`; the file is created with mode `0600`. Demo data is opt-in through `DETECTIVE_CHICKEN_SEED_DEMO=true`.

`migrations/001_init.sql` and `migrations/002_accounts.sql` are the scale-out contract for PostgreSQL 18 and TimescaleDB. The storage boundary stays behind the API package so PostgreSQL can replace the JSON repository without changing Agent or frontend contracts.

Raw upstream JSON is retained on canonical reports. This allows a newer parser to rebuild normalized samples after an upstream schema change without executing a new third-party scan.

## Scale path

1. Replace the single-instance JSON snapshot with PostgreSQL/TimescaleDB while preserving tenant and owner authorization.
2. Move raw reports to S3-compatible object storage.
3. Add Redis for distributed nonce and rate-limit state.
4. Insert NATS JetStream after ingest when synchronous writes become a bottleneck.
5. Add ClickHouse only when high-dimensional fleet analysis exceeds PostgreSQL's operational envelope.
