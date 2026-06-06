CREATE TABLE deployments (
    id SERIAL PRIMARY KEY,
    app_id VARCHAR(36) NOT NULL REFERENCES apps(app_id) ON DELETE CASCADE,
    image VARCHAR(512) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    container_id VARCHAR(64),
    container_name VARCHAR(255),
    error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deployments_app_id ON deployments(app_id);
CREATE INDEX idx_deployments_status ON deployments(status);