DROP INDEX v2.appraisals_device_appraised_at;


CREATE INDEX appraisals_key_created ON v2.appraisals (key_id, received_at);