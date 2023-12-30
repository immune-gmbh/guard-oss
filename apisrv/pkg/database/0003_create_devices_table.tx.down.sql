DROP INDEX v2.devices_cookie_index;
DROP INDEX v2.devices_hwid_active_index;
DROP INDEX v2.devices_fpr_index;
DROP INDEX v2.devices_hwid_index;
DROP INDEX v2.keys_device_id;
DROP INDEX v2.keys_device_id_fpr;

DROP TABLE v2.keys;
DROP TABLE v2.devices;

DROP FUNCTION v2.devices_ensure_row_consistency;
DROP FUNCTION v2.devices_ensure_table_consistency;
