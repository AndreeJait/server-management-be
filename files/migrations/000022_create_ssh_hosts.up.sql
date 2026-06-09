CREATE TABLE ssh_hosts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 22,
    username VARCHAR(255) NOT NULL,
    auth_method VARCHAR(20) NOT NULL DEFAULT 'password',
    password TEXT NOT NULL DEFAULT '',
    private_key TEXT NOT NULL DEFAULT '',
    owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ssh_hosts_owner_id ON ssh_hosts(owner_id);