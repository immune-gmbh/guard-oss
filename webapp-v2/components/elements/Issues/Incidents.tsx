import * as IssuesV1 from 'generated/issuesv1';
import Trans from 'next-translate/Trans';
import useTranslation from 'next-translate/useTranslation';

import { IIncident } from './Index';
import { KV, KVTable } from './Index';

function UefiOptionRomSet({ args }: IssuesV1.UefiOptionRomSet): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/option-rom-set';

  const forensics = args.devices.map(({ name, vendor, address, before, after }) => (
    <KVTable className="mt-4" key={name}>
      <KV k="PCI Device" v={vendor + ' ' + name} />
      <KV k="PCI Address" v={<code>{address}</code>} />
      <KV k="Before" v={<code>{before}</code>} />
      <KV k="After" v={<code>{after}</code>} />
    </KVTable>
  ));

  t(`${incidentKey}.slug`);

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {t(`${incidentKey}.description.d1`, {
            count: forensics.length,
            vendor: args.devices?.[0]?.vendor,
            name: args.devices?.[0]?.name,
          })}
        </p>
        <p className="mb-2">
          <Trans
            i18nKey={`${incidentKey}.description.d2`}
            components={{ thunderstrikeLink: <a href="https://trmm.net/Thunderstrike/" /> }}
          />
        </p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: forensics.length > 1,
    cta: t(`${incidentKey}.cta`),
  };
}

const EK_ISSUE_FIELDS = {
  model: 'Device Model',
  vendor: 'Device Vendor',
  ek_model: 'Model in Certificate',
  ek_vendor: 'Vendor in Certificate',
  ek_version: 'Version in Certificate',
  ek_issuer: 'Certificate Issuer',
};
const EK_ERROR_TEXT = {
  'san-invalid': "The certificate's Subject Alternative Name extension could no be parsed.",
  'san-mismatch':
    "The model and vendor name embedded in the certificate's Subject Alternative Name extension does not match the values read from the device.",
  'no-eku':
    'The certificate is not allowed to be used to identify a TPM. It lacks the TCG Endorsement Key Certificate (2.23.133.8.1) Extended Key Usage flag.',
  'no-signer':
    "We couldn't validate the TPM identity with our vendor CA certificate list. The issuer of the certificate used an untrusted key to sign it.",
  'wrong-signature':
    "The certificate's cryptographic signature did not validate. The certificate was likely manipulated in transit.",
};

function TpmEndorsementCertUnverified({ args }: IssuesV1.TpmEndorsementCertUnverified): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:tpm/endorsement-cert-unverified';

  const forensics = Object.entries(args).flatMap((ary: string[]) => {
    const k = EK_ISSUE_FIELDS[ary[0]];
    return k && ary[1] && <KV key={k} k={k} v={ary[1]} />;
  });

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {t(`${incidentKey}.description.d1`)} {EK_ERROR_TEXT[args.error] || ''}
        </p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <KVTable className="mt-4">{forensics}</KVTable>,
    collapsible: false,
    cta: t(`${incidentKey}.cta`),
  };
}

function CsmeNoUpdate({ args }: IssuesV1.CsmeNoUpdate): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:csme/no-update';

  const forensics = args.components.map(({ version, name, before, after }) => (
    <KVTable className="mt-4" key={name}>
      <KV k="CSME Component" v={name} />
      <KV k="CSME Version" v={version} />
      <KV k="Before" v={<code>{before}</code>} />
      <KV k="After" v={<code>{after}</code>} />
    </KVTable>
  ));

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {t(`${incidentKey}.description.d1`, {
            count: forensics.length,
            name: args.components?.[0]?.name,
          })}
        </p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: forensics.length > 1,
    cta: t(`${incidentKey}.cta`),
  };
}

function CsmeDowngrade({ args }: IssuesV1.CsmeDowngrade): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:csme/downgrade';

  const forensics = (
    <table className="mt-4 table-fixed">
      <thead className="text-left">
        <tr>
          <th className="pr-3">Component</th>
          <th className="pr-3">Before</th>
          <th>Now</th>
        </tr>
      </thead>
      <tbody>
        {args.combined && (
          <tr>
            <td className="pr-3">
              <i>Combined CSME</i>
            </td>
            <td className="pr-3">{args.combined.before}</td>
            <td>{args.combined.after}</td>
          </tr>
        )}
        {args.components?.map(({ name, before, after }) => (
          <tr key={name}>
            <td className="pr-3">{name}</td>
            <td className="pr-3">{before}</td>
            <td>{after}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
  const collapse = (args.combined ? 1 : 0) + args.components?.length > 4;

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {`${t(`${incidentKey}.intro`, {
            count: args.components?.length || 0,
            name: args.components?.[0]?.name,
          })} ${t(`${incidentKey}.description.d1`, { count: args.components?.length || 0 })}`}
        </p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: collapse,
    cta: t(`${incidentKey}.cta`),
  };
}

function UefiIbbNoUpdate({ args }: IssuesV1.UefiIbbNoUpdate): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/ibb-no-update';

  const forensics = (
    <KVTable className="mt-4">
      <KV k="Vendor" v={args.vendor} />
      <KV
        k="Version"
        v={
          <>
            {args.version}, released {args.release_date}
          </>
        }
      />
      <KV k="Before" v={<code>{args.before}</code>} />
      <KV k="After" v={<code>{args.after}</code>} />
    </KVTable>
  );

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: forensics,
    collapsible: false,
    cta: t(`${incidentKey}.cta`),
  };
}

function UefiBootAppSet({ args }: IssuesV1.UefiBootAppSet): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/boot-app-set';

  const forensics = args.apps?.map(({ path, before, after }, index) => {
    let changeRows: JSX.Element;
    if (before && !after) {
      changeRows = <KV k="File Deleted" v={<code>{before}</code>} />;
    } else if (after && !before) {
      changeRows = <KV k="New File" v={<code>{after}</code>} />;
    } else {
      changeRows = (
        <>
          <KV k="Before" key={`before-${index}`} v={<code>{before}</code>} />
          <KV k="After" key={`after-${index}`} v={<code>{after}</code>} />
        </>
      );
    }

    return (
      <KVTable className="mt-4" key={`path-${index}-${path}`}>
        <KV k="UEFI Path" key={`path-${index}`} v={path} />
        {changeRows}
      </KVTable>
    );
  });

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {t(`${incidentKey}.description.d1`, { count: forensics?.length || 0 })}
        </p>
        <p className="mb-2"></p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: (forensics?.length || 0) > 1,
    cta: t(`${incidentKey}.cta`),
  };
}

function UefiGptChanged({ args }: IssuesV1.UefiGptChanged): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/gpt-changed';

  const parts = args.partitions.map(({ guid, name, type, start, end }) => {
    return (
      <KVTable className="mt-4" key={guid}>
        <KV k="Partition" v={(name && `${name} (${guid})`) || guid} />
        <KV k="Type" v={type} />
        <KV k="Size" v={`${start} : ${end}`} />
      </KVTable>
    );
  });
  const forensics = (
    <KVTable className="mt-4">
      <KV k="Disk" v={args.guid} />
      <KV k="Before" v={args.before} />
      <KV k="After" v={args.after} />
    </KVTable>
  );

  return {
    slug: t(`${incidentKey}.slug`),
    description: <p className="mb-2">{t(`${incidentKey}.description`)}</p>,
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: (
      <>
        {forensics}
        {parts}
      </>
    ),
    collapsible: parts.length > 1,
    cta: t(`${incidentKey}.cta`),
  };
}

function UefiSecureBootKeys({ args }: IssuesV1.UefiSecureBootKeys): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/secure-boot-keys';

  const forensics = [args.pk]
    .concat(args.kek)
    .filter((k) => k !== undefined && k !== null)
    .map(({ subject, issuer, not_before, not_after, fpr }, idx: number) => (
      <KVTable className="mt-4" key={`${issuer}-${idx}`}>
        <KV k="Type" v={idx == 0 ? 'Platform Key (PK)' : ' Key Exchange Key (KEK)'} />
        <KV k="Subject" v={subject} />
        <KV k="Issuer" v={issuer} />
        <KV k="Valid From" v={not_before} />
        <KV k="Valid Until" v={not_after} />
        <KV k="Fingerprint" v={<code>{fpr}</code>} />
      </KVTable>
    ));

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: forensics.length > 1,
    cta: t(`${incidentKey}.cta`),
  };
}

function UefiSecureBootVariables({ args }: IssuesV1.UefiSecureBootVariables): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/secure-boot-variables';

  const forensics = Object.entries({
    setup_mode: 'SetupMode',
    audit_mode: 'AuditMode',
    deployed_mode: 'DeployedMode',
    secure_boot: 'SecureBoot',
  }).map((kv: string[]) => (
    <KV key={kv[0]} k={kv[1]} v={args[kv[0]] ? <code>{args[kv[0]]}</code> : <i>Missing</i>} />
  ));

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <KVTable className="mt-4">{forensics}</KVTable>,
    collapsible: forensics.length > 1,
    cta: t(`${incidentKey}.cta`),
  };
}

function UefiSecureBootDbx({ args }: IssuesV1.UefiSecureBootDbx): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/secure-boot-dbx';

  const forensics = args.fprs.map((fpr, idx) => (
    <div key={`${fpr}-${idx}`}>
      <code>{fpr}</code>
    </div>
  ));
  const collapse = forensics.length > 3;

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`, { count: args.fprs?.length })}</p>
        <p className="mb-2">
          <Trans
            i18nKey={`${incidentKey}.description.d2`}
            components={{
              code: <code />,
            }}
          />
        </p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: collapse,
    cta: (
      <Trans
        i18nKey={`${incidentKey}.cta`}
        components={{ uefiLink: <a href="https://uefi.org/revocationlistfile" />, code: <code /> }}
      />
    ),
  };
}

function UefiBootFailure({ args }: IssuesV1.UefiBootFailure): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/boot-failure';

  const byValue = new Map<string, Array<string>>();
  for (const k in args) {
    const v = args[k];
    if (v === '00000000') continue;
    if (!byValue.has(v)) byValue.set(v, []);
    byValue.get(v).push(k);
  }
  const forensics = Array.from(byValue.entries()).map((kv: [string, string[]]) => {
    const val = kv[0];
    const pcrs = <>{kv[1].map((x) => x.replace('pcr', 'PCR #')).join(', ')}</>;
    return <KV key={val} k={pcrs} v={val} />;
  });
  const collapse = forensics.length > 2;

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <KVTable className="mt-4">{forensics}</KVTable>,
    collapsible: collapse,
    cta: t(`${incidentKey}.cta`),
  };
}

const GRUB_CONFIG_KEYS = {
  kernel: 'Linux SHA-256',
  kernel_path: 'Path to Linux',
  initrd: 'Initial RAM disk SHA-256',
  initrd_path: 'Path to initial RAM disk',
  command_line: 'Linux command line',
};
function GrubBootChanged({ args }: IssuesV1.GrubBootChanged): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:grub/boot-changed';

  const forensics = Object.entries(args.before).flatMap((key: [string, string]) => {
    const k = key[0];
    const b = args.before[k];
    const a = args.after[k];
    if (JSON.stringify(a) === JSON.stringify(b)) return [];
    return (
      <div key={k}>
        <p className="mt-4 font-semibold">{GRUB_CONFIG_KEYS[k]}</p>
        <KVTable>
          <KV key="before" k="Before" v={b} />
          <KV key="after" k="Now" v={a} />
        </KVTable>
      </div>
    );
  });
  const kernel =
    args.before.kernel !== args.after.kernel || args.before.kernel_path !== args.after.kernel_path;
  const initrd =
    args.before.initrd !== args.after.initrd || args.before.initrd_path !== args.after.initrd_path;

  const descriptionKey = (): string => {
    if (!kernel && !initrd) return 'notKernelNotInit';
    if (kernel && !initrd) return 'kernelNotInit';
    if (!kernel && initrd) return 'notKernelInit';
    return 'kernelAndInit';
  };

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {t(`${incidentKey}.description.${descriptionKey()}`)}
          {t(`${incidentKey}.description.d1`)}
        </p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: forensics.length > 2,
    cta: t(`${incidentKey}.cta.${descriptionKey()}`),
  };
}

function TpmNoEventlog(): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:tpm/no-eventlog';

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    cta: t(`${incidentKey}.cta`),
  };
}

function TpmInvalidEventlog(): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:tpm/invalid-eventlog';

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    cta: t(`${incidentKey}.cta`),
  };
}

function TpmDummy(): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:tpm/dummy';

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    cta: t(`${incidentKey}.cta`),
  };
}

function EsetDisabled(): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:eset/disabled';

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    cta: t(`${incidentKey}.cta`),
  };
}

function EsetNotStarted({ args }: IssuesV1.EsetNotStarted): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:eset/not-started';

  const forensics = (
    <table className="mt-4 table-fixed">
      <thead className="text-left">
        <tr>
          <th className="pr-3">Component</th>
          <th>Started</th>
        </tr>
      </thead>
      <tbody>
        {args.components.map((f) => (
          <tr key={f.path}>
            <td className="pr-3">
              <code>{f.path}</code>
            </td>
            <td>{f.started ? 'Yes' : <b>No</b>}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
  const ncs = args.components.filter((c) => !c.started);

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {t(`${incidentKey}.description.d1`, { count: ncs?.length || 0, nscPath: ncs?.[0]?.path })}
        </p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: forensics,
    collapsible: true,
    cta: t(`${incidentKey}.cta`),
  };
}

function EsetExcludedSet({
  args: { files: fs, processes: ps },
}: IssuesV1.EsetExcludedSet): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:eset/excluded-set';

  const forensics = (
    <>
      <p className="mt-4 font-semibold">Files</p>
      <div className="flex flex-col">
        {fs.map((f, idx) => (
          <div key={`f-${f}-${idx}`}>{f}</div>
        ))}
      </div>
      <p className="mt-4 font-semibold">Processes</p>
      <div className="flex flex-col">
        {ps.map((f, idx) => (
          <div key={`p-${f}-${idx}`}>{f}</div>
        ))}
      </div>
    </>
  );
  const f0 = fs.length == 0;
  const f1 = fs.length == 1 && fs[0];
  const f2 = fs.length > 1 && fs.length;
  const p0 = ps.length == 0;
  const p1 = ps.length == 1 && ps[0];
  const p2 = ps.length > 1 && ps.length;

  const descriptionInit = (): string | void => {
    if (f0 && p1) return 'noFilesOneProcess';
    if (f0 && p2) return 'noFilesMoreProcesses';
    if (f1 && p0) return 'oneFileNoProcess';
    if (f1 && p1) return 'oneFileOneProcess';
    if (f1 && p2) return 'oneFileMoreProcesses';
    if (f2 && p0) return 'moreFilesNoProcess';
    if (f2 && p1) return 'moreFilesOneProcess';
    if (f2 && p2) return 'moreFilesMoreProcesses';
    return 'noFilesNoProcess';
  };

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {`${t(`${incidentKey}.description.${descriptionInit()}`, {
            process: p1,
            processCount: ps.length,
            file: f1,
            fileCount: fs.length,
          })} ${t(`${incidentKey}.description.d1`)}`}
        </p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: forensics,
    collapsible: fs.length + ps.length > 3,
    cta: t(`${incidentKey}.cta`),
  };
}

function EsetManipulated({ args: { components: cs } }: IssuesV1.EsetManipulated): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:eset/manipulated';

  const forensics = cs.map(({ path, before, after }) => (
    <KVTable key={path} className="mt-4">
      <KV k="File" v={path} />
      <KV k="Before" v={<code>{before}</code>} />
      <KV k="After" v={<code>{after}</code>} />
    </KVTable>
  ));

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {t(`${incidentKey}.description.d1`, { count: cs.length, path: cs[0]?.path })}
        </p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: cs.length > 3,
    cta: t(`${incidentKey}.cta`),
  };
}

function WindowsBootCounterReplay({ args }: IssuesV1.WindowsBootCounterReplay): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:windows/boot-counter-replay';

  const forensics = (
    <KVTable className="mt-4">
      <KV k="Latest observed" v={args.latest} />
      <KV k="Received" v={args.received} />
    </KVTable>
  );

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: false,
    cta: t(`${incidentKey}.cta`),
  };
}

function WindowsBootLogQuotes(): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:windows/boot-log-quotes';

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    collapsible: false,
    cta: t(`${incidentKey}.cta`),
  };
}

function ImaBootAggregate({ args }: IssuesV1.ImaBootAggregate): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:ima/boot-aggregate';

  const forensics = (
    <KVTable className="mt-4">
      <KV k="Received from device" v={<code>{args.computed}</code>} />
      <KV k="Logged" v={<code>{args.logged}</code>} />
    </KVTable>
  );

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: false,
    cta: t(`${incidentKey}.cta`),
  };
}

function ImaRuntimeMeasurements({
  args: { files: fs },
}: IssuesV1.ImaRuntimeMeasurements): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:ima/runtime-measurements';

  const forensics = fs.map(({ path, before, after }) => (
    <KVTable key={path} className="mt-4">
      <KV k="File" v={path} />
      <KV k="Before" v={<code>{before}</code>} />
      <KV k="Now" v={<code>{after}</code>} />
    </KVTable>
  ));

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {t(`${incidentKey}.description.d1`, { count: fs.length, path: fs[0]?.path })}
        </p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: fs.length > 3,
    cta: t(`${incidentKey}.cta`),
  };
}

function ImaInvalidLog({ args }: IssuesV1.ImaInvalidLog): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:ima/invalid-log';

  const forensics = args.pcr.map(({ computed, quoted, number }) => (
    <div key={`pcr-${number}`}>
      <p className="mt-4">PCR #{number}</p>
      <KVTable>
        <KV k="Computed" v={<code>{computed}</code>} />
        <KV k="Received from device" v={<code>{quoted}</code>} />
      </KVTable>
    </div>
  ));

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: args.pcr.length > 3,
    cta: t(`${incidentKey}.cta`),
  };
}

function PolicyIntelTsc(): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:policy/intel-tsc';

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    cta: t(`${incidentKey}.cta`),
  };
}

function PolicyEndpointProtection(): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:policy/endpoint-protection';

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    cta: t(`${incidentKey}.cta`),
  };
}

function TscEndorsementCertificate({ args }: IssuesV1.TscEndorsementCertificate): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:tsc/endorsement-certificate';

  const forensics = (
    <KVTable className="mt-4">
      <KV k="Endorsement key certificate issuer" v={args.ek_issuer} />
      <KV k="Endorsement key certificate serial" v={args.ek_serial} />
      <KV k="Platform certificate holder issuer" v={args.holder_issuer} />
      <KV k="Platform certificate holder serial" v={args.holder_serial} />
      <KV k="TSC meta data serial" v={args.xml_serial} />
    </KVTable>
  );
  // XXX
  const holderIssuer = args.error == 'holder-issuer';
  const xmlSerial = args.error == 'xml-serial';
  const holderSerial = args.error == 'holder-serial';

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">
          {holderIssuer && ''}
          {xmlSerial && ''}
          {holderSerial && ''}
          {t(`${incidentKey}.description`)}
        </p>
        <p className="mb-2"></p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: forensics,
    cta: t(`${incidentKey}.cta`),
  };
}

function TscPcrValues({ args }: IssuesV1.TscPcrValues): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:tsc/pcr-values';

  const forensics = args.values.map(({ tsc, quoted, number: n }, idx) => (
    <KVTable className="mt-4" key={`${n}-${idx}`}>
      <KV k="PCR" v={n} />
      <KV k="Registered at TSC" v={<code>{tsc}</code>} />
      <KV k="Received from device" v={<code>{quoted}</code>} />
    </KVTable>
  ));

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description`)}</p>
        <p className="mb-2"></p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: args.values.length > 3,
    cta: t(`${incidentKey}.cta`),
  };
}

function UefiBootOrder({ args }: IssuesV1.UefiBootOrder): IIncident {
  const { t } = useTranslation();
  const incidentKey = 'incidents:uefi/boot-order';

  const forensics = args.variables.map(({ before, after, name }) => (
    <div key={name}>
      <p className="mt-4 font-semibold">{name}</p>
      <KVTable>
        <KV k="Before" v={<code>{before}</code>} />
        <KV k="Now" v={<code>{after}</code>} />
      </KVTable>
    </div>
  ));

  return {
    slug: t(`${incidentKey}.slug`),
    description: (
      <>
        <p className="mb-2">{t(`${incidentKey}.description.d1`)}</p>
        <p className="mb-2">{t(`${incidentKey}.description.d2`)}</p>
      </>
    ),
    forensicsPre: t(`${incidentKey}.forensicsPre`),
    forensicsPost: <>{forensics}</>,
    collapsible: args.variables.length > 2,
    cta: t(`${incidentKey}.cta`),
  };
}

export const INCIDENTS = {
  'csme/no-update': CsmeNoUpdate,
  'csme/downgrade': CsmeDowngrade,
  'uefi/option-rom-set': UefiOptionRomSet,
  'uefi/boot-app-set': UefiBootAppSet,
  'uefi/ibb-no-update': UefiIbbNoUpdate,
  'uefi/gpt-changed': UefiGptChanged,
  'uefi/secure-boot-keys': UefiSecureBootKeys,
  'uefi/secure-boot-variables': UefiSecureBootVariables,
  'uefi/secure-boot-dbx': UefiSecureBootDbx,
  'uefi/boot-failure': UefiBootFailure,
  'uefi/boot-order': UefiBootOrder,
  'tpm/endorsement-cert-unverified': TpmEndorsementCertUnverified,
  'tpm/no-eventlog': TpmNoEventlog,
  'tpm/invalid-eventlog': TpmInvalidEventlog,
  'tpm/dummy': TpmDummy,
  'grub/boot-changed': GrubBootChanged,
  'eset/disabled': EsetDisabled,
  'eset/not-started': EsetNotStarted,
  'eset/excluded-set': EsetExcludedSet,
  'eset/manipulated': EsetManipulated,
  'windows/boot-log-quotes': WindowsBootLogQuotes,
  'windows/boot-counter-replay': WindowsBootCounterReplay,
  'windows/boot-log': WindowsBootLogQuotes,
  'ima/invalid-log': ImaInvalidLog,
  'ima/boot-aggregate': ImaBootAggregate,
  'ima/runtime-measurements': ImaRuntimeMeasurements,
  'policy/intel-tsc': PolicyIntelTsc,
  'policy/endpoint-protection': PolicyEndpointProtection,
  'tsc/endorsement-certificate': TscEndorsementCertificate,
  'tsc/pcr-values': TscPcrValues,
};
