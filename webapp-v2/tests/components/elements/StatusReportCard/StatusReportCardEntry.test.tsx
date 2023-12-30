/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { render, screen, within } from '@testing-library/react';
import StatusReportCardEntry from 'components/elements/StatusReportCard/StatusReportCardEntry';
import * as IssuesV1 from 'generated/issuesv1';
import I18nProvider from 'next-translate/I18nProvider';
import { mockReportCardIncident } from 'tests/mocks';

import common from 'locales/en/common.json';
import incidents from 'locales/en/incidents.json';
import risks from 'locales/en/risks.json';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

describe('uefiOptionRomSet incident is correctly rendered', () => {
  const device = {
    name: 'deviceName',
    vendor: 'vendorName',
    before: 'beforeValue',
    after: 'afterValue',
  };

  test('singular', () => {
    const incident = mockReportCardIncident({
      id: 'uefi/option-rom-set',
      args: { devices: [device] },
    }) as IssuesV1.UefiOptionRomSet;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(/The firmware of the vendorName deviceName device that is/i),
    ).toBeInTheDocument();
  });

  test('plural', () => {
    const devices = [device, { ...device, name: 'deviceName2' }];
    const incident = mockReportCardIncident({
      id: 'uefi/option-rom-set',
      args: { devices },
    }) as IssuesV1.UefiOptionRomSet;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(/The firmware of 2 devices that is/i),
    ).toBeInTheDocument();
    expect(screen.getByText('Thunderstrike').closest('a')).toHaveAttribute(
      'href',
      'https://trmm.net/Thunderstrike/',
    );
  });
});

describe('csmeNoUpdate incident is correctly rendered', () => {
  const component = {
    name: 'deviceName',
    version: '2.0',
    before: '1.8',
    after: '1.9',
  };

  test('singular', () => {
    const incident = mockReportCardIncident({
      id: 'csme/no-update',
      args: { components: [component] },
    }) as IssuesV1.CsmeNoUpdate;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(
        /The deviceName firmware component of the Converged Security/i,
      ),
    ).toBeInTheDocument();
  });

  test('plural', () => {
    const components = [component, { ...component, name: 'deviceName2' }];
    const incident = mockReportCardIncident({
      id: 'csme/no-update',
      args: { components },
    }) as IssuesV1.UefiOptionRomSet;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(/2 firmware components of the Converged Security/i),
    ).toBeInTheDocument();
  });
});

describe('tpmEndorsementCertUnverified incident is correctly rendered', () => {
  const args = {
    error: 'san-invalid',
    vendor: 'vendorName',
  };

  test('singular', () => {
    const incident = mockReportCardIncident({
      id: 'tpm/endorsement-cert-unverified',
      args,
    }) as IssuesV1.TpmEndorsementCertUnverified;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(/verified. The certificate's Subject Alternative/i),
    ).toBeInTheDocument();
  });
});

describe('csmeDowngrade incident is correctly rendered', () => {
  const component = {
    name: 'componentName',
    before: 'beforeName',
    after: 'afterName',
  };

  test('zero - no components', () => {
    const incident = mockReportCardIncident({
      id: 'csme/downgrade',
      args: { components: null },
    }) as IssuesV1.CsmeDowngrade;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(
        /The device's Converged Security and Management Engine \(CSME\) firmware was reverted/i,
      ),
    ).toBeInTheDocument();
  });

  test('singular', () => {
    const incident = mockReportCardIncident({
      id: 'csme/downgrade',
      args: { components: [component] },
    }) as IssuesV1.CsmeDowngrade;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(
        /The componentName firmware component of the Converged Security and Management Engine \(CSME\) firmware was reverted/i,
      ),
    ).toBeInTheDocument();
  });

  test('plural', () => {
    const components = [component, { ...component, name: 'componentseName2' }];
    const incident = mockReportCardIncident({
      id: 'csme/downgrade',
      args: { components },
    }) as IssuesV1.CsmeDowngrade;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(
        /2 firmware components of the Converged Security and Management Engine \(CSME\) firmware were reverted/i,
      ),
    ).toBeInTheDocument();
  });
});

describe('esetExcludedSet incident is correctly rendered', () => {
  const createIncident = (files: string[], processes: string[]) =>
    mockReportCardIncident({
      id: 'eset/excluded-set',
      args: { files, processes },
    }) as IssuesV1.EsetExcludedSet;

  test('no files - one process', () => {
    const incident = createIncident([], ['process1']);

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(/The process1 process was added/i),
    ).toBeInTheDocument();
  });

  test('no files - more processes', () => {
    const incident = createIncident([], ['process1', 'process2']);

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(/2 processes were added/i),
    ).toBeInTheDocument();
  });

  test('one file - no processes', () => {
    const incident = createIncident(['file1'], []);

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(/The file1 file was added/i),
    ).toBeInTheDocument();
  });

  test('one file - one process', () => {
    const incident = createIncident(['file1'], ['process1']);

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(
        /The file1 file and the process1 process were added/i,
      ),
    ).toBeInTheDocument();
  });

  test('one file - more processes', () => {
    const incident = createIncident(['file1'], ['process1', 'process2']);

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(
        /The file1 file and 2 processes were added/i,
      ),
    ).toBeInTheDocument();
  });

  test('more files - no processes', () => {
    const incident = createIncident(['file1', 'file2'], []);

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(/2 files were added/i),
    ).toBeInTheDocument();
  });

  test('more files - one process', () => {
    const incident = createIncident(['file1', 'file2'], ['process1']);

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(
        /2 files and the process1 process were added/i,
      ),
    ).toBeInTheDocument();
  });

  test('more files - more processes', () => {
    const incident = createIncident(['file1', 'file2'], ['process1', 'process2']);

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(/2 files and 2 processes were added/i),
    ).toBeInTheDocument();
  });
});

describe('UefiSecureBootKeys incident is correctly rendered', () => {
  test('issue #1577', () => {
    const incident = mockReportCardIncident({
      id: 'uefi/secure-boot-keys',
      aspect: 'firmware',
      args: {
        kek: [
          {
            fpr: '627afa0af19e20b75d3d4693ff4e2cc8fc9192e45a4a5a53acd2bccca46ca5f0',
            issuer: 'CN=Key Exchange Key,C=Key Exchange Key',
            not_after: '2026-08-13T13:32:48Z',
            not_before: '2021-08-13T13:32:48Z',
            subject: 'CN=Key Exchange Key,C=Key Exchange Key',
          },
          {
            fpr: '627afa0af19e20b75d3d4693ff4e2cc8fc9192e45a4a5a53acd2bccca46ca5f0',
            issuer: 'CN=Key Exchange Key,C=Key Exchange Key',
            not_after: '2026-08-13T13:32:48Z',
            not_before: '2021-08-13T13:32:48Z',
            subject: 'CN=Key Exchange Key,C=Key Exchange Key',
          },
        ],
      },
    }) as IssuesV1.UefiSecureBootKeys;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(
        /Changes to these keys are extremely rare and do not happen through firmware updates./i,
      ),
    ).toBeInTheDocument();
  });
});

describe('UefiSecureBootDbx incident is correctly rendered', () => {
  test('singular', () => {
    const incident = mockReportCardIncident({
      id: 'uefi/secure-boot-dbx',
      args: { fprs: ['fprs1'] },
    }) as IssuesV1.UefiSecureBootDbx;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(/During the modifications one entry was removed/i),
    ).toBeInTheDocument();
    expect(screen.getByText('its website').closest('a')).toHaveAttribute(
      'href',
      'https://uefi.org/revocationlistfile',
    );
  });

  test('plural', () => {
    const incident = mockReportCardIncident({
      id: 'uefi/secure-boot-dbx',
      args: { fprs: ['fprs1', 'fprs2'] },
    }) as IssuesV1.UefiSecureBootDbx;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(/During the modifications 2 entries were removed/i),
    ).toBeInTheDocument();
  });
});

describe('grubBootChanged incident is correctly rendered', () => {
  const beforeArgs = {
    kernel: 'kernelName',
    kernel_path: 'kernelPath',
    initrd: 'initRdName',
    initrd_path: 'initRdPath',
  };

  const createArgs = ({ kernel, kernel_path, initrd, initrd_path }) => ({
    before: beforeArgs,
    after: {
      kernel,
      kernel_path,
      initrd,
      initrd_path,
    },
  });

  const createIncident = (args: object) =>
    mockReportCardIncident({ id: 'grub/boot-changed', args }) as IssuesV1.GrubBootChanged;

  test('none changed', () => {
    const incident = createIncident(createArgs(beforeArgs));

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(
        /Linux command line set by the bootloader was changed./i,
      ),
    ).toBeInTheDocument();
    expect(
      within(screen.queryByRole('complementary')).queryByText(
        /Verify that the Linux command line was changed/i,
      ),
    ).toBeInTheDocument();
  });

  test('kernel changed', () => {
    const incident = createIncident(createArgs({ ...beforeArgs, kernel: 'anotherName' }));

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(
        /The linux kernel image booted by the bootloader/i,
      ),
    ).toBeInTheDocument();
    expect(
      within(screen.queryByRole('complementary')).queryByText(
        /Verify that the kernel image was changed/i,
      ),
    ).toBeInTheDocument();
  });

  test('init RAM disk changed', () => {
    const incident = createIncident(createArgs({ ...beforeArgs, initrd: 'anotherName' }));

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(
        /The initial RAM disk image booted by the bootloader/i,
      ),
    ).toBeInTheDocument();
    expect(
      within(screen.queryByRole('complementary')).queryByText(
        /Verify that the initial RAM disk image was changed/i,
      ),
    ).toBeInTheDocument();
  });

  test('init RAM disk changed', () => {
    const incident = createIncident(
      createArgs({ ...beforeArgs, kernel: 'anotherNameK', initrd: 'anotherNameI' }),
    );

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    expect(
      within(screen.queryByRole('contentinfo')).queryByText(
        /The linux kernel and initial RAM disk images booted by the bootloader/i,
      ),
    ).toBeInTheDocument();
    expect(
      within(screen.queryByRole('complementary')).queryByText(
        /Verify that the Linux kernel and initial RAM disk images were changed/i,
      ),
    ).toBeInTheDocument();
  });
});

describe('UefiBootAppSet incident is correctly rendered', () => {

  const app = (index: number) => ({
    path: `pathName-${index}`,
    before: `beforeName-${index}`,
    after: `afterName-${index}`,
  });

  const getByTextContent = (parent: HTMLElement, plainText: string) =>
    within(parent).queryByText((_content, node) => node.textContent === plainText)

  test('with app data in args, both before and after', () => {
    const incident = mockReportCardIncident({
      id: 'uefi/boot-app-set',
      args: { apps: [app(0), app(1)] },
    }) as IssuesV1.UefiBootAppSet;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const forensics = screen.queryByRole('switch');
    expect(forensics).toBeInTheDocument();
    expect(getByTextContent(forensics, 'Before:beforeName-0')).toBeInTheDocument()
    expect(getByTextContent(forensics, 'After:afterName-0')).toBeInTheDocument()
  });

  test('with app data in args, with after but NO before', () => {
    const appNewFile = [app(0), {...app(1), before: null}];
    const incident = mockReportCardIncident({
      id: 'uefi/boot-app-set',
      args: { apps: appNewFile },
    }) as IssuesV1.UefiBootAppSet;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const forensics = screen.queryByRole('switch');
    expect(forensics).toBeInTheDocument();
    expect(getByTextContent(forensics, 'New File:afterName-1')).toBeInTheDocument();
  });

  test('with app data in args, with before but NO after', () => {
    const appFileDeleted = [app(0), {...app(1), after: null}];
    const incident = mockReportCardIncident({
      id: 'uefi/boot-app-set',
      args: { apps: appFileDeleted },
    }) as IssuesV1.UefiBootAppSet;

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <StatusReportCardEntry incident={incident} />
      </I18nProvider>,
    );

    const forensics = screen.queryByRole('switch');
    expect(forensics).toBeInTheDocument();
    expect(getByTextContent(forensics, 'File Deleted:beforeName-1')).toBeInTheDocument();
  });
});
