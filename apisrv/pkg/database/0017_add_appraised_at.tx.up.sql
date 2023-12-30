drop view v2.last_appraisals;
drop view v2.devices_appraisals;

alter table v2.appraisals
  rename column created_at to received_at;

alter table v2.appraisals
  alter column received_at set data type timestamptz using
    received_at::timestamp without time zone at time zone 'Europe/Berlin',
  add column appraised_at timestamptz;

update v2.appraisals
  set appraised_at = received_at;

alter table v2.appraisals
  alter column appraised_at set not null,
  add constraint received_before_appraised check (appraised_at >= received_at);

create index appraisals_appraised_at on v2.appraisals (appraised_at);

create view v2.last_appraisals as
  select distinct on (v2.keys.device_id)
    v2.appraisals.*,
    v2.keys.device_id as device_id
  from v2.keys
  left join v2.appraisals on appraisals.id = (
    select id from v2.appraisals
    where appraisals.key_id = keys.id
    order by appraisals.appraised_at desc
    limit 1)
  order by v2.keys.device_id, v2.keys.id desc;

create view v2.devices_appraisals as
  select v2.appraisals.*, v2.devices.id as device_id, v2.constraints.policy_id
  from v2.devices
  inner join v2.keys on v2.devices.id = v2.keys.device_id
  inner join v2.appraisals on v2.appraisals.key_id = v2.keys.id
  inner join v2.constraints on v2.appraisals.constraint_id = v2.constraints.id
  order by v2.appraisals.appraised_at desc;

grant select on table v2.devices_appraisals to apisrv;
grant select on table v2.last_appraisals to apisrv;
