import * as IssuesV1 from 'generated/issuesv1';
import Trans from 'next-translate/Trans';
import useTranslation from 'next-translate/useTranslation';

import { IRisk } from './Index';
import { KV, KVTable } from './Index';

function BinarlyIssue({ id }: IssuesV1.Binarly): IRisk {
  const { t } = useTranslation();

  return {
    score: parseFloat(t(`binarly:${id}.score`)),
    group: t(`binarly:${id}.group`),
    slug: <Trans i18nKey={`binarly:${id}.slug`} />,
    description: (
      <Trans
        i18nKey={`binarly:${id}.description`}
        components={[
          <a key="1" className="underline decoration-solid" href={t(`binarly:${id}.link`)} />,
        ]}
      />
    ),
  };
}

function UefiNoExitBootSrv({ args }: IssuesV1.UefiNoExitBootSrv): IRisk {
  const { t } = useTranslation();

  const forensicsKey = args.entered ? 'entered' : 'not_entered';

  return {
    slug: t('risks:uefi/no-exit-boot-srv.slug'),
    group: 'configuration-failure',
    description: (
      <>
        <p className="mb-2">
          {forensicsKey && (
            <Trans
              i18nKey={`risks:uefi/no-exit-boot-srv.forensics.${forensicsKey}`}
              components={{ code: <code /> }}
            />
          )}
          {t('risks:uefi/no-exit-boot-srv.description.d1')}
        </p>
        <p className="mb-2">
          <Trans
            i18nKey="risks:uefi/no-exit-boot-srv.description.d2"
            components={{ code: <code /> }}
          />
        </p>
      </>
    ),
  };
}

function UefiOfficialDbx({ args }: IssuesV1.UefiSecureBootDbx): IRisk {
  const { t } = useTranslation();

  return {
    slug: t('risks:uefi/official-dbx.slug'),
    group: 'update',
    description: (
      <>
        <p className="mb-2">
          {t('risks:uefi/official-dbx.description.d1', { count: args.fprs.length })}
        </p>
        <p className="mb-2">
          <Trans
            i18nKey="risks:uefi/official-dbx.description.d2"
            components={{
              code: <code />,
              revocationLink: <a href="https://uefi.org/revocationlistfile" />,
            }}
          />
        </p>
      </>
    ),
  };
}

function FirmwareUpdate({ args }: IssuesV1.FirmwareUpdate): IRisk {
  const { t } = useTranslation();

  const forensics = args.updates.map(({ name, current, next }, index) => (
    <div key={`fw-update-${name}-${index}`}>
      <p className="mt-4 font-semibold">{name}</p>
      <KVTable>
        <KV k="Current version" v={<code>{current}</code>} />
        <KV k="Updated version" v={<code>{next}</code>} />
      </KVTable>
    </div>
  ));

  return {
    slug: t('risks:fw/update.slug'),
    group: 'update',
    description: (
      <>
        <p className="mb-2">
          <Trans
            i18nKey="risks:fw/update.description.d1"
            components={{ code: <code /> }}
            values={{ count: args.updates.length }}
          />
        </p>
        <p className="mb-2">{t('risks:fw/update.description.d2')}</p>
        {!!forensics.length && <div className="mb-4">{forensics}</div>}
        <p className="mb-2">
          <Trans
            i18nKey="risks:fw/update.description.d3"
            components={{ lvfsLink: <a href="https://lvfs.org/" /> }}
          />
        </p>
      </>
    ),
  };
}

const WINDOWS_BOOT_CFG = {
  boot_debugging: ['boot debugging', true],
  kernel_debugging: ['kernel debugging', true],
  dep_disabled: ['Data Execution Prevention', true],
  code_integrity_disabled: ['Code Integrity', true],
  test_signing: ['test signing', true],
};

function WindowsBootConfig({ args }: IssuesV1.WindowsBootConfig): IRisk {
  const { t } = useTranslation();

  const forensics = (
    <KVTable className="mt-4">
      {Object.entries(args)
        .filter((kv) => kv[0] in WINDOWS_BOOT_CFG)
        .map((kv) => {
          const k = WINDOWS_BOOT_CFG[kv[0]][0];
          const p = WINDOWS_BOOT_CFG[kv[0]][1] as boolean;
          return (
            <KV key={kv[0]} k={k[0].toUpperCase() + k.slice(1)} v={kv[1] !== p ? 'Yes' : 'No'} />
          );
        })}
    </KVTable>
  );
  const comp = Object.entries(args).filter((kv) => kv[0] in WINDOWS_BOOT_CFG && kv[1]);

  return {
    slug: t(`risks:windows/boot-config.slug`),
    group: 'configuration-failure',
    description: (
      <>
        <p className="mb-2">
          {t(`risks:windows/boot-config.description.d1`, {
            count: comp.length,
            bootCfg: WINDOWS_BOOT_CFG[comp[0]?.[0]],
          })}
        </p>
        <p className="mb-2">{t(`risks:windows/boot-config.description.d2`)}</p>
        <div className="mb-2">{forensics}</div>
      </>
    ),
  };
}

export const RISKS = {
  'fw/update': FirmwareUpdate,
  'uefi/no-exit-boot-srv': UefiNoExitBootSrv,
  'uefi/official-dbx': UefiOfficialDbx,
  'windows/boot-config': WindowsBootConfig,
  'brly/': BinarlyIssue,
};
