CREATE TABLE registry_credentials (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    scope VARCHAR(10) NOT NULL DEFAULT 'project',
    registry_url VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_registry_credentials_project_id ON registry_credentials(project_id);
CREATE INDEX idx_registry_credentials_scope ON registry_credentials(scope);