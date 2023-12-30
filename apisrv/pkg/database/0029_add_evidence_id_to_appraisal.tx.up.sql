drop view v2.last_appraisals;
drop view v2.devices_appraisals;

alter table v2.appraisals
  add column evidence_id v2.ksuid_ref references v2.evidence (id) null;

create index appraisal_evidence on v2.appraisals (evidence_id);

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
  select
    v2.appraisals.*,
    v2.devices.id as device_id,
    v2.constraints.policy_id
  from v2.devices
  inner join v2.keys on v2.devices.id = v2.keys.device_id
  inner join v2.appraisals on v2.appraisals.key_id = v2.keys.id
  left join v2.constraints on v2.appraisals.constraint_id = v2.constraints.id
  order by v2.appraisals.appraised_at desc;

grant select on table v2.devices_appraisals to apisrv;
grant select on table v2.last_appraisals to apisrv;
