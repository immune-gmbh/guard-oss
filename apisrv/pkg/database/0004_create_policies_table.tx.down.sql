DROP INDEX v2.policies_organization_id_index;
DROP INDEX v2.policies_replaced_by_index;

DROP TABLE v2.constraints;
DROP TABLE v2.policies;

DROP FUNCTION v2.policies_ensure_consistency;
