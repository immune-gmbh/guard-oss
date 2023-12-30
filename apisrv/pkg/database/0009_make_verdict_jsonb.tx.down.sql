DROP VIEW v2.devices_appraisals;
DROP VIEW v2.last_appraisals;

ALTER TABLE v2.appraisals
  DROP CONSTRAINT verdict_typed,
  ALTER COLUMN verdict SET DATA TYPE boolean USING
    CASE WHEN verdict->>'type' = 'verdict/1' THEN (verdict->'result')::boolean
         ELSE false
    END;

CREATE VIEW v2.last_appraisals AS
  SELECT DISTINCT ON (v2.keys.device_id)
    v2.appraisals.*,
    v2.keys.device_id AS device_id
  FROM v2.keys
  LEFT JOIN v2.appraisals ON appraisals.id = (
    SELECT id FROM v2.appraisals
    WHERE appraisals.key_id = keys.id
    ORDER BY appraisals.created_at DESC
    LIMIT 1)
  ORDER BY v2.keys.device_id, v2.keys.id DESC;

CREATE VIEW v2.devices_appraisals AS
  SELECT v2.appraisals.*, v2.devices.id AS device_id, v2.constraints.policy_id
  FROM v2.devices
  INNER JOIN v2.keys ON v2.devices.id = v2.keys.device_id
  INNER JOIN v2.appraisals ON v2.appraisals.key_id = v2.keys.id
  INNER JOIN v2.constraints ON v2.appraisals.constraint_id = v2.constraints.id
  ORDER BY v2.appraisals.created_at DESC;

GRANT SELECT ON TABLE v2.last_appraisals TO apisrv;
GRANT SELECT ON TABLE v2.devices_appraisals TO apisrv;
