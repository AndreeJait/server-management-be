ALTER TABLE apps
  DROP COLUMN post_deploy_commands,
  DROP COLUMN volume_mounts,
  DROP COLUMN env_vars;

ALTER TABLE pipeline_steps
  DROP COLUMN output;