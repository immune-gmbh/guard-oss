create table v2.evidence (
  id            v2.ksuid primary key,
  received_at   timestamptz not null,
  -- signer Name, weak ref to v2.keys.fpr
  signed_by     bytea not null
                constraint signer_not_empty check (signed_by <> ''),
  values        jsonb not null
                constraint values_typed check (values::jsonb ? 'type' and values->>'type' <> ''),
  baseline      jsonb not null
                constraint baseline_typed check (baseline::jsonb ? 'type' and baseline->>'type' <> ''),
  image_ref     text null,
  binarly_ref   text null,
  inteltsc_ref  text null,

  check (binarly_ref is null or image_ref is not null)
);

-- for looking up by (finished) job
create index evidence_image_ref on v2.evidence (image_ref);
create index evidence_binarly_ref on v2.evidence (binarly_ref);
create index evidence_inteltsc_ref on v2.evidence (inteltsc_ref);

-- for looking up evidence/pending appraisals by device
create index evidence_signed_by on v2.evidence (signed_by);

-- inmutable
grant select, insert on table v2.evidence TO apisrv;
