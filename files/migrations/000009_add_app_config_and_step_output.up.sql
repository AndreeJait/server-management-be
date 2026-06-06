ALTER TABLE apps
  ADD COLUMN env_vars JSONB NOT NULL DEFAULT '{}',
  ADD COLUMN volume_mounts JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN post_deploy_commands JSONB NOT NULL DEFAULT '[]';

ALTER TABLE pipeline_steps
  ADD COLUMN output TEXT;