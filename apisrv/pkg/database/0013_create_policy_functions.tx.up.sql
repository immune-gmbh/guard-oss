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
