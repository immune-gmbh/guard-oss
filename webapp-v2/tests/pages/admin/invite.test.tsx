/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import * as OrganisationsHook from 'hooks/organisations';
import * as UsersHook from 'hooks/users';
import { useRouter } from 'next/router';
import AdminDashboardUsersInvite from 'pages/admin/invite';
import { serializedOrganisation, serializedOrganisations } from 'tests/mocks';
import { SerializedUser, SerializedMembership } from 'utils/types';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const mockUseOrganisation = jest.spyOn(OrganisationsHook, 'useOrganisation');
const mockUseOrganisations = jest.spyOn(OrganisationsHook, 'useOrganisations');
const mockInviteUser = jest.spyOn(UsersHook, 'useInviteUser');

const getElements = (index = 0) => {
  const rows = screen.queryAllByTestId(/row-[0-9]+/);
  const [orgSelect, roleSelect] = within(rows[index]).queryAllByRole('listbox');

  return {
    inviteUserForm: screen.queryByRole('form'),
    inviteUserButton: screen.queryAllByText('Invite User')[1],
    addUserButton: screen.queryByText(/\+ Add user/i),
    inputs: {
      nameInput: within(rows[index]).queryByPlaceholderText('Name'),
      emailInput: within(rows[index]).queryByPlaceholderText('Email'),
    },
    rowsLength: rows.length,
    orgSelect,
    roleSelect,
  };
};

const newUser = {
  name: 'New',
  email: 'new@user.com',
  invited: true,
  id: 'new-user-id',
  role: 'user',
} as SerializedUser;

const newMembership = {
  id: 'new-membership-id',
  user: newUser,
  role: 'owner',
  notifyDeviceUpdate: false,
  notifyInvoice: false,
} as SerializedMembership;

const inviteUserResponse = {
  message: 'Invitation has been sent.',
  success: true,
  user: newUser,
  membership: newMembership,
};

describe('admin/invite', () => {
  let mutateMemberships: jest.Mock<() => Promise<unknown>>;

  beforeEach(() => {
    mutateMemberships = jest.fn(() => Promise.resolve(inviteUserResponse));

    mockUseRouter.mockReturnValue({
      pathname: '/admin/invite',
      query: {},
    });

    mockUseOrganisation.mockImplementation(() => ({
      data: serializedOrganisation,
      isLoading: false,
      isError: false,
    }));

    mockUseOrganisations.mockImplementation(() => ({
      data: serializedOrganisations,
      isLoading: false,
      isError: false,
    }));

    mockInviteUser.mockImplementation(() => ({
      mutate: mutateMemberships,
      data: inviteUserResponse,
      isError: false,
      isLoading: false,
    }));
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('component renders correctly', () => {
    render(<AdminDashboardUsersInvite />);

    const { orgSelect, roleSelect, inviteUserForm } = getElements();

    expect(screen.queryAllByText('Invite User').length).toBe(2);
    expect(inviteUserForm).toBeInTheDocument();
    expect(orgSelect).toBeInTheDocument();
    expect(roleSelect).toBeInTheDocument();
  });

  test('appends an extra row for adding multiple users', () => {
    render(<AdminDashboardUsersInvite />);

    const { addUserButton, inputs } = getElements();

    expect(inputs.nameInput).toBeInTheDocument();
    expect(inputs.emailInput).toBeInTheDocument();
    fireEvent.click(addUserButton);

    const { inputs: newlyAddedInputs, rowsLength } = getElements(1);
    expect(newlyAddedInputs.nameInput).toBeInTheDocument();
    expect(newlyAddedInputs.emailInput).toBeInTheDocument();
    expect(rowsLength).toBe(2);
  });

  test('invites user successfully', async () => {
    render(<AdminDashboardUsersInvite />);

    const { inputs, orgSelect, roleSelect, inviteUserButton } = getElements();

    fireEvent.change(inputs.nameInput, { target: { value: 'New' } });
    fireEvent.change(inputs.emailInput, { target: { value: 'new@user.com' } });
    fireEvent.change(orgSelect, { target: { value: serializedOrganisations[0].id } });
    fireEvent.change(roleSelect, { target: { value: 'owner' } });

    fireEvent.click(inviteUserButton);

    await waitFor(() => {
      expect(mutateMemberships).toHaveBeenCalledWith({
        membership: expect.objectContaining({
          name: newUser.name,
          email: newUser.email,
          organisation_id: serializedOrganisations[0].id,
        }),
      });

      expect(screen.queryByText('Invitation has been sent.')).toBeInTheDocument();
    });
  });
});
