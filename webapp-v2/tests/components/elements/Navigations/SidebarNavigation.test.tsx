/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';
import SidebarNavigation from 'components/elements/Navigations/SidebarNavigation';
import * as ConfigHook from 'hooks/config';
import * as SessionHook from 'hooks/useSession';
import { useRouter } from 'next/router';
import { mockConfig, mockSessionContext } from 'tests/mocks';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

describe('sidebar navigation', () => {
  const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;
  const mockUseConfig = jest.spyOn(ConfigHook, 'useConfig') as jest.Mock;

  mockUseSession.mockImplementation(() => mockSessionContext);
  mockUseConfig.mockImplementation(() => mockConfig);

  test('renders sidebar navigation correctly', async () => {
    (useRouter as jest.Mock).mockReturnValue({
      query: { did: '123' },
      pathname: '/dashboard',
    });

    render(<SidebarNavigation />);

    expect(screen.queryByText(/Dashboard/i)).toBeInTheDocument();
    expect(screen.queryByText(/My Organisation/i)).toBeInTheDocument();
    expect(screen.queryByText(/Admin/i)).not.toBeInTheDocument();
  });

  test('shows admin option only if user is admin', () => {
    const sessionWithAdmin = {
      ...mockSessionContext,
      session: {
        ...mockSessionContext.session,
        user: {
          ...mockSessionContext.session.user,
          role: 'admin',
        },
      },
    };
    mockUseSession.mockImplementationOnce(() => sessionWithAdmin);

    render(<SidebarNavigation />);

    expect(screen.queryByText(/My Organisation/i)).toBeInTheDocument();
    expect(screen.queryByText(/Admin/i)).toBeInTheDocument();
  });
});
