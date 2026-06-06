CREATE TABLE apps (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    framework_preset VARCHAR(50) NOT NULL DEFAULT 'custom',
    container_count INTEGER NOT NULL DEFAULT 0,
    deploy_token VARCHAR(255) NOT NULL,
    app_id VARCHAR(36) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_apps_project_id ON apps(project_id);
CREATE INDEX idx_apps_app_id ON apps(app_id);