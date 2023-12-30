export enum Aspect {
  SupplyChain = 'supply-chain',
  Configuration = 'configuration',
  Firmware = 'firmware',
  Bootloader = 'bootloader',
  OperatingSystem = 'operating-system',
  EndpointProtection = 'endpoint-protection',
}

export const issuesAspects = {
  'csme/no-update': Aspect.Firmware,
  'csme/downgrade': Aspect.Firmware,
  'uefi/option-rom-set': Aspect.Firmware,
  'uefi/boot-app-set': Aspect.Bootloader,
  'uefi/ibb-no-update': Aspect.Firmware,
  'uefi/no-exit-boot-srv': Aspect.Firmware,
  'uefi/gpt-changed': Aspect.Bootloader,
  'uefi/secure-boot-keys': Aspect.Firmware,
  'uefi/secure-boot-variables': Aspect.Configuration,
  'uefi/secure-boot-dbx': Aspect.Configuration,
  'uefi/official-dbx': Aspect.Configuration,
  'uefi/boot-failure': Aspect.Firmware,
  'uefi/boot-order': Aspect.Configuration,
  'tpm/endorsement-cert-unverified': Aspect.SupplyChain,
  'tpm/no-eventlog': Aspect.Firmware,
  'tpm/invalid-eventlog': Aspect.Firmware,
  'tpm/dummy': Aspect.SupplyChain,
  'grub/boot-changed': Aspect.Bootloader,
  'eset/disabled': Aspect.EndpointProtection,
  'eset/not-started': Aspect.EndpointProtection,
  'eset/excluded-set': Aspect.EndpointProtection,
  'eset/manipulated': Aspect.EndpointProtection,
  'windows/boot-config': Aspect.OperatingSystem,
  'windows/boot-log': Aspect.OperatingSystem,
  'windows/boot-counter-replay': Aspect.OperatingSystem,
  'ima/invalid-log': Aspect.EndpointProtection,
  'ima/boot-aggregate': Aspect.EndpointProtection,
  'ima/runtime-measurements': Aspect.EndpointProtection,
  'tsc/pcr-values': Aspect.SupplyChain,
  'tsc/endorsement-certificate': Aspect.SupplyChain,
  'policy/endpoint-protection': Aspect.EndpointProtection,
  'policy/intel-tsc': Aspect.SupplyChain,
  'fw/update': Aspect.Firmware,
  'brly/': Aspect.Firmware,
};

export const aspectByIssueId = (issueId: string) => {
  const issueKey = issueId.startsWith('brly/') ? 'brly/' : issueId;
  return issuesAspects[issueKey];
};
