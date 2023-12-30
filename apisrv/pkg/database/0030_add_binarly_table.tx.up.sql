create table v2.binarly (
  id           v2.ksuid primary key,
  reference    text not null,
  created_at   timestamptz not null,

  external_id  text null
               constraint external_id_not_empty check (external_id <> ''),
  submitted_at timestamptz null,
               constraint after_create check (submitted_at is null or submitted_at >= created_at),
  constraint   external_id_and_submitted check ((external_id is null) = (submitted_at is null)),

  report       jsonb null
               constraint report_typed check (report is null or (report::jsonb ? 'type' and report->>'type' <> '')),
  finished_at  timestamptz null
               constraint after_submit check (finished_at is null or finished_at >= submitted_at),

  constraint   finished_at_and_report check ((report is null) = (finished_at is null))
);


-- for looking up by (finished) job
create unique index binarly_reference on v2.binarly (reference);
create unique index binarly_external_id on v2.binarly (external_id);

grant select, insert, update on table v2.binarly TO apisrv;
