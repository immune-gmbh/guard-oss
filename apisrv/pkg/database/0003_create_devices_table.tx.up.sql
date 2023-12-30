-- We use the hstore data type for PCR lists and device attributes
CREATE EXTENSION IF NOT EXISTS hstore;

CREATE TABLE v2.devices (
  id BIGSERIAL PRIMARY KEY,

  cookie TEXT NULL,

  name TEXT NOT NULL CHECK (name <> ''),

  attributes HSTORE NOT NULL CHECK (NOT attributes ? '') DEFAULT hstore(''),

  hwid BYTEA NOT NULL CHECK (hwid <> ''),

  fpr BYTEA NOT NULL CHECK (fpr <> ''), 

  organization_id BIGINT REFERENCES v2.organizations (id) NOT NULL,

  -- can only be set if retired is TRUE
  replaced_by BIGINT REFERENCES v2.devices (id) NULL
    CHECK (id IS DISTINCT FROM replaced_by AND (retired OR replaced_by IS NULL))
    DEFAULT NULL,

  retired BOOLEAN NOT NULL DEFAULT FALSE
);

-- Start device primary keys at 1000
ALTER SEQUENCE v2.devices_id_seq RESTART WITH 1000;

--
-- Index devices table by name, replaced_by
--
CREATE UNIQUE INDEX devices_cookie_index ON v2.devices (cookie) WHERE cookie IS NOT NULL;
CREATE UNIQUE INDEX devices_hwid_active_index ON v2.devices (hwid) WHERE retired = FALSE;
CREATE UNIQUE INDEX devices_fpr_index ON v2.devices (fpr) WHERE retired = FALSE;
CREATE INDEX devices_hwid_index ON v2.devices (hwid);

--
-- Devices table update trigger. Makes sure the immutable fields aren't
-- changed and some fields can only be changed in certain 'directions'.
--
CREATE FUNCTION v2.devices_ensure_row_consistency() RETURNS TRIGGER AS $$
BEGIN
  --
  -- Immutablility checks
  --

  -- id is immutable
  IF OLD.id IS DISTINCT FROM NEW.id THEN
    RAISE 'Tried to modify immutable value "id"';
  END IF;
  -- cookie is immutable
  IF OLD.cookie IS DISTINCT FROM NEW.cookie THEN
    RAISE 'Tried to modify immutable value "cookie"';
  END IF;
  -- fpr is immutable
  IF OLD.hwid IS DISTINCT FROM NEW.hwid THEN
    RAISE 'Tried to modify immutable value "hwid"';
  END IF;
  -- fpr_ts is immutable
  IF OLD.fpr IS DISTINCT FROM NEW.fpr THEN
    RAISE 'Tried to modify immutable value "fpr"';
  END IF;
  -- organization_id is immutable
  IF OLD.organization_id IS DISTINCT FROM NEW.organization_id THEN
    RAISE 'Tried to modify immutable value "organization_id"';
  END IF;

  --
  -- Consistency checks
  --

  -- replaced_by can only be set
  IF OLD.replaced_by IS NOT NULL AND NEW.replaced_by IS NULL THEN
    RAISE 'Tried to modify immutable value "replaced_by"';
  END IF;

  -- retired cannot be set back to FALSE
  IF OLD.retired IS TRUE AND NEW.retired IS FALSE THEN
    RAISE 'Tried to set "retired" back to FALSE';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

--
-- Devices table update and insert trigger.
--
CREATE FUNCTION v2.devices_ensure_table_consistency() RETURNS TRIGGER AS $$
DECLARE
  have_dups integer;
BEGIN
  -- only one device per fpr can be unretired
  SELECT 1 INTO have_dups WHERE EXISTS (
    SELECT fpr, count(*)
      FROM v2.devices
      WHERE retired <> FALSE
      GROUP BY fpr
      HAVING count(*) > 1);

  IF have_dups IS NOT NULL THEN
    RAISE 'cannot have more than one active (non retired) per fingerprint';
  END IF;
 
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER devices_ensure_row_consistency
  BEFORE UPDATE ON v2.devices
  FOR EACH ROW
  WHEN (NEW.* IS DISTINCT FROM OLD.*)
  EXECUTE PROCEDURE v2.devices_ensure_row_consistency();

CREATE TABLE v2.keys (
  id BIGSERIAL PRIMARY KEY,

  public BYTEA NOT NULL CHECK (public <> ''),

  name TEXT NOT NULL CHECK (name <> ''),

  fpr BYTEA NOT NULL CHECK (fpr <> ''),

  credential TEXT NOT NULL CHECK (credential <> ''),

  device_id BIGINT REFERENCES v2.devices (id)
);

ALTER SEQUENCE v2.keys_id_seq RESTART WITH 1000;

CREATE INDEX keys_device_id ON v2.keys (device_id);
CREATE UNIQUE INDEX keys_device_id_fpr ON v2.keys (device_id,fpr);

GRANT SELECT, UPDATE, INSERT ON TABLE v2.devices TO apisrv;
GRANT SELECT, INSERT ON TABLE v2.keys TO apisrv;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA v2 TO apisrv;
