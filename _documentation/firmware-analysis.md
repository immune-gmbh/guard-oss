Firmware Analysis
=================

[Individual checks](https://docs.google.com/spreadsheets/d/1DhDo99Za-e1rwkngsGkgthjjBnn2gB9eNgP8ZDPE83o/edit#gid=0)

Data Sources
------------

SPI flash image (16 MB mapped below 4 GB), TPM 2.0 event log, hostname, set of
physical interface MACs, Intel TXT public space, SMBIOS tables.

**UEFI variables**: PK, KEK, PKDefault, KEKDefault, db, dbx, dbt, dbr,
dbDefault, dbxDefault, dbtDefault, dbrDefault, AuditMode, SetupMode,
DeployedMode, SecureBoot.

**Machine specific registers**: 0xc0010010 (sev), 0x3a (sgx)

**CPUID leafs (EAX/ECX)**: 0x8000001f/0x0 (sev), 0x7/0x0 (sgx), 0x12/0x0 to 0x12/0x9 (sgx)

**AMD SEV commands**: PLATFORM\_STATUS, PDH\_CERT\_EXPORT, GET\_ID, PEK\_CSR

**Intel ME commands**: GEN GetFWVersion, HMRFPO Enable, HMRFPO GetStatus,
FWCAPS FeatureState, FWCAPS GetLocalUpdate

**Intel SPS commands**: GetMEBiosInterface, GetVendorLabel

**TPM 2.0 GetCapability (TPM\_PT\_)**: LEVEL, REVISION, DAY\_OF\_YEAR, YEAR,
MANUFACTURER, VENDOR\_STRING\_1-4, FIRMWARE\_VERSION\_1-2

**PCI config space**: 8086:0000.16 (me)
