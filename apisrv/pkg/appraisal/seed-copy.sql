INSERT INTO v2.organizations (id, external, devices, features, updated_at)
VALUES (100, 'ext-id-1', 100, ARRAY[]::v2.organizations_feature[], 'NOW');

INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (100, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', 'Test Device #1', FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (100, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 100);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (100, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (100, 100);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (100, hstore(''), ARRAY[]::TEXT[], 100);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (100, 'NOW', timestamp 'NOW' + interval '24 hours', '{"type":"verdict/1","result":true,"bootchain":true,"firmware":true,"configuration":true}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AXABYABBSTvC+LGXO5+/NFowggwv4IJR4FABPSo0kgbltfWqoqEIzo/Jr9IBiyymhed8X+RrMAAAAWAAAACwEAAAAAAAAAKgAWAARnyCQ9jxFLZERCpLXPjaNCtYMYlQAWAARTezDs+/844vwYBMHl12yx0wyLvQ==", "signature": "ABgACwAgtJbuh/YGOBH/7cfmw7EfHoU7ze+m/1znUCm+s3CJSswAIClMQuy7oLFYYg5D6IJSuaeR7WPYX74i4Uuej+5aE4K0", "algorithm": "20", "pcrs": {"0": "", "1": ""}, "firmware": {"uefi": [], "msrs": [], "cpuid": [], "me": [], "sps": [], "tpm2": [], "pci": [], "smbios":{"data":""}, "txt": {"data": ""}, "flash":{"data":""}, "event_log":{"data":""}, "os": {"hostname": "example.com", "name": "windows"}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 100,100);

-- single constraint
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (110, E'\\x0022000b305c1823252de4490e639ec0c327f8d2da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b305c1823253de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', 'Test Device #1', FALSE, 100, NULL);

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (110, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c328f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 110);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (110, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (110, 110);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (110, hstore('"0"=>"00000000"'), ARRAY[]::TEXT[], 110);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (111, 'Test Policy #1', NULL, NULL, ARRAY[0,1,2]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (110, 111);


