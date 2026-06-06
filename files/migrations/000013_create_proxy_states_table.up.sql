CREATE TABLE IF NOT EXISTS proxy_states (
    id BIGSERIAL PRIMARY KEY,
    app_id VARCHAR(36) UNIQUE NOT NULL REFERENCES apps(app_id) ON DELETE CASCADE,
    active_slot VARCHAR(10) NOT NULL DEFAULT 'blue',
    blue_container_id VARCHAR(64),
    green_container_id VARCHAR(64),
    blue_target VARCHAR(255),
    green_target VARCHAR(255),
    traffic_percent INTEGER NOT NULL DEFAULT 100,
    health_check_path VARCHAR(512) NOT NULL DEFAULT '/health',
    health_check_interval INTEGER NOT NULL DEFAULT 10,
    status VARCHAR(20) NOT NULL DEFAULT 'idle',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_proxy_states_app_id ON proxy_states(app_id);
CREATE INDEX IF NOT EXISTS idx_proxy_states_status ON proxy_states(status);