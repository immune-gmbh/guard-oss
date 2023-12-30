UPDATE v2.appraisals AS a
SET device_id = (
        SELECT k.device_id
        FROM v2.keys AS k
        WHERE k.id = a.key_id
    );


ALTER TABLE v2.appraisals
ADD CONSTRAINT appraisals_device_id_fkey FOREIGN KEY (device_id) REFERENCES v2.devices (id);


ALTER TABLE v2.appraisals
ALTER COLUMN device_id DROP DEFAULT;