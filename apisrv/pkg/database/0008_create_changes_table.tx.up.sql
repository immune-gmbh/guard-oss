CREATE TYPE v2.changes_type AS ENUM (
  'enroll',
  'resurrect',
  'rename',
  'tag',
  'associate',
  'template',
  'new',
  'instanciate',
  'revoke',
  'retire'
);

CREATE TABLE v2.changes (
  id BIGSERIAL PRIMARY KEY,

  organization_id BIGINT REFERENCES v2.organizations NOT NULL,

  type v2.changes_type NOT NULL,

  actor TEXT NULL CHECK (actor <> ''),

  comment TEXT NULL CHECK (comment <> ''),

  device_id BIGINT REFERENCES v2.devices NULL,

  policy_id BIGINT REFERENCES v2.policies NULL,

  constraint_id BIGINT REFERENCES v2.constraints NULL,

  key_id BIGINT REFERENCES v2.keys NULL,

  timestamp TIMESTAMP NOT NULL,

  CHECK (
    (type = 'enroll' AND device_id IS NOT NULL AND key_id IS NOT NULL) OR
    (type = 'resurrect' AND device_id IS NOT NULL) OR
    (type = 'rename' AND (device_id IS NOT NULL OR policy_id IS NOT NULL)) OR
    (type = 'tag' AND device_id IS NOT NULL) OR
    (type = 'associate' AND device_id IS NOT NULL AND policy_id IS NOT NULL) OR
    (type = 'template' AND policy_id IS NOT NULL) OR
    (type = 'new' AND policy_id IS NOT NULL AND constraint_id IS NOT NULL) OR
    (type = 'instanciate' AND policy_id IS NOT NULL AND constraint_id IS NOT NULL) OR
    (type = 'revoke' AND (policy_id IS NOT NULL)) OR
    (type = 'retire' AND (device_id IS NOT NULL)))
);

-- Start device primary keys at 1000
ALTER SEQUENCE v2.changes_id_seq RESTART WITH 1000;

--
-- Index devices table by timestamp
--
CREATE INDEX changes_timestamp_index ON v2.changes (timestamp);

CREATE FUNCTION v2.changes_prevent_cross_organization() RETURNS TRIGGER AS $$
DECLARE
  mismatch integer;
BEGIN
  SELECT 1 INTO mismatch WHERE EXISTS (
    SELECT 1
      FROM v2.changes
      LEFT JOIN v2.constraints ON constraints.id = changes.constraint_id
      LEFT JOIN v2.policies ON policies.id = v2.constraints.policy_id
      LEFT JOIN v2.keys ON keys.id = changes.key_id
      LEFT JOIN v2.devices ON devices.id = v2.keys.device_id
      WHERE devices.organization_id <> policies.organization_id
    UNION SELECT 1
      FROM v2.changes
      LEFT JOIN v2.devices ON devices.id = changes.device_id
      LEFT JOIN v2.policies ON policies.id = changes.policy_id
      WHERE devices.organization_id <> policies.organization_id);

  IF mismatch IS NOT NULL THEN
    RAISE 'cannot have a change spanning multiple organization';
  END IF;
 
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER changes_prevent_cross_organization_trigger
  AFTER UPDATE OR INSERT ON v2.changes
  EXECUTE PROCEDURE v2.changes_prevent_cross_organization();

CREATE VIEW v2.last_device_changes AS
  SELECT changes.*
  FROM v2.devices AS devs
  LEFT JOIN v2.changes ON changes.id = (
    SELECT id FROM v2.changes
    WHERE changes.device_id = devs.id
    ORDER BY changes.timestamp DESC
    LIMIT 1);

CREATE VIEW v2.last_policy_changes AS
  SELECT changes.*
  FROM v2.policies AS pols
  LEFT JOIN v2.changes ON changes.id = (
    SELECT id FROM v2.changes
    WHERE changes.policy_id = pols.id
    ORDER BY changes.timestamp DESC
    LIMIT 1);

GRANT SELECT, INSERT ON TABLE v2.changes TO apisrv;
GRANT SELECT ON TABLE v2.last_device_changes TO apisrv;
GRANT SELECT ON TABLE v2.last_policy_changes TO apisrv;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA v2 TO apisrv;
