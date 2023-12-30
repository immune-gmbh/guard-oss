/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import * as MembershipsHook from 'hooks/memberships';
import * as OrganisationsHook from 'hooks/organisations';
import * as SessionHook from 'hooks/useSession';
import { useRouter } from 'next/router';
import AdminOrganisationUsers from 'pages/admin/organisations/[id]/users';
import { mockSessionContext, serializedOrganisationWithUsers } from 'tests/mocks';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;
const mockUseOrganisation = jest.spyOn(OrganisationsHook, 'useOrganisation');
const mockUpdateMembership = jest.spyOn(MembershipsHook, 'useUpdateMembership');

const firstMembership = serializedOrganisationWithUsers.memberships[0];

const getTableElements = () => {
  const usersTable = screen.queryByRole('table');
  const usersTableBody = within(usersTable).queryAllByRole('rowgroup')[1]; //tbody

  return {
    usersTable,
    usersTableBody,
    userRows: within(usersTableBody).queryAllByRole('row'),
  };
};

describe('admin/organisation/[id]/users', () => {
  let mutateMembership: jest.Mock<() => Promise<unknown>>;

  beforeEach(() => {
    mutateMembership = jest.fn(() => Promise.resolve(firstMembership));

    mockUseRouter.mockReturnValue({
      pathname: '/admin/organisations/[id]/users',
      query: { id: serializedOrganisationWithUsers.id },
    });

    mockUseSession.mockImplementation(() => ({
      ...mockSessionContext,
      session: {
        ...mockSessionContext.session,
        user: serializedOrganisationWithUsers.users[0],
      },
    }));

    mockUseOrganisation.mockImplementation(() => ({
      data: serializedOrganisationWithUsers,
      isLoading: false,
      isError: undefined,
    }));

    mockUpdateMembership.mockImplementation(() => ({
      mutate: mutateMembership,
      data: serializedOrganisationWithUsers.memberships[0],
      isError: false,
      isLoading: false,
    }));
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('renders the component correctly', async () => {
    render(<AdminOrganisationUsers />);

    const { userRows } = getTableElements();

    expect(screen.queryByText("User Management - Kai's Organisation")).toBeInTheDocument();

    expect(userRows.length).toBe(3);
    expect(within(userRows[0]).getByText('Kai Michaelis')).toBeInTheDocument();
    expect(within(userRows[0]).getByText('Owner')).toBeInTheDocument();
  });

  test('allows changing user role from table', async () => {
    render(<AdminOrganisationUsers />);

    const { userRows } = getTableElements();
    expect(within(userRows[0]).getByText('Owner')).toBeInTheDocument();

    const firstUserCells = within(userRows[0]).queryAllByRole('cell');
    const firstUserOptionsCell = firstUserCells.slice(-1)[0];

    const options = firstUserOptionsCell.firstElementChild?.children;
    expect(options.length).toBe(1);

    const editOption = options.item(0);
    fireEvent.click(editOption);

    // options are now check or cancel
    const confirmOptions = firstUserOptionsCell.firstElementChild?.children;
    expect(confirmOptions.length).toBe(2);

    const roleSelect = firstUserCells[2].firstElementChild;
    fireEvent.change(roleSelect, { target: { value: 'admin' } });
    fireEvent.blur(roleSelect);

    await act(async () => {
      fireEvent.click(confirmOptions[0]);

      await waitFor(() => {
        expect(mutateMembership).toHaveBeenCalled();
        expect(within(userRows[0]).getByText('Admin')).toBeInTheDocument();
      });
    });
  });
});
