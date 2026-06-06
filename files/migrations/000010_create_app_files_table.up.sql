CREATE TABLE app_files (
    id BIGSERIAL PRIMARY KEY,
    app_id VARCHAR(255) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_app_files_app_id ON app_files(app_id);