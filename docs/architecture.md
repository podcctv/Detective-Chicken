# Architecture

JiJian treats IP.Check.Place as an external collector process, not as the platform identity or data model.

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

- A node is the business asset.
- An agent is one installation of the collector.
- The agent generates its Ed25519 key pair locally.
- A ten-minute, one-use enrollment token registers only the public key.
- Every ingest request binds method, path, content digest, timestamp, nonce and key ID.
- The server accepts five minutes of clock skew and remembers nonces for ten minutes.

## Storage path

The current MVP uses an in-process store with deterministic demo data so the complete product can be evaluated without infrastructure. `migrations/001_init.sql` is the production data contract for PostgreSQL 18 and TimescaleDB. The storage interface is intentionally kept behind the API package so a PostgreSQL repository can replace it without changing the Agent or frontend contracts.

Raw upstream JSON is retained on canonical reports. This allows a newer parser to rebuild normalized samples after an upstream schema change without executing a new third-party scan.

## Scale path

1. Replace the MVP store with PostgreSQL/TimescaleDB and apply tenant-scoped authorization.
2. Move raw reports to S3-compatible object storage.
3. Add Redis for distributed nonce and rate-limit state.
4. Insert NATS JetStream after ingest when synchronous writes become a bottleneck.
5. Add ClickHouse only when high-dimensional fleet analysis exceeds PostgreSQL's operational envelope.
