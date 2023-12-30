--
-- Devices table update trigger. Makes sure the immutable fields aren't
-- changed and some fields can only be changed in certain 'directions'.
--
create function v2.devices_ensure_row_consistency_v2() returns trigger as $$
begin
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

  RETURN NEW;
end;
$$ language plpgsql;

create trigger devices_ensure_row_consistency_v2
  before update on v2.devices
  for each row
  when (NEW.* is distinct from OLD.*)
  execute procedure v2.devices_ensure_row_consistency_v2();

drop trigger devices_ensure_row_consistency ON v2.devices;
