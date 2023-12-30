drop index v2.devices_hwid_active_index;
create unique index devices_hwid_active_index
  on v2.devices (hwid, organization_id) where retired = false;
