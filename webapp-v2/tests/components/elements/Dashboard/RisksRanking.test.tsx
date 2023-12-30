/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';
import RisksRanking from 'components/elements/Dashboard/RisksRanking';
import common from 'locales/en/common.json';
import dashboard from 'locales/en/dashboard.json';
import incidents from 'locales/en/incidents.json';
import risks from 'locales/en/risks.json';
import I18nProvider from 'next-translate/I18nProvider';
import { mockRisks } from 'tests/mocks';

jest.mock('echarts-for-react', () => () => {
  return <div data-testid="echart" />;
});

describe('dashboard - risks ranking', () => {
  test('renders when there are no risks', async () => {
    render(
      <I18nProvider lang="en" namespaces={{ dashboard, risks, incidents, common }}>
        <RisksRanking risks={[]} />
      </I18nProvider>,
    );

    expect(screen.getByText(/Top Risks/i)).toBeInTheDocument();
    expect(screen.queryByText(/Broken down by affected component/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/No risks found for the device fleet/i)).toBeInTheDocument();
    expect(screen.queryAllByTestId('echart')).toEqual([]);
  });

  test('renders when there are risks', async () => {
    render(
      <I18nProvider lang="en" namespaces={{ dashboard, risks, incidents, common }}>
        <RisksRanking risks={mockRisks} />
      </I18nProvider>,
    );

    const risksList = screen.queryByRole('list');

    expect(risksList).toBeInTheDocument();
    expect(risksList.childNodes.length).toBe(5);

    expect(risksList.firstChild.textContent).toBe('Critical Interfaces Accessible');
    expect(risksList.lastChild.textContent).toBe('Intel CSME Firmware Manipulated');
  });
});
