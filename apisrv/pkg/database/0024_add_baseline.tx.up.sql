alter table v2.devices
  add column baseline jsonb not null default '{"type": "baseline/1"}'::jsonb;

alter table v2.devices
  alter column baseline drop default;
