/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { act, render, screen } from '@testing-library/react';
import * as ConfigHook from 'hooks/config';
import * as DashboardHook from 'hooks/dashboard';
import * as SessionHook from 'hooks/useSession';
import I18nProvider from 'next-translate/I18nProvider';
import { mockConfig, mockDashboardData, mockSessionContext } from 'tests/mocks';

import common from 'locales/en/common.json';
import dashboard from 'locales/en/dashboard.json';
import risks from 'locales/en/risks.json';
import DashboardIndex from 'pages/dashboard/index';

jest.mock('echarts-for-react', () => () => {
  return <div data-testid="echart" />;
});

describe('dashboard index', () => {
  const mockUseDashboard = jest.spyOn(DashboardHook, 'useDashboard') as jest.Mock;
  const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;
  const mockUseConfig = jest.spyOn(ConfigHook, 'useConfig') as jest.Mock;

  mockUseDashboard.mockImplementation(() => ({
    data: mockDashboardData,
    loadData: jest.fn(),
    isLoading: false,
    isError: false,
  }));

  mockUseSession.mockImplementation(() => mockSessionContext);
  mockUseConfig.mockImplementation(() => mockConfig);

  test('renders dashboard correctly', async () => {
    render(
      <I18nProvider lang="en" namespaces={{ dashboard, common, risks }}>
        <DashboardIndex />
      </I18nProvider>,
    );

    expect(screen.getByText(/Top Risks/i)).toBeInTheDocument();
    expect(screen.getByText(/Device Status/i)).toBeInTheDocument();
    expect(screen.queryByRole('banner')).toBeInTheDocument();
    expect(screen.getByText(/10 incidents found in 30 devices./i)).toBeInTheDocument();

    await act(async () => {
      await new Promise((r) => setTimeout(r, 600));
      expect(screen.getAllByTestId('echart').length).toBe(2);
    });
  });

  test('does not show incidents banner if no incidents', () => {
    mockUseDashboard.mockImplementationOnce(() => ({
      data: { ...mockDashboardData, incidents: { count: 0, devices: 20 } },
      loadData: jest.fn(),
      isLoading: false,
      isError: false,
    }));

    render(
      <I18nProvider lang="en" namespaces={{ dashboard, common, risks }}>
        <DashboardIndex />
      </I18nProvider>,
    );

    expect(screen.queryByRole('banner')).not.toBeInTheDocument();
  });
});
