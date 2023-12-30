CREATE TABLE v2.appraisals (
  id BIGSERIAL PRIMARY KEY,

  created_at TIMESTAMP NOT NULL,

  expires TIMESTAMP NOT NULL,

  CHECK (created_at < expires),

  verdict BOOLEAN NOT NULL,

  -- {
  --   type: 'evidence/1'
  --
  --   ...
  -- }
  evidence JSONB NOT NULL
    CHECK (evidence ? 'type' AND evidence->>'type' <> ''), -- Needs to have a type

  -- {
  --   type: 'report/1'
  --
  --   ...
  -- }
  report JSONB NOT NULL
    CHECK (report ? 'type' AND report->>'type' <> ''), -- Needs to have a type

  key_id BIGINT REFERENCES v2.keys (id) NOT NULL,
  constraint_id BIGINT REFERENCES v2.constraints (id) NOT NULL
);

-- Start appraisal primary keys at 1000
ALTER SEQUENCE v2.appraisals_id_seq RESTART WITH 1000;

CREATE FUNCTION v2.appraisals_prevent_cross_organization() RETURNS TRIGGER AS $$
DECLARE
  mismatch integer;
BEGIN
  SELECT 1 INTO mismatch WHERE EXISTS (
    SELECT 1
      FROM v2.appraisals
      JOIN v2.devices ON devices.id =
        (SELECT device_id FROM v2.keys WHERE id = appraisals.key_id)
      JOIN v2.policies ON policies.id =
        (SELECT policy_id FROM v2.constraints WHERE id = appraisals.constraint_id)
      WHERE devices.organization_id <> policies.organization_id);

  IF mismatch IS NOT NULL THEN
    RAISE 'cannot associate devices with policies of another organization';
  END IF;
 
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER appraisals_prevent_cross_organization_trigger
  AFTER UPDATE OR INSERT ON v2.appraisals
  EXECUTE PROCEDURE v2.appraisals_prevent_cross_organization();

CREATE INDEX appraisals_key_created ON v2.appraisals (key_id, created_at);

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

CREATE FUNCTION v2.device_state(retired boolean, has_unretired boolean, aik_id BIGINT, verdict boolean, expires TIMESTAMP, now TIMESTAMP) RETURNS text AS $$
BEGIN
  CASE
    WHEN retired = TRUE AND NOT has_unretired THEN
      RETURN 'resurrectable';
    WHEN retired = TRUE AND has_unretired THEN
      RETURN 'retired';
    WHEN aik_id IS NULL THEN 
      RETURN 'new';
    WHEN verdict IS NULL THEN 
      RETURN 'unseen';
    WHEN expires <= now THEN
      RETURN 'outdated';
    WHEN verdict = TRUE THEN
      RETURN 'trusted';
    ELSE
      RETURN 'vulnerable';
  END CASE;
END;
$$ LANGUAGE plpgsql;

GRANT SELECT, INSERT ON TABLE v2.appraisals TO apisrv;
GRANT SELECT ON TABLE v2.last_appraisals TO apisrv;
GRANT SELECT ON TABLE v2.devices_appraisals TO apisrv;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA v2 TO apisrv;
