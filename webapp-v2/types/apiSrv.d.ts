import * as IssuesV1 from 'generated/issuesv1';

export namespace ApiSrv {
  export interface Change {
    devices?: Device;
    id: string;
    timestamp: string;
    actor: string;
    type:
      | 'enroll'
      | 'resurrect'
      | 'rename'
      | 'tag'
      | 'associate'
      | 'template'
      | 'new'
      | 'instanciate'
      | 'revoke'
      | 'retire';
    comment: string;
  }

  export interface Evidence {}

  export interface Host {
    name: string;
    hostname: string;
    type: string;
    cpu_vendor: string;
  }

  export interface Smbios {
    bios_release_date: string;
    bios_vendor: string;
    bios_version: string;
    manufacturer: string;
    product: string;
    serial: string;
    uuid: string;
  }

  export interface Annotation {
    id: string;
    path: string;
    expected?: string;
    fatal: boolean;
  }

  export interface Nic {
    name: string;
    mac: string;
    ipv4: string[];
    ipv6: string[];
  }

  export interface Values {
    host: Host;
    nics: Nic[];
    smbios: Smbios;
    agent_release: string;
  }

  export interface Report {
    type: string;
    values: Values;
    annotations: Annotation[];
  }

  export type VerdictStatus = 'trusted' | 'unsupported' | 'vulnerable';

  export interface Verdict {
    bootloader: VerdictStatus;
    configuration: VerdictStatus;
    endpoint_protection: VerdictStatus;
    firmware: VerdictStatus;
    operating_system: VerdictStatus;
    result: VerdictStatus;
    supply_chain: VerdictStatus;
    type: string;
  }

  export interface AppraisalData {
    type: string;
    id: string;
    expires: Date;
    appraised: Date;
    report: Report;
    verdict: Verdict;
    issues: Issues;
  }

  export interface Device {
    id: string;
    hwid: string;
    name: string;
    policy: {
      endpoint_protection: 'off' | 'on' | 'if-present';
      intel_tsc: 'off' | 'on' | 'if-present';
    };
    state: 'trusted' | 'vulnerable' | 'new' | 'resurrectable' | 'outdated' | 'unseen' | 'retired';
    appraisals: AppraisalData[];
    tags?: Tag[];
    replaces: string[];
    replaced_by: string[];
    attestation_in_progress?: Date;
  }

  export interface Issues {
    type: 'issues/v1';
    issues: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml[];
  }

  export interface Tag {
    id: string;
    key: string;
    score: double;
  }
}
