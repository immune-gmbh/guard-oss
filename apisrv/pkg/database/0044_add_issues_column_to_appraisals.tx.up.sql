alter table v2.appraisals
  add column issues jsonb null
    check (issues is null or (issues::jsonb ? 'type' and issues->>'type' <> ''));
