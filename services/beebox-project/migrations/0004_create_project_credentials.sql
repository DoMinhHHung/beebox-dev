CREATE TABLE project_credentials (
    project_id TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    secret_hash TEXT NOT NULL,
    status TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, id),
    UNIQUE (secret_hash)
);

CREATE INDEX idx_project_credentials_verify
    ON project_credentials (project_id, kind, status, secret_hash);
