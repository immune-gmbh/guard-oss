-- We use the hstore data type for PCR lists
CREATE EXTENSION IF NOT EXISTS hstore;
-- Used the check PCR templates
CREATE EXTENSION IF NOT EXISTS intarray;

-- Policies table
CREATE TABLE v2.policies (
  id BIGSERIAL PRIMARY KEY,
  
  cookie TEXT NULL,

  -- Human readable name
  name TEXT NOT NULL CHECK(name <> ''), -- No empty name

  -- Validity range. Cannot be changed.
  valid_from TIMESTAMP NULL DEFAULT NULL,
  valid_until TIMESTAMP NULL DEFAULT NULL,

  CHECK (valid_until IS NULL OR valid_from IS NULL OR valid_from <= valid_until),
  
  -- List of PCRs. Cannot be changed.
  pcr_template smallint ARRAY NOT NULL 
    -- Only unsigned ints
    CHECK (pcr_template = ARRAY[]::smallint[] OR (sort(pcr_template)::smallint[])[0] >= 0)
    DEFAULT ARRAY[]::smallint[],

  fw_template TEXT ARRAY NOT NULL
    CHECK (NOT fw_template && ARRAY[''])
    DEFAULT ARRAY[]::text[],

  -- If a new policy was created that this one replaced. For example the user
  -- changes the validity period. Cannot be changed.
  replaced_by BIGINT REFERENCES v2.policies (id) NULL
    CHECK (id IS DISTINCT FROM replaced_by)
    DEFAULT NULL,

  -- Owner of this policies. Cannot be changed.
  organization_id BIGINT REFERENCES v2.organizations (id) NOT NULL,

  -- Whether the policy was revoked before period ended.
  revoked BOOLEAN NOT NULL DEFAULT FALSE
);

-- Start policy primary keys at 1000
ALTER SEQUENCE v2.policies_id_seq RESTART WITH 1000;


--
-- Index policies table by name, replaced_by
--
CREATE INDEX policies_organization_id_index ON v2.policies (organization_id);
CREATE INDEX policies_replaced_by_index ON v2.policies (replaced_by);
--
-- Policies table update trigger. Makes sure the immutable fields aren't
-- changed and some fields can only be changed in certain 'directions'.
--
CREATE FUNCTION v2.policies_ensure_consistency() RETURNS TRIGGER AS $$
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
   -- valid_from is immutable
  IF OLD.valid_from IS DISTINCT FROM NEW.valid_from THEN
    RAISE 'Tried to modify immutable value "valid_from"';
  END IF;
   -- valid_until is immutable
   IF OLD.valid_until IS DISTINCT FROM NEW.valid_until THEN
    RAISE 'Tried to modify immutable value "valid_until"';
  END IF;
  -- pcr_template is immutable
  IF OLD.pcr_template IS DISTINCT FROM NEW.pcr_template THEN
    RAISE 'Tried to modify immutable value "pcr_template"';
  END IF;
  -- fw_template is immutable
  IF OLD.fw_template IS DISTINCT FROM NEW.fw_template THEN
    RAISE 'Tried to modify immutable value "fw_template"';
  END IF;
   -- organization_id is immutable
  IF OLD.organization_id IS DISTINCT FROM NEW.organization_id THEN
    RAISE 'Tried to modify immutable value "organization_id"';
  END IF;
  -- replaces_id is immutable
  IF OLD.replaced_by IS DISTINCT FROM NEW.replaced_by THEN
    RAISE 'Tried to modify immutable value "replaced_by"';
  END IF;

  --
  -- Consistency checks
  --

  -- revoked cannot be set back to FALSE
  IF OLD.revoked IS TRUE AND NEW.revoked IS FALSE THEN
    RAISE 'Tried to set "revoked" back to FALSE';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER policies_ensure_consistency
  BEFORE UPDATE ON v2.policies
  FOR EACH ROW
  WHEN (NEW.* IS DISTINCT FROM OLD.*)
  EXECUTE PROCEDURE v2.policies_ensure_consistency();

CREATE TABLE v2.constraints (
  id BIGSERIAL PRIMARY KEY,

  pcr_values HSTORE NOT NULL
    CHECK (NOT %%pcr_values @> ARRAY['']), -- No empty strings

  firmware TEXT ARRAY NOT NULL,

  policy_id BIGINT REFERENCES v2.policies (id) NOT NULL
);

-- Start constraint primary keys at 1000
ALTER SEQUENCE v2.constraints_id_seq RESTART WITH 1000;

GRANT SELECT, UPDATE, INSERT ON TABLE v2.policies TO apisrv;
GRANT SELECT, INSERT ON TABLE v2.constraints TO apisrv;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA v2 TO apisrv;
