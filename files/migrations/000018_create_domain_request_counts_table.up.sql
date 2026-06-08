CREATE TABLE IF NOT EXISTS domain_request_counts (
    id              BIGSERIAL PRIMARY KEY,
    domain          VARCHAR(512) UNIQUE NOT NULL,
    count           BIGINT NOT NULL DEFAULT 0,
    last_request_at TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_domain_request_counts_domain ON domain_request_counts(domain);