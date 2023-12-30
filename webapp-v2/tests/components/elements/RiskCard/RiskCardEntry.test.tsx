/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { render, screen, within } from '@testing-library/react';
import RiskCardEntry from 'components/elements/RiskCard/RiskCardEntry';
import * as IssuesV1 from 'generated/issuesv1';
import binarly from 'locales/en/binarly.json';
import common from 'locales/en/common.json';
import risks from 'locales/en/risks.json';
import I18nProvider from 'next-translate/I18nProvider';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

describe('send form to request password', () => {
  test('renders the correspondant description w/ translation', async () => {
    const issue = {
      id: 'fw/update',
      incident: false,
      aspect: 'firmware',
      args: {
        updates: [
          {
            name: 'updateName',
            current: '1.0',
            next: '2.0',
          },
          {
            name: 'updateName2',
            current: '1.3',
            next: '1.8',
          },
        ],
      },
    } as IssuesV1.FirmwareUpdate;

    render(
      <I18nProvider lang="en" namespaces={{ risks, binarly, common }}>
        <RiskCardEntry issue={issue} />
      </I18nProvider>,
    );

    const description = screen.queryByRole('contentinfo');

    expect(description).toBeInTheDocument();
    expect(
      within(description).queryByText(/Firmware updates are available for 2 components/i),
    ).toBeInTheDocument();
  });
});
