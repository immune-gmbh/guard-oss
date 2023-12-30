-- drop view and do not re-create b/c it conflicts with device_id column
DROP VIEW v2.devices_appraisals;


ALTER TABLE v2.appraisals
ADD COLUMN device_id BIGINT NOT NULL DEFAULT 0;