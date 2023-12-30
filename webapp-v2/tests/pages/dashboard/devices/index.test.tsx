/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { act, fireEvent, render, screen } from '@testing-library/react';
import * as InfiniteDevicesHook from 'hooks/infinityDevices';
import * as IntersectionObserverHook from 'hooks/intersectionObserver';
import * as SessionHook from 'hooks/useSession';
import dashboard from 'locales/en/dashboard.json';
import incidents from 'locales/en/incidents.json';
import risks from 'locales/en/risks.json';
import I18nProvider from 'next-translate/I18nProvider';
import { useRouter } from 'next/router';
import DevicesIndex from 'pages/dashboard/devices/index';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import { devicesMock } from 'tests/mocks';
import { ApiSrv } from 'types/apiSrv';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseInfiniteDeviceLoading = jest.spyOn(
  InfiniteDevicesHook,
  'useInfiniteDeviceLoading',
) as jest.Mock;
const mockIntersectionObserver = jest.spyOn(IntersectionObserverHook, 'default') as jest.Mock;
const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;

const mockLoadNext = jest.fn(() => ({}));
const mockLoadItems = jest.fn(() => Promise.resolve());

const mockInfiniteResponse = (response: ApiSrv.Device[], hasNext: boolean) => {
  return mockUseInfiniteDeviceLoading.mockImplementation(() => ({
    devices: response,
    hasNext,
    loadNext: mockLoadNext,
    isLoading: false,
    loadItems: mockLoadItems,
  }));
};

describe('devices index', () => {
  const ViewWithContext = (
    <I18nProvider lang="en" namespaces={{ dashboard, incidents, risks }}>
      <SelectedDevicesContext.Provider value={{ items: [], dispatch: () => ({}) }}>
        <DevicesIndex />
      </SelectedDevicesContext.Provider>
    </I18nProvider>
  );

  beforeEach(() => {
    (useRouter as jest.Mock).mockReturnValue({
      query: {},
      replace: () => ({}),
    });

    mockUseSession.mockReturnValue({
      session: {
        user: { id: '1' },
        currentMembership: {
          organisation: {
            id: '2',
          },
        },
      },
      isInitialized: true,
    });
  });

  afterEach(() => {
    mockUseInfiniteDeviceLoading.mockReset();
    mockIntersectionObserver.mockReset();
    mockUseSession.mockReset();
    jest.resetModules();
  });

  test('renders devices view correctly', async () => {
    mockInfiniteResponse(devicesMock, false);

    render(ViewWithContext);

    const devicesTable = screen.queryByRole('table');
    expect(devicesTable).toBeInTheDocument();
    expect(screen.queryAllByRole('listbox').length).toBe(2);
  });

  test('filters by name on the client side', async () => {
    mockInfiniteResponse(devicesMock, false);

    render(ViewWithContext);

    const searchInput = screen.getByRole('search');
    expect(screen.queryAllByRole('listbox').length).toBe(2);

    fireEvent.change(searchInput, { target: { value: 'lud' } });
    expect(screen.queryAllByRole('listbox').length).toBe(1);
    expect(screen.queryByText(/ludmilla/i)).toBeInTheDocument();
  });

  describe('calls loadNext when needed', () => {
    test('theres no next and button is not on viewport', async () => {
      mockInfiniteResponse(devicesMock, false);
      mockIntersectionObserver.mockImplementation(() => ({
        isIntersecting: false,
      }));

      render(ViewWithContext);

      await act(async () => {
        await new Promise((r) => setTimeout(r, 600));
      });

      expect(mockLoadNext).toHaveBeenCalledTimes(0);
    });

    test('theres no next but button is on viewport', async () => {
      mockInfiniteResponse(devicesMock, false);
      mockIntersectionObserver.mockImplementation(() => ({
        isIntersecting: true,
      }));

      render(ViewWithContext);

      await act(async () => {
        await new Promise((r) => setTimeout(r, 600));
      });

      expect(mockLoadNext).toHaveBeenCalledTimes(1);
    });

    test('theres next and button is on viewport', async () => {
      mockInfiniteResponse(devicesMock, true);
      mockIntersectionObserver.mockImplementation(() => ({
        isIntersecting: true,
      }));

      render(ViewWithContext);

      await act(async () => {
        await new Promise((r) => setTimeout(r, 600));
      });

      expect(mockLoadNext).toHaveBeenCalledTimes(2);
    });
  });
});
