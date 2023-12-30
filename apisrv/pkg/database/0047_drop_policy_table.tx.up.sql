DROP VIEW v2.last_policy_changes;


DROP TABLE v2.devices_policies;


DROP VIEW v2.last_device_changes;


ALTER TABLE v2.changes DROP COLUMN constraint_id;


ALTER TABLE v2.changes DROP COLUMN policy_id;


DROP TABLE v2.constraints;


DROP TABLE v2.policies;


DROP FUNCTION v2.new_policy_v1;


DROP FUNCTION v2.new_policy_v2;


DROP FUNCTION v2.policies_ensure_consistency;


DROP FUNCTION v2.policies_ensure_consistency_v2;


DROP FUNCTION v2.devices_policies_prevent_cross_organization CASCADE;


DROP FUNCTION v2.new_template_v1;


DROP FUNCTION v2.new_template_v2;


DROP FUNCTION v2.new_template_v3;


DROP FUNCTION v2.revoke_policies_v1;


DROP TYPE v2.new_policy_v1_return;


DROP FUNCTION v2.changes_prevent_cross_organization() CASCADE;