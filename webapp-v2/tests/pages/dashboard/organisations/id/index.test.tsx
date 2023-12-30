/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { loadStripe } from '@stripe/stripe-js';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import DashboardLayout from 'components/layouts/dashboard';
import { parse } from 'date-fns';
import * as ConfigHook from 'hooks/config';
import * as InvoicesHook from 'hooks/invoices';
import * as OrganisationsHook from 'hooks/organisations';
import * as SubscriptionsHook from 'hooks/subscriptions';
import * as SessionHook from 'hooks/useSession';
import { useRouter } from 'next/router';
import DashboardUsersOrganisation from 'pages/dashboard/organisations/[id]';
import { toast } from 'react-toastify';
import {
  mockConfig,
  mockSessionContext,
  mockStripeFn,
  serializedOrganisation,
  serializedSession,
  serializedSubscription,
  serializedSubscriptionInvoices,
} from 'tests/mocks';
import * as ApiUtils from 'utils/api';
import { SerializedSubscription } from 'utils/types';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

jest.mock('@stripe/stripe-js', () => ({
  loadStripe: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;
const mockUseSubscription = jest.spyOn(SubscriptionsHook, 'useSubscription');
const mockUseOrganisation = jest.spyOn(OrganisationsHook, 'useOrganisation');
const mockUpdateOrganisation = jest.spyOn(OrganisationsHook, 'useUpdateOrganisation');
const mockUseInvoices = jest.spyOn(InvoicesHook, 'useInvoices');
const mockUseConfig = jest.spyOn(ConfigHook, 'useConfig') as jest.Mock;
const mockRefreshSession = jest.spyOn(ApiUtils, 'refreshSession');
const mockLoadStripe = loadStripe as jest.Mock;

const mockToastSuccess = jest.spyOn(toast, 'success');

describe('organisations/show', () => {
  let mutateOrganisation: jest.Mock<() => Promise<unknown>>;
  let errorMutateOrganisation: jest.Mock<() => Promise<unknown>>;
  const tabs = (): HTMLCollection => screen.queryByRole('tabpanel').firstElementChild?.children;
  const saveChangesButton = (): HTMLElement => screen.queryByText('Save Organisation');

  beforeEach(() => {
    mutateOrganisation = jest.fn(() => Promise.resolve(serializedOrganisation));

    mockUseRouter.mockReturnValue({
      pathname: `/dashboard/organisations/${serializedOrganisation.id}`,
      query: { id: serializedOrganisation.id },
      replace: () => ({}),
    });

    mockLoadStripe.mockImplementation(() => mockStripeFn());

    mockUseSession.mockImplementation(() => mockSessionContext);
    mockUseConfig.mockImplementation(() => mockConfig);

    mockUseOrganisation.mockImplementation(() => ({
      data: serializedOrganisation,
      isLoading: false,
      isError: false,
    }));

    mockUseSubscription.mockImplementation(() => ({
      data: { ...serializedSubscription, currentDevicesAmount: 100 },
      isLoading: false,
      isError: true,
    }));

    mockUpdateOrganisation.mockImplementation(() => ({
      mutate: mutateOrganisation,
      data: serializedOrganisation,
      isError: false,
      isLoading: false,
    }));

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

  describe('Regular Information tab', () => {
    const getElements = () => {
      const regularInfoForm = screen.queryByRole('form');

      return {
        regularInfoForm,
        orgaNameInput: within(regularInfoForm).queryByPlaceholderText('Organisation Name'),
      };
    };

    test('renders correctly', () => {
      render(<DashboardUsersOrganisation />);

      const { regularInfoForm } = getElements();

      expect(tabs().length).toBe(3);
      expect(within(regularInfoForm).queryByText('Regular Information')).toBeInTheDocument();
    });

    test('allows changing org names', async () => {
      render(
        <DashboardLayout>
          <DashboardUsersOrganisation />
        </DashboardLayout>,
      );

      const { orgaNameInput } = getElements();

      expect(saveChangesButton()).not.toBeInTheDocument();

      fireEvent.change(orgaNameInput, { target: { value: 'New Orga Name' } });

      expect(saveChangesButton()).toBeInTheDocument();

      fireEvent.click(saveChangesButton());

      await waitFor(() => {
        expect(mutateOrganisation).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedOrganisation.id,
            name: 'New Orga Name',
          }),
        );
      });
    });

    test('handles server errors', async () => {
      errorMutateOrganisation = jest.fn(() =>
        Promise.resolve({
          errors: [{ id: 'address.postal_code', title: 'Test message for error' }],
        }),
      );

      mockUpdateOrganisation.mockImplementation(() => ({
        mutate: errorMutateOrganisation,
        data: serializedOrganisation,
        isError: false,
        isLoading: false,
      }));

      render(
        <DashboardLayout>
          <DashboardUsersOrganisation />
        </DashboardLayout>,
      );

      const { orgaNameInput } = getElements();

      expect(saveChangesButton()).not.toBeInTheDocument();

      fireEvent.change(orgaNameInput, { target: { value: 'New Orga Name' } });

      expect(saveChangesButton()).toBeInTheDocument();

      fireEvent.click(saveChangesButton());

      await waitFor(() => {
        expect(errorMutateOrganisation).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedOrganisation.id,
            name: 'New Orga Name',
          }),
        );

        expect(screen.queryByText('Test message for error')).toBeInTheDocument();
      });
    });
  });

  describe('Billing tab', () => {
    const getElements = () => {
      const invoicesTable = screen.queryByRole('table');
      const usersTableBody = within(invoicesTable).queryAllByRole('rowgroup')[0]; //tbody

      return {
        invoicesTable,
        usersTableBody,
        rows: within(usersTableBody).queryAllByRole('row'),
      };
    };

    beforeEach(() => {
      mockUseInvoices.mockImplementation(() => ({
        data: serializedSubscriptionInvoices,
        isLoading: false,
        isError: false,
      }));
    });

    test('list invoices and calls stripe', async () => {
      render(<DashboardUsersOrganisation />);

      fireEvent.click(tabs()[1]);

      const { invoicesTable, rows } = getElements();

      expect(screen.queryByText('Payment Details')).toBeInTheDocument();
      expect(invoicesTable).toBeInTheDocument;
      expect(rows.length).toBe(3);

      rows.forEach((row, index, rowList) => {
        if (index == 0) return true;

        const current = parse(row.firstElementChild.textContent, 'PPP', new Date());
        const previous = parse(rowList[index - 1].firstElementChild.textContent, 'PPP', new Date());

        expect(previous < current).toBe(true);
      });

      expect(mockLoadStripe).toHaveBeenCalled();
    });

    test('displays subscription details for orgs w/ devices', async () => {
      render(<DashboardUsersOrganisation />);

      fireEvent.click(tabs()[1]);

      expect(screen.queryByText(/billed monthly for 100 devices/i)).toBeInTheDocument();
      expect(screen.queryByText(/\(500€\)/i)).toBeInTheDocument();
      expect(
        screen.getByText((_, element) => element.textContent == 'Next bill: 515€'),
      ).toBeInTheDocument();
    });

    test('displays subscription details for orgs w/ partial free credits', async () => {
      mockUseSubscription.mockImplementation(() => ({
        data: {
          ...serializedSubscription,
          currentDevicesAmount: 10,
          billingDetails: {
            freeCredits: 4200,
          },
        } as SerializedSubscription,
        isLoading: false,
        isError: true,
      }));

      render(<DashboardUsersOrganisation />);

      fireEvent.click(tabs()[1]);

      expect(screen.queryByText(/next bill will be covered partially/i)).toBeInTheDocument();
      expect(screen.queryByText(/credit balance of 42€/i)).toBeInTheDocument();
      expect(
        screen.getByText((_, element) => element.textContent == 'Next bill: 65€'),
      ).toBeInTheDocument();
      expect(
        screen.getByText((_, element) => element.textContent == '42€ free credits remaining'),
      ).toBeInTheDocument();
    });

    test('displays subscription details for orgs w/ free credits', async () => {
      mockUseSubscription.mockImplementation(() => ({
        data: {
          ...serializedSubscription,
          currentDevicesAmount: 10,
          billingDetails: {
            freeCredits: 1004200,
          },
        } as SerializedSubscription,
        isLoading: false,
        isError: true,
      }));

      render(<DashboardUsersOrganisation />);

      fireEvent.click(tabs()[1]);

      expect(screen.queryByText(/credit balance of 10042€/i)).toBeInTheDocument();
      expect(
        screen.getByText((_, element) => element.textContent == 'Next bill: 65€'),
      ).toBeInTheDocument();
      expect(
        screen.getByText((_, element) => element.textContent == '10042€ free credits remaining'),
      ).toBeInTheDocument();
    });

    test('displays subscription details for orgs w/ base fee only', async () => {
      mockUseSubscription.mockImplementation(() => ({
        data: {
          ...serializedSubscription,
          currentDevicesAmount: 10,
          monthlyFeePerDevice: 0,
        } as SerializedSubscription,
        isLoading: false,
        isError: true,
      }));

      render(<DashboardUsersOrganisation />);

      fireEvent.click(tabs()[1]);

      expect(screen.queryByText(/is billed a monthly flat fee of 15€/i)).toBeInTheDocument();
      expect(
        screen.getByText((_, element) => element.textContent == 'Next bill: 15€'),
      ).toBeInTheDocument();
    });

    test('displays subscription details for orgs w/o fees', async () => {
      mockUseSubscription.mockImplementation(() => ({
        data: {
          ...serializedSubscription,
          currentDevicesAmount: 10,
          monthlyFeePerDevice: 0,
          monthlyBaseFee: 0,
        } as SerializedSubscription,
        isLoading: false,
        isError: true,
      }));

      render(<DashboardUsersOrganisation />);

      fireEvent.click(tabs()[1]);

      expect(screen.queryByText(/This organisation is free./i)).toBeInTheDocument();
    });

    test('displays subscription details for orgs w/ freeloader status', async () => {
      mockUseOrganisation.mockReturnValue({
        data: {
          ...serializedOrganisation,
          freeloader: true,
        },
        isLoading: false,
        isError: false,
      });

      render(<DashboardUsersOrganisation />);

      fireEvent.click(tabs()[1]);

      expect(screen.queryByText(/This organisation is free./i)).toBeInTheDocument();
    });

    test('displays subscription details for orgs w/ variable fee only', async () => {
      mockUseSubscription.mockImplementation(() => ({
        data: {
          ...serializedSubscription,
          currentDevicesAmount: 10,
          monthlyBaseFee: 0,
        } as SerializedSubscription,
        isLoading: false,
        isError: true,
      }));

      render(<DashboardUsersOrganisation />);

      fireEvent.click(tabs()[1]);

      expect(screen.queryByText(/i.e. 50€ per month./i)).toBeInTheDocument();
      expect(
        screen.getByText((_, element) => element.textContent == 'Next bill: 50€'),
      ).toBeInTheDocument();
    });
  });

  describe('Notifications', () => {
    const getElements = () => {
      const [splunkGroup, syslogGroup] = screen.queryAllByRole('group');

      return {
        splunk: {
          switch: within(splunkGroup).queryByRole('switch'),
          serverInput: within(splunkGroup).queryByPlaceholderText(/Splunk server/i),
          tokenInput: within(splunkGroup).queryByPlaceholderText(
            /Authentication token for Splunk/i,
          ),
          certCheckbox: within(splunkGroup).queryByLabelText(/accept all server certificates/i),
        },
        syslog: {
          switch: within(syslogGroup).queryByRole('switch'),
          serverInput: within(syslogGroup).queryByPlaceholderText(/Syslog server/i),
          portInput: within(syslogGroup).queryByPlaceholderText(/UDP Port/i),
        },
        submitButton: () => screen.queryByText('Save Organisation'),
      };
    };

    test('splunk can be enabled and configured', async () => {
      render(
        <DashboardLayout>
          <DashboardUsersOrganisation />
        </DashboardLayout>,
      );

      await waitFor(() => fireEvent.click(tabs()[2]));

      expect(screen.queryByText('Splunk')).toBeInTheDocument();
      const { splunk, submitButton } = getElements();

      fireEvent.change(splunk.serverInput, {
        target: { value: 'https://siem.example.com/splunk/hec/42' },
      });
      fireEvent.change(splunk.tokenInput, { target: { value: 'deadbeefcafebabe' } });
      fireEvent.click(splunk.certCheckbox);
      fireEvent.click(submitButton());

      await waitFor(() => {
        expect(mutateOrganisation).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedOrganisation.id,
            splunkAcceptAllServerCertificates: false,
            splunkAuthenticationToken: 'deadbeefcafebabe',
            splunkEventCollectorUrl: 'https://siem.example.com/splunk/hec/42',
          }),
        );

        expect(mockToastSuccess).toHaveBeenCalled();
      });
    });

    test('splunk can be disabled', async () => {
      render(
        <DashboardLayout>
          <DashboardUsersOrganisation />
        </DashboardLayout>,
      );

      await waitFor(() => fireEvent.click(tabs()[2]));

      const { splunk, submitButton } = getElements();

      fireEvent.click(splunk.switch);

      expect(splunk.serverInput).toBeDisabled();
      expect(splunk.tokenInput).toBeDisabled();
      fireEvent.click(submitButton());

      await waitFor(() => {
        expect(mutateOrganisation).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedOrganisation.id,
            splunkEnabled: false,
          }),
        );

        expect(mockToastSuccess).toHaveBeenCalled();
      });
    });

    test('syslog can be enabled and configured', async () => {
      render(
        <DashboardLayout>
          <DashboardUsersOrganisation />
        </DashboardLayout>,
      );

      await waitFor(() => fireEvent.click(tabs()[2]));

      expect(screen.queryByText('Syslog')).toBeInTheDocument();
      const { syslog, submitButton } = getElements();

      fireEvent.change(syslog.serverInput, { target: { value: 'siem.example.com' } });
      fireEvent.change(syslog.portInput, { target: { value: '1234' } });
      fireEvent.click(submitButton());

      await waitFor(() => {
        expect(mutateOrganisation).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedOrganisation.id,
            syslogHostnameOrAddress: 'siem.example.com',
            syslogUdpPort: '1234',
          }),
        );

        expect(mockToastSuccess).toHaveBeenCalled();
      });
    });

    test('syslog can be disabled', async () => {
      render(
        <DashboardLayout>
          <DashboardUsersOrganisation />
        </DashboardLayout>,
      );

      await waitFor(() => fireEvent.click(tabs()[2]));

      const { syslog, submitButton } = getElements();

      fireEvent.click(syslog.switch);

      expect(syslog.serverInput).toBeDisabled();
      expect(syslog.portInput).toBeDisabled();
      fireEvent.click(submitButton());

      await waitFor(() => {
        expect(mutateOrganisation).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedOrganisation.id,
            syslogEnabled: false,
          }),
        );

        expect(mockToastSuccess).toHaveBeenCalled();
      });
    });

    test('handles server errors', async () => {
      errorMutateOrganisation = jest.fn(() =>
        Promise.resolve({ errors: [{ id: 'syslogUdpPort', title: 'Test message for error' }] }),
      );

      mockUpdateOrganisation.mockImplementation(() => ({
        mutate: errorMutateOrganisation,
        data: serializedOrganisation,
        isError: false,
        isLoading: false,
      }));

      render(
        <DashboardLayout>
          <DashboardUsersOrganisation />
        </DashboardLayout>,
      );

      await waitFor(() => fireEvent.click(tabs()[2]));

      const { syslog, submitButton } = getElements();

      fireEvent.change(syslog.serverInput, { target: { value: 'siem.example.com' } });
      fireEvent.change(syslog.portInput, { target: { value: '1234' } });
      fireEvent.click(submitButton());

      await waitFor(() => {
        expect(errorMutateOrganisation).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedOrganisation.id,
            syslogEnabled: true,
          }),
        );

        expect(screen.queryByText('Test message for error')).toBeInTheDocument();
      });
    });
  });
});
