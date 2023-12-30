import * as IssuesV1 from 'generated/issuesv1';

export const examples: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml[] = [
  {
    "id": "csme/no-update",
    "aspect": "firmware",
    "incident": true,
    "args": {
      "components": [
        {
          "before": "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7",
          "after": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "name": "Intel RBE Manifest",
          "version": "17.2.2.1234"
        }
      ]
    }
  },
  {
    "id": "csme/downgrade",
    "incident": true,
    "aspect": "firmware",
    "args": {
      "combined": {
        "before": "17.2.20",
        "after": "17.0.1.1"
      },
      "components": [
        {
          "name": "RBE Manifest",
          "before": "17.2.20",
          "after": "17.0.1.1"
        }
      ]
    }
  },
  {
    "id": "uefi/option-rom-set",
    "incident": true,
    "aspect": "firmware",
    "args": {
      "devices": [
        {
          "before": "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7",
          "after": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "address": "0000:c1:03.1",
          "vendor": "Intel",
          "name": "Sunrise Point"
        }
      ]
    }
  },
  {
    "id": "uefi/boot-app-set",
    "incident": true,
    "aspect": "bootloader",
    "args": {
      "apps": [
        {
          "path": "C:\\\\EFI\\\\HDD1\\\\Windows.efi",
          "before": "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7",
          "after": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"
        }
      ]
    }
  },
  {
    "id": "uefi/ibb-no-update",
    "incident": true,
    "aspect": "firmware",
    "args": {
      "vendor": "AMI",
      "version": "1.2.3",
      "release_date": "22/11/22",
      "before": "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7",
      "after": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"
    }
  },
  {
    "id": "uefi/no-exit-boot-srv",
    "incident": false,
    "aspect": "firmware",
    "args": {
      "entered": true
    }
  },
  {
    "id": "uefi/gpt-changed",
    "incident": true,
    "aspect": "bootloader",
    "args": {
      "guid": "33FD7786-4E80-479F-9DA6-0A5BEC3F6C33",
      "before": "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7",
      "after": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
      "partitions": [
        {
          "guid": "33FD7786-4E80-479F-9DA6-0A5BEC3F6C33",
          "type": "6C53E76C-6A87-461D-B070-6E460A67CA2D",
          "name": null,
          "start": "a",
          "end": "fffffffff"
        }
      ]
    }
  },
  {
    "id": "uefi/secure-boot-keys",
    "incident": true,
    "aspect": "firmware",
    "args": {
      "pk": {
        "subject": "CN=SUSE Linux Enterprise Secure Boot CA,OU=Build Team,O=SUSE Linux Products GmbH,L=Nuremberg,C=DE,1.2.840.113549.1.9.1=#0c",
        "issuer": "CN=Microsoft Root Certificate Authority 2010,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US",
        "not_before": "1985-04-12T23:20:50.52Z",
        "not_after": "1985-04-12T23:20:50.52Z",
        "fpr": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"
      },
      "kek": [
        {
          "subject": "CN=SUSE Linux Enterprise Secure Boot CA,OU=Build Team,O=SUSE Linux Products GmbH,L=Nuremberg,C=DE,1.2.840.113549.1.9.1=#0c",
          "issuer": "CN=Microsoft Root Certificate Authority 2010,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US",
          "not_before": "1985-04-12T23:20:50.52Z",
          "not_after": "1985-04-12T23:20:50.52Z",
          "fpr": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"
        },
        {
          "subject": "CN=SUSE Linux Enterprise Secure Boot CA,OU=Build Team,O=SUSE Linux Products GmbH,L=Nuremberg,C=DE,1.2.840.113549.1.9.1=#0c",
          "issuer": "CN=Microsoft Root Certificate Authority 2010,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US",
          "not_before": "1985-04-12T23:20:50.52Z",
          "not_after": "1985-04-12T23:20:50.52Z",
          "fpr": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"
        },
        {
          "subject": "CN=SUSE Linux Enterprise Secure Boot CA,OU=Build Team,O=SUSE Linux Products GmbH,L=Nuremberg,C=DE,1.2.840.113549.1.9.1=#0c",
          "issuer": "CN=Microsoft Root Certificate Authority 2010,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US",
          "not_before": "1985-04-12T23:20:50.52Z",
          "not_after": "1985-04-12T23:20:50.52Z",
          "fpr": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"
        }
      ]
    }
  },
  {
    "id": "uefi/secure-boot-variables",
    "incident": true,
    "aspect": "configuration",
    "args": {
      "secure_boot": "0",
      "setup_mode": "0",
      "audit_mode": "",
      "deployed_mode": "1"
    }
  },
  {
    "id": "uefi/secure-boot-dbx",
    "incident": true,
    "aspect": "configuration",
    "args": {
      "fprs": [
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"
      ]
    }
  },
  {
    "id": "uefi/official-dbx",
    "incident": false,
    "aspect": "configuration",
    "args": {
      "fprs": [
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"
      ]
    }
  },
  {
    "id": "uefi/boot-failure",
    "incident": true,
    "aspect": "firmware",
    "args": {
      "pcr0": "00000001",
      "pcr1": "00000000",
      "pcr2": "00000000",
      "pcr3": "00000000",
      "pcr4": "00000000",
      "pcr5": "00000000",
      "pcr6": "00000000",
      "pcr7": "00000000"
    }
  },
  {
    "id": "uefi/boot-order",
    "incident": true,
    "aspect": "configuration",
    "args": {
      "variables": [
        {
          "name": "C:\\\\Boot\\\\Boot1.efi",
          "before": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "after": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff"
        },
        {
          "name": "C:\\\\Boot\\\\Boot2.efi",
          "before": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "after": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff"
        },
        {
          "name": "C:\\\\Boot\\\\Boot3.efi",
          "before": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "after": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff"
        }
      ]
    }
  },
  {
    "id": "tpm/endorsement-cert-unverified",
    "incident": true,
    "aspect": "supply-chain",
    "args": {
      "error": "san-invalid",
      "vendor": "IFX",
      "ek_issuer": "Infineon AG"
    }
  },
  {
    "id": "tpm/endorsement-cert-unverified",
    "incident": true,
    "aspect": "supply-chain",
    "args": {
      "error": "san-mismatch",
      "vendor": "IFX",
      "ek_issuer": "Infineon AG",
      "ek_vendor": "Infineon AG",
      "ek_version": "2.00 rev. 102"
    }
  },
  {
    "id": "tpm/endorsement-cert-unverified",
    "incident": true,
    "aspect": "supply-chain",
    "args": {
      "error": "no-eku",
      "vendor": "IFX",
      "ek_issuer": "Infineon AG",
      "ek_vendor": "Infineon AG",
      "ek_version": "2.00 rev. 102"
    }
  },
  {
    "id": "tpm/endorsement-cert-unverified",
    "incident": true,
    "aspect": "supply-chain",
    "args": {
      "error": "invalid-certificate",
      "vendor": "IFX",
      "ek_issuer": "Infineon AG",
      "ek_vendor": "Infineon AG",
      "ek_version": "2.00 rev. 102"
    }
  },
  {
    "id": "tpm/no-eventlog",
    "incident": true,
    "aspect": "firmware"
  },
  {
    "id": "tpm/dummy",
    "incident": true,
    "aspect": "supply-chain"
  },
  {
    "id": "grub/boot-changed",
    "incident": true,
    "aspect": "bootloader",
    "args": {
      "before": {
        "kernel": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "kernel_path": "(hd0,0)/boot/vmlinux-5.12.0-ubuntu12",
        "command_line": [
          "root=/dev/sda1",
          "acpioff",
          "swap=/dev/sdb"
        ],
        "initrd": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
        "initrd_path": "(hd0,0)/boot/initrd-5.12.0-ubuntu12"
      },
      "after": {
        "kernel": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff",
        "kernel_path": "(hd0,0)/boot/vmlinux-5.13.0-ubuntu12",
        "command_line": [
          "root=/dev/sda1",
          "acpioff",
          "swap=/dev/sdb"
        ],
        "initrd": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff",
        "initrd_path": "(hd0,0)/boot/initrd-5.13.0-ubuntu12"
      }
    }
  },
  {
    "id": "eset/disabled",
    "incident": true,
    "aspect": "endpoint-protection"
  },
  {
    "id": "eset/not-started",
    "incident": true,
    "aspect": "endpoint-protection",
    "args": {
      "components": [
        {
          "path": "/opt/eset/EAgent",
          "started": false
        },
        {
          "path": "/opt/eset/libexec/x86_64-linux-gnu/blah-blub",
          "started": true
        },
        {
          "path": "/opt/eset/blag",
          "started": false
        },
        {
          "path": "/opt/eset/lib/modules/eset_rtp.ko",
          "started": true
        },
        {
          "path": "/opt/eset/bin/blub",
          "started": false
        }
      ]
    }
  },
  {
    "id": "eset/excluded-set",
    "aspect": "endpoint-protection",
    "incident": true,
    "args": {
      "files": [
        "/opt/eset/EAgent",
        "/opt/eset/EAgent",
        "/opt/eset/EAgent",
        "/opt/eset/EAgent",
        "/opt/eset/EAgent"
      ],
      "processes": [
        "/bin/bash",
        "/bin/bash",
        "/bin/bash",
        "/bin/bash",
        "/bin/bash"
      ]
    }
  },
  {
    "id": "eset/manipulated",
    "incident": true,
    "aspect": "endpoint-protection",
    "args": {
      "components": [
        {
          "path": "/opt/eset/EAgent",
          "before": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "after": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff"
        }
      ]
    }
  },
  {
    "id": "windows/boot-config",
    "incident": false,
    "aspect": "operating-system",
    "args": {
      "boot_debugging": false,
      "kernel_debugging": false,
      "dep_disabled": false,
      "code_integrity_disabled": false,
      "test_signing": false
    }
  },
  {
    "id": "windows/boot-log",
    "incident": true,
    "aspect": "operating-system",
    "args": {
      "error": "missing-trust-point",
      "log": 2
    }
  },
  {
    "id": "windows/boot-log",
    "incident": true,
    "aspect": "operating-system",
    "args": {
      "error": "wrong-format",
      "log": 2
    }
  },
  {
    "id": "windows/boot-log",
    "incident": true,
    "aspect": "operating-system",
    "args": {
      "error": "wrong-signature",
      "log": 2
    }
  },
  {
    "id": "windows/boot-log",
    "incident": true,
    "aspect": "operating-system",
    "args": {
      "error": "wrong-quote",
      "log": 2
    }
  },
  {
    "id": "windows/boot-counter-replay",
    "incident": true,
    "aspect": "operating-system",
    "args": {
      "latest": "8",
      "received": "2"
    }
  },
  {
    "id": "ima/invalid-log",
    "incident": true,
    "aspect": "endpoint-protection",
    "args": {
      "pcr": [
        {
          "number": "13",
          "computed": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "quoted": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff"
        },
        {
          "number": "12",
          "computed": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "quoted": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff"
        }
      ]
    }
  },
  {
    "id": "ima/boot-aggregate",
    "incident": true,
    "aspect": "endpoint-protection",
    "args": {
      "computed": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
      "logged": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff"
    }
  },
  {
    "id": "ima/runtime-measurements",
    "incident": true,
    "aspect": "endpoint-protection",
    "args": {
      "files": [
        {
          "path": "/opt/eset/EAgent",
          "before": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "after": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff"
        }
      ]
    }
  },
  {
    "id": "tsc/pcr-values",
    "incident": true,
    "aspect": "supply-chain",
    "args": {
      "values": [
        {
          "tsc": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "quoted": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff",
          "number": "0"
        },
        {
          "tsc": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "quoted": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff",
          "number": "0"
        },
        {
          "tsc": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "quoted": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff",
          "number": "0"
        },
        {
          "tsc": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
          "quoted": "7058299627365fc7a3dd7840fd3d56f29306cd30c0f2c13cb500fe79617290ff",
          "number": "0"
        }
      ]
    }
  },
  {
    "id": "tsc/endorsement-certificate",
    "aspect": "supply-chain",
    "incident": true,
    "args": {
      "error": "xml-serial",
      "xml_serial": "026382998",
      "ek_serial": "026382998",
      "holder_serial": "7058299",
      "holder_issuer": "CN=SUSE Linux Enterprise Secure Boot CA,OU=Build Team,O=SUSE Linux Products GmbH,L=Nuremberg,C=DE,1.2.840.113549.1.9.1=#0c",
      "ek_issuer": "CN=Microsoft Root Certificate Authority 2010,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US"
    }
  },
  {
    "id": "tsc/endorsement-certificate",
    "aspect": "supply-chain",
    "incident": true,
    "args": {
      "error": "holder-issuer",
      "xml_serial": "026382998",
      "ek_serial": "026382998",
      "holder_serial": "7058299",
      "holder_issuer": "CN=SUSE Linux Enterprise Secure Boot CA,OU=Build Team,O=SUSE Linux Products GmbH,L=Nuremberg,C=DE,1.2.840.113549.1.9.1=#0c",
      "ek_issuer": "CN=Microsoft Root Certificate Authority 2010,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US"
    }
  },
  {
    "id": "tsc/endorsement-certificate",
    "incident": true,
    "aspect": "supply-chain",
    "args": {
      "error": "holder-serial",
      "xml_serial": "026382998",
      "ek_serial": "026382998",
      "holder_serial": "7058299",
      "holder_issuer": "CN=SUSE Linux Enterprise Secure Boot CA,OU=Build Team,O=SUSE Linux Products GmbH,L=Nuremberg,C=DE,1.2.840.113549.1.9.1=#0c",
      "ek_issuer": "CN=Microsoft Root Certificate Authority 2010,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US"
    }
  },
  {
    "id": "policy/endpoint-protection",
    "aspect": "endpoint-protection",
    "incident": true
  },
  {
    "id": "policy/intel-tsc",
    "aspect": "supply-chain",
    "incident": true
  },
  {
    "id": "fw/update",
    "incident": false,
    "aspect": "firmware",
    "args": {
      "updates": [
        {
          "name": "UEFI DBX",
          "current": "77",
          "next": "122"
        },
        {
          "name": "UEFI DBX",
          "current": "77",
          "next": "122"
        },
        {
          "name": "UEFI DBX",
          "current": "77",
          "next": "122"
        },
        {
          "name": "UEFI DBX",
          "current": "77",
          "next": "122"
        }
      ]
    }
  },
  {
    "id": "brly/2021-001-swsmi-len-65529",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/rkloader",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-025",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-037",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-035",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-032",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-023",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-029",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/intel-bssa-dft",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-050",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-051",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-013",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-022",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-024",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/lojax-secdxe",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/usbrt-intel-sa-00057",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-005",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-021",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/moonbounce-core-dxe",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-053",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-028",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-008",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-040",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-020",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-004",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-016",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/usbrt-cve-2017-5721",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-006",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-019",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-009",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-026",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-039",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-010",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-007",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-031",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/mosaicregressor",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-012",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-033",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-036",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/thinkpwn",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-014",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-018",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-038",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-004",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/especter",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-003",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-042",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/usbrt-swsmi-cve-2020-12301",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-043",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-028-rsbstuffing",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-034",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-030",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-041",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-017",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/usbrt-usbsmi-cve-2020-12301",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-027",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-015",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-011",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-045",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-009-1",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-029-1",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-010-1",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-011-1",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-008-1",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-001",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-027",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-014",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-011",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-013",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-016",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-047",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-015",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-009",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-010",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2022-012",
    "incident": false,
    "aspect": "firmware"
  },
  {
    "id": "brly/2021-046",
    "incident": false,
    "aspect": "firmware"
  }
];

