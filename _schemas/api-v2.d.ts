type Membership = {
  // writable 
  role: "owner" | "user"
  notifications: string[] // "alerts", "invoices"

  // can be PATCH'd between active <-> deleted and invited -> deleted
  state: "invited" | "active" | "deleted"

  // writable once
  user: string
  org: string
  // unique id: (user, org)
  default: boolean // only one default membership per user
}

type User = {
  // writable
  name: string
  email: string
  totp: boolean
  admin: boolean
  organizations: Membership[]

  // write only
  password: string | undefined
  cookie: string

  // read only
  id: string
  provider: "native" | "google" | "github"
  state: "invited" | "active" | "suspended"
}

type Organization = {
  // writable
  name: string
  users: Membership[]

  splunk: Splunk
  syslog: Syslog
  state: "active" | "suspended" | "deleted"

  // writable once
  cookie: string

  // read only
  id: string
  slots: number
  maximum_slots: number
  forecast: number
  fixed_price: number
  variable_price: number
  invoices: Invoice[]
}

type Splunk = {
  // writable
  enabled: boolean
  url: string
  // write only
  token: string
}

type Syslog = {
  // writable
  enabled: boolean
  hostname: string
  port: number
}

type Invoice = {
  year: number
  month: number
  state: "paid" | "void" | "open"
  warning: number // 0 <= x < 3
  pdf: string // link
  sum: number
}

// /v2/devices (apisrv)
type Device = {
  // writable
  name: string
  attributes: Map<string, string | undefined>

  // can only PATCH'd to "retired"
  state: "new" | "unseen" | "vulnerable" | "trusted" | "outdated" | "retired"

  // writeable once
  cookie: string

  // read only
  id: string
  fpr: string

  replaces: string | undefined
  replaced_by: string | undefined

  state_timestamp: string

  last_update_timestamp: string
  last_update_actor: string

  appraisals: Appraisal[]
}

type RenameDevice = (dev: Device, new_name: string) => Error_ | undefined
type RetireDevice = (dev: Device) => Error_ | undefined
// ReplaceDevice(d, nil)
// for p in dev.Policies
//   RevokePolicyIfUsed(p)
type RegisterDevice = (fpr: string, name_hint: string) => Error_ | Device
// oldDevs = GetDevicesByFpr(fpr)
// dev = NewDevice(fpr, name_hint)
// for d in oldDevs
//   ReplaceDevice(d, dev)
//   for p in d.Policies
//     RevokePolicyIfUsed(p)
// return dev
type EnrollDevice = (fpr: string, keys: Key[], name_hint: string) => Error_ | undefined
// dev = RegisterDevice(fpr, name_hint)
// for k in keys:
//   NewKey(k, dev)
// NewTemplatePolicy(dev)
type AttestDevice = (evidence: Evidence) => Appraisal
// tmpls = TemplatePolicies(credential.device_id, evidence.signer_qn, now)
// if len(tmpls) > 0
//   for t,d,k in tmpls
//     if ValidateQuote(evidence, k)
//       InstanciateTemplate(t, d, evidence)
//       return IssueAppraisal(evidence, d, p)
// else
//   pols = GetPolicies(credential.device_id, evidence.signer_qn, now)
//   for p,d,k in pols
//     if ValidateQuote(evidence, k) and MatchPolicy(p, evidence)
//       return IssueGoodAppraisal(evidence, d, p)
//         
// return IssueBadAppraisal(evidence, d, pols)

// /v2/policies (apisrv)
type Policy = {
  // writable
  name: string
  devices: string[]

  valid_since: string | undefined
  valid_until: string | undefined

  // writable once
  pcr_template: Map<string, string> | undefined // PCR -> (sha256 | "")
  cookie: string

  pcrs: Map<string, string> | undefined // PCR -> sha256
  fw_overrides: string[] | undefined // Annotation::id

  // read only
  id: string
  revoked: boolean
}

// /v2/enroll (apisrv)
type EncryptedCredential = {
  name: string
	credential: string // encrypted Credential_
	secret: string
	certificate: string
	nonce: string
}

// /v2/enroll (apisrv)
// JWT send as Bearer token to authenticate /v2/attest
type Credential_ = {
  iss: string // issuing service
  sub: string // device's platform id
  aud: string // apisrv
  nbf: string // now
  scope: "attest"
}

// /v2/enroll (apisrv)
type Enrollment = {
  name_hint: string
  endoresment_key: string
  endoresment_certificate: string | undefined
  root: string
  keys: Key[]
  cookie: string
}

// /v2/enroll (apisrv)
type Key = {
  name: string
  pub: string
	certify_info: string
	certify_signature: string
}

// /v2/attest (apisrv)
type Evidence = {
  quote: string
	// TPMT_SIGNATURE (Base64 encoded, unique)
	signature: string
	// TPMI_ALG_HASH
	algorithm: string
	// PCRs (Hex encoded)
	pcrs: Map<string, string>
	// Unprocessed firmware info
	firmware: FirmwareProperies
  cookie: string
}

type FirmwarePropertiesRequest = {
  // PK, KEK, PKDefault, KEKDefault, db, dbx, dbt, dbr, dbDefault, dbxDefault,
  // dbtDefault, dbrDefault, AuditMode, SetupMode, DeployedMode, SecureBoot
  uefi_variables: UEFIVariable[]

	// 0xc0010010 (sev), 0x3a (sgx)
  msrs: string[]

  // Format: eax or eax/ecx both hex
  // Default: 0x8000001f (sev), 0x7 (sgx), 0x12/0x0 to 0x12/0x9 (sgx)
  cpuid_leaves: CPUIDLeaf[]

  // PLATFORM_STATUS, PDH_CERT_EXPORT, GET_ID, PEK_CSR
  sev_commands: SEVCommand[]

  // GEN GetFWVersion, HMRFPO Enable, HMRFPO GetStatus, FWCAPS FeatureState,
  // FWCAPS GetLocalUpdate
  me_commands: string[]

  // GetMEBiosInterface, GetVendorLabel
  sps_commands: string[]

  // GetCapability (TPM_PT_*): LEVEL, REVISION, DAY_OF_YEAR, YEAR,
  // MANUFACTURER, VENDOR_STRING_1-4, FIRMWARE_VERSION_1-2
  tpm2_capailities: TPM2Capability[]

  // 8086:0.16 (me)
  pci_config_space: PCIDeviceId[]
}

type UEFIVariable = {
  vendor: string
  name: string
}

type CPUIDLeaf = {
  eax: string
  ecx: string | undefined
}

type SEVCommand = "PLATFORM_STATUS" | "PDH_CERT_EXPORT" | "GET_ID" | "PEK_CSR"

type TPM2Capability = {
  capability: string
  proptery: string
  count: string
}

type PCIDeviceId = {
  vendor: string
  device: string
  function: string
}

// /v2/attest (apisrv)
type FirmwareProperies = FirmwareReport
type FirmwareProperies2 = {
  // FirmwarePropertiesRequest
  uefi_variables: Map<string, string>
  msrs: Map<string, string>
  sev: Map<string, string>
  me: Map<string, string>
  sps: Map<string, string>
  tpm: Map<string, string>

  // Hard coded
  smbios: string
  txt_public: string
  flash: string
  event_log: string
  hostname: string
  os: string
}

// /v2/devices (apisrv)
type Appraisal = {
  received: string
  // evidence.signature is unique
  evidence: Evidence

  verdict: boolean
  device: string
  policy: string

  // derived from policy and evidence
  firmware: FirmwareReport
  annotations: {
    good: Annotation[]
    bad: Annotation[]
  }
  pcrs: {
    valid: Map<string, string>
    invalid: Map<string, string>
    ignored: Map<string, string>
  }
}

// /v2/devices (apisrv)
type FirmwareReport = {
  os: {
    hostname: string
    type: string
  }
  smbios: {
    manufacturer: string
    product: string
    bios_release_date: string
    bios_vendor: string
    bios_version: string
  }
  uefi: {
    secureboot: "enabled" | "disabled" | "setup"
    platform_keys: string[]
    exchange_keys: string[]
    permitted_keys: string[]
    forbidden_keys: string[]
  }
  tpm: {
    manufacturer: string
    vendor_id: string
    spec_version: number
    event_log: TPMEvent[]
  }
  csme:{
    variant: "icu" | "txe" | "consumer" | "business" | "light" | "sps" | "unknown"
    version: number[]
    recovery_version: number[]
    fitc_version: number[]
  }
  sgx: {
    version: number
    enabled: boolean
    flc: boolean
    kss: boolean
    enclave_size_32: number
    enclave_size_64: number
    epc: EnclavePageCache[]
    txt:{
      ready: boolean
    }
    sev: {
      enabled: boolean
      version: number[]
      sme: boolean
      es: boolean
      vte: boolean
      snp: boolean
      vmpl: boolean
      guests: number
      min_asid: number
    }  
  }
}

// /v2/devices (apisrv)
type TPMEvent = {
  pcr: number
  value: string
  algorithm: number
  note: string
}

// /v2/devices (apisrv)
type EnclavePageCache = {
  base: number
  size: number
  cir_protection: boolean
}

// /v2/devices (apisrv)
type Annotation = {
  id: string
  expected: string[],
  path: string,
}

// /v2/info (apisrv)
type Info = {
  api_version: "2"
}

type Envelope = {
  code: "ok" | "err"
  errors: Error_[]
  data: {
    devices: Device[] | undefined
    policies: Policy[] | undefined
    credentials: EncryptedCredential[] | undefined
    info: Info | undefined
  }
  meta: {
    next: string | undefined
  }
}

type Error_ = {
  id: string
  msg: string
  path: string | undefined
  fatal: boolean
}

// /v2/devices (apisrv)
// /v2/policies (apisrv)
// /v2/* (authsrv)
// JWT send as Bearer token
type Token = {
  iss: string // issuing service
  exp: string // future,
  sub: string // organization id,
  act: {
    sub: string // service name
  } | undefined,
  aud: string // other service's name
  nbf: string // now
  scope: "event" | "public" // event -> can POST /v2/events
}

// CloudEvents 1.0.1
type Event_<T> = {
  // unique id: (id, source)
  id: string
  source: string // k8s pod id
  specversion: "1.0"
  type: string
  dataschema: string
  subject: string
  time: string // RFC 3339
  data: T
}

// /v2/events (apisrv)
// dataschema: "https://immu.ne/specs/v2/events/quota-update.schema.json"
// subject: organization.id
// type: "ne.immu.v2.quote-update"
type QuotaUpdateEvent = Event_<QuotaUpdateEventPayload>
type QuotaUpdateEventPayload = {
  organization: string
  devices: number
  features: string[]
}

// /v2/events (authsrv)
// dataschema: "https://immu.ne/specs/v2/events/new-appraisal.schema.json"
// subject: device.id
// type: "ne.immu.v2.new-appraisal"
type NewAppraisalEvent = Event_<NewAppraisalEventPayload>
type NewAppraisalEventPayload = {
  device: Device
  previous: Appraisal
  next: Appraisal
}

// /v2/events (authsrv)
// dataschema: "https://immu.ne/specs/v2/events/appraisal-expired.schema.json"
// subject: device.id
// type: "ne.immu.v2.appraisal-expired"
type AppraisalExpiredEvent = Event_<AppraisalExpiredEventPayload>
type AppraisalExpiredEventPayload = {
  device: Device
  previous: Appraisal
}
