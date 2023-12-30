create table v2.jobs (
  id            v2.ksuid primary key,
	reference     text not null,
	type          text not null constraint type_not_empty_string check (type <> ''),
	args          jsonb,
	scheduled_at  timestamptz not null,
	scheduled_ctx jsonb,
  next_run_at   timestamptz not null,

	last_run_at   timestamptz,
	last_run_ctx  jsonb,

	locked_at     timestamptz,
	locked_until  timestamptz,
	locked_by     text constraint locked_by_not_empty_string check (locked_by is null or locked_by <> ''),

  -- all three fields must be set
  constraint locked_fields_consistent check (
    (locked_by is null and locked_at is null and locked_until is null) or
    (locked_by is not null and locked_at is not null and locked_until is not null)
  ),

  -- lock deadline should be in the future
  constraint deadline_in_future check (
    (locked_until is null and locked_at is null) or (locked_until > locked_at)
  ),

	error_count   int not null constraint error_count_positive check(error_count >= 0) default 0,

	successful    boolean,
	finished_at   timestamptz,

  -- both fields need to be set
  constraint successful_and_finished_at_consistent check (
    (successful is null and finished_at is null) or
    (successful is not null and finished_at is not null)
  )
);

-- ensure references are unique
create unique index jobs_reference on v2.jobs (reference); 

-- for looking up next job to process (queue.lockJob)
create index jobs_waiting on v2.jobs (next_run_at)
  where successful is null and locked_by is null;

-- for garbage collecting jobs past lock deadline (queue.unlockExpired)
create index jobs_locked on v2.jobs (locked_until)
  where locked_at is not null or locked_by is not null;

-- for looking up jobs by reference and type (queue.ByReference)
create index jobs_type_reference on v2.jobs (reference, type);

grant select, insert, update on table v2.jobs TO apisrv;
