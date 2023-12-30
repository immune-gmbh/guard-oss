DROP VIEW v2.devices_appraisals;


ALTER TABLE v2.appraisals DROP COLUMN constraint_id;


CREATE VIEW v2.devices_appraisals AS
SELECT v2.appraisals.*,
  v2.devices.id AS device_id
FROM v2.devices
  INNER JOIN v2.keys ON v2.devices.id = v2.keys.device_id
  INNER JOIN v2.appraisals ON v2.appraisals.key_id = v2.keys.id
ORDER BY v2.appraisals.appraised_at DESC;


GRANT SELECT ON TABLE v2.devices_appraisals TO apisrv;