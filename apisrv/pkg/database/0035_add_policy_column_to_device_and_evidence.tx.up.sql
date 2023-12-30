alter table v2.devices
  add column policy jsonb not null
    default '{"type": "policy/1", "endpoint_protection": "if-present", "intel_tsc": "if-present"}'::jsonb
    check (policy::jsonb ? 'type' and policy->>'type' <> '');

alter table v2.evidence
  add column policy jsonb null
    check (policy is null or (policy::jsonb ? 'type' and policy->>'type' <> ''));
