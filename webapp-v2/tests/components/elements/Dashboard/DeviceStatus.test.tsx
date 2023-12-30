/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { act, render, screen } from '@testing-library/react';
import DeviceStatus, { parseBubbleValues } from 'components/elements/Dashboard/DeviceStatus';
import { IDeviceStats } from 'hooks/dashboard';
import * as OrganisationHook from 'hooks/organisations';
import * as SubscriptionHook from 'hooks/subscriptions';
import common from 'locales/en/common.json';
import dashboard from 'locales/en/dashboard.json';
import incidents from 'locales/en/incidents.json';
import risks from 'locales/en/risks.json';
import I18nProvider from 'next-translate/I18nProvider';
import { mockDeviceStats } from 'tests/mocks';

jest.mock('echarts-for-react', () => () => {
  return <div data-testid="echart" />;
});

describe('dashboard - device status', () => {
  const mockUseOrganisation = jest.spyOn(OrganisationHook, 'useOrganisation') as jest.Mock;
  const mockUseSubscription = jest.spyOn(SubscriptionHook, 'useSubscription') as jest.Mock;

  test('renders when there is no valid organisation ID or stats', async () => {
    render(
      <I18nProvider lang="en" namespaces={{ dashboard, risks, incidents, common }}>
        <DeviceStatus organisationId="orgaId" deviceStats={{} as IDeviceStats} />
      </I18nProvider>,
    );

    expect(screen.getByText(/Device Status/i)).toBeInTheDocument();
    expect(screen.queryByRole('status')).not.toBeInTheDocument();

    await act(async () => {
      await new Promise((r) => setTimeout(r, 600));
      expect(screen.getAllByTestId('echart').length).toBe(1);
    });

    expect(parseBubbleValues({} as IDeviceStats)).toEqual([]);
  });

  test('renders when there are stats', async () => {
    mockUseOrganisation.mockImplementation(() => ({
      data: {
        id: 'orgaId',
        subscription: {
          id: 'subsId',
        },
      },
    }));

    mockUseSubscription.mockImplementation(() => ({
      data: {
        id: 'subsId',
        currentDevicesAmount: 80,
        maxDevicesAmount: 100,
      },
    }));

    render(
      <I18nProvider lang="en" namespaces={{ dashboard, risks, incidents, common }}>
        <DeviceStatus organisationId="orgaId" deviceStats={mockDeviceStats} />
      </I18nProvider>,
    );

    await act(async () => {
      await new Promise((r) => setTimeout(r, 600));
      expect(screen.getAllByTestId('echart').length).toBe(1);
    });

    expect(screen.getByRole('status')).toContainHTML('<span class="font-semibold">80 / 100</span>');
    expect(parseBubbleValues(mockDeviceStats)[0]).toEqual(
      expect.objectContaining({
        color: '#e0e9c7',
        id: 'numTrusted',
        x: 0,
        y: 0,
        value: 102,
        size: 120,
      }),
    );
  });
});
