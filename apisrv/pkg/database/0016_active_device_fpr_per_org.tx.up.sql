drop index v2.devices_fpr_index;
create unique index devices_fpr_index on v2.devices (fpr, organization_id) where retired = false;
