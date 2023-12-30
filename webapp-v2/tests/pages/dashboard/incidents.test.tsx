/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { within, render, screen } from '@testing-library/react';
import * as IncidentsHook from 'hooks/issues';
import common from 'locales/en/common.json';
import dashboard from 'locales/en/dashboard.json';
import incidents from 'locales/en/incidents.json';
import utils from 'locales/en/utils.json';
import I18nProvider from 'next-translate/I18nProvider';
import DashboardIncidents from 'pages/dashboard/incidents';
import { mockIncidentList } from 'tests/mocks';

jest.mock('echarts-for-react', () => () => {
  return <div data-testid="echart" />;
});

const mockUseIncidents = jest.spyOn(IncidentsHook, 'useIncidents') as jest.Mock;

describe('incidents list view', () => {
  test('renders incidents view correctly', () => {
    mockUseIncidents.mockImplementation(() => ({
      data: mockIncidentList,
      isLoading: false,
    }));

    render(
      <I18nProvider lang="en" namespaces={{ dashboard, incidents, common, utils }}>
        <DashboardIncidents />
      </I18nProvider>,
    );

    const view = screen.getByRole('main');
    expect(within(view).queryByRole('table')).toBeInTheDocument();
    expect(within(view).queryByTestId('echart')).toBeInTheDocument();
  });

  test('renders incidents view on empty status', () => {
    mockUseIncidents.mockImplementation(() => ({
      data: [],
      isLoading: false,
    }));

    render(
      <I18nProvider lang="en" namespaces={{ dashboard, incidents, common, utils }}>
        <DashboardIncidents />
      </I18nProvider>,
    );

    const view = screen.getByRole('main');
    expect(within(view).queryByRole('table')).not.toBeInTheDocument();
    expect(within(view).queryByTestId('echart')).not.toBeInTheDocument();
    expect(within(view).queryByText(/There are currently no incidents/i)).toBeInTheDocument();
  });
});
