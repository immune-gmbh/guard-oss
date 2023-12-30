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
