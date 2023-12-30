CREATE INDEX IF NOT EXISTS devices_id_index
    ON v2.devices USING btree
    (id ASC NULLS LAST)
    TABLESPACE pg_default;