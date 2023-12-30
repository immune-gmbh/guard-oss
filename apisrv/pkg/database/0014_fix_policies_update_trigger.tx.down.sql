drop trigger policies_ensure_consistency on v2.policies;
create trigger policies_ensure_consistency
  before update on v2.policies
  for each row
  when (NEW.* is distinct from OLD.*)
  execute procedure v2.policies_ensure_consistency();

drop function v2.policies_ensure_consistency_v2();
