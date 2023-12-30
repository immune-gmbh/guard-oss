alter table v2.appraisals
  alter column constraint_id drop not null;

drop view v2.devices_appraisals;
CREATE VIEW v2.devices_appraisals AS
  SELECT v2.appraisals.*, v2.devices.id AS device_id, v2.constraints.policy_id
  FROM v2.devices
  INNER JOIN v2.keys ON v2.devices.id = v2.keys.device_id
  INNER JOIN v2.appraisals ON v2.appraisals.key_id = v2.keys.id
  LEFT JOIN v2.constraints ON v2.appraisals.constraint_id = v2.constraints.id
  ORDER BY v2.appraisals.appraised_at DESC;
GRANT SELECT ON TABLE v2.devices_appraisals TO apisrv;
