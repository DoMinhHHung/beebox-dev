CREATE TABLE project_capabilities (
    project_id TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    module_id TEXT NOT NULL,
    module_version TEXT NOT NULL,
    capability_id TEXT NOT NULL,
    capability_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, module_id, capability_id)
);

CREATE TABLE project_capability_fields (
    project_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    capability_id TEXT NOT NULL,
    field_id TEXT NOT NULL,
    field_version TEXT NOT NULL,
    capability_version TEXT NOT NULL,
    PRIMARY KEY (project_id, module_id, capability_id, field_id, field_version),
    FOREIGN KEY (project_id, module_id, capability_id)
        REFERENCES project_capabilities (project_id, module_id, capability_id)
        ON DELETE CASCADE
);