CREATE TABLE project_configuration_versions (
    project_id TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    module_id TEXT NOT NULL,
    module_version TEXT NOT NULL,
    capability_id TEXT NOT NULL,
    capability_version TEXT NOT NULL,
    data_fields JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, version)
);

CREATE TABLE project_configuration_rollouts (
    project_id TEXT PRIMARY KEY REFERENCES projects (id) ON DELETE CASCADE,
    desired_version INTEGER NOT NULL DEFAULT 0,
    applied_version INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
