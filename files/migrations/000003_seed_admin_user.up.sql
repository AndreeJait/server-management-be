-- Seed admin user with default password "admin123" (bcrypt hash, cost 10)
-- IMPORTANT: Change this password immediately after first login
INSERT INTO users (email, password, name, created_at, updated_at)
VALUES ('admin@server-management.local', '$2a$10$x5LBAVnOlfCYiYqNKVunAO3F3TSgvij.TUe88w9cZTgVMxexFouFm', 'Admin', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_roles (user_id, role)
SELECT id, 'admin' FROM users WHERE email = 'admin@server-management.local'
ON CONFLICT (user_id, role) DO NOTHING;