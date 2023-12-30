/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import DashboardLayout from 'components/layouts/dashboard';
import NextJsRoutes from 'generated/NextJsRoutes';
import * as OrganisationsHook from 'hooks/organisations';
import * as SubscriptionHook from 'hooks/subscriptions';
import * as SessionHook from 'hooks/useSession';
import { useRouter } from 'next/router';
import DashboardUsersNewOrganisation from 'pages/dashboard/organisations/new';
import { toast } from 'react-toastify';
import { mockSessionContext, mockSubscription, serializedOrganisation } from 'tests/mocks';
import { SerializedOrganisation } from 'utils/types';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const mockUseSubscription = jest.spyOn(SubscriptionHook, 'useSubscription');
const mockUseOrganisation = jest.spyOn(OrganisationsHook, 'useOrganisation');
const mockCreateOrganisation = jest.spyOn(OrganisationsHook, 'useCreateOrganisation');
const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;
const mockToastSuccess = jest.spyOn(toast, 'success');

const getElements = () => ({
  newOrgaForm: screen.queryByRole('form'),
  createOrgButton: () => screen.queryByText('Create Organisation'),
  inputs: {
    nameInput: screen.queryByPlaceholderText('Organisation Name'),
    invoiceNameInput: screen.queryByPlaceholderText('Name'),
    streetInput: screen.queryByPlaceholderText('Street & Number'),
    postalCodeInput: screen.queryByPlaceholderText('Postal Code'),
    cityInput: screen.queryByPlaceholderText('City'),
  },
  countrySelect: screen.queryByPlaceholderText('Country'),
});

describe('organisations/new', () => {
  const { location } = window;
  let mutateOrganisation: jest.Mock<() => Promise<unknown>>;
  let errorMutateOrganisation: jest.Mock<() => Promise<unknown>>;
  let mockLocationAssign: jest.Mock;

  beforeEach(() => {
    mutateOrganisation = jest.fn(() => Promise.resolve(serializedOrganisation));
    mockLocationAssign = jest.fn();

    mockUseSession.mockImplementation(() => mockSessionContext);

    mockUseRouter.mockReturnValue({
      pathname: '/dashboard/organisations/new',
      query: {
        id: null,
      },
    });

    mockUseOrganisation.mockReturnValue({
      data: {} as SerializedOrganisation,
      isLoading: false,
      isError: false,
    });

    mockUseSubscription.mockImplementation(() => ({
      data: mockSubscription,
      isLoading: false,
      isError: true,
    }));

    mockCreateOrganisation.mockImplementation(() => ({
      mutate: mutateOrganisation,
      data: serializedOrganisation,
      isError: false,
      isLoading: false,
    }));

    delete window.location;
    window.location = {...location, assign: mockLocationAssign };
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('it renders the component correctly', () => {
    render(<DashboardUsersNewOrganisation />);

    const { newOrgaForm } = getElements();

    expect(screen.queryByText('New Organisation')).toBeInTheDocument();
    expect(newOrgaForm).toBeInTheDocument();
    expect(within(newOrgaForm).queryByText('Regular Information')).toBeInTheDocument();
  });

  test('creates a new organisation successfully', async () => {
    render(
      <DashboardLayout>
        <DashboardUsersNewOrganisation />
      </DashboardLayout>
    );

    const { inputs, countrySelect, createOrgButton } = getElements();

    expect(createOrgButton()).not.toBeInTheDocument();

    fireEvent.change(inputs.nameInput, { target: { value: 'New Orga Name' } });
    fireEvent.change(inputs.invoiceNameInput, { target: { value: 'Testy McTest' } });
    fireEvent.change(inputs.streetInput, { target: { value: 'Teststr. 11' } });
    fireEvent.change(inputs.postalCodeInput, { target: { value: '12345' } });
    fireEvent.change(inputs.cityInput, { target: { value: 'Metropolis' } });
    fireEvent.change(countrySelect, { target: { value: 'Aruba' } });

    expect(createOrgButton()).toBeInTheDocument();

    fireEvent.click(createOrgButton());

    await waitFor(() => {
      expect(mutateOrganisation).toHaveBeenCalledWith(
        expect.objectContaining({
          id: null,
          name: 'New Orga Name',
          invoiceName: 'Testy McTest',
          address: expect.objectContaining({
            country: 'Aruba',
          }),
        }),
      );
      expect(mockToastSuccess).toHaveBeenCalledWith("Successfully created new organisation");
      expect(mockLocationAssign).toHaveBeenCalledWith(NextJsRoutes.dashboardIndexPath);
    });
  });

  test('handles server errors', async () => {
    errorMutateOrganisation = jest.fn(() => Promise.resolve({ errors: [{ id: 'invoice_name', title: 'Test message' }] }));

    mockCreateOrganisation.mockImplementation(() => ({
      mutate: errorMutateOrganisation,
      data: serializedOrganisation,
      isError: false,
      isLoading: false,
    }));

    render(
      <DashboardLayout>
        <DashboardUsersNewOrganisation />
      </DashboardLayout>
    );

    const { inputs, countrySelect, createOrgButton } = getElements();

    expect(createOrgButton()).not.toBeInTheDocument();

    fireEvent.change(inputs.nameInput, { target: { value: 'New Orga Name' } });
    fireEvent.change(inputs.invoiceNameInput, { target: { value: 'Testy McTest' } });
    fireEvent.change(inputs.streetInput, { target: { value: 'Teststr. 11' } });
    fireEvent.change(inputs.postalCodeInput, { target: { value: '12345' } });
    fireEvent.change(inputs.cityInput, { target: { value: 'Metropolis' } });
    fireEvent.change(countrySelect, { target: { value: 'Aruba' } });

    expect(createOrgButton()).toBeInTheDocument();

    fireEvent.click(createOrgButton());

    await waitFor(() => {
      expect(errorMutateOrganisation).toHaveBeenCalledWith(
        expect.objectContaining({
          id: null,
          name: 'New Orga Name',
          invoiceName: 'Testy McTest',
          address: expect.objectContaining({
            country: 'Aruba',
          }),
        }),
      );
      expect(screen.queryByText("Test message")).toBeInTheDocument();
      expect(mockLocationAssign).not.toHaveBeenCalledWith(NextJsRoutes.dashboardIndexPath);
    });
  });
});
