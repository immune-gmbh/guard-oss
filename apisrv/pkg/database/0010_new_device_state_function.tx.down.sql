DROP FUNCTION v2.device_state;
CREATE FUNCTION v2.device_state(retired boolean, has_unretired boolean, aik_id BIGINT, verdict boolean, expires TIMESTAMP, now TIMESTAMP) RETURNS text AS $$
BEGIN
  CASE
    WHEN retired = TRUE AND NOT has_unretired THEN
      RETURN 'resurrectable';
    WHEN retired = TRUE AND has_unretired THEN
      RETURN 'retired';
    WHEN aik_id IS NULL THEN 
      RETURN 'new';
    WHEN verdict IS NULL THEN 
      RETURN 'unseen';
    WHEN expires <= now THEN
      RETURN 'outdated';
    WHEN verdict = TRUE THEN
      RETURN 'trusted';
    ELSE
      RETURN 'vulnerable';
  END CASE;
END;
$$ LANGUAGE plpgsql;

