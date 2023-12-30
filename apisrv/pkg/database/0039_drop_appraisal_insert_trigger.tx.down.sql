CREATE FUNCTION v2.appraisals_prevent_cross_organization() RETURNS TRIGGER AS $$
DECLARE mismatch integer;


BEGIN
SELECT 1 INTO mismatch
WHERE EXISTS (
    SELECT 1
    FROM v2.appraisals
      JOIN v2.devices ON devices.id = (
        SELECT device_id
        FROM v2.keys
        WHERE id = appraisals.key_id
      )
      JOIN v2.policies ON policies.id = (
        SELECT policy_id
        FROM v2.constraints
        WHERE id = appraisals.constraint_id
      )
    WHERE devices.organization_id <> policies.organization_id
  );


IF mismatch IS NOT NULL THEN RAISE 'cannot associate devices with policies of another organization';


END IF;


RETURN NEW;


END;


$$ LANGUAGE plpgsql;


CREATE TRIGGER appraisals_prevent_cross_organization_trigger
AFTER
UPDATE
  OR
INSERT ON v2.appraisals EXECUTE PROCEDURE v2.appraisals_prevent_cross_organization();