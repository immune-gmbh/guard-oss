INSERT INTO v2.organizations (id, external, devices, features, updated_at)
VALUES (100, 'ext-id-1', 100, ARRAY[]::v2.organizations_feature[], 'NOW');

-- TestCreate/NewForVulnerable
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (130, E'\\x0022000b305c1823253de4495e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc9d936433562b7e5a055fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (130, E'\\x0023000b00050072000000100018000b00030010002022e5777a0befc5b49e7bdf1d93599db7b1e86c170dbd6109c754e88b4b757bd6002052d04a37d1ff5f2be8558ddd0a2d10f1859dac5232ec2e9c82d8b0107e4bbcce', 'aik', E'\\x0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 130);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (130, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (130, 130);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (130, hstore('0=>00'), ARRAY[]::TEXT[], 130);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (130, 'NOW'::timestamp - interval '24 hours', '2022-10-10 00:00:00 UTC', '{"type":"verdict/1","result":false,"bootchain":false,"firmware":false,"configuration":false}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 130,130);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (131, 'NOW'::timestamp - interval '5 minutes', '2022-10-10 00:00:00 UTC', '{"type":"verdict/1","result":false,"bootchain":false,"firmware":false,"configuration":false}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}, "annotations":[{"id": "test-blah", "fatal": true, "path": "/test/blah"}]}', 130,130);

-- TestCreate/NewStayVulnerable
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (132, E'\\x0022000b305c1823253de4491e640ec0c327f8d2da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc6d936433562b7e5a056fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), TRUE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (132, E'\\x0023000b00050072000000100018000b00030010002022e5777a0befc5b49e7bdf1d93599db7b1e86c170dbd6109c754e88b4b757bd6002052d04a37d1ff5f2be8558ddd0a2d10f1859dac5232ec2e9c82d8b0107e4bbcce', 'aik', E'\\x0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 132);


INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (132, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (132, 132);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (132, hstore('0=>00'), ARRAY[]::TEXT[], 132);

-- TestCreate/RevokeExclusive
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by) VALUES
  (200, E'\\x0022000b309c1823254de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823257de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL),
  (201, E'\\x0022000b309c1823255de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823258de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #4', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES
  (200, 'Test Policy #1', NULL, NULL, '{1,0,22}'::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL),
  (201, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL),
  (202, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id) VALUES
  (200, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 200),
  (201, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 201),
  (202, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 202);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES
  (200, 200),
  (200, 201),
  (201, 202);


-- TestCreate/RevokeShared
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by) VALUES
  (210, E'\\x0022000b309c1823254de4490e641ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823257de4490e639ec0c327f8d1da14e18636ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL),
  (211, E'\\x0022000b309c1823255de4490e639ec0c327f8d1da15e18634ef9f6c3b57b0ec33c1234a', E'\\x0022000b307c1823258de4490e639ec0c327f8d1da14e18634ef7f9c3b57b0ec34c1234b', 'Test Device #4', FALSE, 100, NULL),
  (212, E'\\x0022000b309c1823255de4490e639ec0c328f8d1da15e18634ef7f6c3b57b1ec33c1234a', E'\\x0022000b307c1823258de4490e639ec0c327f8d1da14e18635ef7f6c3b57b0ec36c1234b', 'Test Device #5 (Shared)', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES
  (210, 'Test Policy #1', NULL, NULL, '{1,0,22}'::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL),
  (211, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL),
  (212, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL),
  (213, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id) VALUES
  (210, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 210),
  (211, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 211),
  (212, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 212),
  (213, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 213);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES
  (210, 210),
  (210, 211),
  (211, 212),
  (212, 212),
  (212, 213);


-- TestCreate/RevokeNone
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by) VALUES
  (220, E'\\x0022000b309c1823254de4490e641ec0c327f8d1da16e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823257de4490e639ec0c328f8d1da14e18636ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL);


-- TestCreate/RevokeRevokedAndExpired
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by) VALUES
  (230, E'\\x0022000b309c1823254de4490e639ec0c327f8d1da17e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823257de4490e642ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES
  (230, 'Test Policy #1', NULL, NULL, '{1,0,22}'::smallint[], ARRAY[]::TEXT[], TRUE, 100, NULL),
  (231, 'Test Policy #1', NULL, 'NOW'::timestamp - interval '24 hours', ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id) VALUES
  (230, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 230),
  (231, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 231);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES
  (230, 230),
  (230, 231);
