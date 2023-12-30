create table v2.blobs (
  id            v2.ksuid primary key,
	snowflake     char(6) not null
                constraint snowflake_not_empty_string check (snowflake <> ''),
	namespace     text not null
                constraint namespace_not_empty_string check (namespace <> ''),
                -- lower case hex string
	digest        char(64) not null
                constraint digest_hex_string check (digest similar to '[a-f0-9]{64}'),
  metadata      jsonb not null
                constraint metadata_typed check (metadata::jsonb ? 'type'),
  created_at    timestamptz not null
);

-- ensure snowflakes are unique
create unique index blobs_snowflake on v2.blobs (snowflake); 

-- for looking up next job to process (queue.lockJob)
create index blobs_ns_digest on v2.blobs (namespace, digest);

grant select, insert, update on table v2.blobs TO apisrv;
