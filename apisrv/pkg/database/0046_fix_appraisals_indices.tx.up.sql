CREATE INDEX appraisals_device_appraised_at ON v2.appraisals USING btree (device_id, appraised_at DESC) INCLUDE (id);


DROP INDEX v2.appraisals_key_created;