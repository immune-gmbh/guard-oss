/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import * as OrganisationsHook from 'hooks/organisations';
import * as SubscriptionHook from 'hooks/subscriptions';
import * as SessionHook from 'hooks/useSession';
import { useRouter } from 'next/router';
import AdminOrganisation from 'pages/admin/organisations/[id]/index';
import {
  mockSessionContext,
  mockSubscription,
  serializedOrganisation,
  serializedSession,
} from 'tests/mocks';
import * as ApiUtils from 'utils/api';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const mockUseSubscription = jest.spyOn(SubscriptionHook, 'useSubscription');
const mockUseOrganisation = jest.spyOn(OrganisationsHook, 'useOrganisation');
const mockRefreshSession = jest.spyOn(ApiUtils, 'refreshSession');
const mockUpdateOrganisation = jest.spyOn(OrganisationsHook, 'useUpdateOrganisation');
const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;

describe('organisations/index', () => {
  let mutateOrganisation: jest.Mock<() => Promise<unknown>>;

  beforeEach(() => {
    mutateOrganisation = jest.fn(() => Promise.resolve(serializedOrganisation));

    mockUseSession.mockImplementation(() => mockSessionContext);

    mockUseRouter.mockReturnValue({
      pathname: `/organisations/${serializedOrganisation.id}`,
      query: {
        id: serializedOrganisation.id,
        edit: '1',
      },
    });

    mockUseOrganisation.mockImplementation(() => ({
      data: serializedOrganisation,
      isLoading: false,
      isError: false,
    }));

    mockUseSubscription.mockImplementation(() => ({
      data: mockSubscription,
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

  test('it renders the component correctly', () => {
    render(<AdminOrganisation />);

    const regularInfoForm = screen.queryByRole('form');

    expect(screen.queryByText("Kai's Organisation")).toBeInTheDocument();
    expect(regularInfoForm).toBeInTheDocument();
    expect(within(regularInfoForm).queryByText('Regular Information')).toBeInTheDocument();
  });

  test('can update name from form', async () => {
    render(<AdminOrganisation />);

    const regularInfoForm = screen.queryByRole('form');
    const orgaNameInput = within(regularInfoForm).queryByPlaceholderText('Organisation Name');
    const saveChangesButton = () => screen.queryByText('Save Organisation');

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
      expect(mockRefreshSession).toHaveBeenCalled();
    });
  });
});
