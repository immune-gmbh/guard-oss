INSERT INTO v2.organizations (id, external, devices, features, updated_at)
VALUES (100, 'ext-id-1', 100, ARRAY[]::v2.organizations_feature[], 'NOW');

-- single constraints
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (100, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', 'Test Device #1', FALSE, 100, NULL);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (100, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (100, 100);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (100, hstore(''), ARRAY[]::TEXT[], 100);

-- multiple constraints
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (110, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (110, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);
INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (110, hstore(''), ARRAY[]::TEXT[], 110);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (111, hstore('0=>aa'), ARRAY[]::TEXT[], 110);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (110, 110);

-- single template policy (empty pcr_template)
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (120, E'\\x0022000b307c1823253de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b308c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (120, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (120, 120);

-- single template policy (overlapping pcr_template)
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (121, E'\\x0022000b307c1823253de4490e639ec0c328f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b308c1823252de4490e639ec1c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (121, 'Test Policy #1', NULL, NULL, '{0,33}'::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (121, 121);

-- single template policy (non-empty fw_template)
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (122, E'\\x0022000b307c1823253de4490e639ec0c327f8d1da16e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b308c1823252de4491e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (122, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (122, 122);

-- multiple policies
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (130, E'\\x0022000b307c1823252de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES
  (130, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL),
  (131, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL),
  (132, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES
  (130, hstore(''), ARRAY[]::TEXT[], 130),
  (131, hstore(''), ARRAY[]::TEXT[], 131),
  (132, hstore(''), ARRAY[]::TEXT[], 132);

INSERT INTO v2.devices_policies (device_id, policy_id)
VALUES (130, 130), (130, 131), (130, 132);

-- multiple template policies
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (150, E'\\x0022000b309c1823252de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823255de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES
  (150, 'Test Policy #1', NULL, NULL, '{1,0,22}'::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL),
  (151, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL),
  (152, 'Test Policy #1', NULL, NULL, '{1,0,22}'::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id)
VALUES (150, 150), (150, 151), (150, 152);

-- mixed template & active policies
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (160, E'\\x0022000b309c1823253de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823256de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES
  (160, 'Test Policy #1', NULL, NULL, '{1,0,22}'::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL),
  (161, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], '{"a","b"}'::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (160, hstore('"0"=>"000000", "1"=>"000000"'), ARRAY[]::TEXT[], 160);

INSERT INTO v2.devices_policies (device_id, policy_id)
VALUES (160, 160), (160, 161);

-- default policy
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (170, E'\\x0022000b309c1823254de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b307c1823256de4491e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 

-- single template policy
-- INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
-- VALUES (150, E'\\x0022000b307c1823253de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b308c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234b', 'Test Device #3', FALSE, 100, NULL); 
-- 
-- INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
-- VALUES (150, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);
-- 
-- INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (120, 120);


-- RepairBrokenDevice
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (230, E'\\x0022000b305c1823253de4495e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc9d936433562b7e5a055fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (230, E'\\x0023000b00050072000000100018000b00030010002022e5777a0befc5b49e7bdf1d93599db7b1e86c170dbd6109c754e88b4b757bd6002052d04a37d1ff5f2be8558ddd0a2d10f1859dac5232ec2e9c82d8b0107e4bbcce', 'aik', E'\\x0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 230);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (230, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (230, 230);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (230, hstore('0=>00'), ARRAY[]::TEXT[], 230);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (230, 'NOW'::timestamp - interval '24 hours', '3022-10-10 00:00:00 UTC', '{"type":"verdict/1","result":false,"bootchain":false,"firmware":false,"configuration":false}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 230,230);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (231, 'NOW'::timestamp - interval '5 minutes', '3022-10-10 00:00:00 UTC', '{"type":"verdict/1","result":false,"bootchain":false,"firmware":false,"configuration":false}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}, "annotations":[{"id": "test-blah", "fatal": true, "path": "/test/blah"}]}', 230,230);


