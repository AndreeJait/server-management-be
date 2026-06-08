INSERT INTO settings (section, key, value, type) VALUES
  ('proxy',  'enabled',            'true',                 'bool'),
  ('proxy',  'health_check_path',  '/health',              'string'),
  ('proxy',  'shift_interval_sec', '30',                   'int'),
  ('proxy',  'tunnel_service_url', 'http://server-management-be:8080', 'string'),
  ('proxy',  'rate_limit_rps',     '0',                    'int'),
  ('docker', 'network',            '',                      'string'),
  ('docker', 'host_base',          '/home/andree/docker/app-server', 'string')
ON CONFLICT (section, key) DO NOTHING;