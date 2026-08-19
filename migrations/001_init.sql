CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE tenants (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE nodes (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    name text NOT NULL,
    provider text,
    region_label text,
    visibility text NOT NULL DEFAULT 'private',
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE agent_keys (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    node_id uuid NOT NULL REFERENCES nodes(id),
    public_key bytea NOT NULL,
    fingerprint text NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz
);

CREATE TABLE heartbeats (
    observed_at timestamptz NOT NULL,
    tenant_id uuid NOT NULL,
    node_id uuid NOT NULL,
    agent_version text,
    source_ip inet,
    reported_ip inet,
    status jsonb NOT NULL DEFAULT '{}'
);
SELECT create_hypertable('heartbeats', by_range('observed_at'), if_not_exists => TRUE);

CREATE TABLE ip_quality_samples (
    observed_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    tenant_id uuid NOT NULL,
    node_id uuid NOT NULL,
    report_id text NOT NULL,
    ip_family smallint NOT NULL CHECK (ip_family IN (4, 6)),
    reported_ip inet,
    source_ip inet,
    asn bigint,
    country_code char(2),
    overall_risk smallint,
    scores jsonb NOT NULL DEFAULT '{}',
    factors jsonb NOT NULL DEFAULT '{}',
    media jsonb NOT NULL DEFAULT '{}',
    mail jsonb NOT NULL DEFAULT '{}',
    schema_version text NOT NULL,
    collector_version text,
    parser_version text NOT NULL,
    raw_report jsonb,
    report_sha256 bytea NOT NULL,
    PRIMARY KEY (observed_at, tenant_id, report_id)
);
SELECT create_hypertable('ip_quality_samples', by_range('observed_at'), if_not_exists => TRUE);
CREATE INDEX ix_samples_node_time ON ip_quality_samples (tenant_id, node_id, observed_at DESC);

ALTER TABLE nodes ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_nodes ON nodes USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

SELECT add_retention_policy('heartbeats', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('ip_quality_samples', INTERVAL '180 days', if_not_exists => TRUE);
