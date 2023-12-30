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
