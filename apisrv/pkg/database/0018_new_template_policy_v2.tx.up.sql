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
