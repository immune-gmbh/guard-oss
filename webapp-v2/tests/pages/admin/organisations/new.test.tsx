/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import * as OrganisationsHook from 'hooks/organisations';
import * as SubscriptionHook from 'hooks/subscriptions';
import * as SessionHook from 'hooks/useSession';
import { useRouter } from 'next/router';
import AdminNewOrganisation from 'pages/admin/organisations/new';
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
  let mutateOrganisation: jest.Mock<() => Promise<unknown>>;
  let routerPush: jest.Mock;

  beforeEach(() => {
    mutateOrganisation = jest.fn(() => Promise.resolve(serializedOrganisation));
    routerPush = jest.fn();

    mockUseSession.mockImplementation(() => mockSessionContext);

    mockUseRouter.mockReturnValue({
      push: routerPush,
      pathname: '/organisations/new',
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
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('it renders the component correctly', () => {
    render(<AdminNewOrganisation />);

    const { newOrgaForm } = getElements();

    expect(screen.queryByText('New Organisation')).toBeInTheDocument();
    expect(newOrgaForm).toBeInTheDocument();
    expect(within(newOrgaForm).queryByText('Regular Information')).toBeInTheDocument();
  });

  test('does not create new org if there are missing fields', async () => {
    render(<AdminNewOrganisation />);

    const { newOrgaForm, inputs, createOrgButton } = getElements();

    expect(createOrgButton()).not.toBeInTheDocument();

    fireEvent.change(inputs.nameInput, { target: { value: 'New Orga Name' } });

    expect(createOrgButton()).toBeInTheDocument();

    fireEvent.click(createOrgButton());

    await waitFor(() => {
      expect(mutateOrganisation).not.toHaveBeenCalled();
      const missingFields = within(newOrgaForm).queryAllByText('Field is required');

      expect(missingFields.length).toBeGreaterThanOrEqual(4);
    });
  });

  test('creates a new organisation successfully', async () => {
    render(<AdminNewOrganisation />);

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
      expect(routerPush).toHaveBeenCalledWith({
        pathname: '/admin/organisations/[id]/users',
        query: { id: serializedOrganisation.id },
      });
    });
  });
});
