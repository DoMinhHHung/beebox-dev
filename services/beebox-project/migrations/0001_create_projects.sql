CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_projects_organization_id ON projects (organization_id);