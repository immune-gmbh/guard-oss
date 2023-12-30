DROP FUNCTION v2.device_state;
CREATE FUNCTION v2.device_state(retired boolean, has_unretired boolean, aik_id BIGINT, verdict JSONB, expires TIMESTAMP, now TIMESTAMP) RETURNS text AS $$
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
    ELSE
      CASE (verdict->'result')::text
        WHEN 'true' THEN 
          RETURN 'trusted';
      WHEN '"trusted"' THEN
          RETURN 'trusted';
        ELSE
          RETURN 'vulnerable';
      END CASE;
  END CASE;
END;
$$ LANGUAGE plpgsql;
