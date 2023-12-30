create index keys_devices on v2.keys (device_id, id);
create index evidence_signed_by_sorted on v2.evidence (signed_by, received_at desc);
create index devices_organization_sorted on v2.devices (organization_id, id desc);
create index key_appraised_at on v2.appraisals (key_id, appraised_at desc);
create index active_devices on v2.devices (organization_id) where retired = false;
