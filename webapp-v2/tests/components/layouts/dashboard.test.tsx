/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import DashboardLayout from 'components/layouts/dashboard';
import common from 'locales/en/common.json';
import dashboard from 'locales/en/dashboard.json';
import risks from 'locales/en/risks.json';
import I18nProvider from 'next-translate/I18nProvider';
import { useRouter } from 'next/router';
import DashboardIndex from 'pages/dashboard/index';
import SessionProvider from 'provider/SessionProvider';
import { serializedSession } from 'tests/mocks';
import * as ApiUtils from 'utils/api';
import * as SessionUtils from 'utils/session';

jest.mock('next/dist/client/router', () => ({
  useRouter: jest.fn(),
}));

const mockFetchSession = jest.spyOn(SessionUtils, 'fetchSession');
const mockRefreshSession = jest.spyOn(ApiUtils, 'refreshSession');
const mockUseRouter = useRouter as jest.Mock;

describe('changes organisation', () => {
  let organisationSelect: HTMLElement;
  let firstOtherOrganisation: HTMLElement;
  let organisationSelectButton: HTMLButtonElement;

  beforeEach(() => {
    mockFetchSession.mockImplementation(() =>
      Promise.resolve({
        ...serializedSession,
        currentMembership: serializedSession.memberships[0],
      }),
    );

    mockRefreshSession.mockImplementation((setSessionFn, membership) => {
      return Promise.resolve(
        setSessionFn({
          ...serializedSession,
          currentMembership: membership,
        }),
      );
    });
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('can change organisation from dropdown', async () => {
    mockUseRouter.mockReturnValue({
      query: {},
      pathname: '/dashboard',
      prefetch: () => Promise.resolve(true),
    });

    render(
      <SessionProvider>
        <I18nProvider lang="en" namespaces={{ dashboard, common, risks }}>
          <DashboardLayout>
            <DashboardIndex />
          </DashboardLayout>
        </I18nProvider>
      </SessionProvider>
    );

    await waitFor(() => {
      organisationSelect = screen.queryByTestId('org-select-menu');
      organisationSelectButton = within(organisationSelect).queryByRole('button');
      expect(
        within(organisationSelectButton).queryByText(/Kai's Organisation/i),
      ).toBeInTheDocument();
    });

    fireEvent.click(organisationSelectButton);

    await waitFor(() => {
      firstOtherOrganisation = within(organisationSelect).queryAllByRole('menuitem')[0];
      expect(firstOtherOrganisation).toBeInTheDocument();
    });

    fireEvent.click(firstOtherOrganisation);

    await waitFor(() => {
      expect(mockRefreshSession).toHaveBeenCalledWith(
        expect.any(Function),
        serializedSession.memberships[1],
      );
    });

    expect(
      within(organisationSelectButton).queryByText(/Testy's Organisation/i),
    ).toBeInTheDocument();
  });

  test('handles organisation query string parameter', async () => {
    mockUseRouter.mockReturnValue({
      query: { organisation: '3ceecfd9-1f59-53a2-b2ff-636900e70668' },
      pathname: '/dashboard',
      prefetch: () => Promise.resolve(true),
    });

    render(
      <SessionProvider>
        <I18nProvider lang="en" namespaces={{ dashboard, common, risks }}>
          <DashboardLayout>
            <DashboardIndex />
          </DashboardLayout>
        </I18nProvider>
      </SessionProvider>
    );

    await waitFor(() => {
      organisationSelect = screen.queryByTestId('org-select-menu');
      organisationSelectButton = within(organisationSelect).queryByRole('button');
      expect(mockRefreshSession).toHaveBeenCalledWith(
        expect.any(Function),
        serializedSession.memberships[1],
      );
    });

    expect(
      within(organisationSelectButton).queryByText(/Testy's Organisation/i),
    ).toBeInTheDocument();
  });

  test('handles invalid organisation query string parameter', async () => {
    mockUseRouter.mockReturnValue({
      query: { organisation: 'non-existing-org-id' },
      pathname: '/dashboard',
      prefetch: () => Promise.resolve(true),
    });

    render(
      <SessionProvider>
        <I18nProvider lang="en" namespaces={{ dashboard, common, risks }}>
          <DashboardLayout>
            <DashboardIndex />
          </DashboardLayout>
        </I18nProvider>
      </SessionProvider>
    );

    await waitFor(() => {
      organisationSelect = screen.queryByTestId('org-select-menu');
      organisationSelectButton = within(organisationSelect).queryByRole('button');
      expect(mockRefreshSession).not.toHaveBeenCalled();
    });

    expect(within(organisationSelectButton).queryByText(/Kai's Organisation/i)).toBeInTheDocument();
  });
});
