CREATE TABLE users (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    username text NOT NULL,
    display_name text NOT NULL,
    password_hash bytea NOT NULL,
    role text NOT NULL CHECK (role IN ('admin', 'user')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, username)
);

CREATE TABLE sessions (
    token_hash bytea PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_sessions_expiry ON sessions (expires_at);

CREATE TABLE password_resets (
    token_hash bytea PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tenant_settings (
    tenant_id uuid PRIMARY KEY REFERENCES tenants(id),
    registration_enabled boolean NOT NULL DEFAULT false,
    updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE nodes ADD COLUMN owner_user_id uuid REFERENCES users(id);
CREATE INDEX ix_nodes_owner ON nodes (tenant_id, owner_user_id) WHERE deleted_at IS NULL;

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_users ON users USING (tenant_id = current_setting('app.tenant_id', true)::uuid);
