import { Device, DeviceState, Policy, ChangeType } from './mock'

function newHwid(): string {
  return "0022000b" + newHash()
}

function newCookie(): string {
  return newHash()
}

function newHash(): string {
  let ret = ""
  while(ret.length < 64) {
    ret += Math.random().toString(16).substring(2)
  }
  return ret.substring(0, 64)
}

function newAttribs(): Record<string, string> {
  const validLocations = [
    "Berlin",
    "useast-a",
    "Raum 2/203",
    "Bochum",
    "Mars",
  ]
  const validOwner = [
    "Kai",
    "test@example.com",
    "Devops",
  ]

  const loc = Math.trunc(Math.random() * 100) % validLocations.length
  const owner = Math.trunc(Math.random() * 100) % validOwner.length

  return {
    loc: validLocations[loc],
    owner: validOwner[owner],
  }
}

function newPublicKey(): Record<string, unknown> {
  switch (Math.trunc(Math.random() ) * 10 % 2) {
    case 0:
      // RSA 2048 SHA256 PKCS#1
      return {
        type: "rsa2048",
        exponent: 0x10001,
        modulos: newHash()+newHash()+newHash()+newHash(),
      }

    case 1:
      // ECDSA NIST P256 SHA256
      return {
        type: "ec-p256",
        x: newHash()+newHash()+newHash()+newHash(),
        y: newHash()+newHash()+newHash()+newHash(),
      }
   }
}

function newX509(): Record<string, unknown> {
  return {
    name: 'Microsoft Windows Production PCA 2011',
    trusted: false,
    fingerprint: newHash(),
    issued_at: (Date.now() - (100 * 3600 * 3600)).toString(),
    issued_by: 'Microsoft Corporation',
    type: 'rsa2048-pkcs1-sha256',
  }
}

function newPCRSet(num: number): Record<number, string> {
  const ret = {}
  for (let i = 0; i < num; i += 1) {
    ret[i] = newHash()
  }
  return ret
}

function newAppraisal(verdict: boolean, policy: string, timestamp: number): Record<string, unknown> {
  const validPaths = [
    "/host/name",
    "/host/hostname",
    "/host/type",
    "/host/cpu_vendor",
    "/smbios/manufacturer",
    "/smbios/product",
    "/smbios/bios_release_date",
    "/smbios/bios_vendor",
    "/smbios/bios_version",
    "/uefi/platform_key",
    "/uefi/exchange_key",
    "/uefi/permitted_keys/certificates[0]",
    "/uefi/permitted_keys/certificates[1]",
    "/uefi/permitted_keys/hashes[0]",
    "/uefi/permitted_keys/hashes[1]",
    "/uefi/permitted_keys/keys[0]",
    "/uefi/permitted_keys/keys[1]",
    "/uefi/forbidden_keys/certificates[0]",
    "/uefi/forbidden_keys/hashes[0]",
    "/uefi/forbidden_keys/hashes[1]",
    "/uefi/mode",
    "/uefi/secureboot",
    "/uefi/option_roms[0]",
    "/tpm/manufacturer",
    "/tpm/spec_version",
    "/tpm/vendor_id",
    "/tpm/eventlog",
    "/me/features",
    "/me/variant",
    "/me/version",
    "/me/recovery_version",
    "/me/fitc_version",
    "/me/api",
    "/me/updatable",
    "/me/chipset_version",
    "/me/chip_id",
    "/me/manufacturer",
    "/me/size",
    "/me/signature",
    "/sgx/version",
    "/sgx/enabled",
    "/sgx/flc",
    "/sgx/kss",
    "/sgx/epc:[0]",
    "/txt/ready",
    "/sev/enabled",
    "/sev/version",
    "/sev/sme",
    "/sev/es",
    "/sev/vte",
    "/sev/snp",
    "/sev/vmpl",
    "/sev/guests",
    "/sev/min_asid",
    "/boot_guard/state",
    "/boot_guard/svn",
    "/boot_guard/bp_signing_key",
    "/boot_guard/ibb_signing_keys[0]",
  ]
 const annotations = []

  if (verdict == false) {
    const n = Math.random() * 5
    for (let i = 0; i < n; i += 1) {
      annotations.push({
        id: 'gen-error',
        path: validPaths[Math.trunc(Math.random() * 100) % validPaths.length],
        expected: "something",
      })
    }
  }

  return {
    id: newHash(),
    received: timestamp.toString(),
    expires: (timestamp + (Date.now() + 3600 * 3600)).toString(),
    evidence: {},
    verdict,
    policy,
    report: {
      type: "report/1",
      host: {
        name: "Ubuntu 18.04.5 LTS",
        hostname: "vision.9elements.com",
        type: "linux",
        cpu_vendor: "intel",
      },
      smbios: {
        manufacturer: "Micro-Star International Co., Ltd.",
        product: "MS-7A38",
        bios_release_date: "09/17/2018",
        bios_vendor: "American Megatrends Inc.",
        bios_version: "A.EF",
      },
      uefi: {
        platform_key: newX509(),
        exchange_key: newX509(),
        permitted_keys: {
          certificates: [newX509(), newX509()],
          hashes: [
            {
              type: "x509-sha256",
              digest: newHash(),
            }, {
              type: "sha256",
              digest: newHash(),
            },
          ],
          keys: [newPublicKey(), newPublicKey()]
        },
        forbidden_keys: {
          certificates: [newX509()],
          hashes: [
            {
              type: "x509-sha256",
              digest: newHash(),
            }, {
              type: "sha256",
              digest: newHash(),
            },
          ],
          keys:[],
        },
        mode: "depoyed",
        secureboot: true,
        option_roms: [
          {
            signed_by: newHash(),
            // XXX
          },
        ]
      },
      tpm: {
        manufacturer: "Infineon",
        vendor_id: "SLB9600",
        spec_version: "2.0",
        eventlog: [
          {
            pcr: 0,
            value: newHash(),
            algorithm: 0xb,
            note: "Initial boot block",
          }, {
            pcr: 1,
            value: newHash(),
            algorithm: 0xb,
            note: "Operating system",
          }
        ],
      },
      me: {
        features: ["ptt"],
        variant: "lightME",
        version: [15,0,1,0],
        recovery_version: [15,0,1,0],
        fitc_version: [15,0,1,0],
        api: [1,2],
        updatable: "password",
        chipset_version: 5,
        chip_id: "012345",
        manufacturer: "MSI",
        size: 10000000,
        signature: "Test",
      },
      sgx: {
        version: 2, // XXX SVN?
        enabled: true,
        flc: true,
        kss: true,
        enclave_size_32: 16400,
        enclave_size_64: 16400,
        epc: [
          {
            base: 0,
            size: 0xffff0000,
            cir_protection: true,
          }
        ]
      },
      txt: {
        ready: true,
      },
      sev: {
        enabled: true,
        version: [1,2,3],
        sme: true,
        es: true,
        vte: true,
        snp: true,
        vmpl: true,
        guests: 32,
        min_asid: 3,
      },
      boot_guard: {
        state: "enabled",
        svn: 1,
        bp_signing_key: newPublicKey(),
        ibb_signing_keys: [newPublicKey()],
      },
      annotations,
    }
  }
}

const actor1 = "1"


// Undo'd retire
const dev100 = {
  id: "100",
  cookie: newCookie(),
  name: "Test Device #1",
  state: "retired" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: [ "100" ],
  replaced_by: [ "101" ],
  replaces: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "retired" as ChangeType,
      comment: "Server removed from datacenter."
    }
  ],
  appraisals: [newAppraisal(true, "100", Date.now() - (3 * 3600 * 1000))]
}
const dev101 = {
  id: "101",
  cookie: newCookie(),
  name: "Test Device #1",
  state: "trusted" as DeviceState,
  attributes: newAttribs(),
  hwid: dev100.hwid,
  policies: [ "100" ],
  replaces: [ "100" ],
  replaced_by: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "resurrected" as ChangeType,
      comment: "Made a mistake."
    }
  ],
  appraisals: [newAppraisal(true, "100", Date.now() - (3 * 3600 * 1000))]
}
const pol100 = {
  id: "100",
  cookie: newCookie(),
  name: "Test Policy #100",
  devices: ["100", "101"],
  valid_from: (Date.now() - (10 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

const dev110 = {
  id: "110",
  cookie: newCookie(),
  name: "Unseen device w/o policy",
  state: "unseen" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: [],
  replaces: [],
  replaced_by: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: []
}

const dev120 = {
  id: "120",
  cookie: newCookie(),
  name: "Unseen device w/ policy",
  state: "unseen" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: [ "120" ],
  replaces: [],
  replaced_by: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: []
}
const pol120 = {
  id: "120",
  cookie: newCookie(),
  name: "Test Policy #120",
  devices: ["120"],
  valid_from: (Date.now() - (2 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

const dev130 = {
  id: "130",
  cookie: newCookie(),
  name: "Trusted device w/ policy",
  state: "trusted" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: [ "130" ],
  replaces: [],
  replaced_by: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: [
    newAppraisal(true, "130", Date.now() - (2 *3600 *3600)),
    newAppraisal(true, "130", Date.now() - (1.6 *3600 *3600)),
    newAppraisal(true, "130", Date.now() - (1.3 *3600 *3600)),
    newAppraisal(true, "130", Date.now() - (1 *3600 *3600)),
    newAppraisal(true, "130", Date.now() - (0.6 *3600 *3600)),
    newAppraisal(true, "130", Date.now() - (0.3 *3600 *3600))
  ]
}
const pol130 = {
  id: "130",
  cookie: newCookie(),
  name: "Test Policy #130",
  devices: ["130"],
  valid_from: (Date.now() - (2 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

const dev140 = {
  id: "140",
  cookie: newCookie(),
  name: "Vuln device w/ policy",
  state: "vulnerable" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: [ "140" ],
  replaces: [],
  replaced_by: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: [
    newAppraisal(false, "140", Date.now() - (2 *3600 *3600)),
    newAppraisal(true, "140", Date.now() - (1.6 *3600 *3600)),
    newAppraisal(false, "140", Date.now() - (1.3 *3600 *3600)),
    newAppraisal(true, "140", Date.now() - (1 *3600 *3600)),
    newAppraisal(true, "140", Date.now() - (0.6 *3600 *3600)),
    newAppraisal(true, "140", Date.now() - (0.3 *3600 *3600))
  ]
}
const pol140 = {
  id: "140",
  cookie: newCookie(),
  name: "Test Policy #140",
  devices: ["140"],
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

const dev150 = {
  id: "150",
  cookie: newCookie(),
  name: "Outdated device w/ policy",
  state: "outdated" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: [ "150" ],
  replaces: [],
  replaced_by: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: [
    newAppraisal(true, "140", Date.now() - (1.6 *3600 *3600)),
  ]
}
const pol150 = {
  id: "150",
  cookie: newCookie(),
  name: "Test Policy #150",
  devices: ["150"],
  valid_from: (Date.now() - (2 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

const dev160 = {
  id: "160",
  cookie: newCookie(),
  name: "Retired, can undo",
  state: "resurrectable" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: [ "160" ],
  replaced_by: [],
  replaces: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "retired" as ChangeType,
      comment: "Server removed from datacenter."
    }
  ],
  appraisals: [newAppraisal(true, "160", Date.now() - (3 * 3600 * 1000))]
}
const pol160 = {
  id: "160",
  cookie: newCookie(),
  name: "Test Policy #160",
  devices: ["160"],
  valid_from: (Date.now() - (10 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

const dev170 = {
  id: "170",
  cookie: newCookie(),
  name: "New device",
  state: "new" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: [],
  replaced_by: [],
  replaces: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "retired" as ChangeType,
      comment: "Server removed from datacenter."
    }
  ],
  appraisals: []
}

// vuln, schedruled updated
const dev180 = {
  id: "180",
  cookie: newCookie(),
  name: "Vulnerable, update schedruled",
  state: "vulnerable" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: ["180", "181"],
  replaced_by: [],
  replaces: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: [newAppraisal(false, "180", Date.now() - (3 * 3600 * 1000))]
}
const pol180 = {
  id: "180",
  cookie: newCookie(),
  name: "Old policy of Dev #180",
  devices: ["180"],
  valid_from: (Date.now() - (10 * 3600 * 3600)).toString(),
  valid_until: (Date.now() + (5 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}
const pol181 = {
  id: "181",
  cookie: newCookie(),
  name: "Schedruled policy of Dev #181",
  devices: ["180"],
  valid_from: (Date.now() + (5 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

const dev190 = {
  id: "190",
  cookie: newCookie(),
  name: "Trusted, update schedruled",
  state: "trusted" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: ["190", "191"],
  replaced_by: [],
  replaces: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: [newAppraisal(true, "190", Date.now() - (3 * 3600 * 1000))]
}
const pol190 = {
  id: "190",
  cookie: newCookie(),
  name: "Old policy of Dev #190",
  devices: ["190"],
  valid_from: (Date.now() - (10 * 3600 * 3600)).toString(),
  valid_until: (Date.now() + (5 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}
const pol191 = {
  id: "191",
  cookie: newCookie(),
  name: "Schedruled policy of Dev #191",
  devices: ["190"],
  valid_from: (Date.now() + (5 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: newPCRSet(8),
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

// unseen, w/ template policy
const dev200 = {
  id: "200",
  cookie: newCookie(),
  name: "Unseen with template policy",
  state: "trusted" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: ["200"],
  replaced_by: [],
  replaces: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: [newAppraisal(true, "200", Date.now() - (3 * 3600 * 1000))]
}
const pol200 = {
  id: "200",
  cookie: newCookie(),
  name: "Template policy for Device #200",
  devices: ["200"],
  valid_from: (Date.now() - (10 * 3600 * 3600)).toString(),
  revoked: false,
	pcr_template:["1","2"],
	fw_template: ["test-1", "test-2"],
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "template" as ChangeType,
    }
  ],
}

// trusted, firmware checks excluded
const dev210 = {
  id: "210",
  cookie: newCookie(),
  name: "Trusted, some checks whitelisted",
  state: "trusted" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: ["210"],
  replaced_by: [],
  replaces: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: [newAppraisal(true, "210", Date.now() - (3 * 3600 * 1000))]
}
const pol210 = {
  id: "210",
  cookie: newCookie(),
  name: "Template policy for Device #210",
  devices: ["210"],
  valid_from: (Date.now() - (10 * 3600 * 3600)).toString(),
  revoked: false,
	pcr_template:["1","2"],
	fw_template: ["test-1", "test-2"],
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "template" as ChangeType,
    }
  ],
}

const dev220 = {
  id: "220",
  cookie: newCookie(),
  name: "Trusted, no constrains",
  state: "trusted" as DeviceState,
  attributes: newAttribs(),
  hwid: newHwid(),
  policies: ["220"],
  replaced_by: [],
  replaces: [],
  changes: [
    {
      timestamp: (Date.now() - (2 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "enroll" as ChangeType,
    }
  ],
  appraisals: [newAppraisal(true, "220", Date.now() - (3 * 3600 * 1000))]
}
const pol220 = {
  id: "220",
  cookie: newCookie(),
  name: "Policy for Device #220",
  devices: ["220"],
  valid_until: (Date.now() + (10 * 3600 * 3600)).toString(),
  revoked: false,
  pcrs: {},
  changes: [
    {
      timestamp: (Date.now() - (10 * 3600 * 1000)).toString(),
      actor: actor1,
      type: "new" as ChangeType,
    }
  ],
}

export const devices: Device[] = [
  dev100, dev101,
  dev110,
  dev120,
  dev130,
  dev140,
  dev150,
  dev160,
  dev170,
  dev180,
  dev190,
  dev200,
  dev210,
  dev220,
]
export const policies: Policy[] = [
  pol100,
  pol120,
  pol130,
  pol140,
  pol150,
  pol160,
  pol180, pol181,
  pol190, pol191,
  pol200,
  pol210,
  pol220,
]
