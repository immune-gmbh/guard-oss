--
-- WARNING: these migrations only restore enough for the downmigration not to fail, but not all triggers and functions are restored
--          and the policy functionality will be defunct; the apisrv code is not using any of this anyway, so it will not pose a problem
--

--- migration 0004

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


--- migration 0005

CREATE TABLE v2.devices_policies (
  device_id BIGINT REFERENCES v2.devices (id),
  policy_id BIGINT REFERENCES v2.policies (id)
);

CREATE UNIQUE INDEX devices_policies_index
  ON v2.devices_policies(device_id, policy_id);

CREATE FUNCTION v2.devices_policies_prevent_cross_organization() RETURNS TRIGGER AS $$
DECLARE
  mismatch integer;
BEGIN
  SELECT 1 INTO mismatch WHERE EXISTS (
    SELECT 1
      FROM v2.devices_policies
      JOIN v2.devices ON devices.id = devices_policies.device_id
      JOIN v2.policies ON policies.id = devices_policies.policy_id
      WHERE devices.organization_id <> policies.organization_id);

  IF mismatch IS NOT NULL THEN
    RAISE 'cannot associate devices with policies of another organization';
  END IF;
 
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER devices_policies_prevent_cross_organization_trigger
  AFTER UPDATE OR INSERT ON v2.devices_policies
  EXECUTE PROCEDURE v2.devices_policies_prevent_cross_organization();

CREATE TRIGGER devices_policies_prevent_cross_organization_trigger
  AFTER UPDATE OR INSERT ON v2.devices
  EXECUTE PROCEDURE v2.devices_policies_prevent_cross_organization();

CREATE TRIGGER devices_policies_prevent_cross_organization_trigger
  AFTER UPDATE OR INSERT ON v2.policies
  EXECUTE PROCEDURE v2.devices_policies_prevent_cross_organization();

GRANT SELECT, INSERT, DELETE, UPDATE ON TABLE v2.devices_policies TO apisrv;

--- migration 0013

create type v2.new_policy_v1_return as (
  policy_id bigint,
  constraint_id bigint
);

-- Insert a new policy with the given values and associates it with v_devices.
create function v2.new_policy_v1(
  v_org_ext text,
  v_cookie text,
  v_name text,
  v_valid_from timestamptz,
  v_valid_until timestamptz,
  v_pcrs hstore,
  v_firmware text[],
  v_now timestamptz,
  v_actor text,
  v_devices bigint[]
) returns v2.new_policy_v1_return as $$
declare
  org_id bigint;
  existing_policy v2.new_policy_v1_return;
  new_policy_id bigint;
  new_constraint_id bigint;
  new_policy v2.new_policy_v1_return;
begin
  -- only v_valid_from and v_valid_until can be null
  case 
    when v_org_ext is null then
      raise exception 'v_org_ext cannot be NULL';
    when v_name is null then
      raise exception 'v_name cannot be NULL';
    when v_pcrs is null then
      raise exception 'v_pcrs cannot be NULL';
    when v_firmware is null then
      raise exception 'v_firmware cannot be NULL';
    when v_now is null then
      raise exception 'v_now cannot be NULL';
    when v_actor is null then
      raise exception 'v_actor cannot be NULL';
    when v_devices is null then 
      raise exception 'v_devices cannot be NULL';
    else
  end case;

  -- organization id
  select id into strict org_id
    from v2.organizations
    where v2.organizations.external = v_org_ext
    limit 1;

  -- short circut if we've finished this request already
  select 
    v2.policies.id as policy_id,
    v2.constraints.id as constraint_id
  into existing_policy
    from v2.policies
    left join v2.constraints on v2.constraints.policy_id = v2.policies.id
    where v2.policies.cookie = v_cookie
      and v_cookie is not null
      and v2.policies.organization_id = org_id;
  if existing_policy is not null then
    return existing_policy;
  end if;

  -- insert a new policy with no template
  insert into v2.policies (
    id,
    cookie,
    name,
    valid_from,
    valid_until,
    organization_id,
    revoked)
  values (
    default,
    v_cookie,
    v_name,
    v_valid_from,
    v_valid_until,
    org_id,
    false)
  returning id into strict new_policy_id;

  -- insert constraint set for above policy
  insert into v2.constraints (
    id,
    pcr_values,
    firmware,
    policy_id)
  values (
    default,
    v_pcrs,
    v_firmware,
    new_policy_id)
  returning id into strict new_constraint_id;

  -- insert change for the new policy/constraint
  insert into v2.changes (
    type,
    policy_id,
    constraint_id,
    organization_id,
    timestamp,
    actor)
  values (
    'new',
    new_policy_id,
    new_constraint_id,
    org_id,
    v_now,
    v_actor);

  -- associate devices with new policy
  insert into v2.devices_policies (
    device_id,
    policy_id)
  select
    unnest as device_id,
    new_policy_id
  from
    unnest(v_devices);

  -- insert changes for the above association
  insert into v2.changes (
    type,
    policy_id,
    device_id,
    organization_id,
    timestamp,
    actor)
  select
    'associate',
    new_policy_id,
    unnest as device_id,
    org_id,
    v_now,
    v_actor
  from
    unnest(v_devices);

  -- return new policy/constraint pair
  select
    new_policy_id as policy_id,
    new_constraint_id as constraint_id
  into strict new_policy;
  return new_policy;
end;
$$ language plpgsql;

-- Insert a new policy template with the given values and associates it with v_devices.
create function v2.new_template_v1(
  v_org_ext text,
  v_cookie text,
  v_name text,
  v_valid_from timestamptz,
  v_valid_until timestamptz,
  v_pcrs int[],
  v_firmware text[],
  v_now timestamptz,
  v_actor text,
  v_devices bigint[]
) returns v2.new_policy_v1_return as $$
declare
  org_id bigint;
  existing_policy v2.new_policy_v1_return;
  new_policy_id bigint;
  new_policy v2.new_policy_v1_return;
begin
  -- only v_valid_from and v_valid_until can be null
  case 
    when v_org_ext is null then
      raise exception 'v_org_ext cannot be NULL';
    when v_name is null then
      raise exception 'v_name cannot be NULL';
    when v_pcrs is null then
      raise exception 'v_pcrs cannot be NULL';
    when v_firmware is null then
      raise exception 'v_firmware cannot be NULL';
    when v_now is null then
      raise exception 'v_now cannot be NULL';
    when v_actor is null then
      raise exception 'v_actor cannot be NULL';
    when v_devices is null then 
      raise exception 'v_devices cannot be NULL';
    else
  end case;

  -- organization id
  select id into strict org_id
    from v2.organizations
    where v2.organizations.external = v_org_ext
    limit 1;

  -- short circut if we've finished this request already
  select 
    v2.policies.id as policy_id,
    v2.constraints.id as constraint_id
  into existing_policy
    from v2.policies
    left join v2.constraints on v2.constraints.policy_id = v2.policies.id
    where v2.policies.cookie = v_cookie
      and v_cookie is not null
      and v2.policies.organization_id = org_id;
  if existing_policy is not null then
    return existing_policy;
  end if;

  -- insert a new policy with no template
  insert into v2.policies (
    id,
    cookie,
    name,
    valid_from,
    valid_until,
    pcr_template,
    fw_template,
    organization_id,
    revoked)
  values (
    default,
    v_cookie,
    v_name,
    v_valid_from,
    v_valid_until,
    v_pcrs,
    v_firmware,
    org_id,
    false)
  returning * into strict new_policy_id;

  -- insert change for the new policy/constraint
  insert into v2.changes (
    type,
    policy_id,
    organization_id,
    timestamp,
    actor)
  values (
    'new',
    new_policy_id,
    org_id,
    v_now,
    v_actor);

  -- associate devices with new policy
  insert into v2.devices_policies (
    device_id,
    policy_id)
  select
    unnest as device_id,
    new_policy_id
  from
    unnest(v_devices);

  -- insert changes for the above association
  insert into v2.changes (
    type,
    policy_id,
    device_id,
    organization_id,
    timestamp,
    actor)
  select
    'associate',
    new_policy_id,
    unnest as device_id,
    org_id,
    v_now,
    v_actor
  from
    unnest(v_devices);

  -- return new policy
  select
    new_policy_id as policy_id
  into strict new_policy;
  return new_policy;
end;
$$ language plpgsql;

-- revokes all active policies of v_devices, optionally adjusts their validity
-- period to stay active until v_valid_until. If the policies are shared with
-- devices not in v_devices policies are copied to isolate changes to the
-- devices in v_devices.
create function v2.revoke_policies_v1(
  v_devices bigint[],
  v_valid_until timestamptz,
  v_replaced_by bigint,
  v_org text,
  v_actor text,
  v_now timestamptz
) returns setof bigint as $$
declare
  org_id bigint;
begin
  -- organization id
  select id into strict org_id
    from v2.organizations
    where v2.organizations.external = v_org
    limit 1;

  -- select all policies of interest
  create temporary table selected_policies on commit drop as
  select distinct on (old_policy_id)
    v2.devices_policies.device_id,
    v2.devices_policies.policy_id as old_policy_id,
    nextval('v2.policies_id_seq') as new_policy_id
  from v2.devices_policies
  inner join v2.policies on v2.policies.id = v2.devices_policies.policy_id
  where v2.devices_policies.device_id = any(v_devices)
    and v2.policies.revoked = false 
    and (v2.policies.valid_from is null or v2.policies.valid_from <= v_now)
    and (v2.policies.valid_until is null or v2.policies.valid_until >= v_now)
  order by v2.devices_policies.policy_id desc;

  create temporary table exclusive_policies on commit drop as
  select distinct
    v2.devices_policies.policy_id as old_policy_id
  from v2.devices_policies
  inner join v2.policies on v2.policies.id = v2.devices_policies.policy_id
  where v2.devices_policies.device_id = any(v_devices)
    and v2.policies.revoked = false 
    and (v2.policies.valid_from is null or v2.policies.valid_from <= v_now)
    and (v2.policies.valid_until is null or v2.policies.valid_until >= v_now)
    and not exists (
    select 1 from v2.devices_policies as dp2
    where dp2.policy_id = v2.devices_policies.policy_id
      and not (dp2.device_id = any(v_devices)));

  if v_valid_until is not null then
    -- if a grace period is set create new policies
    insert into v2.policies (
      id,
      cookie,
      name,
      valid_from,
      valid_until,
      pcr_template,
      fw_template,
      replaced_by,
      organization_id,
      revoked)
    select 
      selected_policies.new_policy_id,
      v2.policies.cookie,
      v2.policies.name,
      v2.policies.valid_from,
      v_valid_until,
      v2.policies.pcr_template,
      v2.policies.fw_template,
      v2.policies.replaced_by,
      v2.policies.organization_id,
      v2.policies.revoked
    from selected_policies
    inner join v2.policies on selected_policies.old_policy_id = v2.policies.id;

    -- new constraints and changes for above
    with new_constraints as (
      insert into v2.constraints (
        id,
        pcr_values,
        firmware,
        policy_id)
      select 
        nextval('v2.constraints_id_seq'),
        v2.constraints.pcr_values,
        v2.constraints.firmware,
        selected_policies.new_policy_id
      from v2.constraints
      inner join selected_policies on selected_policies.old_policy_id = v2.constraints.policy_id
      returning id, policy_id

    )
    insert into v2.changes (
      type,
      policy_id,
      constraint_id,
      organization_id,
      actor,
      timestamp)
    select
     'new',
      new_constraints.policy_id,
      new_constraints.id,
      org_id,
      v_actor,
      v_now
    from new_constraints;

    delete from v2.devices_policies 
    where v2.devices_policies.device_id = any(v_devices)
      and exists (
        select 1 from selected_policies
        where selected_policies.old_policy_id = v2.devices_policies.policy_id);

    insert into v2.devices_policies (
      device_id,
      policy_id)
    select
      selected_policies.device_id,
      selected_policies.new_policy_id
    from selected_policies
    where selected_policies.device_id = any(v_devices);

    insert into v2.changes (
      type,
      device_id,
      policy_id,
      organization_id,
      actor,
      timestamp)
    select
      'associate',
      selected_policies.device_id,
      selected_policies.new_policy_id,
      org_id,
      v_actor,
      v_now
    from selected_policies
    where selected_policies.device_id = any(v_devices);
  end if;

  -- revoke non shared policies
  create temporary table revoked_policies on commit drop as
  with t as (
    update v2.policies
    set revoked = true, replaced_by = v_replaced_by
    where v2.policies.revoked = false
      and (v2.policies.valid_from is null or v2.policies.valid_from <= v_now)
      and (v2.policies.valid_until is null or v2.policies.valid_until >= v_now)
      and exists (
        select 1
          from exclusive_policies
          where exclusive_policies.old_policy_id = v2.policies.id)
    returning v2.policies.id
  )
  select t.* from t;

  insert into v2.changes (
    type,
    policy_id,
    organization_id,
    actor,
    timestamp)
  select
    'revoke',
    revoked_policies.id,
    org_id,
    v_actor,
    v_now
  from revoked_policies;

  return query select id from v2.devices
    where id = any(v_devices)
      and organization_id = org_id;
end;
$$ language plpgsql;

--- migration 0014

--
-- policies table update trigger. makes sure the immutable fields aren't
-- changed and some fields can only be changed in certain 'directions'.
--
create function v2.policies_ensure_consistency_v2() returns trigger as $$
begin
  --
  -- immutablility checks
  --

  -- id is immutable
  if OLD.id is distinct from NEW.id then
    raise 'tried to modify immutable value "id"';
  end if;
  -- cookie is immutable
  if OLD.cookie is distinct from NEW.cookie then
    raise 'tried to modify immutable value "cookie"';
  end if;
   -- valid_from is immutable
  if OLD.valid_from is distinct from NEW.valid_from then
    raise 'tried to modify immutable value "valid_from"';
  end if;
   -- valid_until is immutable
   if OLD.valid_until is distinct from NEW.valid_until then
    raise 'tried to modify immutable value "valid_until"';
  end if;
  -- pcr_template is immutable
  if OLD.pcr_template is distinct from NEW.pcr_template then
    raise 'tried to modify immutable value "pcr_template"';
  end if;
  -- fw_template is immutable
  if OLD.fw_template is distinct from NEW.fw_template then
    raise 'tried to modify immutable value "fw_template"';
  end if;
   -- organization_id is immutable
  if OLD.organization_id is distinct from NEW.organization_id then
    raise 'tried to modify immutable value "organization_id"';
  end if;

  --
  -- consistency checks
  --

  -- revoked cannot be set back to false
  if OLD.revoked is true and NEW.revoked is false then
    raise 'tried to set "revoked" back to false';
  end if;
  -- replaced_by is immutable
  if (OLD.replaced_by is not null) and (OLD.replaced_by is distinct from NEW.replaced_by) then
    raise 'tried to reset "replaced_by" after it has been set';
  end if;

  return NEW;
end;
$$ language plpgsql;

-- set new trigger
drop trigger policies_ensure_consistency on v2.policies;
create trigger policies_ensure_consistency
  before update on v2.policies
  for each row
  when (NEW.* is distinct from OLD.*)
  execute procedure v2.policies_ensure_consistency_v2();

--- migration 0018

-- Insert a new policy template with the given values and associates it with v_devices.
create function v2.new_template_v2(
  v_org_ext text,
  v_cookie text,
  v_name text,
  v_valid_from timestamptz,
  v_valid_until timestamptz,
  v_pcrs int[],
  v_firmware text[],
  v_now timestamptz,
  v_actor text,
  v_devices bigint[]
) returns v2.new_policy_v1_return as $$
declare
  org_id bigint;
  existing_policy v2.new_policy_v1_return;
  new_policy_id bigint;
  new_policy v2.new_policy_v1_return;
begin
  -- only v_valid_from and v_valid_until can be null
  case 
    when v_org_ext is null then
      raise exception 'v_org_ext cannot be NULL';
    when v_name is null then
      raise exception 'v_name cannot be NULL';
    when v_pcrs is null then
      raise exception 'v_pcrs cannot be NULL';
    when v_firmware is null then
      raise exception 'v_firmware cannot be NULL';
    when v_now is null then
      raise exception 'v_now cannot be NULL';
    when v_actor is null then
      raise exception 'v_actor cannot be NULL';
    when v_devices is null then 
      raise exception 'v_devices cannot be NULL';
    else
  end case;

  -- organization id
  select id into strict org_id
    from v2.organizations
    where v2.organizations.external = v_org_ext
    limit 1;

  -- short circut if we've finished this request already
  select 
    v2.policies.id as policy_id,
    v2.constraints.id as constraint_id
  into existing_policy
    from v2.policies
    left join v2.constraints on v2.constraints.policy_id = v2.policies.id
    where v2.policies.cookie = v_cookie
      and v_cookie is not null
      and v2.policies.organization_id = org_id;
  if existing_policy is not null then
    return existing_policy;
  end if;

  -- insert a new policy template
  insert into v2.policies (
    id,
    cookie,
    name,
    valid_from,
    valid_until,
    pcr_template,
    fw_template,
    organization_id,
    revoked)
  values (
    default,
    v_cookie,
    v_name,
    v_valid_from,
    v_valid_until,
    v_pcrs,
    v_firmware,
    org_id,
    false)
  returning id into strict new_policy_id;

  -- insert change for the new policy/constraint
  insert into v2.changes (
    type,
    policy_id,
    organization_id,
    timestamp,
    actor)
  values (
    'template',
    new_policy_id,
    org_id,
    v_now,
    v_actor);

  -- associate devices with new policy
  insert into v2.devices_policies (
    device_id,
    policy_id)
  select
    unnest as device_id,
    new_policy_id
  from
    unnest(v_devices);

  -- insert changes for the above association
  insert into v2.changes (
    type,
    policy_id,
    device_id,
    organization_id,
    timestamp,
    actor)
  select
    'associate',
    new_policy_id,
    unnest as device_id,
    org_id,
    v_now,
    v_actor
  from
    unnest(v_devices);

  -- return new policy
  select
    new_policy_id as policy_id
  into strict new_policy;
  return new_policy;
end;
$$ language plpgsql;

--- migration 0019

-- Insert a new policy template with the given values and associates it with v_devices.
create function v2.new_template_v3(
  v_org_ext text,
  v_cookie text,
  v_name text,
  v_valid_from timestamptz,
  v_valid_until timestamptz,
  v_pcrs int[],
  v_firmware text[],
  v_now timestamptz,
  v_actor text,
  v_devices bigint[],
  v_comment text
) returns v2.new_policy_v1_return as $$
declare
  org_id bigint;
  existing_policy v2.new_policy_v1_return;
  new_policy_id bigint;
  new_policy v2.new_policy_v1_return;
begin
  -- only v_valid_from, v_valid_until and v_comment can be null
  case 
    when v_org_ext is null then
      raise exception 'v_org_ext cannot be NULL';
    when v_name is null then
      raise exception 'v_name cannot be NULL';
    when v_pcrs is null then
      raise exception 'v_pcrs cannot be NULL';
    when v_firmware is null then
      raise exception 'v_firmware cannot be NULL';
    when v_now is null then
      raise exception 'v_now cannot be NULL';
    when v_actor is null then
      raise exception 'v_actor cannot be NULL';
    when v_devices is null then 
      raise exception 'v_devices cannot be NULL';
    else
  end case;

  -- organization id
  select id into strict org_id
    from v2.organizations
    where v2.organizations.external = v_org_ext
    limit 1;

  -- short circut if we've finished this request already
  select 
    v2.policies.id as policy_id,
    v2.constraints.id as constraint_id
  into existing_policy
    from v2.policies
    left join v2.constraints on v2.constraints.policy_id = v2.policies.id
    where v2.policies.cookie = v_cookie
      and v_cookie is not null
      and v2.policies.organization_id = org_id;
  if existing_policy is not null then
    return existing_policy;
  end if;

  -- insert a new policy template
  insert into v2.policies (
    id,
    cookie,
    name,
    valid_from,
    valid_until,
    pcr_template,
    fw_template,
    organization_id,
    revoked)
  values (
    default,
    v_cookie,
    v_name,
    v_valid_from,
    v_valid_until,
    v_pcrs,
    v_firmware,
    org_id,
    false)
  returning id into strict new_policy_id;

  -- insert change for the new policy/constraint
  insert into v2.changes (
    type,
    policy_id,
    organization_id,
    comment,
    timestamp,
    actor)
  values (
    'template',
    new_policy_id,
    org_id,
    v_comment,
    v_now,
    v_actor);

  -- associate devices with new policy
  insert into v2.devices_policies (
    device_id,
    policy_id)
  select
    unnest as device_id,
    new_policy_id
  from
    unnest(v_devices);

  -- insert changes for the above association
  insert into v2.changes (
    type,
    policy_id,
    device_id,
    organization_id,
    timestamp,
    actor)
  select
    'associate',
    new_policy_id,
    unnest as device_id,
    org_id,
    v_now,
    v_actor
  from
    unnest(v_devices);

  -- return new policy
  select
    new_policy_id as policy_id
  into strict new_policy;
  return new_policy;
end;
$$ language plpgsql;

--- migration 0020

-- Insert a new policy with the given values and associates it with v_devices.
create function v2.new_policy_v2(
  v_org_ext text,
  v_cookie text,
  v_name text,
  v_valid_from timestamptz,
  v_valid_until timestamptz,
  v_pcrs hstore,
  v_firmware text[],
  v_now timestamptz,
  v_actor text,
  v_devices bigint[],
  v_comment text
) returns v2.new_policy_v1_return as $$
declare
  org_id bigint;
  existing_policy v2.new_policy_v1_return;
  new_policy_id bigint;
  new_constraint_id bigint;
  new_policy v2.new_policy_v1_return;
begin
  -- only v_valid_from, v_valid_until and v_comment can be null
  case 
    when v_org_ext is null then
      raise exception 'v_org_ext cannot be NULL';
    when v_name is null then
      raise exception 'v_name cannot be NULL';
    when v_pcrs is null then
      raise exception 'v_pcrs cannot be NULL';
    when v_firmware is null then
      raise exception 'v_firmware cannot be NULL';
    when v_now is null then
      raise exception 'v_now cannot be NULL';
    when v_actor is null then
      raise exception 'v_actor cannot be NULL';
    when v_devices is null then 
      raise exception 'v_devices cannot be NULL';
    else
  end case;

  -- organization id
  select id into strict org_id
    from v2.organizations
    where v2.organizations.external = v_org_ext
    limit 1;

  -- short circut if we've finished this request already
  select 
    v2.policies.id as policy_id,
    v2.constraints.id as constraint_id
  into existing_policy
    from v2.policies
    left join v2.constraints on v2.constraints.policy_id = v2.policies.id
    where v2.policies.cookie = v_cookie
      and v_cookie is not null
      and v2.policies.organization_id = org_id;
  if existing_policy is not null then
    return existing_policy;
  end if;

  -- insert a new policy with no template
  insert into v2.policies (
    id,
    cookie,
    name,
    valid_from,
    valid_until,
    organization_id,
    revoked)
  values (
    default,
    v_cookie,
    v_name,
    v_valid_from,
    v_valid_until,
    org_id,
    false)
  returning id into strict new_policy_id;

  -- insert constraint set for above policy
  insert into v2.constraints (
    id,
    pcr_values,
    firmware,
    policy_id)
  values (
    default,
    v_pcrs,
    v_firmware,
    new_policy_id)
  returning id into strict new_constraint_id;

  -- insert change for the new policy/constraint
  insert into v2.changes (
    type,
    policy_id,
    constraint_id,
    organization_id,
    comment,
    timestamp,
    actor)
  values (
    'new',
    new_policy_id,
    new_constraint_id,
    org_id,
    v_comment,
    v_now,
    v_actor);

  -- associate devices with new policy
  insert into v2.devices_policies (
    device_id,
    policy_id)
  select
    unnest as device_id,
    new_policy_id
  from
    unnest(v_devices);

  -- insert changes for the above association
  insert into v2.changes (
    type,
    policy_id,
    device_id,
    organization_id,
    timestamp,
    actor)
  select
    'associate',
    new_policy_id,
    unnest as device_id,
    org_id,
    v_now,
    v_actor
  from
    unnest(v_devices);

  -- return new policy/constraint pair
  select
    new_policy_id as policy_id,
    new_constraint_id as constraint_id
  into strict new_policy;
  return new_policy;
end;
$$ language plpgsql;

--- migration 0008 partially

ALTER TABLE v2.changes ADD COLUMN constraint_id BIGINT;
ALTER TABLE v2.changes ADD COLUMN policy_id BIGINT;

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

GRANT SELECT ON TABLE v2.last_device_changes TO apisrv;
GRANT SELECT ON TABLE v2.last_policy_changes TO apisrv;