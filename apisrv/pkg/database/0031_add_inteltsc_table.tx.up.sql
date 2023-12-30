create table v2.inteltsc (
  id           v2.ksuid primary key,
  reference    text not null,
  created_at   timestamptz not null,

  data         xml null,
  certificates text[] null,
  finished_at  timestamptz null
               constraint after_submit check (finished_at is null or finished_at >= created_at),

  constraint   finished_at_and_data check (((data is null) and (certificates is null)) or (finished_at is not null))
);

-- for looking up by (finished) job
create unique index inteltsc_reference on v2.inteltsc (reference);

grant select, insert, update on table v2.inteltsc TO apisrv;
