/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import DashboardLayout from 'components/layouts/dashboard';
import * as MembershipsHook from 'hooks/memberships';
import * as OrganisationsHook from 'hooks/organisations';
import * as SessionHook from 'hooks/useSession';
import * as UsersHook from 'hooks/users';
import { useRouter } from 'next/router';
import DashboardUsersAccount from 'pages/dashboard/users/account';
import SessionProvider from 'provider/SessionProvider';
import { toast } from 'react-toastify';
import {
  mockSessionContext,
  serializedMembership,
  serializedMemberships,
  serializedOrganisation,
  serializedSession,
} from 'tests/mocks';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const serializedUser = serializedSession.user;
const changePasswordResponse = { success: true, message: 'some message' };
const mutateResponse = {
  mutate: () => ({}),
  data: {},
  isLoading: false,
  isError: false,
};

const mockUseChangePasswordUser = jest.spyOn(UsersHook, 'useChangePasswordUser');
const mockUseDeleteUser = jest.spyOn(UsersHook, 'useDeleteUser');
const mockUseUpdateUser = jest.spyOn(UsersHook, 'useUpdateUser');
const mockUseDeleteOrganisation = jest.spyOn(OrganisationsHook, 'useDeleteOrganisation');
const mockUseMemberships = jest.spyOn(MembershipsHook, 'useMemberships');
const mockUseUpdateMembership = jest.spyOn(MembershipsHook, 'useUpdateMembership');
const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;

describe('organisations/users/account', () => {
  let mutateUser: jest.Mock<() => Promise<unknown>>;
  let mutatePasswordUser: jest.Mock<() => Promise<unknown>>;
  let mutateDeleteOrganisation: jest.Mock<() => Promise<unknown>>;
  const mockToastError = jest.spyOn(toast, 'error');
  const tabs = (): HTMLCollection => screen.queryByRole('tabpanel').firstElementChild?.children;

  beforeEach(() => {
    mutateUser = jest.fn(() => Promise.resolve(serializedUser));
    mutatePasswordUser = jest.fn(() => Promise.resolve(changePasswordResponse));
    mutateDeleteOrganisation = jest.fn(() => Promise.resolve(serializedOrganisation));

    mockUseSession.mockImplementation(() => ({
      ...mockSessionContext,
      session: {
        ...mockSessionContext.session,
        ...serializedSession,
      },
    }));

    mockUseRouter.mockReturnValue({
      pathname: '/dashboard/users/account',
      query: {},
    });

    mockUseUpdateUser.mockImplementation(() => ({
      ...mutateResponse,
      mutate: mutateUser,
      data: serializedUser,
    }));

    mockUseDeleteUser.mockImplementation(() => ({
      ...mutateResponse,
      mutate: mutateUser,
      data: serializedUser,
    }));

    mockUseChangePasswordUser.mockImplementation(() => ({
      ...mutateResponse,
      mutate: mutatePasswordUser,
      data: changePasswordResponse,
    }));

    mockUseDeleteOrganisation.mockImplementation(() => ({
      ...mutateResponse,
      mutate: mutateDeleteOrganisation,
      data: serializedOrganisation,
    }));

    mockUseMemberships.mockImplementation(() => ({
      ...mutateResponse,
      mutate: jest.fn(),
      data: serializedMemberships,
    }));
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe('Profile information tab', () => {
    const getElements = () => {
      const passwordInputs = screen.queryAllByPlaceholderText('******');

      return {
        profileInfoForm: screen.queryByRole('form'),
        tabs: screen.queryByRole('tabpanel').firstElementChild?.children,
        inputs: {
          nameInput: screen.queryByPlaceholderText('Name'),
          emailInput: screen.queryByPlaceholderText('Email'),
          currentPasswordInput: passwordInputs[0],
          newPasswordInput: passwordInputs[1],
          confirmPasswordInput: passwordInputs[2],
        },
        deleteAccountButton: screen.queryByText('Delete Account'),
        saveChangesButton: screen.queryByText('Save changes'),
      };
    };

    test('it renders the component correctly', () => {
      render(<DashboardUsersAccount />);

      const { profileInfoForm } = getElements();

      expect(screen.queryByText('Settings')).toBeInTheDocument();
      expect(within(profileInfoForm).queryByText('Profile information')).toBeInTheDocument();

      expect(within(profileInfoForm).queryByDisplayValue('Kai Michaelis')).toBeInTheDocument();
      expect(
        within(profileInfoForm).queryByDisplayValue('kai.michaelis@immu.ne'),
      ).toBeInTheDocument();
    });

    test('allows changing my name and email', async () => {
      render(
        <SessionProvider>
          <DashboardLayout>
            <DashboardUsersAccount />
          </DashboardLayout>
        </SessionProvider>,
      );

      const { inputs, saveChangesButton } = getElements();

      fireEvent.change(inputs.nameInput, { target: { value: 'New Name' } });
      fireEvent.change(inputs.emailInput, { target: { value: 'new@email.com' } });
      fireEvent.click(saveChangesButton);

      await waitFor(() => {
        expect(mutateUser).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedUser.id,
            name: 'New Name',
          }),
        );
      });
    });

    test('handles errors changing my name and email', async () => {
      const errorMutateUser = jest.fn(() =>
        Promise.resolve({ errors: [{ id: 'name', title: 'Test message' }] }),
      );

      mockUseUpdateUser.mockImplementation(() => ({
        mutate: errorMutateUser,
        data: serializedUser,
        isLoading: false,
        isError: true,
      }));

      render(
        <SessionProvider>
          <DashboardLayout>
            <DashboardUsersAccount />
          </DashboardLayout>
        </SessionProvider>,
      );

      const { inputs, saveChangesButton } = getElements();

      fireEvent.change(inputs.nameInput, { target: { value: 'New Name' } });
      fireEvent.change(inputs.emailInput, { target: { value: 'new@email.com' } });
      fireEvent.click(saveChangesButton);

      await waitFor(() => {
        expect(errorMutateUser).toHaveBeenCalledWith(
          expect.objectContaining({
            id: serializedUser.id,
            name: 'New Name',
            email: 'new@email.com',
          }),
        );
        expect(screen.queryByText('Test message')).toBeInTheDocument();
      });
    });

    test('handles password changes', async () => {
      render(
        <SessionProvider>
          <DashboardLayout>
            <DashboardUsersAccount />
          </DashboardLayout>
        </SessionProvider>,
      );

      const { inputs, saveChangesButton } = getElements();

      fireEvent.change(inputs.currentPasswordInput, { target: { value: '123456' } });
      fireEvent.change(inputs.newPasswordInput, { target: { value: 'abcdef' } });
      fireEvent.change(inputs.confirmPasswordInput, { target: { value: 'abcdef' } });
      fireEvent.click(saveChangesButton);

      await waitFor(() => {
        expect(mutatePasswordUser).toHaveBeenCalledWith(
          expect.objectContaining({
            currentPassword: '123456',
            password: 'abcdef',
          }),
        );
      });
    });

    test('handles error password changes', async () => {
      const errorMutateUser = jest.fn(() =>
        Promise.resolve({ errors: [{ id: 'current_password', title: 'Test message' }] }),
      );

      mockUseChangePasswordUser.mockImplementation(() => ({
        mutate: errorMutateUser,
        data: { ...changePasswordResponse, success: false },
        isLoading: false,
        isError: true,
      }));

      render(
        <SessionProvider>
          <DashboardLayout>
            <DashboardUsersAccount />
          </DashboardLayout>
        </SessionProvider>,
      );

      const { inputs, saveChangesButton } = getElements();

      fireEvent.change(inputs.currentPasswordInput, { target: { value: '123456' } });
      fireEvent.change(inputs.newPasswordInput, { target: { value: 'abcdef' } });
      fireEvent.change(inputs.confirmPasswordInput, { target: { value: 'abcdef' } });
      fireEvent.click(saveChangesButton);

      await waitFor(() => {
        expect(errorMutateUser).toHaveBeenCalledWith(
          expect.objectContaining({
            currentPassword: '123456',
            password: 'abcdef',
          }),
        );

        expect(mockToastError).toHaveBeenCalledWith('Test message');
      });
    });
  });

  describe('Organisations tab', () => {
    const getElements = () => {
      const orgsTable = screen.queryByRole('table');
      const orgsTableBody = within(orgsTable).queryAllByRole('rowgroup')[0]; //tbody

      return {
        orgsTable,
        orgsTableBody,
        orgRows: within(orgsTableBody).queryAllByRole('row'),
      };
    };

    test('displays my organisations info', () => {
      render(<DashboardUsersAccount />);

      expect(tabs().length).toBe(3);
      fireEvent.click(tabs()[1]);

      const { orgsTable } = getElements();

      expect(screen.queryAllByText('Organisations')[1]).toBeInTheDocument();
      expect(orgsTable).toBeInTheDocument();

      expect(within(orgsTable).queryByText("Kai's Organisation")).toBeInTheDocument();
      expect(within(orgsTable).queryByText("Testy's Organisation")).toBeInTheDocument();
      expect(within(orgsTable).queryByText('Payment reminder Organisation')).toBeInTheDocument();
    });

    test('can delete organisations', async () => {
      render(<DashboardUsersAccount />);

      fireEvent.click(tabs()[1]);

      const { orgRows } = getElements();
      expect(within(orgRows[0]).getByText("Kai's Organisation")).toBeInTheDocument();

      const firstOrgCells = within(orgRows[0]).queryAllByRole('cell');
      const firstOrgOptionsCell = firstOrgCells.slice(-1)[0];

      const options = firstOrgOptionsCell.firstElementChild?.children;
      expect(options.length).toBe(3);
      const deleteOption = options.item(0);

      waitFor(async () => {
        fireEvent.click(deleteOption);

        //it opens confirm modal
        const dialog = screen.getByRole('dialog');
        expect(dialog).toBeInTheDocument();

        const confirmDeleteInput = screen.findByLabelText('confirm-del');
        expect(confirmDeleteInput).toBeInTheDocument();

        fireEvent.change(await confirmDeleteInput, { target: { value: "Kai's Organisation" } });
        fireEvent.click(screen.queryByText('Delete'));

        expect(mutateDeleteOrganisation).toHaveBeenCalledWith({
          id: serializedOrganisation.id,
        });
      });
    });
  });

  describe('Notifications tab', () => {
    const getElements = () => {
      const orgsTable = screen.queryByRole('table');
      const orgsTableBody = within(orgsTable).queryAllByRole('rowgroup')[0]; //tbody

      return {
        orgsTable,
        orgsTableBody,
        orgRows: within(orgsTableBody).queryAllByRole('row'),
      };
    };

    test('can change email prefs', async () => {
      const mutateUpdateMembership = jest.fn(() =>
        Promise.resolve({
          ...serializedMembership,
          notifyInvoice: true,
        }),
      );

      mockUseUpdateMembership.mockImplementation(() => ({
        ...mutateResponse,
        data: serializedMembership,
        mutate: mutateUpdateMembership,
      }));

      render(<DashboardUsersAccount />);

      fireEvent.click(tabs()[2]);

      const { orgRows } = getElements();
      expect(within(orgRows[0]).getByText("Kai's Organisation")).toBeInTheDocument();

      const firstOrgCells = within(orgRows[0]).queryAllByRole('cell');
      const billingSwitch = within(firstOrgCells[1]).queryByRole('switch', { checked: false });

      expect(billingSwitch).toBeInTheDocument();

      fireEvent.click(billingSwitch);

      await waitFor(() => {
        expect(mutateUpdateMembership).toHaveBeenCalledWith({
          id: serializedMembership.id,
          membership: {
            notifyInvoice: true,
          },
        });
      });
    });
  });
});
