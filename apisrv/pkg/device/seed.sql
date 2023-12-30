INSERT INTO v2.organizations (id, external, devices, features, updated_at)
VALUES (100, 'ext-id-1', 100, ARRAY[]::v2.organizations_feature[], 'NOW');

INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (100, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', 'Test Device #1', FALSE, 100, NULL); 

INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (101, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234b', 'Test Device #2', hstore(''), TRUE, 100, 100); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (100, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 100);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (100, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (100, 100);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (101, 100);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (100, hstore(''), ARRAY[]::TEXT[], 100);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (100, 'NOW', '3022-10-10 00:00:00 UTC', '{"type":"verdict/1","result":true,"bootchain":true,"firmware":true,"configuration":true}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AXABYABBSTvC+LGXO5+/NFowggwv4IJR4FABPSo0kgbltfWqoqEIzo/Jr9IBiyymhed8X+RrMAAAAWAAAACwEAAAAAAAAAKgAWAARnyCQ9jxFLZERCpLXPjaNCtYMYlQAWAARTezDs+/844vwYBMHl12yx0wyLvQ==", "signature": "ABgACwAgtJbuh/YGOBH/7cfmw7EfHoU7ze+m/1znUCm+s3CJSswAIClMQuy7oLFYYg5D6IJSuaeR7WPYX74i4Uuej+5aE4K0", "algorithm": "20", "pcrs": {"0": "", "1": ""}, "firmware": {"uefi": [], "msrs": [], "cpuid": [], "me": [], "sps": [], "tpm2": [], "pci": [], "smbios":{"data":""}, "txt": {"data": ""}, "flash":{"data":""}, "event_log":{"data":""}, "os": {"hostname": "example.com", "name": "windows"}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 100,100);

INSERT INTO v2.changes (id, timestamp, device_id, key_id, organization_id, type)
VALUES (100, 'NOW'::timestamp, 100, 100,100, 'enroll');

INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (102, E'\\x0022000b305c1823252de4490e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b305c1823253de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (102, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823255de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 102);


INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (103, E'\\x0022000b305c1823253de4490e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b305c1823256de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (103, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823255de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234a', '{}', 103);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (103, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (103, 103);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (103, hstore(''), ARRAY[]::TEXT[], 103);

-- TestDeviceCRUD/ResurrectTrusted
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (104, E'\\x0022000b305c1823253de4491e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc6d936433562b7e5a055fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), TRUE, 100, NULL); 

-- Private key (MarshalECPrivateKey) MHcCAQEEIFl1D6VQ380l/DxPz9P17Caxh3kQaNjHjodPQBPVOYsZoAoGCCqGSM49AwEHoUQDQgAEIfJlZd+mBsmp8eJ2X/zQOMsvdpiI7E8/4QKkMo/5oX6zw9Q1ejO5iD6Dx2MyMZJPyuH2llzwwy/vFHbmACfkyw==
INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (104, E'\\x0023000dda874829002039e5a0380863aa65feb91dd49cbf5e2a53b1aa00c88c35080ecc0b6b09f9e74f0006008000420018000b00030021000c002021f26565dfa606c9a9f1e2765ffcd038cb2f769888ec4f3fe102a4328ff9a17e0020b3c3d4357a33b9883e83c7633231924fcae1f6965cf0c32fef1476e60027e4cb', 'aik', E'\\x0022000bf7f5c2fac339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 104);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (104, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (104, 104);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (104, hstore(''), ARRAY[]::TEXT[], 104);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (104, 'NOW'::timestamp - interval '24 hours', '3022-10-10 00:00:00 UTC', '{"type":"verdict/1","result":true,"bootchain":true,"firmware":true,"configuration":true}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 104,104);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (105, 'NOW'::timestamp - interval '5 minutes', 'NOW'::timestamp + interval '24 hours', '{"type":"verdict/1","result":true,"bootchain":true,"firmware":true,"configuration":true}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 104,104);


-- TestDeviceCRUD/ResurrectVuln
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (105, E'\\x0022000b305c1823253de4491e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc6d936433562b7e5a055fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), TRUE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (105, E'\\x0023000b00050072000000100018000b00030010002022e5777a0befc5b49e7bdf1d93599db7b1e86c170dbd6109c754e88b4b757bd6002052d04a37d1ff5f2be8558ddd0a2d10f1859dac5232ec2e9c82d8b0107e4bbcce', 'aik', E'\\x0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 105);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (105, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (105, 105);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (105, hstore('0=>00'), ARRAY[]::TEXT[], 105);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (107, 'NOW'::timestamp - interval '24 hours', '3022-10-10 00:00:00 UTC', '{"type":"verdict/1","result":false,"bootchain":false,"firmware":false,"configuration":false}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 105,105);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (106, 'NOW'::timestamp - interval '5 minutes', '3022-10-10 00:00:00 UTC', '{"type":"verdict/1","result":false,"bootchain":false,"firmware":false,"configuration":false}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 105,105);


-- TestDeviceCRUD/ResurrectUnseen
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (106, E'\\x0022000b305c1823253de4491e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc6d936433562b7e5a055fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), TRUE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (106, E'\\x0023000b00050072000000100018000b00030010002022e5777a0befc5b49e7bdf1d93599db7b1e86c170dbd6109c754e88b4b757bd6002052d04a37d1ff5f2be8558ddd0a2d10f1859dac5232ec2e9c82d8b0107e4bbcce', 'aik', E'\\x0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 106);


INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (106, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (106, 106);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (106, hstore('0=>00'), ARRAY[]::TEXT[], 106);


-- TestDeviceCRUD/ResurrectOutdated
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (107, E'\\x0022000b305c1823253de4491e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc6d936433562b7e5a055fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), TRUE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (107, E'\\x0023000b00050072000000100018000b00030010002022e5777a0befc5b49e7bdf1d93599db7b1e86c170dbd6109c754e88b4b757bd6002052d04a37d1ff5f2be8558ddd0a2d10f1859dac5232ec2e9c82d8b0107e4bbcce', 'aik', E'\\x0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 107);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (107, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (107, 107);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (107, hstore(''), ARRAY[]::TEXT[], 107);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (108, 'NOW'::timestamp - interval '72 hours', 'NOW'::timestamp - interval '48 hours', '{"type":"verdict/1","result":false,"bootchain":false,"firmware":false,"configuration":false}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 107,107);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (109, 'NOW'::timestamp - interval '48 hours', 'NOW'::timestamp - interval '24 hours', '{"type":"verdict/1","result":false,"bootchain":false,"firmware":false,"configuration":false}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 107,107);


-- ResurrectNotLast
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (110, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b14f5a595f211843ad59b46d4c58c838be5a5de5e32944f3db0aed2c546eb0552', 'Test Device #2', hstore(''), TRUE, 100, NULL);

INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (111, E'\\x0022000b305c1823253de4491e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc6d936433562b7e5a055fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), TRUE, 100, 110); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (111, E'\\x0023000b00050072000000100018000b00030010002022e5777a0befc5b49e7bdf1d93599db7b1e86c170dbd6109c754e88b4b757bd6002052d04a37d1ff5f2be8558ddd0a2d10f1859dac5232ec2e9c82d8b0107e4bbcce', 'aik', E'\\x0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 111);

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (110, E'\\x0023000b000400520020978a7bc3e0495b5712fefdbe0c102887e100dacd2d885f692cb607da00a11c1c00100018000b000300100020d9a07127a54bb71053054c3d72c1c373f5655c1bc12ab83ed7c2fff998bee20c0020986feb7e11a0d846c7220469d2577b671ad7056dae2fe3e5c58c1ff3695099cb', 'aik', E'\\x0022000b1c8e4fecc32d4c5d842ec3afb384cd4633cc22b59c91fde94339d3c2366d3dba', '{}', 110);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (110, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (110, 110);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (111, 110);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (110, hstore(''), ARRAY[]::TEXT[], 110);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (110, 'NOW'::timestamp - interval '5 minutes', 'NOW'::timestamp + interval '24 hours', '{"type":"verdict/1","result":true,"bootchain":true,"firmware":true,"configuration":true}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 111,110);


-- ResurrectMultipleKeys
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (120, E'\\x0022000b305c1823253de4491e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000bd178cb97d0377bc6d936433562b7e5a055fadf5bba48f6b7737ff3a2b53675dc', 'Test Device #1', TRUE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (120, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 120);

-- Private key (MarshalECPrivateKey) MHcCAQEEIFl1D6VQ380l/DxPz9P17Caxh3kQaNjHjodPQBPVOYsZoAoGCCqGSM49AwEHoUQDQgAEIfJlZd+mBsmp8eJ2X/zQOMsvdpiI7E8/4QKkMo/5oX6zw9Q1ejO5iD6Dx2MyMZJPyuH2llzwwy/vFHbmACfkyw==
INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (121, E'\\x0023000dda874829002039e5a0380863aa65feb91dd49cbf5e2a53b1aa00c88c35080ecc0b6b09f9e74f0006008000420018000b00030021000c002021f26565dfa606c9a9f1e2765ffcd038cb2f769888ec4f3fe102a4328ff9a17e0020b3c3d4357a33b9883e83c7633231924fcae1f6965cf0c32fef1476e60027e4cb', 'aik', E'\\x0022000bf7f5c2feb339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7', '{}', 120);

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (120, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (120, 120);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (120, hstore(''), ARRAY[]::TEXT[], 120);

INSERT INTO v2.appraisals (id, created_at, expires, verdict, evidence, report, key_id, constraint_id)
VALUES (120, 'NOW'::timestamp - interval '5 minutes', 'NOW'::timestamp + interval '24 hours', '{"type":"verdict/1","result":true,"bootchain":true,"firmware":true,"configuration":true}'::jsonb, '{"type": "evidence/1", "quote": "/1RDR4AYACIAC/f1wv6zOfnMQfL63ZHA2GugsMmgW9TkPMeHYWk8zcmnACDsHZ6rL6g0oEQdZL7CDfxNMqyIK4DJZW5x8F9Tb2KY5AAAAAAAAAAGAAAAAQAAAAABIBcGGQAWNjYAAAABAAsD/wEAACAtVWX7SD2OpFJaepIpZ30QOK00tuIsjVFS4df3uYF1lw==", "signature": "ABgACwAgHz8c5cIN5yk9Y8TTvMnpRoB6Ua5XU/OCZs5vDYfSiZgAINo5nf6QCqc6TvKCj/io7/VqiRDCNaeGsbvFed80AGzq", "algorithm": "11", "pcrs": {"0":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","2":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","3":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","4":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","5":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","6":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","7":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","8":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, "firmware": {"acpi":{},"cpuid":[],"event_log":{},"flash":{},"mac":{"addrs":null},"me":[],"memory":{},"msrs":[],"os":{"hostname":"example.com","name":"windows"},"pci":[],"sev":null,"smbios":{},"tpm2_nvram":null,"tpm2_properties":null,"txt":{},"uefi":[],"vtd":{}}, "cookie": ""}', '{"type": "report/1", "host": {"name": "Windows 10 RC2", "hostname": "example.com", "type": "windows", "cpu_vendor": "GenuineIntel"}, "smbios": {"manufacturer": "Lenovo", "product": "ThinkPad X230", "bios_vendor": "Lenovo", "bios_version": "1.0", "bios_release_date": "11/2/2011"}, "uefi":{"mode": "deployed", "secureboot": true, "platform_keys": [], "exchange_keys": [], "permitted_keys": [], "forbidden_keys": []}, "tpm": {"manufacturer": "IFX", "vendor_id": "Opti", "spec_version": "2", "eventlog": []}, "csme": {"variant": "consumer", "version": [1,2,3], "recovery_version": [1,2,3], "fitc_version": [1,2,3]}, "sgx": {"version": 2, "enabled": true, "flc": true, "kss": true, "enclave_size_64": 10000, "enclave_size_32": 1000, "epc": []}, "txt": {"ready": true}, "sev": {"enabled": true, "version": [1,18], "sme":true, "es": true, "vte": true, "snp": true, "vmpl": true, "guests": 3, "min_asid": 1}}', 121,120);


-- ResurrectNoAIK
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (121, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da15e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b14f5a595f211843ad59b46d4c58c838be5a5de5e32944f3db0aed2c546eb0552', 'Test Device #1', TRUE, 100, NULL); 

INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (121, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (121, 121);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (121, hstore(''), ARRAY[]::TEXT[], 121);


-- TestDeviceCRUD/ResurrectNoPolicies
INSERT INTO v2.devices (id, hwid, fpr, name, attributes, retired, organization_id, replaced_by)
VALUES (122, E'\\x0022000b305c1823253de4491e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', E'\\x0022000b14f5a595f211843ad59b46d4c58c838be5a5de5e32944f3db0aed2c546eb0552', 'Test Device #1', hstore('a=>a, b=>b, c=>c'), TRUE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (122, E'\\x0023000b000400520020978a7bc3e0495b5712fefdbe0c102887e100dacd2d885f692cb607da00a11c1c00100018000b000300100020d9a07127a54bb71053054c3d72c1c373f5655c1bc12ab83ed7c2fff998bee20c0020986feb7e11a0d846c7220469d2577b671ad7056dae2fe3e5c58c1ff3695099cb', 'aik', E'\\x0022000b1c8e4fecc32d4c5d842ec3afb384cd4633cc22b59c91fde94339d3c2366d3dba', '{}', 122);


-- TestDeviceCRUD/ListUpdateUntil

-- no pols
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (130, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1233a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1233a', 'Test Device #1', FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (130, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 130);

-- 1 active pols
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (131, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1238a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1238a', 'Test Device #1', FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (131, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 131);

-- 3 active (2 valid_until), 1 revoked, 1 template pol
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (132, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1237a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1237a', 'Test Device #1', FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (132, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 132);

-- 1 active, 1 template pol
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (133, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1236a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1236a', 'Test Device #1', FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (133, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 133);

-- 2 template, 2 active pol, valid_until, 1 revoked
INSERT INTO v2.devices (id, hwid, fpr, name, retired, organization_id, replaced_by)
VALUES (134, E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1235a', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1235a', 'Test Device #1', FALSE, 100, NULL); 

INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (134, E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251', 'aik', E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a', '{}', 134);


-- active 1
INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (130, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (130, hstore(''), ARRAY[]::TEXT[], 130);

-- active 2, valid_until
INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (131, 'Test Policy #1', NULL, 'NOW'::timestamp + interval '24 hours', ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (131, hstore(''), ARRAY[]::TEXT[], 131);

-- active 3, valid_until
INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (132, 'Test Policy #1', NULL, 'NOW'::timestamp + interval '48 hours', ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (132, hstore(''), ARRAY[]::TEXT[], 132);

-- template 1
INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (133, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

-- template 2
INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (134, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], FALSE, 100, NULL);

-- revoked 1
INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (135, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], TRUE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (135, hstore(''), ARRAY[]::TEXT[], 135);

-- revoked 2
INSERT INTO v2.policies (id, name, valid_from, valid_until, pcr_template, fw_template, revoked, organization_id, replaced_by)
VALUES (136, 'Test Policy #1', NULL, NULL, ARRAY[]::smallint[], ARRAY[]::TEXT[], TRUE, 100, NULL);

INSERT INTO v2.constraints (id, pcr_values, firmware, policy_id)
VALUES (136, hstore(''), ARRAY[]::TEXT[], 136);

-- 1 active pols
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (131, 130);

-- 3 active (2 valid_until), 1 revoked, 1 template pol
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (132, 130);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (132, 131);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (132, 132);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (132, 133);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (132, 135);

-- 1 active, 1 template pol
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (133, 131);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (133, 132);

-- 2 template, 2 active pol, valid_until, 1 revoked
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (134, 133);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (134, 134);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (134, 131);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (134, 132);
INSERT INTO v2.devices_policies (device_id, policy_id) VALUES (134, 136);
