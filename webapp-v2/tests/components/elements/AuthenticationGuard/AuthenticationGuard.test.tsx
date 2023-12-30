/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { render, screen, waitFor } from '@testing-library/react';
import AuthenticationGuard from 'components/elements/AuthenticationGuard/AuthenticationGuard';
import * as DashboardHook from 'hooks/dashboard';
import { useRouter } from 'next/router';
import AdminUsers from 'pages/admin/users';
import DashboardIndex from 'pages/dashboard/index';
import { contextDefaultValues, SessionContext } from 'provider/SessionProvider';
import { toast } from 'react-toastify';
import { mockSessionContext, mockDashboardData } from 'tests/mocks';
import { ISession } from 'utils/session';
import { SerializedUser } from 'utils/types';

jest.mock('next/dist/client/router', () => ({
  useRouter: jest.fn(),
}));

jest.mock('echarts-for-react', () => () => {
  return <div data-testid="echart" />;
});

const mockUseRouter = useRouter as jest.Mock;
const mockUseDashboard = jest.spyOn(DashboardHook, 'useDashboard') as jest.Mock;

describe('authentication guard', () => {
  const mockRouterPush = jest.fn(() => Promise.resolve(true));
  const mockToastError = jest.spyOn(toast, 'error');
  const notAuthorizedText = 'You are currently not authorized to access this feature.';

  beforeEach(() => {
    mockUseDashboard.mockImplementation(() => ({
      data: mockDashboardData,
      loadData: jest.fn(),
      isLoading: false,
      isError: false,
    }));
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe('user is not admin', () => {
    const initSessionValues = {
      ...contextDefaultValues,
      isInitialized: true,
      session: mockSessionContext.session as ISession,
    };

    test('does not allow going to admin view, logs user out', async () => {
      mockUseRouter.mockReturnValue({
        pathname: '/admin/users',
        push: mockRouterPush,
      });

      render(
        <SessionContext.Provider value={initSessionValues}>
          <AuthenticationGuard>
            <AdminUsers />
          </AuthenticationGuard>
        </SessionContext.Provider>,
      );

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(notAuthorizedText);
        expect(mockRouterPush).toHaveBeenCalledWith('/logout');
      });
    });

    test('allows entering dashboard page', async () => {
      mockUseRouter.mockReturnValue({
        pathname: '/dashboard',
        push: mockRouterPush,
        prefetch: () => Promise.resolve(true),
      });

      render(
        <SessionContext.Provider value={initSessionValues}>
          <AuthenticationGuard>
            <DashboardIndex />
          </AuthenticationGuard>
        </SessionContext.Provider>,
      );

      await waitFor(() => {
        expect(mockToastError).not.toHaveBeenCalledWith(notAuthorizedText);
        expect(mockRouterPush).not.toHaveBeenCalledWith('/logout');

        expect(
          screen.queryAllByText(/dashboard:[incidents|risksRanking].*/i).length,
        ).toBeGreaterThan(2);
      });
    });
  });

  describe('user is admin', () => {
    const initSessionValues = {
      ...contextDefaultValues,
      isInitialized: true,
      session: {
        ...(mockSessionContext.session as ISession),
        user: { id: '1', role: 'admin' } as SerializedUser,
      },
    };

    test('allows entering admin page', async () => {
      mockUseRouter.mockReturnValue({
        pathname: '/admin/users',
        push: mockRouterPush,
        query: {},
        prefetch: () => Promise.resolve(true),
      });

      render(
        <SessionContext.Provider value={initSessionValues}>
          <AuthenticationGuard>
            <AdminUsers />
          </AuthenticationGuard>
        </SessionContext.Provider>,
      );

      await waitFor(() => {
        expect(mockToastError).not.toHaveBeenCalledWith(notAuthorizedText);
        expect(mockRouterPush).not.toHaveBeenCalledWith('/logout');
        expect(screen.queryByText('ADMIN')).toBeInTheDocument();
        expect(screen.queryByText('User Management')).toBeInTheDocument();
      });
    });

    test('allows entering dashboard page', async () => {
      mockUseRouter.mockReturnValue({
        pathname: '/dashboard',
        push: mockRouterPush,
        prefetch: () => Promise.resolve(true),
      });

      render(
        <SessionContext.Provider value={initSessionValues}>
          <AuthenticationGuard>
            <DashboardIndex />
          </AuthenticationGuard>
        </SessionContext.Provider>,
      );

      await waitFor(() => {
        expect(mockToastError).not.toHaveBeenCalledWith(notAuthorizedText);
        expect(mockRouterPush).not.toHaveBeenCalledWith('/logout');

        expect(
          screen.queryAllByText(/dashboard:[incidents|risksRanking].*/i).length,
        ).toBeGreaterThan(2);
      });
    });
  });
});
