CREATE VIEW v2.last_appraisals AS
SELECT DISTINCT ON (v2.keys.device_id) v2.appraisals.*,
    v2.keys.device_id AS device_id
FROM v2.keys
    LEFT JOIN v2.appraisals ON appraisals.id = (
        SELECT id
        FROM v2.appraisals
        WHERE appraisals.key_id = KEYS.id
        ORDER BY appraisals.appraised_at DESC
        LIMIT 1
    )
ORDER BY v2.keys.device_id,
    v2.keys.id DESC;


GRANT SELECT ON TABLE v2.last_appraisals TO apisrv;