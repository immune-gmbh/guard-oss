UPDATE v2.evidence AS e
SET device_id = (
        SELECT k.device_id
        FROM v2.keys AS k
        WHERE k.fpr = e.signed_by
    );


CREATE INDEX evidence_device_id_fkey ON v2.evidence USING btree (device_id);