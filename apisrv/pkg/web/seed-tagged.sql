insert into
  v2.organizations (id, external, devices, features, updated_at)
values
  (100, 'ext-id-1', 100, array[]::v2.organizations_feature[], 'NOW');

insert into
  v2.devices (id, hwid, fpr, name, attributes, baseline, retired, organization_id)
values (
  100,
  E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
  E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
  'Test Device #1',
  hstore(''),
  '{"type": "baseline/3"}',
  false,
  100
), (
  101,
  E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234b',
  E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234b',
  'Test Device #2',
  hstore(''),
  '{"type": "baseline/3"}',
  false,
  100
), (
  102,
  E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234c',
  E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234c',
  'Test Device #3',
  hstore(''),
  '{"type": "baseline/3"}',
  false,
  100
);

insert into
  v2.tags (id, key, metadata, organization_id)
values (
  '0ujtsYcgvSTl8PAuAdqWYSMnLOv',
  'Test tag #1',
  '{"type": "tag/1"}',
  100
), (
  '0ujtsYcgvSTl8PAuAdqWYSMnLOw',
 'Test tag #2',
  '{"type": "tag/1"}',
  100
);

insert into v2.devices_tags (device_id, tag_id)
values
  (100, '0ujtsYcgvSTl8PAuAdqWYSMnLOv'),
  (100, '0ujtsYcgvSTl8PAuAdqWYSMnLOw'),
  (101, '0ujtsYcgvSTl8PAuAdqWYSMnLOw');
