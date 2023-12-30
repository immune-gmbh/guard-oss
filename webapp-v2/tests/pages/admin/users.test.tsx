/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { render, screen, within } from '@testing-library/react';
import * as SessionHook from 'hooks/useSession';
import * as UsersHook from 'hooks/users';
import { useRouter } from 'next/router';
import AdminUsers from 'pages/admin/users';
import { mockSessionContext, serializedUsersWithOrgs } from 'tests/mocks';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;

const getTableElements = () => {
  const usersTable = screen.queryByRole('table');
  const usersTableBody = within(usersTable).queryAllByRole('rowgroup')[1]; //tbody

  return {
    usersTable,
    usersTableBody,
    userRows: within(usersTableBody).queryAllByRole('row'),
  };
};

describe('admin/users', () => {
  beforeEach(() => {
    mockUseRouter.mockReturnValue({
      pathname: '/admin/users',
      query: {},
    });

    mockUseSession.mockImplementation(() => mockSessionContext);
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('renders the component correctly, when still loading', async () => {
    jest.spyOn(UsersHook, 'useUsers').mockImplementationOnce(() => ({
      data: [],
      isLoading: true,
      isError: undefined,
    }));

    render(<AdminUsers />);

    expect(screen.queryByText('User Management')).toBeInTheDocument();
    expect(screen.queryByRole('status')).toBeInTheDocument();
  });

  test('renders the component correctly, when no loading anymore', async () => {
    jest.spyOn(UsersHook, 'useUsers').mockImplementationOnce(() => ({
      data: serializedUsersWithOrgs,
      isLoading: false,
      isError: undefined,
    }));

    render(<AdminUsers />);

    const { userRows } = getTableElements();

    expect(screen.queryByText('User Management')).toBeInTheDocument();
    expect(screen.queryByRole('status')).not.toBeInTheDocument();

    expect(userRows.length).toBe(3);
    expect(within(userRows[0]).getByText('Kai Michaelis')).toBeInTheDocument();
    expect(
      within(userRows[0]).getByText(
        "Kai's Organisation, Testy's Organisation, Payment reminder Organisation, New Orga",
      ),
    ).toBeInTheDocument();
  });
});
