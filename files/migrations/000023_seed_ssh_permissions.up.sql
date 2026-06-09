INSERT INTO role_permissions (role, permission) VALUES
    ('admin', 'ssh:read'),
    ('admin', 'ssh:write'),
    ('admin', 'ssh:connect'),
    ('operator', 'ssh:read'),
    ('operator', 'ssh:write'),
    ('operator', 'ssh:connect'),
    ('viewer', 'ssh:read')
ON CONFLICT (role, permission) DO NOTHING;