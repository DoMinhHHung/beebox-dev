CREATE TABLE field_definition_schemas (
    project_id TEXT NOT NULL REFERENCES projects(id),
    version INTEGER NOT NULL,
    fields JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, version)
);