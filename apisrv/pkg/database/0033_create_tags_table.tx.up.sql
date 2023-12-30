create extension pg_trgm;

create table v2.tags (
  id v2.ksuid primary key,
  key text not null check (key <> ''),
  value text null,
  metadata jsonb not null,
  organization_id bigint references v2.organizations (id) not null,

  constraint metadata_typed check (metadata::jsonb ? 'type' and metadata->>'type' <> '')
);

create unique index tags_unique_key on v2.tags (key, organization_id);
create index tags_organization on v2.tags (organization_id);
create index tags_organization_id on v2.tags (organization_id desc, id desc);
create index tags_key_trigram on v2.tags using gist (key gist_trgm_ops);

create table v2.devices_tags (
  device_id bigint references v2.devices (id) not null,
  tag_id v2.ksuid references v2.tags (id) not null,

  primary key (device_id, tag_id)
);

create index devices_tags_device on v2.devices_tags (device_id desc);
create index devices_tags_tag on v2.devices_tags (tag_id desc);

grant select, insert, update, delete on table v2.tags TO apisrv;
grant select, insert, update, delete on table v2.devices_tags TO apisrv;
