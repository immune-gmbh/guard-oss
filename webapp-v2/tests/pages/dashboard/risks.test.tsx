/**
 * @jest-environment jsdom
 */

import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { within, render, screen } from '@testing-library/react';
import * as RisksHook from 'hooks/issues';
import DashboardRisks from 'pages/dashboard/risks';

import common from 'locales/en/common.json';
import utils from 'locales/en/utils.json';
import dashboard from 'locales/en/dashboard.json';
import risks from 'locales/en/risks.json';
import I18nProvider from 'next-translate/I18nProvider';
import { mockRiskList } from 'tests/mocks';

jest.mock('echarts-for-react', () => () =>  {
  return <div data-testid="echart" />
})

const mockUseIncidents = jest.spyOn(RisksHook, 'useRisks') as jest.Mock;

describe('risks list view', () => {

  test('renders incidents view correctly', () => {
    mockUseIncidents.mockImplementation(() => ({
      data: mockRiskList,
      isLoading: false
    }))

    render(
      <I18nProvider lang="en" namespaces={{ dashboard, risks, common, utils }}>
        <DashboardRisks />
      </I18nProvider>
    )

    const view = screen.getByRole('main')
    expect(within(view).queryByRole('table')).toBeInTheDocument()
    expect(within(view).queryByTestId('echart')).toBeInTheDocument()
  })

  test('renders risks view on empty status', () => {
    mockUseIncidents.mockImplementation(() => ({
      data: [],
      isLoading: false
    }))

    render(
      <I18nProvider lang="en" namespaces={{ dashboard, risks, common, utils }}>
        <DashboardRisks />
      </I18nProvider>
    )

    const view = screen.getByRole('main')
    expect(within(view).queryByRole('table')).not.toBeInTheDocument()
    expect(within(view).queryByTestId('echart')).not.toBeInTheDocument()
    expect(within(view).queryByText(/There are currently no risks/i)).toBeInTheDocument()
  })

})
